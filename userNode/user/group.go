package user

import (
	"context"
	"math/big"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/peer"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	stoped       int8 = iota //stoped
	starting                 //begin to start
	collecting               // broadcast UserInit
	collectDone              // notify keeper
	deploying                // deploy contracts
	depoyDone                // connect
	groupStarted             // done
)

type keeperInfo struct {
	keeperID  string
	sessionID uuid.UUID // for one write
	connected bool
}

type providerInfo struct {
	sync.Mutex
	providerID string
	connected  bool
	sessionID  uuid.UUID // for one write
	chanItem   *role.ChannelItem
	offerItem  *role.OfferItem
}

// group stores use's groupinfo
type groupInfo struct {
	sync.RWMutex
	groupID   string // id format of query address
	userID    string // id format of user address
	shareToID string // shareToID = userID when not share
	privKey   string // EthString of shareTo
	rootID    string // Root contract addr
	state     int8   // atomic?
	ds        data.Service

	keepers   map[string]*keeperInfo
	providers map[string]*providerInfo

	storeDays     int64 //表示部署合约时的存储数据时间，单位是“天”
	storeSize     int64 //表示部署合约时的存储数据大小，单位是“MB”
	storePrice    int64 //表示部署合约时的存储价格大小，单位是“wei”
	keeperSLA     int   //表示部署合约时的keeper参数，目前是keeper数量
	providerSLA   int   //表示部署合约时的provider参数，目前是provider数量
	stPayCycle    int64
	reDeploy      bool // 是否重新部署offer
	force         bool
	tempKeepers   []string // for seletcting during init phase
	tempProviders []string

	sessionID uuid.UUID

	upKeepingItem *role.UpKeepingItem
	queryItem     *role.QueryItem
}

func newGroup(uid, shareTo, sk string, duration, capacity, price int64, ks, ps int, d data.Service) *groupInfo {
	return &groupInfo{
		userID:      uid,
		rootID:      uid,
		shareToID:   shareTo,
		privKey:     sk,
		ds:          d,
		state:       stoped,
		storeDays:   duration,
		storeSize:   capacity,
		storePrice:  price,
		stPayCycle:  utils.DefaultCycle,
		keeperSLA:   ks,
		providerSLA: ps,
		keepers:     make(map[string]*keeperInfo, ks),
		providers:   make(map[string]*providerInfo, ps),
	}
}

