/*
Package core implements the MefsNode object and related methods.

Packages underneath core/ provide a (relatively) stable, low-level API
to carry out most MEFS-related tasks.  For more details on the other
interfaces and how core/... fits into the bigger MEFS picture, see:

  $ godoc github.com/memoio/go-mefs
*/
package core

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"
	goprocess "github.com/jbenet/goprocess"
	libp2p "github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	ifconnmgr "github.com/libp2p/go-libp2p-core/connmgr"
	p2phost "github.com/libp2p/go-libp2p-core/host"
	metrics "github.com/libp2p/go-libp2p-core/metrics"
	smux "github.com/libp2p/go-libp2p-core/mux"
	peer "github.com/libp2p/go-libp2p-core/peer"
	pstore "github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/routing"
	mplex "github.com/libp2p/go-libp2p-mplex"
	pnet "github.com/libp2p/go-libp2p-pnet"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	yamux "github.com/libp2p/go-libp2p-yamux"
	discovery "github.com/libp2p/go-libp2p/p2p/discovery"
	rhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	identify "github.com/libp2p/go-libp2p/p2p/protocol/identify"
	mafilter "github.com/libp2p/go-maddr-filter"

	"github.com/btcsuite/btcd/btcec"
	cy "github.com/libp2p/go-libp2p-core/crypto"
	p2pbhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	version "github.com/memoio/go-mefs"
	config "github.com/memoio/go-mefs/config"
	p2p "github.com/memoio/go-mefs/p2p"
	repo "github.com/memoio/go-mefs/repo"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/source/data"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	bstore "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	dhtopts "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/opts"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	ma "github.com/multiformats/go-multiaddr"
	mamask "github.com/whyrusleeping/multiaddr-filter"
)

const discoveryConnTimeout = time.Second * 30

var log = logging.Logger("core")

type mode int

const (
	// zero value is not a valid mode, must be explicitly set
	localMode mode = iota
	offlineMode
	onlineMode
)

func init() {
	identify.ClientVersion = "go-mefs/" + version.CurrentVersionNumber + "/" + version.CurrentCommit
}

// LocalNode uses global var to pass
var LocalNode *MefsNode

// MefsNode is MEFS Core module. It represents an MEFS instance.
type MefsNode struct {

	// Self
	Identity peer.ID // the local node's identity
	password string  // to decrypt the PrivateKey

	Repo repo.Repo

	// Local node
	PrivateKey      string // the local node's private Key(with format of eth, without prefix "0x")
	PNetFingerprint []byte // fingerprint of private network
	NetKey          string

	// Services
	Peerstore  pstore.Peerstore  // storage for other Peer instances
	Blockstore bstore.Blockstore // the raw blockstore, no filestore wrapping
	Data       data.Service      // the block service, get/add blocks.
	Reporter   metrics.Reporter
	Discovery  discovery.Service

	// Online
	PeerHost     p2phost.Host    // the network host (server+client)
	Bootstrapper io.Closer       // the periodic bootstrapper
	Routing      routing.Routing // the routing system. recommend ipfs-dht
	Inst         instance.Service

	P2P *p2p.P2P

	proc goprocess.Process
	ctx  context.Context

	mode         mode
	localModeSet bool
}

