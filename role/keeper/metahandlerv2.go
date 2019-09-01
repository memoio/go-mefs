package keeper

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

// KeeperHandlerV2 keeper角色回调接口的实现，
type KeeperHandlerV2 struct {
	Role string
}

// HandleMetaMessage keeper角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (keeper *KeeperHandlerV2) HandleMetaMessage(metaKey, metaValue, from string) (string, error) {
	km, err := metainfo.GetKeyMeta(metaKey) //注意 这里对metakey已经进行过一次查错,保证传入的数据长度是满足keytype要求的
	if err != nil {
		return "", err
	}
	keytype := km.GetKeyType()
	switch keytype {
	case metainfo.UserInitReq: //user初始化
		go handleUserInitReq(km, from)
	case metainfo.UserInitNotif: //user初始化确认
		go handleNewUserNotif(km, metaValue, from)
	case metainfo.UserDeployedContracts: //user部署好合约
		go handleUserDeloyedContracts(km, metaValue, from)
	case metainfo.DeleteBlock: //user删除块
		go handleDeleteBlockMeta(km, from)
	case metainfo.NewKPReq: //user申请新的provider
		return handleNewProviderReq(km, metaValue)
	case metainfo.BlockMetaInfo: //user发送块元数据
		go handleBlockMeta(km, metaValue, from)
	case metainfo.Proof: //provider 挑战回复
		go handleProofResultBls12(km, metaValue, from)
	case metainfo.RepairRes: //provider 修复回复
		go handleRepairResponse(km, metaValue, from)
	case metainfo.Sync: //keeper 同步信息
		go handleSync(km, metaValue, from)
	case metainfo.StorageSync:
		go handleStorageSync(km, metaValue, from)
	case metainfo.Query: //user查询信息
		return handleQueryInfo(km)
	case metainfo.GetPeerAddr:
		return handlePeerAddr(km)
	case metainfo.Test:
		go handleTest(km)
	default: //没有匹配的信息，丢弃
		return "", nil
	}
	return metainfo.MetaHandlerComplete, nil
}

// GetRole 获取这个节点的角色信息，返回错误说明keeper还没有启动好
func (keeper *KeeperHandlerV2) GetRole() (string, error) {
	if keeper == nil {
		return "", ErrKeeperServiceNotReady
	}
	return keeper.Role, nil
}

