package keeper

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

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