// startGroupService starts group
// step1: broadcast init message(query address) to keeper
// step2: handle init message from keeper
// step3: sync send notify to keeper and handle keeper's notify
// step4: deploy upkeeping contract and channel contracts(need modify)
func (g *groupInfo) start(ctx context.Context) (bool, error) {
	// getUK
	if g.upKeepingItem != nil {
		uItem := g.upKeepingItem
		utils.MLogger.Info("start user: ", g.userID, " and its lfs: ", g.groupID)
		g.keeperSLA = int(uItem.KeeperSLA)
		g.providerSLA = int(uItem.ProviderSLA)

		var keepers []string
		var providers []string
		for _, keeper := range uItem.Keepers {
			kid, err := address.GetIDFromAddress(keeper.Addr.String())
			if err != nil {
				return false, err
			}
			keepers = append(keepers, kid)
		}

		for _, provider := range uItem.Providers {
			pid, err := address.GetIDFromAddress(provider.Addr.String())
			if err != nil {
				return false, err
			}
			providers = append(providers, pid)
		}

		g.tempKeepers = keepers
		g.tempProviders = providers
		g.storeDays = uItem.Duration / (24 * 60 * 60)
		g.storePrice = uItem.Price
		g.storeSize = uItem.Capacity
		g.stPayCycle = uItem.Cycle
		g.state = depoyDone
		err := g.connect(ctx)
		if err != nil {
			return true, err
		}
		return true, nil
	}

	if g.userID == g.groupID {
		kmUser, err := metainfo.NewKey(g.groupID, mpb.KeyType_LFS, g.userID)
		if err != nil {
			return false, err
		}

		res, err := g.ds.GetKey(ctx, kmUser.ToString(), "local")
		if err == nil {
			utils.MLogger.Info("Test user: ", g.userID, " has keepers and providers: ", string(res))

			splitedMeta := strings.Split(string(res), metainfo.DELIMITER)
			if len(splitedMeta) == 2 {
				count := 0
				keepers := splitedMeta[0]
				for i := 0; i < len(keepers)/utils.IDLength; i++ {
					kid := keepers[i*utils.IDLength : (i+1)*utils.IDLength]
					_, err := peer.IDB58Decode(kid)
					if err != nil {
						continue
					}

					if !utils.CheckDup(g.tempKeepers, kid) {
						continue
					}

					g.tempKeepers = append(g.tempKeepers, kid)
					count++
				}

				g.keeperSLA = count

				count = 0
				providers := splitedMeta[1]
				for i := 0; i < len(providers)/utils.IDLength; i++ {
					pid := providers[i*utils.IDLength : (i+1)*utils.IDLength]

					_, err := peer.IDB58Decode(pid)
					if err != nil {
						continue
					}

					if !utils.CheckDup(g.tempProviders, pid) {
						continue
					}

					g.tempProviders = append(g.tempProviders, pid)
					count++
				}

				g.providerSLA = count

				utils.MLogger.Info("Start test user: ", g.userID, " and its lfs:", g.groupID)

				g.state = depoyDone
				err = g.connect(ctx)
				if err != nil {
					return true, err
				}
				return true, nil
			}
		}
	}

	utils.MLogger.Info("Initialize user:", g.userID, " and its lfs: ", g.groupID)
	err := g.initGroup(ctx)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (g *groupInfo) connect(ctx context.Context) error {
	if g.state != depoyDone {
		return role.ErrWrongState
	}

	g.Lock()
	defer g.Unlock()

	utils.MLogger.Info("Connect keepers and providers for user: ", g.userID)
	var wg sync.WaitGroup
	for _, kid := range g.tempKeepers {
		tempKeeper := &keeperInfo{
			keeperID: kid,
		}
		g.keepers[kid] = tempKeeper
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			g.ds.Connect(ctx, pid)
		}(kid)
	}

	for _, pid := range g.tempProviders {
		tempPro := &providerInfo{
			providerID: pid,
		}
		g.providers[pid] = tempPro
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			g.ds.Connect(ctx, pid)
		}(pid)
	}

	wg.Wait()

	connectTryCount := 5
	failNum := 0
	for i := 0; i < connectTryCount; i++ {
		failNum = 0
		for _, kinfo := range g.keepers {
			if !g.ds.Connect(ctx, kinfo.keeperID) {
				failNum++
				kinfo.connected = false
				if i == connectTryCount-1 {
					utils.MLogger.Warn("Connect to keeper: ", kinfo.keeperID, " failed.")
				}
				time.Sleep(time.Minute)
			} else {
				kinfo.connected = true
			}
		}

		if failNum > 0 {
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	// all fails
	if failNum == g.keeperSLA {
		return ErrNoEnoughKeeper
	}

	failNum = 0
	for i := 0; i < connectTryCount; i++ {
		failNum = 0
		for _, pinfo := range g.providers {
			if !g.ds.Connect(ctx, pinfo.providerID) {
				failNum++
				pinfo.connected = false
				if i == connectTryCount-1 {
					utils.MLogger.Warn("Connect to provider: ", pinfo.providerID, " failed.")
				}
			} else {
				pinfo.connected = true
			}
		}

		if failNum > 0 {
			time.Sleep(time.Minute)
			continue
		}
		break
	}

	// all fails
	if failNum == g.providerSLA {
		return ErrNoEnoughProvider
	}

	// send pubKey to all kps
	pubKey, err := id.GetCompressPubByte(g.privKey)
	if err != nil {
		utils.MLogger.Error("Get publickey for: ", g.shareToID, " fail: ", err)
		return err
	}

	kmp, err := metainfo.NewKey(g.shareToID, mpb.KeyType_PublicKey)
	if err != nil {
		return err
	}

	g.putToAll(ctx, kmp.ToString(), pubKey)

	newID := uuid.New()
	g.sessionID = newID

	force := "0"
	if g.force {
		force = "1"
	}

	// key: queryID/"UserStart"/userID/kc/pc/id
	kmc, err := metainfo.NewKey(g.groupID, mpb.KeyType_UserStart, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA), newID.String(), force)
	if err != nil {
		return err
	}

	var res strings.Builder
	for _, keeper := range g.keepers {
		res.WriteString(keeper.keeperID)
	}

	res.WriteString(metainfo.DELIMITER)

	for _, provider := range g.providers {
		res.WriteString(provider.providerID)
	}

	kms := kmc.ToString()
	val := []byte(res.String())

	sig, err := id.SignForKey(g.privKey, kms, val)
	if err != nil {
		return err
	}

	if role.Debug {
		ok := g.ds.VerifyKey(ctx, kms, val, sig)
		if !ok {
			utils.MLogger.Errorf("key signature is wrong for %s", kms)
			return nil
		}
	}

	for _, kinfo := range g.keepers {
		kinfo.sessionID = newID
		resp, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kms, val, sig, kinfo.keeperID)
		if err != nil {
			utils.MLogger.Warn("Send keeper: ", kinfo.keeperID, " err: ", err)
			continue
		}

		uuidtmp, err := uuid.ParseBytes(resp)
		if err != nil {
			utils.MLogger.Warn("uuid ParseBytes: ", string(resp), "  err: ", err)
			continue
		}
		kinfo.sessionID = uuidtmp
	}

	for _, pinfo := range g.providers {
		pinfo.sessionID = newID
		resp, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kms, val, sig, pinfo.providerID)
		if err != nil {
			utils.MLogger.Warn("Send provider: ", pinfo.providerID, "  err: ", err)
		}

		uuidtmp, err := uuid.ParseBytes(resp)
		if err != nil {
			utils.MLogger.Warn("uuid ParseBytes: ", string(resp), "  err: ", err)
			continue
		}
		pinfo.sessionID = uuidtmp
	}

	utils.MLogger.Info("Group Service is ready for: ", g.userID)

	g.loadContracts(ctx, "")

	g.state = groupStarted
	return nil
}