func (n *MefsNode) startOnlineServices(ctx context.Context, routingOption RoutingOption, hostOption HostOption, do DiscoveryOption, mplex bool) error {
	if n.PeerHost != nil { // already online.
		return errors.New("node already online")
	}

	// load private key
	if err := n.LoadPrivateKey(); err != nil {
		return err
	}

	// get undialable addrs from config
	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	var libp2pOpts []libp2p.Option
	for _, s := range cfg.Swarm.AddrFilters {
		f, err := mamask.NewMask(s)
		if err != nil {
			return fmt.Errorf("incorrectly formatted address filter in config: %s", s)
		}
		libp2pOpts = append(libp2pOpts, libp2p.FilterAddresses(f))
	}

	if !cfg.Swarm.DisableBandwidthMetrics {
		// Set reporter
		n.Reporter = metrics.NewBandwidthCounter()
		libp2pOpts = append(libp2pOpts, libp2p.BandwidthReporter(n.Reporter))
	}

	if n.NetKey != "" {
		fmt.Println("Using private network: ", n.NetKey)
		snKey := crypto.Keccak256([]byte(n.NetKey))
		snkeyPrefix := []byte("/key/swarm/psk/1.0.0/\n/base16/\n" + hex.EncodeToString(snKey))
		protec, err := pnet.NewProtector(bytes.NewReader(snkeyPrefix))
		if err != nil {
			return fmt.Errorf("failed to configure private network: %s", err)
		}
		n.PNetFingerprint = protec.Fingerprint()
		go func() {
			t := time.NewTicker(30 * time.Second)
			<-t.C // swallow one tick
			for {
				select {
				case <-t.C:
					if ph := n.PeerHost; ph != nil {
						if len(ph.Network().Peers()) == 0 {
							log.Warning("We are in private network and have no peers.")
							log.Warning("This might be configuration mistake.")
						}
					}
				case <-n.Process().Closing():
					t.Stop()
					return
				}
			}
		}()

		libp2pOpts = append(libp2pOpts, libp2p.PrivateNetwork(protec))
	}

	addrsFactory, err := makeAddrsFactory(cfg.Addresses)
	if err != nil {
		return err
	}
	if !cfg.Swarm.DisableRelay {
		addrsFactory = composeAddrsFactory(addrsFactory, filterRelayAddrs)
	}
	libp2pOpts = append(libp2pOpts, libp2p.AddrsFactory(addrsFactory))

	connm, err := constructConnMgr(cfg.Swarm.ConnMgr)
	if err != nil {
		return err
	}
	libp2pOpts = append(libp2pOpts, libp2p.ConnectionManager(connm))

	libp2pOpts = append(libp2pOpts, makeSmuxTransportOption(mplex))

	if !cfg.Swarm.DisableNatPortMap {
		libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())
	}

	// disable the default listen addrs
	libp2pOpts = append(libp2pOpts, libp2p.NoListenAddrs)

	if cfg.Swarm.DisableRelay {
		// Enabled by default.
		libp2pOpts = append(libp2pOpts, libp2p.DisableRelay())
	} else {
		relayOpts := []circuit.RelayOpt{circuit.OptDiscovery}
		if cfg.Swarm.EnableRelayHop {
			relayOpts = append(relayOpts, circuit.OptHop)
		}
		libp2pOpts = append(libp2pOpts, libp2p.EnableRelay(relayOpts...))
	}

	// explicitly enable the default transports
	libp2pOpts = append(libp2pOpts, libp2p.DefaultTransports)

	if cfg.Experimental.QUIC {
		libp2pOpts = append(libp2pOpts, libp2p.Transport(quic.NewTransport))
	}

	peerhost, err := hostOption(ctx, n.Identity, n.Peerstore, libp2pOpts...)

	if err != nil {
		return err
	}

	if err := n.startOnlineServicesWithHost(ctx, peerhost, routingOption); err != nil {
		return err
	}

	// Ok, now we're ready to listen.
	if err := startListening(n.PeerHost, cfg); err != nil {
		return err
	}

	n.P2P = p2p.NewP2P(n.Identity, n.PeerHost, n.Peerstore)

	// setup local discovery
	if do != nil {
		service, err := do(ctx, n.PeerHost)
		if err != nil {
			log.Error("mdns error: ", err)
		} else {
			service.RegisterNotifee(n)
			n.Discovery = service
		}
	}

	return n.Bootstrap(DefaultBootstrapConfig)
}

