package coreapi

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"time"

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

	err = api.node.Data.PutBlock(b)
	if err != nil {
		return nil, err
	}

	return &BlockStat{cid: b.Cid(), size: len(data)}, nil
}

func (api *BlockAPI) Putto(scid string, pid string, ctx context.Context) (coreiface.BlockStat, error) {

	var ncid cid.Cid
	ncid = cid.NewCidV2([]byte(scid))

	block, err := api.node.Data.GetBlock(ctx, ncid)
	if err != nil {
		return nil, err
	}
	err = api.node.Data.PutBlockTo(block, pid)
	if err != nil {
		return nil, err
	}
	return &BlockStat{cid: block.Cid(), size: len(block.RawData())}, nil
}

func (api *BlockAPI) Get(ctx context.Context, p string) (io.Reader, error) {
	var ncid cid.Cid
	ncid = cid.NewCidV2([]byte(p))

	b, err := api.node.Data.GetBlock(ctx, ncid)
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
	b, err := api.node.Data.GetBlockFrom(ctx, peerid, p, 2*time.Minute, sig)
	if err != nil || b == nil {
		return nil, err
	}

	return bytes.NewReader(b.RawData()), nil
}

func (api *BlockAPI) Stat(ctx context.Context, p string) (coreiface.BlockStat, error) {
	ncid := cid.NewCidV2([]byte(p))

	b, err := api.node.Data.GetBlock(ctx, ncid)
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