// user init
// key: queryID/"UserInit"/userID/keeperCount/providerCount
// for test: queryID = userID
func (g *groupInfo) initGroup(ctx context.Context) error {
	//构造init信息并发送 此时，初始化阶段为collecting
	kmInit, err := metainfo.NewKey(g.groupID, mpb.KeyType_UserInit, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA))
	if err != nil {
		return err
	}

	kmes := kmInit.ToString()

	g.state = collecting
	go g.ds.BroadcastMessage(ctx, kmes)

	// wait 20 minutes for collecting
	timeOutCount := 0
	tick := time.Tick(60 * time.Second)
	for {
		select {
		case <-tick:
			if timeOutCount >= 30 {
				return role.ErrTimeOut
			}
			switch g.state {
			case collecting:
				timeOutCount++
				ok := g.collect(ctx)
				if ok {
					go g.ds.BroadcastMessage(ctx, kmes)
				}
			case collectDone:
				g.notify(ctx)
			case deploying:
				g.deployContract(ctx)
			case depoyDone:
				return g.connect(ctx)
			default:
				return nil
			}
		case <-ctx.Done():
			return role.ErrCancel
		}
	}
}

// handleUserInit handle replys from keepers
// key: queryID/"UserInit"/userID/keepercount/providercount,
// value: kid1kid2..../pid1pid2
func (g *groupInfo) handleUserInit(ctx context.Context, km *metainfo.Key, metaValue []byte, from string) {
	g.Lock()
	defer g.Unlock()

	if g.state != collecting {
		return
	}

	utils.MLogger.Info("Receive InitResponse，from：", from, ", value is：", string(metaValue))
	splitedMeta := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitedMeta) != 2 {
		return
	}

	kcount := 0
	pcount := 0
	keepers := splitedMeta[0]
	for i := 0; i < len(keepers)/utils.IDLength; i++ {
		kid := keepers[i*utils.IDLength : (i+1)*utils.IDLength]
		_, err := peer.IDB58Decode(kid)
		if err != nil {
			continue
		}
		if !utils.CheckDup(g.tempKeepers, kid) {
			continue
		}

		if g.ds.Connect(ctx, kid) {
			g.tempKeepers = append(g.tempKeepers, kid)
			kcount++
		}
	}

	providers := splitedMeta[1]
	for i := 0; i < len(providers)/utils.IDLength; i++ {
		pid := providers[i*utils.IDLength : (i+1)*utils.IDLength]
		if !utils.CheckDup(g.tempProviders, pid) {
			continue
		}

		if g.ds.Connect(ctx, pid) {
			g.tempProviders = append(g.tempProviders, pid)
			pcount++
		}
	}

	if kcount >= g.keeperSLA && pcount >= g.providerSLA {
		g.state = collectDone
	}
}

