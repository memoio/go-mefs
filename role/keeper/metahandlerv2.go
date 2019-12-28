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
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

// HandlerV2 keeper角色回调接口的实现，
type HandlerV2 struct {
	Role string
}

// HandleMetaMessage callback
func (keeper *HandlerV2) HandleMetaMessage(opType int, metaKey string, metaValue []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetDType()
	switch dtype {
	case metainfo.UserInit: //user初始化
		go handleUserInit(km, from)
	case metainfo.UserNotify: //user初始化确认
		return handleUserNotify(km, metaValue, from)
	case metainfo.Contract: //user部署好合约
		go handleUserContracts(km, from)
	case metainfo.BlockPos:
		switch opType {
		case metainfo.Put:
			go handleAddBlockPos(km, metaValue, from)
		case metainfo.Delete:
			go handleDeleteBlockPos(km)
		}
	case metainfo.Challenge:
		go handleProof(km, metaValue, from)
	case metainfo.Repair: //provider 修复回复
		go handleRepairResult(km, metaValue, from)
	case metainfo.Storage:
		go handleStorage(km, metaValue, from)
	case metainfo.ExternalAddress:
		return handleExternalAddr(km)
	case metainfo.ChalTime:
		return handleChalTime(km)
	case metainfo.Pos:
		switch opType {
		case metainfo.Put:
			go handlePosAdd(km, metaValue, from)
		case metainfo.Delete:
			go handlePosDelete(km, metaValue, from)
		}
	default: //没有匹配的信息，丢弃
		return nil, errors.New("Beyond the capacity")
	}
	return []byte(instance.MetaHandlerComplete), nil
}

// GetRole 获取这个节点的角色信息，返回错误说明keeper还没有启动好
func (keeper *HandlerV2) GetRole() (string, error) {
	if keeper == nil {
		return "", errKeeperServiceNotReady
	}
	return keeper.Role, nil
}

// handleUserInit collect keepers and providers for user
// return kv, key: "UserInit"/userID/queryAddr/keepercount/providercount;
// value: kid1kid2../pid1pid2..
func handleUserInit(km *metainfo.KeyMeta, from string) {
	log.Println("NewUserInit: ", km.ToString(), " From: ", from)
	userID := km.GetMid()
	options := km.GetOptions()
	if len(options) < 3 {
		return
	}

	queryAddr := options[0]
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
	log.Println("New user: ", userID, " keeperCount: ", keeperCount, "providerCount: ", providerCount, "price: ", price)

	localNode.Data.PutKey(context.Background(), km.ToString(), []byte(response), "local")
	localNode.Data.SendMetaRequest(context.Background(), int32(metainfo.Put), km.ToString(), []byte(response), nil, from)
}

func handleUserNotify(km *metainfo.KeyMeta, metaValue []byte, from string) ([]byte, error) {
	log.Println("NewUserNotify: ", km.ToString(), metaValue, "From:", from)

	ctx := context.Background()

	userID := km.GetMid()

	go fillPinfo(userID, metaValue, from)

	var res []byte
	if !localPeerInfo.enableBft {
		res = []byte("simple")
		localNode.Data.PutKey(ctx, km.ToString(), res, "local")
		log.Println("use simple mode，userID:", userID)
		return res, nil
	}

	// todo
	res = []byte("bft|ip")
	localNode.Data.PutKey(ctx, km.ToString(), res, "local")
	log.Println("use bft mode，userID:", userID)
	return res, nil
}

func handleUserContracts(km *metainfo.KeyMeta, from string) {
	log.Println("NewUserDeployedContracts", km.ToString(), "From:", from)
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

func handleAddBlockPos(km *metainfo.KeyMeta, metaValue []byte, from string) {
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}

	err = localNode.Data.PutKey(context.Background(), km.ToString(), metaValue, "local")
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}

	splitedValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitedValue) < 2 {
		log.Println("handleBlockPos err: ", metainfo.ErrIllegalValue)
		return
	}
	offset, err := strconv.Atoi(splitedValue[1])
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}
	pid := splitedValue[0]
	addCredit(pid)

	err = addBlocktoMem(bm.GetUid(), pid, blockID, offset)
	if err != nil {
		log.Println("handleBlockPos err: ", err)
	}
	return
}

func handleDeleteBlockPos(km *metainfo.KeyMeta) {
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}

	ctx := context.Background()
	value, err := localNode.Data.GetKey(ctx, km.ToString(), "local")
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}

	splitedValue := strings.Split(string(value), metainfo.DELIMITER)
	if len(splitedValue) < 2 {
		log.Println("handleDeleteBlockMeta err: ", metainfo.ErrIllegalValue)
		return
	}
	err = localNode.Data.DeleteKey(ctx, km.ToString(), "local")
	if err != nil && err != ds.ErrNotFound {
		log.Println("handleDeleteBlockMeta err: ", err)
	}

	// send to other keepers?

	deleteBlockFromMem(bm.GetUid(), splitedValue[0], blockID)
}

func handleStorage(km *metainfo.KeyMeta, value []byte, pid string) {
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

func handleExternalAddr(km *metainfo.KeyMeta) ([]byte, error) {
	peerID := km.GetMid()
	conns := localNode.PeerHost.Network().Conns()
	for _, c := range conns {
		pid := c.RemotePeer()
		if pid.Pretty() == peerID {
			addr := c.RemoteMultiaddr()
			log.Println("handlePeerAddr: ", addr.String())
			return addr.Bytes(), nil
		}
	}
	return nil, errors.New("Donot have this peer")
}

func handleChalTime(km *metainfo.KeyMeta) ([]byte, error) {
	blockID := km.GetMid()
	log.Println("handle get last challenge time of block: ", blockID)
	if len(blockID) < utils.IDLength {
		return nil, errUnmatchedPeerID
	}
	userIDstr := blockID[:utils.IDLength]
	kmReq, err := metainfo.NewKeyMeta(blockID, metainfo.Block)
	if err != nil {
		return nil, errBlockNotExist
	}

	value, err := localNode.Data.GetKey(context.Background(), kmReq.ToString(), "local")
	if err != nil || value == nil {
		return nil, errBlockNotExist
	}

	proIDspl := strings.Split(string(value), metainfo.DELIMITER)
	if len(proIDspl) < 2 {
		return nil, errBlockNotExist
	}
	proID := proIDspl[0]
	pu := puKey{
		pid: proID,
		uid: userIDstr,
	}

	cidString, err := metainfo.GetCidFromBlock(blockID)
	if err != nil {
		return nil, err
	}

	if thischalinfo, ok := getChalinfo(pu); ok {
		if thiscidinfo, ok := thischalinfo.cidMap.Load(cidString); ok {
			return []byte(utils.UnixToString(thiscidinfo.(*cidInfo).availtime)), nil
		}
	}
	return nil, errBlockNotExist
}
