package dataformat

import (
	"errors"
	"strconv"

	bf "github.com/memoio/go-mefs/source/go-block-format"
)

const (
	RsPolicy           = 1
	MulPolicy          = 2
	DefaultTagSize     = 48
	DefaultSegmentSize = 4096
	DefaultLength      = 256
)

var (
	ErrVarintBufferShort = errors.New("uvarint: buffer too small")
	ErrVarintTooLong     = errors.New("uvarint: varint too big (max 64bit)")
	ErrWrongTagFlag      = errors.New("no such tag flag")
	ErrWrongPolicy       = errors.New("no such policy")
	ErrDataTooShort      = errors.New("data is too short")
	ErrDataBroken        = errors.New("data format is wrong")
	ErrWrongField        = errors.New("error Wrong Field to append")
	ErrCannotGetSegment  = errors.New("error cannot get segment")
	ErrDataToolong       = errors.New("input Data is too long for a block")
	ErrRepairCrash       = errors.New("repair crash")
	ErrRecoverData       = errors.New("The recovered data is incorrect")
)

//VerifyBlockLength：检查一个块的长度，从beginoffset开始，一个block至少要存的数据量，要么为dif对应的offset，要么填满
//整个块
func VerifyBlockLength(data []byte, start, length int) (bool, error) {
	if data == nil {
		return false, ErrDataTooShort
	}
	pre, err := bf.PrefixDecode(data)
	if err != nil {
		return false, err
	}

	preData, err := bf.PrefixEncode(pre)
	if err != nil {
		return false, err
	}

	preLen := len(preData)

	tagCount := 2 + (pre.ParityCount-1)/pre.DataCount

	fieldSize := pre.SegmentSize + tagCount*pre.TagSize

	dataLen := len(data)

	if int32(dataLen-preLen)/fieldSize < int32(start+length) {
		return false, nil
	}

	return true, nil
}

//对数据进行验证，VerifyBlock传进来一个带前缀的完整块
//模拟挑战证明聚合验证，0.04s一个块
func (d *DataCoder) VerifyBlock(data []byte, ncid string) bool {
	if data == nil || len(data) == 0 {
		return false
	}

	pre, err := bf.PrefixDecode(data)
	if err != nil {
		return false
	}

	preData, err := bf.PrefixEncode(pre)
	if err != nil {
		return false
	}

	preLen := len(preData)

	tagCount := 2 + (pre.ParityCount-1)/pre.DataCount

	fieldSize := int(pre.SegmentSize + tagCount*pre.TagSize)

	noPreRawdata := data[preLen:]

	count := (len(noPreRawdata) - 1) / fieldSize

	segments := make([][]byte, count)
	tags := make([][]byte, count)
	indices := make([]string, count)
	for i := 0; i < count; i++ {
		indices[i] = ncid + "_" + strconv.Itoa(i)
		segments[i] = noPreRawdata[i*fieldSize : i*fieldSize+int(pre.SegmentSize)]
		tags[i] = noPreRawdata[i*fieldSize+int(pre.SegmentSize) : i*fieldSize+int(pre.SegmentSize+pre.TagSize)]
	}

	ok, err := d.BlsKey.VerifyDataForUser(indices, segments, tags, 32)
	if !ok || err != nil {
		return false
	}
	return true
}
