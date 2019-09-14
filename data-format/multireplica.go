package dataformat

import (
	"errors"
	"strconv"

	mcl "github.com/memoio/go-mefs/bls12"
)

//将数据生成给定参数的多副本Stripe
func DataEncodeToMul(data []byte, ncidPrefix string, dataCount, parityCount int32, segmentSize, tagFlag uint64, keyset *mcl.KeySet) ([][]byte, uint64, error) {
	if segmentSize < DefaultSegmentSize { //最小为4k
		segmentSize = DefaultSegmentSize
	}
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}
	tagSize, ok := DefaultLengths[tagFlag]
	if !ok {
		return nil, 0, ErrWrongTagFlag
	}
	BlockCount := dataCount + parityCount

	prefix, err := PrefixEncode(MulPolicy, uint64(dataCount), uint64(parityCount), tagFlag, segmentSize, tagSize)
	if err != nil {
		return nil, 0, err
	}

	segmentCount := uint64(len(data)-1)/segmentSize + 1 //此data分解成segment的数量

	stripe := make([][]byte, BlockCount)
	fieldSize := segmentSize + tagSize*uint64(BlockCount)
	blockSize := fieldSize*segmentCount + uint64(len(prefix))

	for i := 0; i < int(BlockCount); i++ {
		stripe[i] = make([]byte, 0, blockSize)
		stripe[i] = append(stripe[i], prefix...)
	}
	remainder := uint64(len(data)) % segmentSize
	if remainder != 0 {
		data = append(data, make([]byte, segmentSize-remainder)...) ////TODO:目前都用零补全，以后为了安全，应用随机数
	}

	for i := 0; uint64(i) < segmentCount; i++ {
		segment := data[uint64(i)*segmentSize : uint64(i+1)*segmentSize]
		tags := make([][]byte, BlockCount)
		//填充此segment并生成tags
		for j := 0; j < int(BlockCount); j++ {
			stripe[j] = append(stripe[j], segment...)
			index := []byte(ncidPrefix + "_" + strconv.Itoa(j) + "_" + strconv.Itoa(i))
			tag, err := GenTagForSegment(segment, index, tagFlag, segmentSize, keyset)
			if err != nil {
				return nil, 0, err
			}
			tags[j] = tag
		}
		//将tags放进stripe
		//规则为，第一个为当前BlockID_offset对应的tag，其他的按顺序追加（除自己的）
		for j := 0; j < int(BlockCount); j++ {
			stripe[j] = append(stripe[j], tags[j]...)
			for k := 0; k < int(BlockCount); k++ {
				if j == k {
					continue
				}
				stripe[j] = append(stripe[j], tags[k]...)
			}
		}
	}

	return stripe, segmentCount - 1, nil
}

//生成的数据用于追加
func DataEncodeToMulForAppend(data []byte, ncidPrefix string, dataCount, parityCount int32, segmentSize, tagFlag uint64, beginOffset int, keyset *mcl.KeySet) ([][]byte, uint64, error) {
	if segmentSize < DefaultSegmentSize { //最小为4k
		segmentSize = DefaultSegmentSize
	}
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}
	tagSize, ok := DefaultLengths[tagFlag]
	if !ok {
		return nil, 0, ErrWrongTagFlag
	}
	BlockCount := dataCount + parityCount

	segmentCount := uint64(len(data)-1)/segmentSize + 1 //此data分解成segment的数量

	stripe := make([][]byte, BlockCount)
	fieldSize := segmentSize + tagSize*uint64(BlockCount) //一个segment及附带的tag占的大小
	blockSize := fieldSize * segmentCount

	for i := 0; i < int(BlockCount); i++ { //提前分配好内存
		stripe[i] = make([]byte, 0, blockSize)
	}
	remainder := uint64(len(data)) % segmentSize
	if remainder != 0 {
		data = append(data, make([]byte, segmentSize-remainder)...) ////TODO:目前都用零补全，以后为了安全，应用随机数
	}

	for i := beginOffset; i < beginOffset+int(segmentCount); i++ {
		segment := data[uint64(i-beginOffset)*segmentSize : uint64(i-beginOffset+1)*segmentSize]
		tags := make([][]byte, BlockCount)
		//讲Segment放进Stripe顺便生成tags
		for j := 0; j < int(BlockCount); j++ {
			stripe[j] = append(stripe[j], segment...)
			index := []byte(ncidPrefix + "_" + strconv.Itoa(j) + "_" + strconv.Itoa(i))
			tag, err := GenTagForSegment(segment, index, tagFlag, segmentSize, keyset)
			if err != nil {
				return nil, 0, err
			}
			tags[j] = tag
		}
		//将tags放进stripe
		//规则为，第一个为当前BlockID_offset对应的tag，其他的按顺序追加（除自己的）
		for j := 0; j < int(BlockCount); j++ {
			stripe[j] = append(stripe[j], tags[j]...)
			for k := 0; k < int(BlockCount); k++ {
				if j == k {
					continue
				}
				stripe[j] = append(stripe[j], tags[k]...)
			}
		}
	}

	return stripe, uint64(beginOffset) + segmentCount - 1, nil

}

