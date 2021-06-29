package config

import (
	"errors"
	"fmt"

	iaddr "github.com/ipfs/go-ipfs-addr"
	// Needs to be imported so that users can import this package directly
	// and still parse the bootstrap addresses.
	_ "github.com/multiformats/go-multiaddr-dns"
)

// DefaultBootstrapAddresses are the hardcoded bootstrap addresses
// for MEFS. they are nodes run by the MEFS team. docs on these later.
// As with all p2p networks, bootstrap is an important security concern.
//
// NOTE: This is here -- and not inside cmd/ipfs/init.go -- because of an
// import dependency issue. TODO: move this into a config/default/ package.
var DefaultBootstrapAddresses = []string{
	"/ip4/103.39.231.220/tcp/14069/p2p/8MHSTe6UWXsBYLgFKdEU3VnrnVPRUp",
	"/ip4/103.39.231.220/tcp/14061/p2p/8MGRqoRZisBforqyu8eCim2e1MiohV",
	"/ip4/103.39.231.220/tcp/14064/p2p/8MHsfVNmRYiCNUaofa1b34VoymnRm4",
	"/ip4/39.98.240.7/tcp/4001/p2p/8MK6qHvAfayLQy4d2684NTLAVLpQFk",
	"/ip4/39.100.0.162/tcp/4001/p2p/8MJ5cAWfAP86cHmAcC3dxqzK41dh4a",
	"/ip4/47.92.5.51/tcp/4001/p2p/8MH4Woxb2FkM5nFr86dHj21fLgEybi",
}

// BootstrapPeer is a peer used to bootstrap the network.
type BootstrapPeer iaddr.IPFSAddr

// ErrInvalidPeerAddr signals an address is not a valid peer address.
var ErrInvalidPeerAddr = errors.New("invalid peer address")

func (c *Config) BootstrapPeers() ([]BootstrapPeer, error) {
	return ParseBootstrapPeers(c.Bootstrap)
}

// DefaultBootstrapPeers returns the (parsed) set of default bootstrap peers.
// if it fails, it returns a meaningful error for the user.
// This is here (and not inside cmd/ipfs/init) because of module dependency problems.
func DefaultBootstrapPeers() ([]BootstrapPeer, error) {
	ps, err := ParseBootstrapPeers(DefaultBootstrapAddresses)
	if err != nil {
		return nil, fmt.Errorf(`failed to parse hardcoded bootstrap peers: %s
This is a problem with the mefs codebase. Please report it to the dev team.`, err)
	}
	return ps, nil
}

func (c *Config) SetBootstrapPeers(bps []BootstrapPeer) {
	c.Bootstrap = BootstrapPeerStrings(bps)
}

func ParseBootstrapPeer(addr string) (BootstrapPeer, error) {
	ia, err := iaddr.ParseString(addr)
	if err != nil {
		return nil, err
	}
	return BootstrapPeer(ia), err
}

func ParseBootstrapPeers(addrs []string) ([]BootstrapPeer, error) {
	peers := make([]BootstrapPeer, len(addrs))
	var err error
	for i, addr := range addrs {
		peers[i], err = ParseBootstrapPeer(addr)
		if err != nil {
			return nil, err
		}
	}
	return peers, nil
}

func BootstrapPeerStrings(bps []BootstrapPeer) []string {
	bpss := make([]string, len(bps))
	for i, p := range bps {
		bpss[i] = p.String()
	}
	return bpss
}
