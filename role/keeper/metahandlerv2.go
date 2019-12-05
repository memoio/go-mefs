package keeper

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

// HandlerV2 keeper角色回调接口的实现，
type HandlerV2 struct {
	Role string
}

// HandleMetaMessage keeper角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (keeper *HandlerV2) HandleMetaMessage(metaKey, metaValue, from string) (string, error) {
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
		go handleDeleteBlockMeta(km)
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
	case metainfo.PosAdd:
		go handlePosAdd(km, metaValue, from)
	case metainfo.PosDelete:
		go handlePosDelete(km, metaValue, from)
	case metainfo.Test:
		go handleTest(km)
	default: //没有匹配的信息，丢弃
		return "", nil
	}
	return metainfo.MetaHandlerComplete, nil
}

// GetRole 获取这个节点的角色信息，返回错误说明keeper还没有启动好
func (keeper *HandlerV2) GetRole() (string, error) {
	if keeper == nil {
		return "", errKeeperServiceNotReady
	}
	return keeper.Role, nil
}

// handleUserInitReq 收到user发来的初始化请求的回调函数
//return kv, key: userID/"UserInitReq"/keepercount/providercount; value:kid1kid2../pid1pid2..
func handleUserInitReq(km *metainfo.KeyMeta, from string) {
	log.Println("handleUserInitReq: ", km.ToString(), " From: ", from)
	userID := km.GetMid()
	options := km.GetOptions()
	if len(options) < 3 {
		return
	}

	queryAddr := options[2]
	log.Println("Query合约信息：", queryAddr)
	var keeperCount, providerCount int
	var price int64
	var response string
	if queryAddr == contracts.InvalidAddr {
		log.Println("No query contracts，use k/p numbers in init request")
		ks, err := strconv.Atoi(options[0])
		if err != nil {
			log.Println("handleUserInitReq: ", err)
			return
		}
		ps, err := strconv.Atoi(options[1])
		if err != nil {
			log.Println("handleUserInitReq: ", err)
			return
		}
		keeperCount = ks
		providerCount = ps
		price = int64(utils.STOREPRICEPEDOLLAR)
		response, err = initUser(userID, keeperCount, providerCount, price)
		if err != nil {
			if err != nil { //硬盘查找也出错 就直接返回
				log.Println("handleUserInitReq err: ", err)
				return
			}
		}
	} else {
		log.Println("Get k/p numbers from query contract of user: ", userID)
		localAddr, _ := ad.GetAddressFromID(localNode.Identity.Pretty())
		item, err := contracts.GetQueryInfo(localAddr, common.HexToAddress(queryAddr))
		if item.Completed || err != nil {
			log.Println("complete:", item.Completed, "error:", err)
			return
		}
		keeperCount = int(item.KeeperNums)
		providerCount = int(item.ProviderNums)
		price = item.Price
		if pos.GetPosId() == userID {
			price = int64(utils.STOREPRICEPEDOLLAR)
		}

		response, err = userNewInit(userID, keeperCount, providerCount, price)
		if err != nil {
			if err != nil { //硬盘查找也出错 就直接返回
				log.Println("handleUserInitReq err: ", err)
				return
			}
		}
	}
	log.Println(userID, " keeperCount: ", keeperCount, "providerCount: ", providerCount, "price: ", price)
	//查询出user的keeper和provider
	//首先看看内存里是否有该节点

	km.SetKeyType(metainfo.UserInitRes)
	putKeyTo(km.ToString(), response, "local")
	sendMetaRequest(km, response, from)
}

func handleNewUserNotif(km *metainfo.KeyMeta, metaValue, from string) {
	log.Println("NewUserNotif", km.ToString(), metaValue, "From:", from)
	userID := km.GetMid()

	var keepers []string
	var providers []string
	//将value切分，生成好对应的keepers和providers列表
	splited := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splited) < 2 {
		log.Println("handleNewUserNotif value is not correct: ", metaValue)
		return
	}
	kids := splited[0]
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(keeper)
		if err != nil {
			continue
		}
		keepers = append(keepers, keeper)
	}
	pids := splited[1]
	for i := 0; i < len(pids)/utils.IDLength; i++ {
		providerID := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(providerID)
		if err != nil {
			continue
		}
		providers = append(providers, providerID)
	}

	go fillPinfo(userID, keepers, providers, from)

	//如果ledgerinfo中有该user的信息，则清除。
}

func handleUserDeloyedContracts(km *metainfo.KeyMeta, metaValue, from string) {
	log.Println("NewUserDeployedContracts", km.ToString(), metaValue, "From:", from)

	userID := km.GetMid()

	err := saveUpkeeping(userID, false)
	if err != nil {
		log.Println("Save ", userID, "'s Upkeeping err", err)
	}
	log.Println("Save ", userID, "'s Upkeeping success")

	err = saveQuery(userID, false)
	if err != nil {
		log.Println("Save ", userID, "'s Query err", err)
	}
	log.Println("Save ", userID, "'s Query success")

	_, err = getUserBLS12Config(userID)
	if err != nil {
		log.Println("Save ", userID, "'s BLSconfig err", err)
	}
	log.Println("Save ", userID, "'s BLSconfig success")
}

