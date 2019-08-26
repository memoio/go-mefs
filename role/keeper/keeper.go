package keeper

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/core"
	pb "github.com/memoio/go-mefs/role/pb"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

type KeeperType uint8

const (
	UnKnow      KeeperType = 0
	IsMaster    KeeperType = 1
	IsNotMaster KeeperType = 2
)

const (
	EXPIRETIME       = int64(30 * 60) //超过这个时间，触发修复，单位：秒
	CHALTIME         = 5 * time.Minute
	CHECKTIME        = 6 * time.Minute
	PERSISTTIME      = 3 * time.Minute
	STORAGESYNCTIME  = 10 * time.Minute
	SPACETIMEPAYTIME = time.Hour
	CONPEERTIME      = 5 * time.Minute
)

//一个组中的keeper信息
type KeeperInGroup struct {
	KID        string
	IP         string
	ID         string
	PubKey     string
	MasterType KeeperType
}

//单个节点“拥有的”K P的对应关系，是对user和本地keeper的描述
type GroupsInfo struct {
	Keepers     []*KeeperInGroup
	Providers   []string
	User        string
	GroupID     string
	LocalKeeper *KeeperInGroup
	upkeeping   contracts.UpKeepingItem
}

//PInfo 存放U-K-P的对应关系，key为userid，value中存放与User相关的Group的信息
//var PInfo map[string]*GroupsInfo

var PInfo sync.Map

//存本节点的相关信息的结构
type PeerInfo struct {
	Keepers          []string
	StoredProviders  []string // 提供存储服务的provider
	Providers        []string
	Storage          sync.Map
	Credit           sync.Map
	UserCache        *lru.Cache //收到Init请求，但未确认的User先记录在这里,长时间不相应则删除user在本地的信息
	enableTendermint bool
	offerBook        sync.Map // 存储连接的provider部署的Offer条约，K-provider的id，V-Offer实例
	queryBook        sync.Map // 存储连接的user部署的Query条约，K-user的id，V-Query实例
}

type storageInfo struct {
	maxSpace       string
	actulDataSpace uint64
	rawDataSpace   uint64
}

var localPeerInfo *PeerInfo

type UserBLS12Config struct {
	PubKey *mcl.PublicKey
}

var localNode *core.MefsNode

//===========================PInfo数据结构操作============================

//getGroupsInfo 从Pinfo中取GropupSinfo 返回时已经类型转换，在代码上显得更简洁一点. 若没取到，返回nil，在调用时需要进行 !ok的判断
func getGroupsInfo(groupid string) (*GroupsInfo, bool) {
	thisGroupinfo, ok := PInfo.Load(groupid)
	if !ok {
		fmt.Println("getGroupsInfo err,groupid:", groupid)
		return nil, false
	}
	out, ok := thisGroupinfo.(*GroupsInfo) //做类型断言的检查，接口的类型转换出错说明数据有问题，报错
	if !ok {
		fmt.Println("thisGroupinfo.(*GroupsInfo) err！", thisGroupinfo)
		return nil, false
	}
	return out, true
}

//getLocalKeeperInGroup 用于获取某个组的本地节点信息，传入参数为组名
//关于LocalKeeper属性，指向本组中本地节点的结构，同一个keeper在不同组中的角色可能不同，避免多次同步请求的重复查找
func getLocalKeeperInGroup(groupid string) (*KeeperInGroup, error) {
	if !IsKeeperServiceRunning() {
		fmt.Println("keeper service not running")
		return nil, ErrKeeperServiceNotReady
	}
	thisGroupInfo, ok := getGroupsInfo(groupid)
	if !ok {
		fmt.Println("getGroupsInfo err! groupid:", groupid)
		return nil, ErrNoGroupsInfo
	}
	if thisGroupInfo.LocalKeeper == nil {
		localID := localNode.Identity.Pretty()
		for _, keeper := range thisGroupInfo.Keepers {
			if strings.Compare(keeper.KID, localID) == 0 {
				thisGroupInfo.LocalKeeper = keeper
			}
		}
	}
	if thisGroupInfo.LocalKeeper == nil {
		return nil, ErrNotKeeperInThisGroup
	}
	return thisGroupInfo.LocalKeeper, nil
}

