package provider

import (
	"log"

	"github.com/memoio/go-mefs/source/instance"
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
	case metainfo.Contract:
		go p.handleUserDeployedContracts(km, metaKey, from)
	case metainfo.Challenge:
		go p.handleChallengeBls12(km, metaValue, from)
	case metainfo.Repair:
		go p.handleRepair(km, metaValue, from)
	case metainfo.Block:
		switch optype {
		case metainfo.Put:
			err := p.handlePutBlock(km, metaValue, from)
			if err != nil {
				log.Println("put Blcok Error: ", err)
				return nil, err
			}
		case metainfo.Get:
			res, err := p.handleGetBlock(km, from)
			if err != nil {
				log.Println("getBlcokError: ", err)
			} else {
				return res, nil
			}
		case metainfo.Append:
			err := p.handleAppendBlock(km, metaValue, from)
			if err != nil {
				log.Println("put Blcok Error: ", err)
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
