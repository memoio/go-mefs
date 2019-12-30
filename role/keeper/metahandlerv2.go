package keeper

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	ds "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage callback
func (k *Info) HandleMetaMessage(opType int, metaKey string, metaValue []byte, from string) ([]byte, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetDType()
	switch dtype {
	case metainfo.UserInit: //user初始化
		go k.handleUserInit(km, from)
	case metainfo.UserNotify: //user初始化确认
		return k.handleUserNotify(km, metaValue, from)
	case metainfo.Contract: //user部署好合约
		go k.handleContracts(km, from)
	case metainfo.BlockPos:
		switch opType {
		case metainfo.Put:
			go k.handleAddBlockPos(km, metaValue, from)
		case metainfo.Delete:
			go k.handleDeleteBlockPos(km)
		}
	case metainfo.Challenge:
		mkey, err := k.ukpManager.getUserBLS12Config(km.GetMid())
		if err != nil {
			go k.lManager.handleProof(km, metaValue, from, mkey)
		}
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

func (k *Info) handleAddBlockPos(km *metainfo.KeyMeta, metaValue []byte, from string) {
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}

	err = k.ds.PutKey(context.Background(), km.ToString(), metaValue, "local")
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

	err = k.addBlockMeta(bm.GetQid(), pid, bm.ToShortStr(), offset)
	if err != nil {
		log.Println("handleBlockPos err: ", err)
	}
	return
}

func (k *Info) handleDeleteBlockPos(km *metainfo.KeyMeta) {
	blockID := km.GetMid()
	bm, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("handleDeleteBlockMeta err: ", err)
		return
	}

	// send to other keepers?

	k.deleteBlockMeta(bm.GetQid(), bm.ToShortStr())
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