//localKeeperIsMaster 判断本地节点是否为master节点，取本地节点信息进行判断，若是，返回true，若不是，返回false。若状态未定，则通过相应规则进行判断，并且保存状态
func localKeeperIsMaster(groupid string) bool {
	localKeeper, err := getLocalKeeperInGroup(groupid)
	if err != nil {
		fmt.Println("getLocalKeeperInGroup err.", err)
		return false
	}
	var kidList []string
	if localKeeper.MasterType == UnKnow { //本地节点状态未定,先确定状态
		thisGroupsInfo, ok := getGroupsInfo(groupid)
		if !ok {
			fmt.Println("localkeeperIsMaster err!There is no information in Pinfo,groupid:", groupid)
			return false
		}
		for _, keeper := range thisGroupsInfo.Keepers {
			kidList = append(kidList, keeper.KID)
		}
		masterID := getMasterID(kidList)
		if strings.Compare(masterID, localKeeper.KID) == 0 {
			localKeeper.MasterType = IsMaster
		} else {
			localKeeper.MasterType = IsNotMaster
		}
	}
	return localKeeper.MasterType == IsMaster
}

//getMasterID  根据传入的keeper列表，选出一个master，返回其id
//目前的做法是简单排序后选排在中间的节点，选master策略可以进一步修改
func getMasterID(kidlist []string) string {
	sort.Strings(kidlist)
	return kidlist[len(kidlist)/2]
}

//===============================

//TODO:Keeper出问题重启后，应该能自动将所有user的信息恢复到内存中
func StartKeeperService(ctx context.Context, node *core.MefsNode, enableTendermint bool) error {
	//初始化各类结构体
	localNode = node
	userCache, err := lru.New(100)
	if err != nil {
		return err
	}
	var credit, storage sync.Map
	localPeerInfo = &PeerInfo{
		UserCache: userCache,
		Credit:    credit,
		Storage:   storage,
	}

	err = loadAllUser() //加载本地保存的数据
	if err != nil {
		localNode = nil
		localPeerInfo = nil
		return err
	}
	fmt.Println("Keeper Service is ready")
	err = SearchAllKeepersAndProviders(ctx) //连接节点
	if err != nil {
		fmt.Println("SearchAllKeepersAndProviders()err:", err)
		localNode = nil
		localPeerInfo = nil
		return err
	}
	//tendermint启动相关
	localPeerInfo.enableTendermint = enableTendermint
	if !localPeerInfo.enableTendermint {
		fmt.Println("本节点不使用Tendermint")
	}

	go persistLocalPeerInfoRegular(ctx)
	go challengeRegular(ctx) //挑战
	go cleanTestUsersRegular(ctx)
	go checkrepairlist(ctx)
	go checkLedger(ctx)
	go spaceTimePayRegular(ctx)
	go checkStorage(ctx)
	go checkPeers(ctx)
	fmt.Println("StartKeeperService() complete!")
	return nil
}

func persistLocalPeerInfoRegular(ctx context.Context) {
	fmt.Println("persistLocalPeerInfoRegular() start!")
	ticker := time.NewTicker(PERSISTTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := PersistlocalPeerInfo()
			if err != nil {
				log.Println("PersistlocalPeerInfo err:", err)
			}
		}
	}
}