func constructConnMgr(cfg config.ConnMgr) (ifconnmgr.ConnManager, error) {
	switch cfg.Type {
	case "":
		// 'default' value is the basic connection manager
		return connmgr.NewConnManager(config.DefaultConnMgrLowWater, config.DefaultConnMgrHighWater, config.DefaultConnMgrGracePeriod), nil
	case "none":
		return nil, nil
	case "basic":
		grace, err := time.ParseDuration(cfg.GracePeriod)
		if err != nil {
			return nil, fmt.Errorf("parsing Swarm.ConnMgr.GracePeriod: %s", err)
		}

		return connmgr.NewConnManager(cfg.LowWater, cfg.HighWater, grace), nil
	default:
		return nil, fmt.Errorf("unrecognized ConnMgr.Type: %q", cfg.Type)
	}
}

func makeAddrsFactory(cfg config.Addresses) (p2pbhost.AddrsFactory, error) {
	var annAddrs []ma.Multiaddr
	for _, addr := range cfg.Announce {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		annAddrs = append(annAddrs, maddr)
	}

	filters := mafilter.NewFilters()
	noAnnAddrs := map[string]bool{}
	for _, addr := range cfg.NoAnnounce {
		f, err := mamask.NewMask(addr)
		if err == nil {
			filters.AddDialFilter(f)
			continue
		}
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		noAnnAddrs[maddr.String()] = true
	}

	return func(allAddrs []ma.Multiaddr) []ma.Multiaddr {
		var addrs []ma.Multiaddr
		if len(annAddrs) > 0 {
			addrs = annAddrs
		} else {
			addrs = allAddrs
		}

		var out []ma.Multiaddr
		for _, maddr := range addrs {
			// check for exact matches
			ok, _ := noAnnAddrs[maddr.String()]
			// check for /ipcidr matches
			if !ok && !filters.AddrBlocked(maddr) {
				out = append(out, maddr)
			}
		}
		return out
	}, nil
}

func makeSmuxTransportOption(mplexExp bool) libp2p.Option {
	const yamuxID = "/yamux/1.0.0"
	const mplexID = "/mplex/6.7.0"

	ymxtpt := *yamux.DefaultTransport
	ymxtpt.AcceptBacklog = 512

	if os.Getenv("YAMUX_DEBUG") != "" {
		ymxtpt.LogOutput = os.Stderr
	}

	muxers := map[string]smux.Multiplexer{yamuxID: &ymxtpt}
	if mplexExp {
		muxers[mplexID] = mplex.DefaultTransport
	}

	// Allow muxer preference order overriding
	order := []string{yamuxID, mplexID}
	if prefs := os.Getenv("LIBP2P_MUX_PREFS"); prefs != "" {
		order = strings.Fields(prefs)
	}

	opts := make([]libp2p.Option, 0, len(order))
	for _, id := range order {
		tpt, ok := muxers[id]
		if !ok {
			log.Warning("unknown or duplicate muxer in LIBP2P_MUX_PREFS: %s", id)
			continue
		}
		delete(muxers, id)
		opts = append(opts, libp2p.Muxer(id, tpt))
	}

	return libp2p.ChainOptions(opts...)
}

func setupDiscoveryOption(d config.Discovery) DiscoveryOption {
	if d.MDNS.Enabled {
		return func(ctx context.Context, h p2phost.Host) (discovery.Service, error) {
			if d.MDNS.Interval == 0 {
				d.MDNS.Interval = 5
			}
			return discovery.NewMdnsService(ctx, h, time.Duration(d.MDNS.Interval)*time.Second, discovery.ServiceTag)
		}
	}
	return nil
}

// HandlePeerFound attempts to connect to peer from `PeerInfo`, if it fails
// logs a warning log.
func (n *MefsNode) HandlePeerFound(p peer.AddrInfo) {
	log.Warning("trying peer info: ", p)
	ctx, cancel := context.WithTimeout(n.Context(), discoveryConnTimeout)
	defer cancel()
	if err := n.PeerHost.Connect(ctx, p); err != nil {
		log.Warning("Failed to connect to peer found by discovery: ", err)
	}
}