func (g *groupInfo) collect(ctx context.Context) bool {
	g.Lock()
	defer g.Unlock()

	if g.state != collecting {
		return false
	}

	kcount := 0
	pcount := 0

	for _, kid := range g.tempKeepers {
		if g.ds.Connect(ctx, kid) {
			kcount++
		}
	}

	for _, kid := range g.tempProviders {
		if g.ds.Connect(ctx, kid) {
			pcount++
		}
	}

	if kcount >= g.keeperSLA && pcount >= g.providerSLA {
		g.state = collectDone
		return false
	}
	utils.MLogger.Infof("No enough keepers and providers, have k:%d p:%d, want k:%d p:%d, collecting...", kcount, pcount, g.keeperSLA, g.providerSLA)
	return true
}

// key: queryID/"UserNotify"/userID/kc/pc
func (g *groupInfo) notify(ctx context.Context) {
	// in case other change temp
	g.Lock()

	if g.state != collectDone {
		g.Unlock()
		return
	}

	utils.MLogger.Info("Has enough Keeper and Providers, choosing...")
	keepers := make([]string, 0, g.keeperSLA)
	providers := make([]string, 0, g.providerSLA)
	g.tempKeepers = utils.DisorderArray(g.tempKeepers)
	i := 0
	for _, kidStr := range g.tempKeepers {
		if i >= g.keeperSLA {
			break
		}

		if !g.ds.Connect(ctx, kidStr) {
			continue
		}
		i++
		keepers = append(keepers, kidStr)
	}

	if len(keepers) < g.keeperSLA {
		g.state = collecting
		g.Unlock()
		utils.MLogger.Info("Keeper is not enough, collecting...")
		return
	}

	g.tempProviders = utils.DisorderArray(g.tempProviders)
	i = 0
	for _, pidStr := range g.tempProviders {
		if i >= g.providerSLA {
			break
		}

		if !g.ds.Connect(ctx, pidStr) {
			continue
		}

		i++
		providers = append(providers, pidStr)
	}

	if len(providers) < g.providerSLA {
		g.state = collecting
		g.Unlock()
		utils.MLogger.Info("Provider is not enough, collecting...")
		return
	}

	g.tempKeepers = keepers
	g.tempProviders = providers

	utils.MLogger.Info("Choose completed")

	var res strings.Builder
	for _, kid := range g.tempKeepers {
		res.WriteString(kid)
	}

	res.WriteString(metainfo.DELIMITER)

	for _, pid := range g.tempProviders {
		res.WriteString(pid)
	}

	g.Unlock()

	kmNotify, err := metainfo.NewKey(g.groupID, mpb.KeyType_UserNotify, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA))
	if err != nil {
		return
	}

	kmes := kmNotify.ToString()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		count := 0
		for _, kid := range keepers { //循环发消息
			wg.Add(1)
			utils.MLogger.Info("Notify keeper: ", kid)
			go func(kid string) {
				defer wg.Done()
				retry := 0
				// retry
				for retry < 10 {
					res, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kmes, []byte(res.String()), nil, kid)
					if err != nil || string(res) != "ok" {
						retry++
						time.Sleep(30 * time.Second)
					} else {
						g.Lock()
						count++
						g.Unlock()
						return
					}
				}

			}(kid)

		}
		wg.Wait()

		//all keepers are online
		if count == g.keeperSLA {
			utils.MLogger.Info("Receive all keepers' response")
			g.Lock()
			g.state = deploying
			g.Unlock()
			return
		}
	}

	// re-collecting
	g.Lock()
	g.state = collecting
	g.Unlock()
}

