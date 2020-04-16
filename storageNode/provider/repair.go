package provider

import (
	"strconv"
	"strings"

	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
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

	block, err := p.ds.GetBlock(ctx, bid.ToString(), nil, "local")
	if err == nil {
		ok := df.VerifyBlock(block.RawData(), blockID, pubKey)
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
			pid := splitcpid[1]
			if splitcpid[0] != chunkID {
				blk, err := p.ds.GetBlock(ctx, blkid, sig, pid)
				if blk != nil && err == nil {
					right := df.VerifyBlock(blk.RawData(), blkid, pubKey)
					if right {
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
					} else {
						utils.MLogger.Warn("block verify failed, error: ", err)
					}
				} else {
					utils.MLogger.Warn("GetBlock error :", err)
				}
			} else {
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
			}
		}
	}

	newstripe, off, err := df.Repair(stripe)
	if err != nil {
		utils.MLogger.Info("repair ", blockID, " failed: ", err)
		return err
	}

	right := df.VerifyBlock(newstripe[nbid], blockID, pubKey)
	if !right {
		utils.MLogger.Warnf("Block %s is not right", blockID)
		return nil
	}

	ncid := cid.NewCidV2([]byte(blockID))
	newblk, err := blocks.NewBlockWithCid(newstripe[nbid], ncid)
	if err != nil {
		utils.MLogger.Error("New block failed, error:", err)
		return err
	}
	err = p.ds.BlockStore().Put(newblk)
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
