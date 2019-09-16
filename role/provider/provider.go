package provider

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

var localNode *core.MefsNode

var usersConfigs sync.Map

var ProContracts *ProviderContracts

//StartProviderService start provider service
func StartProviderService(ctx context.Context, node *core.MefsNode, capacity int64, duration int64, price int64, reDeployOffer bool) (err error) {
	localNode = node
	ProContracts = &ProviderContracts{}
	if cfg, _ := node.Repo.Config(); !cfg.Test {
		//部署resolver和offer
		for {
			err = providerDeployResolverAndOffer(node, capacity, duration, price, reDeployOffer)
			if err != nil {
				log.Println("provider deploying resolver and offer failed!")
				time.Sleep(2 * time.Minute)
			} else {
				break
			}
		}

		err = saveOffer()
		if err != nil {
			log.Println("Save ", localNode.Identity.Pretty(), "'s Offer err", err)
		} else {
			log.Println("Save ", localNode.Identity.Pretty(), "'s Offer success")
		}
	}

	go getKpMapRegular(ctx)
	go sendStorageRegular(ctx)

	log.Println("Provider Service is ready")
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
		log.Println("持久化channel:", channel.ChannelAddr, channel.Value.String())
	}
	posKM, err := metainfo.NewKeyMeta(posID, metainfo.PosMeta)
	if err != nil {
		return err
	}
	posValue := posCidPrefix
	log.Println("posKM :", posKM.ToString(), "\nposValue :", posValue)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(posKM.ToString(), posValue, "local")
	if err != nil {
		log.Println("CmdPutTo posKM error :", err)
		return err
	}
	return nil
}

func storageSync(ctx context.Context) error {
	actulDataSpace, err := getDiskUsage()
	if err != nil {
		return err
	}

	maxSpace := getDiskTotal()

	klist, ok := contracts.GetKeepersOfPro(localNode.Identity.Pretty())
	if !ok {
		return nil
	}

	km, err := metainfo.NewKeyMeta(localNode.Identity.Pretty(), metainfo.StorageSync)
	if err != nil {
		log.Println("construct StorageSync KV error :", err)
		return err
	}

	value := strconv.FormatUint(maxSpace, 10) + metainfo.DELIMITER + strconv.FormatUint(actulDataSpace, 10)

	for _, kid := range klist {
		_, err = sendMetaRequest(km, value, kid)
		if err != nil {
			log.Println("storage info send to", kid, "error: ", err)
			if sc.ConnectTo(ctx, localNode, kid) {
				sendMetaRequest(km, value, kid)
			}
		}
	}

	return nil
}

func getKpMapRegular(ctx context.Context) {
	log.Println("Get kpMap from chain start!")
	peerID := localNode.Identity.Pretty()
	contracts.SaveKpMap(peerID)
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				contracts.SaveKpMap(peerID)
			}()
		}
	}
}

func sendStorageRegular(ctx context.Context) {
	log.Println("Send storages to keepers start!")
	time.Sleep(time.Minute)
	storageSync(ctx)
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				storageSync(ctx)
			}()
		}
	}
}
