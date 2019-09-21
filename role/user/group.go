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
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

// ConstructGroupService Constructs GroupService
func ConstructGroupService(userid string, privKey []byte, duration int64, capacity int64, price int64, ks int, ps int, redeploy bool) *GroupService {
	if privKey == nil {
		return nil
	}
	return &GroupService{
		Userid:         userid,
		PrivateKey:     privKey,
		localPeersInfo: PeersInfo{},
		storeDays:      duration,
		storeSize:      capacity,
		storePrice:     price,
		keeperSLA:      ks,
		providerSLA:    ps,
		reDeploy:       redeploy,
	}
}

// StartGroupService starts gp
// TODO:在使用provider之前都应该检查一下连接性
func (gp *GroupService) StartGroupService(ctx context.Context, pwd string, isInit bool) error {
	if gp == nil {
		return ErrCannotConnectNetwork
	}
	gp.password = pwd
	// getbalance
	uaddr, err := address.GetAddressFromID(gp.Userid)
	if err != nil {
		return err
	}
	config, err := localNode.Repo.Config()
	if err != nil {
		return err
	}
	balance, err := contracts.QueryBalance(uaddr.Hex())
	if err != nil {
		if config.Test {
			balance = big.NewInt(0)
		} else {
			return err
		}

	}
	log.Println(uaddr.String(), " balance", balance)
	// 说明该user没钱，该user为testuser;否则为有金额的实际用户
	if balance.Cmp(big.NewInt(0)) <= 0 {
		// 判断是否为初始化启动
		if isInit {
			err := gp.findKeeperAndProviderInit(ctx)
			return err
		}
		// 尝试以inited方式启动
		err = gp.findKeeperAndProviderNotInit(ctx)
		if err != nil {
			log.Println("Can't connect the user's keepers, so we think the user hasn't inited, begin to init user...", gp.Userid)
			err = gp.findKeeperAndProviderInit(ctx)
			if err != nil {
				return err
			}
		}
	} else {
		// getUK
		_, uk, err := contracts.GetUKFromResolver(uaddr)
		switch err {
		case nil: //部署过
			log.Println("begin to find keepers and providers to start user : ", gp.Userid)
			item, err := contracts.GetUpkeepingInfo(uaddr, uk)
			if err != nil {
				return err
			}
			// keeper数量、provider的数量应以合约约定为主
			gp.keeperSLA = int(item.KeeperSLA)
			gp.providerSLA = int(item.ProviderSLA)
			err = gp.connectKeepersAndProviders(ctx, item.KeeperIDs, item.ProviderIDs)
			if err != nil {
				return err
			}
		case contracts.ErrNotDeployedMapper, contracts.ErrNotDeployedUk: //没有部署过
			log.Println("begin to init user : ", gp.Userid)
			err := gp.findKeeperAndProviderInit(ctx)
			if err != nil {
				return err
			}
		default:
			return err
		}
	}
	return nil
}

