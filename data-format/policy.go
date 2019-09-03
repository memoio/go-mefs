package dataformat

import (
	"fmt"

	mcl "github.com/memoio/go-mefs/bls12"
)

type DataEncoder struct {
	Policy      int32 // flase为多副本，true为RScode
	DataCount   int32 // 若为多副本，则副本总数等于DataCount+ParityCount
	ParityCount int32 // 若为多副本，
	TagFlag     int32
	SegmentSize uint64
	KeySet      *mcl.KeySet
}

// 构建一个dataformat配置
func NewDataEncoder(policy int32, dataCount, pairtyCount, tagflag int32, segmentSize uint64, keyset *mcl.KeySet) *DataEncoder {
	if segmentSize < DefaultSegmentSize {
		segmentSize = DefaultSegmentSize
	}

	newOpt := &DataEncoder{
		Policy:      policy,
		DataCount:   dataCount,
		ParityCount: pairtyCount,
		TagFlag:     tagflag,
		SegmentSize: segmentSize,
		KeySet:      keyset,
	}
	return newOpt
}

// Encode的策略选择
// 传入数据，返回编码后的stripe或者多副本
func (opt *DataEncoder) Encode(data []byte, ncidPrefix string, beginOffset int32) ([][]byte, int, error) {
	var stripe [][]byte
	var offset int32
	var err error
	switch opt.Policy {
	case RsPolicy:
		stripe, offset, err = opt.rsEncode(data, ncidPrefix, beginOffset)
		if err != nil {
			return nil, 0, err
		}
	case MulPolicy:
		stripe, offset, err = opt.mulEncode(data, ncidPrefix, beginOffset)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, ErrWrongPolicy
	}
	return stripe, int(offset), nil
}

func (opt *DataEncoder) mulEncode(data []byte, ncidPrefix string, beginOffset int32) ([][]byte, int32, error) {
	if beginOffset == 0 {
		stripe, offset, err := DataEncodeToMul(data, ncidPrefix, opt.DataCount, opt.ParityCount, opt.SegmentSize, uint64(opt.TagFlag), opt.KeySet)
		if err != nil {
			return nil, 0, err
		}

		return stripe, int32(offset), nil
	}
	stripe, offset, err := DataEncodeToMulForAppend(data, ncidPrefix, opt.DataCount, opt.ParityCount, opt.SegmentSize, uint64(opt.TagFlag), int(beginOffset), opt.KeySet)
	if err != nil {
		return nil, 0, err
	}
	return stripe, int32(offset), nil
}

func (opt *DataEncoder) rsEncode(data []byte, ncidPrefix string, beginOffset int32) ([][]byte, int32, error) {
	if beginOffset == 0 {
		stripe, offset, err := EncodeDataToPreStripe(data, ncidPrefix, int(opt.DataCount), int(opt.ParityCount), int(opt.TagFlag), opt.SegmentSize, opt.KeySet)
		if err != nil {
			return nil, 0, err
		}
		return stripe, int32(offset), nil
	}
	stripe, offset, err := EncodeDataToNoPreStripe(data, ncidPrefix, int(opt.DataCount), int(opt.ParityCount), int(opt.TagFlag), int(beginOffset), opt.SegmentSize, opt.KeySet)
	if err != nil {
		return nil, 0, err
	}
	return stripe, int32(offset), nil
}

//DataDecoder 用于用一个stripe中获取纯数据，无tag
type DataDecoder struct {
	Policy      int32 // flase为多副本，true为RScode
	DataCount   int32 // 若为多副本，则副本总数等于DataCount+ParityCount
	ParityCount int32 // 若为多副本，
}

func NewDataDecoder(policy, dataCount, pairtyCount int32) (*DataDecoder, error) {
	return &DataDecoder{
		Policy:      policy,
		DataCount:   dataCount,
		ParityCount: pairtyCount,
	}, nil
}

func (dec *DataDecoder) Decode(datas [][]byte, offsetStart int, needRepair bool) ([]byte, error) {
	var data []byte
	var err error
	switch dec.Policy {
	case RsPolicy:
		if needRepair {
			recoveredData, err := RecoverData(datas, int(dec.DataCount), int(dec.Policy), -1)
			if err != nil {
				return nil, err
			}
			data, err = GetFileDataFromSripe(recoveredData, int(dec.DataCount), int(offsetStart), -1)
			if err != nil {
				return nil, err
			}
		} else {
			data, err = GetFileDataFromSripe(datas, int(dec.DataCount), int(offsetStart), -1)
			if err != nil {
				return nil, err
			}
		}
	case MulPolicy:
		if datas == nil || len(datas) < 1 {
			return nil, ErrDataTooShort
		}
		data, err = GetSegsFromData(datas[0], int(offsetStart), -1)
	default:
		return nil, ErrWrongPolicy
	}
	return data, nil
}

// 修复的策略选择
// 传入欲修复的数据组，返回修复后的数据组
// 注意：①传入的是多副本数据，则原副本应该放在正确的stripe位置上，例如:stripe为5副本的策略，以stripe中的第2个副本data2进行修复，则stripe[1] = data2，stripe其他数量为nil，返回修复好的stripe
// ②传入的是rs纠删码，则stripe只需要保留足够的块以便恢复，例如：stripe为3+2的冗余策略，stripe[0]~stripe[2]含有数据，其他为nil，则返回修复好的数据
func Repair(stripe [][]byte) ([][]byte, error) {
	prefix, _, _, avaNum, err := decodeStripe(stripe)
	if err != nil {
		return nil, err
	}
	// 如果stripe中实际含有的DataCount数量小于前缀要求的DataCount数量，修复不了，报错
	if int(prefix.DataCount) > avaNum {
		fmt.Println("prefix.DataCount :", prefix.DataCount, "\navaNum :", avaNum)
		return nil, ErrRepairCrash
	}
	// 如果avaNum>=DataCount，但是stripe数量不够DataCount+ParityCount，可修复但得补够
	if len(stripe) < int(prefix.DataCount+prefix.ParityCount) {
		for i := len(stripe); i < int(prefix.DataCount+prefix.ParityCount); i++ {
			stripe = append(stripe, nil)
		}
	}
	switch prefix.Policy {
	case MulPolicy:
		for i := 0; i < len(stripe); i++ {
			if len(stripe[i]) != 0 {
				stripe, err := mulRepair(stripe[i], i)
				if err != nil {
					return nil, err
				}
				return stripe, nil
			}
		}
	case RsPolicy:
		stripe, err := rsRepair(stripe)
		if err != nil {
			return nil, err
		}
		return stripe, nil
	default:
		return nil, ErrWrongPolicy
	}
	return nil, ErrRepairCrash
}

// TODO:修改成多ptag修复模式
func mulRepair(data []byte, blockOrd int) ([][]byte, error) {
	stripe, err := RecoverMul(data, blockOrd)
	if err != nil {
		return nil, err
	}
	return stripe, nil
}

func rsRepair(stripe [][]byte) ([][]byte, error) {
	stripe, err := RecoverStripe(stripe)
	if err != nil {
		return nil, err
	}
	return stripe, nil
}
