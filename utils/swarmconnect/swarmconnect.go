package swarmconnect

import (
	"context"
	"fmt"
	"strings"

	peer "github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/core"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//连接试三次
func ConnectTo(ctx context.Context, node *core.MefsNode, to string) bool {
	if node.PeerHost == nil {
		return false
	}

	id, err := peer.IDB58Decode(to)
	if err != nil {
		return false
	}

	var retry = false
	connectTryCount := 3
	for i := 0; i <= connectTryCount; i++ {
		if retry { // retry three times
			return getAddrAndConnect(node, to)
		}

		pi, err := node.Routing.FindPeer(ctx, id)
		if err != nil {
			fmt.Printf("findpeer err: %s\n", err)
			return false
		}

		if swrm, ok := node.PeerHost.Network().(*swarm.Swarm); ok {
			swrm.Backoff().Clear(pi.ID)
		}

		err = node.PeerHost.Connect(ctx, pi)
		if err == nil {
			return true
		}
		retry = true
	}
	return false
}

func getAddrAndConnect(node *core.MefsNode, to string) bool {
	km, err := metainfo.NewKeyMeta(to, metainfo.GetPeerAddr)
	if err != nil {
		return false
	}

	id, err := peer.IDB58Decode(to)
	if err != nil {
		return false
	}

	pi := peer.AddrInfo{}

	for _, defaultBootstrapAddress := range config.DefaultBootstrapAddresses {
		addr := strings.Split(defaultBootstrapAddress, "/")
		peerID := addr[len(addr)-1]
		res, err := node.Routing.(*dht.IpfsDHT).SendMetaRequest(km.ToString(), "", peerID, "GetPeerAddr")
		if err != nil {
			continue
		}
		pi.ID = id
		fmt.Println("get: ", res)
	}
	return false
}
