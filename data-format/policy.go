package dataformat

import (
	"errors"
	"log"
	"strconv"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/data-format/reedsolomon"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	pb "github.com/memoio/go-mefs/source/go-block-format/pb"
	"github.com/memoio/go-mefs/utils/metainfo"
)

type DataCoder struct {
	Prefix     *pb.Prefix
	BlsKey     *mcl.KeySet
	Repair     bool
	DataSize   int
	blockCount int
	tagCount   int
	tagSize    int
	segSize    int
	fieldSize  int
	prefixSize int
}

// NewDefaultDataCoder creates a new datacode with default
func NewDefaultDataCoder(policy, dataCount, pairtyCount int32, keyset *mcl.KeySet) *DataCoder {
	return NewDataCoder(policy, dataCount, pairtyCount, DefaultTagFlag, DefaultSegmentSize, DefaultLength, keyset)
}

// 构建一个dataformat配置
func NewDataCoder(policy, dataCount, parityCount, tagSize, segmentSize, length int32, keyset *mcl.KeySet) *DataCoder {
	if segmentSize < DefaultSegmentSize {
		segmentSize = DefaultSegmentSize
	}

	switch policy {
	case RsPolicy:
	case MulPolicy:
		parityCount = dataCount + parityCount - 1
		dataCount = 1
	default:
		return nil
	}

	pre := &pb.Prefix{
		Policy:      int32(policy),
		DataCount:   int32(dataCount),
		ParityCount: int32(parityCount),
		TagSize:     int32(tagSize),
		SegmentSize: int32(segmentSize),
		Length:      int32(length),
	}

	d := &DataCoder{
		Prefix: pre,
		BlsKey: keyset,
	}
	d.PreCompute()
	return d
}

func (d *DataCoder) PreCompute() {
	preLen := proto.Size(d.Prefix)
	d.prefixSize = preLen + proto.SizeVarint(uint64(preLen))
	dc := int(d.Prefix.DataCount)
	pc := int(d.Prefix.ParityCount)
	d.blockCount = dc + pc
	d.tagCount = 2 + (pc-1)/dc
	d.tagSize = int(d.Prefix.TagSize)
	d.segSize = int(d.Prefix.SegmentSize)
	d.fieldSize = d.segSize + d.tagSize*d.tagCount
}

func (d *DataCoder) Encode(data []byte, ncidPrefix string, start int) ([][]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}

	dc := int(d.Prefix.DataCount)
	pc := int(d.Prefix.ParityCount)
	bc := dc + pc
	tagNum := 1 + (bc-1)/dc

	segSize := int(d.Prefix.SegmentSize)
	endSegment := (len(data) - 1) / (segSize * dc)

	tagSize := int(d.Prefix.TagSize)
	fieldSize := segSize + tagSize*tagNum
	blockSize := fieldSize * (endSegment + 1)

	blockGroup := make([][]byte, bc)
	for i := 0; i < len(blockGroup); i++ {
		blockGroup[i] = make([]byte, 0, blockSize)
	}

	var stripe [][]byte
	if start == 0 {
		preData, err := bf.PrefixEncode(d.Prefix)
		if err != nil {
			return nil, 0, err
		}

		preLen := len(preData)

		stripe = creatGroup(bc, blockSize+preLen)
		for i := 0; i < len(stripe); i++ {
			stripe[i] = append(stripe[i], preData...)
		}
	} else {
		stripe = creatGroup(bc, blockSize)
	}

	// 生成临时块组保存data切分后的segment
	dataGroup := creatGroup(bc, segSize)
	// 生成taggroup装一组的tag+tagP
	tagGroup := creatGroup(bc*tagNum, tagSize)

	enc, err := reedsolomon.New(int(dc), int(pc))
	if err != nil {
		return nil, 0, err
	}

	encP, err := reedsolomon.New(int(bc), int(bc*(tagNum-1)))
	if err != nil {
		return nil, 0, err
	}

	for i := 0; i <= int(endSegment) && data != nil; i++ {
		clearGroup(dataGroup)
		clearGroup(tagGroup)
		for j := 0; j < int(dc); j++ {
			// 填充数据
			if len(data) < int(segSize) {
				copy(dataGroup[j], data)
				data = nil
				break
			}
			copy(dataGroup[j], data[:segSize])
			data = data[segSize:]
		}

		switch d.Prefix.Policy {
		case MulPolicy:
			for j := dc; j < bc; j++ {
				copy(dataGroup[j], dataGroup[0])
			}
		case RsPolicy:
			err = enc.Encode(dataGroup)
			if err != nil {
				return nil, 0, err
			}
		default:
			return nil, 0, ErrWrongPolicy
		}

		var res strings.Builder
		for j := 0; j < int(bc); j++ {
			// 生成tag并装进taggroup，index为peerid_bucketid_stripeid_blockid_offsetid
			res.Reset()
			res.WriteString(ncidPrefix)
			res.WriteString(metainfo.DELIMITER)
			res.WriteString(strconv.Itoa(j))
			res.WriteString(metainfo.DELIMITER)
			res.WriteString(strconv.Itoa(i))

			tag, err := d.GenTagForSegment([]byte(res.String()), dataGroup[j])
			if err != nil {
				return nil, 0, err
			}
			copy(tagGroup[j], tag)
		}
		// 生成tag+tagP的taggroup格式
		err = encP.Encode(tagGroup)
		if err != nil {
			return nil, 0, err
		}
		// 生成Field结构，此时beginOffset为下一个Field的起始偏移
		stripe = createFields(stripe, dataGroup, tagGroup)
	}
	return stripe, start + endSegment, nil
}

