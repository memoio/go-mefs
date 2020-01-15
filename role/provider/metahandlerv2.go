package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage provider角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (p *Info) HandleMetaMessage(optype int, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetDType()
	switch dtype {
	case metainfo.UserStart:
		return p.handleUserStart(km, metaValue, from)
	case metainfo.Challenge:
		if optype == metainfo.Get {
			go p.handleChallengeBls12(km, metaValue, from)
		}
	case metainfo.Repair:
		if optype == metainfo.Get {
			go p.handleRepair(km, metaValue, from)
		}
	case metainfo.Block:
		switch optype {
		case metainfo.Put:
			err := p.handlePutBlock(km, metaValue, from)
			if err != nil {
				utils.MLogger.Error("put blcok error: ", err)
				return nil, err
			}
		case metainfo.Get:
			res, err := p.handleGetBlock(km, metaValue, sig, from)
			if err != nil {
				utils.MLogger.Error("get blcok error: ", err)
			} else {
				return res, nil
			}
		case metainfo.Append:
			err := p.handleAppendBlock(km, metaValue, from)
			if err != nil {
				utils.MLogger.Info("append blcok error: ", err)
				return nil, err
			}
		case metainfo.Delete:
			go p.handleDeleteBlock(km, from)
		}
	default: //没有匹配的信息，报错
		return nil, metainfo.ErrWrongType
	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (p *Info) handleUserStart(km *metainfo.KeyMeta, metaValue []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleUserStart: ", km.ToString(), " from: ", from)

	gid := km.GetMid()
	ops := km.GetOptions()
	if len(ops) != 3 {
		return nil, errors.New("wrong key")
	}

	splitValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitValue) != 2 {
		return nil, errors.New("wrong value")
	}

	var keepers []string
	var pros []string
	kids := splitValue[0]
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(keeper)
		if err != nil {
			continue
		}
		keepers = append(keepers, keeper)
	}

	pids := splitValue[1]
	has := false
	for i := 0; i < len(pids)/utils.IDLength; i++ {
		pid := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(pid)
		if err != nil {
			continue
		}
		if pid == p.localID {
			has = true
		}
		pros = append(pros, pid)
	}

	if !has {
		return nil, errors.New("Not my user")
	}

	uid := ops[0]

	_, ok := p.fsGroup.Load(gid)
	if !ok {
		gp := newGroup(p.localID, uid, gid, keepers, pros)
		p.fsGroup.Store(gid, gp)
	}

	ui, ok := p.users.Load(uid)
	if !ok {
		ui := &uInfo{
			userID: uid,
		}
		ui.setQuery(gid)
		p.users.Store(uid, ui)
	} else {
		ui.(*uInfo).setQuery(groupID)
	}

	kmkps, _ := metainfo.NewKeyMeta(gid, metainfo.LogFS, uid)
	p.ds.PutKey(context.Background(), kmkps.ToString(), metaValue, "local")

	p.loadChannelValue(uid, gid)

	return []byte("ok"), nil
}
