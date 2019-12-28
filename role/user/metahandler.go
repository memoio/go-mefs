package user

import (
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage User角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (u *Info) HandleMetaMessage(dt int, metaKey string, metaValue []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	keytype := km.GetDType()

	switch keytype {
	case metainfo.UserInit: //handle init response from keeper
		if dt == metainfo.Put {
			uInfo := u.GetUser(km.GetOptions()[0])
			go uInfo.(*LfsInfo).gInfo.handleUserInit(km, metaValue, from)
		}
	default: //没有匹配的信息，报错
		return nil, metainfo.ErrWrongType
	}
	return []byte(instance.MetaHandlerComplete), nil
}

// GetRole gets role
func (u *Info) GetRole() (string, error) {
	return u.role, nil
}
