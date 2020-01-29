package keeper

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// HandleMetaMessage callback
func (k *Info) HandleMetaMessage(opType int, metaKey string, metaValue, sig []byte, from string) ([]byte, error) {
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
	case metainfo.UserStart: //user部署好合约
		go k.handleUserStart(km, metaValue, from)
	case metainfo.BlockPos:
		switch opType {
		case metainfo.Put:
			go k.handleAddBlockPos(km, metaValue, from)
		case metainfo.Delete:
			go k.handleDeleteBlockPos(km)
		}
	case metainfo.Challenge:
		if opType == metainfo.Put {
			go k.handleProof(km, metaValue)
		}
	case metainfo.Repair: //provider 修复回复
		if opType == metainfo.Put {
			go k.handleRepairResult(km, metaValue, from)
		}
	case metainfo.Storage:
		go k.handleStorage(km, metaValue, from)
	case metainfo.ExternalAddress:
		return k.handleExternalAddr(km)
	case metainfo.ChalTime:
		return k.handleChalTime(km)
	case metainfo.Pos:
		switch opType {
		case metainfo.Put:
			go k.handlePosAdd(km, metaValue, from)
		case metainfo.Delete:
			go k.handlePosDelete(km, metaValue, from)
		}
	default: //没有匹配的信息，丢弃
		return nil, errors.New("Beyond the capacity")
	}
	return []byte(instance.MetaHandlerComplete), nil
}

// key: blockID/"BlockPos"
// value: pid/offset
func (k *Info) handleAddBlockPos(km *metainfo.KeyMeta, metaValue []byte, from string) {
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

	err = k.ds.PutKey(context.Background(), km.ToString(), metaValue, "local")
	if err != nil {
		utils.MLogger.Info("handleBlockPos err: ", err)
		return
	}

	bids := strings.SplitN(blockID, metainfo.BLOCK_DELIMITER, 2)
	err = k.addBlockMeta(bids[0], bids[1], sValue[0], offset)
	if err != nil {
		utils.MLogger.Info("handleBlockPos err: ", err)
	}
	return
}

func (k *Info) handleDeleteBlockPos(km *metainfo.KeyMeta) {
	utils.MLogger.Info("handleDeleteBlockPos: ", km.ToString())
	blockID := km.GetMid()

	// delete from local
	err := k.ds.DeleteKey(context.Background(), km.ToString(), "local")
	if err != nil {
		utils.MLogger.Info("handleBlockPos err: ", err)
		return
	}

	// delete from mem
	bids := strings.SplitN(blockID, metainfo.BLOCK_DELIMITER, 2)
	// send to other keepers?
	k.deleteBlockMeta(bids[0], bids[1], false)
}

// key: "Storage"/pid; value: total/used
func (k *Info) handleStorage(km *metainfo.KeyMeta, value []byte, pid string) {
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

func (k *Info) handleExternalAddr(km *metainfo.KeyMeta) ([]byte, error) {
	utils.MLogger.Info("handleExternnalAddr: ", km.ToString())
	peerID := km.GetMid()
	return k.ds.GetExternalAddr(peerID)
}

func (k *Info) handleChalTime(km *metainfo.KeyMeta) ([]byte, error) {
	utils.MLogger.Info("handleChalTime: ", km.ToString())

	blockID := km.GetMid()
	if len(blockID) < utils.IDLength {
		return nil, errUnmatchedPeerID
	}

	sValue := strings.SplitN(string(blockID), metainfo.BLOCK_DELIMITER, 2)
	qid := sValue[0]
	bid := sValue[1]
	avail, err := k.getBlockAvail(qid, bid)
	if err != nil {
		return nil, err
	}

	return []byte(utils.UnixToString(avail)), nil
}