func (gp *GroupService) connectKeepersAndProviders(ctx context.Context, keepers, providers []string) error {
	log.Println("Begin to connect user's keepers and providers:", gp.Userid)
	waitTime := 0 //进行网络连接
	for {
		if waitTime > 60 { //连不上网？
			log.Println(ErrCannotConnectNetwork, "please restart and retry.")
			return ErrCannotConnectNetwork
		}
		if connPeers := localNode.PeerHost.Network().Peers(); len(connPeers) != 0 { //刚启动还没连接节点，等等
			break //连上网了，退出
		} else {
			log.Println(ErrCannotConnectNetwork, "waiting...")
			time.Sleep(10 * time.Second) //没联网，等联网
		}
		waitTime++
	}

	if len(gp.localPeersInfo.Keepers) >= gp.keeperSLA {
		return nil
	}

	connectTryCount := 5
	// 第一次对所有keeper进行连接，第二次对连接失败的keeper进行连接，依次类推
	for i := 0; i < connectTryCount; i++ {
		var unsuccess []string
		for _, kid := range keepers {
			// 连接失败加入unsuccess
			if !sc.ConnectTo(ctx, localNode, kid) {
				unsuccess = append(unsuccess, kid)
				log.Println("Connect to keeper", kid, "failed.")
				continue
			}
			tempKeeper := &KeeperInfo{
				KeeperID: kid,
			}
			kmKid, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypeKid)
			if err != nil {
				return err
			}
			res, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmKid.ToString(), kid)
			if err == nil && res != nil {
				resStr := string(res)
				splitRes := strings.Split(resStr, metainfo.DELIMITER)
				if len(splitRes) > 3 {
					if splitRes[1] == metainfo.SyncTypeBft {
						tempKeeper.IsBFT = true
					}
				}
			}
			// 检查该keeper是否已添加
			var repeat int
			for repeat = 0; repeat < len(gp.localPeersInfo.Keepers); repeat++ {
				if kid == gp.localPeersInfo.Keepers[repeat].KeeperID {
					break
				}
			}
			if repeat == len(gp.localPeersInfo.Keepers) {
				gp.localPeersInfo.Keepers = append(gp.localPeersInfo.Keepers, tempKeeper)
			}
			if tempKeeper.IsBFT {
				log.Println("Connect to keeper", tempKeeper.KeeperID, "use bft mode")
			} else {
				log.Println("Connect to keeper", tempKeeper.KeeperID, "use simple mode")
			}
		}

		keepers = unsuccess
		// 每一个keeper连接后的判断，若连接数量足够后就立即进行而不是全部连接后再进行
		if len(gp.localPeersInfo.Keepers) < KeeperSLA {
			time.Sleep(time.Minute)
		}
	}

	if len(gp.localPeersInfo.Keepers) <= 0 {
		return ErrNoEnoughKeeper
	}

	for i := 0; i < connectTryCount; i++ {
		if gp.KeySet == nil {
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
			if utils.CheckDup(gp.localPeersInfo.Providers, pid) {
				gp.localPeersInfo.Providers = append(gp.localPeersInfo.Providers, pid) //将Provider加入内存缓冲
			}
		}

		if len(gp.localPeersInfo.Providers) <= 0 {
			time.Sleep(time.Minute)
			continue
		}

		// 构造key告诉keeper和provider自己已经启动
		kmPid, err := metainfo.NewKeyMeta(gp.Userid, metainfo.UserDeployedContracts)
		if err != nil {
			log.Println("Construct Deployed key error", err)
			return err
		}
		for _, keeper := range gp.localPeersInfo.Keepers {
			_, err = sendMetaRequest(kmPid, gp.Userid, keeper.KeeperID)
			if err != nil {
				log.Println("Send keeper", keeper.KeeperID, " err:", err)
			}
		}
		for _, provider := range gp.localPeersInfo.Providers {
			_, err = sendMetaRequest(kmPid, gp.Userid, provider)
			if err != nil {
				log.Println("Send provider", provider, " err:", err)
			}
		}
		log.Println(gp.Userid + ":Group Service is ready")
		err = SetUserState(gp.Userid, GroupStarted)
		if err != nil {
			return err
		}
		return nil
	}

	if len(gp.localPeersInfo.Providers) <= 0 {
		return ErrNoEnoughProvider
	}

	return nil
}