func PersistlocalPeerInfo() error { //每次退出前将现有的本地PeerInfo持久化一次新的PeersInfo
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	}
	localID := localNode.Identity.Pretty() //本地id

	kmKid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return err
	}
	kmPid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return err
	}
	kmUid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeUID)
	if err != nil {
		return err
	}

	kmLedger, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeLedger)
	if err != nil {
		return err
	}
	kmCredit, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeCredit)
	if err != nil {
		return err
	}

	var kids bytes.Buffer
	var pids bytes.Buffer
	var users bytes.Buffer
	//var ledgers bytes.Buffer

	for i := 0; i < len(localPeerInfo.Keepers); i++ {
		kids.WriteString(localPeerInfo.Keepers[i])
	} //整理连接的keeper 信息
	for i := 0; i < len(localPeerInfo.Providers); i++ {
		pids.WriteString(localPeerInfo.Providers[i])
	} //整理连接的provider信息

	tmpCredit := make(map[string]uint32)
	localPeerInfo.Credit.Range(func(key, value interface{}) bool {
		pid := key.(string)
		sco := value.(int)
		tmpCredit[pid] = uint32(sco)
		return true
	})
	creditProto := &pb.Credit{
		Scores: tmpCredit,
	}
	creditByte, err := proto.Marshal(creditProto)
	if err != nil {
		fmt.Println("Credit marshal failed")
	}

	PInfo.Range(func(uid, groupsinfo interface{}) bool { //循环PInfo整理连接的user信息
		thisuid, ok := uid.(string)
		if ok { //类型断言检查
			if thisuid != localID {
				users.WriteString(thisuid)
			}
			return true
		}
		fmt.Println("uid.(string) false!uid:", uid)
		return false
	})

	tmpLedgerinfo := make(map[string]*pb.Chalin)
	var ledgerByte []byte

	LedgerInfo.Range(func(k, v interface{}) bool {
		pu := k.(PU)
		thischalinfo := v.(*chalinfo)
		tmpCid := make(map[string]*pb.Cidin)
		puProto := &pb.Pu{
			Provider: pu.pid,
			User:     pu.uid,
		}
		puByte, err := proto.Marshal(puProto) //*格式修改
		if err != nil {
			fmt.Println("proto.Marshal error:", err)
		}
		thischalinfo.Cid.Range(func(k, v interface{}) bool {
			tmpCidin := &pb.Cidin{
				Res:      v.(*cidInfo).res,
				Repair:   v.(*cidInfo).repair,
				Offset:   int64(v.(*cidInfo).offset),
				Avaltime: utils.UnixToString(v.(*cidInfo).availtime),
			}
			tmpCid[k.(string)] = tmpCidin
			return true
		})
		chalinProto := &pb.Chalin{
			Cidin:     tmpCid,
			Maxlength: thischalinfo.maxlength,
		}
		tmpLedgerinfo[string(puByte)] = chalinProto
		ledgerin := &pb.LedgerInfo{
			Chalinfo: tmpLedgerinfo,
		}
		ledgerByte, err = proto.Marshal(ledgerin) //*格式修改
		if err != nil {
			fmt.Println("proto.Marshal error:", err)
		}
		return true
	})

	//整理好的元数据KV信息 存入持久化介质
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmUid.ToString(), users.String(), "local")
	if err != nil {
		return err
	}
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmKid.ToString(), kids.String(), "local")
	if err != nil {
		return err
	}
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmPid.ToString(), pids.String(), "local")
	if err != nil {
		return err
	}

	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmLedger.ToString(), string(ledgerByte), "local")
	if err != nil {
		return err
	}
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmCredit.ToString(), string(creditByte), "local")
	if err != nil {
		return nil
	}

	kids.Reset()
	pids.Reset()
	users.Reset() //释放buffer
	//ledgers.Reset()
	return nil
}

//此函数仅在内测阶段需要，会在每天 1~5点期间，将测试User的信息删掉
func cleanTestUsersRegular(ctx context.Context) {
	fmt.Println("cleanTestUsersRegular() start!")
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tNow := time.Now()
			t1 := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 1, 0, 0, 0, tNow.Location())
			t2 := t1.Add(4 * time.Hour)
			//在一点和五点之间，清理testUsers
			if tNow.After(t1) && tNow.Before(t2) {
				fmt.Println("Begin to clean testUser")
				go func() {
					cleanTestUsers()
				}()
			}
		}
	}
}

func cleanTestUsers() {
	testUsers := make(map[PU]struct{})
	LedgerInfo.Range(func(k, v interface{}) bool { //对PU对进行循环
		pu := k.(PU)
		//没有部署合约的定期清理数据
		addr, err := address.GetAddressFromID(pu.uid)
		if err != nil {
			return true
		}
		config, err := localNode.Repo.Config()
		if err != nil {
			return true
		}
		endPoint := config.Eth //获取endPoint
		_, _, err = contracts.GetUKFromResolver(endPoint, addr)
		//部署过合约的不清理
		if err != contracts.ErrNotDeployedMapper && err != contracts.ErrNotDeployedUk {
			return true
		}
		fmt.Println(pu.uid, "is a test User, clean its data")
		testUsers[pu] = struct{}{}
		thischalinfo := v.(*chalinfo)
		thischalinfo.Cid.Range(func(key, value interface{}) bool { //对该PU对中provider保存的块循环
			blockID := key.(string)
			fmt.Println("Delete testUser block-", blockID)
			//先通知Provider删除块
			km, err := metainfo.NewKeyMeta(blockID, metainfo.DeleteBlock)
			if err != nil {
				fmt.Println("construct delete block KV error :", err)
				return false
			}
			_, err = sendMetaRequest(km, "", pu.pid)
			if err != nil {
				retryCount := 3
				for i := 0; i < retryCount; i++ {
					_, err = sendMetaRequest(km, "", pu.pid)
					if err == nil {
						break
					}
				}
				if err != nil {
					fmt.Println("Delete testUser block failed-", blockID, "error:", err)
				}
			}

			//再在本地删除记录
			kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
			if err != nil {
				fmt.Println("NewKeyMeta()error!", err, "blockID:", blockID)
			}
			err = localNode.Routing.(*dht.IpfsDHT).DeleteLocal(kmBlock.ToString())
			if err != nil {
				log.Println("Delete local Message error:", err)
			}
			return true
		})
		return true
	})
	//将其从账本中删除
	for pu := range testUsers {
		kmKid, err := metainfo.NewKeyMeta(pu.uid, metainfo.Local, metainfo.SyncTypeKid)
		if err != nil {
			return
		}
		err = localNode.Routing.(*dht.IpfsDHT).DeleteLocal(kmKid.ToString())
		if err != nil {
			log.Println("Delete local Message error:", err)
		}
		kmPid, err := metainfo.NewKeyMeta(pu.uid, metainfo.Local, metainfo.SyncTypePid)
		if err != nil {
			return
		}
		err = localNode.Routing.(*dht.IpfsDHT).DeleteLocal(kmPid.ToString())
		if err != nil {
			log.Println("Delete local Message error:", err)
		}
		PInfo.Delete(pu.uid)
		LedgerInfo.Delete(pu)
	}
}

