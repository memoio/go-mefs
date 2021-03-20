package config

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"strings"
	"time"

	id "github.com/memoio/go-mefs/crypto/identity"
	"github.com/memoio/go-mefs/utils/address"
)

//Init init config
func Init(out io.Writer, sk string) (*Config, string, error) {
	var identity Identity

	identity, err := CreateIdentity(out, sk)
	if err != nil {
		return nil, "", err
	}

	bootstrapPeers, err := DefaultBootstrapPeers()
	if err != nil {
		return nil, "", err
	}

	datastore := DefaultDatastoreConfig()

	conf := &Config{
		Role: "user",
		API: API{
			HTTPHeaders: map[string][]string{},
		},

		// setup the node's default addresses.
		// NOTE: two swarm listen addrs, one tcp, one utp.
		Addresses: addressesConfig(),

		Datastore: datastore,
		Bootstrap: BootstrapPeerStrings(bootstrapPeers),
		PeerID:    identity.PeerID,
		Discovery: Discovery{
			MDNS: MDNS{
				Enabled:  true,
				Interval: 10,
			},
		},

		Routing: Routing{
			Type: "dht",
		},

		IsInit: true,
		Eth:    "http://119.147.213.219:8101",
		Test:   false,

		Gateway: Gateway{
			RootRedirect: "",
			Writable:     false,
			PathPrefixes: []string{},
			HTTPHeaders: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Allow-Methods": {"GET"},
				"Access-Control-Allow-Headers": {"X-Requested-With", "Range"},
			},
			APICommands: []string{},
		},
		Swarm: SwarmConfig{
			ConnMgr: ConnMgr{
				LowWater:    DefaultConnMgrLowWater,
				HighWater:   DefaultConnMgrHighWater,
				GracePeriod: DefaultConnMgrGracePeriod.String(),
				Type:        "basic",
			},
		},
	}

	return conf, identity.PrivKey, nil
}

//InitTestnet init config
func InitTestnet(out io.Writer, sk string) (*Config, string, error) {
	identity, err := CreateIdentity(out, sk)
	if err != nil {
		return nil, "", err
	}

	bootstrapPeers, err := DefaultBootstrapPeers()
	if err != nil {
		return nil, "", err
	}

	datastore := DefaultDatastoreConfig()

	conf := &Config{
		Role: "user",
		API: API{
			HTTPHeaders: map[string][]string{},
		},

		// setup the node's default addresses.
		// NOTE: two swarm listen addrs, one tcp, one utp.
		Addresses: addressesConfig(),

		Datastore: datastore,
		Bootstrap: BootstrapPeerStrings(bootstrapPeers),
		PeerID:    identity.PeerID,
		Discovery: Discovery{
			MDNS: MDNS{
				Enabled:  true,
				Interval: 10,
			},
		},

		Routing: Routing{
			Type: "dht",
		},

		IsInit: true,
		Eth:    "http://119.147.213.219:8101",
		Test:   false,

		Gateway: Gateway{
			RootRedirect: "",
			Writable:     false,
			PathPrefixes: []string{},
			HTTPHeaders: map[string][]string{
				"Access-Control-Allow-Origin":  {"*"},
				"Access-Control-Allow-Methods": {"GET"},
				"Access-Control-Allow-Headers": {"X-Requested-With", "Range"},
			},
			APICommands: []string{},
		},
		Swarm: SwarmConfig{
			ConnMgr: ConnMgr{
				LowWater:    DefaultConnMgrLowWater,
				HighWater:   DefaultConnMgrHighWater,
				GracePeriod: DefaultConnMgrGracePeriod.String(),
				Type:        "basic",
			},
		},
	}

	return conf, identity.PrivKey, nil
}

// DefaultConnMgrHighWater is the default value for the connection managers
// 'high water' mark
const DefaultConnMgrHighWater = 900

// DefaultConnMgrLowWater is the default value for the connection managers 'low
// water' mark
const DefaultConnMgrLowWater = 600

// DefaultConnMgrGracePeriod is the default value for the connection managers
// grace period
const DefaultConnMgrGracePeriod = time.Second * 20

func addressesConfig() Addresses {
	return Addresses{
		Swarm: []string{
			"/ip4/0.0.0.0/tcp/4001",
			// "/ip4/0.0.0.0/udp/4002/utp", // disabled for now.
			"/ip6/::/tcp/4001",
		},
		Announce:   []string{},
		NoAnnounce: []string{},
		API:        Strings{"/ip4/127.0.0.1/tcp/5001"},
		Gateway:    Strings{"/ip4/127.0.0.1/tcp/8080"},
	}
}

// DefaultDatastoreConfig is an internal function exported to aid in testing.
func DefaultDatastoreConfig() Datastore {
	return Datastore{
		StorageMax:         "1000GB",
		StorageGCWatermark: 90, // 90%
		GCPeriod:           "1h",
		BloomFilterSize:    0,
		Spec: map[string]interface{}{
			"type": "mount",
			"mounts": []interface{}{
				map[string]interface{}{
					"mountpoint": "/data",
					"type":       "measure",
					"prefix":     "flatfs.datastore",
					"child": map[string]interface{}{
						"type":      "flatfs",
						"path":      "data",
						"sync":      true,
						"shardFunc": "/repo/flatfs/shard/v1/next-to-last/2",
					},
				},
				map[string]interface{}{
					"mountpoint": "/",
					"type":       "measure",
					"prefix":     "leveldb.datastore",
					"child": map[string]interface{}{
						"type":        "levelds",
						"path":        "datastore",
						"compression": "none",
					},
				},
			},
		},
	}
}

// CreateIdentity initializes a new identity.
func CreateIdentity(out io.Writer, hexSk string) (Identity, error) {
	// TODO guard higher up
	ident := Identity{}

	var sk *ecdsa.PrivateKey
	if hexSk == "" {
		fmt.Fprintf(out, "generating Secp256k1 keypair...")
		tsk, err := id.Create()
		if err != nil {
			return ident, err
		}
		fmt.Fprintf(out, "done\n")
		sk = tsk
	} else {
		tsk, err := id.ECDSAStringToSk(hexSk)
		if err != nil {
			return ident, err
		}
		sk = tsk
	}

	pk, err := id.GetPubByteFromECDSA(sk)
	if err != nil {
		return ident, err
	}

	pid, err := id.GetIDFromPubKey(pk)
	if err != nil {
		return ident, err
	}
	fmt.Fprintf(out, "peer identity: %s\n", pid)
	addr, err := address.GetAddressFromID(pid)
	if err != nil {
		return ident, err
	}
	fmt.Fprintf(out, "peer address: %s\n", addr.String())

	npid, err := address.GetIDFromAddress(addr.String())
	if err != nil {
		return ident, err
	}
	if strings.Compare(pid, npid) != 0 {
		fmt.Fprintf(out, "peer identity from address is: %s\n", npid)
	}
	ident.PeerID = pid
	ident.PrivKey = id.ECDSAByteToString(id.ToECDSAByte(sk))
	fmt.Fprintf(out, "peer secretKey: %s\n", ident.PrivKey)

	return ident, nil
}
