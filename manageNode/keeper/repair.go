package keeper

import (
	"context"
	"strconv"
	"strings"
	"time"

	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

func (k *Info) checkLedgerV2(ctx context.Context) {
	utils.MLogger.Info("Check Ledger start!")
	time.Sleep(2 * chalTime)
	ticker := time.NewTicker(chalRepairTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.MLogger.Info("Repair starts!")
			pus := k.getQUKeys()
			for _, pu := range pus {
				// not repair pos blocks
				if pu.uid == pos.GetPosId() {
					continue
				}

				gp := k.getGroupInfo(pu.uid, pu.qid, false)
				if gp == nil {
					continue
				}

				utils.MLogger.Infof("check repair for user %s fsID %s", pu.uid, pu.qid)

				pre := pu.uid + metainfo.BlockDelimiter + pu.qid + metainfo.BlockDelimiter
				bucketNum := gp.bucketNum

				var res strings.Builder
				// superbucket
				for i := 0; i <= int(bucketNum); i++ {
					binfo := gp.getBucketInfo(strconv.Itoa(-i), true)
					if binfo == nil {
						utils.MLogger.Infof("missing bucket %d info", -i)
						continue
					}
					// superbucket 3 chunks and 4k segment
					for j := 0; j < binfo.chunkNum; j++ {
						res.Reset()
						res.WriteString("0")
						res.WriteString(metainfo.BlockDelimiter)
						res.WriteString(strconv.Itoa(j))
						scid := res.String()
						cInfo, ok := binfo.stripes.Load(scid)
						if ok {
							thisinfo := cInfo.(*blockInfo)
							eclasped := time.Now().Unix() - thisinfo.availtime
							switch thisinfo.repair {
							case 0:
								if expireTime < eclasped {
									res.Reset()
									res.WriteString(pre)
									res.WriteString(strconv.Itoa(-i))
									res.WriteString(metainfo.BlockDelimiter)
									res.WriteString(scid)
									cid := res.String()
									utils.MLogger.Info("Need repair cid first time: ", cid)
									thisinfo.repair++
									k.repch <- cid
								}
							case 1:
								if 4*expireTime < eclasped {
									res.Reset()
									res.WriteString(pre)
									res.WriteString(strconv.Itoa(-i))
									res.WriteString(metainfo.BlockDelimiter)
									res.WriteString(scid)
									cid := res.String()
									utils.MLogger.Info("Need repair cid second time: ", cid)
									thisinfo.repair++
									k.repch <- cid
								}
							case 2:
								if 16*expireTime < eclasped {
									res.Reset()
									res.WriteString(pre)
									res.WriteString(strconv.Itoa(-i))
									res.WriteString(metainfo.BlockDelimiter)
									res.WriteString(scid)
									cid := res.String()
									utils.MLogger.Info("Need repair cid third time: ", cid)
									thisinfo.repair++
									k.repch <- cid
								}
							default:
								// > 30 days; we donnot repair
								if 480*expireTime >= eclasped {
									// try every 32 hours
									if int64(64*thisinfo.repair-2)*expireTime < eclasped {
										res.Reset()
										res.WriteString(pre)
										res.WriteString(strconv.Itoa(-i))
										res.WriteString(metainfo.BlockDelimiter)
										res.WriteString(scid)
										cid := res.String()
										utils.MLogger.Info("Need repair cid tried: ", cid)
										thisinfo.repair++
										k.repch <- cid
									}
								}
							}
						}
					}
				}

				// challenge buckets
				for i := 1; i <= int(bucketNum); i++ {
					binfo := gp.getBucketInfo(strconv.Itoa(i), false)
					if binfo == nil {
						utils.MLogger.Infof("missing bucket %d info", i)
						continue
					}

					count := binfo.curStripes
					for j := 0; j <= count; j++ {
						for l := 0; l < binfo.chunkNum; l++ {
							res.Reset()
							res.WriteString(strconv.Itoa(j))
							res.WriteString(metainfo.BlockDelimiter)
							res.WriteString(strconv.Itoa(l))
							scid := res.String()
							cInfo, ok := binfo.stripes.Load(scid)
							if ok {
								thisinfo := cInfo.(*blockInfo)
								eclasped := time.Now().Unix() - thisinfo.availtime
								switch thisinfo.repair {
								case 0:
									if expireTime < eclasped {
										res.Reset()
										res.WriteString(pre)
										res.WriteString(strconv.Itoa(i))
										res.WriteString(metainfo.BlockDelimiter)
										res.WriteString(scid)
										cid := res.String()
										utils.MLogger.Info("Need repair cid first time: ", cid)
										thisinfo.repair++
										k.repch <- cid
									}
								case 1:
									if 4*expireTime < eclasped {
										res.Reset()
										res.WriteString(pre)
										res.WriteString(strconv.Itoa(i))
										res.WriteString(metainfo.BlockDelimiter)
										res.WriteString(scid)
										cid := res.String()
										utils.MLogger.Info("Need repair cid second time: ", cid)
										thisinfo.repair++
										k.repch <- cid
									}
								case 2:
									if 16*expireTime < eclasped {
										res.Reset()
										res.WriteString(pre)
										res.WriteString(strconv.Itoa(i))
										res.WriteString(metainfo.BlockDelimiter)
										res.WriteString(scid)
										cid := res.String()
										utils.MLogger.Info("Need repair cid third time: ", cid)
										thisinfo.repair++
										k.repch <- cid
									}
								default:
									// > 30 days; we donnot repair
									if 480*expireTime >= eclasped {
										// try every 32 hours
										if int64(64*thisinfo.repair-2)*expireTime < eclasped {
											res.Reset()
											res.WriteString(pre)
											res.WriteString(strconv.Itoa(i))
											res.WriteString(metainfo.BlockDelimiter)
											res.WriteString(scid)
											cid := res.String()
											utils.MLogger.Info("Need repair cid tried: ", cid)
											thisinfo.repair++
											k.repch <- cid
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func (k *Info) checkLedger(ctx context.Context) {
	utils.MLogger.Info("Check Ledger start!")
	time.Sleep(2 * chalTime)
	ticker := time.NewTicker(chalRepairTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.MLogger.Info("Repair starts!")
			pus := k.getQUKeys()
			for _, pu := range pus {
				// not repair pos blocks
				if pu.uid == pos.GetPosId() {
					continue
				}

				gp := k.getGroupInfo(pu.uid, pu.qid, false)
				if gp == nil {
					continue
				}

				utils.MLogger.Infof("check repair for user %s fsID ", pu.uid, pu.qid)

				for _, proID := range gp.providers {
					// only master repair
					if !gp.isMaster(proID) {
						utils.MLogger.Debug(proID, " check repair is not msater for user: ", pu.uid)
						continue
					}

					thislinfo := gp.getLInfo(proID, false)
					if thislinfo == nil {
						utils.MLogger.Debug(proID, "check repair has no legerinfo for user: ", pu.uid)
						continue
					}

					nowtime := time.Now().Unix()
					pre := pu.uid + metainfo.BlockDelimiter + pu.qid + metainfo.BlockDelimiter
					thislinfo.blockMap.Range(func(key, value interface{}) bool {
						thisinfo := value.(*blockInfo)
						eclasped := nowtime - thisinfo.availtime
						switch thisinfo.repair {
						case 0:
							if expireTime < eclasped {
								cid := pre + key.(string)
								utils.MLogger.Info("Need repair cid first time: ", cid)
								thisinfo.repair++
								k.repch <- cid
							}
						case 1:
							if 4*expireTime < eclasped {
								cid := pre + key.(string)
								utils.MLogger.Info("Need repair cid second time: ", cid)
								thisinfo.repair++
								k.repch <- cid
							}
						case 2:
							if 16*expireTime < eclasped {
								cid := pre + key.(string)
								utils.MLogger.Info("Need repair cid third time: ", cid)
								thisinfo.repair++
								k.repch <- cid
							}
						default:
							// > 30 days; we donnot repair
							if 480*expireTime >= eclasped {
								// try every 32 hours
								if int64(64*thisinfo.repair-2)*expireTime < eclasped {
									cid := pre + key.(string)
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
// value: chunkID1_pid1/chunkID2_pid2/...
func (k *Info) repairBlock(ctx context.Context, rBlockID string) {
	utils.MLogger.Info("Repair blocks:", rBlockID)
	var response, oldpid string
	// uid_qid_bid_sid_cid
	blkinfo := strings.Split(rBlockID, metainfo.BlockDelimiter)
	if len(blkinfo) < 5 {
		return
	}

	blockID := strings.Join(blkinfo[1:], metainfo.BlockDelimiter)

	uid := blkinfo[0]
	qid := blkinfo[1]

	gp := k.getGroupInfo(uid, qid, false)
	if gp == nil {
		return
	}

	thisbucket := k.getBucketInfo(qid, qid, blkinfo[2], false)
	if thisbucket == nil {
		return
	}

	count := int(thisbucket.chunkNum)
	utils.MLogger.Infof("blockID has %d chunks", thisbucket.chunkNum)

	cpids := make([]string, 0, count)
	ugid := make([]string, 0, count)

	var res strings.Builder
	for i := 0; i < count; i++ {
		res.Reset()
		res.WriteString(blkinfo[3])
		res.WriteString(metainfo.BlockDelimiter)
		res.WriteString(strconv.Itoa(i))
		thisinfo, ok := thisbucket.stripes.Load(res.String())
		if !ok {
			continue
		}

		res.Reset()
		res.WriteString(strconv.Itoa(i))
		res.WriteString(metainfo.BlockDelimiter)

		pid := thisinfo.(*blockInfo).storedOn

		// recheck the status
		if strconv.Itoa(i) == blkinfo[4] {
			if thisinfo.(*blockInfo).repair == 0 {
				return
			}
			response = pid
			oldpid = pid
		}

		res.WriteString(pid)
		cpids = append(cpids, res.String())
		ugid = append(ugid, pid)
	}

	if len(ugid) == 0 {
		utils.MLogger.Infof("Repair %s: no enough informations", rBlockID)
		return
	}

	credit := 0
	if len(response) > 0 {
		proInfo, ok := k.providers.Load(response)
		if ok {
			credit = proInfo.(*pInfo).credit
		}

		if !k.ds.Connect(ctx, response) {
			utils.MLogger.Info("Repair: need choose a new provider to replace old: ", response)
			response = ""
		}
	}

	if credit < 0 || len(response) == 0 || response == "" {
		response = k.searchNewProvider(ctx, qid, ugid)
		if response == "" {
			utils.MLogger.Info("Repair failed, no available provider")
			return
		}
	}

	// cid1_pid1/cid2_pid2
	metaValue := strings.Join(cpids, metainfo.DELIMITER)

	km, err := metainfo.NewKey(blockID, mpb.KeyType_Repair, uid)
	if err != nil {
		utils.MLogger.Info("construct repair KV error: ", err)
		return
	}

	utils.MLogger.Infof("%s has cpids: %s on %s repairs on %s", rBlockID, cpids, oldpid, response)
	k.ds.SendMetaRequest(k.context, int32(mpb.OpType_Get), km.ToString(), []byte(metaValue), nil, response)
}

// key: queryID_bucketID_stripeID_chunkID/"Repair"/uid
// value: "ok"/pid/offset or "fail"
func (k *Info) handleRepairResult(km *metainfo.Key, metaValue []byte, provider string) {
	utils.MLogger.Info("handleRepairResult: ", km.ToString(), " From:", provider)
	blockID := km.GetMid()
	splitedValue := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splitedValue) != 3 {
		return
	}
	splitedKey := strings.SplitN(blockID, metainfo.BlockDelimiter, 2)
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

		k.deleteBlockMeta(qid, bid, true)
		k.addBlockMeta(qid, bid, newPid, newOffset, true)

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
				proInfo, ok := k.providers.Load(tmpPro)
				if ok {
					if proInfo.(*pInfo).credit < 0 {
						continue
					}
				}
				response = tmpPro
				break
			}
		}
	}

	return response
}
