package dataformat

import (
	"errors"
	"sort"
	"strconv"
	"strings"

	proto "github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/pdp"
	"github.com/memoio/go-mefs/data-format/reedsolomon"
	mpb "github.com/memoio/go-mefs/pb"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

type DataCoder struct {
	Prefix     *mpb.BlockOptions
	BlsKey     pdp.KeySet
	Repair     bool
	RLength    int // recover how long
	blockCount int
	tagCount   int
	tagSize    int
	segSize    int
	fieldSize  int
	prefixSize int
}

// NewDataCoder 构建一个dataformat配置
func NewDataCoder(keyset pdp.KeySet, policy, dataCount, parityCount, version, tagFlag, segmentSize, segCount, encrpto int, userID, fsID string) (*DataCoder, error) {
	if segmentSize < DefaultSegmentSize {
		segmentSize = DefaultSegmentSize
	}

	switch policy {
	case RsPolicy:
	case MulPolicy:
		parityCount = dataCount + parityCount - 1
		dataCount = 1
	default:
		return nil, ErrWrongPolicy
	}

	bo := &mpb.BucketOptions{
		Version:      int32(version),
		Policy:       int32(policy),
		DataCount:    int32(dataCount),
		ParityCount:  int32(parityCount),
		TagFlag:      int32(tagFlag),
		SegmentSize:  int32(segmentSize),
		SegmentCount: int32(segCount),
		Encryption:   int32(encrpto),
	}

	return NewDataCoderWithBopts(keyset, bo, userID, fsID)
}

// NewDataCoderWithDefault creates a new datacoder with default
func NewDataCoderWithDefault(keyset pdp.KeySet, policy, dataCount, pairtyCount int, userID, fsID string) (*DataCoder, error) {
	return NewDataCoder(keyset, policy, dataCount, pairtyCount, CurrentVersion, DefaultTagFlag, DefaultSegmentSize, DefaultSegmentCount, DefaultCrypt, userID, fsID)
}

// NewDataCoderWithBopts contructs a new datacoder with bucketops
func NewDataCoderWithBopts(keyset pdp.KeySet, bo *mpb.BucketOptions, userID, fsID string) (*DataCoder, error) {
	pre := &mpb.BlockOptions{
		Bopts:   bo,
		Start:   0,
		UserID:  userID,
		QueryID: fsID,
	}

	return NewDataCoderWithPrefix(keyset, pre)
}

// NewDataCoderWithPrefix creates a new datacoder with prefix
func NewDataCoderWithPrefix(keyset pdp.KeySet, p *mpb.BlockOptions) (*DataCoder, error) {
	d := &DataCoder{
		Prefix: p,
		BlsKey: keyset,
	}
	err := d.PreCompute()
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (d *DataCoder) PreCompute() error {
	preLen := proto.Size(d.Prefix)
	d.prefixSize = preLen + proto.SizeVarint(uint64(preLen))
	dc := int(d.Prefix.Bopts.DataCount)
	pc := int(d.Prefix.Bopts.ParityCount)
	if dc < 1 || pc < 1 {
		return ErrWrongPolicy
	}

	d.blockCount = dc + pc
	d.tagCount = 2 + (pc-1)/dc

	s, ok := pdp.TagMap[int(d.Prefix.Bopts.TagFlag)]
	if !ok {
		s = 48
	}

	d.tagSize = int(s)
	d.segSize = int(d.Prefix.Bopts.SegmentSize)
	d.fieldSize = d.segSize + d.tagSize*d.tagCount

	return nil
}

func (d *DataCoder) Encode(data []byte, ncidPrefix string, start int) ([][]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, ErrDataTooShort
	}

	dc := int(d.Prefix.Bopts.DataCount)
	pc := int(d.Prefix.Bopts.ParityCount)

	endSegment := 1 + (len(data)-1)/(d.segSize*dc)

	blockSize := d.fieldSize * endSegment //加上tag之后的size

	d.Prefix.Start = int32(start) //curoffset
	preData, preLen, err := bf.PrefixEncode(d.Prefix)
	if err != nil {
		return nil, 0, err
	}

	stripe := make([][]byte, d.blockCount)
	for i := 0; i < d.blockCount; i++ {
		stripe[i] = make([]byte, blockSize+preLen)
		copy(stripe[i], preData)
	}

	// 生成临时块组保存data切分后的segment
	dataGroup := make([][]byte, d.blockCount)
	// 生成taggroup装一组的tag+tagP
	tagGroup := make([][]byte, d.blockCount*d.tagCount)

	enc, err := reedsolomon.New(int(dc), int(pc))
	if err != nil {
		return nil, 0, err
	}

	encP, err := reedsolomon.New(d.blockCount, d.blockCount*(d.tagCount-1))
	if err != nil {
		return nil, 0, err
	}

	for i := start; i < start+endSegment && len(data) != 0; i++ {
		beginOffset := preLen + (i-start)*d.fieldSize
		for j := 0; j < dc; j++ {
			dataGroup[j] = stripe[j][beginOffset : beginOffset+d.segSize]
			// 填充数据
			if len(data) < d.segSize {
				copy(dataGroup[j], data)
				data = data[:0]
			} else {
				copy(dataGroup[j], data[:d.segSize])
				data = data[d.segSize:]
			}
		}

		switch d.Prefix.Bopts.Policy {
		case MulPolicy:
			for j := dc; j < d.blockCount; j++ {
				dataGroup[j] = stripe[j][beginOffset : beginOffset+d.segSize]
				res := copy(dataGroup[j], dataGroup[0])
				if res != d.segSize {
					utils.MLogger.Error("copied: ", res, " is less than: ", d.segSize)
				}
			}
		case RsPolicy:
			for j := dc; j < d.blockCount; j++ {
				dataGroup[j] = stripe[j][beginOffset : beginOffset+d.segSize]
			}
			err = enc.Encode(dataGroup)
			if err != nil {
				return nil, 0, err
			}
		default:
			return nil, 0, ErrWrongPolicy
		}

		var res strings.Builder
		for j := 0; j < d.blockCount; j++ {
			tagGroup[j] = stripe[j][beginOffset+d.segSize : beginOffset+d.segSize+d.tagSize]
			// 生成tag并装进taggroup，index为peerid_bucketid_stripeid_blockid_offsetid
			res.Reset()
			res.WriteString(ncidPrefix)
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(j))
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(i))

			tag, err := d.GenTagForSegment([]byte(res.String()), dataGroup[j])
			if err != nil {
				utils.MLogger.Error("Gen tag for: ", res.String(), " fails: ", err)
				return nil, 0, err
			}
			copy(tagGroup[j], tag)
		}

		for j := 1; j < d.tagCount; j++ {
			for k := 0; k < d.blockCount; k++ {
				tagGroup[j*d.blockCount+k] = stripe[k][beginOffset+d.segSize+j*d.tagSize : beginOffset+d.segSize+(j+1)*d.tagSize]
			}
		}
		// 生成tag+tagP的taggroup格式
		err = encP.Encode(tagGroup)
		if err != nil {
			return nil, 0, err
		}
		// 生成Field结构，此时beginOffset为下一个Field的起始偏移
	}
	// endoffset
	return stripe, start + endSegment, nil
}

