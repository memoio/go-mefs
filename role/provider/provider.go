package provider

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
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
	fsGroup      sync.Map // key: queryID, value: *groupInfo
	users        sync.Map // key: userID, value: *uInfo
	keepers      sync.Map // key: keeperID, value: *kInfo
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
	providers    []string
	blsKey       *mcl.KeySet
	upkeeping    *role.UpKeepingItem
	channel      *role.ChannelItem
	query        *role.QueryItem
}

// store user information
type uInfo struct {
	sync.RWMutex
	userID string
	querys map[string]struct{} // key is queryID
}

func (u *uInfo) setQuery(qid string) {
	u.Lock()
	defer u.Unlock()
	if u.querys == nil {
		u.querys = make(map[string]struct{})
	}
	_, ok := u.querys[qid]
	if !ok {
		u.querys[qid] = struct{}{}
	}
}

func (u *uInfo) getQuery() []string {
	u.RLock()
	defer u.RUnlock()
	var res []string
	for id := range u.querys {
		res = append(res, id)
	}

	return res
}

type kInfo struct {
	keeperID  string
	online    bool
	availTime int64
	keepItem  *role.KeeperItem
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
				utils.MLogger.Error("provider deploying resolver and offer failed: ", err)
				time.Sleep(2 * time.Minute)
			} else {
				break
			}
		}

		err = m.loadContracts()
		if err != nil {
			utils.MLogger.Info("Save ", m.localID, " 's provider info err: ", err)
		}
	}()

	utils.MLogger.Info("Get ", m.localID, "'s contract info success")

	go m.getKpMapRegular(ctx)
	go m.sendStorageRegular(ctx)
	go m.saveRegular(ctx)

	m.state = true

	utils.MLogger.Info("Provider Service is ready")
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

func newGroup(localID, uid, gid string, kps []string, pros []string) *groupInfo {
	g := &groupInfo{
		userID:    uid,
		groupID:   gid,
		keepers:   kps,
		providers: pros,
		sessionID: uuid.New(),
	}

	g.loadContracts(localID)

	return g
}

