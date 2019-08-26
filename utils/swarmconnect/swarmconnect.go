package swarmconnect

import (
	"context"
	"fmt"

	peer "github.com/libp2p/go-libp2p-core/peer"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/memoio/go-mefs/core"
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
			ctx = context.WithValue(ctx, "ExternIP", true)
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
