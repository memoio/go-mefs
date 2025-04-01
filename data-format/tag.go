package dataformat

import (
	"encoding/binary"
	"hash/crc32"

	"github.com/memoio/go-mefs/crypto/pdp"
)

//GenTagForSegment 根据指定段大小生成标签，index是生成BLS-tag的需要
func (d *DataCoder) GenTagForSegment(index, data []byte) ([]byte, error) {
	switch d.Prefix.GetBopts().TagFlag {
	case pdp.CRC32:
		return uint32ToBytes(crc32.ChecksumIEEE(data)), nil
	case pdp.BLS:
		return nil, ErrWrongTagFlag
	case pdp.PDPV0:
		res, err := d.BlsKey.GenTag(index, data, 0, 32, true)
		if err != nil {
			return nil, err
		}
		return res, nil
	case pdp.PDPV1:
		return nil, ErrWrongTagFlag
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
