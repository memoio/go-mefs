package provider

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/routing"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils/metainfo"
)

type Info struct {
	netID      string
	sk         string
	state      bool
	ds         data.Service
	conManager *pContracts
	blsConfigs sync.Map
}

//New start provider service
func New(ctx context.Context, id, sk string, ds data.Service, rt routing.Routing, capacity, duration, price int64, reDeployOffer bool) (instance.Service, error) {
	m := &Info{
		netID: id,
		sk:    sk,
		ds:    ds,
	}
	err := rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
		return nil, err
	}

	m.conManager = &pContracts{}
	for {
		err = providerDeployResolverAndOffer(id, sk, capacity, duration, price, reDeployOffer)
		if err != nil {
			log.Println("provider deploying resolver and offer failed!")
			time.Sleep(2 * time.Minute)
		} else {
			break
		}
	}

	err = m.saveProInfo()
	if err != nil {
		log.Println("Save ", m.netID, "'s provider info err", err)
		return nil, err
	}

	log.Println("Save ", m.netID, "'s provider info success")

	err = m.saveOffer()
	if err != nil {
		log.Println("Save ", m.netID, "'s Offer err", err)
		return nil, err
	}
	log.Println("Save ", m.netID, "'s Offer success")

	go m.getKpMapRegular(ctx)
	go m.sendStorageRegular(ctx)

	m.state = true

	log.Println("Provider Service is ready")
	return m, nil
}

func (p *Info) Online() bool {
	return p.state
}

func (p *Info) GetRole() string {
	return metainfo.RoleProvider
}

func (p *Info) Stop() error {
	return p.save(context.Background())
}

func (p *Info) save(ctx context.Context) error {
	if !p.state {
		return errProviderServiceNotReady
	}
	channels := p.getChannels()
	for _, channel := range channels {
		// 保存本地形式：K-userID，V-channel此时的value
		km, err := metainfo.NewKeyMeta(channel.ChannelAddr, metainfo.Channel)
		if err != nil {
			return err
		}
		err = p.ds.PutKey(context.Background(), km.ToString(), []byte(channel.Value.String()), "local")
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
	err = p.ds.PutKey(context.Background(), posKM.ToString(), []byte(posValue), "local")
	if err != nil {
		log.Println("CmdPutTo posKM error :", err)
		return err
	}
	return nil
}

func (p *Info) getKpMapRegular(ctx context.Context) {
	log.Println("Get kpMap from chain start!")
	peerID := p.netID
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

func (p *Info) sendStorageRegular(ctx context.Context) {
	log.Println("Send storages to keepers start!")
	time.Sleep(time.Minute)
	p.storageSync(ctx)
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				p.storageSync(ctx)
			}()
		}
	}
}

func (p *Info) storageSync(ctx context.Context) error {
	actulDataSpace, err := p.getDiskUsage()
	if err != nil {
		return err
	}

	maxSpace := p.getDiskTotal()

	klist, ok := contracts.GetKeepersOfPro(p.netID)
	if !ok {
		return nil
	}

	km, err := metainfo.NewKeyMeta(p.netID, metainfo.Storage)
	if err != nil {
		log.Println("construct StorageSync KV error :", err)
		return err
	}

	value := strconv.FormatUint(maxSpace, 10) + metainfo.DELIMITER + strconv.FormatUint(actulDataSpace, 10)

	for _, kid := range klist {
		_, err = p.ds.SendMetaRequest(ctx, int32(metainfo.Put), km.ToString(), []byte(value), nil, kid)
		if err != nil {
			log.Println("storage info send to", kid, "error: ", err)
		}
	}

	return nil
}
