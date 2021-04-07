package provider

import (
	"context"
	"errors"
	"math/big"
	"net"
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
	"github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/contracts"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
	ma "github.com/multiformats/go-multiaddr"
	mdns "github.com/multiformats/go-multiaddr-dns"
	mnet "github.com/multiformats/go-multiaddr-net"
)

// Info tracks provider's information
type Info struct {
	localID           string
	sk                string
	state             bool
	enablePos         bool
	ds                data.Service
	StorageTotal      uint64
	StorageUsed       uint64
	StoragePosUsed    uint64
	LocalStorageTotal uint64
	LocalStorageFree  uint64
	TotalIncome       *big.Int
	ReadIncome        *big.Int
	StorageIncome     *big.Int
	PosIncome         *big.Int
	context           context.Context
	fsGroup           sync.Map // key: queryID, value: *groupInfo
	users             sync.Map // key: userID, value: *uInfo
	keepers           sync.Map // key: keeperID, value: *kInfo
	providers         sync.Map // key: proID, value: *kInfo
	offers            []*role.OfferItem
	proContract       *role.ProviderItem
	userConfigs       *lru.ARCCache
	ms                *measure
	ExtAddr           string
	serverAddr        string
	serverTime        time.Time
	StartTime         time.Time
}

type ProviderStartOption struct {
	Capacity      int64
	Duration      int64
	DepositSize   int64
	Price         *big.Int
	ReDeployOffer bool
	EnablePos     bool
	Gc            bool
	ExtAddr       string
}

type measure struct {
	balance     metrics.Gauge
	storageUsed metrics.Gauge
	storageFree metrics.Gauge
	groupNum    metrics.Gauge
	userNum     metrics.Gauge
	providerNum metrics.Gauge
	keeperNum   metrics.Gauge
}

