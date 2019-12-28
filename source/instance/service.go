package instance

import "errors"

var (
	ErrMetaHandlerNotAssign = errors.New("MetaMessageHandler not assign") // ErrMetaHandlerNotAssign 节点没有挂载接口时调用，报这个错
	ErrMetaHandlerFailed    = errors.New("meta Handler err")              //ErrMetaHandlerNotAssign 进行回调函数出错，没有特定错误的时候，报这个错
)

const (
	MetaHandlerComplete = "complete"
)

const (
	RoleKeeper   = "keeper"
	RoleUser     = "user"
	RoleProvider = "provider"
)

type Service interface {
	HandleMetaMessage(int, string, []byte, string) ([]byte, error) //传入Key Value 和发送信息的节点id
	GetRole() (string, error)                                      //获取本节点的角色信息
}
