package dataformat

import (
	"errors"
	"fmt"
	"strconv"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/data-format/reedsolomon"
)

// 将传入的n个数据块编码成n+m个数据冗余块组的形式，并返回该数据冗余块组
func EncodeData(segments [][]byte, dataCount, parityCount int) ([][]byte, error) {
	enc, err := reedsolomon.New(dataCount, parityCount)
	if err != nil {
		return nil, err
	}
	perSize := len(segments[0])
	// 构建块组，数据块+冗余块
	newsegments := make([][]byte, dataCount+parityCount)
	for i := 0; i < len(newsegments); i++ {
		newsegments[i] = make([]byte, perSize)
	}
	for i := 0; i < dataCount; i++ {
		copy(newsegments[i], segments[i])
	}
	err = enc.Encode(newsegments)
	if err != nil {
		return nil, err
	}
	return newsegments, nil
}

// 将传入的数据冗余块组恢复，返回想要恢复的块，若index为-1则返回个块组
func RecoverData(datas [][]byte, dataCount, parityCount int, index ...int) ([][]byte, error) {
	// 根据传入的参数，恢复整个块组
	enc, err := reedsolomon.New(dataCount, parityCount)
	if err != nil {
		return nil, err
	}
	ok, err := enc.Verify(datas)
	if err == reedsolomon.ErrShardNoData || err == reedsolomon.ErrTooFewShards {
		return nil, err
	}
	if !ok {
		err = enc.Reconstruct(datas)
		if err != nil {
			return nil, err
		}
		ok, err = enc.Verify(datas)
		if !ok {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
	}
	// 根据传入的index，返回第几个块，若为-1则全返回
	if index[0] == -1 && len(index) == 1 {
		return datas, nil
	}
	blockgroup := make([][]byte, len(index))
	for i := 0; i < len(blockgroup); i++ {
		blockgroup[i] = make([]byte, len(datas[0]))
	}
	for key, val := range index {
		copy(blockgroup[key], datas[val])
	}
	return blockgroup, nil
}

// 将传入的Rawdata数据块编码成含前缀的规范化块组Stripe，并返回该Stripe及结束时的offset
func EncodeDataToPreStripe(data []byte, ncidPrefix string, dataCount, parityCount, tagflag int, segmentSize uint64, keyset *mcl.KeySet) ([][]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}
	tagNum := 1 + (parityCount-1)/dataCount + 1
	enc, err := reedsolomon.New(dataCount, parityCount)
	if err != nil {
		return nil, 0, err
	}
	encP, err := reedsolomon.New(dataCount+parityCount, (dataCount+parityCount)*(tagNum-1))
	if err != nil {
		return nil, 0, err
	}
	// 生成块组Stripe
	var stripe [][]byte
	realOffset := (len(data) - 1) / (int(segmentSize) * dataCount)
	tagSize, ok := DefaultLengths[uint64(tagflag)]
	if !ok {
		return nil, 0, ErrWrongTagFlag
	}
	prefix, err := PrefixEncode(RsPolicy, uint64(dataCount), uint64(parityCount), uint64(tagflag), segmentSize, tagSize)
	if err != nil {
		return nil, 0, err
	}
	stripe = createStripeHavePrefix(segmentSize, tagSize, prefix, dataCount, parityCount, realOffset+1)

	// 生成临时块组保存data切分后的segment
	tmpdata := creatGroup(dataCount+parityCount, segmentSize)
	// 生成taggroup装一组的tag+tagP
	taggroup := creatGroup((dataCount+parityCount)*tagNum, tagSize)

	for i := 0; i <= realOffset && data != nil; i++ {
		clearGroup(tmpdata)
		clearGroup(taggroup)
		for j := 0; j < dataCount; j++ {
			// 填充数据
			if segmentSize > uint64(len(data)) {
				copy(tmpdata[j], data)
				data = nil
				break
			}
			copy(tmpdata[j], data[:segmentSize])
			data = data[segmentSize:]
		}
		err = enc.Encode(tmpdata)
		if err != nil {
			return nil, 0, err
		}
		for j := 0; j < dataCount+parityCount; j++ {
			// 生成tag并装进taggroup，index为peerid_bucketid_stripeid_blockid_offsetid
			tag, err := GenTagForSegment(tmpdata[j], []byte(ncidPrefix+"_"+strconv.Itoa(j)+"_"+strconv.Itoa(i)), uint64(tagflag), segmentSize, keyset)
			if err != nil {
				return nil, 0, err
			}
			copy(taggroup[j], tag)
		}
		// 生成tag+tagP的taggroup格式
		err = encP.Encode(taggroup)
		if err != nil {
			return nil, 0, err
		}
		// 生成Field结构，此时beginOffset为下一个Field的起始偏移
		stripe = createFields(stripe, tmpdata, taggroup)
	}
	return stripe, realOffset, nil
}

