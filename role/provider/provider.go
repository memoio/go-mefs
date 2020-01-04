package provider

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/routing"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// Info tracks provider's information
type Info struct {
	localID      string
	sk           string
	state        bool
	ds           data.Service
	storageUsed  uint64
	storageTotal uint64
	users        sync.Map // key: queryID, value: *groupInfo
	offers       []*role.OfferItem
	proContract  *role.ProviderItem
}

type groupInfo struct {
	sessionID    uuid.UUID
	userID       string
	groupID      string
	storageUsed  uint64
	storageTotal uint64
	keepers      []string
	blsPubKey    *mcl.PublicKey
	upkeeping    *role.UpKeepingItem
	channel      *role.ChannelItem
	query        *role.QueryItem
}

//New start provider service
func New(ctx context.Context, id, sk string, ds data.Service, rt routing.Routing, capacity, duration, price int64, reDeployOffer bool) (instance.Service, error) {
	m := &Info{
		localID: id,
		sk:      sk,
		ds:      ds,
		offers:  make([]*role.OfferItem, 1),
	}
	err := rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			_, err := role.DeployOffer(id, sk, capacity, duration, price, reDeployOffer)
			if err != nil {
				log.Println("provider deploying resolver and offer failed!")
				time.Sleep(2 * time.Minute)
			} else {
				break
			}
		}
	}()

	err = m.getContracts()
	if err != nil {
		log.Println("Save ", m.localID, "'s provider info err", err)
		return nil, err
	}

	log.Println("Get ", m.localID, "'s contract info success")

	go m.getKpMapRegular(ctx)
	go m.sendStorageRegular(ctx)
	go m.saveRegular(ctx)

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

func newGroup(localID, uid, gid string, kps []string) *groupInfo {
	g := &groupInfo{
		userID:  uid,
		groupID: gid,
		keepers: kps,
	}

	if gid != uid {
		g.getContracts(localID)
	}

	return g
}

func (p *Info) getGroupInfo(userID, groupID string, mode bool) *groupInfo {
	groupI, ok := p.users.Load(groupID)
	if !ok {
		if mode {
			return newGroup(p.localID, userID, groupID, []string{userID})
		}
		return nil
	}

	return groupI.(*groupInfo)
}

type quKey struct {
	uid string
	qid string
}

func (p *Info) getGroups() []quKey {
	var res []quKey
	p.users.Range(func(key, value interface{}) bool {
		tmp := quKey{
			uid: value.(*groupInfo).userID,
			qid: key.(string),
		}
		res = append(res, tmp)
		return true
	})

	return res
}

func (p *Info) saveRegular(ctx context.Context) {
	time.Sleep(time.Minute)
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.save(ctx)
		}
	}
}

func (p *Info) save(ctx context.Context) error {
	if !p.state {
		return errProviderServiceNotReady
	}

	res := p.getGroups()
	for _, qu := range res {
		p.saveChannelValue(qu.qid, qu.uid, p.localID)
	}
	posKM, err := metainfo.NewKeyMeta(groupID, metainfo.PosMeta)
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
	peerID := p.localID
	role.SaveKpMap(peerID)
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				role.SaveKpMap(peerID)
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

	klist, ok := role.GetKeepersOfPro(p.localID)
	if !ok {
		return nil
	}

	km, err := metainfo.NewKeyMeta(p.localID, metainfo.Storage)
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
