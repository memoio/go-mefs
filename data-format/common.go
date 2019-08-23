package dataformat

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"
	"strconv"

	"github.com/memoio/go-mefs/bls12"
)

const (
	DefaultSegmentSize = 4 * 1024 //(4k)
	MAXOFFSET          = 255      // 一个Stripe最多有256个field，最大offset为255
	RsPolicy           = 1
	MulPolicy          = 2
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

//Prefix represent a rawdata's structure
type Prefix struct {
	Policy      uint64 // 1为纠删冗余，2为多副本
	DataCount   uint64
	ParityCount uint64
	TagFlag     uint64
	SegmentSize uint64
	TagSize     uint64
	Len         int
}

//根据传入数据，生成相应的prefix，policy=1则为纠删冗余
func PrefixEncode(policy, dataCount, paritycount, tagFlag, segmentSize, tagSize uint64) ([]byte, error) {
	if _, ok := Codes[tagFlag]; !ok {
		return nil, ErrWrongTagFlag
	}
	if DefaultLengths[tagFlag] != tagSize {
		return nil, ErrWrongTagFlag
	}
	if policy != RsPolicy && policy != MulPolicy {
		return nil, ErrWrongPolicy
	}
	prefix := make([]byte, 6*binary.MaxVarintLen64)
	temp := prefix
	n := binary.PutUvarint(temp, policy)
	temp = prefix[n:]
	n += binary.PutUvarint(temp, dataCount)
	temp = prefix[n:]
	n += binary.PutUvarint(temp, paritycount)
	temp = prefix[n:]
	n += binary.PutUvarint(temp, tagFlag)
	temp = prefix[n:]
	n += binary.PutUvarint(temp, segmentSize)
	temp = prefix[n:]
	n += binary.PutUvarint(temp, tagSize)

	return prefix[:n], nil
}

//从[]byte头部解析出Prefix，并返回去前缀的数据
func PrefixDecode(rawdata []byte) (*Prefix, []byte, error) {
	pre := &Prefix{}
	var err error
	var count int
	pre.Policy, rawdata, count, err = uvarint(rawdata)
	if err != nil {
		return nil, nil, err
	}
	if pre.Policy != MulPolicy && pre.Policy != RsPolicy {
		return nil, nil, ErrWrongPolicy
	}
	pre.Len += count
	pre.DataCount, rawdata, count, err = uvarint(rawdata)
	if err != nil {
		return nil, nil, err
	}
	pre.Len += count
	pre.ParityCount, rawdata, count, err = uvarint(rawdata)
	if err != nil {
		return nil, nil, err
	}
	pre.Len += count
	pre.TagFlag, rawdata, count, err = uvarint(rawdata)
	if err != nil {
		return nil, nil, err
	}
	pre.Len += count
	pre.SegmentSize, rawdata, count, err = uvarint(rawdata)
	if err != nil {
		return nil, nil, err
	}
	pre.Len += count
	pre.TagSize, rawdata, count, err = uvarint(rawdata)
	if err != nil {
		return nil, nil, err
	}
	pre.Len += count
	if DefaultLengths[pre.TagFlag] != pre.TagSize {
		return nil, nil, ErrWrongTagFlag
	}
	return pre, rawdata, nil
}

func uvarint(buf []byte) (uint64, []byte, int, error) {
	n, c := binary.Uvarint(buf) //从buf解码unit64的值，并返回该值和读取的字节数

	if c == 0 {
		return n, buf, 0, ErrVarintBufferShort
	} else if c < 0 {
		return n, buf[-c:], c, ErrVarintTooLong
	} else {
		return n, buf[c:], c, nil
	}
}

//具体实现
func (prefix *Prefix) GetSegAndTagFromRawdata(noPreRawdata []byte, offset int) ([]byte, []byte, [][]byte, error) {
	var TagCount uint64
	switch prefix.Policy {
	case RsPolicy:
		TagCount = 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
	case MulPolicy:
		TagCount = prefix.DataCount + prefix.ParityCount
	default:
		return nil, nil, nil, ErrWrongPolicy
	}

	fieldSize := prefix.SegmentSize + TagCount*prefix.TagSize
	start := uint64(offset) * (fieldSize)

	if uint64(len(noPreRawdata)) < start+fieldSize {
		return nil, nil, nil, ErrDataTooShort
	}

	segment := noPreRawdata[start : start+prefix.SegmentSize]
	start += prefix.SegmentSize
	tag := noPreRawdata[start : start+prefix.TagSize]
	start += prefix.TagSize
	ptags := make([][]byte, TagCount-1)
	for i := 0; i < int(TagCount)-1; i++ {
		ptags[i] = noPreRawdata[start : start+prefix.TagSize]
		start += prefix.TagSize
	}

	return segment, tag, ptags, nil
}

//可用于Provider生成证明，直接从文件读取偏移位置数据，不用读取整块到内存中
func GetSegAndTagFromFile(path string, offset uint64) ([]byte, []byte, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	fileSize := fInfo.Size()

	fReader := bufio.NewReader(f)
	prefix := make([]byte, 6*binary.MaxVarintLen64)
	_, err = fReader.Read(prefix)
	if err != nil {
		return nil, nil, err
	}
	pre, _, err := PrefixDecode(prefix)
	if err != nil {
		return nil, nil, err
	}
	var TagCount uint64
	switch pre.Policy {
	case RsPolicy:
		TagCount = 1 + (pre.ParityCount-1)/pre.DataCount + 1
	case MulPolicy:
		TagCount = pre.DataCount + pre.ParityCount
	default:
		return nil, nil, ErrWrongPolicy
	}

	fieldSize := pre.SegmentSize + TagCount*pre.TagSize

	start := uint64(pre.Len) + offset*(fieldSize)
	if uint64(fileSize) < start+fieldSize {
		return nil, nil, ErrDataTooShort
	}

	segment := make([]byte, pre.SegmentSize)
	tag := make([]byte, pre.TagSize)
	//-------
	n, err := f.ReadAt(segment, int64(start))
	if err != nil {
		return nil, nil, err
	}
	if n != int(pre.SegmentSize) {
		return nil, nil, ErrCannotGetSegment
	}
	start += uint64(n)
	//-------
	n, err = f.ReadAt(tag, int64(start))
	if err != nil {
		return nil, nil, err
	}
	if n != int(pre.TagSize) {
		return nil, nil, ErrCannotGetSegment
	}
	return segment, tag, nil
}

//VerifyBlockLength：检查一个块的长度，从beginoffset开始，一个block至少要存的数据量，要么为dif对应的offset，要么填满
//整个块
func VerifyBlockLength(blockData []byte, beginoffset, tagFlag, segmentSize, dataCount, parityCount, dif int, policy int32) (bool, error) {
	if blockData == nil {
		return false, ErrDataTooShort
	}
	pre, noPreRawdata, err := PrefixDecode(blockData)
	if err != nil {
		return false, err
	}
	if pre.DataCount != uint64(dataCount) || pre.ParityCount != uint64(parityCount) ||
		pre.TagFlag != uint64(tagFlag) || pre.Policy != uint64(policy) ||
		pre.SegmentSize != uint64(segmentSize) {
		return false, ErrWrongField
	}

	//不同策略的标签数目不同
	var TagCount uint64
	switch pre.Policy {
	case RsPolicy:
		TagCount = 1 + (pre.ParityCount-1)/pre.DataCount + 1
	case MulPolicy:
		TagCount = pre.DataCount + pre.ParityCount
	default:
		return false, ErrWrongPolicy
	}
	fieldSize := pre.SegmentSize + TagCount*pre.TagSize
	if len(noPreRawdata)%int(fieldSize) != 0 { //fields有问题
		return false, ErrDataBroken
	}
	if dif+beginoffset*int(pre.SegmentSize)*dataCount > (MAXOFFSET+1)*int(pre.SegmentSize)*dataCount {
		if len(noPreRawdata)/int(fieldSize) < MAXOFFSET+1 {
			return false, ErrDataBroken
		}
	} else {
		//首先将数据分给每个块，计算每个块的数据量
		dataForEachBlock := (dif-1)/dataCount + 1
		//然后在每个块计算切成segment的数量
		count := (dataForEachBlock-1)/int(pre.SegmentSize) + 1
		if len(noPreRawdata)/int(fieldSize) < beginoffset+count {
			return false, ErrDataBroken
		}
	}

	return true, nil
}

//对数据进行验证，VerifyBlock传进来一个带前缀的完整块
//模拟挑战证明聚合验证，0.04s一个块
func VerifyBlock(blockData []byte, ncid string, pk *mcl.PublicKey) bool {
	if pk == nil || blockData == nil || len(blockData) == 0 {
		return false
	}

	var err error

	pre, noPreRawdata, err := PrefixDecode(blockData)
	if err != nil {
		return false
	}

	//不同策略的标签数目不同
	var TagCount uint64
	switch pre.Policy {
	case RsPolicy:
		TagCount = 1 + (pre.ParityCount-1)/pre.DataCount + 1
	case MulPolicy:
		TagCount = pre.DataCount + pre.ParityCount
	default:
		return false
	}

	fieldSize := pre.SegmentSize + TagCount*pre.TagSize
	if len(noPreRawdata)%int(fieldSize) != 0 { //fields有问题
		return false
	}
	count := (len(noPreRawdata) - 1) / int(fieldSize)
	segments := make([][]byte, count)
	tags := make([][]byte, count)
	indices := make([]string, count)
	for i := 0; i < count; i++ {
		indices[i] = ncid + "_" + strconv.Itoa(i)
		segments[i] = noPreRawdata[uint64(i)*fieldSize : uint64(i)*fieldSize+pre.SegmentSize]
		tags[i] = noPreRawdata[uint64(i)*fieldSize+pre.SegmentSize : uint64(i)*fieldSize+pre.SegmentSize+pre.TagSize]
	}
	ok, err := mcl.VerifyDataForUser(pk, indices, segments, tags)
	if !ok || err != nil {
		return false
	}
	return true
}