func (d *DataCoder) Decode(rawData [][]byte, start, length int) ([]byte, error) {
	var data [][]byte
	var err error
	switch d.Prefix.Policy {
	case RsPolicy:
		if d.Repair {
			data, err = d.recover(rawData)
			if err != nil {
				return nil, err
			}
		}
	case MulPolicy:
	default:
		return nil, ErrWrongPolicy
	}

	tagNum := 2 + (d.Prefix.ParityCount-1)/d.Prefix.DataCount
	fieldSize := int(d.Prefix.SegmentSize + d.Prefix.TagSize*tagNum)

	priData, err := bf.PrefixEncode(d.Prefix)
	if err != nil {
		return nil, err
	}

	priLen := len(priData)

	// 根据offset构建空的文件数据
	realOffset := len(data[0])/int(fieldSize) - 1
	if length == -1 {
		length = realOffset - start
	}

	if length <= 0 || start+length > realOffset {
		return nil, ErrDataTooShort
	}

	res := make([]byte, 0, length*int(d.Prefix.DataCount*d.Prefix.SegmentSize))
	// 根据offset从每个块中提取Field的data
	for i := start; i <= start+length; i++ {
		for j := 0; j < int(d.Prefix.DataCount); j++ {
			res = append(res, data[j][priLen+i*fieldSize:priLen+i*fieldSize+int(d.Prefix.SegmentSize)]...)
		}
	}
	return res, nil
}

func Repair(stripe [][]byte) ([][]byte, error) {
	prefix, err := decodeStripe(stripe)
	if err != nil {
		return nil, err
	}

	// 如果avaNum>=DataCount，但是stripe数量不够DataCount+ParityCount，可修复但得补够
	if len(stripe) < int(prefix.DataCount+prefix.ParityCount) {
		for i := len(stripe); i < int(prefix.DataCount+prefix.ParityCount); i++ {
			stripe = append(stripe, nil)
		}
	}

	switch prefix.Policy {
	case MulPolicy:
	case RsPolicy:
	default:
		return nil, ErrWrongPolicy
	}
	return nil, ErrRepairCrash
}

