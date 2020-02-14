package metainfo

import (
	"strings"

	peer "github.com/libp2p/go-libp2p-core/peer"
)

// BlockMeta is
type BlockMeta struct {
	queryID  string
	bucketID string
	stripeID string
	chunkID  string
}

func (bm *BlockMeta) GetQid() string {
	if bm == nil {
		return ""
	}
	return bm.queryID
}

func (bm *BlockMeta) GetBid() string {
	if bm == nil {
		return ""
	}
	return bm.bucketID
}

func (bm *BlockMeta) GetSid() string {
	if bm == nil {
		return ""
	}
	return bm.stripeID
}

func (bm *BlockMeta) GetCid() string {
	if bm == nil {
		return ""
	}
	return bm.chunkID
}

func (bm *BlockMeta) SetCid(cid string) {
	if bm == nil {
		return
	}
	bm.chunkID = cid
}

// ToString 将BlockMeta结构体转换成字符串格式，进行传输
func (bm *BlockMeta) ToString(prefix ...int) string {
	if bm == nil {
		return ""
	}
	outLength := 4
	if len(prefix) > 0 && prefix[0] > 0 && prefix[0] < 4 {
		outLength = prefix[0]
	}
	res := strings.Join([]string{bm.queryID, bm.bucketID, bm.stripeID, bm.chunkID}[:outLength], BlockDelimiter)
	return res
}

func NewBlockMeta(qid, bid, sid, cid string) (*BlockMeta, error) {
	_, err := peer.IDB58Decode(qid)
	if err != nil {
		return nil, err
	}

	return &BlockMeta{
		queryID:  qid,
		bucketID: bid,
		stripeID: sid,
		chunkID:  cid,
	}, nil
}

//GetBlockMeta 对于传入的key进行整理，返回结构体KeyMeta
func GetBlockMeta(key string) (*BlockMeta, error) {
	splitedKey := strings.Split(key, BlockDelimiter)
	if len(splitedKey) < 3 {
		return nil, ErrIllegalKey
	}

	if len(splitedKey) == 3 {
		return NewBlockMeta(splitedKey[0], splitedKey[0], splitedKey[1], splitedKey[2])
	}

	return NewBlockMeta(splitedKey[0], splitedKey[1], splitedKey[2], splitedKey[3])
}

//GetCidFromBlock retur cid:
// if key == uid_bucketid_sid_blockid, returns bucketid_sid_blockid
// if key == bucketid_sid_blockid, returns bucketid_sid_blockid
func GetCidFromBlock(key string) (string, error) {
	splitedKey := strings.Split(key, BlockDelimiter)

	if len(splitedKey) == 3 {
		return key, nil
	}

	if len(splitedKey) == 4 {
		return strings.Join(splitedKey[1:], BlockDelimiter), nil
	}

	return "", ErrIllegalKey
}

// GetIDsFromBlock returns bucketid;
// if key == uid_bucketid_sid_blockid, returns bucketid
// if key == bucketid_sid_blockid, returns bucketid
func GetIDsFromBlock(key string) (string, string, string, error) {
	splitedKey := strings.Split(key, BlockDelimiter)
	if len(splitedKey) == 3 {
		return splitedKey[0], splitedKey[1], splitedKey[2], nil
	}
	if len(splitedKey) == 4 {
		return splitedKey[1], splitedKey[2], splitedKey[3], nil
	}
	return "", "", "", ErrIllegalKey
}
