package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"

	ci "github.com/libp2p/go-libp2p-core/crypto"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

//CreateID create other ID, the out can be os.Stdout, nBitsForKeypair is useless, but the default is 2048 by nitsForKeypairDefault
func CreateID(out io.Writer, nBitsForKeypair int) (identity Identity, err error) {
	identity, err = identityConfig(out, nBitsForKeypair)
	if err != nil {
		return identity, err
	}
	return identity, nil
}

//Init init config
func Init(out io.Writer, nBitsForKeypair int, sk string) (*Config, string, error) {
	var identity Identity
	var err error
	if sk == "" {
		identity, err = identityConfig(out, nBitsForKeypair)
		if err != nil {
			return nil, "", err
		}
	} else {
		identity, err = identityConfigSK(out, sk)
		if err != nil {
			return nil, "", err
		}
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
		Eth:    "http://47.92.5.51:8101",
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

//Init init config
func InitTestnet(out io.Writer, nBitsForKeypair int, sk string) (*Config, string, error) {
	var identity Identity
	var err error
	if sk == "" {
		identity, err = identityConfig(out, nBitsForKeypair)
		if err != nil {
			return nil, "", err
		}
	} else {
		identity, err = identityConfigSK(out, sk)
		if err != nil {
			return nil, "", err
		}
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
		Eth:    "http://47.92.5.51:8101",
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

// identityConfig initializes a new identity.
func identityConfig(out io.Writer, nbits int) (Identity, error) {
	// TODO guard higher up
	ident := Identity{}
	if nbits < 1024 {
		return ident, errors.New("bitsize less than 1024 is considered unsafe")
	}

	fmt.Fprintf(out, "generating Secp256k1 keypair...")
	sk, pk, err := ci.GenerateKeyPair(ci.Secp256k1, nbits)
	if err != nil {
		return ident, err
	}
	fmt.Fprintf(out, "done\n")

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	//以太坊的私钥是先将ECDSA产生的privatekey转换成2进制字节，再转换成带0x前缀的16进制字符串
	skbytes, err := sk.Bytes()
	if err != nil {
		return ident, err
	}
	ident.PrivKey = base64.StdEncoding.EncodeToString(skbytes)
	id, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return ident, err
	}
	ident.PeerID = id.Pretty()
	fmt.Fprintf(out, "peer identity: %s\n", ident.PeerID)
	addr, err := address.GetAddressFromID(ident.PeerID)
	if err != nil {
		return ident, err
	}
	fmt.Fprintf(out, "peer address: %s\n", addr.String())

	pid, err := address.GetIDFromAddress(addr.String())
	if err != nil {
		return ident, err
	}
	if strings.Compare(ident.PeerID, pid) != 0 {
		fmt.Fprintf(out, "peer identity from address is: %s\n", pid)
	}

	return ident, nil
}

func identityConfigSK(out io.Writer, hexsk string) (Identity, error) {
	ident := Identity{}

	sk, err := utils.HexskToIPFSsk(hexsk)
	if err != nil {
		return ident, err
	}

	// 根据sk获取pk
	skBytes, err := base64.StdEncoding.DecodeString(sk)
	if err != nil {
		return ident, err
	}
	prik, err := ci.UnmarshalPrivateKey(skBytes)
	if err != nil {
		return ident, err
	}
	pubk := prik.GetPublic()
	id, err := peer.IDFromPublicKey(pubk)
	if err != nil {
		return ident, err
	}
	peerID := id.Pretty()
	ident.PrivKey = sk
	ident.PeerID = peerID
	fmt.Fprintf(out, "peer identity: %s\n", ident.PeerID)

	addr, err := address.GetAddressFromID(ident.PeerID)
	if err != nil {
		return ident, err
	}
	fmt.Fprintf(out, "peer address: %s\n", addr.String())

	pid, err := address.GetIDFromAddress(addr.String())
	if err != nil {
		return ident, err
	}

	if strings.Compare(ident.PeerID, pid) != 0 {
		fmt.Fprintf(out, "peer identity from address is: %s\n", pid)
	}

	return ident, nil
}