// startOnlineServicesWithHost  is the set of services which need to be
// initialized with the host and _before_ we start listening.
func (n *MefsNode) startOnlineServicesWithHost(ctx context.Context, host p2phost.Host, routingOption RoutingOption) error {
	// setup routing service
	r, err := routingOption(ctx, host, n.Repo.Datastore())
	if err != nil {
		return err
	}
	n.Routing = r

	// Wrap standard peer host with routing system to allow unknown peer lookups
	n.PeerHost = rhost.Wrap(host, n.Routing)
	return nil
}

// Process returns the Process object
func (n *MefsNode) Process() goprocess.Process {
	return n.proc
}

// Close calls Close() on the Process object
func (n *MefsNode) Close() error {
	return n.proc.Close()
}

// Context returns the MefsNode context
func (n *MefsNode) Context() context.Context {
	if n.ctx == nil {
		n.ctx = context.TODO()
	}
	return n.ctx
}

// Process returns the Process object
func (n *MefsNode) SetPassWord(pwd string) {
	n.password = pwd
}

// teardown closes owned children. If any errors occur, this function returns
// the first error.
func (n *MefsNode) teardown() error {
	log.Debug("core is shutting down...")
	// owned objects are closed in this teardown to ensure that they're closed
	// regardless of which constructor was used to add them to the node.
	var closers []io.Closer

	// NOTE: The order that objects are added(closed) matters, if an object
	// needs to use another during its shutdown/cleanup process, it should be
	// closed before that other object

	if n.Routing != nil {
		closers = append(closers, n.Routing.(*dht.KadDHT).Process())
	}

	if n.Bootstrapper != nil {
		closers = append(closers, n.Bootstrapper)
	}

	if n.PeerHost != nil {
		closers = append(closers, n.PeerHost)
	}

	// Repo closed last, most things need to preserve state here
	closers = append(closers, n.Repo)

	var errs []error
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// OnlineMode returns whether or not the MefsNode is in OnlineMode.
func (n *MefsNode) OnlineMode() bool {
	return n.mode == onlineMode
}

// SetLocal will set the MefsNode to local mode
func (n *MefsNode) SetLocal(isLocal bool) {
	if isLocal {
		n.mode = localMode
	}
	n.localModeSet = true
}

// LocalMode returns whether or not the MefsNode is in LocalMode
func (n *MefsNode) LocalMode() bool {
	if !n.localModeSet {
		// programmer error should not happen
		panic("local mode not set")
	}
	return n.mode == localMode
}

// Bootstrap will set and call the IpfsNodes bootstrap function.
func (n *MefsNode) Bootstrap(cfg BootstrapConfig) error {
	// TODO what should return value be when in offlineMode?
	if n.Routing == nil {
		return nil
	}

	if n.Bootstrapper != nil {
		n.Bootstrapper.Close() // stop previous bootstrap process.
	}

	// if the caller did not specify a bootstrap peer function, get the
	// freshest bootstrap peers from config. this responds to live changes.
	if cfg.BootstrapPeers == nil {
		cfg.BootstrapPeers = func() []peer.AddrInfo {
			ps, err := n.loadBootstrapPeers()
			if err != nil {
				log.Warning("failed to parse bootstrap peers from config")
				return nil
			}
			return ps
		}
	}

	var err error
	n.Bootstrapper, err = Bootstrap(n, cfg)
	return err
}

func (n *MefsNode) loadID() error {
	if n.Identity != "" {
		return errors.New("identity already loaded")
	}

	cfg, err := n.Repo.Config()
	if err != nil {
		return err
	}

	cid := cfg.PeerID
	if cid == "" {
		return errors.New("identity was not set in config (was 'mefs init' run?)")
	}
	if len(cid) == 0 {
		return errors.New("no peer ID in config! (was 'mefs init' run?)")
	}

	id, err := peer.IDB58Decode(cid)
	if err != nil {
		return fmt.Errorf("peer ID invalid: %s", err)
	}

	n.Identity = id
	return nil
}

//LoadPrivateKey load privatekey from keystore to setup MefsNode
func (n *MefsNode) LoadPrivateKey() error {
	if n.Identity == "" || n.Peerstore == nil {
		return errors.New("loaded private key out of order")
	}

	if n.PrivateKey != "" {
		log.Warning("private key already loaded")
		return nil
	}

	sk, err := fsrepo.GetPrivateKeyFromKeystore(n.Identity.Pretty(), n.password) //format of eth without prefix "0x"
	if err != nil {
		return err
	}

	n.PrivateKey = sk

	skEcdsa, err := utils.EthskToECDSAsk(sk)
	prik := (*cy.Secp256k1PrivateKey)((*btcec.PrivateKey)(skEcdsa))
	n.Peerstore.AddPrivKey(n.Identity, prik)
	n.Peerstore.AddPubKey(n.Identity, prik.GetPublic())
	return nil
}

func (n *MefsNode) loadBootstrapPeers() ([]peer.AddrInfo, error) {
	cfg, err := n.Repo.Config()
	if err != nil {
		return nil, err
	}

	parsed, err := cfg.BootstrapPeers()
	if err != nil {
		return nil, err
	}
	return toPeerInfos(parsed), nil
}

func listenAddresses(cfg *config.Config) ([]ma.Multiaddr, error) {
	var listen []ma.Multiaddr
	for _, addr := range cfg.Addresses.Swarm {
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			return nil, fmt.Errorf("failure to parse config.Addresses.Swarm: %s", cfg.Addresses.Swarm)
		}
		listen = append(listen, maddr)
	}

	return listen, nil
}