func (d *DataCoder) RecoverStripe(stripe [][]byte) ([][]byte, error) {
	dc := int(d.Prefix.DataCount)
	pc := int(d.Prefix.ParityCount)
	bc := int(dc + pc)
	tagNum := int(1 + (bc-1)/dc)
	fieldSize := int(d.Prefix.SegmentSize) + int(d.Prefix.TagSize)*tagNum

	preData, err := bf.PrefixEncode(d.Prefix)
	if err != nil {
		return nil, err
	}

	preLen := len(preData)
	avai := 0
	// 构建丢失的块并加上prefix
	for i := 0; i < bc; i++ {
		if len(stripe[i]) == 0 {
			stripe[i] = make([]byte, 0, preLen+int(d.Prefix.Length)*fieldSize)
			stripe[i] = append(stripe[i], preData...)
		} else {
			avai = i
		}
	}

	// 创建临时Data组用以恢复segment，临时tag组用以恢复tag和tagp
	tmpData := creatGroup(bc, int(d.Prefix.SegmentSize))
	tmpTag := creatGroup(bc*tagNum, int(d.Prefix.TagSize))
	// 解析出data、tag、tagP
	for i := 0; i < d.Size; i++ {
		clearGroup(tmpData)
		clearGroup(tmpTag)
		for j := 0; j < bc; j++ {
			if len(stripe[j]) != preLen+fieldSize*i {
				rawdata, tags := d.decodeField(stripe[j][preLen+i*fieldSize:preLen+(i+1)*fieldSize], tagNum)
				copy(tmpData[j], rawdata)
				for k := 0; k < tagNum; k++ {
					copy(tmpTag[j+bc*k], tags[k])
				}
			}
		}

		// recover data
		switch d.Prefix.Policy {
		case MulPolicy:
			for j := 0; j < bc; j++ {
				if len(stripe[j]) == preLen+fieldSize*i {
					copy(tmpData[j], tmpData[avai])
				}
			}
		case RsPolicy:
			tmpData, err = d.recover(tmpData)
			if err != nil {
				return nil, err
			}
		default:
			return nil, ErrWrongPolicy

		}

		// recover tag
		tmpTag, err = d.recover(tmpTag)
		if err != nil {
			return nil, err
		}

		// 将恢复的数据放回stripe内
		for j := 0; j < bc; j++ {
			if len(stripe[j]) == preLen+fieldSize*i {
				stripe[j] = append(stripe[j], tmpData[j]...)
				for k := 0; k < int(tagNum); k++ {
					stripe[j] = append(stripe[j], tmpTag[j+bc*k]...)
				}
			}
		}
	}
	return stripe, nil
}

// 将传入的数据冗余块组恢复，返回想要恢复的块，若index为-1则返回个块组
func (d *DataCoder) recover(data [][]byte) ([][]byte, error) {
	// 根据传入的参数，恢复整个块组
	enc, err := reedsolomon.New(int(d.Prefix.DataCount), int(d.Prefix.ParityCount))
	if err != nil {
		return nil, err
	}
	ok, err := enc.Verify(data)
	if err == reedsolomon.ErrShardNoData || err == reedsolomon.ErrTooFewShards {
		return nil, err
	}
	if !ok {
		err = enc.Reconstruct(data)
		if err != nil {
			return nil, err
		}
		ok, err = enc.Verify(data)
		if !ok {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (d *DataCoder) decodeField(data []byte, tagNum int) ([]byte, [][]byte) {
	rawdata := data[:d.Prefix.SegmentSize]
	tag := make([][]byte, 0, tagNum)
	for i := 0; i < tagNum; i++ {
		tag[i] = data[d.Prefix.SegmentSize+int32(i)*d.Prefix.TagSize : d.Prefix.SegmentSize+int32(i+1)*d.Prefix.TagSize]
	}
	return rawdata, tag
}

// 创建空的数据冗余块组
func creatGroup(count, blockSize int) [][]byte {
	blockgroup := make([][]byte, count)
	for i := 0; i < len(blockgroup); i++ {
		blockgroup[i] = make([]byte, 0, blockSize)
	}
	return blockgroup
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
func createFields(stripe, dataGroup, tagGroup [][]byte) [][]byte {
	for i := 0; i < len(stripe); i++ {
		stripe[i] = append(stripe[i], dataGroup[i]...)
	}
	for i := 0; i < len(tagGroup); {
		for j := 0; j < len(stripe); j++ {
			stripe[j] = append(stripe[j], tagGroup[i]...)
			i++
		}
	}
	return stripe
}

// 解析一个Stripe中单独一个BLock的Prefix、大小、最多拥有的Field数量和该Stripe实际含有的未D丢失数据块的数量
func decodeStripe(data [][]byte) (*pb.Prefix, error) {
	var prefix *pb.Prefix
	var avaNum int32
	for i := 0; i < len(data); i++ {
		if len(data[i]) != 0 {
			if prefix == nil {
				pre, err := bf.PrefixDecode(data[i])
				if err != nil {
					continue
				}
				prefix = pre
			}
			avaNum++
		}
	}

	if avaNum == 0 {
		return nil, errors.New("no available block")
	}

	if prefix != nil && prefix.DataCount > avaNum {
		log.Println("repair crash: need data:", prefix.DataCount, ", but got avaNum: ", avaNum)
		return nil, ErrRepairCrash
	}

	return prefix, nil
}
