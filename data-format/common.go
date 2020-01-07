package dataformat

import (
	"errors"
	"strconv"

	mcl "github.com/memoio/go-mefs/bls12"
	bf "github.com/memoio/go-mefs/source/go-block-format"
)

const (
	RsPolicy            = 1
	MulPolicy           = 2
	DefaultSegmentSize  = 4096
	DefaultSegmentCount = 256
	DefaultTagFlag      = BLS12
	CurrentVersion      = 1
)

var (
	ErrWrongTagFlag     = errors.New("no such tag flag")
	ErrWrongPolicy      = errors.New("no such policy")
	ErrDataTooShort     = errors.New("data is too short")
	ErrDataBroken       = errors.New("data format is wrong")
	ErrWrongField       = errors.New("error Wrong Field to append")
	ErrCannotGetSegment = errors.New("error cannot get segment")
	ErrDataToolong      = errors.New("input Data is too long for a block")
	ErrRepairCrash      = errors.New("repair crash")
	ErrRecoverData      = errors.New("The recovered data is incorrect")
)

//VerifyBlockLength  verify blocks length
func VerifyBlockLength(data []byte, start, length int) (bool, error) {
	if data == nil {
		return false, ErrDataTooShort
	}
	pre, preLen, err := bf.PrefixDecode(data)
	if err != nil {
		return false, err
	}

	tagSize, ok := TagMap[int(pre.TagFlag)]
	if !ok {
		return false, ErrWrongTagFlag
	}

	tagCount := int(2 + (pre.ParityCount-1)/pre.DataCount)

	fieldSize := int(pre.SegmentSize) + tagCount*tagSize

	dataLen := len(data)

	if (dataLen-preLen)/fieldSize < start+length {
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

	pre, preLen, err := bf.PrefixDecode(data)
	if err != nil {
		return false
	}

	d.Prefix = pre

	d.PreCompute()

	noPreRawdata := data[preLen:]

	count := 1 + (len(noPreRawdata)-1)/d.fieldSize

	segments := make([][]byte, count)
	tags := make([][]byte, count)
	indices := make([]string, count)
	for i := 0; i < count; i++ {
		indices[i] = ncid + "_" + strconv.Itoa(i)
		segments[i] = noPreRawdata[i*d.fieldSize : i*d.fieldSize+d.segSize]
		tags[i] = noPreRawdata[i*d.fieldSize+d.segSize : i*d.fieldSize+d.segSize+d.tagSize]
	}

	ok, err := d.BlsKey.VerifyDataForUser(indices, segments, tags, 32)
	if !ok || err != nil {
		return false
	}
	return true
}

func VerifyBlock(data []byte, ncid string, k *mcl.KeySet) bool {
	if data == nil || len(data) == 0 {
		return false
	}

	d := &DataCoder{
		BlsKey: k,
	}
	return d.VerifyBlock(data, ncid)
}
