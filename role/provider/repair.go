package provider

import (
	"log"
	"strconv"
	"strings"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	rs "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/role/user"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func handleRepair(km *metainfo.KeyMeta, rpids, keeper string) error {
	var nbid int
	var cids []string
	var ret string
	sig, err := user.BuildSignMessage()
	if err != nil {
		return err
	}
	blockID := km.GetMid()
	userID := blockID[:utils.IDLength]
	tmpPubKey, ok := usersConfigs.Load(userID)
	if !ok {
		tmpPubKey, err := getNewUserConfig(userID, keeper)
		if err != nil {
			log.Println("get new user`s config failed,error :", err)
			return err
		}
		usersConfigs.Store(userID, tmpPubKey)
	}

	pubKey := tmpPubKey.(*mcl.PublicKey)

	cpids := strings.Split(rpids, metainfo.DELIMITER)
	stripe := make([][]byte, len(cpids)-1)
	for _, cpid := range cpids {
		if len(cpid) > 0 {
			splitcpid := strings.Split(cpid, metainfo.REPAIR_DELIMETER)
			cids = append(cids, splitcpid[0])
			if strings.Compare(splitcpid[0], blockID) != 0 {
				pid := splitcpid[1]
				blk, err := localNode.Blocks.GetBlockFrom(localNode.Context(), pid, splitcpid[0], time.Minute, sig)
				if blk != nil && err == nil {
					right := rs.VerifyBlock(blk.RawData(), splitcpid[0], pubKey)
					if right {
						blkMeta, err := metainfo.GetBlockMeta(splitcpid[0])
						if err != nil {
							log.Println("get block meta error :", err)
							return err
						}
						i, err := strconv.Atoi(blkMeta.GetBid())
						if err != nil {
							log.Println("strconv.Atoi error :", err)
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
						log.Println("block rawdata verify failed, error :", err)
						return err
					}
				} else {
					log.Println("GetBlock error :", err)
				}
			} else {
				cidMeta, err := metainfo.GetBlockMeta(blockID)
				if err != nil {
					log.Println("get block meta error :", err)
					return err
				}
				nbid, err = strconv.Atoi(cidMeta.GetBid())
				if err != nil {
					log.Println("strconv.Atoi error :", err)
					return err
				}
				//ret = cid|pid|offset
				ret = splitcpid[0] + metainfo.DELIMITER + splitcpid[1] + metainfo.DELIMITER + splitcpid[2]
				if nbid >= len(stripe) {
					for j := len(stripe); j <= nbid; j++ {
						stripe = append(stripe, nil)
					}
				}
			}
		}
	}

	retKm, err := metainfo.NewKeyMeta(blockID, metainfo.RepairRes)
	if err != nil {
		return err
	}
	newstripe, err := rs.Repair(stripe)
	if err != nil {
		log.Println("repair ", blockID, " failed, error: ", err)
		retMetaValue := RepairFailed + metainfo.DELIMITER + ret
		log.Println("repair response metavalue :", retMetaValue)
		_, err = sendMetaRequest(retKm, retMetaValue, keeper)
		if err != nil {
			return err
		}
		//删除修复时get到的block；
		for _, tmpCid := range cids {
			if len(tmpCid) > 0 {
				temcid := cid.NewCidV2([]byte(tmpCid))
				err = localNode.Blocks.DeleteBlock(temcid)
				if err != nil && err != bs.ErrNotFound {
					log.Println("delete error :", err)
					return err
				}
			}
		}
		return nil
	}
	ncid := cid.NewCidV2([]byte(blockID))
	newblk, err := blocks.NewBlockWithCid(newstripe[nbid], ncid)
	if err != nil {
		log.Println("New block failed, error :", err)
		return err
	}
	//删除修复时get到的block；
	for _, tmpCid := range cids {
		if len(tmpCid) > 0 {
			temcid := cid.NewCidV2([]byte(tmpCid))
			err = localNode.Blocks.DeleteBlock(temcid)
			if err != nil && err != bs.ErrNotFound {
				log.Println("delete error :", err)
				return err
			}
		}
	}
	//把修复好的block放到本地；
	err = localNode.Blocks.PutBlock(newblk)
	if err != nil {
		log.Println("add block failed, error :", err)
		return err
	}
	retMetaValue := RepairSuccess + metainfo.DELIMITER + ret
	log.Println("repair response metavalue :", retMetaValue)
	log.Println("repair success：", blockID)
	_, err = sendMetaRequest(retKm, retMetaValue, keeper)
	if err != nil {
		log.Println("repair response err :", err)
		return err
	}
	return nil
}
