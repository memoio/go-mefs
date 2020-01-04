package user

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	starting     int8 = iota
	collecting        // broadcast UserInit
	collectDone       // notify keeper
	deploying         // deploy contracts
	depoyDone         // connect
	groupStarted      // done
)

//keeperInfo 此结构体记录Keeper的信息，存储Tendermint地址，让user也能访问链上数据
type keeperInfo struct {
	keeperID  string
	sessionID uuid.UUID // for user
	connected bool
}

type providerInfo struct {
	providerID string
	connected  bool
	sessionID  uuid.UUID // for user
	chanItem   *role.ChannelItem
	offerItem  *role.OfferItem
}

// group stores use's groupinfo
type groupInfo struct {
	sync.Mutex
	groupID string // query address
	userID  string // user address
	privKey string // utils.EthSkByteToEthString(getSk(userID))
	state   int8   // atomic?
	ds      data.Service

	keepers   map[string]*keeperInfo
	providers map[string]*providerInfo

	storeDays     int64 //表示部署合约时的存储数据时间，单位是“天”
	storeSize     int64 //表示部署合约时的存储数据大小，单位是“MB”
	storePrice    int64 //表示部署合约时的存储价格大小，单位是“wei”
	keeperSLA     int   //表示部署合约时的keeper参数，目前是keeper数量
	providerSLA   int   //表示部署合约时的provider参数，目前是provider数量
	count         int
	reDeploy      bool     //是否重新部署offer
	tempKeepers   []string // for seletcting during init phase
	tempProviders []string

	upKeepingItem *role.UpKeepingItem
	queryItem     *role.QueryItem
}