func (d *DataCoder) Decode(stripe [][]byte, start, length int) ([]byte, error) {
	var data [][]byte

	_, _, minLen, err := decodeStripe(stripe)
	if err != nil {
		return nil, err
	}

	if d.Repair {
		d.RLength = minLen
		data, _, err = d.recover(stripe)
		if err != nil {
			return nil, err
		}
	} else {
		data = stripe
	}

	segStart := start
	segLength := 1 + (length-1)/(int(d.Prefix.Bopts.DataCount)*d.segSize)

	if length == -1 {
		segLength = 1 + (minLen-d.prefixSize-1)/d.fieldSize - segStart
	}

	res := make([]byte, 0, segLength*int(d.Prefix.Bopts.DataCount)*d.segSize)
	// 根据offset从每个块中提取Field的data
	for i := segStart; i < segStart+segLength; i++ {
		for j := 0; j < int(d.Prefix.Bopts.DataCount); j++ {
			res = append(res, data[j][d.prefixSize+i*d.fieldSize:d.prefixSize+i*d.fieldSize+d.segSize]...)
		}
	}

	return res, nil
}

// Repair stripes
func Repair(stripe [][]byte) ([][]byte, int, error) {
	prefix, _, minLen, err := decodeStripe(stripe)
	if err != nil {
		return nil, 0, err
	}

	coder, err := NewDataCoderWithPrefix(nil, prefix)
	if err != nil {
		return nil, 0, err
	}

	coder.RLength = minLen

	return coder.recover(stripe)
}

