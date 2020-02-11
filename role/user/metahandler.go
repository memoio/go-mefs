package user

import (
	"context"

	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage User角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (u *Info) HandleMetaMessage(opType int, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	keytype := km.GetDType()

	switch keytype {
	case metainfo.UserInit: //handle init response from keeper
		switch opType {
		case metainfo.Put:
			fs, ok := u.fsMap.Load(km.GetMid())
			if !ok {
				utils.MLogger.Warn("no lfs for: ", km.GetMid())
			}
			go fs.(*LfsInfo).gInfo.handleUserInit(km, metaValue, from)
		default:
			return nil, metainfo.ErrWrongType
		}
	case metainfo.UserStart:
	case metainfo.UserNotify:
	default: //没有匹配的信息，报错
		switch opType {
		case metainfo.Put:
			go u.handlePutKey(km, metaValue, sig, from)
		case metainfo.Get:
			return u.handleGetKey(km, metaValue, sig, from)
		case metainfo.Delete:
			go u.handleDeleteKey(km, metaValue, sig, from)
		default:
			return nil, metainfo.ErrWrongType
		}
	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (u *Info) handlePutKey(km *metainfo.KeyMeta, metaValue, sig []byte, from string) {
	ctx := context.Background()
	ok := u.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	u.ds.PutKey(ctx, km.ToString(), metaValue, sig, "local")
}

func (u *Info) handleGetKey(km *metainfo.KeyMeta, metaValue, sig []byte, from string) ([]byte, error) {
	ctx := context.Background()

	return u.ds.GetKey(ctx, km.ToString(), "local")
}

func (u *Info) handleDeleteKey(km *metainfo.KeyMeta, metaValue, sig []byte, from string) {
	ctx := context.Background()
	ok := u.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	u.ds.DeleteKey(ctx, km.ToString(), "local")
}
