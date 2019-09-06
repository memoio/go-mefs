//节点之间以KV对的形式交互信息的key的格式
// key：mainid/keytype/operator1/operator2 ...  分隔符用\t或其他不会重复的

package metainfo

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrWrongType      = errors.New("mismatch type")
	ErrIllegalKey     = errors.New("this key is illegal")
	ErrWrongKeyLength = errors.New("this key's length is wrong")
	ErrIllegalValue   = errors.New("this metavalue is illegal")
)

//DELIMITER 作为信息中的分隔符，不能与信息中的字符重复
const DELIMITER = "|"
const REPAIR_DELIMETER = "/"

// MetaKeyType 操作码的类型
type MetaKeyType int32

//这部分是操作码
const (
	Wrong MetaKeyType = iota
	UserInitReq
	UserInitRes
	UserInitNotif
	UserInitNotifRes
	UserDeployedContracts
	NewKPReq
	Sync
	Local
	TendermintRestart
	Challenge
	Proof
	Repair
	RepairRes
	DeleteBlock
	HasBlock
	StorageSync
	BlockMetaInfo
	Query
	PutBlock
	GetBlock
	GetPeerAddr
	PosAdd
	PosDelete
	PosMeta
	Test MetaKeyType = 99
)

// KeyTypeMap 记录每个keytype对应的key长度，若小于这个长度，则会报错
var KeyTypeMap = map[MetaKeyType]int{
	Wrong:                 0,
	UserInitReq:           5,
	UserInitRes:           4,
	UserInitNotif:         4,
	UserInitNotifRes:      3,
	UserDeployedContracts: 2,
	NewKPReq:              3,
	Sync:                  3,
	TendermintRestart:     2,
	Challenge:             3,
	Proof:                 4,
	Repair:                2,
	RepairRes:             2,
	DeleteBlock:           2,
	Local:                 2,
	HasBlock:              0,
	StorageSync:           2,
	BlockMetaInfo:         2,
	Query:                 3,
	PutBlock:              1,
	GetBlock:              1,
	GetPeerAddr:           0,
	PosAdd:                2,
	PosDelete:             2,
	PosMeta:               0,
	Test:                  0,
}

//同步操作中区分信息类别的参数，mid/"Local"or"Sync"/SyncType
const (
	SyncTypeUID          = "uid"
	SyncTypeKid          = "kid"
	SyncTypePid          = "pid"
	SyncTypeBlock        = "block"
	SyncTypeChalRes      = "chalres"
	SyncTypeChalPay      = "chalpay"
	SyncTypeLastPay      = "lastpay"
	SyncTypeTInfo        = "tendermintinfo"
	SyncTypeBft          = "bft"
	SyncTypeRole         = "roleinfo"
	SyncTypeCfg          = "config"
	SyncTypeLedger       = "ledgerinfo"
	SyncTypeCredit       = "credit"
	SyncTypeChannelValue = "channelvalue"
)

// 节点设置相关的信息  mid/"Local"/"SyncTypeCfg"/CfgType
const (
	CfgTypeBls12 = "bls12"
)

// Query中查询信息的类别 mid/"Query"/QueryType
const (
	QueryTypeLastChal = "lastchal"
)

// KeyMeta 解析key中信息 整理成的数据结构
type KeyMeta struct {
	mid     string      // main id = peerID or blockID 这条信息主要相关的单位，数据块or节点
	keytype MetaKeyType //操作码 标识这条信息用来做什么操作，或某种状态
	options []string    //顺序存放本信息中的操作码
}

//以下是keyMeta数据结构的操作，获取和修改其中的元素
func (km *KeyMeta) GetMid() string {
	if km == nil {
		return ""
	}
	return km.mid
}
func (km *KeyMeta) GetKeyType() MetaKeyType {
	if km == nil {
		return Wrong
	}
	return km.keytype
}
func (km *KeyMeta) GetOptions() []string {
	if km == nil {
		return nil
	}
	return km.options
}

// TODO:修改keytype时，要求的key长度可能会变化，这里需要做容错？
func (km *KeyMeta) SetKeyType(keyType MetaKeyType) {
	if km == nil {
		return
	}
	km.keytype = keyType
}

// ToByte 将KeyMeta结构体转换成byte，进行传输
func (km *KeyMeta) ToByte() []byte {
	return []byte(km.ToString())
}

// ToString 将KeyMeta结构体转换成字符串格式，进行传输
func (km *KeyMeta) ToString() string {
	if km == nil {
		return ""
	}
	if km.mid == "" {
		return ""
	}
	res := strings.Join([]string{km.mid, strconv.Itoa(int(km.keytype))}, DELIMITER)
	for _, option := range km.options {
		res += DELIMITER + option
	}
	return res
}

//NewKeyMeta 获取新的keymeta结构体
func NewKeyMeta(mainID string, keyType MetaKeyType, options ...string) (*KeyMeta, error) {
	minLength, ok := KeyTypeMap[keyType]
	if !ok || (len(options)+2) < minLength { //检查对应keytype的参数是否足够
		return nil, ErrWrongKeyLength
	}

	return &KeyMeta{
		mid:     mainID,
		keytype: keyType,
		options: options,
	}, nil
}

// GetKeyMeta 对于传入的key进行整理，返回结构体KeyMeta
func GetKeyMeta(key string) (*KeyMeta, error) {
	splitedKey := strings.Split(key, DELIMITER)
	if len(splitedKey) < 2 {
		return nil, ErrIllegalKey
	}
	keyTypeInt, err := strconv.Atoi(splitedKey[1])
	if err != nil {
		return nil, ErrWrongType
	}
	keyType := MetaKeyType(keyTypeInt)

	minLength, ok := KeyTypeMap[keyType]
	if !ok || len(splitedKey) < minLength { //没有记录这一类的key,或者缺少参数 报错
		return nil, ErrWrongKeyLength
	}

	km := &KeyMeta{
		mid:     splitedKey[0],
		keytype: keyType,
	}
	for i := 2; i < len(splitedKey); i++ { //从第2号元素开始，添加这个信息的操作数
		km.options = append(km.options, splitedKey[i])
	}
	return km, nil
}
