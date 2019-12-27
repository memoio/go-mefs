package coreapi

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"

	coreiface "github.com/memoio/go-mefs/core/coreapi/interface"
	caopts "github.com/memoio/go-mefs/core/coreapi/interface/options"
	"github.com/memoio/go-mefs/role/user"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
)

type BlockAPI CoreAPI

type BlockStat struct {
	cid  cid.Cid
	size int
}

const NEWCIDLENGTH int = 30

func (api *BlockAPI) Put(ctx context.Context, src io.Reader, ncid string, opts ...caopts.BlockPutOption) (coreiface.BlockStat, error) {
	data, err := ioutil.ReadAll(src)
	if err != nil {
		return nil, err
	}

	bcid := cid.NewCidV2([]byte(ncid))

	b, err := blocks.NewBlockWithCid(data, bcid)
	if err != nil {
		return nil, err
	}

	err = api.node.Data.PutBlock(ctx, ncid, data, "local")
	if err != nil {
		return nil, err
	}

	return &BlockStat{cid: b.Cid(), size: len(data)}, nil
}

func (api *BlockAPI) Putto(scid string, pid string, ctx context.Context) (coreiface.BlockStat, error) {
	block, err := api.node.Data.GetBlock(ctx, scid, nil, "local")
	if err != nil {
		return nil, err
	}
	err = api.node.Data.PutBlock(ctx, scid, block.RawData(), pid)
	if err != nil {
		return nil, err
	}
	return &BlockStat{cid: block.Cid(), size: len(block.RawData())}, nil
}

func (api *BlockAPI) Get(ctx context.Context, p string) (io.Reader, error) {
	b, err := api.node.Data.GetBlock(ctx, p, nil, "local")
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(b.RawData()), nil
}

func (api *BlockAPI) GetFrom(ctx context.Context, p string, peerid string) (io.Reader, error) {
	sig, err := user.BuildSignMessage()
	if err != nil {
		return nil, err
	}
	b, err := api.node.Data.GetBlock(ctx, p, sig, peerid)
	if err != nil || b == nil {
		return nil, err
	}

	return bytes.NewReader(b.RawData()), nil
}

func (api *BlockAPI) Stat(ctx context.Context, p string) (coreiface.BlockStat, error) {
	b, err := api.node.Data.GetBlock(ctx, p, nil, "local")
	if err != nil {
		return nil, err
	}

	return &BlockStat{
		cid:  b.Cid(),
		size: len(b.RawData()),
	}, nil
}

func (bs *BlockStat) Size() int {
	return bs.size
}

func (bs *BlockStat) Cid() cid.Cid {
	return bs.cid
}

func (api *BlockAPI) core() coreiface.CoreAPI {
	return (*CoreAPI)(api)
}
