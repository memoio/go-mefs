package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	goprocess "github.com/jbenet/goprocess"
	procctx "github.com/jbenet/goprocess/context"
	periodicproc "github.com/jbenet/goprocess/periodic"
	host "github.com/libp2p/go-libp2p-core/host"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	lgbl "github.com/libp2p/go-libp2p-loggables"
	config "github.com/memoio/go-mefs/config"
)

// ErrNotEnoughBootstrapPeers signals that we do not have enough bootstrap
// peers to bootstrap correctly.
var ErrNotEnoughBootstrapPeers = errors.New("not enough bootstrap peers to bootstrap")

// BootstrapConfig specifies parameters used in an MefsNode's network
// bootstrapping process.
type BootstrapConfig struct {

	// MinPeerThreshold governs whether to bootstrap more connections. If the
	// node has less open connections than this number, it will open connections
	// to the bootstrap nodes. From there, the routing system should be able
	// to use the connections to the bootstrap nodes to connect to even more
	// peers. Routing systems like the KadDHT do so in their own Bootstrap
	// process, which issues random queries to find more peers.
	MinPeerThreshold int

	// Period governs the periodic interval at which the node will
	// attempt to bootstrap. The bootstrap process is not very expensive, so
	// this threshold can afford to be small (<=30s).
	Period time.Duration

	// ConnectionTimeout determines how long to wait for a bootstrap
	// connection attempt before cancelling it.
	ConnectionTimeout time.Duration

	// BootstrapPeers is a function that returns a set of bootstrap peers
	// for the bootstrap process to use. This makes it possible for clients
	// to control the peers the process uses at any moment.
	BootstrapPeers func() []peer.AddrInfo
}

// DefaultBootstrapConfig specifies default sane parameters for bootstrapping.
var DefaultBootstrapConfig = BootstrapConfig{
	MinPeerThreshold:  5,
	Period:            2 * time.Minute,
	ConnectionTimeout: (30 * time.Second) / 3, // Perod / 3
}

func BootstrapConfigWithPeers(pis []peer.AddrInfo) BootstrapConfig {
	cfg := DefaultBootstrapConfig
	cfg.BootstrapPeers = func() []peer.AddrInfo {
		return pis
	}
	return cfg
}

// Bootstrap kicks off MefsNode bootstrapping. This function will periodically
// check the number of open connections and -- if there are too few -- initiate
// connections to well-known bootstrap peers. It also kicks off subsystem
// bootstrapping (i.e. routing).
func Bootstrap(n *MefsNode, cfg BootstrapConfig) (io.Closer, error) {

	// make a signal to wait for one bootstrap round to complete.
	doneWithRound := make(chan struct{})

	if len(cfg.BootstrapPeers()) == 0 {
		// We *need* to bootstrap but we have no bootstrap peers
		// configured *at all*, inform the user.
		log.Error("no bootstrap nodes configured: go-mefs may have difficulty connecting to the network")
	}

	// the periodic bootstrap function -- the connection supervisor
	periodic := func(worker goprocess.Process) {
		ctx := procctx.OnClosingContext(worker)
		defer log.EventBegin(ctx, "periodicBootstrap", n.Identity).Done()

		if err := bootstrapRound(ctx, n.PeerHost, cfg); err != nil {
			log.Event(ctx, "bootstrapError", n.Identity, lgbl.Error(err))
			log.Debugf("%s bootstrap error: %s", n.Identity, err)
		}

		<-doneWithRound
	}

	// kick off the node's periodic bootstrapping
	proc := periodicproc.Tick(cfg.Period, periodic)
	proc.Go(periodic) // run one right now.

	// kick off Routing.Bootstrap?
	// run dht.Bootstrap, we need to reach out
	if n.Routing != nil {
		ctx := procctx.OnClosingContext(proc)
		if err := n.Routing.Bootstrap(ctx); err != nil {
			proc.Close()
			return nil, err
		}
	}

	doneWithRound <- struct{}{}
	close(doneWithRound) // it no longer blocks periodic
	return proc, nil
}

