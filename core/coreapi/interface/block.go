package iface

import (
	"context"
	"io"

	options "github.com/memoio/go-mefs/core/coreapi/interface/options"
	cid "github.com/memoio/go-mefs/source/go-cid"
)

// BlockStat contains information about a block
type BlockStat interface {
	// Size is the size of a block
	Size() int

	// Path returns path to the block
	Cid() cid.Cid
}

// BlockAPI specifies the interface to the block layer
type BlockAPI interface {
	// Put imports raw block data, hashing it using specified settings.
	Put(context.Context, io.Reader, string, ...options.BlockPutOption) (BlockStat, error)

	Putto(string, string, context.Context) (BlockStat, error)

	// Get attempts to resolve the path and return a reader for data in the block
	Get(context.Context, string) (io.Reader, error)

	GetFrom(context.Context, string, string) (io.Reader, error)
	// Rm removes the block specified by the path from local blockstore.
	// By default an error will be returned if the block can't be found locally.
	//
	// NOTE: If the specified block is pinned it won't be removed and no error
	// will be returned
	// Rm(context.Context, string, ...options.BlockRmOption) error

	// Stat returns information on
	Stat(context.Context, string) (BlockStat, error)
}
