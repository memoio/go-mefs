package provider

import (
	"errors"
	"log"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage provider角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (p *Info) HandleMetaMessage(optype int, metaKey string, metaValue []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetDType()
	switch dtype {
	case metainfo.UserStart:
		log.Println("handle user start: ", metaKey)
		go p.handleUserStart(km, metaKey, from)
	case metainfo.Challenge:
		log.Println("handle challenge: ", metaKey)
		go p.handleChallengeBls12(km, metaValue, from)
	case metainfo.Repair:
		log.Println("handle repair: ", metaKey)
		go p.handleRepair(km, metaValue, from)
	case metainfo.Block:
		switch optype {
		case metainfo.Put:
			log.Println("handle put block: ", metaKey)
			err := p.handlePutBlock(km, metaValue, from)
			if err != nil {
				log.Println("put Blcok Error: ", err)
				return nil, err
			}
		case metainfo.Get:
			log.Println("handle get block: ", metaKey)
			res, err := p.handleGetBlock(km, from)
			if err != nil {
				log.Println("getBlcokError: ", err)
			} else {
				return res, nil
			}
		case metainfo.Append:
			log.Println("handle append block: ", metaKey)
			err := p.handleAppendBlock(km, metaValue, from)
			if err != nil {
				log.Println("put Blcok Error: ", err)
				return nil, err
			}
		case metainfo.Delete:
			log.Println("handle delete block: ", metaKey)
			go p.handleDeleteBlock(km, from)
		}
	default: //没有匹配的信息，报错
		return nil, metainfo.ErrWrongType
	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (p *Info) handleUserStart(km *metainfo.KeyMeta, metaValue, from string) ([]byte, error) {
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
	kids := splitValue[0]
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(keeper)
		if err != nil {
			continue
		}
		keepers = append(keepers, keeper)
	}

	uid := ops[0]
	gp := newGroup(p.localID, uid, gid, keepers)

	p.users.Store(gid, gp)

	p.loadChannelValue(uid, gid)

	return []byte("ok"), nil
}