func (g *groupInfo) deployContract(ctx context.Context) error {
	g.Lock()
	defer g.Unlock()

	if g.state != deploying {
		return role.ErrWrongState
	}

	var res strings.Builder
	// sort for signs
	sort.Strings(g.tempKeepers)
	for _, kid := range g.tempKeepers {
		res.WriteString(kid)
	}

	res.WriteString(metainfo.DELIMITER)

	sort.Strings(g.tempProviders)
	for _, pid := range g.tempProviders {
		res.WriteString(pid)
	}

	if g.userID != g.groupID {
		ukID, err := role.DeployUpKeeping(g.userID, g.groupID, g.privKey, g.tempKeepers, g.tempProviders, g.storeDays, g.storeSize, g.storePrice, g.stPayCycle, true)
		if err != nil {
			utils.MLogger.Error("Deploy UpKeeping failed: ", err)
		}

		uItem, err := role.GetUpkeepingInfo(g.userID, ukID)
		if err != nil {
			utils.MLogger.Error("Get UpKeeping failed: ", err)
		}

		g.upKeepingItem = &uItem

		rootID, err := role.DeployRoot(g.privKey, g.userID, g.groupID, true)
		if err != nil {
			utils.MLogger.Error("Deploy root contract failed: ", err)
		}

		g.rootID = rootID

		var wg sync.WaitGroup
		for _, proID := range g.tempProviders {
			wg.Add(1)
			go func(proID string) {
				defer wg.Done()
				for i := 0; i < 5; i++ {
					tdelay := rand.Int63n(int64(i+1) * 60000000000)
					time.Sleep(time.Duration(tdelay))
					_, err := role.DeployChannel(g.userID, g.groupID, proID, g.privKey, g.storeDays, g.storeSize, true)
					if err != nil {
						continue
					}
					utils.MLogger.Infof("deploy channel contract for %s success", proID)
				}
				// need persist
			}(proID)
		}
		wg.Wait()

		balance, err := role.QueryBalance(g.userID)
		if err == nil {
			utils.MLogger.Infof("%s has balance: %s", g.userID, balance)
		}
	} else {
		kmUser, err := metainfo.NewKey(g.groupID, mpb.KeyType_LFS, g.userID)
		if err != nil {
			return err
		}

		g.ds.PutKey(ctx, kmUser.ToString(), []byte(res.String()), nil, "local")
	}
	g.state = depoyDone

	return nil
}

func (g *groupInfo) stop(ctx context.Context) error {
	if g.state != groupStarted {
		return role.ErrWrongState
	}

	g.Lock()
	defer g.Unlock()

	utils.MLogger.Info("Stop user: ", g.userID)

	// key: queryID/"UserStop"/userID/kc/pc/id
	kmc, err := metainfo.NewKey(g.groupID, mpb.KeyType_UserStop, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA), g.sessionID.String())
	if err != nil {
		return err
	}

	for _, kinfo := range g.keepers {
		_, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kmc.ToString(), nil, nil, kinfo.keeperID)
		if err != nil {
			utils.MLogger.Warn("Send keeper: ", kinfo.keeperID, " err: ", err)
			continue
		}
	}

	for _, pinfo := range g.providers {
		_, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kmc.ToString(), nil, nil, pinfo.providerID)
		if err != nil {
			utils.MLogger.Warn("Send provider: ", pinfo.providerID, "  err: ", err)
		}
	}

	utils.MLogger.Info("Group Service is stop for: ", g.userID)

	g.state = stoped
	return nil
}

func (g *groupInfo) heartbeat(ctx context.Context) error {
	if g.state != groupStarted {
		return role.ErrWrongState
	}

	g.RLock()
	defer g.RUnlock()

	utils.MLogger.Info("Send heartbeat for user: ", g.userID)

	// key: queryID/"UserStart"/userID/kc/pc/id
	kmc, err := metainfo.NewKey(g.groupID, mpb.KeyType_HeartBeat, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA), g.sessionID.String())
	if err != nil {
		return err
	}

	for _, kid := range g.tempKeepers {
		res, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kmc.ToString(), nil, nil, kid)
		if err != nil {
			continue
		}

		uuidtmp, err := uuid.ParseBytes(res)
		if err != nil {
			utils.MLogger.Warn("uuid ParseBytes: ", string(res), " err: ", err)
			continue
		}

		if uuidtmp != uuid.Nil && g.sessionID != uuidtmp {
			return ErrLfsReadOnly
		}
	}

	for _, pid := range g.tempProviders {
		g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), kmc.ToString(), nil, nil, pid)
	}

	return nil
}

