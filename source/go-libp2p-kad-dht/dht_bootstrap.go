package dht

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"

	u "github.com/ipfs/go-ipfs-util"
	"github.com/multiformats/go-multiaddr"
	_ "github.com/multiformats/go-multiaddr-dns"
)

var DefaultBootstrapPeers []multiaddr.Multiaddr

func init() {
	for _, s := range []string{
		"/ip4/119.3.159.159/tcp/4001/p2p/8MG8mghNNtP7LCJdoqWkrnUmfvNgKZ",
		"/ip4/39.100.145.251/tcp/4001/p2p/8MGRZbvn8caS431icB2P1uT74B3EHh",
		"/ip4/47.90.212.117/tcp/4001/p2p/8MKX58Ko5vBeJUkfgpkig53jZzwqoW",
		"/ip4/39.98.240.7/tcp/4001/p2p/8MHYzNkm6dF9SWU5u7Py8MJ31vJrzS",
		"/ip4/39.100.0.162/tcp/4001/p2p/8MJ5cAWfAP86cHmAcC3dxqzK41dh4a",
		"/ip4/47.90.252.189/tcp/4001/p2p/8MH4Woxb2FkM5nFr86dHj21fLgEybi",
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		DefaultBootstrapPeers = append(DefaultBootstrapPeers, ma)
	}
}

// BootstrapConfig specifies parameters used bootstrapping the DHT.
//
// Note there is a tradeoff between the bootstrap period and the
// number of queries. We could support a higher period with less
// queries.
type BootstrapConfig struct {
	Queries int           // how many queries to run per period
	Period  time.Duration // how often to run periodic bootstrap.
	Timeout time.Duration // how long to wait for a bootstrap query to run
}

var DefaultBootstrapConfig = BootstrapConfig{
	// For now, this is set to 1 query.
	// We are currently more interested in ensuring we have a properly formed
	// DHT than making sure our dht minimizes traffic. Once we are more certain
	// of our implementation's robustness, we should lower this down to 8 or 4.
	Queries: 1,

	// For now, this is set to 5 minutes, which is a medium period. We are
	// We are currently more interested in ensuring we have a properly formed
	// DHT than making sure our dht minimizes traffic.
	Period: time.Duration(5 * time.Minute),

	Timeout: time.Duration(10 * time.Second),
}

// A method in the IpfsRouting interface. It calls BootstrapWithConfig with
// the default bootstrap config.
func (dht *KadDHT) Bootstrap(ctx context.Context) error {
	return dht.BootstrapWithConfig(ctx, DefaultBootstrapConfig)
}

// Runs cfg.Queries bootstrap queries every cfg.Period.
func (dht *KadDHT) BootstrapWithConfig(ctx context.Context, cfg BootstrapConfig) error {
	// Because this method is not synchronous, we have to duplicate sanity
	// checks on the config so that callers aren't oblivious.
	if cfg.Queries <= 0 {
		return fmt.Errorf("invalid number of queries: %d", cfg.Queries)
	}
	go func() {
		for {
			err := dht.runBootstrap(ctx, cfg)
			if err != nil {
				logger.Warningf("error bootstrapping: %s", err)
			}
			select {
			case <-time.After(cfg.Period):
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

// This is a synchronous bootstrap. cfg.Queries queries will run each with a
// timeout of cfg.Timeout. cfg.Period is not used.
func (dht *KadDHT) BootstrapOnce(ctx context.Context, cfg BootstrapConfig) error {
	if cfg.Queries <= 0 {
		return fmt.Errorf("invalid number of queries: %d", cfg.Queries)
	}
	return dht.runBootstrap(ctx, cfg)
}

func newRandomPeerId() peer.ID {
	id := make([]byte, 32) // SHA256 is the default. TODO: Use a more canonical way to generate random IDs.
	rand.Read(id)
	id = u.Hash(id) // TODO: Feed this directly into the multihash instead of hashing it.
	return peer.ID(id)
}

// Traverse the DHT toward the given ID.
func (dht *KadDHT) walk(ctx context.Context, target peer.ID) (peer.AddrInfo, error) {
	// TODO: Extract the query action (traversal logic?) inside ,
	// don't actually call through the FindPeer machinery, which can return
	// things out of the peer store etc.
	return dht.FindPeer(ctx, target)
}

// Traverse the DHT toward a random ID.
func (dht *KadDHT) randomWalk(ctx context.Context) error {
	id := newRandomPeerId()
	p, err := dht.walk(ctx, id)
	switch err {
	case routing.ErrNotFound:
		return nil
	case nil:
		// We found a peer from a randomly generated ID. This should be very
		// unlikely.
		logger.Warningf("random walk toward %s actually found peer: %s", id, p)
		return nil
	default:
		return err
	}
}

// Traverse the DHT toward the self ID
func (dht *KadDHT) selfWalk(ctx context.Context) error {
	_, err := dht.walk(ctx, dht.self)
	if err == routing.ErrNotFound {
		return nil
	}
	return err
}

// runBootstrap builds up list of peers by requesting random peer IDs
func (dht *KadDHT) runBootstrap(ctx context.Context, cfg BootstrapConfig) error {
	doQuery := func(n int, target string, f func(context.Context) error) error {
		logger.Infof("starting bootstrap query (%d/%d) to %s (routing table size was %d)",
			n, cfg.Queries, target, dht.routingTable.Size())
		defer func() {
			logger.Infof("finished bootstrap query (%d/%d) to %s (routing table size is now %d)",
				n, cfg.Queries, target, dht.routingTable.Size())
		}()
		queryCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		err := f(queryCtx)
		if err == context.DeadlineExceeded && queryCtx.Err() == context.DeadlineExceeded && ctx.Err() == nil {
			return nil
		}
		return err
	}

	// Do all but one of the bootstrap queries as random walks.
	for i := 0; i < cfg.Queries; i++ {
		err := doQuery(i, "random ID", dht.randomWalk)
		if err != nil {
			return err
		}
	}

	// Find self to distribute peer info to our neighbors.
	return doQuery(cfg.Queries, fmt.Sprintf("self: %s", dht.self), dht.selfWalk)
}

func (dht *KadDHT) BootstrapRandom(ctx context.Context) error {
	return dht.randomWalk(ctx)
}

func (dht *KadDHT) BootstrapSelf(ctx context.Context) error {
	return dht.selfWalk(ctx)
}