// 将传入的Rawdata数据块编码成不含前缀的块组Stripe，并返回该Stripe。便于发送给provider进行append操作
func EncodeDataToNoPreStripe(data []byte, ncidPrefix string, dataCount, parityCount, tagflag, beginOffset int, segmentSize uint64, keyset *mcl.KeySet) ([][]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}
	tagNum := 1 + (parityCount-1)/dataCount + 1
	enc, err := reedsolomon.New(dataCount, parityCount)
	if err != nil {
		return nil, 0, err
	}
	encP, err := reedsolomon.New(dataCount+parityCount, (dataCount+parityCount)*(tagNum-1))
	if err != nil {
		return nil, 0, err
	}
	// 生成块组Stripe
	var stripe [][]byte
	realOffset := (len(data) - 1) / (int(segmentSize) * dataCount)
	tagSize, ok := DefaultLengths[uint64(tagflag)]
	if !ok {
		return nil, 0, ErrWrongTagFlag
	}
	stripe = createStripeNoPrefix(segmentSize, tagSize, dataCount, parityCount, realOffset+1)
	// 生成临时块组保存data切分后的segment
	tmpdata := creatGroup(dataCount+parityCount, segmentSize)
	// 生成taggroup装一组的tag+tagP
	taggroup := creatGroup((dataCount+parityCount)*tagNum, tagSize)

	maxOffset := realOffset + beginOffset
	for i := beginOffset; i <= maxOffset && data != nil; i++ {
		clearGroup(tmpdata)
		clearGroup(taggroup)
		for j := 0; j < dataCount; j++ {
			// 填充数据
			if segmentSize > uint64(len(data)) {
				copy(tmpdata[j], data)
				data = nil
				break
			}
			copy(tmpdata[j], data[:segmentSize])
			data = data[segmentSize:]
		}
		err = enc.Encode(tmpdata)
		if err != nil {
			return nil, 0, err
		}
		for j := 0; j < dataCount+parityCount; j++ {
			// 生成tag并装进taggroup，index为peerid_bucketid_stripeid_blockid_offsetid
			tag, err := GenTagForSegment(tmpdata[j], []byte(ncidPrefix+"_"+strconv.Itoa(j)+"_"+strconv.Itoa(i)), uint64(tagflag), segmentSize, keyset)
			if err != nil {
				return nil, 0, err
			}
			copy(taggroup[j], tag)
		}
		// 生成tag+tagP的taggroup格式
		err = encP.Encode(taggroup)
		if err != nil {
			return nil, 0, err
		}
		// 生成Field结构，此时beginOffset为下一个Field的起始偏移
		stripe = createFields(stripe, tmpdata, taggroup)
	}
	return stripe, beginOffset + realOffset, nil
}

