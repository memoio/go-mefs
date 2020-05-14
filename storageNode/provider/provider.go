package provider

import (
	"context"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru"
	metrics "github.com/ipfs/go-metrics-interface"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
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
	context      context.Context
	fsGroup      sync.Map // key: queryID, value: *groupInfo
	users        sync.Map // key: userID, value: *uInfo
	keepers      sync.Map // key: keeperID, value: *kInfo
	providers    sync.Map // key: proID, value: *kInfo
	offers       []*role.OfferItem
	proContract  *role.ProviderItem
	userConfigs  *lru.ARCCache
	ms           *measure
}

type measure struct {
	balance     metrics.Gauge
	storageUsed metrics.Gauge
	groupNum    metrics.Gauge
	userNum     metrics.Gauge
	providerNum metrics.Gauge
	keeperNum   metrics.Gauge
}

type groupInfo struct {
	sessionID    uuid.UUID
	sessionTime  int64
	userID       string
	groupID      string
	storageUsed  uint64
	storageTotal uint64
	keepers      []string
	providers    []string
	upkeeping    *role.UpKeepingItem
	channel      sync.Map //key is channelID
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

type pInfo struct {
	providerID string
	online     bool
	availTime  int64
}

//New start provider service
func New(ctx context.Context, id, sk string, ds data.Service, rt routing.Routing, capacity, duration, depositSize int64, price *big.Int, reDeployOffer, enablePos, gc bool) (instance.Service, error) {
	mea := &measure{
		balance:     metrics.New("provider.balance", "Balance of this provider").Gauge(),
		groupNum:    metrics.New("provider.group_num", "Group number").Gauge(),
		userNum:     metrics.New("provider.user_num", "User number").Gauge(),
		keeperNum:   metrics.New("provider.keeper_num", "Keeper number").Gauge(),
		providerNum: metrics.New("provider.provider_num", "Providers number").Gauge(),
		storageUsed: metrics.New("provider.storage_used", "Storage used(bytes)").Gauge(),
	}

	m := &Info{
		localID: id,
		sk:      sk,
		ds:      ds,
		context: ctx,
		ms:      mea,
		offers:  make([]*role.OfferItem, 0, 1),
	}
	err := rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
		return nil, err
	}

	// cache userconfigs, key is queryID
	ucache, err := lru.NewARC(1024)
	if err != nil {
		utils.MLogger.Error("new lru err:", err)
		return nil, err
	}
	m.userConfigs = ucache

	balance := role.GetBalance(m.localID)
	ba, _ := new(big.Float).SetInt(balance).Float64()
	m.ms.balance.Set(ba)

	usedCapacity, err := m.getDiskUsage()
	if err == nil {
		m.ms.storageUsed.Set(float64(usedCapacity))
	}

	m.ms.providerNum.Inc()

	err = m.loadContracts(capacity, duration, depositSize, price, reDeployOffer)
	if err != nil {
		utils.MLogger.Error("provider load contarct failed: ", err)
		return nil, err
	}

	utils.MLogger.Info("Get ", m.localID, "'s contract info success")

	go m.getFromChainRegular(ctx)
	go m.sendStorageRegular(ctx)
	go m.saveRegular(ctx)

	m.state = true
	if enablePos {
		go func() {
			err := m.PosService(ctx, gc)
			if err != nil {
				utils.MLogger.Errorf("start pos err: %s ", err)
			}
		}()
	}

	utils.MLogger.Info("Provider Service is ready")
	return m, nil
}

func (p *Info) Online() bool {
	return p.state
}

func (p *Info) GetRole() string {
	return metainfo.RoleProvider
}

func (p *Info) Close() error {
	return p.save(p.context)
}

func (p *Info) GetStorageInfo() (int64, uint64) {
	return p.proContract.Capacity, p.storageUsed
}

//GetIncomeInfo get upkeepingAddress and channelAddress of this provider to filter logs in chain
func (p *Info) GetIncomeInfo() ([]common.Address, []common.Address) {
	ukAddr := []common.Address{}
	channelAddr := []common.Address{}
	p.fsGroup.Range(func(key, value interface{}) bool {
		gInfo, ok := value.(*groupInfo)
		if !ok {
			return true
		}
		ukid := gInfo.upkeeping.UpKeepingID
		tmp, err := address.GetAddressFromID(ukid)
		if err != nil {
			return false
		}
		ukAddr = append(ukAddr, tmp)

		gInfo.channel.Range(func(key, value interface{}) bool {
			cItem, ok := value.(*role.ChannelItem)
			if !ok {
				return true
			}
			if cItem.ProID == p.localID {
				channelAddr = append(channelAddr) //此处需要return false终止遍历吗
			}
			return true
		})
		return true
	})
	return ukAddr, channelAddr
}

