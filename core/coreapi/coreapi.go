/*
Package coreapi provides direct access to the core commands in MEFS. If you are
embedding MEFS directly in your Go program, this package is the public
interface you should use to read and write files or otherwise control MEFS.

If you are running MEFS as a separate process, you should use `go-ipfs-api` to
work with it via HTTP. As we finalize the interfaces here, `go-ipfs-api` will
transparently adopt them so you can use the same code with either package.

**NOTE: this package is experimental.** `go-ipfs` has mainly been developed
as a standalone application and library-style use of this package is still new.
Interfaces here aren't yet completely stable.
*/
package coreapi

import (
	"context"

	core "github.com/memoio/go-mefs/core"
	coreiface "github.com/memoio/go-mefs/core/coreapi/interface"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("core/coreapi")

type CoreAPI struct {
	node *core.MefsNode
}

// NewCoreAPI creates new instance of MEFS CoreAPI backed by go-mefs Node.
func NewCoreAPI(n *core.MefsNode) coreiface.CoreAPI {
	api := &CoreAPI{n}
	return api
}

// Block returns the BlockAPI interface implementation backed by the go-mefs node
func (api *CoreAPI) Block() coreiface.BlockAPI {
	return (*BlockAPI)(api)
}

// Swarm returns the SwarmAPI interface implementation backed by the go-mefs node
func (api *CoreAPI) Swarm() coreiface.SwarmAPI {
	return (*SwarmAPI)(api)
}

// getSession returns new api backed by the same node with a read-only session DAG
func (api *CoreAPI) getSession(ctx context.Context) *CoreAPI {
	return &CoreAPI{api.node}
}
