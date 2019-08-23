// package blockservice implements a BlockService interface that provides
// a single GetBlock/AddBlock interface that seamlessly retrieves data either
// locally or from a remote peer through the exchange.
package blockservice

import (
	"context"
	"errors"
	"strings"
	"time"

	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	blockstore "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"

	logging "github.com/ipfs/go-log"
	routing "github.com/libp2p/go-libp2p-routing"
)

var log = logging.Logger("blockservice")

var ErrNotFound = errors.New("blockservice: key not found")

const NEWCIDLENGTH int = 30

// BlockService is a hybrid block datastore. It stores data in a local
// datastore and may retrieve data from a remote Exchange.
// It uses an internal `datastore.Datastore` instance to store values.
type BlockService interface {
	// GetBlock gets the requested block.
	GetBlock(ctx context.Context, c cid.Cid) (blocks.Block, error)

	GetBlockFrom(ctx context.Context, peerid string, ncid string, tim time.Duration, sig []byte) (blocks.Block, error)

	// AddBlock puts a given block to the underlying datastore
	PutBlock(o blocks.Block) error

	PutBlockTo(o blocks.Block, id string) error

	// DeleteBlock deletes the given block from the blockservice.
	DeleteBlock(o cid.Cid) error

	// Blockstore returns a reference to the underlying blockstore
	Blockstore() blockstore.Blockstore
}

type blockService struct {
	blockstore blockstore.Blockstore

	rt routing.IpfsRouting
	// If checkFirst is true then first check that a block doesn't
	// already exist to avoid republishing the block on the exchange.
	checkFirst bool
}

// NewBlockService creates a BlockService with given datastore instance.
func New(bs blockstore.Blockstore, rt routing.IpfsRouting) BlockService {
	if rt == nil {
		log.Warning("blockservice running in local (offline) mode.")
	}

	return &blockService{
		blockstore: bs,
		rt:         rt,
		checkFirst: true,
	}
}

// NewWriteThrough ceates a BlockService that guarantees writes will go
// through to the blockstore and are not skipped by cache checks.
func NewWriteThrough(bs blockstore.Blockstore, rt routing.IpfsRouting) BlockService {
	return &blockService{
		blockstore: bs,
		rt:         rt,
		checkFirst: false,
	}
}

// Blockstore returns the blockstore behind this blockservice.
func (s *blockService) Blockstore() blockstore.Blockstore {
	return s.blockstore
}

// AddBlock adds a particular block to the service, Putting it into the datastore.
// TODO pass a context into this if the remote.HasBlock is going to remain here.
func (s *blockService) PutBlock(o blocks.Block) error {
	c := o.Cid()
	if s.checkFirst {
		if has, err := s.blockstore.Has(c); has || err != nil {
			return err
		}
	}

	if err := s.blockstore.Put(o); err != nil {
		return err
	}

	return nil
}

func (s *blockService) PutBlockTo(o blocks.Block, pid string) error {
	if s.rt != nil {
		ncid := o.Cid().String()
		splitedNcid := strings.Split(ncid, metainfo.DELIMITER)
		if len(splitedNcid) < 2 {
			km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock)
			ncid = km.ToString()
		}
		_, err := s.rt.(*dht.IpfsDHT).SendMetaRequest(ncid, string(o.RawData()), pid, "putBlockTo")
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("Routing is nil")
}

// GetBlock retrieves a particular block from the service,
// Getting it from the datastore using the key (hash).
func (s *blockService) GetBlock(ctx context.Context, c cid.Cid) (blocks.Block, error) {
	block, err := s.blockstore.Get(c)
	if err == nil {
		return block, nil
	}

	return nil, err
}

func (s *blockService) GetBlockFrom(ctx context.Context, pid string, ncid string, tim time.Duration, sig []byte) (blocks.Block, error) {
	if s.rt != nil {
		km, err := metainfo.NewKeyMeta(ncid, metainfo.GetBlock, string(sig))
		if err != nil {
			return nil, err
		}
		bdata, err := s.rt.(*dht.IpfsDHT).SendMetaRequest(km.ToString(), "", pid, "getBlockFrom")
		if err != nil {
			return nil, err
		}
		c := cid.NewCidV2([]byte(ncid))
		b, err := blocks.NewBlockWithCid([]byte(bdata), c)
		if err != nil {
			return nil, err
		}
		return b, nil
	}
	return nil, errors.New("Routing is nil")
}

// DeleteBlock deletes a block in the blockservice from the datastore
func (s *blockService) DeleteBlock(c cid.Cid) error {
	err := s.blockstore.DeleteBlock(c)
	if err == nil {
		log.Event(context.TODO(), "BlockService.BlockDeleted", c)
	}
	return err
}