func (g *groupInfo) getBlockProviders(ctx context.Context, blockID string) (string, int, error) {
	var pidstr string
	var offset int

	kmBlock, err := metainfo.NewKey(blockID, mpb.KeyType_BlockPos)
	if err != nil {
		return "", 0, err
	}
	blockMeta := kmBlock.ToString()
	for _, kp := range g.tempKeepers {
		pidAndOffset, err := g.ds.GetKey(ctx, blockMeta, kp)
		if err != nil || pidAndOffset == nil {
			continue
		}
		//成功收到
		splitedValue := strings.Split(string(pidAndOffset), metainfo.DELIMITER)
		if len(splitedValue) < 2 {
			continue
		}
		pidstr = splitedValue[0]
		offset, err = strconv.Atoi(splitedValue[1])
		if err != nil {
			continue
		}

		return pidstr, offset, nil
	}
	return "", 0, ErrNoProviders
}

func (g *groupInfo) CheckKeepersConn(ctx context.Context) (int, error) {
	if g == nil {
		return 0, ErrLfsServiceNotReady
	}
	count := 0
	for _, kp := range g.tempKeepers {
		if g.ds.Connect(ctx, kp) { //连接不上此keeper
			count++
		}
	}
	return count, nil
}

func (g *groupInfo) GetKeepers(ctx context.Context, count int) ([]string, []string, error) {
	if g == nil {
		return nil, nil, ErrLfsServiceNotReady
	}
	num := count
	if count < 0 {
		num = len(g.tempKeepers)
	}

	unconKeepers := make([]string, 0, num)
	conKeepers := make([]string, 0, num)

	i := 0
	for _, kp := range g.tempKeepers {
		if i >= num {
			break
		}

		if !g.ds.Connect(ctx, kp) { //连接不上此keeper
			unconKeepers = append(unconKeepers, kp)
		} else {
			conKeepers = append(conKeepers, kp)
			i++
		}
	}

	if len(conKeepers) < num && count > 0 {
		return conKeepers, unconKeepers, ErrNoEnoughKeeper
	}

	return conKeepers, unconKeepers, nil
}

func (g *groupInfo) GetProviders(ctx context.Context, count int) ([]string, []string, error) {
	num := count
	if count < 0 {
		num = len(g.tempProviders)
	}

	i := 0

	unconPro := make([]string, 0, num)
	conPro := make([]string, 0, num)
	for _, pro := range g.tempProviders {
		if i >= num {
			break
		}

		if !g.ds.Connect(ctx, pro) { //连接不上此provider
			unconPro = append(unconPro, pro)
			continue
		} else {
			conPro = append(conPro, pro)
			i++
		}
	}

	if len(conPro) < num && count > 0 {
		return conPro, unconPro, ErrNoEnoughProvider
	}

	return conPro, unconPro, nil
}

func (g *groupInfo) putToAll(ctx context.Context, key string, value []byte) {
	g.ds.PutKey(ctx, key, value, nil, "local")

	sig, err := id.SignForKey(g.privKey, key, value)
	if err != nil {
		utils.MLogger.Error("sign for key %s fails: %s", key, err)
		return
	}

	var wg sync.WaitGroup
	//put config to
	for _, kid := range g.tempKeepers {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			retry := 0
			for retry < 10 {
				err := g.ds.PutKey(ctx, key, value, sig, pid)
				if err != nil {
					retry++
					if retry >= 10 {
						utils.MLogger.Warn("Put bls config to: ", pid, " failed: ", err)
					}
					time.Sleep(60 * time.Second)
				}
				break
			}
		}(kid)
	}

	for _, kid := range g.tempProviders {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			retry := 0
			for retry < 10 {
				err := g.ds.PutKey(ctx, key, value, sig, pid)
				if err != nil {
					retry++
					if retry >= 10 {
						utils.MLogger.Warn("Put bls config to: ", pid, " failed: ", err)
					}
					time.Sleep(60 * time.Second)
				}
				break
			}
		}(kid)
	}
	wg.Wait()
	return
}

