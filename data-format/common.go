package dataformat

import (
	"errors"
	"strconv"

	mcl "github.com/memoio/go-mefs/crypto/bls12"
	mpb "github.com/memoio/go-mefs/pb"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	RsPolicy            = 1
	MulPolicy           = 2
	DefaultSegmentSize  = 32 * 1024
	DefaultSegmentCount = 64
	DefaultTagFlag      = BLS12
	CurrentVersion      = 1
	DefaultCrypt        = 1
	BlockSize           = DefaultSegmentSize * DefaultSegmentCount
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

// DefaultBucketOptions is default bucket option
func DefaultBucketOptions() *mpb.BucketOptions {
	return &mpb.BucketOptions{
		Version:      1,
		Policy:       RsPolicy,
		DataCount:    3,
		ParityCount:  2,
		SegmentSize:  DefaultSegmentSize,
		TagFlag:      BLS12,
		SegmentCount: DefaultSegmentCount,
		Encryption:   1,
	}
}

// DefaultSuperBucketOptions is default supberbucket option
func DefaultSuperBucketOptions() *mpb.BucketOptions {
	return &mpb.BucketOptions{
		Version:      1,
		Policy:       MulPolicy,
		DataCount:    1,
		ParityCount:  2,
		SegmentSize:  DefaultSegmentSize,
		TagFlag:      BLS12,
		SegmentCount: DefaultSegmentCount,
		Encryption:   0,
	}
}

//VerifyBlockLength verify blocks length
func VerifyBlockLength(data []byte, start, length int) (bool, error) {
	if data == nil {
		return false, ErrDataTooShort
	}
	pre, preLen, err := bf.PrefixDecode(data)
	if err != nil || pre.GetBopts().GetVersion() == 0 || pre.GetBopts().GetDataCount() == 0 {
		return false, err
	}

	if int(pre.Start) != start {
		utils.MLogger.Error("VerifyBlockLength has start: ", pre.Start, ", need start: ", start)
		return false, errors.New("wrong data")
	}

	dataLen := len(data) - preLen

	s, ok := TagMap[int(pre.GetBopts().GetTagFlag())]
	if !ok {
		s = 48
	}

	fieldSize := int(pre.GetBopts().GetSegmentSize()) + s*int(2+(pre.GetBopts().ParityCount-1)/pre.GetBopts().DataCount)

	if dataLen != length*fieldSize {
		utils.MLogger.Error("VerifyBlockLength has length: ", dataLen, ", need: ", length*fieldSize)
		return false, errors.New("wrong data")
	}

	return true, nil
}

//VerifyBlock is 传进来一个带前缀的完整块
//模拟挑战证明聚合验证，0.04s一个块
func (d *DataCoder) VerifyBlock(data []byte, ncid string) ([][]byte, [][]byte, int, bool) {
	if data == nil || len(data) == 0 {
		return nil, nil, 0, false
	}

	pre, preLen, err := bf.PrefixDecode(data)
	if err != nil || pre.GetBopts().GetVersion() == 0 || pre.GetBopts().GetDataCount() == 0 {
		utils.MLogger.Error("prefix is not good: ", pre)
		return nil, nil, 0, false
	}

	d.Prefix = pre

	d.PreCompute()

	noPreRawdata := data[preLen:]

	count := len(noPreRawdata) / d.fieldSize
	if count <= 0 {
		utils.MLogger.Errorf("%s has short len: %d", ncid, len(noPreRawdata))
		return nil, nil, 0, false
	}

	segments := make([][]byte, count)
	tags := make([][]byte, count)
	indices := make([]string, count)
	for i := 0; i < count; i++ {
		indices[i] = ncid + metainfo.BlockDelimiter + strconv.Itoa(int(pre.Start)+i)
		segments[i] = append(segments[i], noPreRawdata[i*d.fieldSize:i*d.fieldSize+d.segSize]...)
		tags[i] = append(tags[i], noPreRawdata[i*d.fieldSize+d.segSize:i*d.fieldSize+d.segSize+d.tagSize]...)
	}

	ok, err := d.BlsKey.VerifyDataForUser(indices, segments, tags, 32)
	if !ok || err != nil {
		utils.MLogger.Error("Tag is wrong: ", err)
		return nil, nil, 0, false
	}
	return segments, tags, int(pre.Start), true
}

func VerifyBlock(data []byte, ncid string, k *mcl.KeySet) bool {
	if data == nil || len(data) == 0 || k == nil || k.Pk == nil {
		return false
	}

	d := &DataCoder{
		BlsKey: k,
	}

	_, _, _, ok := d.VerifyBlock(data, ncid)
	return ok
}

func GetSegAndTag(data []byte, ncid string, k *mcl.KeySet) ([][]byte, [][]byte, int, bool) {
	if data == nil || len(data) == 0 || k == nil || k.Pk == nil {
		return nil, nil, 0, false
	}

	d := &DataCoder{
		BlsKey: k,
	}
	return d.VerifyBlock(data, ncid)
}
