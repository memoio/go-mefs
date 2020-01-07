package dataformat

import (
	"encoding/binary"
	"hash/crc32"
	"log"
)

// Tag constants
const (
	CRC32 = 1
	BLS   = 2
	BLS12 = 3
)

// DefaultLengths maps a hash code to it's default length
var TagMap = map[int]int{
	CRC32: 4,
	BLS:   32,
	BLS12: 48,
}

//根据指定段大小生成标签，index是生成BLS-tag的需要
func (d *DataCoder) GenTagForSegment(index, data []byte) ([]byte, error) {
	switch d.Prefix.TagFlag {
	case CRC32:
		return uint32ToBytes(crc32.ChecksumIEEE(data)), nil
	case BLS:
		return nil, ErrWrongTagFlag
	case BLS12:
		res, err := d.BlsKey.GenTag(index, data, 0, 32, true)
		if err != nil {
			return nil, err
		}
		indice := make([]string, 1)
		indice[0] = string(index)
		segs := make([][]byte, 1)
		segs[0] = data
		tags := make([][]byte, 1)
		tags[0] = res
		ok, err := d.BlsKey.VerifyDataForUser(indice, segs, tags, 32)
		if err != nil || !ok {
			log.Println("gentag and verify fails")
		}
		return res, nil
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
