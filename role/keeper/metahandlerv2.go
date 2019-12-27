package keeper

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
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
func (keeper *HandlerV2) HandleMetaMessage(opType int, metaKey string, metaValue []byte, from string) (string, error) {
	km, err := metainfo.GetKeyMeta(metaKey) //注意 这里对metakey已经进行过一次查错,保证传入的数据长度是满足keytype要求的
	if err != nil {
		return "", err
	}
	dtype := km.GetDType()
	switch dtype {
	case metainfo.UserInit: //user初始化
		go handleUserInitReq(km, from)
	case metainfo.UserNotify: //user初始化确认
		return handleNewUserNotif(km, metaValue, from)
	case metainfo.Contract: //user部署好合约
		go handleUserDeloyedContracts(km, from)
	case metainfo.Block: //user删除块
		go handleDeleteBlockMeta(km)
	case metainfo.BlockPos: //user发送块元数据
		go handleBlockMeta(km, metaValue, from)
	case metainfo.Challenge:
		go handleProofResultBls12(km, metaValue, from)
	case metainfo.Repair: //provider 修复回复
		go handleRepairResponse(km, metaValue, from)
	case metainfo.Storage:
		go handleStorageSync(km, metaValue, from)
	case metainfo.ExteralAddress:
		return handlePeerAddr(km)
	case metainfo.Pos:
		go handlePosAdd(km, metaValue, from)
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

	localNode.Data.PutKey(context.Background(), km.ToString(), []byte(response), "local")
	localNode.Data.SendMetaRequest(context.Background(), int32(metainfo.Put), km.ToString(), []byte(response), nil, from)
}

func handleNewUserNotif(km *metainfo.KeyMeta, metaValue []byte, from string) (string, error) {
	log.Println("NewUserNotif", km.ToString(), metaValue, "From:", from)

	ctx := context.Background()

	userID := km.GetMid()

	go fillPinfo(userID, metaValue, from)

	if !localPeerInfo.enableBft {
		resValue := "simple"
		localNode.Data.PutKey(ctx, km.ToString(), []byte(resValue), "local")
		log.Println("use simple mode，userID:", userID)
		return resValue, nil
	}

	// todo
	resValue := "bft|ip"
	localNode.Data.PutKey(ctx, km.ToString(), []byte(resValue), "local")
	log.Println("use bft mode，userID:", userID)
	return resValue, nil
}

func handleUserDeloyedContracts(km *metainfo.KeyMeta, from string) {
	log.Println("NewUserDeployedContracts", km.ToString(), "From:", from)

	ctx := context.Background()

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

func handleBlockMeta(km *metainfo.KeyMeta, metaValue []byte, from string) {
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleBlockMeta err: ", err)
		return
	}

	err = localNode.Data.PutKey(context.Background(), km.ToString(), metaValue, "local")
	if err != nil {
		log.Println("handleBlockMeta err: ", err)
		return
	}

	splitedValue := strings.Split(string(metaValue), metainfo.DELIMITER)
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

func handleStorages(km *metainfo.KeyMeta, value []byte, pid string) {
	vals := strings.Split(string(value), metainfo.DELIMITER)
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

	//获取保存这个块的provider
	metavalueByte, err := localNode.Data.GetKey(context.Background(), kmBlock.ToString(), "local")
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}

	splitedValue := strings.Split(string(metavalueByte), metainfo.DELIMITER)
	if len(splitedValue) < 2 {
		log.Println("handleDeleteBlockMeta err: ", metainfo.ErrIllegalValue)
		return
	}
	err = localNode.Data.DeleteKey(context.Background(), km.ToString(), "local")
	if err != nil && err != ds.ErrNotFound {
		log.Println("handleDeleteBlockMeta err: ", err)
	}
	deleteBlockFromMem(bm.GetUid(), splitedValue[0], blockID)
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
		blockMetaValue, err := localNode.Data.GetKey(kmReq.ToString(), "local")
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