type groupInfo struct {
	status       bool
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
func New(ctx context.Context, id, sk string, ds data.Service, rt routing.Routing, capacity, duration, depositSize int64, price *big.Int, reDeployOffer, enablePos, gc bool, extAddr string) (instance.Service, error) {
	mea := &measure{
		balance:     metrics.New("provider.balance", "Balance of this provider").Gauge(),
		groupNum:    metrics.New("provider.group_num", "Group number").Gauge(),
		userNum:     metrics.New("provider.user_num", "User number").Gauge(),
		keeperNum:   metrics.New("provider.keeper_num", "Keeper number").Gauge(),
		providerNum: metrics.New("provider.provider_num", "Providers number").Gauge(),
		storageUsed: metrics.New("provider.storage_used", "Storage used(bytes)").Gauge(),
		storageFree: metrics.New("provider.storage_free", "Local available storage(bytes)").Gauge(),
	}

	m := &Info{
		localID:       id,
		sk:            sk,
		ds:            ds,
		context:       ctx,
		ms:            mea,
		enablePos:     enablePos,
		TotalIncome:   big.NewInt(0),
		ReadIncome:    big.NewInt(0),
		StorageIncome: big.NewInt(0),
		PosIncome:     big.NewInt(0),
		offers:        make([]*role.OfferItem, 0, 1),
		StartTime:     time.Now(),
	}

	err := rt.(*dht.KadDHT).AssignmetahandlerV2(m)
	if err != nil {
		return nil, err
	}

	err = m.loadContracts(capacity, duration, depositSize, price, reDeployOffer)
	if err != nil {
		utils.MLogger.Error("provider load contarct failed: ", err)
		return nil, err
	}

	if m.enablePos {
		go func() {
			err := m.PosService(ctx, gc)
			if err != nil {
				utils.MLogger.Errorf("start pos err: %s ", err)
			}
		}()
	}

	if extAddr != "" {
		utils.MLogger.Info("extAddress is set to: ", extAddr)
		eaddr := strings.Split(extAddr, ":")
		if len(eaddr) == 2 {
			if net.ParseIP(eaddr[0]) != nil {
				ips := strings.Split(eaddr[0], ".")
				if len(ips) == 4 {
					// example:= /ip4/123.123.234.123/tcp/50272/p2p/8MH3B8DT14cJrVFpZ8TjmJ6NRUfVew
					m.ExtAddr = "/ip4/" + eaddr[0] + "/tcp/" + eaddr[1] + "/p2p/" + m.localID
				} else {
					// example:= /ip4/123.123.234.123/tcp/50272/p2p/8MH3B8DT14cJrVFpZ8TjmJ6NRUfVew
					m.ExtAddr = "/ip6/" + eaddr[0] + "/tcp/" + eaddr[1] + "/p2p/" + m.localID
				}
			} else {
				// example:= /dns/239v39e500.zicp.vip/tcp/50272/p2p/8MH3B8DT14cJrVFpZ8TjmJ6NRUfVew
				m.ExtAddr = "/dns/" + eaddr[0] + "/tcp/" + eaddr[1] + "/p2p/" + m.localID
			}
		}
	}

	return m, nil
}

func (p *Info) Start(ctx context.Context, opts interface{}) error {
	// cache userconfigs, key is queryID
	ucache, err := lru.NewARC(1024)
	if err != nil {
		utils.MLogger.Error("new lru err:", err)
		return err
	}
	p.userConfigs = ucache

	balance := role.GetBalance(p.localID)
	ba, _ := new(big.Float).SetInt(balance).Float64()
	p.ms.balance.Set(ba)

	usedCapacity, err := p.getDiskUsage()
	if err == nil {
		p.ms.storageUsed.Set(float64(usedCapacity))
	}

	p.StorageUsed = usedCapacity

	lsinfo, err := role.GetDiskSpaceInfo()
	if err != nil {
		return err
	}

	p.LocalStorageTotal = lsinfo.Total
	p.LocalStorageFree = lsinfo.Free
	p.ms.storageFree.Set(float64(p.LocalStorageFree))

	p.StorageTotal = p.getDiskTotal()

	if p.LocalStorageTotal < p.StorageTotal {
		utils.MLogger.Errorf("%s has pledge space %d, but local storage has %d", p.localID, p.StorageTotal, p.LocalStorageTotal)
	}

	utils.MLogger.Info("Get ", p.localID, "'s contract info success")

	p.ms.providerNum.Inc()
	utils.MLogger.Info("Take charge of network handler")

	err = p.load(ctx)
	if err != nil {
		utils.MLogger.Error("provider load local info failed: ", err)
		return err
	}

	go p.getFromChainRegular(ctx)
	go p.sendStorageRegular(ctx)
	go p.saveRegular(ctx)

	p.extAddrSync(ctx)
	p.GetPublicAddress()

	p.state = true

	utils.MLogger.Info("Provider Service is ready")
	return nil
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
func (p *Info) GetPublicAddress() (string, error) {
	if p.serverAddr != "" && p.serverTime.Add(time.Hour).After(time.Now()) {
		return p.serverAddr, nil
	}

	e, err := p.ds.GetPublicAddr(p.context, p.localID)
	if err != nil {
		return "", err
	}

	eAddr := e.String()
	if strings.Contains(eAddr, "dns") {
		ctx, cancle := context.WithTimeout(p.context, 30*time.Second)
		defer cancle()
		addrs, _ := mdns.Resolve(ctx, e)
		for _, maddr := range addrs {
			ok := mnet.IsPrivateAddr(maddr)
			if !ok {
				eAddr = maddr.String()
				break
			}
		}
	}

	if strings.Contains(eAddr, p.localID) {
		p.serverAddr = eAddr
		p.serverTime = time.Now()
	} else {
		p.serverAddr = eAddr + "/p2p/" + p.localID
		p.serverTime = time.Now()
	}

	return p.serverAddr, nil
}

func newGroup(localID, uid, gid string, kps []string, pros []string) *groupInfo {
	utils.MLogger.Infof("Create user %s fsID %s groupInfo", uid, gid)
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
		go p.loadChannelValue(userID, groupID)
		for _, kid := range gp.keepers {
			p.getKInfo(kid, true)
		}

		for _, pid := range gp.providers {
			go p.getPInfo(pid, true)
		}

		ui := p.getUserInfo(userID)
		if ui != nil {
			ui.setQuery(gp.groupID)
		}

		kmsess, err := metainfo.NewKey(groupID, mpb.KeyType_Session, userID)
		if err != nil {
			return nil
		}

		sessByte, err := p.ds.GetKey(p.context, kmsess.ToString(), "local")
		if err == nil && len(sessByte) > 0 {
			sID, err := uuid.ParseBytes(sessByte)
			if err == nil {
				gp.sessionID = sID
				gp.sessionTime = time.Now().Unix() - 1500 //
			}
		}

		gp.status = true
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
			querys: make(map[string]struct{}),
		}

		utils.MLogger.Infof("add new user %s", pid)
		p.ms.userNum.Inc()
		p.users.Store(pid, ui)
		return ui
	}

	return ui.(*uInfo)
}

