package keeper

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

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
		if err != nil {
			go k.handleProof(km, metaValue)
		}
	case metainfo.Repair: //provider 修复回复
		go k.handleRepairResult(km, metaValue, from)
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
	blockID := km.GetMid()

	err := k.ds.PutKey(context.Background(), km.ToString(), metaValue, "local")
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}

	sValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(sValue) < 2 {
		log.Println("handleBlockPos err: ", metainfo.ErrIllegalValue)
		return
	}
	offset, err := strconv.Atoi(sValue[1])
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}

	bids := strings.SplitN(blockID, metainfo.BLOCK_DELIMITER, 2)
	err = k.addBlockMeta(bids[0], bids[1], sValue[0], offset)
	if err != nil {
		log.Println("handleBlockPos err: ", err)
	}
	return
}

func (k *Info) handleDeleteBlockPos(km *metainfo.KeyMeta) {
	blockID := km.GetMid()

	// delete from local
	err := k.ds.DeleteKey(context.Background(), km.ToString(), "local")
	if err != nil {
		log.Println("handleBlockPos err: ", err)
		return
	}

	// delete from mem
	bids := strings.SplitN(blockID, metainfo.BLOCK_DELIMITER, 2)
	// send to other keepers?
	k.deleteBlockMeta(bids[0], bids[1], false)
}

// key: "Storage"/pid; value: total/used
func (k *Info) handleStorage(km *metainfo.KeyMeta, value []byte, pid string) {
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

	thisInfo, err := k.getPInfo(pid)
	if err != nil {
		return
	}
	thisInfo.maxSpace = total
	thisInfo.usedSpace = used
}

func (k *Info) handleExternalAddr(km *metainfo.KeyMeta) ([]byte, error) {
	peerID := km.GetMid()
	return k.ds.GetExternalAddr(peerID)
}

func (k *Info) handleChalTime(km *metainfo.KeyMeta) ([]byte, error) {
	blockID := km.GetMid()
	log.Println("handle get last challenge time of block: ", blockID)
	if len(blockID) < utils.IDLength {
		return nil, errUnmatchedPeerID
	}

	sValue := strings.SplitN(string(blockID), metainfo.BLOCK_DELIMITER, 2)
	qid := sValue[0]
	bid := sValue[1]
	pid, err := k.getBlockPos(qid, bid)
	if err != nil {
		return nil, err
	}

	thisl := k.getLInfo(qid, qid, pid, false)
	if thisl != nil {
		thiscidinfo, ok := thisl.blockMap.Load(bid)
		if ok {
			return []byte(utils.UnixToString(thiscidinfo.(*blockInfo).availtime)), nil
		}
	}

	return nil, errBlockNotExist
}
