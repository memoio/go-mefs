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
const DELIMITER = "/"
const BLOCK_DELIMITER = "_"
const REPAIR_DELIMETER = "|"

const version = 100

//这部分是操作码
const (
	Wrong int = iota
	Store
	Put
	Get
	Append
	Delete
	Test = 99
)

const (
	Role int = iota
	Config
	PublicKey
	Keepers
	Providers
	Users
	Managedkeepers
	ManagedProviders
	ManagedUsers
	UserInit
	UserNotify
	Contract
	Query
	Offer
	UpKeeping
	Channel
	Block    // provider handle block data
	BlockPos // keeper handle block pos
	ExteralAddress
	Challenge
	Repair
	Storage
	Pos
)

// KeyMeta is datatype/mid/ops
type KeyMeta struct {
	version int
	mid     string   // main id = peerID or blockID
	dType   int      // indicates which data type
	options []string // extra options
}

func (km *KeyMeta) GetMid() string {
	if km == nil {
		return ""
	}
	return km.mid
}

func (km *KeyMeta) GetDType() int {
	if km == nil {
		return Wrong
	}
	return km.dType
}

func (km *KeyMeta) GetOptions() []string {
	if km == nil {
		return nil
	}
	return km.options
}

// TODO:修改keytype时，要求的key长度可能会变化，这里需要做容错？
func (km *KeyMeta) SetDType(keyType int) {
	if km == nil {
		return
	}
	km.dType = keyType
}

// ToByte 将KeyMeta结构体转换成byte，进行传输
func (km *KeyMeta) ToByte() []byte {
	return []byte(km.ToString())
}

// ToString 将KeyMeta结构体转换成字符串格式，进行传输
// datatype/mid/op1/op2/...
func (km *KeyMeta) ToString() string {
	var res strings.Builder

	res.WriteString(strconv.Itoa(km.dType))
	res.WriteString(DELIMITER)
	res.WriteString(km.mid)

	for _, option := range km.options {
		res.WriteString(DELIMITER)
		res.WriteString(option)
	}
	return res.String()
}

//NewKeyMeta 获取新的keymeta结构体
func NewKeyMeta(mainID string, dt int, ops ...string) (*KeyMeta, error) {
	return &KeyMeta{
		mid:     mainID,
		dType:   dt,
		options: ops,
	}, nil
}

// GetKeyMeta 对于传入的key进行整理，返回结构体KeyMeta
func GetKeyMeta(key string) (*KeyMeta, error) {
	splitedKey := strings.Split(key, DELIMITER)
	if len(splitedKey) < 2 {
		return nil, ErrIllegalKey
	}

	dt, err := strconv.Atoi(splitedKey[0])
	if err != nil {
		return nil, ErrWrongType
	}

	km := &KeyMeta{
		mid:   splitedKey[1],
		dType: dt,
	}

	for i := 2; i < len(splitedKey); i++ { //从第2号元素开始，添加这个信息的操作数
		km.options = append(km.options, splitedKey[i])
	}
	return km, nil
}