type ConstructPeerHostOpts struct {
	AddrsFactory      p2pbhost.AddrsFactory
	DisableNatPortMap bool
	DisableRelay      bool
	EnableRelayHop    bool
	ConnectionManager ifconnmgr.ConnManager
}

type HostOption func(ctx context.Context, id peer.ID, ps pstore.Peerstore, options ...libp2p.Option) (p2phost.Host, error)

var DefaultHostOption HostOption = constructPeerHost

// isolates the complex initialization steps
func constructPeerHost(ctx context.Context, id peer.ID, ps pstore.Peerstore, options ...libp2p.Option) (p2phost.Host, error) {
	pkey := ps.PrivKey(id)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.Pretty())
	}
	options = append([]libp2p.Option{libp2p.Identity(pkey), libp2p.Peerstore(ps)}, options...)
	return libp2p.New(ctx, options...)
}

func filterRelayAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	var raddrs []ma.Multiaddr
	for _, addr := range addrs {
		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			continue
		}
		raddrs = append(raddrs, addr)
	}
	return raddrs
}

func composeAddrsFactory(f, g p2pbhost.AddrsFactory) p2pbhost.AddrsFactory {
	return func(addrs []ma.Multiaddr) []ma.Multiaddr {
		return f(g(addrs))
	}
}

// startListening on the network addresses
func startListening(host p2phost.Host, cfg *config.Config) error {
	listenAddrs, err := listenAddresses(cfg)
	if err != nil {
		return err
	}

	// Actually start listening:
	if err := host.Network().Listen(listenAddrs...); err != nil {
		return err
	}

	// list out our addresses
	addrs, err := host.Network().InterfaceListenAddresses()
	if err != nil {
		return err
	}
	log.Infof("Swarm listening at: %s", addrs)
	return nil
}

func constructDHTRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching) (routing.Routing, error) {
	return dht.New(
		ctx, host,
		dhtopts.Datastore(dstore),
	)
}

func constructClientDHTRouting(ctx context.Context, host p2phost.Host, dstore ds.Batching) (routing.Routing, error) {
	return dht.New(
		ctx, host,
		dhtopts.Client(true),
		dhtopts.Datastore(dstore),
	)
}

type RoutingOption func(context.Context, p2phost.Host, ds.Batching) (routing.Routing, error)

type DiscoveryOption func(context.Context, p2phost.Host) (discovery.Service, error)

var DHTOption RoutingOption = constructDHTRouting
var DHTClientOption RoutingOption = constructClientDHTRouting
