package provider

import (
	"context"
	"strconv"
	"strings"

	rs "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/role/user"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) handleRepair(km *metainfo.KeyMeta, rpids []byte, keeper string) error {
	utils.MLogger.Info("handleRepair: ", km.ToString(), "from: ", keeper)

	var nbid int
	sig, err := user.BuildSignMessage()
	if err != nil {
		return err
	}

	blockID := km.GetMid()
	ops := km.GetOptions()
	if len(ops) < 1 {
		return nil
	}
	blkInfo := strings.Split(blockID, metainfo.BLOCK_DELIMITER)
	if len(blkInfo) < 4 {
		return nil
	}

	userID := ops[0]
	fsID := blkInfo[0]
	chunkID := blkInfo[3]

	pubKey, err := p.getNewUserConfig(userID, fsID)
	if err != nil {
		utils.MLogger.Warn("get new user`s config failed,error :", err)
		return err
	}

	ctx := context.Background()

	cpids := strings.Split(string(rpids), metainfo.DELIMITER)
	stripe := make([][]byte, len(cpids)+1)
	for _, cpid := range cpids {
		if len(cpid) > 0 {
			splitcpid := strings.Split(cpid, metainfo.BLOCK_DELIMITER)
			if len(splitcpid) != 2 {
				continue
			}
			blkInfo[3] = splitcpid[0]
			blkid := strings.Join(blkInfo, metainfo.BLOCK_DELIMITER)
			pid := splitcpid[1]
			if splitcpid[0] != chunkID {
				blk, err := p.ds.GetBlock(ctx, blkid, sig, pid)
				if blk != nil && err == nil {
					right := rs.VerifyBlock(blk.RawData(), blkid, pubKey)
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

	newstripe, off, err := rs.Repair(stripe)
	if err != nil {
		utils.MLogger.Info("repair ", blockID, " failed: ", err)
		return err
	}

	right := rs.VerifyBlock(newstripe[nbid], blockID, pubKey)
	if !right {
		utils.MLogger.Warn("Block %s is not right", blockID)
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

	retMetaValue := "RepairSuccess" + metainfo.DELIMITER + p.localID + metainfo.DELIMITER + strconv.Itoa(off-1)
	_, err = p.ds.SendMetaRequest(context.Background(), int32(metainfo.Put), km.ToString(), []byte(retMetaValue), nil, keeper)
	if err != nil {
		utils.MLogger.Error("repair response err :", err)
		return err
	}
	return nil
}
