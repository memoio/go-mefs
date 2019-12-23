package user

import (
	"context"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	inet "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

//keeperInfo 此结构体记录Keeper的信息，存储Tendermint地址，让user也能访问链上数据
type keeperInfo struct {
	isBFT     bool //标识Keeper组采取的同步方法
	keeperID  string
	connected bool
}

type providerInfo struct {
	providerID string
	connected  bool
	chanItem   *contracts.ChannelItem
	offerItem  *contracts.OfferItem
}

type group interface {
	start(ctx context.Context) error
	connect(ctx context.Context) error
	// broadcast init information
	initGroup(ctx context.Context) error
	// notify keepers and providers
	notify(km *metainfo.KeyMeta)
	// confirm all keepers
	confirm(ctx context.Context)
	done(ctx context.Context)
}

// group stores use's groupinfo
type groupInfo struct {
	userID        string
	keepers       []*keeperInfo
	providers     []*providerInfo
	upKeepingItem *contracts.UpKeepingItem
	queryItem     *contracts.QueryItem

	initResMutex sync.Mutex //目前同一时间只回复一个Keeper避免冲突
	storeDays    int64      //表示部署合约时的存储数据时间，单位是“天”
	storeSize    int64      //表示部署合约时的存储数据大小，单位是“MB”
	storePrice   int64      //表示部署合约时的存储价格大小，单位是“wei”
	keeperSLA    int        //表示部署合约时的keeper参数，目前是keeper数量
	providerSLA  int        //表示部署合约时的provider参数，目前是provider数量
	reDeploy     bool       //是否重新部署offer

	tempKeepers   []string // for seletcting during init phase
	tempProviders []string
}

func newGroup(uid string, duration int64, capacity int64, price int64, ks int, ps int, redeploy bool) *groupInfo {
	return &groupInfo{
		userID:      uid,
		storeDays:   duration,
		storeSize:   capacity,
		storePrice:  price,
		keeperSLA:   ks,
		providerSLA: ps,
		reDeploy:    redeploy,
	}
}

// startGroupService starts group
// step1: deploy query contract
// step2: send init message(query address) to keeper
// step3: handle init message from keeper
// step4: sync send notify to keeper and hanle keeper's notify
// step5: init userconfig, deploy upkeeping contract and channel contracts(need modify)
func (g *groupInfo) start(ctx context.Context) (bool, error) {
	// getUK
	_, uk, err := contracts.GetUKFromResolver(uaddr)
	switch err {
	case nil: //部署过
		log.Println("begin to start user : ", g.userID)
		item, err := contracts.GetUpkeepingInfo(uaddr, uk)
		if err != nil {
			return err
		}
		// keeper数量、provider的数量应以合约约定为主
		g.keeperSLA = int(item.KeeperSLA)
		g.providerSLA = int(item.ProviderSLA)
		g.upKeepingItem = &item
		g.tempKeepers = item.KeeperIDs
		g.tempProviders = item.ProviderIDs
		err = g.connect(ctx)
		if err != nil {
			return false, err
		}
		return true, nil
	case contracts.ErrNotDeployedMapper, contracts.ErrNotDeployedUk: //没有部署过
		log.Println("begin to init user : ", g.userID)
		err := gp.initGroup(ctx)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return false, nil
}

func (g *groupInfo) connect(ctx context.Context) error {
	log.Println("Connect keepers and providers for use: ", g.userID)

	err = testConnect()
	if err != nil {
		return err
	}

	connectTryCount := 5
	for i := 0; i < connectTryCount; i++ {
		var unsuccess []string
		for _, kid := range g.tempKeepers {
			// 连接失败加入unsuccess
			if !sc.ConnectTo(ctx, localNode, kid) {
				unsuccess = append(unsuccess, kid)
				log.Println("Connect to keeper", kid, "failed.")
				continue
			}
			tempKeeper := &keeperInfo{
				keeperID:  kid,
				connected: true,
			}
			kmKid, err := metainfo.NewKeyMeta(g.userID, metainfo.Local, metainfo.SyncTypeKid)
			if err != nil {
				return err
			}
			res, err := getKeyFrom(kmKid.ToString(), kid)
			if err == nil && res != nil {
				resStr := string(res)
				splitRes := strings.Split(resStr, metainfo.DELIMITER)
				if len(splitRes) > 3 {
					if splitRes[1] == metainfo.SyncTypeBft {
						tempKeeper.isBFT = true
					}
				}
			}
			// 检查该keeper是否已添加
			exist := false
			for _, kinfo := range g.keepers {
				if kid == kinfo.keeperID {
					exist = true
					break
				}
			}
			if !exist {
				g.keepers = append(g.keepers, tempKeeper)
			}
			if tempKeeper.isBFT {
				log.Println("Connect to keeper", tempKeeper.keeperID, "use bft mode")
			} else {
				log.Println("Connect to keeper", tempKeeper.keeperID, "use simple mode")
			}
		}

		if len(g.keepers) == g.keeperSLA {
			break
		}

		keepers = unsuccess
		// 每一个keeper连接后的判断，若连接数量足够后就立即进行而不是全部连接后再进行
		if len(g.keepers) < KeeperSLA {
			time.Sleep(time.Minute)
		}
	}

	if len(g.keepers) <= 0 {
		return ErrNoEnoughKeeper
	}

	for i := 0; i < connectTryCount; i++ {
		for _, pid := range g.tempProviders {
			if sc.ConnectTo(ctx, localNode, pid) {
				log.Println("Connect to provider-", pid, "success.")
			} else {
				log.Println("Connect to provider-", pid, "failed.")
			}

			tempP := &providerInfo{
				providerID: pid,
			}

			exist := false
			for _, kinfo := range g.providers {
				if pid == kinfo.providerID {
					exist = true
					break
				}
			}

			if !exist {
				g.providers = append(g.providers, tempP)
			}
		}

		if len(g.providers) == g.providerSLA {
			break
		}

		if len(g.providers) <= 0 {
			time.Sleep(time.Minute)
			continue
		}
	}

	if len(g.providers) <= 0 {
		return ErrNoEnoughProvider
	}

	// 构造key告诉keeper和provider自己已经启动
	kmPid, err := metainfo.NewKeyMeta(gp.userID, metainfo.UserDeployedContracts)
	if err != nil {
		log.Println("Construct Deployed key error", err)
		return err
	}
	for _, keeper := range g.keepers {
		_, err = sendMetaRequest(kmPid, g.userID, keeper.keeperID)
		if err != nil {
			log.Println("Send keeper", keeper.keeperID, " err:", err)
		}
	}
	for _, provider := range g.providers {
		_, err = sendMetaRequest(kmPid, g.userID, provider.providerID)
		if err != nil {
			log.Println("Send provider", provider, " err:", err)
		}
	}
	log.Println(g.userID + ":Group Service is ready")

	err = setState(g.userID, groupStarted)
	if err != nil {
		return err
	}
	return nil
}

// user init
func (g *groupInfo) initGroup(ctx context.Context) error {
	addr, err := address.GetAddressFromID(uid)
	if err != nil {
		return err
	}
	// getbalance
	balance, err := contracts.QueryBalance(addr.Hex())
	if err != nil {
		return err
	}
	log.Println(addr.String(), " has balance (wei): ", balance)

	//判断账户余额能否部署query合约 + upKeeping + channel合约
	var moneyPerDay = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(g.storePrice), big.NewInt(g.storeSize))
	var moneyAccount = new(big.Int)
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(g.storeDays))

	deployPrice := big.NewInt(int64(740621000000000))
	deployPrice.Add(big.NewInt(1128277), big.NewInt(int64(652346*gp.providerSLA)))
	var leastMoney = new(big.Int)
	leastMoney = leastMoney.Add(moneyAccount, deployPrice)
	if balance.Cmp(leastMoney) < 0 { //余额不足
		log.Println(addr.String(), " need more balance to start")
		return ErrBalance
	}

	// deploy query

	sk := utils.EthSkByteToEthString(getSk(g.userID))

	queryAddr, err = contracts.DeployQuery(addr, sk, g.storeSize, g.storeDays, g.storePrice, g.keeperSLA, g.providerSLA, gp.reDeploy)
	if err != nil {
		log.Println("fail to deploy query contract")
		return err
	}

	log.Println("Begin to find keepers for init...", "\nthe uid is: "+g.userID+"\nthe addr is: "+addr.String())

	err = testConnect()
	if err != nil {
		return err
	}

	//构造init信息并发送 此时，初始化阶段为collecting
	kmInit, err := metainfo.NewKeyMeta(g.userID, metainfo.UserInitReq, strconv.Itoa(g.keeperSLA), strconv.Itoa(g.providerSLA), queryAddr.String())
	if err != nil {
		log.Println("gp connect: NewKeyMeta error!")
		return err
	}
	go broadcastMetaMessage(kmInit, "")
	err = setState(g.userID, collecting)
	if err != nil {
		return err
	}
	// wait 20 minutes for collecting
	timeOutCount := 0
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case <-tick: //每过30s 检查是否收到了足够的KP信息，如果不足，继续发送初始化请求，足够的时候进行KP的选择和确认
			if timeOutCount >= 40 {
				return ErrTimeOut
			}
			userState, err := getState(g.userID)
			if err != nil {
				return err
			}
			switch userState {
			case collecting:
				timeOutCount++
				if len(g.tempKeepers) >= g.keeperSLA && len(g.tempProviders) >= g.providerSLA {
					//收集到足够的keeper和Provider 进行挑选并给keeper发送确认信息，初始化阶段变为collectComplete
					err := setState(g.userID, collectCompleted)
					if err != nil {
						return err
					}
					g.notify(kmInit)
				} else {
					log.Printf("Timeout, No enough keepers and Providers,Have k:%d p:%d,want k:%d p:%d, retrying...\n", len(gp.tempKeepers), len(gp.tempProviders), gp.keeperSLA, gp.providerSLA)
					go broadcastMetaMessage(kmInit, "")
				}
			case collectCompleted:
				timeOutCount++
				//TODO：等待keeper的第四次握手超时怎么办，目前继续等待
				log.Printf("Timeout, waiting keeper response\n")
				for _, keeperInfo := range gp.keepers {
					if !keeperInfo.connected {
						log.Printf("Keeper %s not response, waiting...", keeperInfo.keeperID)
					}
				}
			case onDeploy:
				g.handleRest()
			default:
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

//userInitNotIf 收集齐KP信息之后， 选择keeper和provider，构造确认信息发给keeper
func (g *groupInfo) notify(km *metainfo.KeyMeta) {
	userState, err := getState(g.userID)
	if err != nil {
		return err
	}
	if userState != collectCompleted {
		return ErrWrongInitState
	}
	log.Println("Has enough Keeper and Providers, choosing...")
	g.keepers = make([]*keeperInfo, 0, g.keeperSLA)
	g.providers = make([]*providerInfo, 0, g.providerSLA)
	//选择keeper
	g.tempKeepers = utils.DisorderArray(g.tempKeepers)
	i := 0
	for _, kidStr := range g.tempKeepers {
		if i >= g.keeperSLA {
			break
		}
		kidStr := g.tempKeepers[i]
		kidB58, _ := peer.IDB58Decode(kidStr)
		//判断是否链接，如果连不上，则从备选中删除，看下一个
		if localNode.PeerHost.Network().Connectedness(kidB58) != inet.Connected {
			if !sc.ConnectTo(context.Background(), localNode, kidStr) {
				continue
			}
		}
		tempK := &keeperInfo{
			keeperID:  kidStr,
			connected: false, // set true when receive notify response
		}
		i++
		g.keepers = append(g.keepers, tempK)
	}
	if len(g.keepers) < g.keeperSLA {
		err := setState(g.userID, collecting)
		if err != nil {
			log.Println("userInitNotif()setState()err:", err)
		}
		return
	}

	//选择provider
	g.tempProviders = utils.DisorderArray(g.tempProviders)
	i = 0
	for _, pidStr := range g.tempProviders {
		if i >= g.providerSLA {
			break
		}
		pidStr := g.tempProviders[i]
		pidB58, _ := peer.IDB58Decode(pidStr)
		//判断是否链接，如果连不上，则从备选中删除，看下一个
		if localNode.PeerHost.Network().Connectedness(pidB58) != inet.Connected {
			if !sc.ConnectTo(context.Background(), localNode, pidStr) {
				continue
			}
		}
		i++

		tempP := &providerInfo{
			providerID: pidStr,
			connected:  true,
		}

		g.providers = append(g.providers, tempP)
	}

	if len(g.providers) < g.providerSLA {
		err := setState(g.userID, collecting)
		if err != nil {
			log.Println("userInitNotif()setState()err:", err)
		}
		return
	}

	log.Println("Choose completed")

	//构造本节点keeper信息和provider信息 放入硬盘 id1id2id3.......(无分隔符)
	var assignedKeeper, assignedProvider string
	for _, keeper := range g.keepers {
		assignedKeeper += keeper.keeperID
	}

	for _, provider := range g.providers {
		assignedProvider += provider.providerID
	}
	//构造发给keeper的初始化确认信息并发送给自己的所有keeper
	assignedKP := assignedKeeper + metainfo.DELIMITER + assignedProvider
	km.SetKeyType(metainfo.UserInitNotif)

	var wg sync.WaitGroup
	for _, keeper := range g.keepers { //循环发消息
		wg.Add(1)
		log.Println("Notify keeper:", keeper.keeperID)
		go func(keeper string) {
			defer wg.Done()
			retry := 0
			// retry
			for retry < 10 {
				res, err := sendMetaRequest(km, assignedKP, keeper) //发送确认信息
				if err != nil {
					retry++
					time.Sleep(30 * time.Second)
				} else {
					g.confirm(res, keeper)
					return
				}
			}

		}(keeper.keeperID)

	}
	wg.Wait()

	log.Println("Waiting for keepers' response")
}

// confirm 第四次握手 确认Keeper启动完毕
// PeerID/"user_init_notif_res"/"bft","simple"或 IP:p2pport/IP:rpcport
func (g *groupInfo) confirm(keeper string, initRes string) {
	g.initResMutex.Lock()
	defer g.initResMutex.Unlock()
	splitInitRes := strings.Split(initRes, metainfo.DELIMITER)
	var count int
	//将发来信息的keeper记录为连接成功
	for _, kp := range g.keepers {
		if strings.Compare(kp.keeperID, keeper) == 0 && !kp.connected {
			if len(splitInitRes) >= 2 {
				kp.isBFT = true
			}
			kp.connected = true
			log.Printf("Receive %s's response, waiting for other keepers\n", kp.keeperID)
		}
		if kp.connected {
			count++
		}
	}
	//与所有keeper都连接成功了
	if count == g.keeperSLA {
		log.Println("Receive all keepers' response")
		err := localNode.Repo.SetConfigKey("IsInit", false) //初始化过后，设置IsInit
		if err != nil {
			log.Println("gp.localNode.Repo.SetConfigKey failed :", err)
		}

		// 状态改为部署合约中
		err = setState(gp.userid, onDeploy)
		if err != nil {
			log.Println("setState failed :", err)
			return
		}

	}
	return
}

func (g *groupInfo) done() error {

	//部署合约
	err = deployUpKeepingAndChannel()
	if err != nil {
		log.Println("deployUpKeepingAndChannel failed :", err)
	}

	saveContracts(g.userID)

	// 构造key告诉keeper和provider自己已经部署好合约
	kmPid, err := metainfo.NewKeyMeta(g.userID, metainfo.UserDeployedContracts)
	if err != nil {
		log.Println("Construct Deployed key error", err)
		return err
	}
	for _, keeper := range gp.keepers {
		_, err = sendMetaRequest(kmPid, gp.userid, keeper.keeperID)
		if err != nil {
			log.Println("Send keeper", keeper.keeperID, " err:", err)
		}
	}
	for _, provider := range gp.providers {
		_, err = sendMetaRequest(kmPid, gp.userid, provider.providerID)
		if err != nil {
			log.Println("Send provider", provider, " err:", err)
		}
	}
	log.Println(gp.userID + ": group is ready")
	err = setState(gp.userID, groupStarted)
	if err != nil {
		log.Println("setState failed :", err)
		return err
	}
	return nil
}
