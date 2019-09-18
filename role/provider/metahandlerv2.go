package provider

import (
	"log"

	cid "github.com/memoio/go-mefs/source/go-cid"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	RepairFailed  = "Repair Failed"
	RepairSuccess = "Repair Successes"
)

// ProviderHandlerV2 provider角色回调接口的实现，
type ProviderHandlerV2 struct {
	Role string
}

// HandleMetaMessage provider角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (provider *ProviderHandlerV2) HandleMetaMessage(metaKey, metaValue, from string) (string, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return "", err
	}
	keytype := km.GetKeyType()
	switch keytype {
	case metainfo.Test:
		go handleTest(km)
	case metainfo.UserInitReq:
		log.Println("keytype：UserInitReq, do not handle it")
	case metainfo.UserDeployedContracts:
		go handleUserDeployedContracts(km, metaKey, from)
	case metainfo.Challenge:
		go handleChallengeBls12(km, metaValue, from)
	case metainfo.Repair:
		go handleRepair(km, metaValue, from)
	case metainfo.DeleteBlock:
		go handleDeleteBlock(km, from)
	case metainfo.GetBlock:
		res, err := handleGetBlock(km, from)
		if err != nil {
			log.Println("getBlcokError: ", err)
		} else {
			return res, nil
		}
	case metainfo.PutBlock:
		go handlePutBlock(km, metaValue, from)
	default: //没有匹配的信息，报错
		return "", metainfo.ErrWrongType
	}
	return metainfo.MetaHandlerComplete, nil
}

// GetRole 获取这个节点的角色信息，返回错误说明provider还没有启动好
func (provider *ProviderHandlerV2) GetRole() (string, error) {
	return provider.Role, nil
}

func handleDeleteBlock(km *metainfo.KeyMeta, from string) error {
	blockID := km.GetMid()
	bcid := cid.NewCidV2([]byte(blockID))
	err := localNode.Blocks.DeleteBlock(bcid)
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	return nil
}

func handleTest(km *metainfo.KeyMeta) {
	log.Println("测试用回调函数")
	log.Println("km.mid:", km.GetMid())
	log.Println("km.options", km.GetOptions())
}
