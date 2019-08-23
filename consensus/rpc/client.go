package rpc

import (
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/client"
)

func GetHTTPClient(rpcAddr string) *client.HTTP {
	return client.NewHTTP(rpcAddr, "/websocket")
}
func GetLocalClient(node *nm.Node) *client.Local {
	return client.NewLocal(node)
}