func (g *groupInfo) putDataToKeepers(ctx context.Context, key string, value []byte) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}

	sig, err := id.SignForKey(g.privKey, key, value)
	if err != nil {
		utils.MLogger.Error("sign for key %s fails: %s", key, err)
		return err
	}

	var wg sync.WaitGroup

	count := int32(0)
	for _, keeper := range g.tempKeepers {
		wg.Add(1)
		go func(pid string) {
			defer wg.Done()
			i := 0
			for {
				_, err := g.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), key, value, sig, pid)
				if err != nil {
					i++
					if i == 5 {
						utils.MLogger.Error("Send meta message to: ", pid, " error : ", err)
						atomic.AddInt32(&count, 1)
						return
					}
					time.Sleep(30 * time.Second)
				}
				break
			}
		}(keeper)
	}

	wg.Wait()

	if int(count) == len(g.tempKeepers) {
		return ErrNoEnoughKeeper
	}

	return nil
}

func (g *groupInfo) putDataMetaToKeepers(ctx context.Context, blockID string, provider string, offset int) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	kmBlock, err := metainfo.NewKey(blockID, mpb.KeyType_BlockPos)
	if err != nil {
		return err
	}
	metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
	return g.putDataToKeepers(ctx, kmBlock.ToString(), []byte(metaValue))
}

//删除块
func (g *groupInfo) deleteBlocksFromProvider(ctx context.Context, blockID string, updateMeta bool) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	provider, _, err := g.getBlockProviders(ctx, blockID)
	if err == ErrNoProviders { //Noprovider说明此块还不存在，不用删除
		utils.MLogger.Warnf("Get block: %s's location error, no exist or keepers lost it.", blockID)
		return nil
	} else if err != nil {
		return err
	}

	km, err := metainfo.NewKey(blockID, mpb.KeyType_Block)
	if err != nil {
		return err
	}

	if updateMeta { //这个需要等待返回
		g.ds.DeleteBlock(ctx, km.ToString(), provider)
	} else {
		go g.ds.DeleteBlock(ctx, km.ToString(), provider)
	}

	// or sent by provider?
	km.KType = mpb.KeyType_BlockPos
	for _, kp := range g.tempKeepers {
		go g.ds.DeleteKey(ctx, km.ToString(), kp)
	}

	return nil
}