func (p *Info) newGroupWithFS(userID, groupID string, kpids string, flag bool) *groupInfo {
	if kpids == "" && flag {
		ctx := context.Background()
		kmkps, err := metainfo.NewKeyMeta(groupID, metainfo.LogFS, userID)
		if err != nil {
			return nil
		}

		res, err := p.ds.GetKey(ctx, kmkps.ToString(), "local")
		if err != nil {
			return nil
		}
		kpids = string(res)
	}

	splitedMeta := strings.Split(kpids, metainfo.DELIMITER)
	var tmpKps []string
	var tmpPros []string
	if len(splitedMeta) == 2 {
		kps := splitedMeta[0]
		for i := 0; i < len(kps)/utils.IDLength; i++ {
			kid := string(kps[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(kid)
			if err != nil {
				continue
			}
			tmpKps = append(tmpKps, kid)
		}

		kps = splitedMeta[1]
		for i := 0; i < len(kps)/utils.IDLength; i++ {
			kid := string(kps[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(kid)
			if err != nil {
				continue
			}
			tmpPros = append(tmpPros, kid)
		}
	}

	if len(tmpKps) == 0 {
		tmpKps = append(tmpKps, groupID)
	}

	if len(tmpPros) == 0 {
		tmpPros = append(tmpPros, groupID)
	}

	gp := newGroup(p.localID, userID, groupID, tmpKps, tmpPros)
	if gp != nil {
		p.fsGroup.Store(groupID, gp)
	}
	return gp
}

func (p *Info) getGroupInfo(userID, groupID string, mode bool) *groupInfo {
	groupI, ok := p.fsGroup.Load(groupID)
	if !ok {
		return p.newGroupWithFS(userID, groupID, "", mode)
	}

	return groupI.(*groupInfo)
}

type quKey struct {
	uid string
	qid string
}

func (p *Info) getGroups() []quKey {
	var res []quKey
	p.fsGroup.Range(func(key, value interface{}) bool {
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

func (p *Info) load(ctx context.Context) error {
	localID := p.localID
	// load keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Keepers)
	if err != nil {

		return err
	}

	kids, err := p.ds.GetKey(ctx, kmKID.ToString(), "local")

	if err == nil && len(kids) > 0 {
		utils.MLogger.Info(localID, " has keepers: ", string(kids))
		for i := 0; i < len(kids)/utils.IDLength; i++ {
			tmpKid := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(tmpKid)
			if err != nil {
				continue
			}

			if p.ds.Connect(ctx, tmpKid) {
				thisKinfo := &kInfo{
					keeperID: tmpKid,
				}
				p.keepers.Store(tmpKid, thisKinfo)
			}
		}
	}

	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Users)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	users, err := p.ds.GetKey(ctx, kmUID.ToString(), "local")

	if err == nil && len(users) > 0 {
		for i := 0; i < len(users)/utils.IDLength; i++ {
			userID := string(users[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(userID)
			if err != nil {
				continue
			}

			utils.MLogger.Info("Load user: ", userID, " 's infomations")
			wg.Add(1)
			go func(userID string) {
				defer wg.Done()
				kmfs, err := metainfo.NewKeyMeta(userID, metainfo.Query)
				if err != nil {
					return
				}

				qs, err := p.ds.GetKey(ctx, kmfs.ToString(), "local")
				if err != nil {
					return
				}

				ui := &uInfo{
					userID: userID,
				}

				p.users.Store(userID, ui)

				for i := 0; i < len(qs)/utils.IDLength; i++ {
					qid := string(qs[i*utils.IDLength : (i+1)*utils.IDLength])
					_, err := peer.IDB58Decode(qid)
					if err != nil {
						continue
					}

					ui.setQuery(qid)

					p.getGroupInfo(userID, qid, true)
				}
			}(userID)
		}
	}

	wg.Wait()

	return nil
}

func (p *Info) save(ctx context.Context) error {
	if !p.state {
		return errProviderServiceNotReady
	}

	localID := p.localID

	var pids strings.Builder

	// persist keepers
	kmKID, err := metainfo.NewKeyMeta(localID, metainfo.Keepers)
	if err != nil {
		return err
	}

	p.keepers.Range(func(key, value interface{}) bool {
		pids.WriteString(key.(string))
		return true
	})

	if pids.Len() > 0 {
		err = p.ds.PutKey(ctx, kmKID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	pids.Reset()
	kmUID, err := metainfo.NewKeyMeta(localID, metainfo.Users)
	if err != nil {
		return err
	}

	p.users.Range(func(key, value interface{}) bool {
		pids.WriteString(key.(string))
		return true
	})

	if pids.Len() > 0 {
		err = p.ds.PutKey(ctx, kmUID.ToString(), []byte(pids.String()), "local")
		if err != nil {
			return err
		}
	}

	p.users.Range(func(key, value interface{}) bool {
		pids.Reset()
		uid := key.(string)
		ui := value.(*uInfo)
		qus := ui.getQuery()
		if len(qus) > 0 {
			kmQID, err := metainfo.NewKeyMeta(uid, metainfo.Query)
			if err != nil {
				return true
			}

			for _, qid := range qus {
				pids.WriteString(qid)
			}

			if pids.Len() > 0 {
				err = p.ds.PutKey(ctx, kmQID.ToString(), []byte(pids.String()), "local")
				if err != nil {
					return true
				}
			}
		}
		return true
	})

	// store keepers and channel value
	res := p.getGroups()
	pids.Reset()
	for _, qu := range res {
		p.saveChannelValue(qu.uid, qu.qid, p.localID)
	}

	return nil
}

func (p *Info) getKpMapRegular(ctx context.Context) {
	utils.MLogger.Info("Get kpMap from chain start!")
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
	utils.MLogger.Info("Send storages to keepers start!")
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
		utils.MLogger.Info("construct StorageSync KV error :", err)
		return err
	}

	value := strconv.FormatUint(maxSpace, 10) + metainfo.DELIMITER + strconv.FormatUint(actulDataSpace, 10)

	for _, kid := range klist {
		_, err = p.ds.SendMetaRequest(ctx, int32(metainfo.Put), km.ToString(), []byte(value), nil, kid)
		if err != nil {
			utils.MLogger.Info("storage info send to", kid, "error: ", err)
		}
	}

	return nil
}
