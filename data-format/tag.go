package dataformat

import (
	"encoding/binary"
	"hash/crc32"
)

// Tag constants
const (
	CRC32 = 1
	BLS   = 2
	BLS12 = 3
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
var TagMap = map[int]int{
	4:  CRC32,
	32: BLS,
	48: BLS12,
}

//根据指定段大小生成标签，index是生成BLS-tag的需要
func (d *DataCoder) GenTagForSegment(index, data []byte) ([]byte, error) {
	tagFlag, ok := TagMap[int(d.Prefix.TagSize)]
	if !ok {
		tagFlag = BLS12
	}

	switch tagFlag {
	case CRC32:
		return uint32ToBytes(crc32.ChecksumIEEE(data)), nil
	case BLS:
		return nil, ErrWrongTagFlag
	case BLS12:
		return d.BlsKey.GenTag(index, data, 0, 32, true)
	default:
		return nil, ErrWrongTagFlag
	}
}

//将uint32切片转成[]byte
func uint32ToBytes(vs uint32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, vs)
	return buf
}