//重启后重新恢复User现场 读取本地存储的U-K-P信息，构建PInfo结构
func loadAllUser() error {
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	}
	fmt.Println("loadAllUser")
	localID := localNode.Identity.Pretty() //本地id

	kmUid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeUID)
	if err != nil {
		return err
	}
	usersLocal := kmUid.ToString()

	kmLedger, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeLedger)
	if err != nil {
		return err
	}
	kmCredit, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeCredit)
	if err != nil {
		return err
	}

	//将硬盘中保存的K、U、P信息取出，形成PInfo结构。
	if users, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(usersLocal, "local"); users != nil && err == nil {
		if remain := len(users) % utils.IDLength; remain != 0 {
			users = users[:len(users)-remain]
		}
		fmt.Println("Load-User", string(users))
		for i := 0; i < len(users)/utils.IDLength; i++ { //对user进行循环，逐个恢复user信息
			userID := string(users[i*utils.IDLength : (i+1)*utils.IDLength])
			var userPeersInfo GroupsInfo
			PInfo.Store(userID, &userPeersInfo)
			kmKid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeKid)
			if err != nil {
				return err
			}
			kmPid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypePid)
			if err != nil {
				return err
			}
			userkidsMeta := kmKid.ToString()
			userpidsMeta := kmPid.ToString()

			//填写peersinfo.keepers信息
			//TODO:检查连接性，但由于还没写没连接上该怎么处理的逻辑，先不检查
			if userKids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userkidsMeta, "local"); userKids != nil && err == nil {
				if remain := len(userKids) % utils.IDLength; remain != 0 {
					userKids = userKids[:len(userKids)-remain]
				}
				for i := 0; i < len(userKids)/utils.IDLength; i++ {
					keeperid := string(userKids[i*utils.IDLength : (i+1)*utils.IDLength])
					keeper := &KeeperInGroup{
						KID: keeperid,
					}
					userPeersInfo.Keepers = append(userPeersInfo.Keepers, keeper)
				}
			}

			//填写peersinfo.providers信息
			if userPids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userpidsMeta, "local"); userPids != nil && err == nil {
				if remain := len(userPids) % utils.IDLength; remain != 0 {
					userPids = userPids[:len(userPids)-remain]
				}
				for i := 0; i < len(userPids)/utils.IDLength; i++ {
					provider := string(userPids[i*utils.IDLength : (i+1)*utils.IDLength])
					userPeersInfo.Providers = append(userPeersInfo.Providers, provider)
				}
			}
			// 保存Upkeeping信息
			err = SaveUpkeeping(&userPeersInfo, userID)
			if err != nil {
				fmt.Println("Save ", userID, "'s Upkeeping error in loadAllUser", err)
			} else {
				fmt.Println("Save ", userID, "'s Upkeeping success in loadAllUser")
			}
			// 保存Query信息
			err = SaveQuery(userID)
			if err != nil {
				fmt.Println("Save ", userID, "'s Query error in loadAllUser", err)
			} else {
				fmt.Println("Save ", userID, "'s Query success in loadAllUser")
			}
			// 保存Offer信息
			for _, provider := range userPeersInfo.Providers {
				err = SaveOffer(provider)
				if err != nil {
					fmt.Println("Save ", provider, "'s Offer error in loadAllUser", err)
				} else {
					fmt.Println("Save ", provider, "'s Offer success in loadAllUser")
				}
			}
		}
	}

	//取硬盘中的LedgerInfo结构
	if ledgers, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmLedger.ToString(), "local"); ledgers != nil && err == nil {
		LedgerinProto := &pb.LedgerInfo{}
		err = proto.Unmarshal(ledgers, LedgerinProto)
		if err != nil {
			return err
		}
		for pustr, thischalinfoinProto := range LedgerinProto.Chalinfo {
			puinProto := &pb.Pu{}
			err = proto.Unmarshal([]byte(pustr), puinProto)
			if err != nil {
				return err
			}
			newpu := PU{
				pid: puinProto.Provider,
				uid: puinProto.User,
			}
			var cidMap, timeMap sync.Map
			for blockid, thiscidinfoinProto := range thischalinfoinProto.Cidin {
				newcidinfo := &cidInfo{
					res:       thiscidinfoinProto.Res,
					repair:    thiscidinfoinProto.Repair,
					availtime: utils.StringToUnix(thiscidinfoinProto.Avaltime),
					offset:    int(thiscidinfoinProto.Offset),
				}
				cidMap.Store(blockid, newcidinfo)
			}
			newchalinfo := &chalinfo{
				Cid:       cidMap,
				Time:      timeMap,
				maxlength: thischalinfoinProto.Maxlength,
			}
			LedgerInfo.Store(newpu, newchalinfo)
		}
	}

	if credits, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmCredit.ToString(), "local"); credits != nil && err == nil {
		creditProto := &pb.Credit{}
		err = proto.Unmarshal(credits, creditProto)
		if err != nil {
			return err
		}
		for pid, cre := range creditProto.Scores {
			localPeerInfo.Credit.Store(pid, int(cre))
		}
	}
	return nil
}

