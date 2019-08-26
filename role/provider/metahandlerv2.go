package provider

import (
	"fmt"
	"strconv"

	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
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
		fmt.Println("keytype：UserInitReq 不处理")
	case metainfo.UserDeployedContracts:
		go handleUserDeployedContracts(km, metaKey, from)
	case metainfo.Challenge:
		go handleChallengeBls12(km, metaValue, from)
	case metainfo.Repair:
		go handleRepair(km, metaValue, from)
	case metainfo.StorageSync:
		go hanldeStorageSync(from)
	case metainfo.DeleteBlock:
		go handleDeleteBlock(km, from)
	case metainfo.GetBlock:
		res, err := handleGetBlock(km, from)
		if err == nil {
			return res, nil
		}
	case metainfo.PutBlock:
		handlePutBlock(km, metaValue, from)
	default: //没有匹配的信息，报错
		return "", metainfo.ErrWrongType
	}
	return metainfo.MetaHandlerComplete, nil
}

// 获取这个节点的角色信息，返回错误说明provider还没有启动好
func (provider *ProviderHandlerV2) GetRole() (string, error) {
	return provider.Role, nil
}

func hanldeStorageSync(kid string) error {
	cfg, err := localNode.Repo.Config()
	if err != nil {
		fmt.Println("get config failed :", err)
		return err
	}
	maxSpace := cfg.Datastore.StorageMax
	dataStore := localNode.Repo.Datastore()
	actulDataSpace, err := ds.DiskUsage(dataStore)
	if err != nil {
		fmt.Println("get disk usage failed :", err)
		return err
	}
	rawDataSpace := actulDataSpace
	km, err := metainfo.NewKeyMeta(kid, metainfo.StorageSync)
	if err != nil {
		fmt.Println("construct StorageSync KV error :", err)
		return err
	}
	value := maxSpace + metainfo.DELIMITER + strconv.FormatUint(actulDataSpace, 10) + metainfo.DELIMITER + strconv.FormatUint(rawDataSpace, 10)
	_, err = sendMetaRequest(km, value, kid)
	if err != nil {
		fmt.Println("send error :", err)
		return err
	}
	return nil
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
	fmt.Println("测试用回调函数")
	fmt.Println("km.mid:", km.GetMid())
	fmt.Println("km.options", km.GetOptions())
}
