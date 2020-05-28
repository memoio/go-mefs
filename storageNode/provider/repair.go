package provider

import (
	"strconv"
	"strings"

	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) handleRepair(km *metainfo.Key, rpids []byte, keeper string) error {
	utils.MLogger.Info("handleRepair: ", km.ToString(), " from: ", keeper)

	blockID := km.GetMid()
	ops := km.GetOptions()
	if len(ops) < 2 {
		return role.ErrWrongKey
	}

	blkInfo := strings.Split(blockID, metainfo.BlockDelimiter)
	if len(blkInfo) < 4 {
		return role.ErrWrongKey
	}

	userID := ops[0]
	fsID := blkInfo[0]
	chunkID := blkInfo[3]

	pubKey, err := p.getNewUserConfig(userID, fsID)
	if err != nil {
		utils.MLogger.Warn("get new user`s config failed,error :", err)
		return err
	}

	ctx := p.context

	bid, err := metainfo.NewKey(blockID, mpb.KeyType_Block, "0", ops[1])
	if err != nil {
		return err
	}

	segNeed, err := strconv.Atoi(ops[1])
	if err != nil {
		return err
	}
	block, err := p.ds.GetBlock(ctx, bid.ToString(), nil, "local")
	if err == nil {
		ok, err := df.VerifyBlockLength(block.RawData(), 0, segNeed)
		if ok {
			ok = df.VerifyBlock(block.RawData(), blockID, pubKey)
			if ok {
				retMetaValue := "ok" + metainfo.DELIMITER + p.localID + metainfo.DELIMITER + ops[1]
				_, err = p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), km.ToString(), []byte(retMetaValue), nil, keeper)
				if err != nil {
					utils.MLogger.Error("repair response err :", err)
					return err
				}
				return nil
			}
		}
	}

	var nbid int
	sig, err := role.BuildSignMessage()
	if err != nil {
		return err
	}

	cpids := strings.Split(string(rpids), metainfo.DELIMITER)
	stripe := make([][]byte, len(cpids)+1)
	for _, cpid := range cpids {
		if len(cpid) > 0 {
			splitcpid := strings.Split(cpid, metainfo.BlockDelimiter)
			if len(splitcpid) != 2 {
				continue
			}
			blkInfo[3] = splitcpid[0]
			blkid := strings.Join(blkInfo, metainfo.BlockDelimiter)

			bid, err := metainfo.NewKey(blkid, mpb.KeyType_Block, "0", ops[1])
			if err != nil {
				continue
			}

			if splitcpid[0] == chunkID {
				nbid, err = strconv.Atoi(chunkID)
				if err != nil {
					utils.MLogger.Info("strconv.Atoi error :", err)
					return err
				}

				if nbid >= len(stripe) {
					for j := len(stripe); j <= nbid; j++ {
						stripe = append(stripe, nil)
					}
				}
				continue
			}

			pid := splitcpid[1]
			blk, err := p.ds.GetBlock(ctx, bid.ToString(), sig, pid)
			if err != nil || blk == nil {
				continue
			}
			ok, err := df.VerifyBlockLength(blk.RawData(), 0, segNeed)
			if err != nil || !ok {
				continue
			}

			ok = df.VerifyBlock(blk.RawData(), blkid, pubKey)
			if !ok {
				continue
			}
			chNum, err := strconv.Atoi(splitcpid[0])
			if err != nil {
				utils.MLogger.Info("strconv.Atoi error :", err)
				return err
			}

			if chNum >= len(stripe) {
				for j := len(stripe); j <= chNum; j++ {
					stripe = append(stripe, nil)
				}
			}
			stripe[chNum] = blk.RawData()
			p.ds.DeleteBlock(ctx, blkid, "local")
		}
	}

	newstripe, off, err := df.Repair(stripe)
	if err != nil {
		utils.MLogger.Info("repair ", blockID, " failed: ", err)
		return err
	}

	if off != segNeed {
		utils.MLogger.Warnf("Block %s length is not right, need %d, but got %d", blockID, segNeed, off)
		return role.ErrEmptyData
	}

	ok, err := df.VerifyBlockLength(newstripe[nbid], 0, segNeed)
	if err != nil || !ok {
		utils.MLogger.Warnf("Block %s length is not right", blockID)
		return err
	}

	ok = df.VerifyBlock(newstripe[nbid], blockID, pubKey)
	if !ok {
		utils.MLogger.Warnf("Block %s is not right", blockID)
		return nil
	}

	err = p.ds.PutBlock(p.context, blockID, newstripe[nbid], "local")
	if err != nil {
		utils.MLogger.Error("put block to local failed, error : ", err)
		return err
	}

	utils.MLogger.Info("repair success: ", blockID)

	retMetaValue := "ok" + metainfo.DELIMITER + p.localID + metainfo.DELIMITER + strconv.Itoa(off)
	_, err = p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), km.ToString(), []byte(retMetaValue), nil, keeper)
	if err != nil {
		utils.MLogger.Error("repair response err :", err)
		return err
	}
	return nil
}