func SearchAllKeepersAndProviders(ctx context.Context) error {
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	} //只有角色为Keeper才传递

	loadKnownKeepersAndProviders(ctx) //先加载持久化的keeper和Provider看看，有助于快速恢复
	//go newConnPeerRole(PeerIDch, ctx) //此协程不断处理新连接的节点
	err := checkConnectedPeer() //查看当前连接的节点的角色
	fmt.Println("checkConnectedPeer()complete")
	if err != nil {
		return err
	}
	return nil
}

//查找本地持久化保存的U-K-P信息，并与这些节点尝试连接
func loadKnownKeepersAndProviders(ctx context.Context) error {
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	}
	localID := localNode.Identity.Pretty() //本地id
	kmKid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return err
	}
	kmPid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return err
	}
	//尝试链接持久化保存的keeper信息
	if kids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmKid.ToString(), "local"); kids != nil && err == nil {
		if remain := len(kids) % utils.IDLength; remain != 0 {
			kids = kids[:len(kids)-remain]
		}
		for i := 0; i < len(kids)/utils.IDLength; i++ {
			kid := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
			if sc.ConnectTo(ctx, localNode, kid) {
				var j int
				for j = 0; j < len(localPeerInfo.Keepers); j++ { //看本地已有此节点记录
					if kid == localPeerInfo.Keepers[j] {
						break
					}
				}
				if j == len(localPeerInfo.Keepers) {
					log.Println("Connect to known keeper: ", kid)
					localPeerInfo.Keepers = append(localPeerInfo.Keepers, kid)
				}
			}
		}
	} //连接其他keeper的过程

	if pids, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmPid.ToString(), "local"); pids != nil && err == nil {
		if remain := len(pids) % utils.IDLength; remain != 0 {
			pids = pids[:len(pids)-remain]
		}

		for i := 0; i < len(pids)/utils.IDLength; i++ {
			pid := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
			if sc.ConnectTo(ctx, localNode, pid) {
				var j int
				for j = 0; j < len(localPeerInfo.Providers); j++ {
					if pid == localPeerInfo.Providers[j] { //不要重复了
						break
					}
				}
				if j == len(localPeerInfo.Providers) {
					log.Println("Connect to known provider: ", pid)
					localPeerInfo.Providers = append(localPeerInfo.Providers, pid)
				}
			}
		}
	} //连接其他provider的过程
	return nil
}

