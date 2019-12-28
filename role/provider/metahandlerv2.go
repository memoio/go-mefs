package provider

import (
	"log"

	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandlerV2 provider角色回调接口的实现，
type HandlerV2 struct {
	Role string
}

// HandleMetaMessage provider角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (provider *HandlerV2) HandleMetaMessage(optype int, metaKey string, metaValue []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetDType()
	switch dtype {
	case metainfo.Contract:
		go handleUserDeployedContracts(km, metaKey, from)
	case metainfo.Challenge:
		go handleChallengeBls12(km, metaValue, from)
	case metainfo.Repair:
		go handleRepair(km, metaValue, from)
	case metainfo.Block:
		switch optype {
		case metainfo.Put:
			err := handlePutBlock(km, metaValue, from)
			if err != nil {
				log.Println("put Blcok Error: ", err)
				return nil, err
			}
		case metainfo.Get:
			res, err := handleGetBlock(km, from)
			if err != nil {
				log.Println("getBlcokError: ", err)
			} else {
				return res, nil
			}
		case metainfo.Append:
			err := handleAppendBlock(km, metaValue, from)
			if err != nil {
				log.Println("put Blcok Error: ", err)
				return nil, err
			}
		case metainfo.Delete:
			go handleDeleteBlock(km, from)
		}
	default: //没有匹配的信息，报错
		return nil, metainfo.ErrWrongType
	}
	return []byte(instance.MetaHandlerComplete), nil
}

// GetRole 获取这个节点的角色信息，返回错误说明provider还没有启动好
func (provider *HandlerV2) GetRole() (string, error) {
	return provider.Role, nil
}
