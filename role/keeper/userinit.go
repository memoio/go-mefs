package keeper

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

// handleUserInit collect keepers and providers for user
// return kv, key: queryID/"UserInit"/userID/keepercount/providercount;
// value: kid1kid2../pid1pid2..
func (k *Info) handleUserInit(km *metainfo.KeyMeta, from string) {
	log.Println("NewUserInit: ", km.ToString(), " From: ", from)
	options := km.GetOptions()
	if len(options) != 3 {
		return
	}

	kc, err := strconv.Atoi(options[1])
	if err != nil {
		log.Println("handleUserInitReq: ", err)
		return
	}

	pc, err := strconv.Atoi(options[2])
	if err != nil {
		log.Println("handleUserInitReq: ", err)
		return
	}

	uid := options[0]
	qid := km.GetMid()
	price := int64(utils.STOREPRICEPEDOLLAR)
	var response string
	if qid != uid {
		log.Println("Get k/p numbers from query contract of user: ", uid)
		queryAddr, _ := ad.GetAddressFromID(qid)
		item, err := contracts.GetQueryInfo(queryAddr, queryAddr)
		if item.Completed || err != nil {
			log.Println("complete:", item.Completed, "error:", err)
			return
		}
		kc = int(item.KeeperNums)
		pc = int(item.ProviderNums)
		price = item.Price
	}

	if pos.GetPosId() == uid {
		price = int64(utils.STOREPRICEPEDOLLAR)
	}

	response, err = k.initUser(qid, uid, kc, pc, price)
	if err != nil {
		if err != nil {
			log.Println("handleUserInitReq err: ", err)
			return
		}
	}
	log.Println("New user: ", qid, " keeperCount: ", kc, "providerCount: ", pc, "price: ", price)

	k.ds.SendMetaRequest(context.Background(), int32(metainfo.Put), km.ToString(), []byte(response), nil, from)
}

//response: kid1kid2../pid1pid2..
func (k *Info) initUser(qid, uid string, kc, pc int, price int64) (string, error) {
	var newResponse strings.Builder

	thisInfo, ok := k.ukpManager.gMap.Load(qid)
	if !ok {
		localID := k.netID
		// fill self
		newResponse.WriteString(localID)
		kc--
		//fill other keepers
		k.keepers.Range(func(k, v interface{}) bool {
			if kc == 0 {
				return false
			}

			key := k.(string)
			if key == localID {
				return true
			}

			thisinfo := v.(*kInfo)
			if thisinfo.online == true {
				newResponse.WriteString(key)
				kc--
			}
			return true
		})

		newResponse.WriteString(metainfo.DELIMITER)

		// fill providers
		k.providers.Range(func(k, v interface{}) bool {
			if pc == 0 {
				return false
			}
			key := k.(string)
			thisinfo := v.(*pInfo)
			if thisinfo.online == true {
				newResponse.WriteString(key)
				pc--
			}
			return true
		})

		return newResponse.String(), nil
	}

	thisGroupsInfo := thisInfo.(*groupInfo)
	// user has init
	for _, pid := range thisGroupsInfo.keepers {
		newResponse.WriteString(pid)
	}

	newResponse.WriteString(metainfo.DELIMITER)

	for _, pid := range thisGroupsInfo.providers {
		newResponse.WriteString(pid)
	}
	return newResponse.String(), nil
}

// handleUserNotify return kv,
// key: queryID/"UserNotify"/userID/keepercount/providercount;
// value: kid1kid2../pid1pid2..
func (k *Info) handleUserNotify(km *metainfo.KeyMeta, metaValue []byte, from string) ([]byte, error) {
	log.Println("NewUserNotify: ", km.ToString(), metaValue, "From:", from)

	options := km.GetOptions()
	if len(options) != 3 {
		return nil, errors.New("Wrong key")
	}

	kc, err := strconv.Atoi(options[1])
	if err != nil {
		log.Println("handleUserInitReq: ", err)
		return nil, err
	}

	pc, err := strconv.Atoi(options[2])
	if err != nil {
		log.Println("handleUserInitReq: ", err)
		return nil, err
	}

	uid := options[0]
	qid := km.GetMid()

	go k.fillPinfo(qid, uid, pc, kc, metaValue, from)

	return []byte("ok"), nil
}

// fillPinfo fill user's uInfo, groupInfo in ukpMap
func (k *Info) fillPinfo(groupID, userID string, kc, pc int, metaValue []byte, from string) {
	//将value切分，生成好对应的keepers和providers列表
	splited := strings.Split(string(metaValue), metainfo.DELIMITER)
	if len(splited) < 2 {
		log.Println("UserNotif value is not correct: ", metaValue)
		return
	}

	keepers := make([]string, kc)
	providers := make([]string, pc)

	kids := splited[0]
	for i := 0; i < len(kids)/utils.IDLength; i++ {
		keeper := string(kids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(keeper)
		if err != nil {
			continue
		}
		keepers = append(keepers, keeper)
	}

	pids := splited[1]
	for i := 0; i < len(pids)/utils.IDLength; i++ {
		providerID := string(pids[i*utils.IDLength : (i+1)*utils.IDLength])
		_, err := peer.IDB58Decode(providerID)
		if err != nil {
			continue
		}
		providers = append(providers, providerID)
	}

	tempInfo, err := k.fillGroup(groupID, userID, keepers, providers)
	if err != nil {
		return
	}

	kmKid, err := metainfo.NewKeyMeta(groupID, metainfo.Keepers)
	if err != nil {
		log.Println("handleNewUserNotif err: ", err)
		return
	}

	kmPid, err := metainfo.NewKeyMeta(groupID, metainfo.Providers)
	if err != nil {
		log.Println("handleNewUserNotif err: ", err)
		return
	}

	var pidstrings strings.Builder
	for _, keeperID := range tempInfo.keepers {
		pidstrings.WriteString(keeperID)
	}
	ctx := context.Background()
	k.ds.PutKey(ctx, kmKid.ToString(), []byte(pidstrings.String()), "local")

	pidstrings.Reset()
	for _, proID := range tempInfo.providers {
		pidstrings.WriteString(proID)
		// replace ledgerinfo
		thisPU := pqKey{
			qid: groupID,
			pid: proID,
		}
		newChal := &chalinfo{}
		k.lManager.lMap.Store(thisPU, newChal)
	}

	k.ds.PutKey(ctx, kmPid.ToString(), []byte(pidstrings.String()), "local")

	return
}

func (k *Info) handleContracts(km *metainfo.KeyMeta, from string) {
	log.Println("New User", km.ToString(), "From:", from)
	qid := km.GetMid()
	ops := km.GetOptions()
	if len(ops) != 1 {
		return
	}

	err := k.ukpManager.saveUpkeeping(qid, false)
	if err != nil {
		log.Println("Save ", qid, "'s Upkeeping err", err)
	}
	log.Println("Save ", qid, "'s Upkeeping success")

	err = k.ukpManager.saveQuery(qid, false)
	if err != nil {
		log.Println("Save ", qid, "'s Query err", err)
	}
	log.Println("Save ", qid, "'s Query success")

	uid := ops[0]
	k.setQuery(uid, qid)
}
