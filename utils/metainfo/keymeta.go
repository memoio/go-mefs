//节点之间以KV对的形式交互信息的key的格式
// key：mainid/keytype/operator1/operator2 ...  分隔符用\t或其他不会重复的

package metainfo

import (
	"errors"
	"strconv"
	"strings"

	mpb "github.com/memoio/go-mefs/pb"
)

var (
	ErrWrongType      = errors.New("mismatch type")
	ErrIllegalKey     = errors.New("this key is illegal")
	ErrWrongKeyLength = errors.New("this key's length is wrong")
	ErrIllegalValue   = errors.New("this metavalue is illegal")
)

const (
	// DELIMITER seps message
	DELIMITER = "/"
	// BlockDelimiter sep block
	BlockDelimiter = "_"
)

const (
	RoleKeeper   = "keeper"
	RoleUser     = "user"
	RoleProvider = "provider"
)

// Key is mid/keyType/ops
type Key struct {
	mpb.KeyMeta
}

//NewKey creates a new key
func NewKey(mainID string, dt mpb.KeyType, ops ...string) (*Key, error) {
	km := &Key{
		KeyMeta: mpb.KeyMeta{
			MainID:  mainID,
			KeyType: dt,
		},
	}

	for i := 0; i < len(ops); i++ {
		km.Options = append(km.Options, ops[i])
	}

	return km, nil
}

// NewKeyFromString convert string to key
func NewKeyFromString(key string) (*Key, error) {
	splitedKey := strings.Split(key, DELIMITER)
	if len(splitedKey) < 2 {
		return nil, ErrIllegalKey
	}

	dt, err := strconv.Atoi(splitedKey[1])
	if err != nil {
		return nil, ErrWrongType
	}

	km := &Key{
		KeyMeta: mpb.KeyMeta{
			MainID:  splitedKey[0],
			KeyType: mpb.KeyType(dt),
		},
	}

	for i := 2; i < len(splitedKey); i++ {
		km.Options = append(km.Options, splitedKey[i])
	}
	return km, nil
}

// ToString converts key to string: mid/keyType/op1/op2/...
func (k *Key) ToString() string {
	var res strings.Builder

	res.WriteString(k.GetMainID())
	res.WriteString(DELIMITER)
	res.WriteString(strconv.Itoa(int(k.GetKeyType())))

	ops := k.GetOptions()
	for i := 0; i < len(ops); i++ {
		res.WriteString(DELIMITER)
		res.WriteString(ops[i])
	}

	return res.String()
}