func (d *DataCoder) recover(stripe [][]byte) ([][]byte, int, error) {
	if len(stripe) < d.blockCount {
		for i := len(stripe); i < d.blockCount; i++ {
			stripe = append(stripe, nil)
		}
	}

	preData, preLen, err := bf.PrefixEncode(d.Prefix)
	if err != nil {
		return nil, 0, err
	}

	fieldStripe := make([][]byte, d.blockCount)
	fieldSize := d.fieldSize
	fieldCount := 1 + (d.RLength-preLen-1)/fieldSize

	for i := 0; i < d.blockCount; i++ {
		if stripe[i] == nil {
			stripe[i] = make([]byte, 0, preLen)
			stripe[i] = append(stripe[i], preData...)
		}
	}

	for i := 0; i < fieldCount; i++ {
		for j := 0; j < d.blockCount; j++ {
			if len(stripe[j]) >= preLen+(i+1)*fieldSize {
				fieldStripe[j] = stripe[j][preLen+i*fieldSize : preLen+(i+1)*fieldSize]
			} else {
				fieldStripe[j] = nil
			}

		}
		fieldStripe, err = d.recoverField(fieldStripe)
		if err != nil {
			return nil, i, err
		}
		for j := 0; j < d.blockCount; j++ {
			if len(stripe[j]) == preLen+i*fieldSize {
				stripe[j] = append(stripe[j], fieldStripe[j]...)
			}
		}
	}

	return stripe, fieldCount, nil
}

func (d *DataCoder) recoverField(stripe [][]byte) ([][]byte, error) {
	tmpData := make([][]byte, d.blockCount)
	// 解析出data、tag、tagP
	for i := 0; i < d.blockCount; i++ {
		if stripe[i] != nil {
			tmpData[i] = stripe[i][:d.segSize]
		} else {
			tmpData[i] = nil
		}
	}

	datas, err := d.recoverData(tmpData, int(d.Prefix.Bopts.DataCount), int(d.Prefix.Bopts.ParityCount))
	if err != nil {
		return nil, err
	}

	tmpTag := make([][]byte, d.blockCount*d.tagCount)
	// 解析出data、tag、tagP
	for j := 0; j < d.tagCount; j++ {
		for i := 0; i < d.blockCount; i++ {
			if stripe[i] != nil {
				tmpTag[i+j*d.blockCount] = stripe[i][d.segSize+j*d.tagSize : d.segSize+(j+1)*d.tagSize]
			} else {
				tmpTag[i+j*d.blockCount] = nil
			}
		}
	}

	tags, err := d.recoverData(tmpTag, d.blockCount, d.blockCount*(d.tagCount-1))
	if err != nil {
		return nil, err
	}

	for j := 0; j < d.blockCount*d.tagCount; {
		for k := 0; k < d.blockCount; k++ {
			datas[k] = append(datas[k], tags[j]...)
			j++
		}
	}

	return datas, nil
}

// 将传入的数据冗余块组恢复，返回想要恢复的块，若index为-1则返回个块组
func (d *DataCoder) recoverData(data [][]byte, dc, pc int) ([][]byte, error) {
	switch {
	case dc > 1:
		enc, err := reedsolomon.New(dc, pc)
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
			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, ErrDataBroken
			}
		}
		return data, nil
	case dc == 1:
		var i int
		for i = 0; i < d.blockCount; i++ {
			if data[i] != nil {
				break
			}
		}

		if i == d.blockCount {
			return nil, errors.New("no available")
		}

		for j := 0; j < d.blockCount; j++ {
			if data[j] == nil {
				data[j] = make([]byte, 0)
				data[j] = append(data[j], data[i]...)
			}
		}
		return data, nil
	default:
		return nil, ErrWrongPolicy
	}
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

// decode stripe returns prefix, min len
func decodeStripe(data [][]byte) (*mpb.BlockOptions, int, int, error) {
	var prefix *mpb.BlockOptions
	var avaNum, preLen int
	lengths := make([]int, len(data))
	for i := 0; i < len(data); i++ {
		lengths[i] = len(data[i])
		if len(data[i]) != 0 {
			if prefix == nil || preLen == 0 {
				pre, pLen, err := bf.PrefixDecode(data[i])
				if err != nil {
					continue
				}
				preLen = pLen
				prefix = pre
			}
			avaNum++
		}
	}

	if avaNum == 0 {
		return nil, 0, 0, errors.New("no available block")
	}

	if prefix != nil && (int(prefix.Bopts.DataCount) > avaNum || int(prefix.Bopts.DataCount) > len(lengths)) {
		utils.MLogger.Error("repair crash, need data count: ", prefix.Bopts.DataCount, ", but got avaNum: ", avaNum)
		return nil, 0, 0, ErrRepairCrash
	}

	sort.Sort(sort.Reverse(sort.IntSlice(lengths)))

	if lengths[prefix.Bopts.DataCount-1] <= 0 {
		utils.MLogger.Error("repair crash after sort: need count: ", prefix.Bopts.DataCount, ", but got avaNum again: ", avaNum)
		return nil, 0, 0, ErrRepairCrash
	}

	return prefix, preLen, lengths[prefix.Bopts.DataCount-1], nil
}