//获得指定偏移位置以后指定数目的Segment拼合，可指定从此位置取多少个segment拼合的数据
//count为-1表示从指定offset取到最后
// 返回拼接的数据包括原数据在末尾补的零，需要用文件长度截取到正确的数据
func GetSegsFromData(rawdata []byte, beginOffset, count int) ([]byte, error) {
	pre, noPreRawdata, err := PrefixDecode(rawdata)
	if err != nil {
		return nil, err
	}
	return pre.getSegsFromRawdata(noPreRawdata, beginOffset, count)
}

//具体实现
func (prefix *Prefix) getSegsFromRawdata(noPreRawdata []byte, beginOffset, count int) ([]byte, error) {
	var TagCount uint64
	switch prefix.Policy {
	case RsPolicy:
		TagCount = 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
	case MulPolicy:
		TagCount = prefix.DataCount + prefix.ParityCount
	default:
		return nil, ErrWrongPolicy
	}

	fieldSize := prefix.SegmentSize + TagCount*prefix.TagSize
	start := uint64(beginOffset) * (fieldSize)
	if count == -1 {
		count = (len(noPreRawdata)-1)/int(fieldSize) + 1 - beginOffset //取到最后
	}
	if uint64(len(noPreRawdata)) < start+uint64(count)*fieldSize {
		return nil, ErrDataTooShort
	}
	if count > 0 {
		segments := make([]byte, uint64(count)*prefix.SegmentSize)
		for i := 0; i < count; i++ {
			segment := noPreRawdata[start+uint64(i)*fieldSize : start+uint64(i)*fieldSize+prefix.SegmentSize]
			copy(segments[uint64(i)*prefix.SegmentSize:], segment)
		}
		return segments, nil
	}
	return nil, errors.New("count error")
}

//从编码数据中获得所有数据的拼合，元数据块多备份模式可用
// 返回拼接的数据包括原数据在末尾补的零，需要用文件长度截取到正确的数据
func GetDataFromRawData(rawdata []byte) ([]byte, error) {
	pre, rawdata, err := PrefixDecode(rawdata)
	if err != nil {
		return nil, err
	}
	return pre.getDataFromRawdata(rawdata)
}

//从编码数据中获得所有数据的拼合，元数据块多备份模式可用
func (prefix *Prefix) getDataFromRawdata(noPreRawdata []byte) ([]byte, error) {
	var TagCount uint64
	switch prefix.Policy {
	case RsPolicy:
		TagCount = 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
	case MulPolicy:
		TagCount = prefix.DataCount + prefix.ParityCount
	default:
		return nil, ErrWrongPolicy
	}

	fieldSize := prefix.SegmentSize + TagCount*prefix.TagSize

	count := (len(noPreRawdata)-1)/int(fieldSize) + 1 //取到最后

	if uint64(len(noPreRawdata)) < uint64(count)*fieldSize {
		return nil, ErrDataTooShort
	}
	if count > 0 {
		segments := make([]byte, uint64(count)*prefix.SegmentSize)
		for i := 0; i < count; i++ {
			segment := noPreRawdata[uint64(i)*fieldSize : uint64(i)*fieldSize+prefix.SegmentSize]
			copy(segments[uint64(i)*prefix.SegmentSize:], segment)
		}
		return segments, nil
	}
	return nil, errors.New("count error")
}

// 多副本修复策略，传入需要恢复的其中一个副本块，以及该副本块在整个stripe中的块号
func RecoverMul(rawdata []byte, blockOrd int) ([][]byte, error) {
	// 解析前缀，获取副本数，segmentsize和tagsize
	pre, _, err := PrefixDecode(rawdata)
	if err != nil {
		return nil, err
	}
	stripeNum := pre.DataCount + pre.ParityCount
	fieldSize := pre.SegmentSize + pre.TagSize*stripeNum
	maxOffset := (len(rawdata)-pre.Len-1)/int(fieldSize) + 1
	// 创建含前缀的stripe
	stripe := make([][]byte, stripeNum)
	for i := 0; i < len(stripe); i++ {
		stripe[i] = make([]byte, 0, len(rawdata))
		stripe[i] = append(stripe[i], rawdata[:pre.Len]...)
	}
	for i := 0; i < maxOffset; i++ {
		// 获取指定offset的segment、tag
		segment, tag, tagP, err := pre.GetSegAndTagFromRawdata(rawdata[pre.Len:], i)
		if err != nil {
			return nil, err
		}
		// 将tag排好序
		tagGroup := make([][]byte, stripeNum)
		m := 0
		for j := 0; j < len(stripe); j++ {
			if j == blockOrd {
				tagGroup[blockOrd] = tag
			} else {
				tagGroup[j] = tagP[m]
				m++
			}
		}
		for j := 0; j < len(stripe); j++ {
			tmpData := segment
			stripe[j] = append(stripe[j], tmpData...)
			stripe[j] = append(stripe[j], tagGroup[j]...)
			for k := 0; k < len(stripe); k++ {
				if j == k {
					continue
				}
				stripe[j] = append(stripe[j], tagGroup[k]...)
			}
		}
	}
	return stripe, nil
}
