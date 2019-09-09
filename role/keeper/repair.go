package keeper

import (
	"context"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/memoio/go-mefs/contracts"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	sc "github.com/memoio/go-mefs/utils/swarmconnect"
)

var repch chan string

const (
	// RepairFailed ...
	RepairFailed = "Repair Failed"
)

func checkrepairlist(ctx context.Context) {
	log.Println("Check Repairlist start!")
	repch = make(chan string, 1024)
	go func() {
		for {
			select {
			case cid := <-repch:
				uid := cid[:utils.IDLength]
				if localKeeperIsMaster(uid) {
					log.Println("repairing cid: ", cid)
					RepairBlock(ctx, uid, cid)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

//RepairBlock works in 3 steps 1.search a new provider,we do it in func SearchNewProvider
//2.put chunk to this provider 3.change metainfo and sync
func RepairBlock(ctx context.Context, userID string, blockID string) {
	var cpids, ugid []string
	var offset int
	var response string
	blkinfo, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		log.Println("Get Block Meta error :", blockID, err)
		return
	}

	LedgerInfo.Range(func(k, v interface{}) bool {
		pu := k.(PU)
		thischalinfo := v.(*chalinfo)
		if pu.uid == userID && thischalinfo != nil {
			thischalinfo.Cid.Range(func(key, value interface{}) bool {
				cid := key.(string)
				blockMeta, err := metainfo.GetBlockMeta(cid)
				if err != nil {
					log.Println("GetBlockMeta error :", err)
					return false
				}

				if strings.Compare(cid, blockID) == 0 {
					offset = value.(*cidInfo).offset
					cpids = append(cpids, cid+metainfo.REPAIR_DELIMETER+pu.pid+metainfo.REPAIR_DELIMETER+strconv.Itoa(offset))
					ugid = append(ugid, pu.pid)
					response = pu.pid
				}

				if blkinfo.GetGid() == blockMeta.GetGid() && blkinfo.GetSid() == blockMeta.GetSid() && blkinfo.GetBid() != blockMeta.GetBid() {
					cpids = append(cpids, cid+metainfo.REPAIR_DELIMETER+pu.pid)
					ugid = append(ugid, pu.pid)
				}

				return true
			})
		}
		return true
	})

	if len(response) > 0 {
		if !sc.ConnectTo(ctx, localNode, response) {
			response = SearchNewProvider(ctx, userID, ugid)
			if response == "" {
				log.Println("Repair failed, no extra provider")
				return
			}
		}
	}

	var rpids string
	for _, cpid := range cpids {
		rpids += cpid
		rpids += metainfo.DELIMITER
	}
	km, err := metainfo.NewKeyMeta(blockID, metainfo.Repair)
	if err != nil {
		log.Println("construct repair KV error: ", err)
		return
	}
	metaValue := rpids
	log.Println("cpids: ", cpids, "\nrpids: ", rpids, " \nrepairs on: ", response)
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
	oldpu := PU{
		pid: oldPid,
		uid: uid,
	}
	if strings.Compare(splitedValue[0], RepairFailed) == 0 {
		log.Println("repair failed, cid is: ", blockID)
	} else {
		log.Println("repair success, cid is: ", blockID)
		newpu := PU{
			pid: provider,
			uid: uid,
		}

		newcidinfo := &cidInfo{
			repair:    0,
			availtime: utils.GetUnixNow(),
			offset:    offset,
		}

		if thischalinfo, ok := getChalinfo(newpu); ok {
			if thischalinfo.inChallenge == 1 {
				thischalinfo.tmpCid.Store(blockID, newcidinfo)
			} else if thischalinfo.inChallenge == 0 {
				thischalinfo.Cid.Store(blockID, newcidinfo)
			}
		} else {
			isTestUser := false
			addr, err := address.GetAddressFromID(newpu.uid)
			if err == nil {
				_, _, err = contracts.GetUKFromResolver(addr)
				if err != nil {
					isTestUser = true
				}
			}

			var newCid, newTime sync.Map
			newCid.Store(blockID, newcidinfo)
			newchalinfo := &chalinfo{
				Time:     newTime,
				Cid:      newCid,
				testuser: isTestUser,
			}
			LedgerInfo.Store(newpu, newchalinfo)
		}

		oldchalinfo, isExist := getChalinfo(oldpu)
		if isExist {
			oldchalinfo.Cid.Delete(blockID)
		}

		addCredit(provider)

		var NewPids string
		var flag int
		thisGroupsInfo, ok := getGroupsInfo(uid)
		if !ok {
			log.Println(ErrNoGroupsInfo)
			return
		}
		for _, Pid := range thisGroupsInfo.Providers {
			if strings.Compare(Pid, provider) == 0 {
				break
			} else {
				flag++
				NewPids += Pid
			}
		}

		// new provider, now there is no new provider
		if flag == len(thisGroupsInfo.Providers) {
			thisGroupsInfo.Providers = append(thisGroupsInfo.Providers, provider)
			NewPids += provider

			kmPid, err := metainfo.NewKeyMeta(uid, metainfo.Sync, metainfo.SyncTypePid)
			if err != nil {
				log.Println("construct SyncPidsK error :", err)
				return
			}
			metaSyncTo(kmPid, NewPids)
			kmPid.SetKeyType(metainfo.Local) //将数据格式转换为local 保存在本地
			err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmPid.ToString(), NewPids, "local")
			if err != nil {
				log.Println("construct SyncPidsK error :", err)
				return
			}
			//更新block的meta信息
			kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Sync, metainfo.SyncTypeBlock)
			if err != nil {
				log.Println("construct Syncblock KV error :", err)
				return
			}
			metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
			metaSyncTo(kmBlock, metaValue)
			kmBlock.SetKeyType(metainfo.Local) //将数据格式转换为local 保存在本地
			err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmBlock.ToString(), metaValue, "local")
			if err != nil {
				log.Println("construct SyncPidsK error :", err)
				return
			}
		}

	}
	return

}

//SearchNewProvider find a NEW provider for user
//TODO:how to improve the search algorithm
//TODO:is a Timer needed？
func SearchNewProvider(ctx context.Context, uid string, ugid []string) string {
	var response string
	gp, ok := getGroupsInfo(uid)
	if !ok {
		log.Println("SearchNewProvider getGroupsInfo() error")
		return response
	}

	lenp := len(gp.Providers)

	if lenp == 0 || lenp <= len(ugid) {
		return response
	}

	retry := 0
	//return the provider id we need
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //r is a time seed.we use it to create the random number
	for {
		retry++
		response = gp.Providers[r.Intn(lenp)] //first we find a random provider
		var j, flag int
		flag = 0
		for j = 0; j < len(ugid); j++ { //this provider may belong to this stripe already
			if response != ugid[j] {
				flag++
			}
		}
		if flag == len(ugid) {
			if sc.ConnectTo(ctx, localNode, response) {
				break
			}
		}

		if retry > 5*len(ugid) {
			return ""
		}
	}
	return response
}
