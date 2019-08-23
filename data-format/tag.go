package dataformat

import (
	"encoding/binary"
	"errors"
	"hash/crc32"

	mcl "github.com/memoio/go-mefs/bls12"
)

// Tag constants
const (
	CRC32 = 0x00
	BLS   = 0x01
	BLS12 = 0X02
)

// Names maps the name of a tagFlag to the code
var Names = map[string]uint64{
	"CRC32": CRC32,
	"BLS":   BLS,
	"BLS12": BLS12,
}

// Codes maps a tagFlag to it's name
var Codes = map[uint64]string{
	CRC32: "CRC32",
	BLS:   "BLS",
	BLS12: "BLS12",
}

// DefaultLengths maps a hash code to it's default length
var DefaultLengths = map[uint64]uint64{
	CRC32: 4,
	BLS:   128,
	BLS12: 48,
}

//根据指定段大小生成标签，index是生成BLS-tag的需要
func GenTagForSegment(segment, index []byte, tagFlag, segmentSize uint64, keyset *mcl.KeySet) ([]byte, error) {
	if segmentSize < DefaultSegmentSize { //
		segmentSize = DefaultSegmentSize
	}
	if uint64(len(segment)) < segmentSize { //TODO:目前都用零补全，以后为了安全，应用随机数
		segment = append(segment, make([]byte, segmentSize-uint64(len(segment)))...)
	}
	switch tagFlag {
	case CRC32:
		return uint32ToBytes(crc32.ChecksumIEEE(segment)), nil
	case BLS:
		return nil, ErrWrongTagFlag
	case BLS12:
		return genBLS12Tag(keyset, segment, index)
	default:
		return nil, ErrWrongTagFlag
	}
}

func genBLS12Tag(keySet *mcl.KeySet, segment, index []byte) ([]byte, error) {
	if keySet == nil {
		return nil, errors.New("bls12 keyset not construct")
	}
	return mcl.GenTag(keySet, segment, index)
}

//将uint32切片转成[]byte
func uint32ToBytes(vs uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, vs)
	return buf
}