func checkConnectedPeer() error {
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	}

	fmt.Println("checkConnectedPeer")

	localID := localNode.Identity.Pretty() //本地id

	kmKid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return err
	}
	kmPid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return err
	}
	connPeers := localNode.PeerHost.Network().Peers() //the list of peers we are connected to

	for _, ID := range connPeers {
		id := ID.Pretty() //连接结点id的base58编码
		fmt.Println("try to get roleinfo from: ", id)
		kmRole, err := metainfo.NewKeyMeta(id, metainfo.Local, metainfo.SyncTypeRole)
		if err != nil {
			return err
		}
		val, _ := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmRole.ToString(), id) //全网查该节点的角色
		if string(val) == metainfo.RoleKeeper {
			var i int
			for i = 0; i < len(localPeerInfo.Keepers); i++ { //看本地已有此节点记录
				if id == localPeerInfo.Keepers[i] {
					break
				}
			}
			if i == len(localPeerInfo.Keepers) {
				log.Println("Connect to connected keeper: ", id)
				localPeerInfo.Keepers = append(localPeerInfo.Keepers, id)
				err := localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmKid.ToString(), id, "local") //把当前连接的所有keepers信息存到本地的leveldb中
				if err != nil {
					return err
				}
			}
		} else if string(val) == metainfo.RoleProvider {
			var i int
			for i = 0; i < len(localPeerInfo.Providers); i++ {
				if id == localPeerInfo.Providers[i] {
					break
				}
			}
			if i == len(localPeerInfo.Providers) {
				log.Println("Connect to connected provider: ", id)
				localPeerInfo.Providers = append(localPeerInfo.Providers, id)
				err := localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmPid.ToString(), id, "local") //把当前连接的所有providers信息存到本地的leveldb中
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func newConnPeerRole(peerIDch chan string, ctx context.Context) error { //处理新连接的节点
	if !IsKeeperServiceRunning() {
		return ErrKeeperServiceNotReady
	}
	localID := localNode.Identity.Pretty() //本地id
	kmKid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return err
	}
	kmPid, err := metainfo.NewKeyMeta(localID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return err
	}

	for {
		select {
		case id := <-peerIDch:
			kmRole, err := metainfo.NewKeyMeta(id, metainfo.Local, metainfo.SyncTypeRole)
			if err != nil {
				return err
			}
			val, _ := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmRole.ToString(), id) //查询节点角色
			if string(val) == metainfo.RoleKeeper {
				var i int
				for i = 0; i < len(localPeerInfo.Keepers); i++ { //看本地已有此节点记录
					if id == localPeerInfo.Keepers[i] {
						break
					}
				}
				if i == len(localPeerInfo.Keepers) {
					log.Println("Connect to new connect keeper: ", id)
					localPeerInfo.Keepers = append(localPeerInfo.Keepers, id)
					err := localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmKid.ToString(), id, "local")
					if err != nil {
						log.Println("Append keeper meta error:", err)
					}
				}
			} else if string(val) == metainfo.RoleProvider {
				var i int
				for i = 0; i < len(localPeerInfo.Providers); i++ {
					if id == localPeerInfo.Providers[i] {
						break
					}
				}
				if i == len(localPeerInfo.Providers) {
					log.Println("Connect to new connect provider: ", id)
					err := SaveOffer(id)
					if err != nil {
						fmt.Println("Save ", id, "'s Offer err in newConnPeerRole", err)
					} else {
						fmt.Println("Save ", id, "'s Offer success in newConnPeerRole")
					}
					localPeerInfo.Providers = append(localPeerInfo.Providers, id)
					err = localNode.Routing.(*dht.IpfsDHT).CmdAppendTo(kmPid.ToString(), id, "local")
					if err != nil {
						log.Println("Append provider meta error:", err)
					}
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func IsKeeperServiceRunning() bool {
	return localNode != nil && localPeerInfo != nil
}

func checkStorage(ctx context.Context) {
	fmt.Println("CheckStorage() start!")
	ticker := time.NewTicker(STORAGESYNCTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				for _, pid := range localPeerInfo.Providers {
					km, err := metainfo.NewKeyMeta(pid, metainfo.StorageSync)
					if err != nil {
						fmt.Println("construct Storage sync KV error :", err)
						return
					}
					_, err = sendMetaRequest(km, "", pid)
					if err != nil {
						log.Println("sendMetaRequest error:", err)
					}
				}
			}()
		}
	}
}

func checkPeers(ctx context.Context) {
	fmt.Println("CheckConnectedPeer() start!")
	ticker := time.NewTicker(CONPEERTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				checkConnectedPeer()
			}()
		}
	}
}
