package keeper

import (
	"context"
	"strconv"
	"strings"

	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage callback
func (k *Info) HandleMetaMessage(opType mpb.OpType, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
	km, err := metainfo.NewKeyFromString(metaKey)
	if err != nil {
		return nil, err
	}
	dtype := km.GetKType()
	switch dtype {
	case mpb.KeyType_UserInit: //user初始化
		go k.handleUserInit(km, from)
	case mpb.KeyType_UserNotify: //user初始化确认
		return k.handleUserNotify(km, metaValue, from)
	case mpb.KeyType_UserStart: //user部署好合约
		return k.handleUserStart(km, metaValue, sig, from)
	case mpb.KeyType_UserStop: //user部署好合约
		go k.handleUserStop(km, metaValue, from)
	case mpb.KeyType_HeartBeat:
		go k.handleHeartBeat(km, metaValue, from)
	case mpb.KeyType_BlockPos:
		switch opType {
		case mpb.OpType_Put:
			go k.handleAddBlockPos(km, metaValue, from)
		case mpb.OpType_Get:
			return k.handleGetKey(km, metaValue, sig, from)
		case mpb.OpType_Delete:
			go k.handleDeleteBlockPos(km)
		}
	case mpb.KeyType_Challenge:
		if opType == mpb.OpType_Put {
			go k.handleProof(km, metaValue)
		}
	case mpb.KeyType_Repair: //provider 修复回复
		if opType == mpb.OpType_Put {
			go k.handleRepairResult(km, metaValue, from)
		}
	case mpb.KeyType_Storage:
		go k.handleStorage(km, metaValue, from)
	case mpb.KeyType_ExternalAddress:
		return k.handleExternalAddr(km)
	case mpb.KeyType_ChalTime:
		return k.handleChalTime(km)
	case mpb.KeyType_Pos:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePosAdd(km, metaValue, from)
		case mpb.OpType_Delete:
			go k.handlePosDelete(km, metaValue, from)
		}
	default:
		switch opType {
		case mpb.OpType_Put:
			go k.handlePutKey(km, metaValue, sig, from)
		case mpb.OpType_Get:
			return k.handleGetKey(km, metaValue, sig, from)
		case mpb.OpType_Delete:
			go k.handleDeleteKey(km, metaValue, sig, from)
		default:
			return nil, metainfo.ErrWrongType
		}
	}
	return []byte(instance.MetaHandlerComplete), nil
}

func (k *Info) handlePutKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handlePutKey: ", km.ToString())
	ctx := context.Background()
	ok := k.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	k.ds.PutKey(ctx, km.ToString(), metaValue, sig, "local")
}

func (k *Info) handleGetKey(km *metainfo.Key, metaValue, sig []byte, from string) ([]byte, error) {
	utils.MLogger.Info("handleGetKey: ", km.ToString())
	ctx := context.Background()

	return k.ds.GetKey(ctx, km.ToString(), "local")
}

func (k *Info) handleDeleteKey(km *metainfo.Key, metaValue, sig []byte, from string) {
	utils.MLogger.Info("handleDeleteKey: ", km.ToString())
	ctx := context.Background()
	ok := k.ds.VerifyKey(ctx, km.ToString(), metaValue, sig)
	if !ok {
		return
	}

	k.ds.DeleteKey(ctx, km.ToString(), "local")
}

// key: blockID/"BlockPos"
// value: pid/offset
func (k *Info) handleAddBlockPos(km *metainfo.Key, metaValue []byte, from string) {
	utils.MLogger.Info("handleAddBlockPos: ", km.ToString())

	blockID := km.GetMid()

	sValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(sValue) != 2 {
		utils.MLogger.Info("handleBlockPos err: ", metainfo.ErrIllegalValue)
		return
	}
	offset, err := strconv.Atoi(sValue[1])
	if err != nil {
		utils.MLogger.Info("handleBlockPos err: ", err)
		return
	}

	bids := strings.SplitN(blockID, metainfo.BlockDelimiter, 2)
	err = k.addBlockMeta(bids[0], bids[1], sValue[0], offset, true)
	if err != nil {
		utils.MLogger.Error("handleBlockPos err: ", err)
	}
	return
}

func (k *Info) handleDeleteBlockPos(km *metainfo.Key) {
	utils.MLogger.Info("handleDeleteBlockPos: ", km.ToString())
	blockID := km.GetMid()

	// delete from local
	k.ds.DeleteKey(context.Background(), km.ToString(), "local")

	// delete from mem
	bids := strings.SplitN(blockID, metainfo.BlockDelimiter, 2)
	// send to other keepers?
	k.deleteBlockMeta(bids[0], bids[1], false)
}

// key: "Storage"/pid; value: total/used
func (k *Info) handleStorage(km *metainfo.Key, value []byte, pid string) {
	utils.MLogger.Info("handleStorage: ", km.ToString())
	vals := strings.Split(string(value), metainfo.DELIMITER)
	if len(vals) < 2 {
		return
	}

	total, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		utils.MLogger.Info("handleStorageSync err: ", err)
		return
	}

	used, err := strconv.ParseUint(vals[1], 10, 64)
	if err != nil {
		utils.MLogger.Info("handleStorageSync err: ", err)
		return
	}

	thisInfo, err := k.getPInfo(pid)
	if err != nil {
		return
	}
	thisInfo.maxSpace = total
	thisInfo.usedSpace = used
}

func (k *Info) handleExternalAddr(km *metainfo.Key) ([]byte, error) {
	utils.MLogger.Info("handleExternnalAddr: ", km.ToString())
	peerID := km.GetMid()
	addr, err := k.ds.GetExternalAddr(peerID)
	if err != nil {
		return nil, err
	}
	return addr.Bytes(), nil
}

func (k *Info) handleChalTime(km *metainfo.Key) ([]byte, error) {
	utils.MLogger.Info("handleChalTime: ", km.ToString())

	blockID := km.GetMid()
	if len(blockID) < utils.IDLength {
		return nil, errUnmatchedPeerID
	}

	sValue := strings.SplitN(string(blockID), metainfo.BlockDelimiter, 2)
	qid := sValue[0]
	bid := sValue[1]
	avail, err := k.getBlockAvail(qid, bid)
	if err != nil {
		return nil, err
	}

	return []byte(utils.UnixToString(avail)), nil
}