// user初始化流程
func (gp *GroupService) findKeeperAndProviderInit(ctx context.Context) error {
	addr, err := address.GetAddressFromID(gp.Userid)
	if err != nil {
		return err
	}
	//user部署query合约
	config, err := localNode.Repo.Config()
	if err != nil {
		log.Println(err)
		return err
	}

	var queryAddr common.Address
	if !config.Test {
		balance, _ := contracts.QueryBalance(addr.Hex()) //获得用户的账户余额
		log.Println("balance:", balance)
		//判断账户余额能否部署query合约
		deployPrice := big.NewInt(int64(740621000000000))
		if balance.Cmp(deployPrice) < 0 { //余额不足
			log.Println(ErrBalance)
			return ErrBalance
		}

		queryAddr, err = contracts.DeployQuery(addr, address.SkByteToString(gp.PrivateKey), gp.storeSize, gp.storeDays, gp.storePrice, gp.keeperSLA, gp.providerSLA, gp.reDeploy)
		if err != nil {
			return err
		}
	} else {
		log.Println("测试环境，将不部署query合约")
	}

	log.Println("Begin to find Keepers for init...", "\nthe uid is: "+gp.Userid+"\nthe addr is: "+addr.String())

	waitTime := 0 //进行网络连接
	for {
		if waitTime > 60 { //连不上网？
			log.Println(ErrCannotConnectNetwork, "please restart and retry.")
			return ErrCannotConnectNetwork
		}
		if connPeers := localNode.PeerHost.Network().Peers(); len(connPeers) != 0 { //刚启动还没连接节点，等等
			break //连上网了，退出
		} else {
			log.Println(ErrCannotConnectNetwork, "waiting...")
			time.Sleep(10 * time.Second) //没联网，等联网
		}
		waitTime++
	}

	//user的BLS12 config相关
	kmBls, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return err
	}

	userBLS12config, err := gp.userBLS12ConfigInit() //初始化配置
	if err != nil {
		log.Println("Cannot init BLS Config-", err)
		return err
	}
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmBls.ToString(), string(userBLS12config), "local") //先在本地保存一份
	if err != nil {
		log.Println("CmdPutTo()err")
		return err
	}

	//构造init信息并发送 此时，初始化阶段为collecting
	kmInit, err := metainfo.NewKeyMeta(gp.Userid, metainfo.UserInitReq, strconv.Itoa(gp.keeperSLA), strconv.Itoa(gp.providerSLA), queryAddr.String())
	if err != nil {
		log.Println("findKeeperAndProviderInit()NewKeyMeta()error!")
		return err
	}
	go broadcastMetaMessage(kmInit, "")
	err = SetUserState(gp.Userid, Collecting)
	if err != nil {
		return err
	}
	// 二十分钟超时
	timeOutCount := 0
	tick := time.Tick(30 * time.Second)
	for {
		select {
		case <-tick: //每过30s 检查是否收到了足够的KP信息，如果不足，继续发送初始化请求，足够的时候进行KP的选择和确认
			if timeOutCount >= 40 {
				return ErrTimeOut
			}
			userState, err := GetUserServiceState(gp.Userid)
			if err != nil {
				return err
			}
			switch userState {
			case Collecting:
				timeOutCount++
				if len(gp.tempKeepers) >= gp.keeperSLA && len(gp.tempProviders) >= gp.providerSLA {
					//收集到足够的keeper和Provider 进行挑选并给keeper发送确认信息，初始化阶段变为collectComplete
					err := SetUserState(gp.Userid, CollectCompleted)
					if err != nil {
						return err
					}
					gp.userInitNotif(kmInit, kmBls.ToString(), string(userBLS12config))
				} else {
					log.Printf("Timeout, No enough Keepers and Providers,Have k:%d p:%d,want k:%d p:%d, retrying...\n", len(gp.tempKeepers), len(gp.tempProviders), gp.keeperSLA, gp.providerSLA)
					go broadcastMetaMessage(kmInit, "")
				}
			case CollectCompleted:
				timeOutCount++
				//TODO：等待keeper的第四次握手超时怎么办，目前继续等待
				log.Printf("Timeout, waiting keeper response\n")
				for _, keeperInfo := range gp.localPeersInfo.Keepers {
					if !keeperInfo.Connected {
						log.Printf("Keeper %s not response, waiting...", keeperInfo.KeeperID)
					}
				}
			case OnDeploy:
			default:
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

//  AddKeepersAndProviders 把keeper和provider的id加入groupservice的备选中
func (gp *GroupService) addKeepersAndProviders(keepers, providers string) {
	if remain := len(keepers) % utils.IDLength; remain != 0 {
		keepers = keepers[:len(keepers)-remain]
	}
	for i := 0; i < len(keepers)/utils.IDLength; i++ {
		kid := keepers[i*utils.IDLength : (i+1)*utils.IDLength]
		if !utils.CheckDup(gp.tempKeepers, kid) {
			continue
		}
		if sc.ConnectTo(context.Background(), localNode, kid) {
			gp.tempKeepers = append(gp.tempKeepers, kid)
		}
	}
	if remain := len(providers) % utils.IDLength; remain != 0 {
		providers = providers[:len(providers)-remain]
	}
	for i := 0; i < len(providers)/utils.IDLength; i++ {
		pid := providers[i*utils.IDLength : (i+1)*utils.IDLength]
		if !utils.CheckDup(gp.tempProviders, pid) {
			continue
		}
		if sc.ConnectTo(context.Background(), localNode, pid) {
			gp.tempProviders = append(gp.tempProviders, pid)
		}
	}
}

//userInitNotIf 收集齐KP信息之后， 选择keeper和provider，构造确认信息发给keeper
func (gp *GroupService) userInitNotif(km *metainfo.KeyMeta, userBLS12configkey, userBLS12configstring string) {
	err := gp.chooseKeepersAndProviders()
	if err != nil {
		if err == ErrNoEnoughKeeper { //有keeper没有连接上，重新设置初始化状态为collecting
			err := SetUserState(gp.Userid, Collecting)
			if err != nil {
				log.Println("userInitNotif()SetUserState()err:", err)
			}
		}
		log.Println(err)
		return
	}

	//构造本节点keeper信息和provider信息 放入硬盘 id1id2id3.......(无分隔符)
	var assignedKeeper, assignedProvider string
	for _, keeper := range gp.localPeersInfo.Keepers {
		assignedKeeper += keeper.KeeperID
	}

	for _, provider := range gp.localPeersInfo.Providers {
		assignedProvider += provider
	}
	//构造发给keeper的初始化确认信息并发送给自己的所有keeper
	assignedKP := assignedKeeper + metainfo.DELIMITER + assignedProvider
	km.SetKeyType(metainfo.UserInitNotif)

	var wg sync.WaitGroup
	notif := func(wg *sync.WaitGroup, keeper string) {
		defer wg.Done()
		log.Println("Notify keeper:", keeper)
		_, err := sendMetaRequest(km, assignedKP, keeper) //发送确认信息
		if err != nil {
			log.Println("gp.localNode.Routing.MetaNewUserNotif failed :", err)
		}
	}
	//TODO:Keeper有没有错误的逻辑需要考虑，如果有一部分回复completed，一部分回复error怎么办
	for _, keeper := range gp.localPeersInfo.Keepers { //循环发消息
		wg.Add(1)
		go notif(&wg, keeper.KeeperID)
	}
	wg.Wait()

	//最后发送本节点的BLS12公钥到自己的keeper上保存
	for _, keeper := range gp.localPeersInfo.Keepers {
		err := localNode.Routing.(*dht.IpfsDHT).CmdPutTo(userBLS12configkey, userBLS12configstring, keeper.KeeperID)
		if err != nil {
			log.Println("gp.localNode.Routing.CmdPut failed :", err)
		}
	}
	log.Println("Waiting for keepers' response")
}

//chooseKeepersAndProviders 启动流程，收集到足够KP信息，从备选中 选出本节点的keeper和provider，选择keeper的时候，检查连接性
func (gp *GroupService) chooseKeepersAndProviders() error {
	userState, err := GetUserServiceState(gp.Userid)
	if err != nil {
		return err
	}
	if userState != CollectCompleted {
		return ErrWrongInitState
	}
	log.Println("Has enough Keeper and Providers, choosing...")
	gp.localPeersInfo.Keepers = make([]*KeeperInfo, 0, gp.keeperSLA)
	gp.localPeersInfo.Providers = make([]string, 0, gp.providerSLA)
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
		tempKeeper := &KeeperInfo{
			KeeperID:  kidStr,
			Connected: false,
		}
		i++
		gp.localPeersInfo.Keepers = append(gp.localPeersInfo.Keepers, tempKeeper)
	}
	if len(gp.localPeersInfo.Keepers) < gp.keeperSLA {
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
		gp.localPeersInfo.Providers = append(gp.localPeersInfo.Providers, pidStr)
	}

	log.Println("Choose completed, Providers:", gp.localPeersInfo.Providers)

	return nil
}

// keeperConfirm 第四次握手 确认Keeper启动完毕
// PeerID/"user_init_notif_res"/"bft","simple"或 IP:p2pport/IP:rpcport
func (gp *GroupService) keeperConfirm(keeper string, initRes string) error {
	splitInitRes := strings.Split(initRes, metainfo.DELIMITER)
	var ConnectedCount int
	//将发来信息的keeper记录为连接成功
	for _, keeperInfo := range gp.localPeersInfo.Keepers {
		if strings.Compare(keeperInfo.KeeperID, keeper) == 0 && !keeperInfo.Connected {
			if len(splitInitRes) >= 2 {
				keeperInfo.IsBFT = true
			}
			keeperInfo.Connected = true
			log.Printf("Receive %s's response, waiting for other keepers\n", keeper)
		}
		if keeperInfo.Connected {
			ConnectedCount++
		}
	}
	//与所有keeper都连接成功了
	if ConnectedCount == len(gp.localPeersInfo.Keepers) {
		log.Println("Receive all keepers' response")
		err := localNode.Repo.SetConfigKey("IsInit", false) //初始化过后，设置IsInit
		if err != nil {
			log.Println("gp.localNode.Repo.SetConfigKey failed :", err)
		}

		// 状态改为部署合约中
		err = SetUserState(gp.Userid, OnDeploy)
		if err != nil {
			log.Println("SetUserState failed :", err)
			return err
		}

		//部署合约
		err = gp.deployUpKeepingAndChannel()
		if err != nil {
			log.Println("gp.deployUpKeepingAndChannel failed :", err)
		}
		// 构造key告诉keeper和provider自己已经部署好合约
		kmPid, err := metainfo.NewKeyMeta(gp.Userid, metainfo.UserDeployedContracts)
		if err != nil {
			log.Println("Construct Deployed key error", err)
			return err
		}
		for _, keeper := range gp.localPeersInfo.Keepers {
			_, err = sendMetaRequest(kmPid, gp.Userid, keeper.KeeperID)
			if err != nil {
				log.Println("Send keeper", keeper.KeeperID, " err:", err)
			}
		}
		for _, provider := range gp.localPeersInfo.Providers {
			_, err = sendMetaRequest(kmPid, gp.Userid, provider)
			if err != nil {
				log.Println("Send provider", provider, " err:", err)
			}
		}
		log.Println(gp.Userid + ":Group Service is ready")
		err = SetUserState(gp.Userid, GroupStarted)
		if err != nil {
			log.Println("SetUserState failed :", err)
			return err
		}
	}
	return nil
}

func (gp *GroupService) deployUpKeepingAndChannel() error {
	hexPK, localAddress, keepers, providers, err := buildUKParams(gp.Userid, gp.password, gp.localPeersInfo)
	if err != nil {
		log.Println("getParams:", err)
		return err
	}

	config, err := localNode.Repo.Config()
	if err != nil {
		log.Println(err)
		return err
	}
	balance, err := contracts.QueryBalance(localAddress.Hex()) //获得用户的账户余额
	if err != nil {
		if config.Test {
			balance = big.NewInt(0)
		}
	}
	log.Println("balance:", balance)

	d := gp.storeDays
	s := gp.storeSize
	price := gp.storePrice

	var moneyPerDay = new(big.Int)
	var moneyAccount = new(big.Int)
	moneyPerDay = moneyPerDay.Mul(big.NewInt(price), big.NewInt(s))
	moneyAccount = moneyAccount.Mul(moneyPerDay, big.NewInt(d)) //部署合约的储蓄金额，默认是10x100x100

	//判断账户余额能否部署upKeeping以及channel合约
	var deployPrice = new(big.Int)
	deployPrice = deployPrice.Add(big.NewInt(1128277), big.NewInt(int64(652346*len(providers))))
	var leastMoney = new(big.Int)
	leastMoney = leastMoney.Add(moneyAccount, deployPrice)
	if balance.Cmp(leastMoney) < 0 { //余额不足
		log.Println(ErrBalance)
		return ErrBalance
	}

	err = contracts.DeployUpkeeping(hexPK, localAddress, keepers, providers, d, s, price, moneyAccount)
	if err != nil {
		return err
	}

	//部署好upKeeping合约后，将user部署的query合约的completed参数设为true
	queryAddr, err := contracts.GetMarketAddr(localAddress, localAddress, contracts.Query)
	if err != nil {
		return err
	}
	err = contracts.SetQueryCompleted(hexPK, localAddress, queryAddr)
	if err != nil {
		return err
	}

	//依次与各provider签署channel合约
	timeOut := big.NewInt(int64(d * 24 * 60 * 60)) //秒，存储时间
	var moneyToChannel = new(big.Int)
	moneyToChannel = moneyToChannel.Mul(big.NewInt(s), big.NewInt(int64(utils.READPRICEPERMB))) //暂定往每个channel合约中存储金额为：存储大小 x 每MB单价

	var wg sync.WaitGroup

	for _, proAddr := range providers {
		wg.Add(1)
		providerAddr := proAddr
		go func() {
			defer wg.Done()
			channelAddr, err := contracts.DeployChannelContract(hexPK, localAddress, providerAddr, timeOut, moneyToChannel)
			if err != nil {
				return
			}
			//设置channel的value初始值为0
			//存到本地
			channelValueKeyMeta, err := metainfo.NewKeyMeta(channelAddr.String(), metainfo.Local, metainfo.SyncTypeChannelValue)
			if err != nil {
				return
			}
			key := channelValueKeyMeta.ToString() // hexChannelAddress|13|channelvalue
			err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(key, strconv.FormatInt(0, 10), "local")
			if err != nil {
				return
			}
			//存到provider上
			providerID, err := address.GetIDFromAddress(providerAddr.String())
			if err != nil {
				return
			}
			err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(key, strconv.FormatInt(0, 10), providerID)
			if err != nil {
				return
			}
		}()
	}
	wg.Wait()
	log.Println("user has deployed all channel-contract successfully!")
	return nil
}

func (gp *GroupService) findKeeperAndProviderNotInit(ctx context.Context) error {
	log.Println("Begin to find Keeper for start user service:", gp.Userid)
	km, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		log.Println("findKeeperAndProviderNotInit()NewKeyMeta()error:", err)
		return err
	}

	// 判断是否联网
	waitTime := 0
	for {
		if waitTime > 10 { //连不上网？
			log.Println("Cannot connect to other peer, please restart and try again.")
			return ErrCannotConnectNetwork
		}
		if connPeers := localNode.PeerHost.Network().Peers(); len(connPeers) != 0 { //刚启动还没连接节点，等等
			break //连上网了，退出
		} else {
			time.Sleep(10 * time.Second) //没联网，等联网
		}
		log.Println("Cannot connect to other peer, retrying...")
		waitTime++
	}
	waitTime = 0
	for {
		if waitTime > 10 { //尝试多次依然连接不上？
			log.Println(gp.Userid, "Cannot find my keeper,please restart and try again.")
			return ErrCannotConnectKeeper
		}
		if kids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(km.ToString(), ""); kids != nil && err == nil { //先看本地有没有
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
							tempKeeper := &KeeperInfo{
								KeeperID: kid,
							}
							kmKid, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypeKid)
							if err != nil {
								return err
							}
							res, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmKid.ToString(), kid)
							if err == nil && res != nil {
								resStr := string(res)
								splitRes := strings.Split(resStr, metainfo.DELIMITER)
								if len(splitRes) > 3 {
									if splitRes[1] == metainfo.SyncTypeBft {
										tempKeeper.IsBFT = true
									}
								}
							}
							var flag int
							for flag = 0; flag < len(gp.localPeersInfo.Keepers); flag++ {
								if kid == gp.localPeersInfo.Keepers[flag].KeeperID {
									break
								}
							}
							if flag == len(gp.localPeersInfo.Keepers) {
								gp.localPeersInfo.Keepers = append(gp.localPeersInfo.Keepers, tempKeeper)
							}
							if tempKeeper.IsBFT {
								log.Println("Connect to keeper", tempKeeper.KeeperID, "use bft mode")
							} else {
								log.Println("Connect to keeper", tempKeeper.KeeperID, "use simple mode")
							}
						} else {
							log.Println("Connect to keeper", kid, "failed.")
						}

						if len(gp.localPeersInfo.Keepers) >= gp.keeperSLA {
							if gp.KeySet == nil {
								err = gp.loadBLS12Config()
								if err != nil {
									log.Println("Load BLS12 Config error:", err)
									return err
								}
							}
							kmPid, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypePid)
							if err != nil {
								return err
							}
							if pids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmPid.ToString(), ""); pids != nil && err == nil { //只尝试连接一次provider
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
									gp.localPeersInfo.Providers = append(gp.localPeersInfo.Providers, pid)
								}
								log.Println(gp.Userid + ":Group Service is ready")
								err = SetUserState(gp.Userid, GroupStarted)
								if err != nil {
									log.Println("SetUserState failed :", err)
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
			log.Println(gp.Userid, "Cannot find my keeper, retrying...")
			time.Sleep(10 * time.Second)
		}
		waitTime++
	}
}
