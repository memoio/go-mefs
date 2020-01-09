package provider

import (
	"context"
	"strconv"
	"strings"

	rs "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/role/user"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (p *Info) handleRepair(km *metainfo.KeyMeta, rpids []byte, keeper string) error {
	var nbid int
	var cids []string
	var ret string
	sig, err := user.BuildSignMessage()
	if err != nil {
		return err
	}

	blockID := km.GetMid()
	userID := blockID[:utils.IDLength]

	pubKey, err := p.getNewUserConfig(userID, keeper)
	if err != nil {
		utils.MLogger.Info("get new user`s config failed,error :", err)
		return err
	}

	ctx := context.Background()

	cpids := strings.Split(string(rpids), metainfo.DELIMITER)
	stripe := make([][]byte, len(cpids)+1)
	for _, cpid := range cpids {
		if len(cpid) > 0 {
			splitcpid := strings.Split(cpid, metainfo.REPAIR_DELIMETER)
			blkid := userID + metainfo.BLOCK_DELIMITER + splitcpid[0]
			cids = append(cids, blkid)
			pid := splitcpid[1]
			if blkid != blockID {
				blk, err := p.ds.GetBlock(ctx, blkid, sig, pid)
				if blk != nil && err == nil {
					right := rs.VerifyBlock(blk.RawData(), blkid, pubKey)
					if right {
						blkMeta, err := metainfo.GetBlockMeta(blkid)
						if err != nil {
							utils.MLogger.Info("get block meta error :", err)
							return err
						}
						i, err := strconv.Atoi(blkMeta.GetBid())
						if err != nil {
							utils.MLogger.Info("strconv.Atoi error :", err)
							return err
						}

						if i >= len(stripe) {
							for j := len(stripe); j <= i; j++ {
								stripe = append(stripe, nil)
							}
						}
						stripe[i] = make([]byte, len(blk.RawData()))
						stripe[i] = blk.RawData()
					} else {
						utils.MLogger.Info("block rawdata verify failed, error :", err)
						return err
					}
				} else {
					utils.MLogger.Info("GetBlock error :", err)
				}
			} else {
				cidMeta, err := metainfo.GetBlockMeta(blockID)
				if err != nil {
					utils.MLogger.Info("get block meta error :", err)
					return err
				}
				nbid, err = strconv.Atoi(cidMeta.GetBid())
				if err != nil {
					utils.MLogger.Info("strconv.Atoi error :", err)
					return err
				}

				if nbid >= len(stripe) {
					for j := len(stripe); j <= nbid; j++ {
						stripe = append(stripe, nil)
					}
				}

				//ret = blockId|pid|offset
				ret = strings.Join(splitcpid[:2], metainfo.DELIMITER)
			}
		}
	}

	retKm, err := metainfo.NewKeyMeta(blockID, metainfo.Repair)
	if err != nil {
		return err
	}
	newstripe, err := rs.Repair(stripe)
	if err != nil {
		utils.MLogger.Info("repair ", blockID, " failed, error: ", err)
		retMetaValue := "RepairFailed" + metainfo.DELIMITER + ret
		utils.MLogger.Info("repair response metavalue :", retMetaValue)
		_, err = p.ds.SendMetaRequest(context.Background(), int32(metainfo.Put), retKm.ToString(), []byte(retMetaValue), nil, keeper)
		if err != nil {
			return err
		}
		//删除修复时get到的block；
		for _, tmpCid := range cids {
			if len(tmpCid) > 0 {
				tmpcid := cid.NewCidV2([]byte(tmpCid))
				err = p.ds.BlockStore().DeleteBlock(tmpcid)
				if err != nil && err != bs.ErrNotFound {
					utils.MLogger.Info("delete error :", err)
					return err
				}
			}
		}
		return nil
	}
	ncid := cid.NewCidV2([]byte(blockID))
	newblk, err := blocks.NewBlockWithCid(newstripe[nbid], ncid)
	if err != nil {
		utils.MLogger.Info("New block failed, error :", err)
		return err
	}
	//删除修复时get到的block；
	for _, tmpCid := range cids {
		if len(tmpCid) > 0 {
			temcid := cid.NewCidV2([]byte(tmpCid))
			err = p.ds.BlockStore().DeleteBlock(temcid)
			if err != nil && err != bs.ErrNotFound {
				utils.MLogger.Info("delete error :", err)
				return err
			}
		}
	}
	//把修复好的block放到本地；
	err = p.ds.BlockStore().Put(newblk)
	if err != nil {
		utils.MLogger.Info("add block failed, error :", err)
		return err
	}
	retMetaValue := "RepairSuccess" + metainfo.DELIMITER + ret
	utils.MLogger.Info("repair response metavalue :", retMetaValue)
	utils.MLogger.Info("repair success：", blockID)
	_, err = p.ds.SendMetaRequest(context.Background(), int32(metainfo.Put), retKm.ToString(), []byte(retMetaValue), nil, keeper)
	if err != nil {
		utils.MLogger.Info("repair response err :", err)
		return err
	}
	return nil
}
