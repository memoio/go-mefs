package keeper

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

func (k *Info) checkLedger(ctx context.Context) {
	log.Println("Check Ledger start!")
	time.Sleep(2 * CHALTIME)
	ticker := time.NewTicker(CHECKTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("Repair starts!")
			pus := k.ukpManager.getPUKeys()
			for _, pu := range pus {
				// not repair pos blocks
				if pu.qid == pos.GetPosId() {
					continue
				}

				// only master repair
				if !k.ukpManager.isMasterKeeper(pu.qid, pu.pid) {
					continue
				}

				thisInfo, ok := k.lManager.lMap.Load(pu)
				if !ok {
					continue
				}

				thischalinfo := thisInfo.(*chalinfo)

				thischalinfo.cidMap.Range(func(key, value interface{}) bool {
					thisinfo := value.(*cidInfo)
					eclasped := utils.GetUnixNow() - thisinfo.availtime
					switch thisinfo.repair {
					case 0:
						if EXPIRETIME < eclasped {
							cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
							log.Println("Need repair cid first time: ", cid)
							thisinfo.repair++
							k.repch <- cid
						}
					case 1:
						if 4*EXPIRETIME < eclasped {
							cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
							log.Println("Need repair cid second time: ", cid)
							thisinfo.repair++
							k.repch <- cid
						}
					case 2:
						if 16*EXPIRETIME < eclasped {
							cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
							log.Println("Need repair cid third time: ", cid)
							thisinfo.repair++
							k.repch <- cid
						}
					default:
						// > 30 days; we donnot repair
						if 480*EXPIRETIME >= eclasped {
							// try every 32 hours
							if int64(64*thisinfo.repair-2)*EXPIRETIME < eclasped {
								cid := pu.qid + metainfo.BLOCK_DELIMITER + key.(string)
								log.Println("Need repair cid tried: ", cid)
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

func (k *Info) repairRegular(ctx context.Context) {
	log.Println("Check repairlist start!")
	go func() {
		for {
			select {
			case cid := <-k.repch:
				log.Println("repairing cid: ", cid)
				k.ukpManager.repairBlock(ctx, cid)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// repairBlock works in 3 steps:
// 1.search a new provider,we do it in func SearchNewProvider
// 2.put chunk to this provider
// 3.change metainfo and sync
func (u *ukp) repairBlock(ctx context.Context, blockID string) {

	var response string
	// uid_bid_sid_bid
	blkinfo := strings.Split(blockID, metainfo.BLOCK_DELIMITER)
	if len(blkinfo) < 4 {
		return
	}

	userID := blkinfo[0]

	thisbucket, ok := u.getBucketInfo(userID, blkinfo[1])
	if !ok {
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

		offset := thisinfo.(*cidInfo).offset
		pid := thisinfo.(*cidInfo).storedOn

		// recheck the status
		if strconv.Itoa(i) == blkinfo[3] {
			if thisinfo.(*cidInfo).repair == 0 {
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
		log.Println("Repair: no enough informations")
		return
	}

	if len(response) > 0 {
		if !u.ds.Connect(ctx, response) {
			log.Println("Repair: need choose a new provider to replace old: ", response)
			response = ""
		}
	}

	if len(response) == 0 || response == "" {
		response = u.searchNewProvider(ctx, userID, ugid)
		if response == "" {
			log.Println("Repair failed, no available provider")
			return
		}
	}

	// cid1_pid1_offset1|cid1_pid1_offset1
	metaValue := strings.Join(cpids, metainfo.DELIMITER)

	km, err := metainfo.NewKeyMeta(blockID, metainfo.Repair)
	if err != nil {
		log.Println("construct repair KV error: ", err)
		return
	}

	log.Println("cpids: ", cpids, ",repairs on: ", response)
	_, err = u.ds.SendMetaRequest(context.Background(), int32(metainfo.Get), km.ToString(), []byte(metaValue), nil, response)
	if err != nil {
		log.Println("err: ", err)
	}
}

func (k *Info) handleRepairResult(km *metainfo.KeyMeta, metaValue []byte, provider string) {
	blockID := km.GetMid()
	splitedValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitedValue) != 4 {
		log.Println("handleRepairResponse err: ", metainfo.ErrIllegalValue, metaValue)
		return
	}
	// old provider
	oldPid := splitedValue[2]
	offset, err := strconv.Atoi(splitedValue[3])
	if err != nil {
		log.Println("strconv.Atoi offset error: ", err)
		return
	}

	uid := blockID[:utils.IDLength]
	oldpu := pqKey{
		pid: oldPid,
		qid: uid,
	}

	if strings.Compare(splitedValue[0], "RepairSuccess") == 0 {
		log.Println("repair success, cid is: ", blockID)

		k.deleteBlockMeta(oldpu.qid, blockID)
		k.addBlockMeta(uid, provider, blockID, offset)

		//更新block的meta信息
		kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockPos)
		if err != nil {
			log.Println("construct Syncblock KV error :", err)
			return
		}
		metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)

		err = k.ds.PutKey(context.Background(), kmBlock.ToString(), []byte(metaValue), "local")
		if err != nil {
			log.Println("construct SyncPidsK error :", err)
			return
		}
		return
	}

	log.Println("repair failed, cid is: ", blockID)

	return

}

//searchNewProvider find a NEW provider for user
func (u *ukp) searchNewProvider(ctx context.Context, gid string, ugid []string) string {
	response := ""
	gp, ok := u.getGroupsInfo(gid)
	if !ok {
		log.Println("SearchNewProvider getGroupsInfo() error")
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
			if u.ds.Connect(ctx, tmpPro) {
				response = tmpPro
				break
			}
		}
	}

	return response
}