// handleUserInitReq 收到user发来的初始化请求的回调函数
//返回keeper和provider id组成的字符串，格式为 userID/"UserInitReq"/keepercount/providercount kid1kid2../pid1pid2..
//TODO:可以从合约中查询KUP关系
func handleUserInitReq(km *metainfo.KeyMeta, from string) {
	fmt.Println("handleUserInitReq()", km.ToString(), "From:", from)
	userID := km.GetMid()
	options := km.GetOptions()
	queryAddr := options[2]
	fmt.Println("Query合约信息：", queryAddr)
	var keeperCount, providerCount int
	if queryAddr == contracts.InvalidAddr {
		fmt.Println("没有部署query合约，使用init信息中要求的KP数量")
		ks, err := strconv.Atoi(options[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		ps, err := strconv.Atoi(options[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		keeperCount = ks
		providerCount = ps
	} else {
		fmt.Println("部署过query合约，从合约中查询需求")
		localAddr, _ := ad.GetAddressFromID(localNode.Identity.Pretty())
		config, err := localNode.Repo.Config()
		if err != nil {
			fmt.Println("get config err:", err)
			return
		}
		item, err := contracts.GetQueryInfo(config.Eth, localAddr, common.HexToAddress(queryAddr))
		if item.Completed || err != nil {
			fmt.Println("complete:", item.Completed, "error:", err)
			return
		}
		keeperCount = int(item.KeeperNums)
		providerCount = int(item.ProviderNums)
	}
	fmt.Println("keeperCount:", keeperCount, "providerCount:", providerCount)
	//查询出user的keeper和provider
	//首先看看内存里是否有该节点
	response, err := userInitInMem(userID, keeperCount, providerCount)
	if err != nil { //内存查找出错，在硬盘中找
		response, err = userInitInLocal(userID, keeperCount, providerCount)
		if err != nil { //硬盘查找也出错 就直接返回
			fmt.Println(err)
			return
		}
	}
	if response == "" { //没错，但是结果是空，为新user
		response, err = newUserInit(userID, keeperCount, providerCount)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	km.SetKeyType(metainfo.UserInitRes)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(km.ToString(), response, "local") //在本地保存一份，这里keytype为UserInitRes
	if err != nil {
		fmt.Println(err)
		return
	}
	sendMetaRequest(km, response, from)
}

func handleNewUserNotif(km *metainfo.KeyMeta, metaValue, from string) {
	fmt.Println("NewUserNotif", km.ToString(), metaValue, "From:", from)
	userID := km.GetMid()
	kmKid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		fmt.Println(err)
		return
	}
	kmPid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		fmt.Println(err)
		return
	}

	var keepers []*KeeperInGroup
	var providers []string
	//将value切分，生成好对应的keepers和providers列表
	splited := strings.Split(metaValue, metainfo.DELIMITER)
	kids := splited[0]
	if remain := len(kids) % utils.IDLength; remain != 0 {
		kids = kids[:len(kids)-remain]
	}
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := &KeeperInGroup{
			KID: string(kids[i*utils.IDLength : (i+1)*utils.IDLength]),
		}
		keepers = append(keepers, keeper)
	}
	pids := splited[1]
	if remain := len(pids) % utils.IDLength; remain != 0 {
		pids = pids[:len(pids)-remain]
	}
	for i := 0; i < len(pids)/utils.IDLength; i++ {
		providerID := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
		providers = append(providers, providerID)
	}

	// 收到的信息整理完成，接下来开始分情况填充PInfo,若本节点是第一个收到user信息的，则负责转发

	_, ok := getGroupsInfo(userID)

	if ok { //本地已有保存好的user信息,通知usertendermint的状态
		kmRes, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeBft)
		if err != nil {
			fmt.Println(err)
		}
		var resValue string
		if !localPeerInfo.enableTendermint {
			resValue = "simple"
			fmt.Println("本节点不使用Tendermint，GroupID:", userID)
		}
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmRes.ToString(), resValue, "local") //放在本地供User或Provider启动的时候查询
		if err != nil {
			fmt.Println(err)
		}
		kmRes.SetKeyType(metainfo.UserInitNotifRes)
		_, err = sendMetaRequest(kmRes, resValue, from)
		if err != nil {
			fmt.Println(err)
		}
		return //直接返回

	}
	//没有保存好的user信息，填充Pinfo
	go fillPinfo(userID, keepers, providers, from)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmKid.ToString(), splited[0], "local") //替换本地的User信息
	if err != nil {
		fmt.Println(err)
		return
	}
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmPid.ToString(), splited[1], "local")
	if err != nil {
		fmt.Println(err)
		return
	}
	if _, ok := localPeerInfo.UserCache.Get(userID); ok { //本地没有保存好的user信息 但是waitlist里有
		localPeerInfo.UserCache.Remove(userID)
	}

	//如果ledgerinfo中有该user的信息，则清除。
	LedgerInfo.Range(func(key, value interface{}) bool {
		if key.(PU).uid == userID {
			LedgerInfo.Delete(key)
		}
		return true
	})
}

func handleUserDeloyedContracts(km *metainfo.KeyMeta, metaValue, from string) {
	fmt.Println("NewUserDeployedContracts", km.ToString(), metaValue, "From:", from)
	tempInfo, ok := getGroupsInfo(km.GetMid())
	if !ok {
		fmt.Println("Can't find ", km.GetMid(), "'s GroupInfo")
		return
	}
	err := SaveUpkeeping(tempInfo, km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Upkeeping err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Upkeeping success")
	}
	err = SaveQuery(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Query err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Query success")
	}
}

//handleSync 同步操作的回调函数，同步信息中，第一个option为这个信息的类别，根据信息的类别做不同的同步操作
func handleSync(km *metainfo.KeyMeta, metaValue, from string) {
	options := km.GetOptions()
	if len(options) < 1 {
		fmt.Println("handleSync()error:", metainfo.ErrIllegalKey, km.ToString())
	}
	syncType := options[0]
	var err error
	switch syncType { //TODO:检查参数是否完整
	case metainfo.SyncTypeBlock:
		err = syncBlock(km, metaValue)
	case metainfo.SyncTypeChalPay:
		err = syncChalPay(km, metaValue)
	case metainfo.SyncTypeChalRes:
		err = syncChalres(km, metaValue)
	case metainfo.SyncTypeUID, metainfo.SyncTypePid, metainfo.SyncTypeKid:
		err = syncKUPIDs(km, metaValue)
	default:
		err = ErrorWrongSyncType
	}
	if err != nil {
		fmt.Printf("handleSync()error:%s\nmetakey:%s\nmetavalue:%s\nfrom:%s\n", err, km.ToString(), metaValue, from)
	}
}

