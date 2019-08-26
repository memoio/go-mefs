package keeper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var repch chan string

const (
	RepairFailed = "Repair Failed"
)

func checkrepairlist(ctx context.Context) {
	fmt.Println("Checkrepairlist() start!")
	repch = make(chan string, 1024)
	go func() {
		for {
			select {
			case cid := <-repch:
				uid := cid[:IDLength]
				fmt.Println("get need repair cid :", cid)
				if localKeeperIsMaster(uid) {
					fmt.Println("repairing cid :", cid)
					RepairBlock(uid, cid)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

//Repair works in 3 steps 1.search a new provider,we do it in func SearchNewProvider
//2.put chunk to this provider 3.change metainfo and sync
func RepairBlock(userID string, blockID string) {
	var cpids, ugid []string
	var offset int
	blkinfo, err := metainfo.GetBlockMeta(blockID)
	if err != nil {
		fmt.Println("Get Block Meta error :", blockID, err)
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
					fmt.Println("GetBlockMeta error :", err)
					return false
				}
				if blkinfo.GetGid() == blockMeta.GetGid() && blkinfo.GetSid() == blockMeta.GetSid() && blkinfo.GetBid() != blockMeta.GetBid() {
					cpids = append(cpids, cid+metainfo.REPAIR_DELIMETER+pu.pid)
					ugid = append(ugid, pu.pid)
				} else if strings.Compare(cid, blockID) == 0 {
					offset = value.(*cidInfo).offset
					cpids = append(cpids, cid+metainfo.REPAIR_DELIMETER+pu.pid+metainfo.REPAIR_DELIMETER+strconv.Itoa(offset))
					ugid = append(ugid, pu.pid)
				}
				return true
			})
		}
		return true
	})
	response := SearchNewProvider(ugid)
	if response == "" {
		log.Println("Repair failed, no provider")
		return
	}
	fmt.Println("response provider :", response)
	var rpids string
	for _, cpid := range cpids {
		rpids += cpid
		rpids += metainfo.DELIMITER
	}
	km, err := metainfo.NewKeyMeta(blockID, metainfo.Repair)
	if err != nil {
		fmt.Println("construct repair KV error :", err)
		return
	}
	metaValue := rpids
	fmt.Println("cpids :", cpids, "\nrpids :", rpids)
	_, err = sendMetaRequest(km, metaValue, response)
	if err != nil {
		fmt.Println("err :", err)
	}
}

func handleRepairResponse(km *metainfo.KeyMeta, metaValue, provider string) {
	blockID := km.GetMid()
	splitedValue := strings.Split(metaValue, metainfo.DELIMITER)
	if len(splitedValue) != 4 {
		fmt.Println(metainfo.ErrIllegalValue, metaValue)
		return
	}
	pid := splitedValue[2]
	offset, err := strconv.Atoi(splitedValue[3])
	if err != nil {
		fmt.Println("strconv.Atoi offset error :", err)
		return
	}
	uid := blockID[:IDLength]
	pu := PU{
		pid: pid,
		uid: uid,
	}
	if strings.Compare(splitedValue[0], RepairFailed) == 0 {
		fmt.Println("修复失败 cid :", blockID)
		thischalinfo, ok := getChalinfo(pu)
		if ok {
			if thiscidinfo, ok := thischalinfo.Cid.Load(blockID); ok {
				thiscidinfo.(*cidInfo).res = false
				thiscidinfo.(*cidInfo).repair = 0
			}
		} else {
			fmt.Println("!ok blockID :", blockID, "\npid :", pid, "\nuid :", uid)
			newcidinfo := &cidInfo{
				repair: 0,
				offset: offset,
				res:    false,
			}
			var newCid, newTime sync.Map
			newCid.Store(blockID, newcidinfo)
			newchalinfo := &chalinfo{
				Time: newTime,
				Cid:  newCid,
			}
			LedgerInfo.Store(pu, newchalinfo)
		}
	} else {
		pu1 := PU{
			pid: provider,
			uid: uid,
		}
		fmt.Println("修复成功 cid :", blockID)
		newcidinfo := &cidInfo{
			repair:    0,
			availtime: utils.GetUnixNow(),
			offset:    offset,
		}

		if thischalinfo, ok := getChalinfo(pu1); ok {
			if thischalinfo.inChallenge == 1 {
				thischalinfo.tmpCid.Store(blockID, newcidinfo)
			} else if thischalinfo.inChallenge == 0 {
				thischalinfo.Cid.Store(blockID, newcidinfo)
			}
		} else {
			var newCid, newTime sync.Map
			newCid.Store(blockID, newcidinfo)
			newchalinfo := &chalinfo{
				Time: newTime,
				Cid:  newCid,
			}
			LedgerInfo.Store(pu1, newchalinfo)
		}

		oldchalinfo, isExist := getChalinfo(pu)
		if isExist {
			oldchalinfo.Cid.Delete(blockID)
		}

		addCredit(provider)

		var NewPids string
		var flag int
		thisGroupsInfo, ok := getGroupsInfo(uid)
		if !ok {
			fmt.Println(ErrNoGroupsInfo)
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
		if flag == len(thisGroupsInfo.Providers) {
			thisGroupsInfo.Providers = append(thisGroupsInfo.Providers, provider)
		}
		NewPids += provider

		kmPid, err := metainfo.NewKeyMeta(uid, metainfo.Sync, metainfo.SyncTypePid)
		if err != nil {
			fmt.Println("construct SyncPidsK error :", err)
			return
		}
		metaSyncTo(kmPid, NewPids)
		kmPid.SetKeyType(metainfo.Local) //将数据格式转换为local 保存在本地
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmPid.ToString(), NewPids, "local")
		if err != nil {
			fmt.Println("construct SyncPidsK error :", err)
			return
		}
		//更新block的meta信息
		kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.Sync, metainfo.SyncTypeBlock)
		if err != nil {
			fmt.Println("construct Syncblock KV error :", err)
			return
		}
		metaValue := provider + metainfo.DELIMITER + strconv.Itoa(offset)
		metaSyncTo(kmBlock, metaValue)
		kmBlock.SetKeyType(metainfo.Local) //将数据格式转换为local 保存在本地
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmBlock.ToString(), metaValue, "local")
		if err != nil {
			fmt.Println("construct SyncPidsK error :", err)
			return
		}
	}
	return

}

//SearchNewProvider find a NEW provider for user
//TODO:how to improve the search algorithm
//TODO:is a Timer needed？
func SearchNewProvider(ugid []string) string {
	var response string                                  //return the provider id we need
	r := rand.New(rand.NewSource(time.Now().UnixNano())) //r is a time seed.we use it to create the random number
	for {
		if len(localPeerInfo.Providers) == 0 {
			return ""
		}
		response = localPeerInfo.Providers[r.Intn(len(localPeerInfo.Providers))] //first we find a random provider
		var j, flag int
		flag = 0
		for j = 0; j < len(ugid); j++ { //this provider may belong to this user already
			if response != ugid[j] {
				flag++
			}
		}
		if flag == len(ugid) {
			break
		}
	}
	return response
}
