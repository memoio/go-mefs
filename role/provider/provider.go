package provider

import (
	"fmt"
	"sync"
	"time"

	"github.com/memoio/go-mefs/core"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var localNode *core.MefsNode

var usersConfigs sync.Map

var ProContracts *ProviderContracts

//StartProviderService start provider service
func StartProviderService(node *core.MefsNode, capacity int64, duration int64, price int64, reDeployOffer bool) (err error) {
	localNode = node
	ProContracts = &ProviderContracts{}
	if cfg, _ := node.Repo.Config(); !cfg.Test {
		//部署resolver和offer
		for {
			err = providerDeployResolverAndOffer(node, capacity, duration, price, reDeployOffer)
			if err != nil {
				fmt.Println("provider deploying resolver and offer failed!")
				time.Sleep(2 * time.Minute)
			} else {
				break
			}
		}
	}

	fmt.Println("Provider Service is ready")
	return nil
}

func PersistBeforeExit() error {
	if localNode == nil || ProContracts == nil {
		return ErrProviderServiceNotReady
	}
	channels := GetChannels()
	for _, channel := range channels {
		// 保存本地形式：K-userID，V-channel此时的value
		km, err := metainfo.NewKeyMeta(channel.ChannelAddr, metainfo.Local, metainfo.SyncTypeChannelValue)
		if err != nil {
			return err
		}
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(km.ToString(), channel.Value.String(), "local")
		if err != nil {
			return err
		}
		fmt.Println("持久化channel:", channel.ChannelAddr, channel.Value.String())
	}
	posKM, err := metainfo.NewKeyMeta(posUid, metainfo.Local)
	if err != nil {
		return err
	}
	posValue := opt.CidPrefix
	fmt.Println("posKM :", posKM.ToString(), "\nposValue :", posValue)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(posKM.ToString(), posValue, "local")
	if err != nil {
		fmt.Println("CmdPutTo posKM error :", err)
		return err
	}
	return nil
}
