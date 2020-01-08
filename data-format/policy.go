package dataformat

import (
	"errors"
	"log"
	"sort"
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
	DataCount  int // recover how many fields
	blockCount int
	tagCount   int
	tagSize    int
	segSize    int
	fieldSize  int
	prefixSize int
}

// NewDefaultDataCoder creates a new datacoder with default
func NewDefaultDataCoder(policy, dataCount, pairtyCount int, keyset *mcl.KeySet) *DataCoder {
	return NewDataCoder(policy, dataCount, pairtyCount, CurrentVersion, DefaultTagFlag, DefaultSegmentSize, DefaultSegmentCount, keyset)
}

// 构建一个dataformat配置
func NewDataCoder(policy, dataCount, parityCount, version, tagFlag, segmentSize, segCount int, keyset *mcl.KeySet) *DataCoder {
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
		Version:      int32(version),
		Policy:       int32(policy),
		DataCount:    int32(dataCount),
		ParityCount:  int32(parityCount),
		TagFlag:      int32(tagFlag),
		SegmentSize:  int32(segmentSize),
		SegmentCount: int32(segCount),
	}

	return NewDataCoderWithPrefix(pre, keyset)
}

// NewDataCoderWithPrefix creates a new datacoder with prefix
func NewDataCoderWithPrefix(p *pb.Prefix, k *mcl.KeySet) *DataCoder {
	d := &DataCoder{
		Prefix: p,
		BlsKey: k,
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

	s, ok := TagMap[int(d.Prefix.TagFlag)]
	if !ok {
		s = 48
	}

	d.tagSize = int(s)
	d.segSize = int(d.Prefix.SegmentSize)
	d.fieldSize = d.segSize + d.tagSize*d.tagCount
}

func (d *DataCoder) Encode(data []byte, ncidPrefix string, start int) ([][]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}

	dc := int(d.Prefix.DataCount)
	pc := int(d.Prefix.ParityCount)

	endSegment := 1 + (len(data)-1)/(d.segSize*dc)

	blockSize := d.fieldSize * endSegment

	var stripe [][]byte
	if start == 0 {
		preData, preLen, err := bf.PrefixEncode(d.Prefix)
		if err != nil {
			return nil, 0, err
		}

		log.Println(preData)

		stripe = make([][]byte, d.blockCount)
		for i := 0; i < d.blockCount; i++ {
			stripe[i] = make([]byte, 0, blockSize+preLen)
			stripe[i] = append(stripe[i], preData...)
		}
	} else {
		stripe = make([][]byte, d.blockCount)
		for i := 0; i < len(stripe); i++ {
			stripe[i] = make([]byte, 0, blockSize)
		}
	}

	// 生成临时块组保存data切分后的segment
	dataGroup := createGroup(d.blockCount, d.segSize)
	// 生成taggroup装一组的tag+tagP
	tagGroup := createGroup(d.blockCount*d.tagCount, d.tagSize)

	enc, err := reedsolomon.New(int(dc), int(pc))
	if err != nil {
		return nil, 0, err
	}

	encP, err := reedsolomon.New(d.blockCount, d.blockCount*(d.tagCount-1))
	if err != nil {
		return nil, 0, err
	}

	for i := start; i < start+endSegment && len(data) != 0; i++ {
		clearGroup(dataGroup)
		clearGroup(tagGroup)
		for j := 0; j < dc; j++ {
			// 填充数据
			if len(data) < d.segSize {
				copy(dataGroup[j], data)
				data = data[:0]
			} else {
				copy(dataGroup[j], data[:d.segSize])
				data = data[d.segSize:]
			}
		}

		switch d.Prefix.Policy {
		case MulPolicy:
			for j := dc; j < d.blockCount; j++ {
				res := copy(dataGroup[j], dataGroup[0])
				if res != d.segSize {
					log.Println("copied: ", res)
				}
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
		for j := 0; j < d.blockCount; j++ {
			// 生成tag并装进taggroup，index为peerid_bucketid_stripeid_blockid_offsetid
			res.Reset()
			res.WriteString(ncidPrefix)
			res.WriteString(metainfo.BLOCK_DELIMITER)
			res.WriteString(strconv.Itoa(j))
			res.WriteString(metainfo.BLOCK_DELIMITER)
			res.WriteString(strconv.Itoa(i))

			tag, err := d.GenTagForSegment([]byte(res.String()), dataGroup[j])
			if err != nil {
				log.Println("gentag err for: ", res.String())
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

		for j := 0; j < d.blockCount; j++ {
			stripe[j] = append(stripe[j], dataGroup[j]...)
		}

		for j := 0; j < d.blockCount*d.tagCount; {
			for k := 0; k < d.blockCount; k++ {
				stripe[k] = append(stripe[k], tagGroup[j]...)
				j++
			}
		}
	}
	// endoffset
	return stripe, start + endSegment - 1, nil
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
		} else {
			data = rawData
		}
	case MulPolicy:
		data = rawData
	default:
		return nil, ErrWrongPolicy
	}

	segStart := start
	segLength := 1 + (length-1)/(int(d.Prefix.DataCount)*d.segSize)

	if length == -1 {
		segLength = 1 + (len(data[0])-d.prefixSize-1)/d.fieldSize - segStart
	}

	res := make([]byte, 0, segLength*int(d.Prefix.DataCount)*d.segSize)
	// 根据offset从每个块中提取Field的data
	for i := segStart; i < segStart+segLength; i++ {
		for j := 0; j < int(d.Prefix.DataCount); j++ {
			res = append(res, data[j][d.prefixSize+i*d.fieldSize:d.prefixSize+i*d.fieldSize+d.segSize]...)
		}
	}

	return res, nil
}

func Repair(stripe [][]byte) ([][]byte, error) {
	prefix, _, err := decodeStripe(stripe)
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
	preData, preLen, err := bf.PrefixEncode(d.Prefix)
	if err != nil {
		return nil, err
	}

	avai := 0
	// 构建丢失的块并加上prefix
	for i := 0; i < d.blockCount; i++ {
		if len(stripe[i]) == 0 {
			stripe[i] = make([]byte, 0, d.prefixSize+d.DataCount*d.fieldSize)
			stripe[i] = append(stripe[i], preData...)
		} else {
			avai = i
		}
	}

	// 创建临时Data组用以恢复segment，临时tag组用以恢复tag和tagp
	tmpData := createGroup(d.blockCount, int(d.segSize))
	tmpTag := createGroup(d.blockCount*d.tagCount, d.tagSize)
	// 解析出data、tag、tagP
	for i := 0; i < d.DataCount; i++ {
		clearGroup(tmpData)
		clearGroup(tmpTag)
		for j := 0; j < d.blockCount; j++ {
			if len(stripe[j]) != preLen+d.fieldSize*i {
				rawdata, tags := d.decodeField(stripe[j][preLen+i*d.fieldSize:preLen+(i+1)*d.fieldSize], d.tagCount)
				copy(tmpData[j], rawdata)
				for k := 0; k < d.tagCount; k++ {
					copy(tmpTag[j+d.blockCount*k], tags[k])
				}
			}
		}

		// recover data
		switch d.Prefix.Policy {
		case MulPolicy:
			for j := 0; j < d.blockCount; j++ {
				if len(stripe[j]) == preLen+i*d.fieldSize {
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
		for j := 0; j < d.blockCount; j++ {
			if len(stripe[j]) == preLen+i*d.fieldSize {
				stripe[j] = append(stripe[j], tmpData[j]...)
				for k := 0; k < d.tagCount; k++ {
					stripe[j] = append(stripe[j], tmpTag[j+k*d.blockCount]...)
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
	rawdata := data[:d.segSize]
	tag := make([][]byte, 0, tagNum)
	for i := 0; i < tagNum; i++ {
		tag[i] = data[d.segSize+i*d.tagSize : d.segSize+(i+1)*d.tagSize]
	}
	return rawdata, tag
}

// 创建空的数据冗余块组
func createGroup(count, blockSize int) [][]byte {
	blockgroup := make([][]byte, count)
	for i := 0; i < len(blockgroup); i++ {
		blockgroup[i] = make([]byte, blockSize)
	}
	return blockgroup
}

// 使冗余块组内数据置0
func clearGroup(group [][]byte) {
	for i := 0; i < len(group); i++ {
		for j := 0; j < len(group[i]); j++ {
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
func decodeStripe(data [][]byte) (*pb.Prefix, int, error) {
	var prefix *pb.Prefix
	var avaNum int
	lengths := make([]int, len(data))
	for i := 0; i < len(data); i++ {
		lengths[i] = len(data[i])
		if len(data[i]) != 0 {
			if prefix == nil {
				pre, _, err := bf.PrefixDecode(data[i])
				if err != nil {
					continue
				}
				prefix = pre
			}
			avaNum++
		}
	}

	if avaNum == 0 {
		return nil, 0, errors.New("no available block")
	}

	if prefix != nil && (int(prefix.DataCount) > avaNum || int(prefix.DataCount) > len(lengths)) {
		log.Println("repair crash: need data:", prefix.DataCount, ", but got avaNum: ", avaNum)
		return nil, 0, ErrRepairCrash
	}

	sort.Sort(sort.Reverse(sort.IntSlice(lengths)))

	if lengths[prefix.DataCount] <= 0 {
		log.Println("repair crash: need data:", prefix.DataCount, ", but got avaNum again: ", avaNum)
		return nil, 0, ErrRepairCrash
	}

	return prefix, lengths[prefix.DataCount], nil
}