func bootstrapRound(ctx context.Context, host host.Host, cfg BootstrapConfig) error {

	ctx, cancel := context.WithTimeout(ctx, cfg.ConnectionTimeout)
	defer cancel()
	id := host.ID()

	// get bootstrap peers from config. retrieving them here makes
	// sure we remain observant of changes to client configuration.
	peers := cfg.BootstrapPeers()
	// determine how many bootstrap connections to open
	numToDial := cfg.MinPeerThreshold

	// filter out bootstrap nodes we are already connected to
	var notConnected []peer.AddrInfo
	for _, p := range peers {
		if host.Network().Connectedness(p.ID) != inet.Connected {
			notConnected = append(notConnected, p)
		}
	}

	// if connected to all bootstrap peer candidates, exit
	if len(notConnected) < 1 {
		log.Debugf("%s no more bootstrap peers to create %d connections", id, numToDial)
		return ErrNotEnoughBootstrapPeers
	}

	// connect to a random susbset of bootstrap candidates
	randSubset := randomSubsetOfPeers(notConnected, numToDial)

	defer log.EventBegin(ctx, "bootstrapStart", id).Done()
	log.Debugf("%s bootstrapping to %d nodes: %s", id, numToDial, randSubset)
	return bootstrapConnect(ctx, host, randSubset)
}

func bootstrapConnect(ctx context.Context, ph host.Host, peers []peer.AddrInfo) error {
	if len(peers) < 1 {
		return ErrNotEnoughBootstrapPeers
	}

	errs := make(chan error, len(peers))
	var wg sync.WaitGroup
	for _, p := range peers {

		// performed asynchronously because when performed synchronously, if
		// one `Connect` call hangs, subsequent calls are more likely to
		// fail/abort due to an expiring context.
		// Also, performed asynchronously for dial speed.

		wg.Add(1)
		go func(p peer.AddrInfo) {
			defer wg.Done()
			defer log.EventBegin(ctx, "bootstrapDial", ph.ID(), p.ID).Done()
			log.Debugf("%s bootstrapping to %s", ph.ID(), p.ID)

			ph.Peerstore().AddAddrs(p.ID, p.Addrs, pstore.PermanentAddrTTL)
			if err := ph.Connect(ctx, p); err != nil {
				log.Event(ctx, "bootstrapDialFailed", p.ID)
				log.Debugf("failed to bootstrap with %v: %s", p.ID, err)
				errs <- err
				return
			}
			log.Event(ctx, "bootstrapDialSuccess", p.ID)
			log.Infof("bootstrapped with %v", p.ID)
		}(p)
	}
	wg.Wait()

	// our failure condition is when no connection attempt succeeded.
	// So drain the errs channel, counting the results.
	close(errs)
	count := 0
	var err error
	for err = range errs {
		if err != nil {
			count++
		}
	}
	if count == len(peers) {
		return fmt.Errorf("failed to bootstrap. %s", err)
	}
	return nil
}

func toPeerInfos(bpeers []config.BootstrapPeer) []peer.AddrInfo {
	pinfos := make(map[peer.ID]*peer.AddrInfo)
	for _, bootstrap := range bpeers {
		pinfo, ok := pinfos[bootstrap.ID()]
		if !ok {
			pinfo = new(peer.AddrInfo)
			pinfos[bootstrap.ID()] = pinfo
			pinfo.ID = bootstrap.ID()
		}

		pinfo.Addrs = append(pinfo.Addrs, bootstrap.Transport())
	}

	var peers []peer.AddrInfo
	for _, pinfo := range pinfos {
		peers = append(peers, *pinfo)
	}

	return peers
}

func randomSubsetOfPeers(in []peer.AddrInfo, max int) []peer.AddrInfo {
	n := max
	if max > len(in) {
		n = len(in)
	}

	var out []peer.AddrInfo
	for _, val := range rand.Perm(len(in)) {
		out = append(out, in[val])
		if len(out) >= n {
			break
		}
	}
	return out
}
