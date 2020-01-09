package keeper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

func (k *Info) checkLedger(ctx context.Context) {
	utils.MLogger.Info("Check Ledger start!")
	time.Sleep(2 * CHALTIME)
	ticker := time.NewTicker(CHECKTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.MLogger.Info("Repair starts!")
			pus := k.getUQKeys()
			for _, pu := range pus {
				// not repair pos blocks
				if pu.uid == pos.GetPosId() {
					continue
				}

				gp := k.getGroupInfo(pu.uid, pu.qid, false)
				if gp == nil {
					continue
				}

				for _, proID := range gp.providers {
					// only master repair
					if !gp.isMaster(proID) {
						continue
					}

					thislinfo := gp.getLInfo(proID, false)
					if thislinfo == nil {
						continue
					}

					thislinfo.blockMap.Range(func(key, value interface{}) bool {
						thisinfo := value.(*blockInfo)
						eclasped := utils.GetUnixNow() - thisinfo.availtime
						switch thisinfo.repair {
						case 0:
							if EXPIRETIME < eclasped {
								cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
								utils.MLogger.Info("Need repair cid first time: ", cid)
								thisinfo.repair++
								k.repch <- cid
							}
						case 1:
							if 4*EXPIRETIME < eclasped {
								cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
								utils.MLogger.Info("Need repair cid second time: ", cid)
								thisinfo.repair++
								k.repch <- cid
							}
						case 2:
							if 16*EXPIRETIME < eclasped {
								cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
								utils.MLogger.Info("Need repair cid third time: ", cid)
								thisinfo.repair++
								k.repch <- cid
							}
						default:
							// > 30 days; we donnot repair
							if 480*EXPIRETIME >= eclasped {
								// try every 32 hours
								if int64(64*thisinfo.repair-2)*EXPIRETIME < eclasped {
									cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
									utils.MLogger.Info("Need repair cid tried: ", cid)
									thisinfo.repair++
									k.repch <- cid
								}
							}
						}

						return true
					})
				}
			}
		}
	}
}

func (k *Info) repairRegular(ctx context.Context) {
	utils.MLogger.Info("Check repairlist start!")
	go func() {
		for {
			select {
			case cid := <-k.repch:
				utils.MLogger.Info("repairing cid: ", cid)
				k.repairBlock(ctx, cid)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// repairBlock works in 3 steps:
// 1.search a new provider,we do it in func SearchNewProvider
// 2.put chunk to this provider
// key: queryID_bucketID_stripeID_chunkID/"Repair"/uid
// value: chunkID1_pid1_offset1/chunkID2_pid2_offset2/...
func (k *Info) repairBlock(ctx context.Context, blockID string) {

	var response string
	// qid_bid_sid_bid
	blkinfo := strings.Split(blockID, metainfo.BLOCK_DELIMITER)
	if len(blkinfo) < 4 {
		return
	}

	qid := blkinfo[0]

	gp := k.getGroupInfo(qid, qid, false)
	if gp == nil {
		return
	}

	thisbucket := k.getBucketInfo(qid, qid, blkinfo[1], false)
	if thisbucket == nil {
		return
	}

	count := int(thisbucket.chunkNum)

	cpids := make([]string, count)
	ugid := make([]string, count)

	var res strings.Builder
	for i := 0; i < count; i++ {
		res.Reset()
		res.WriteString(blkinfo[2])
		res.WriteString(metainfo.BLOCK_DELIMITER)
		res.WriteString(strconv.Itoa(i))
		thisinfo, ok := thisbucket.stripes.Load(res.String())
		if !ok {
			continue
		}

		res.Reset()
		res.WriteString(strconv.Itoa(i))
		res.WriteString(metainfo.BLOCK_DELIMITER)

		offset := thisinfo.(*blockInfo).offset
		pid := thisinfo.(*blockInfo).storedOn

		// recheck the status
		if strconv.Itoa(i) == blkinfo[3] {
			if thisinfo.(*blockInfo).repair == 0 {
				return
			}
			response = pid
		}

		res.WriteString(metainfo.BLOCK_DELIMITER)
		res.WriteString(pid)
		res.WriteString(metainfo.BLOCK_DELIMITER)
		res.WriteString(strconv.Itoa(offset))
		cpids = append(cpids, res.String())
		ugid = append(ugid, pid)
	}

	if len(ugid) == 0 {
		utils.MLogger.Info("Repair: no enough informations")
		return
	}

	if len(response) > 0 {
		if !k.ds.Connect(ctx, response) {
			utils.MLogger.Info("Repair: need choose a new provider to replace old: ", response)
			response = ""
		}
	}

	if len(response) == 0 || response == "" {
		response = k.searchNewProvider(ctx, qid, ugid)
		if response == "" {
			utils.MLogger.Info("Repair failed, no available provider")
			return
		}
	}

	// cid1_pid1_offset1|cid1_pid1_offset1
	metaValue := strings.Join(cpids, metainfo.DELIMITER)

	km, err := metainfo.NewKeyMeta(blockID, metainfo.Repair)
	if err != nil {
		utils.MLogger.Info("construct repair KV error: ", err)
		return
	}

	utils.MLogger.Info("cpids: ", cpids, ",repairs on: ", response)
	_, err = k.ds.SendMetaRequest(context.Background(), int32(metainfo.Get), km.ToString(), []byte(metaValue), nil, response)
	if err != nil {
		utils.MLogger.Info("err: ", err)
	}
}

// key: queryID_bucketID_stripeID_chunkID/"Repair"/uid
// value: "ok" or "fail"/pid/offset
func (k *Info) handleRepairResult(km *metainfo.KeyMeta, metaValue []byte, provider string) {
	blockID := km.GetMid()
	splitedValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitedValue) != 3 {
		return
	}
	splitedKey := strings.SplitN(blockID, metainfo.BLOCK_DELIMITER, 2)
	qid := splitedKey[0]
	bid := splitedKey[1]

	if strings.Compare(splitedValue[0], "ok") == 0 {
		utils.MLogger.Info("repair success, cid is: ", blockID)
		newPid := splitedValue[1]
		newOffset, err := strconv.Atoi(splitedValue[2])
		if err != nil {
			utils.MLogger.Info("strconv.Atoi offset error: ", err)
			return
		}

		splitedValue := strings.Split(string(metaValue), metainfo.DELIMITER)
		if len(splitedValue) != 3 {
			return
		}

		k.deleteBlockMeta(qid, bid, true)
		k.addBlockMeta(qid, bid, newPid, newOffset)

		return
	}

	utils.MLogger.Info("repair failed, block is: ", blockID)

	return

}

//searchNewProvider find a NEW provider for user
func (k *Info) searchNewProvider(ctx context.Context, gid string, ugid []string) string {
	response := ""
	gp := k.getGroupInfo(gid, gid, false)
	if gp == nil {
		return response
	}

	lenp := len(gp.providers)

	if lenp == 0 || lenp <= len(ugid) {
		return response
	}

	tmpProvider := utils.DisorderArray(gp.providers)
	for _, tmpPro := range tmpProvider {
		flag := 0
		for j := 0; j < len(ugid); j++ { //this provider may belong to this stripe already
			if tmpPro != ugid[j] {
				flag++
			}
		}

		if flag == len(ugid) {
			if k.ds.Connect(ctx, tmpPro) {
				response = tmpPro
				break
			}
		}
	}

	return response
}