func newGroup(localID, uid, gid string, kps []string, pros []string) *groupInfo {
	g := &groupInfo{
		userID:    uid,
		groupID:   gid,
		keepers:   kps,
		providers: pros,
		sessionID: uuid.Nil,
	}

	g.loadContracts(localID, true)

	return g
}

func (p *Info) newGroupWithFS(userID, groupID string, kpids string) *groupInfo {
	var tmpKps []string
	var tmpPros []string

	if kpids == "" && userID == groupID {
		ctx := p.context
		kmkps, err := metainfo.NewKey(groupID, mpb.KeyType_LFS, userID)
		if err != nil {
			return nil
		}

		res, _ := p.ds.GetKey(ctx, kmkps.ToString(), "local")
		kpids = string(res)
	}

	if userID == groupID {
		splitedMeta := strings.Split(kpids, metainfo.DELIMITER)

		has := false
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
				pid := string(kps[i*utils.IDLength : (i+1)*utils.IDLength])
				_, err := peer.IDB58Decode(pid)
				if err != nil {
					continue
				}

				if pid == p.localID {
					has = true
				}

				tmpPros = append(tmpPros, pid)
			}
		}

		if len(tmpKps) == 0 || len(tmpPros) == 0 {
			utils.MLogger.Warn(groupID, " has no keeper or providers")
			return nil
		}

		if !has {
			utils.MLogger.Warn(groupID, " is not my user")
			return nil
		}
	}
	gp := newGroup(p.localID, userID, groupID, tmpKps, tmpPros)
	if gp != nil {
		p.ms.groupNum.Inc()
		p.fsGroup.Store(groupID, gp)
		p.loadChannelValue(userID, groupID)
		for _, kid := range gp.keepers {
			p.getKInfo(kid, true)
		}

		for _, pid := range gp.providers {
			p.getPInfo(pid, true)
		}

		ui := p.getUserInfo(userID)
		if ui != nil {
			ui.setQuery(gp.groupID)
		}
	}

	return gp
}

func (p *Info) getGroupInfo(userID, groupID string, mode bool) *groupInfo {
	groupI, ok := p.fsGroup.Load(groupID)
	if !ok {
		if mode {
			return p.newGroupWithFS(userID, groupID, "")
		}
		return nil
	}

	return groupI.(*groupInfo)
}

func (p *Info) getUserInfo(pid string) *uInfo {
	ui, ok := p.users.Load(pid)
	if !ok {
		ui := &uInfo{
			userID: pid,
		}

		p.ms.userNum.Inc()
		p.users.Store(pid, ui)
		return ui
	}

	return ui.(*uInfo)
}

func (p *Info) getKInfo(pid string, managed bool) *kInfo {
	ui, ok := p.keepers.Load(pid)
	if !ok {

		has, err := role.IsKeeper(pid)
		if err != nil || !has {
			return nil
		}

		kItem, err := role.GetKeeperInfo(p.localID, pid)
		if err != nil {
			return nil
		}

		ui := &kInfo{
			keeperID: pid,
			keepItem: &kItem,
		}

		if p.ds.Connect(p.context, pid) {
			ui.availTime = time.Now().Unix()
			ui.online = true
			p.ms.keeperNum.Inc()
			p.keepers.Store(pid, ui)
			return ui
		}

		if managed {
			p.ms.keeperNum.Inc()
			p.keepers.Store(pid, ui)
			return ui
		}

		return nil
	}

	return ui.(*kInfo)
}