//handleSync 同步操作的回调函数，同步信息中，第一个option为这个信息的类别，根据信息的类别做不同的同步操作
func handleSync(km *metainfo.KeyMeta, metaValue, from string) {
	options := km.GetOptions()
	if len(options) < 1 {
		log.Println("handleSync()error:", metainfo.ErrIllegalKey, km.ToString())
	}
	syncType := options[0]
	var err error
	switch syncType { //TODO:检查参数是否完整
	case metainfo.SyncTypeBlock:
		err = handleSyncBlock(km, metaValue)
	case metainfo.SyncTypeChalPay:
		err = handleSyncChalPay(km, metaValue)
	case metainfo.SyncTypeChalRes:
		err = handleSyncChalres(km, metaValue)
	case metainfo.SyncTypeUID, metainfo.SyncTypePid, metainfo.SyncTypeKid:
		err = syncKUPIDs(km, metaValue)
	default:
		err = errorWrongSyncType
	}
	if err != nil {
		log.Printf("handleSync()error:%s\nmetakey:%s\nmetavalue:%s\nfrom:%s\n", err, km.ToString(), metaValue, from)
	}
}

func handleBlockMeta(km *metainfo.KeyMeta, metaValue, from string) {
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleBlockMeta err: ", err)
		return
	}

	km.SetKeyType(metainfo.Local)
	err = putKeyTo(km.ToString(), metaValue, "local")
	if err != nil {
		log.Println("handleBlockMeta err: ", err)
		return
	}

	splitedValue := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splitedValue) < 2 {
		log.Println("handleBlockMeta err: ", metainfo.ErrIllegalValue)
		return
	}
	offset, err := strconv.Atoi(splitedValue[1])
	if err != nil {
		log.Println("handleBlockMeta err: ", err)
		return
	}
	pid := splitedValue[0]
	addCredit(pid)

	err = addBlocktoMem(bm.GetUid(), pid, blockID, offset)
	if err != nil {
		log.Println("handleBlockMeta err: ", err)
	}
	return
}

func handleStorageSync(km *metainfo.KeyMeta, value, pid string) {
	vals := strings.Split(value, metainfo.DELIMITER)
	if len(vals) < 2 {
		return
	}

	total, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		log.Println("handleStorageSync err: ", err)
		return
	}
	used, err := strconv.ParseUint(vals[1], 10, 64)
	if err != nil {
		log.Println("handleStorageSync err: ", err)
		return
	}

	thisInfo, err := getPInfo(pid)
	if err != nil {
		return
	}
	thisInfo.maxSpace = total
	thisInfo.usedSpace = used
}

func handleDeleteBlockMeta(km *metainfo.KeyMeta) { //立即删除某些块的元数据 由user发送给所有keeper
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}
	kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}

	//获取保存这个块的provider
	metavalueByte, err := getKeyFrom(kmBlock.ToString(), "local")
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}
	splitedValue := strings.Split(string(metavalueByte), metainfo.DELIMITER)
	if len(splitedValue) < 2 {
		log.Println("handleDeleteBlockMeta err: ", metainfo.ErrIllegalValue)
		return
	}
	err = deleteFrom(kmBlock.ToString(), "local")
	if err != nil && err != ds.ErrNotFound {
		log.Println("handleDeleteBlockMeta err: ", err)
	}
	deleteBlockFromMem(bm.GetUid(), splitedValue[0], blockID)
}

func handleNewProviderReq(km *metainfo.KeyMeta, metaValue string) (string, error) {
	userID := km.GetMid()
	options := km.GetOptions()
	_, ok := getGroupsInfo(userID)
	if !ok {
		return "", errUnmatchedPeerID
	}
	_, err := strconv.Atoi(options[0])
	if err != nil {
		return "", err
	}
	var res string

	if remain := len(metaValue) % utils.IDLength; remain != 0 {
		metaValue = metaValue[:len(metaValue)-remain]
	}

	for i := 0; i < len(metaValue)/utils.IDLength; i++ {
		res += string(metaValue[i*utils.IDLength : (i+1)*utils.IDLength])
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
			log.Println("handlePeerAddr: ", addr.String())
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
		log.Println("handle get last challenge time")
		if len(blockID) < utils.IDLength {
			return "", errUnmatchedPeerID
		}
		userIDstr := blockID[:utils.IDLength]
		kmReq, err := metainfo.NewKeyMeta(blockID, metainfo.Local, metainfo.SyncTypeBlock)
		if err != nil {
			return "", errBlockNotExist
		}
		blockMetaValue, err := getKeyFrom(kmReq.ToString(), "local")
		if err != nil || blockMetaValue == nil {
			return "", errBlockNotExist
		}
		providerIDspl := strings.Split(string(blockMetaValue), metainfo.DELIMITER)
		if len(providerIDspl) < 1 {
			return "", errBlockNotExist
		}
		providerIDstr := providerIDspl[0]
		pu := puKey{
			pid: providerIDstr,
			uid: userIDstr,
		}

		cidString, err := metainfo.GetCidFromBlock(blockID)
		if err != nil {
			return "", err
		}

		if thischalinfo, ok := getChalinfo(pu); ok {
			if thiscidinfo, ok := thischalinfo.cidMap.Load(cidString); ok {
				return utils.UnixToString(thiscidinfo.(*cidInfo).availtime), nil
			}
		}

	}
	return "", errBlockNotExist
}

func handleTest(km *metainfo.KeyMeta) {
	log.Println("测试用回调函数")
	log.Println("km.mid:", km.GetMid())
	log.Println("km.options", km.GetOptions())
}