func newGroup(uid, sk string, duration, capacity, price int64, ks, ps int, redeploy bool, d data.Service) *groupInfo {
	return &groupInfo{
		userID:      uid,
		privKey:     sk,
		ds:          d,
		state:       starting,
		storeDays:   duration,
		storeSize:   capacity,
		storePrice:  price,
		keeperSLA:   ks,
		providerSLA: ps,
		count:       0,
		reDeploy:    redeploy,
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
		log.Println("start user:", g.userID, "'s lfs:", g.groupID)
		g.keeperSLA = int(uItem.KeeperSLA)
		g.providerSLA = int(uItem.ProviderSLA)
		g.tempKeepers = uItem.KeeperIDs
		g.tempProviders = uItem.ProviderIDs
		g.state = depoyDone
		err := g.connect(ctx)
		if err != nil {
			return true, err
		}
		return true, nil
	}

	log.Println("init user:", g.userID, "'s lfs:", g.groupID)
	err := g.initGroup(ctx)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (g *groupInfo) connect(ctx context.Context) error {
	if g.state != depoyDone {
		return errors.New("Wrong state")
	}

	g.Lock()
	defer g.Unlock()

	log.Println("Connect for user: ", g.userID)
	for _, kid := range g.tempKeepers {
		tempKeeper := &keeperInfo{
			keeperID: kid,
		}
		g.keepers[kid] = tempKeeper
	}

	connectTryCount := 5
	failNum := 0
	for i := 0; i < connectTryCount; i++ {
		failNum = 0
		for _, kinfo := range g.keepers {
			if !g.ds.Connect(ctx, kinfo.keeperID) {
				failNum++
				kinfo.connected = false
				if i == connectTryCount-1 {
					log.Println("Connect to keeper", kinfo.keeperID, "failed.")
				}
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

	for _, pid := range g.tempProviders {
		tempPro := &providerInfo{
			providerID: pid,
		}
		g.providers[pid] = tempPro
	}

	failNum = 0
	for i := 0; i < connectTryCount; i++ {
		failNum = 0
		for _, pinfo := range g.providers {
			if !g.ds.Connect(ctx, pinfo.providerID) {
				failNum++
				pinfo.connected = false
				if i == connectTryCount-1 {
					log.Println("Connect to provider", pinfo.providerID, "failed.")
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

	// key: queryID/"UserStart"/userID/kc/pc
	kmc, err := metainfo.NewKeyMeta(g.groupID, metainfo.UserStart, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA))
	if err != nil {
		log.Println("Construct Deployed key error", err)
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

	for _, kinfo := range g.keepers {
		resp, err := g.ds.SendMetaRequest(ctx, int32(metainfo.Put), kmc.ToString(), []byte(res.String()), nil, kinfo.keeperID)
		if err != nil {
			log.Println("Send keeper", kinfo.keeperID, " err:", err)
			continue
		}
		uuidtmp, err := uuid.FromBytes(resp)
		if err != nil {
			continue
		}
		kinfo.sessionID = uuidtmp
	}

	for _, pinfo := range g.providers {
		resp, err := g.ds.SendMetaRequest(ctx, int32(metainfo.Put), kmc.ToString(), []byte(res.String()), nil, pinfo.providerID)
		if err != nil {
			log.Println("Send provider", pinfo.providerID, " err:", err)
		}

		uuidtmp, err := uuid.FromBytes(resp)
		if err != nil {
			continue
		}
		pinfo.sessionID = uuidtmp
	}

	log.Println("Group Service is ready for: ", g.userID)

	g.state = groupStarted
	return nil
}

// user init
// key: queryID/"UserInit"/userID/keeperCount/providerCount
// for test: queryID = userID
func (g *groupInfo) initGroup(ctx context.Context) error {
	//构造init信息并发送 此时，初始化阶段为collecting
	kmInit, err := metainfo.NewKeyMeta(g.groupID, metainfo.UserInit, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA))
	if err != nil {
		log.Println("gp connect: NewKeyMeta error!")
		return err
	}

	kmes := kmInit.ToString()

	g.state = collecting
	go g.ds.BroadcastMessage(ctx, kmes)

	// wait 20 minutes for collecting
	timeOutCount := 0
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case <-tick:
			if timeOutCount >= 40 {
				return ErrTimeOut
			}
			switch g.state {
			case collecting:
				timeOutCount++
				log.Printf("No enough keepers and providers, have k:%d p:%d, want k:%d p:%d, collecting...\n", len(g.tempKeepers), len(g.tempProviders), g.keeperSLA, g.providerSLA)
				go g.ds.BroadcastMessage(ctx, kmes)
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
			return nil
		}
	}
}

// handleUserInit handle replys from keepers
// key: queryID/"UserInit"/userID/keepercount/providercount,
// value: kid1kid2..../pid1pid2
func (g *groupInfo) handleUserInit(km *metainfo.KeyMeta, metaValue []byte, from string) {
	if g.state != collecting {
		return
	}
	g.Lock()
	defer g.Unlock()

	log.Println("Receive InitResponse，from：", from, ", value is：", string(metaValue))
	splitedMeta := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitedMeta) != 2 {
		return
	}

	ctx := context.Background()
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
		}
	}

	if len(g.tempKeepers) >= g.keeperSLA && len(g.tempProviders) >= g.providerSLA {
		g.state = collectDone
	}
}

// key: queryID/"UserNotify"/userID/kc/pc
func (g *groupInfo) notify(ctx context.Context) {
	if g.state != collectDone {
		return
	}

	// in case other change temp
	g.Lock()

	log.Println("Has enough Keeper and Providers, choosing...")
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
		log.Println("Keeper is not enough, collecting...")
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
		log.Println("Provider is not enough, collecting...")
		return
	}

	g.tempKeepers = keepers
	g.tempProviders = providers

	log.Println("Choose completed")

	var res strings.Builder
	for _, kid := range g.tempKeepers {
		res.WriteString(kid)
	}

	res.WriteString(metainfo.DELIMITER)

	for _, pid := range g.tempProviders {
		res.WriteString(pid)
	}

	g.Unlock()

	kmNotify, err := metainfo.NewKeyMeta(g.groupID, metainfo.UserNotify, g.userID, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA))
	if err != nil {
		log.Println("gp notify: NewKeyMeta error!")
		return
	}

	kmes := kmNotify.ToString()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		g.Lock()
		g.count = 0
		g.Unlock()
		for _, kid := range keepers { //循环发消息
			wg.Add(1)
			log.Println("Notify keeper:", kid)
			go func(kid string) {
				defer wg.Done()
				retry := 0
				// retry
				for retry < 10 {
					res, err := g.ds.SendMetaRequest(ctx, int32(metainfo.Put), kmes, []byte(res.String()), nil, kid)
					if err != nil || string(res) != "ok" {
						retry++
						time.Sleep(30 * time.Second)
					} else {
						g.Lock()
						g.count++
						g.Unlock()
						return
					}
				}

			}(kid)

		}
		wg.Wait()

		//all keepers are online
		if g.count == g.keeperSLA {
			log.Println("Receive all keepers' response")
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
	if g.state != deploying {
		return errors.New("State is wrong")
	}

	if g.userID != g.groupID {
		ukID, err := role.DeployUpKeeping(g.userID, g.groupID, g.privKey, g.tempKeepers, g.tempProviders, g.storeDays, g.storeSize, g.storePrice, true)
		if err != nil {
			log.Println("deploy UpKeeping failed :", err)
		}

		uItem, err := role.GetUpkeepingInfo(g.userID, ukID)
		if err != nil {
			log.Println("get UpKeeping failed :", err)
		}

		g.upKeepingItem = &uItem

		var wg sync.WaitGroup
		for _, proID := range g.tempProviders {
			wg.Add(1)
			go func(proID string) {
				defer wg.Done()
				_, err := role.DeployChannel(g.userID, g.groupID, proID, g.privKey, g.storeDays, g.storeSize, true)
				if err != nil {
					return
				}
				// need persist
			}(proID)
		}
		wg.Wait()
	}
	g.state = depoyDone

	return nil
}

func (g *groupInfo) getBlockProviders(blockID string) (string, int, error) {
	var pidstr string
	var offset int

	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
	if err != nil {
		return "", 0, err
	}
	blockMeta := kmBlock.ToString()
	ctx := context.Background()
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
			log.Println("Offset decode error-", pidstr, err)
			continue
		}

		if !g.ds.Connect(ctx, pidstr) { //连接不上此provider
			log.Println("Cannot connect to provider-", pidstr)
			return pidstr, offset, ErrNoProviders
		}
		return pidstr, offset, nil
	}
	return "", 0, ErrNoProviders
}

