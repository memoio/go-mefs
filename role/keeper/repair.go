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
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

var repch chan string

func checkLedger(ctx context.Context) {
	log.Println("Check Ledger start!")
	time.Sleep(2 * CHALTIME)
	ticker := time.NewTicker(CHECKTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			doCheckLedgerForRepair()
		}
	}
}

func doCheckLedgerForRepair() {
	pus := getPUKeysFromukpInfo()
	for _, pu := range pus {
		// not repair pos blocks
		if pu.uid == pos.GetPosId() {
			continue
		}

		// only master repair
		if !isMasterKeeper(pu.uid, pu.pid) {
			continue
		}

		thisInfo, ok := ledgerInfo.Load(pu)
		if !ok {
			continue
		}

		thischalinfo := thisInfo.(*chalinfo)

		thischalinfo.cidMap.Range(func(key, value interface{}) bool {
			eclasped := utils.GetUnixNow() - value.(*cidInfo).availtime
			switch value.(*cidInfo).repair {
			case 0:
				if EXPIRETIME < eclasped {
					log.Println("Need repair cid first time: ", key.(string))
					value.(*cidInfo).repair++
					repch <- (pu.uid + metainfo.BLOCK_DELIMITER + key.(string))
				}
			case 1:
				if 4*EXPIRETIME < eclasped {
					log.Println("Need repair cid second time: ", key.(string))
					value.(*cidInfo).repair++
					repch <- (pu.uid + metainfo.BLOCK_DELIMITER + key.(string))
				}
			case 2:
				if 16*EXPIRETIME < eclasped {
					log.Println("Need repair cid third time: ", key.(string))
					value.(*cidInfo).repair++
					repch <- (pu.uid + metainfo.BLOCK_DELIMITER + key.(string))
				}
			default:
				// > 30 days; we donnot repair
				if 480*EXPIRETIME >= eclasped {
					// try every 32 hours
					if int64(64*value.(*cidInfo).repair-2)*EXPIRETIME < eclasped {
						repch <- (pu.uid + metainfo.BLOCK_DELIMITER + key.(string))
						value.(*cidInfo).repair++
					}

				}
			}

			return true
		})
	}
}

func checkrepairlist(ctx context.Context) {
	log.Println("Check repairlist start!")
	repch = make(chan string, 1024)
	go func() {
		for {
			select {
			case cid := <-repch:
				log.Println("repairing cid: ", cid)
				repairBlock(ctx, cid)
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
func repairBlock(ctx context.Context, blockID string) {
	var cpids, ugid []string
	var response string

	blkinfo := strings.Split(blockID, metainfo.BLOCK_DELIMITER)
	if len(blkinfo) < 4 {
		return
	}

	userID := blkinfo[0]

	thisbucket, ok := getBucketInfo(userID, blkinfo[1])
	if !ok {
		return
	}

	cidPrefix := strings.Join(blkinfo[1:2], metainfo.BLOCK_DELIMITER)

	for i := 0; i < int(thisbucket.dataCount+thisbucket.parityCount); i++ {
		blockid := cidPrefix + metainfo.BLOCK_DELIMITER + strconv.Itoa(i)
		thisinfo, ok := thisbucket.stripes.Load(blockid)
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

		cpids = append(cpids, blockid+metainfo.REPAIR_DELIMETER+pid+metainfo.REPAIR_DELIMETER+strconv.Itoa(offset))
		ugid = append(ugid, pid)
	}

	if len(response) > 0 {
		if !sc.ConnectTo(ctx, localNode, response) {
			response = ""
		}
	}

	if response == "" {
		response = SearchNewProvider(ctx, userID, ugid)
		if response == "" {
			log.Println("Repair failed, no extra provider")
			return
		}
	}

	// cid1/pid1/offset1|cid1/pid1/offset1
	metaValue := strings.Join(cpids, metainfo.DELIMITER)

	km, err := metainfo.NewKeyMeta(blockID, metainfo.Repair)
	if err != nil {
		log.Println("construct repair KV error: ", err)
		return
	}

	log.Println("cpids: ", cpids, "\nrpids: ", metaValue, " \nrepairs on: ", response)
	_, err = sendMetaRequest(km, metaValue, response)
	if err != nil {
		log.Println("err: ", err)
	}
}

func handleRepairResponse(km *metainfo.KeyMeta, metaValue, provider string) {
	blockID := km.GetMid()
	splitedValue := strings.Split(metaValue, metainfo.DELIMITER)
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
	oldpu := puKey{
		pid: oldPid,
		uid: uid,
	}

	if strings.Compare(splitedValue[0], "RepairSuccess") == 0 {
		log.Println("repair success, cid is: ", blockID)
		newcidinfo := &cidInfo{
			repair:    0,
			availtime: utils.GetUnixNow(),
			offset:    offset,
			storedOn:  provider,
		}

		addCidinfotoMem(uid, provider, blockID, newcidinfo)
		deleteBlockFromMem(oldpu.uid, oldpu.pid, blockID)

		//更新block的meta信息
		kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Sync, metainfo.SyncTypeBlock)
		if err != nil {
			log.Println("construct Syncblock KV error :", err)
			return
		}
		metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
		metaSyncTo(kmBlock, metaValue)
		kmBlock.SetKeyType(metainfo.Local) //将数据格式转换为local 保存在本地
		err = putKeyTo(kmBlock.ToString(), metaValue, "local")
		if err != nil {
			log.Println("construct SyncPidsK error :", err)
			return
		}
		return
	}

	log.Println("repair failed, cid is: ", blockID)

	return

}

//SearchNewProvider find a NEW provider for user
//TODO:how to improve the search algorithm
//TODO:is a Timer needed？
func SearchNewProvider(ctx context.Context, uid string, ugid []string) string {
	response := ""
	gp, ok := getGroupsInfo(uid)
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
			if sc.ConnectTo(ctx, localNode, response) {
				response = tmpPro
				break
			}
		}
	}

	return response
}