// 将传入的Stripe(可丢失)恢复成完整的Stripe
// 例：传入5个block的块组Stripe，其中stripe[1] = nil，stripe[3] = nil，DataCount = 3，ParityCount = 2，
// 则返回恢复后的块组，并且该块组满足prefix加多个field的结构
func RecoverStripe(stripe [][]byte) ([][]byte, error) {
	if len(stripe) == 0 {
		return nil, ErrRepairCrash
	}
	// 在非空Stripe中识别prefix、块大小blockSize、块中所含的segment数量
	prefix, blockSize, offset, _, err := decodeStripe(stripe)
	if err != nil {
		return nil, err
	}
	DataCount := int(prefix.DataCount)
	ParityCount := int(prefix.ParityCount)
	tagNum := 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
	fieldSize := prefix.SegmentSize + prefix.TagSize*tagNum

	if len(stripe) != DataCount+ParityCount {
		fmt.Print("the len(stripe) is:", len(stripe))
		fmt.Println("the Datacount+ParityCount is:", DataCount+ParityCount)
		return nil, ErrRepairCrash
	}
	// 构建丢失的块并加上prefix
	for i := 0; i < DataCount+ParityCount; i++ {
		if len(stripe[i]) == 0 {
			stripe[i] = make([]byte, 0, blockSize)
			tmpPrefix, err := PrefixEncode(RsPolicy, uint64(DataCount), uint64(ParityCount), prefix.TagFlag, prefix.SegmentSize, prefix.TagSize)
			if err != nil {
				return nil, err
			}
			stripe[i] = append(stripe[i], tmpPrefix...)
		}
	}
	// 创建临时Data组用以恢复segment，临时tag组用以恢复tag和tagp
	tmpData := creatGroup(DataCount+ParityCount, prefix.SegmentSize)
	tmpTag := creatGroup((DataCount+ParityCount)*int(tagNum), prefix.TagSize)
	// 解析出data、tag、tagP
	for i := 0; i < offset; i++ {
		clearGroup(tmpData)
		clearGroup(tmpTag)
		for j := 0; j < len(stripe); j++ {
			if len(stripe[j]) != prefix.Len+int(fieldSize)*i {
				rawData, tag := prefix.decodeField(stripe[j], uint64(i), fieldSize)
				copy(tmpData[j], rawData)
				for k := 0; k < int(tagNum); k++ {
					copy(tmpTag[j+(DataCount+ParityCount)*k], tag[k])
				}
			} else {
				tmpData[j] = nil
				for k := 0; k < int(tagNum); k++ {
					tmpTag[j+(DataCount+ParityCount)*k] = nil
				}
			}
		}
		// 恢复data、tag、tagP
		tmpData, err = RecoverData(tmpData, DataCount, ParityCount, -1)
		if err != nil {
			return nil, err
		}
		tmpTag, err = RecoverData(tmpTag, DataCount+ParityCount, (int(tagNum)-1)*(DataCount+ParityCount), -1)
		if err != nil {
			return nil, err
		}
		// 将恢复的数据放回stripe内
		for j := 0; j < DataCount+ParityCount; j++ {
			if len(stripe[j]) == prefix.Len+int(fieldSize)*i {
				stripe[j] = append(stripe[j], tmpData[j]...)
				for k := 0; k < int(tagNum); k++ {
					stripe[j] = append(stripe[j], tmpTag[j+(DataCount+ParityCount)*k]...)
				}
			}
		}
	}
	return stripe, nil
}

// 根据传入的Stripe(可只含数据块)和偏移量，将偏移量内的每个Segment部分拼接返回(含有补足的0)
func GetFileDataFromSripe(stripe [][]byte, dataCount, beginOffset, endOffset int) ([]byte, error) {
	// 解析prefix
	prefix, _, err := PrefixDecode(stripe[0])
	if err != nil {
		return nil, err
	}
	tagNum := 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
	fieldSize := int(prefix.SegmentSize + prefix.TagSize*tagNum)
	// 根据offset构建空的文件数据
	var offset int
	realOffset := (len(stripe[0])-prefix.Len-1)/fieldSize + 1
	if endOffset == -1 {
		offset = realOffset - beginOffset
	} else {
		offset = endOffset - beginOffset + 1
	}
	data := make([]byte, 0, uint64(offset*dataCount)*prefix.SegmentSize)
	// 根据offset从每个块中提取Field的data
	for i := beginOffset; i < beginOffset+offset; i++ {
		for j := 0; j < dataCount; j++ {
			data = append(data, stripe[j][prefix.Len+i*fieldSize:prefix.Len+i*fieldSize+int(prefix.SegmentSize)]...)
		}
	}
	return data, nil
}