func (g *groupInfo) GetKeepers(count int) ([]string, []string, error) {
	num := count
	if count < 0 {
		num = len(g.tempKeepers)
	}

	unconKeepers := make([]string, 0, num)
	conKeepers := make([]string, 0, num)

	ctx := context.Background()

	i := 0
	for _, kp := range g.tempKeepers {
		if i >= num {
			break
		}

		if !g.ds.Connect(ctx, kp) { //连接不上此keeper
			unconKeepers = append(unconKeepers, kp)
			continue
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

func (g *groupInfo) GetProviders(count int) ([]string, []string, error) {
	num := count
	if count < 0 {
		num = len(g.tempProviders)
	}

	i := 0

	unconPro := make([]string, 0, num)
	conPro := make([]string, 0, num)
	ctx := context.Background()
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

func (g *groupInfo) putDataToKeepers(key string, value []byte) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	ctx := context.Background()
	count := 0
	for _, keeper := range g.tempKeepers {
		_, err := g.ds.SendMetaRequest(ctx, int32(metainfo.Put), key, value, nil, keeper)
		if err != nil {
			log.Println("send metaMessage to ", keeper, " error :", err)
			count++
		}
	}
	if count == len(g.tempKeepers) {
		return ErrNoKeepers
	}
	return nil
}

func (g *groupInfo) putDataMetaToKeepers(blockID string, provider string, offset int) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
	if err != nil {
		log.Println("construct put blockMeta KV error :", err)
		return err
	}
	metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
	return g.putDataToKeepers(kmBlock.ToString(), []byte(metaValue))
}

//删除块
func (g *groupInfo) deleteBlocksFromProvider(blockID string, updateMeta bool) error {
	if g.state < groupStarted {
		return ErrLfsServiceNotReady
	}
	provider, _, err := g.getBlockProviders(blockID)
	if err == ErrNoProviders { //Noprovider说明此块还不存在，不用删除
		log.Printf("Get block:%s's location error, no exist or keepers lost it.\n", blockID)
		return nil
	} else if err != nil {
		return err
	}

	km, err := metainfo.NewKeyMeta(blockID, metainfo.Block)
	if err != nil {
		log.Println("construct delete block KV error :", err)
		return err
	}

	ctx := context.Background()

	if updateMeta { //这个需要等待返回
		err := g.ds.DeleteBlock(ctx, km.ToString(), provider)
		if err != nil {
			log.Println("Cannot delete Block-", blockID, err)
			return ErrCannotDeleteMetaBlock
		}
	} else {
		go g.ds.DeleteBlock(ctx, km.ToString(), provider)
	}

	// or sent by provider?
	km.SetDType(metainfo.BlockPos)
	for _, kp := range g.tempKeepers {
		go g.ds.DeleteKey(ctx, km.ToString(), kp)
	}

	return nil
}
