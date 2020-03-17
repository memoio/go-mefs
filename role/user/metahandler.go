package user

import (
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage User角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (u *Info) HandleMetaMessage(opType mpb.OpType, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
	km, err := metainfo.NewKeyFromString(metaKey)
	if err != nil {
		return nil, err
	}

	keytype := km.GetKType()
	switch keytype {
	case mpb.KeyType_UserInit: //handle init response from keeper
		switch opType {
		case mpb.OpType_Put:
			fs, ok := u.fsMap.Load(km.GetMid())
			if !ok {
				utils.MLogger.Warn("no lfs for: ", km.GetMid())
			}
			go fs.(*LfsInfo).gInfo.handleUserInit(u.context, km, metaValue, from)
		default:
			return nil, metainfo.ErrWrongType
		}
	case mpb.KeyType_UserStart:
	case mpb.KeyType_UserNotify:
	default: //没有匹配的信息，报错
		switch opType {
		case mpb.OpType_Put:
			go u.handlePutKey(km, metaValue, sig, from)
		case mpb.OpType_Get:
			return u.handleGetKey(km, metaValue, sig, from)
		case mpb.OpType_Delete:
			go u.handleDeleteKey(km, metaValue, sig, from)
		default:
			return nil, metainfo.ErrWrongType
		}
	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (u *Info) handlePutKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	ctx := u.context
	ok := u.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	u.ds.PutKey(ctx, km.ToString(), metaValue, sig, "local")
}

func (u *Info) handleGetKey(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	return u.ds.GetKey(u.context, km.ToString(), "local")
}

func (u *Info) handleDeleteKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	ctx := u.context
	ok := u.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	u.ds.DeleteKey(ctx, km.ToString(), "local")
}
