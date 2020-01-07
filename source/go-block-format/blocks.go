// Package blocks contains the lowest level of IPLD data structures.
// A block is raw data accompanied by a CID. The CID contains the multihash
// corresponding to the block.
package blocks

import (
	"errors"
	"fmt"
	"log"

	proto "github.com/gogo/protobuf/proto"
	u "github.com/ipfs/go-ipfs-util"
	pb "github.com/memoio/go-mefs/source/go-block-format/pb"
	cid "github.com/memoio/go-mefs/source/go-cid"
	mh "github.com/multiformats/go-multihash"
)

// ErrWrongHash is returned when the Cid of a block is not the expected
// according to the contents. It is currently used only when debugging.
var ErrWrongHash = errors.New("data did not match given hash")

//identify current block check method, default is 1(CRC32)
const defaultFlag int32 = 1
const checkSize = 4 * 1024 //(4k)

// Block provides abstraction for blocks implementations.
type Block interface {
	RawData() []byte
	Cid() cid.Cid
	String() string
	Loggable() map[string]interface{}
}

// A BasicBlock is a singular block of data in ipfs. It implements the Block
// interface.
type BasicBlock struct {
	cid  cid.Cid
	data []byte
}

// NewBlock creates a Block object from opaque data. It will hash the data.
func NewBlock(data []byte) *BasicBlock {
	// TODO: fix assumptions
	return &BasicBlock{data: data, cid: cid.NewCidV0(u.Hash(data))}
}

// NewBlockWithCid creates a new block when the hash of the data
// is already known, this is used to save time in situations where
// we are able to be confident that the data is correct.
func NewBlockWithCid(data []byte, c cid.Cid) (*BasicBlock, error) {
	if u.Debug {
		chkc, err := c.Prefix().Sum(data)
		if err != nil {
			return nil, err
		}

		if !chkc.Equals(c) {
			return nil, ErrWrongHash
		}
	}
	return &BasicBlock{data: data, cid: c}, nil
}

// Multihash returns the hash contained in the block CID.
func (b *BasicBlock) Multihash() mh.Multihash {
	return b.cid.Hash()
}

// RawData returns the block raw contents as a byte slice.
func (b *BasicBlock) RawData() []byte {
	return b.data
}

// Cid returns the content identifier of the block.
func (b *BasicBlock) Cid() cid.Cid {
	return b.cid
}

// String provides a human-readable representation of the block CID.
func (b *BasicBlock) String() string {
	return fmt.Sprintf("[Block %s]", b.Cid())
}

// Loggable returns a go-log loggable item.
func (b *BasicBlock) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"block": b.Cid().String(),
	}
}

func (b *BasicBlock) Prefix() (*pb.Prefix, int, error) {
	return PrefixDecode(b.RawData())
}

func PrefixLen(data []byte) (int, int, error) {
	len, n := proto.DecodeVarint(data[:10])
	if n <= 0 {
		return 0, 0, errors.New("wrong proto prefix message")
	}

	return int(len), n + int(len), nil
}

func PrefixDecode(data []byte) (*pb.Prefix, int, error) {
	log.Println(data[:60])

	x, n := proto.DecodeVarint(data[:10])
	if n <= 0 || x == 0 {
		log.Println("wrong proto prefix message:", x, n)
		return nil, 0, errors.New("wrong proto prefix message")
	}

	if n+int(x) > len(data) {
		log.Println("short proto prefix message:", x, n)
		return nil, 0, errors.New("short proto prefix message")
	}

	pre := new(pb.Prefix)
	err := proto.Unmarshal(data[n:n+int(x)], pre)
	if err != nil {
		return nil, 0, err
	}
	return pre, n + int(x), nil
}

func PrefixEncode(pre *pb.Prefix) ([]byte, int, error) {
	preData, err := proto.Marshal(pre)
	if err != nil {
		fmt.Println(err)
		return nil, 0, err
	}

	buf := proto.EncodeVarint(uint64(len(preData)))

	buf = append(buf, preData...)

	return buf, len(buf), nil
}
