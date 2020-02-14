package instance

import (
	"errors"

	pb "github.com/memoio/go-mefs/proto"
)

var (
	// ErrMetaHandlerNotAssign 节点没有挂载接口时调用，报这个错
	ErrMetaHandlerNotAssign = errors.New("MetaMessageHandler not assign")
	//ErrMetaHandlerFailed 进行回调函数出错，没有特定错误的时候，报这个错
	ErrMetaHandlerFailed = errors.New("meta Handler err")
)

const (
	// MetaHandlerComplete returns
	MetaHandlerComplete = "complete"
)

// Service is
type Service interface {
	// type, key, value, from
	HandleMetaMessage(pb.OpType, string, []byte, []byte, string) ([]byte, error)
	GetRole() string
	Stop() error
}