func (g *groupInfo) loadContracts(ctx context.Context, pid string) error {
	if g.groupID == g.userID {
		return nil
	}

	if g.queryItem == nil {
		qItem, err := role.GetQueryInfo(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.queryItem = &qItem
	}

	if g.upKeepingItem == nil {
		uItem, err := role.GetUpKeeping(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.upKeepingItem = &uItem
	}

	if g.rootID == g.userID {
		rID, err := role.GetRoot(g.userID, g.groupID)
		if err != nil {
			return err
		}
		g.rootID = rID
	}

	for _, proInfo := range g.providers {
		proID := proInfo.providerID

		if proInfo.offerItem == nil {
			oItem, err := role.GetLatestOffer(g.userID, proID)
			if err != nil {
				continue
			}
			proInfo.offerItem = &oItem
		}
	}

	var wg sync.WaitGroup
	for _, pInfo := range g.providers {
		wg.Add(1)
		go func(pinfo *providerInfo) {
			defer wg.Done()
			proID := pinfo.providerID
			cItem := pinfo.chanItem
			if cItem != nil {
				if pid == proID {
					cItem.Money = role.GetBalance(cItem.ChannelID)
				}
			} else {
				gotItem, err := role.GetLatestChannel(g.shareToID, g.groupID, proID)
				if err == nil {
					cItem = &gotItem
				}
			}

			if cItem != nil {
				if time.Now().Unix()-cItem.StartTime < cItem.Duration {
					if cItem.Money.Cmp(big.NewInt(0)) != 0 {
						km, err := metainfo.NewKey(cItem.ChannelID, mpb.KeyType_Channel)
						if err != nil {
							return
						}

						cSign := &mpb.ChannelSign{}
						valueByte, err := g.ds.GetKey(ctx, km.ToString(), "local")
						if err == nil && len(valueByte) > 0 {
							err = proto.Unmarshal(valueByte, cSign)
							if err == nil {
								ok := role.VerifyChannelSign(cSign)
								if ok {
									value := new(big.Int).SetBytes(cSign.GetValue())
									utils.MLogger.Info("channel value in local is:", value.String())
									pinfo.Lock()
									if value.Cmp(cItem.Value) > 0 {
										cItem.Value = value
										cItem.Sig = valueByte
									}
									pinfo.Unlock()
								}
							}
						}
						utils.MLogger.Info("try to get channel value from remote: ", proID)
						valueRemote, err := g.ds.GetKey(ctx, km.ToString(), proID)
						if err == nil {
							err = proto.Unmarshal(valueRemote, cSign)
							if err == nil {
								ok := role.VerifyChannelSign(cSign)
								if ok {
									value := new(big.Int).SetBytes(cSign.GetValue())
									utils.MLogger.Info("channel value from remote is:", value.String())
									pinfo.Lock()
									if value.Cmp(cItem.Value) > 0 {
										cItem.Value = value
										cItem.Sig = valueRemote
									}
									pinfo.Unlock()
								}
							}
						}

						g.ds.PutKey(ctx, km.ToString(), cItem.Sig, nil, "local")
						if cItem.Value.Cmp(cItem.Money) < 0 {
							pinfo.Lock()
							pinfo.chanItem = cItem
							pinfo.Unlock()
							return
						}
					}
				} else {
					err := role.KillChannel(cItem.ChannelID, g.privKey)
					if err != nil {
						utils.MLogger.Errorf("close channel %s fails: %s", cItem.ChannelID, err)
					}
				}
			}

			// need redeploy
			_, err := role.DeployChannel(g.shareToID, g.groupID, proID, g.privKey, g.storeDays, g.storeSize, true)
			if err != nil {
				return
			}

			gotItem, err := role.GetLatestChannel(g.shareToID, g.groupID, proID)
			if err != nil {
				utils.MLogger.Warn("got channel fails for: ", g.shareToID)
				return
			}

			pinfo.Lock()
			pinfo.chanItem = &gotItem
			pinfo.Unlock()
		}(pInfo)
	}
	wg.Wait()
	return nil
}

func (g *groupInfo) loadChannelValue(ctx context.Context) error {
	if g.groupID == g.userID {
		return nil
	}

	var wg sync.WaitGroup
	for _, pInfo := range g.providers {
		wg.Add(1)
		go func(proInfo *providerInfo) {
			defer wg.Done()
			if proInfo.chanItem != nil {
				proID := proInfo.providerID
				km, err := metainfo.NewKey(proInfo.chanItem.ChannelID, mpb.KeyType_Channel)
				if err != nil {
					return
				}

				valueByte, err := g.ds.GetKey(ctx, km.ToString(), "local")
				if err == nil && len(valueByte) > 0 {
					cSign := &mpb.ChannelSign{}
					err = proto.Unmarshal(valueByte, cSign)
					if err == nil {
						ok := role.VerifyChannelSign(cSign)
						if ok {
							value := new(big.Int).SetBytes(cSign.GetValue())
							utils.MLogger.Info("channel value in local is:", value.String())
							proInfo.Lock()
							if value.Cmp(proInfo.chanItem.Value) > 0 {
								proInfo.chanItem.Value = value
								proInfo.chanItem.Sig = valueByte
							}
							proInfo.Unlock()
						}
					}

					utils.MLogger.Info("try to get channel value from remote: ", proID)
					valueRemote, err := g.ds.GetKey(ctx, km.ToString(), proID)
					if err == nil {
						err = proto.Unmarshal(valueRemote, cSign)
						if err == nil {
							ok := role.VerifyChannelSign(cSign)
							if ok {
								value := new(big.Int).SetBytes(cSign.GetValue())
								utils.MLogger.Info("channel value from remote is:", value.String())

								proInfo.Lock()
								if value.Cmp(proInfo.chanItem.Value) > 0 {
									proInfo.chanItem.Value = value
									proInfo.chanItem.Sig = valueRemote
								}
								proInfo.Unlock()
							}
						}
					}
				}
			}
		}(pInfo)
	}

	wg.Wait()
	return nil
}

func (g *groupInfo) saveChannelValue(ctx context.Context) error {
	if g.groupID == g.userID {
		return nil
	}

	for _, proInfo := range g.providers {
		if proInfo.chanItem != nil && proInfo.chanItem.Sig != nil && proInfo.chanItem.Dirty {
			km, err := metainfo.NewKey(proInfo.chanItem.ChannelID, mpb.KeyType_Channel)
			if err != nil {
				continue
			}

			g.ds.PutKey(ctx, km.ToString(), proInfo.chanItem.Sig, nil, "local")
		}
	}

	return nil
}