func (p *Info) getPInfo(pid string, managed bool) *pInfo {
	ui, ok := p.providers.Load(pid)
	if !ok {
		has, err := role.IsProvider(pid)
		if err != nil || !has {
			return nil
		}

		ui := &pInfo{
			providerID: pid,
		}

		if p.ds.Connect(p.context, pid) {
			ui.availTime = time.Now().Unix()
			ui.online = true
			p.ms.providerNum.Inc()
			p.providers.Store(pid, ui)
			return ui
		}

		if managed {
			p.ms.providerNum.Inc()
			p.providers.Store(pid, ui)
			return ui
		}

		return nil
	}

	return ui.(*pInfo)
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
	ticker := time.NewTicker(10 * time.Minute)
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
	kmKID, err := metainfo.NewKey(localID, mpb.KeyType_Keepers)
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

			p.getKInfo(tmpKid, false)
		}
	}

	kmUID, err := metainfo.NewKey(localID, mpb.KeyType_Users)
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
				kmfs, err := metainfo.NewKey(userID, mpb.KeyType_Query)
				if err != nil {
					return
				}

				qs, err := p.ds.GetKey(ctx, kmfs.ToString(), "local")
				if err != nil {
					return
				}

				ui := p.getUserInfo(userID)
				if ui == nil {
					return
				}

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
		return role.ErrServiceNotReady
	}

	localID := p.localID

	var kids strings.Builder

	// persist keepers
	kmKID, err := metainfo.NewKey(localID, mpb.KeyType_Keepers)
	if err != nil {
		return err
	}

	p.keepers.Range(func(key, value interface{}) bool {
		kids.WriteString(key.(string))
		return true
	})

	if kids.Len() > 0 {
		err = p.ds.PutKey(ctx, kmKID.ToString(), []byte(kids.String()), nil, "local")
		if err != nil {
			return err
		}
	}

	kids.Reset()
	kmUID, err := metainfo.NewKey(localID, mpb.KeyType_Users)
	if err != nil {
		return err
	}

	p.users.Range(func(key, value interface{}) bool {
		kids.WriteString(key.(string))
		return true
	})

	if kids.Len() > 0 {
		err = p.ds.PutKey(ctx, kmUID.ToString(), []byte(kids.String()), nil, "local")
		if err != nil {
			return err
		}
	}

	p.users.Range(func(key, value interface{}) bool {
		kids.Reset()
		uid := key.(string)
		ui := value.(*uInfo)
		qus := ui.getQuery()
		if len(qus) > 0 {
			kmQID, err := metainfo.NewKey(uid, mpb.KeyType_Query)
			if err != nil {
				return true
			}

			for _, qid := range qus {
				kids.WriteString(qid)
			}

			if kids.Len() > 0 {
				err = p.ds.PutKey(ctx, kmQID.ToString(), []byte(kids.String()), nil, "local")
				if err != nil {
					return true
				}
			}
		}
		return true
	})

	// store keepers and channel value
	res := p.getGroups()
	kids.Reset()
	for _, qu := range res {
		p.saveChannelValue(qu.uid, qu.qid, p.localID)
		p.loadChannelValue(qu.uid, qu.qid)
	}

	return nil
}

func (p *Info) loadPeersFromChain() error {
	keepers, _, err := role.GetAllKeepers(p.localID)
	if err != nil {
		return err
	}

	for _, kItem := range keepers {
		p.getKInfo(kItem.KeeperID, false)
	}

	pros, _, err := role.GetAllProviders(p.localID)
	if err != nil {
		return err
	}

	for _, pItem := range pros {
		p.getPInfo(pItem.ProviderID, false)
	}

	role.SaveKpMap(p.localID)

	return nil
}

func (p *Info) getFromChainRegular(ctx context.Context) {
	utils.MLogger.Info("Get infos from chain start!")

	p.loadPeersFromChain()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.loadPeersFromChain()
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
			p.storageSync(ctx)
		}
	}
}

func (p *Info) storageSync(ctx context.Context) error {
	balance := role.GetBalance(p.localID)
	ba, _ := new(big.Float).SetInt(balance).Float64()
	p.ms.balance.Set(ba)

	actulDataSpace, err := p.getDiskUsage()
	if err != nil {
		return err
	}

	p.ms.storageUsed.Set(float64(actulDataSpace))

	maxSpace := p.getDiskTotal()

	p.storageUsed = actulDataSpace
	p.storageTotal = maxSpace

	klist, ok := role.GetKeepersOfPro(p.localID)
	if !ok {
		return nil
	}

	km, err := metainfo.NewKey(p.localID, mpb.KeyType_Storage)
	if err != nil {
		utils.MLogger.Info("construct StorageSync KV error :", err)
		return err
	}

	value := strconv.FormatUint(maxSpace, 10) + metainfo.DELIMITER + strconv.FormatUint(actulDataSpace, 10)

	for _, kid := range klist {
		_, err = p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), km.ToString(), []byte(value), nil, kid)
		if err != nil {
			utils.MLogger.Info("storage info send to", kid, "error: ", err)
		}
	}

	return nil
}
