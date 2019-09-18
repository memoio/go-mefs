package metainfo

import "errors"

var (
	ErrMetaHandlerNotAssign = errors.New("MetaMessageHandler not assign") // ErrMetaHandlerNotAssign 节点没有挂载接口时调用，报这个错
	ErrMetaHandlerFailed    = errors.New("meta Handler err")              // ErrMetaHandlerNotAssign 进行回调函数出错，没有特定错误的时候，报这个错
)

const (
	MetaHandlerComplete = "complete"
	MetaPutBlockErr     = "PutBlockErr"
)

const (
	RoleKeeper   = "keeper"
	RoleUser     = "user"
	RoleProvider = "provider"
)

// MetaMessageHandlerV2 接口，用于进行节点交互信息的回调操作，节点启动时，根据角色启动不同的接口实例，用Routing.Assignmetahandler挂接
type MetaMessageHandlerV2 interface {
	HandleMetaMessage(string, string, string) (string, error) //传入Key Value 和发送信息的节点id
	GetRole() (string, error)                                 //获取本节点的角色信息
}