func handleBlockMeta(km *metainfo.KeyMeta, metaValue, from string) {
	blockID := km.GetMid()
	if len(blockID) <= utils.IDLength {
		fmt.Println(ErrUnmatchedPeerID)
		return
	}

	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		fmt.Println(err)
		return
	}

	km.SetKeyType(metainfo.Local)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(km.ToString(), metaValue, "local")
	if err != nil {
		fmt.Println(err)
		return
	}

	splitedValue := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splitedValue) < 2 {
		fmt.Println(metainfo.ErrIllegalValue)
		return
	}
	offset, err := strconv.Atoi(splitedValue[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	pid := splitedValue[0]
	if _, ok := localPeerInfo.Credit.Load(pid); !ok {
		tmp := 100
		localPeerInfo.Credit.Store(pid, tmp)
	}

	err = doAddBlocktoLedger(splitedValue[0], bm.GetUid(), blockID, offset)
	if err != nil {
		fmt.Println(err)
	}
	return
}

func handleStorageSync(km *metainfo.KeyMeta, value, pid string) {
	ops := strings.Split(value, metainfo.DELIMITER)
	if len(ops) < 3 {
		return
	}
	tmpmaxSpace := ops[0]
	actulDataSpacestr, err := strconv.ParseUint(ops[1], 10, 64)
	if err != nil {
		fmt.Println("strconv dataSpace error :", err)
		return
	}
	rawDataSpacestr, err := strconv.ParseUint(ops[2], 10, 64)
	if err != nil {
		fmt.Println("strconv rawdataSpace error :", err)
		return
	}
	tmpStorageInfo := &storageInfo{
		maxSpace:       tmpmaxSpace,
		actulDataSpace: actulDataSpacestr,
		rawDataSpace:   rawDataSpacestr,
	}
	localPeerInfo.Storage.Store(pid, tmpStorageInfo)
}

func handleDeleteBlockMeta(km *metainfo.KeyMeta, from string) { //立即删除某些块的元数据
	blockID := km.GetMid()
	if len(blockID) <= utils.IDLength {
		fmt.Println(ErrUnmatchedPeerID)
		return
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = localNode.Routing.(*dht.IpfsDHT).DeleteLocal(kmBlock.ToString())
	if err != nil && err != ds.ErrNotFound {
		fmt.Println(err)
	}
}

func handleNewProviderReq(km *metainfo.KeyMeta, metaValue string) (string, error) {
	userID := km.GetMid()
	options := km.GetOptions()
	thisGroupsInfo, ok := getGroupsInfo(userID)
	if !ok {
		return "", ErrUnmatchedPeerID
	}
	count, err := strconv.Atoi(options[0])
	if err != nil {
		return "", err
	}
	var res string
	var flag int

	if remain := len(metaValue) % utils.IDLength; remain != 0 {
		metaValue = metaValue[:len(metaValue)-remain]
	}

	var providers []string
	for i := 0; i < len(metaValue)/utils.IDLength; i++ {
		provider := string(metaValue[i*utils.IDLength : (i+1)*utils.IDLength])
		//添加到返回值
		providers = append(providers, provider)
	}

	for _, provider := range localPeerInfo.Providers {
		if flag >= count {
			break
		}
		if utils.CheckDup(providers, provider) {
			if sc.ConnectTo(context.Background(), localNode, provider) {
				if utils.CheckDup(thisGroupsInfo.Providers, provider) {
					thisGroupsInfo.Providers = append(thisGroupsInfo.Providers, provider)
				}
				res += provider
				flag++
			}
		}
	}
	return res, nil
}

func handlePeerAddr(km *metainfo.KeyMeta) (string, error) {
	peerID := km.GetMid()
	conns := localNode.PeerHost.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		if pid.Pretty() == peerID {
			addr := c.RemoteMultiaddr()
			fmt.Println(addr.String())
			return addr.String(), nil
		}
	}
	return "", errors.New("Donot have this peer")
}

func handleQueryInfo(km *metainfo.KeyMeta) (string, error) {
	options := km.GetOptions()
	if len(options) < 1 {
		return "", metainfo.ErrIllegalKey
	}
	blockID := km.GetMid()
	queryType := options[0]
	switch queryType {
	case metainfo.QueryTypeLastChal:
		if len(blockID) < utils.IDLength {
			return "", ErrUnmatchedPeerID
		}
		userIDstr := blockID[:utils.IDLength]
		kmReq, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
		if err != nil {
			return "", ErrBlockNotExist
		}
		blockMetaValue, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmReq.ToString(), "local")
		if err != nil || blockMetaValue == nil {
			return "", ErrBlockNotExist
		}
		providerIDspl := strings.Split(string(blockMetaValue), metainfo.DELIMITER)
		if len(providerIDspl) < 1 {
			return "", ErrBlockNotExist
		}
		providerIDstr := providerIDspl[0]
		pu := PU{
			pid: providerIDstr,
			uid: userIDstr,
		}
		if thischalinfo, ok := getChalinfo(pu); ok {
			if thiscidinfo, ok := thischalinfo.Cid.Load(blockID); ok {
				return utils.UnixToString(thiscidinfo.(*cidInfo).availtime), nil
			}
		}

	}
	return "", ErrUnmatchedPeerID
}

func handleTest(km *metainfo.KeyMeta) {
	fmt.Println("测试用回调函数")
	fmt.Println("km.mid:", km.GetMid())
	fmt.Println("km.options", km.GetOptions())
}