// 创建空的数据冗余块组
func creatGroup(count int, blockSize uint64) [][]byte {
	blockgroup := make([][]byte, count)
	for i := 0; i < len(blockgroup); i++ {
		blockgroup[i] = make([]byte, blockSize)
	}
	return blockgroup
}

// 创建含Prefix的Stripe
func createStripeHavePrefix(segmentSize, tagSize uint64, prefixByte []byte, dataCount, parityCount, segmentCount int) [][]byte {
	blockSize := (segmentSize+tagSize*uint64(2+(parityCount-1)/dataCount))*uint64(segmentCount) + uint64(len(prefixByte))
	blockGroup := make([][]byte, dataCount+parityCount)
	for i := 0; i < len(blockGroup); i++ {
		blockGroup[i] = make([]byte, 0, blockSize)
		blockGroup[i] = append(blockGroup[i], prefixByte...)
	}
	return blockGroup
}

func createStripeNoPrefix(segmentSize, tagSize uint64, dataCount, parityCount, segmentCount int) [][]byte {
	blockSize := (segmentSize + tagSize*uint64(2+(parityCount-1)/dataCount)) * uint64(segmentCount)
	blockGroup := make([][]byte, dataCount+parityCount)
	for i := 0; i < len(blockGroup); i++ {
		blockGroup[i] = make([]byte, 0, blockSize)
	}
	return blockGroup
}

// 使冗余块组内数据置0
func clearGroup(group [][]byte) {
	for i := 0; i < len(group); i++ {
		for j := 0; j < len(group[0]); j++ {
			group[i][j] = 0
		}
	}
}

// 在Stripe内创建Fields
func createFields(Stripe, data, taggroup [][]byte) [][]byte {
	for i := 0; i < len(Stripe); i++ {
		Stripe[i] = append(Stripe[i], data[i]...)
	}
	for i := 0; i < len(taggroup); {
		for j := 0; j < len(Stripe); j++ {
			Stripe[j] = append(Stripe[j], taggroup[i]...)
			i++
		}
	}
	return Stripe
}

// 解析一个Field中的Segment、tag、tagP
func (prefix *Prefix) decodeField(data []byte, offset, segmentLen uint64) ([]byte, [][]byte) {
	prefixLen := uint64(prefix.Len)
	segment := data[prefixLen+offset*segmentLen : prefixLen+prefix.SegmentSize+offset*segmentLen]

	tagNum := 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
	tag := make([][]byte, tagNum)
	for i := 0; i < int(tagNum); i++ {
		tag[i] = data[prefixLen+prefix.SegmentSize+prefix.TagSize*uint64(i)+offset*segmentLen : prefixLen+prefix.SegmentSize+prefix.TagSize*uint64(i+1)+offset*segmentLen]
	}
	return segment, tag
}

// 解析一个Stripe中单独一个BLock的Prefix、大小、最多拥有的Field数量和该Stripe实际含有的未丢失数据块的数量
func decodeStripe(blocks [][]byte) (*Prefix, uint64, int, int, error) {
	var prefix *Prefix
	var blockSize uint64
	var num int
	var avaNum = 0
	for i := 0; i < len(blocks); i++ {
		if len(blocks[i]) != 0 {
			if avaNum == 0 {
				pre, rawData, err := PrefixDecode(blocks[i])
				if err != nil {
					// 有可能这一块数据受损，但后面的是可以的
					continue
				}
				prefix = pre
				blockSize = uint64(len(blocks[i]))
				tagNum := 1 + (prefix.ParityCount-1)/prefix.DataCount + 1
				num = (len(rawData)-1)/int(prefix.SegmentSize+prefix.TagSize*tagNum) + 1
			}
			avaNum++
		}
	}
	if avaNum == 0 {
		return prefix, blockSize, num, avaNum, errors.New("no available block")
	}
	return prefix, blockSize, num, avaNum, nil
}