func (p *Info) getKInfo(pid string, managed bool) *kInfo {
	ui, ok := p.keepers.Load(pid)
	if !ok {

		r := contracts.NewCR(p.localID, "")
		has, err := r.IsKeeper(pid)
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

		if _, success := p.ds.Connect(p.context, pid); success {
			utils.MLogger.Infof("add new keeper %s", pid)
			ui.availTime = time.Now().Unix()
			ui.online = true
			p.ms.keeperNum.Inc()
			p.keepers.Store(pid, ui)
			return ui
		}

		if managed {
			utils.MLogger.Infof("add new keeper %s", pid)
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
	r := contracts.NewCR(p.localID, "")
	if !ok {
		has, err := r.IsProvider(pid)
		if err != nil || !has {
			return nil
		}

		ui := &pInfo{
			providerID: pid,
		}

		if _, success := p.ds.Connect(p.context, pid); success {
			utils.MLogger.Infof("add new provider %s", pid)
			ui.availTime = time.Now().Unix()
			ui.online = true
			p.ms.providerNum.Inc()
			p.providers.Store(pid, ui)
			return ui
		}

		if managed {
			utils.MLogger.Infof("add new provider %s", pid)
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

	var wg sync.WaitGroup
	kids, err := p.ds.GetKey(ctx, kmKID.ToString(), "local")
	if err == nil && len(kids) > 0 {
		utils.MLogger.Info(localID, " has keepers: ", string(kids))
		for i := 0; i < len(kids)/utils.IDLength; i++ {
			tmpKid := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(tmpKid)
			if err != nil {
				continue
			}

			wg.Add(1)
			go func(keepID string) {
				defer wg.Done()
				p.getKInfo(keepID, false)
			}(tmpKid)
		}
	}

	wg.Wait()

	utils.MLogger.Info("Load all users' infomations begin")
	kmUID, err := metainfo.NewKey(localID, mpb.KeyType_Users)
	if err != nil {
		return err
	}

	users, err := p.ds.GetKey(ctx, kmUID.ToString(), "local")
	if err == nil && len(users) > 0 {
		for i := 0; i < len(users)/utils.IDLength; i++ {
			userID := string(users[i*utils.IDLength : (i+1)*utils.IDLength])
			_, err := peer.IDB58Decode(userID)
			if err != nil {
				continue
			}

			// pos user is init in pos service
			if userID == pos.GetPosId() {
				continue
			}

			utils.MLogger.Info("Load user: ", userID, " 's infomations begin")
			go func(userID string) {
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

				for j := 0; j < len(qs)/utils.IDLength; j++ {
					qid := string(qs[j*utils.IDLength : (j+1)*utils.IDLength])
					_, err := peer.IDB58Decode(qid)
					if err != nil {
						continue
					}

					ui.setQuery(qid)

					p.getGroupInfo(userID, qid, true)
				}

				utils.MLogger.Info("Load user: ", userID, " 's infomations finished")
			}(userID)
		}
	}

	utils.MLogger.Info("Load all users' infomations...")

	return nil
}

func (p *Info) loadChannel(ctx context.Context) {
	// todo
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

//GetIncomeAddress get upkeepingAddress and channelAddress of this provider to filter logs in chain
func (p *Info) GetIncomeAddress() ([]common.Address, []common.Address, []common.Address) {
	ukAddr := []common.Address{}
	posAddr := []common.Address{}
	channelAddr := []common.Address{} // missing some
	pus := p.getGroups()
	for _, pu := range pus {
		if pu.uid == pu.qid {
			continue
		}

		gp := p.getGroupInfo(pu.uid, pu.qid, false)
		if gp == nil || gp.upkeeping == nil {
			continue
		}

		tmp, err := address.GetAddressFromID(gp.upkeeping.UpKeepingID)
		if err != nil {
			continue
		}

		if pu.uid == pos.GetPosId() {
			posAddr = append(posAddr, tmp)
			continue
		} else {
			ukAddr = append(ukAddr, tmp)
		}

		for _, proID := range gp.providers {
			chanIDs, err := role.GetAllChannels(pu.uid, pu.qid, proID)
			if err != nil {
				continue
			}
			for _, chanID := range chanIDs {
				ba, err := role.QueryBalance(chanID)
				if err != nil {
					continue
				}

				if ba.Sign() > 0 {
					continue
				}

				chanAddr, err := address.GetAddressFromID(chanID)
				if err != nil {
					continue
				}
				channelAddr = append(channelAddr, chanAddr)
			}
		}
	}

	if len(posAddr) == 0 {
		qItem, err := role.GetLatestQuery(pos.GetPosId())
		if err != nil {
			return ukAddr, posAddr, channelAddr
		}
		uItem, err := role.GetUpKeeping(pos.GetPosId(), qItem.QueryID)
		if err != nil {
			return ukAddr, posAddr, channelAddr
		}
		localAddr, err := address.GetAddressFromID(p.localID)
		if err != nil {
			return ukAddr, posAddr, channelAddr
		}
		for _, pi := range uItem.Providers {
			if pi.Addr.String() == localAddr.String() {
				uAddr, err := address.GetAddressFromID(uItem.UpKeepingID)
				if err != nil {
					return ukAddr, posAddr, channelAddr
				}
				posAddr = append(posAddr, uAddr)
			}
		}

	}

	return ukAddr, posAddr, channelAddr
}

func (p *Info) getIncome(localAddr common.Address, pBlock int64) (int64, error) {
	b, err := contracts.GetLatestBlock()
	if err != nil {
		return 0, err
	}

	latestBlock := b.Number().Int64()
	endBlock := b.Number().Int64()
	ukaddrs, posAddrs, chanAddrs := p.GetIncomeAddress()

	storageBlock := pBlock
	if len(ukaddrs) > 0 && latestBlock > storageBlock {
		utils.MLogger.Infof("get storage income from chain")
		endBlock = latestBlock

		for endBlock <= latestBlock {
			if endBlock > storageBlock+1024 {
				endBlock = storageBlock + 1024
			}

			sIncome, _, err := contracts.GetStorageIncome(ukaddrs, localAddr, storageBlock, endBlock)
			if err != nil {
				utils.MLogger.Info("get ukpay log err:", err)
				break
			}

			p.StorageIncome.Add(p.StorageIncome, sIncome)
			p.TotalIncome.Add(p.TotalIncome, sIncome)
			storageBlock = endBlock

			if endBlock == latestBlock {
				break
			}

			if endBlock < latestBlock {
				endBlock = latestBlock
			}
		}
	}

	posBlock := pBlock

	if len(posAddrs) > 0 && latestBlock > posBlock {
		utils.MLogger.Infof("get pos income from chain")

		endBlock = latestBlock

		for endBlock <= latestBlock {
			if endBlock > posBlock+1024 {
				endBlock = posBlock + 1024
			}

			sIncome, _, err := contracts.GetStorageIncome(posAddrs, localAddr, posBlock, endBlock)
			if err != nil {
				utils.MLogger.Info("get pos ukpay log err:", err)
				break
			}

			p.PosIncome.Add(p.PosIncome, sIncome)
			p.TotalIncome.Add(p.TotalIncome, sIncome)
			posBlock = endBlock

			if endBlock == latestBlock {
				break
			}

			if endBlock < latestBlock {
				endBlock = latestBlock
			}
		}
	}

	readBlock := pBlock

	if len(chanAddrs) > 0 && latestBlock > readBlock {
		utils.MLogger.Infof("get read income from chain")

		endBlock = latestBlock
		for endBlock <= latestBlock {
			if endBlock > readBlock+1024 {
				endBlock = readBlock + 1024
			}

			sIncome, _, err := contracts.GetReadIncome(chanAddrs, localAddr, readBlock, endBlock)
			if err != nil {
				utils.MLogger.Info("get readpay log err:", err)
				break
			}

			p.ReadIncome.Add(p.ReadIncome, sIncome)
			p.TotalIncome.Add(p.TotalIncome, sIncome)
			readBlock = endBlock

			if endBlock == latestBlock {
				break
			}

			if endBlock < latestBlock {
				endBlock = latestBlock
			}
		}
	}

	km, err := metainfo.NewKey(p.localID, mpb.KeyType_Income)
	if err == nil {
		var res strings.Builder
		res.WriteString(strconv.FormatInt(latestBlock, 10))
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(p.TotalIncome.String())
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(p.StorageIncome.String())
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(p.ReadIncome.String())
		res.WriteString(metainfo.DELIMITER)
		res.WriteString(p.PosIncome.String())

		p.ds.PutKey(p.context, km.ToString(), []byte(res.String()), nil, "local")
	}

	utils.MLogger.Infof("get income from chain finished at block %d", latestBlock)
	return latestBlock, nil
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

	localAddr, err := address.GetAddressFromID(p.localID)
	if err != nil {
		return
	}

	lastBlock := int64(0)

	km, err := metainfo.NewKey(p.localID, mpb.KeyType_Income)
	if err == nil {
		res, err := p.ds.GetKey(p.context, km.ToString(), "local")
		if err == nil && len(res) > 0 {
			utils.MLogger.Infof("Load %s income info: %s", km.ToString(), string(res))
			ins := strings.Split(string(res), metainfo.DELIMITER)
			if len(ins) == 5 {
				lb, err := strconv.ParseInt(ins[0], 10, 0)
				if err == nil {
					lastBlock = lb
				}

				ti, ok := new(big.Int).SetString(ins[1], 10)
				if ok {
					p.TotalIncome = ti
				}

				si, ok := new(big.Int).SetString(ins[2], 10)
				if ok {
					p.StorageIncome = si
				}

				ri, ok := new(big.Int).SetString(ins[3], 10)
				if ok {
					p.ReadIncome = ri
				}

				pi, ok := new(big.Int).SetString(ins[4], 10)
				if ok {
					p.PosIncome = pi
				}
			}
		}
	}

	p.loadPeersFromChain()

	time.Sleep(7 * time.Minute)
	lb, err := p.getIncome(localAddr, lastBlock)
	if err == nil {
		lastBlock = lb
	}

	ticker := time.NewTicker(37 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			lb, err := p.getIncome(localAddr, lastBlock)
			if err == nil {
				lastBlock = lb
			}
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
			p.extAddrSync(ctx)
		}
	}
}

func (p *Info) extAddrSync(ctx context.Context) error {
	if p.ExtAddr == "" {
		return nil
	}

	km, err := metainfo.NewKey(p.localID, mpb.KeyType_ExternalAddress)
	if err != nil {
		utils.MLogger.Info("construct StorageSync KV error :", err)
		return err
	}

	for _, defaultBootstrapAddress := range config.DefaultBootstrapAddresses {
		bi, err := ma.NewMultiaddr(defaultBootstrapAddress)
		if err != nil {
			continue
		}

		pi, err := peer.AddrInfoFromP2pAddr(bi)
		if err != nil {
			continue
		}

		ok := p.ds.FastConnect(p.context, pi.ID.Pretty())
		if !ok {
			continue
		}

		p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(p.ExtAddr), nil, pi.ID.Pretty())
	}

	p.keepers.Range(func(key, value interface{}) bool {
		go p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(p.ExtAddr), nil, key.(string))
		return true
	})

	return nil
}

func (p *Info) storageSync(ctx context.Context) error {
	balance := role.GetBalance(p.localID)
	ba, _ := new(big.Float).SetInt(balance).Float64()
	p.ms.balance.Set(ba)

	actulDataSpace, err := p.getDiskUsage()
	if err != nil {
		return err
	}

	maxSpace := p.getDiskTotal()
	p.StorageUsed = actulDataSpace
	p.StorageTotal = maxSpace

	p.ms.storageUsed.Set(float64(actulDataSpace))

	lsinfo, err := role.GetDiskSpaceInfo()
	if err != nil {
		return err
	}

	p.LocalStorageTotal = lsinfo.Total
	p.LocalStorageFree = lsinfo.Free

	p.ms.storageFree.Set(float64(p.LocalStorageFree))

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

//SetProviderStop send 'quit' message to keepers, so they can set provider stop in upkeeping
func (p *Info) SetProviderStop(ctx context.Context, groupID string, response string) error {
	utils.MLogger.Info("Send 'quit' message to keeper start!")
	gpInfo := p.getGroupInfo("", groupID, false)
	if gpInfo == nil {
		return ErrGroupNotReady
	}

	km, err := metainfo.NewKey(groupID, mpb.KeyType_ProQuit, gpInfo.userID)
	if err != nil {
		return err
	}
	key := km.ToString()

	var wg sync.WaitGroup
	for _, keeperID := range gpInfo.keepers {
		wg.Add(1)
		go func(k, kID string, data []byte) {
			defer wg.Done()
			utils.MLogger.Info("Send 'quit' message to keeper:", kID)
			p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), k, data, nil, kID)
		}(key, keeperID, []byte(response))
	}
	wg.Wait()

	localAddr, err := address.GetAddressFromID(p.localID)
	if err != nil {
		return err
	}

	index := 0
	for i, pItem := range gpInfo.upkeeping.Providers {
		if pItem.Addr.String() == localAddr.String() {
			index = i
			break
		}
	}

	i := 0
	for ; i < 10; i++ {
		time.Sleep(30 * time.Second)
		if gpInfo.upkeeping.Providers[index].Stop {
			break
		}
	}
	if i > 9 {
		return errors.New("set provider stop failed")
	}

	utils.MLogger.Info("have been successfully set stop in upkeeping ", gpInfo.upkeeping.UpKeepingID)
	return nil
}

//MoveData provider want to quit group, so need move its data to another provider
func (p *Info) MoveData(ctx context.Context, groupID string) (string, error) {
	utils.MLogger.Info("move data to another provider start!")
	gpInfo := p.getGroupInfo("", groupID, false)
	if gpInfo == nil || len(gpInfo.keepers) <= 0 {
		return "", ErrGroupNotReady
	}

	km, err := metainfo.NewKey(groupID, mpb.KeyType_MoveData, p.localID)
	if err != nil {
		return "", err
	}

	//get blockMeta from keepers
	var res []byte
	for _, keeperID := range gpInfo.keepers {
		res, err = p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), km.ToString(), nil, nil, keeperID)
		if err == nil && res != nil {
			break
		}
	}

	if err != nil {
		utils.MLogger.Error("get move data response err:", err)
		return "", err
	}
	// res: bid_sid_cid_offset_newPid/bid_sid_cid_offset_newPid..
	if res == nil {
		utils.MLogger.Info("get move data response is nil")
		return "", nil
	}
	utils.MLogger.Info("get move data infomation: ", string(res))
	response := strings.Split(string(res), metainfo.DELIMITER)

	for _, r := range response {
		//binfo: [bid sid cid offset newPid]
		binfo := strings.Split(r, metainfo.BlockDelimiter)
		if len(binfo) != 5 {
			utils.MLogger.Warn("get length of moveData binfo error ", r)
			return "", errors.New("get length of moveData binfo error")
		}

		//send block to another provider: 1. get block from local
		bm, err := metainfo.NewBlockMeta(groupID, binfo[0], binfo[1], binfo[2])
		if err != nil {
			utils.MLogger.Warn("get bm of moveData error: ", err)
			return "", err
		}

		bk, err := metainfo.NewKey(bm.ToString(), mpb.KeyType_Block, "0", binfo[3])
		if err != nil {
			utils.MLogger.Warn("get newKey of moveData-getBlock error: ", err)
			return "", err
		}

		b, err := p.ds.GetBlock(ctx, bk.ToString(), nil, "local")
		if err != nil {
			utils.MLogger.Warn("get block ", bk.ToString(), "of moveData error: ", err)
			return "", err
		}
		utils.MLogger.Info("get block", bk.ToString(), "from local to move")

		//send block to another provider: 2. move block to another provider
		cid := bm.ToString()
		km, err := metainfo.NewKey(cid, mpb.KeyType_Block)
		if err != nil {
			utils.MLogger.Warn("get newKey of moveData-putBlock error: ", err)
			return "", err
		}

		for i := 0; i < 10; i++ {
			err = p.ds.PutBlock(ctx, km.ToString(), b.RawData(), binfo[4])
			if err != nil {
				utils.MLogger.Warn("put block", km.ToString(), "of moveData-putBlock error: ", err)
				continue
			} else {
				utils.MLogger.Info("put block ", km.ToString(), " for moving data")
				break
			}
		}
		if err != nil {
			return "", err
		}
	}

	//delete local block
	for _, r := range response {
		//binfo: [bid sid cid offset newPid]
		binfo := strings.Split(r, metainfo.BlockDelimiter)

		bm, err := metainfo.NewBlockMeta(groupID, binfo[0], binfo[1], binfo[2])
		if err != nil {
			utils.MLogger.Warn("get bm of moveData-deleteBlock error: ", err)
			return "", err
		}

		err = p.ds.DeleteBlock(p.context, bm.ToString(), "local")
		if err != nil {
			utils.MLogger.Warn("delete block ", bm.ToString(), "from local error: ", err)
			return "", err
		}
		utils.MLogger.Info("delete block", bm.ToString(), "from local success")
	}
	return string(res), nil
}
