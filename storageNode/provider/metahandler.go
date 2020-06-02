package provider

import (
	"time"

	"github.com/google/uuid"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage provider角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (p *Info) HandleMetaMessage(opType mpb.OpType, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
	if !p.Online() {
		return nil, role.ErrServiceNotReady
	}

	km, err := metainfo.NewKeyFromString(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetKeyType()
	switch dtype {
	case mpb.KeyType_UserStart:
		return p.handleUserStart(km, metaValue, sig, from)
	case mpb.KeyType_UserStop:
		return p.handleUserStop(km, metaValue, from)
	case mpb.KeyType_Challenge:
		if opType == mpb.OpType_Get {
			go p.handleChallengeBls12(km, metaValue, from)
		}
	case mpb.KeyType_Repair:
		if opType == mpb.OpType_Get {
			go p.handleRepair(km, metaValue, from)
		}
	case mpb.KeyType_Block:
		switch opType {
		case mpb.OpType_Put:
			err := p.handlePutBlock(km, metaValue, from)
			if err != nil {
				utils.MLogger.Error("put block error: ", err)
				return nil, err
			}
		case mpb.OpType_Get:
			return p.handleGetBlock(km, metaValue, sig, from)
		case mpb.OpType_Append:
			err := p.handleAppendBlock(km, metaValue, from)
			if err != nil {
				utils.MLogger.Info("append block error: ", err)
				return nil, err
			}
		case mpb.OpType_Delete:
			go p.handleDeleteBlock(km, from)
		}
	case mpb.KeyType_UserInit:
		return nil, metainfo.ErrWrongType
	default: //没有匹配的信息，报错
		switch opType {
		case mpb.OpType_Put:
			go p.handlePutKey(km, metaValue, sig, from)
		case mpb.OpType_Get:
			return p.handleGetKey(km, metaValue, sig, from)
		case mpb.OpType_Delete:
			go p.handleDeleteKey(km, metaValue, sig, from)
		default:
			return nil, metainfo.ErrWrongType
		}

	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (p *Info) handlePutKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	ctx := p.context
	ok := p.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	p.ds.PutKey(ctx, km.ToString(), metaValue, sig, "local")
}

func (p *Info) handleGetKey(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	ctx := p.context

	return p.ds.GetKey(ctx, km.ToString(), "local")
}

func (p *Info) handleDeleteKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	ctx := p.context
	ok := p.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	p.ds.DeleteKey(ctx, km.ToString(), "local")
}

func (p *Info) handleUserStart(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleUserStart: ", km.ToString(), " from: ", from)

	gid := km.GetMainID()
	ops := km.GetOptions()
	if len(ops) != 5 {
		return nil, role.ErrWrongKey
	}

	uid := ops[0]
	_, ok := p.fsGroup.Load(gid)
	if !ok {
		gp := p.newGroupWithFS(uid, gid, string(metaValue))
		if gp == nil {
			return nil, role.ErrNotMyUser
		}
	}

	ui := p.getUserInfo(uid)
	if ui != nil {
		ui.setQuery(gid)
	}

	kmkps, _ := metainfo.NewKey(gid, mpb.KeyType_LFS, uid)
	p.ds.PutKey(p.context, kmkps.ToString(), metaValue, nil, "local")

	gp := p.getGroupInfo(uid, gid, false)
	if gp != nil {
		if ops[4] == "0" && gp.sessionID != uuid.Nil && time.Now().Unix()-gp.sessionTime < role.SessionExpTime {
			return []byte(gp.sessionID.String()), nil
		}
		ok := p.ds.VerifyKey(p.context, km.ToString(), metaValue, sig)
		if !ok {
			utils.MLogger.Info("key signature is wrong for: ", km.ToString())
			return []byte(uuid.Nil.String()), nil
		}
		sessID, err := uuid.Parse(ops[3])
		if err != nil {
			return nil, err
		}
		gp.sessionID = sessID
		gp.sessionTime = time.Now().Unix()

		kmsess, err := metainfo.NewKey(gp.groupID, mpb.KeyType_Session, gp.userID)
		if err != nil {
			return nil, err
		}

		p.ds.PutKey(p.context, kmsess.ToString(), []byte(sessID.String()), nil, "local")

		return []byte(sessID.String()), nil
	}

	return nil, role.ErrNotMyUser
}

func (p *Info) handleUserStop(km *metainfo.Key, metaValue []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleUserStop: ", km.ToString(), " from: ", from)

	ops := km.GetOptions()
	if len(ops) != 4 {
		return nil, role.ErrWrongKey
	}
	gid := km.GetMainID()

	uid := ops[0]
	gp := p.getGroupInfo(uid, gid, false)
	if gp != nil {
		sessID, err := uuid.Parse(ops[3])
		if err != nil {
			return nil, err
		}

		if gp.sessionID == sessID {
			gp.sessionID = uuid.Nil
			gp.sessionTime = time.Now().Unix()
		}

		return []byte("ok"), nil
	}

	return nil, role.ErrNotMyUser
}

func (p *Info) handleHeartBeat(km *metainfo.Key, metaValue []byte, from string) {
	utils.MLogger.Info("handleUserStop: ", km.ToString(), " from:", from)

	ops := km.GetOptions()
	if len(ops) != 4 {
		return
	}

	uid := ops[0]
	qid := km.GetMainID()

	gp := p.getGroupInfo(uid, qid, false)
	if gp != nil {
		sessID, err := uuid.Parse(ops[3])
		if err != nil {
			return
		}

		if gp.sessionID == uuid.Nil {
			gp.sessionID = sessID
		}

		if gp.sessionID == sessID {
			gp.sessionTime = time.Now().Unix()
		}
		return
	}

	return
}
