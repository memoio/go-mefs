package user

import (
	"context"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	inet "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	mcl "github.com/memoio/go-mefs/bls12"
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

// groupService stores use's groupinfo
type groupService struct {
	userid        string
	keepers       []*keeperInfo
	providers     []*providerInfo
	upKeepingItem *contracts.UpKeepingItem
	queryItem     *contracts.QueryItem
	keySet        *mcl.KeySet
	initResMutex  sync.Mutex //目前同一时间只回复一个Keeper避免冲突
	privateKey    []byte
	storeDays     int64 //表示部署合约时的存储数据时间，单位是“天”
	storeSize     int64 //表示部署合约时的存储数据大小，单位是“MB”
	storePrice    int64 //表示部署合约时的存储价格大小，单位是“wei”
	keeperSLA     int   //表示部署合约时的keeper参数，目前是keeper数量
	providerSLA   int   //表示部署合约时的provider参数，目前是provider数量
	reDeploy      bool  //是否重新部署offer

	tempKeepers   []string // for seletcting during init phase
	tempProviders []string
}

// constructGroupService Constructs groupService
func constructGroupService(userid string, privKey []byte, duration int64, capacity int64, price int64, ks int, ps int, redeploy bool) *groupService {
	if privKey == nil {
		return nil
	}
	return &groupService{
		userid:      userid,
		privateKey:  privKey,
		storeDays:   duration,
		storeSize:   capacity,
		storePrice:  price,
		keeperSLA:   ks,
		providerSLA: ps,
		reDeploy:    redeploy,
	}
}

// startGroupService starts group
func (gp *groupService) startGroupService(ctx context.Context, isInit bool) error {
	if gp == nil {
		return ErrCannotConnectNetwork
	}
	// getbalance
	uaddr, err := address.GetAddressFromID(gp.userid)
	if err != nil {
		return err
	}

	balance, err := contracts.QueryBalance(uaddr.Hex())
	if err != nil {
		config, err := localNode.Repo.Config()
		if err != nil {
			return err
		}
		if config.Test {
			balance = big.NewInt(0)
		}
	}
	log.Println(uaddr.String(), " has balance (wei): ", balance)
	// 说明该user没钱，该user为testuser;否则为有金额的实际用户
	if balance.Cmp(big.NewInt(0)) <= 0 {
		// 判断是否为初始化启动
		if isInit {
			err := gp.findKeeperAndProviderInit(ctx, balance)
			return err
		}
		// 尝试以inited方式启动
		err = gp.findKeeperAndProviderNotInit(ctx)
		if err != nil {
			log.Println("Can't connect the user's keepers, so we think the user hasn't inited, begin to init user...", gp.userid)
			err = gp.findKeeperAndProviderInit(ctx, balance)
			if err != nil {
				return err
			}
		}
	} else {
		// getUK
		_, uk, err := contracts.GetUKFromResolver(uaddr)
		switch err {
		case nil: //部署过
			log.Println("begin to find keepers and providers to start user : ", gp.userid)
			item, err := contracts.GetUpkeepingInfo(uaddr, uk)
			if err != nil {
				return err
			}
			// keeper数量、provider的数量应以合约约定为主
			gp.keeperSLA = int(item.KeeperSLA)
			gp.providerSLA = int(item.ProviderSLA)
			gp.upKeepingItem = &item
			err = gp.connectKeepersAndProviders(ctx, item.KeeperIDs, item.ProviderIDs)
			if err != nil {
				return err
			}
		case contracts.ErrNotDeployedMapper, contracts.ErrNotDeployedUk: //没有部署过
			log.Println("begin to init user : ", gp.userid)
			err := gp.findKeeperAndProviderInit(ctx, balance)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

// user init
func (gp *groupService) findKeeperAndProviderInit(ctx context.Context, balance *big.Int) error {
	addr, err := address.GetAddressFromID(gp.userid)
	if err != nil {
		return err
	}
	//deploy query first
	config, err := localNode.Repo.Config()
	if err != nil {
		log.Println(err)
		return err
	}

	var queryAddr common.Address
	if !config.Test {
		//判断账户余额能否部署query合约 + upKeeping + channel合约
		var moneyPerDay = new(big.Int)
		moneyPerDay = moneyPerDay.Mul(big.NewInt(gp.storePrice), big.NewInt(gp.storeSize))
		var moneyAccount = new(big.Int)
		moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(gp.storeDays))

		deployPrice := big.NewInt(int64(740621000000000))
		deployPrice.Add(big.NewInt(1128277), big.NewInt(int64(652346*gp.providerSLA)))
		var leastMoney = new(big.Int)
		leastMoney = leastMoney.Add(moneyAccount, deployPrice)
		if balance.Cmp(leastMoney) < 0 { //余额不足
			log.Println(ErrBalance, ", you need more balance ")
			return ErrBalance
		}

		queryAddr, err = contracts.DeployQuery(addr, utils.EthSkByteToEthString(gp.privateKey), gp.storeSize, gp.storeDays, gp.storePrice, gp.keeperSLA, gp.providerSLA, gp.reDeploy)
		if err != nil {
			log.Println("fail to deploy query contract")
			return err
		}
	} else {
		log.Println("测试环境，将不部署query合约")
	}

	log.Println("Begin to find keepers for init...", "\nthe uid is: "+gp.userid+"\nthe addr is: "+addr.String())

	err = testConnect()
	if err != nil {
		return err
	}

	//构造init信息并发送 此时，初始化阶段为collecting
	kmInit, err := metainfo.NewKeyMeta(gp.userid, metainfo.UserInitReq, strconv.Itoa(gp.keeperSLA), strconv.Itoa(gp.providerSLA), queryAddr.String())
	if err != nil {
		log.Println("findKeeperAndProviderInit()NewKeyMeta()error!")
		return err
	}
	go broadcastMetaMessage(kmInit, "")
	err = setUserState(gp.userid, collecting)
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
			userState, err := getUserState(gp.userid)
			if err != nil {
				return err
			}
			switch userState {
			case collecting:
				timeOutCount++
				if len(gp.tempKeepers) >= gp.keeperSLA && len(gp.tempProviders) >= gp.providerSLA {
					//收集到足够的keeper和Provider 进行挑选并给keeper发送确认信息，初始化阶段变为collectComplete
					err := setUserState(gp.userid, collectCompleted)
					if err != nil {
						return err
					}
					gp.userInitNotif(kmInit)
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
				gp.handleRest()
			default:
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

//userInitNotIf 收集齐KP信息之后， 选择keeper和provider，构造确认信息发给keeper
func (gp *groupService) userInitNotif(km *metainfo.KeyMeta) {
	err := gp.chooseKeepersAndProviders()
	if err != nil {
		if err == ErrNoEnoughKeeper { //有keeper没有连接上，重新设置初始化状态为collecting
			err := setUserState(gp.userid, collecting)
			if err != nil {
				log.Println("userInitNotif()setUserState()err:", err)
			}
		}
		log.Println(err)
		return
	}

	//构造本节点keeper信息和provider信息 放入硬盘 id1id2id3.......(无分隔符)
	var assignedKeeper, assignedProvider string
	for _, keeper := range gp.keepers {
		assignedKeeper += keeper.keeperID
	}

	for _, provider := range gp.providers {
		assignedProvider += provider.providerID
	}
	//构造发给keeper的初始化确认信息并发送给自己的所有keeper
	assignedKP := assignedKeeper + metainfo.DELIMITER + assignedProvider
	km.SetKeyType(metainfo.UserInitNotif)

	var wg sync.WaitGroup
	for _, keeper := range gp.keepers { //循环发消息
		wg.Add(1)
		log.Println("Notify keeper:", keeper)
		go func(keeper string) {
			defer wg.Done()
			retry := 0
			// retry
			for retry < 10 {
				_, err := sendMetaRequest(km, assignedKP, keeper) //发送确认信息
				if err != nil {
					retry++
					time.Sleep(30 * time.Second)
				} else {
					break
				}
			}

		}(keeper.keeperID)

	}
	wg.Wait()

	log.Println("Waiting for keepers' response")
}

//  AddKeepersAndProviders 把keeper和provider的id加入groupservice的备选中
func (gp *groupService) addKeepersAndProviders(keepers, providers string) {
	for i := 0; i < len(keepers)/utils.IDLength; i++ {
		kid := keepers[i*utils.IDLength : (i+1)*utils.IDLength]
		_, err := peer.IDB58Decode(kid)
		if err != nil {
			continue
		}
		if !utils.CheckDup(gp.tempKeepers, kid) {
			continue
		}
		if sc.ConnectTo(context.Background(), localNode, kid) {
			gp.tempKeepers = append(gp.tempKeepers, kid)
		}
	}
	for i := 0; i < len(providers)/utils.IDLength; i++ {
		pid := providers[i*utils.IDLength : (i+1)*utils.IDLength]
		_, err := peer.IDB58Decode(pid)
		if err != nil {
			continue
		}
		if !utils.CheckDup(gp.tempProviders, pid) {
			continue
		}
		if sc.ConnectTo(context.Background(), localNode, pid) {
			gp.tempProviders = append(gp.tempProviders, pid)
		}
	}
}

//chooseKeepersAndProviders 启动流程，收集到足够KP信息，从备选中 选出本节点的keeper和provider，选择keeper的时候，检查连接性
func (gp *groupService) chooseKeepersAndProviders() error {
	userState, err := getUserState(gp.userid)
	if err != nil {
		return err
	}
	if userState != collectCompleted {
		return ErrWrongInitState
	}
	log.Println("Has enough Keeper and Providers, choosing...")
	gp.keepers = make([]*keeperInfo, 0, gp.keeperSLA)
	gp.providers = make([]*providerInfo, 0, gp.providerSLA)
	//选择keeper
	gp.tempKeepers = utils.DisorderArray(gp.tempKeepers)
	for i := 0; i < len(gp.tempKeepers); {
		if i >= gp.keeperSLA {
			break
		}
		kidStr := gp.tempKeepers[i]
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
		gp.keepers = append(gp.keepers, tempK)
	}
	if len(gp.keepers) < gp.keeperSLA {
		return ErrNoEnoughKeeper
	}

	//选择provider
	gp.tempProviders = utils.DisorderArray(gp.tempProviders)
	for i := 0; i < len(gp.tempProviders); {
		if i >= gp.providerSLA {
			break
		}
		pidStr := gp.tempProviders[i]
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

		gp.providers = append(gp.providers, tempP)
	}

	log.Println("Choose completed")

	return nil
}

// keeperConfirm 第四次握手 确认Keeper启动完毕
// PeerID/"user_init_notif_res"/"bft","simple"或 IP:p2pport/IP:rpcport
func (gp *groupService) keeperConfirm(keeper string, initRes string) error {
	splitInitRes := strings.Split(initRes, metainfo.DELIMITER)
	var ConnectedCount int
	//将发来信息的keeper记录为连接成功
	for _, kp := range gp.keepers {
		if strings.Compare(kp.keeperID, keeper) == 0 && !kp.connected {
			if len(splitInitRes) >= 2 {
				kp.isBFT = true
			}
			kp.connected = true
			log.Printf("Receive %s's response, waiting for other keepers\n", kp.keeperID)
		}
		if kp.connected {
			ConnectedCount++
		}
	}
	//与所有keeper都连接成功了
	if ConnectedCount == len(gp.keepers) {
		log.Println("Receive all keepers' response")
		err := localNode.Repo.SetConfigKey("IsInit", false) //初始化过后，设置IsInit
		if err != nil {
			log.Println("gp.localNode.Repo.SetConfigKey failed :", err)
		}

		// 状态改为部署合约中
		err = setUserState(gp.userid, onDeploy)
		if err != nil {
			log.Println("setUserState failed :", err)
			return err
		}

	}
	return nil
}

func (gp *groupService) handleRest() error {
	err := gp.userBLS12ConfigInit()
	if err != nil {
		return nil
	}

	gp.putUserConfig()

	//部署合约
	err = gp.deployUpKeepingAndChannel()
	if err != nil {
		log.Println("gp.deployUpKeepingAndChannel failed :", err)
	}

	saveContracts(gp.userid)

	// 构造key告诉keeper和provider自己已经部署好合约
	kmPid, err := metainfo.NewKeyMeta(gp.userid, metainfo.UserDeployedContracts)
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
	log.Println(gp.userid + ":Group Service is ready")
	err = setUserState(gp.userid, groupStarted)
	if err != nil {
		log.Println("setUserState failed :", err)
		return err
	}
	return nil
}

func (gp *groupService) connectKeepersAndProviders(ctx context.Context, keepers, providers []string) error {
	log.Println("Begin to connect user's keepers and providers:", gp.userid)
	waitTime := 0
	for {
		if waitTime > 60 {
			log.Println(ErrCannotConnectNetwork, "please restart and retry.")
			return ErrCannotConnectNetwork
		}
		if connPeers := localNode.PeerHost.Network().Peers(); len(connPeers) != 0 {
			break
		} else {
			log.Println(ErrCannotConnectNetwork, "waiting...")
			time.Sleep(10 * time.Second)
		}
		waitTime++
	}

	connectTryCount := 5
	for i := 0; i < connectTryCount; i++ {
		var unsuccess []string
		for _, kid := range keepers {
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
			kmKid, err := metainfo.NewKeyMeta(gp.userid, metainfo.Local, metainfo.SyncTypeKid)
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
			for _, kinfo := range gp.keepers {
				if kid == kinfo.keeperID {
					exist = true
					break
				}
			}
			if !exist {
				gp.keepers = append(gp.keepers, tempKeeper)
			}
			if tempKeeper.isBFT {
				log.Println("Connect to keeper", tempKeeper.keeperID, "use bft mode")
			} else {
				log.Println("Connect to keeper", tempKeeper.keeperID, "use simple mode")
			}
		}

		if len(gp.keepers) == gp.keeperSLA {
			break
		}

		keepers = unsuccess
		// 每一个keeper连接后的判断，若连接数量足够后就立即进行而不是全部连接后再进行
		if len(gp.keepers) < KeeperSLA {
			time.Sleep(time.Minute)
		}
	}

	if len(gp.keepers) <= 0 {
		return ErrNoEnoughKeeper
	}

	for i := 0; i < connectTryCount; i++ {
		if gp.keySet == nil {
			err := gp.loadBLS12Config()
			if err != nil {
				log.Println("Load BLS12 Config error:", err)
				return err
			}
		}
		for _, pid := range providers {
			if sc.ConnectTo(ctx, localNode, pid) {
				log.Println("Connect to provider-", pid, "success.")
			} else {
				log.Println("Connect to provider-", pid, "failed.")
			}

			tempP := &providerInfo{
				providerID: pid,
			}

			exist := false
			for _, kinfo := range gp.providers {
				if pid == kinfo.providerID {
					exist = true
					break
				}
			}

			if !exist {
				gp.providers = append(gp.providers, tempP)
			}
		}

		if len(gp.providers) == gp.providerSLA {
			break
		}

		if len(gp.providers) <= 0 {
			time.Sleep(time.Minute)
			continue
		}
	}

	if len(gp.providers) <= 0 {
		return ErrNoEnoughProvider
	}

	// 构造key告诉keeper和provider自己已经启动
	kmPid, err := metainfo.NewKeyMeta(gp.userid, metainfo.UserDeployedContracts)
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
	log.Println(gp.userid + ":Group Service is ready")

	saveContracts(gp.userid)

	err = setUserState(gp.userid, groupStarted)
	if err != nil {
		return err
	}
	return nil
}

func (gp *groupService) findKeeperAndProviderNotInit(ctx context.Context) error {
	log.Println("Begin to find Keeper for start user service:", gp.userid)
	km, err := metainfo.NewKeyMeta(gp.userid, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		log.Println("findKeeperAndProviderNotInit()NewKeyMeta()error:", err)
		return err
	}

	// 判断是否联网
	err = testConnect()
	if err != nil {
		return err
	}

	waitTime := 0
	for {
		if waitTime > 10 { //尝试多次依然连接不上？
			log.Println(gp.userid, "Cannot find my keeper,please restart and try again.")
			return ErrCannotConnectKeeper
		}
		if kids, err := getKeyFrom(km.ToString(), ""); kids != nil && err == nil { //先看本地有没有
			if remain := len(kids) % utils.IDLength; remain != 0 {
				kids = kids[:len(kids)-remain]
			}
			//有的keeper连接不上
			tick := time.Tick(10 * time.Second)
			tickCount := 0
			for {
				select {
				case <-tick: //持续尝试连接已有的keeper节点
					if tickCount > 10 { //尝试超过十分钟还连不上
						log.Println("Cannot connect to some keeper, please restart...")
						return ErrCannotConnectKeeper
					}
					tickCount++
					for i := 0; i < len(kids)/utils.IDLength; i++ {
						kid := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
						if sc.ConnectTo(ctx, localNode, kid) {
							tempKeeper := &keeperInfo{
								keeperID: kid,
							}
							kmKid, err := metainfo.NewKeyMeta(gp.userid, metainfo.Local, metainfo.SyncTypeKid)
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
							var flag int
							for flag = 0; flag < len(gp.keepers); flag++ {
								if kid == gp.keepers[flag].keeperID {
									break
								}
							}
							if flag == len(gp.keepers) {
								gp.keepers = append(gp.keepers, tempKeeper)
							}
							if tempKeeper.isBFT {
								log.Println("Connect to keeper", tempKeeper.keeperID, "use bft mode")
							} else {
								log.Println("Connect to keeper", tempKeeper.keeperID, "use simple mode")
							}
						} else {
							log.Println("Connect to keeper", kid, "failed.")
						}

						if len(gp.keepers) >= gp.keeperSLA {
							if gp.keySet == nil {
								err = gp.loadBLS12Config()
								if err != nil {
									log.Println("Load BLS12 Config error:", err)
									return err
								}
							}
							kmPid, err := metainfo.NewKeyMeta(gp.userid, metainfo.Local, metainfo.SyncTypePid)
							if err != nil {
								return err
							}
							if pids, err := getKeyFrom(kmPid.ToString(), ""); pids != nil && err == nil { //只尝试连接一次provider
								if remain := len(pids) % utils.IDLength; remain != 0 {
									pids = pids[:len(pids)-remain]
								}
								providers := string(pids)
								for i := 0; i < len(providers)/utils.IDLength; i++ {
									pid := providers[i*utils.IDLength : (i+1)*utils.IDLength]
									if sc.ConnectTo(ctx, localNode, pid) {
										log.Println("Connect to provider-", pid, "success.")
									} else {
										log.Println("Connect to provider-", pid, "failed.")
									}
									tmp := &providerInfo{
										providerID: pid,
									}
									gp.providers = append(gp.providers, tmp)
								}
								log.Println(gp.userid + ":Group Service is ready")
								err = setUserState(gp.userid, groupStarted)
								if err != nil {
									log.Println("setUserState failed :", err)
								}
							}
							return nil
						}
					}
				case <-ctx.Done():
					return nil
				}
			}
		} else {
			log.Println(gp.userid, "Cannot find my keeper, retrying...")
			time.Sleep(10 * time.Second)
		}
		waitTime++
	}
}
