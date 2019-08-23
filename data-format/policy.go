package dataformat

import (
	"fmt"

	mcl "github.com/memoio/go-mefs/bls12"
)

type DataformatOption struct {
	Policy      int32 // flase为多副本，true为RScode
	CidPrefix   string
	DataCount   int32 // 若为多副本，则副本总数等于DataCount+ParityCount
	ParityCount int32 // 若为多副本，
	TagFlag     int32
	BeginOffset int32 // 若StripeHasPrefix为true，则BeginOffset为0
	SegmentSize uint64
	KeySet      *mcl.KeySet
}

// 构建一个dataformat配置
func NewDataformat(policy int32, cidPrefix string, dataCount, pairtyCount, tagflag, beginOffset int32, segmentSize uint64, keyset *mcl.KeySet) DataformatOption {
	if segmentSize < DefaultSegmentSize {
		segmentSize = DefaultSegmentSize
	}
	if tagflag == CRC32 {
		cidPrefix = ""
	}
	newOpt := DataformatOption{
		Policy:      policy,
		CidPrefix:   cidPrefix,
		DataCount:   dataCount,
		ParityCount: pairtyCount,
		TagFlag:     tagflag,
		BeginOffset: beginOffset,
		SegmentSize: segmentSize,
		KeySet:      keyset,
	}
	return newOpt
}

// Encode的策略选择
// 传入数据，返回编码后的stripe或者多副本
func (opt *DataformatOption) Encode(data []byte) ([][]byte, int, error) {
	var stripe [][]byte
	var offset int32
	var err error
	switch opt.Policy {
	case RsPolicy:
		stripe, offset, err = opt.rsEncode(data)
		if err != nil {
			return nil, 0, err
		}
	case MulPolicy:
		stripe, offset, err = opt.mulEncode(data)
		if err != nil {
			return nil, 0, err
		}
	default:
		return nil, 0, ErrWrongPolicy
	}
	return stripe, int(offset), nil
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

func (opt *DataformatOption) mulEncode(data []byte) ([][]byte, int32, error) {
	if opt.BeginOffset == 0 {
		stripe, offset, err := DataEncodeToMul(data, opt.CidPrefix, opt.DataCount, opt.ParityCount, opt.SegmentSize, uint64(opt.TagFlag), opt.KeySet)
		if err != nil {
			return nil, 0, err
		}

		return stripe, int32(offset), nil
	}
	stripe, offset, err := DataEncodeToMulForAppend(data, opt.CidPrefix, opt.DataCount, opt.ParityCount, opt.SegmentSize, uint64(opt.TagFlag), int(opt.BeginOffset), opt.KeySet)
	if err != nil {
		return nil, 0, err
	}
	return stripe, int32(offset), nil
}

func (opt *DataformatOption) rsEncode(data []byte) ([][]byte, int32, error) {
	if opt.BeginOffset == 0 {
		stripe, offset, err := EncodeDataToPreStripe(data, opt.CidPrefix, int(opt.DataCount), int(opt.ParityCount), int(opt.TagFlag), opt.SegmentSize, opt.KeySet)
		if err != nil {
			return nil, 0, err
		}
		return stripe, int32(offset), nil
	}
	stripe, offset, err := EncodeDataToNoPreStripe(data, opt.CidPrefix, int(opt.DataCount), int(opt.ParityCount), int(opt.TagFlag), int(opt.BeginOffset), opt.SegmentSize, opt.KeySet)
	if err != nil {
		return nil, 0, err
	}
	return stripe, int32(offset), nil
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
