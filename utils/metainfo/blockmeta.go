package metainfo

import (
	"strings"

	peer "github.com/libp2p/go-libp2p-core/peer"
)

type BlockMeta struct {
	uid string
	gid string
	sid string
	bid string
}

const BLOCK_DELIMITER = "_"

func (bm *BlockMeta) GetUid() string {
	if bm == nil {
		return ""
	}
	return bm.uid
}

func (bm *BlockMeta) GetGid() string {
	if bm == nil {
		return ""
	}
	return bm.gid
}

func (bm *BlockMeta) GetSid() string {
	if bm == nil {
		return ""
	}
	return bm.sid
}

func (bm *BlockMeta) GetBid() string {
	if bm == nil {
		return ""
	}
	return bm.bid
}

func (bm *BlockMeta) SetBid(bid string) {
	if bm == nil {
		return
	}
	bm.bid = bid
}

// ToString 将BlockMeta结构体转换成字符串格式，进行传输
func (bm *BlockMeta) ToString(prefix ...int) string {
	if bm == nil {
		return ""
	}
	outLength := 4
	if len(prefix) > 0 {
		outLength = prefix[0]
	}
	res := strings.Join([]string{bm.uid, bm.gid, bm.sid, bm.bid}[:outLength], BLOCK_DELIMITER)
	return res
}

func NewBlockMeta(uidString, gidString, sidString, bidString string) (*BlockMeta, error) {
	_, err := peer.IDB58Decode(uidString)
	if err != nil {
		return nil, err
	}

	return &BlockMeta{
		uid: uidString,
		gid: gidString,
		sid: sidString,
		bid: bidString,
	}, nil
}

//GetBlockMeta 对于传入的key进行整理，返回结构体KeyMeta
func GetBlockMeta(key string) (*BlockMeta, error) {
	splitedKey := strings.Split(key, BLOCK_DELIMITER)
	if len(splitedKey) < 4 {
		return nil, ErrIllegalKey
	}
	return NewBlockMeta(splitedKey[0], splitedKey[1], splitedKey[2], splitedKey[3])

}
