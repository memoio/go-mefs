package keeper

import (
	"errors"
	"log"
	"strings"

	"github.com/memoio/go-mefs/utils/metainfo"
)

//response: kid1kid2../pid1pid2..
func initUser(userID string, keeperCount, providerCount int, price int64) (string, error) {
	thisInfo, ok := ukpInfo.Load(userID)
	if !ok {
		return userNewInit(userID, keeperCount, providerCount, price)
	}

	var responseExisted strings.Builder
	thisGroupsInfo := thisInfo.(*groupsInfo)
	// user has init
	for _, pid := range thisGroupsInfo.keepers {
		responseExisted.WriteString(pid)
	}

	responseExisted.WriteString(metainfo.DELIMITER)

	for _, pid := range thisGroupsInfo.providers {
		responseExisted.WriteString(pid)
	}
	return responseExisted.String(), nil
}

func userNewInit(userID string, keeperCount, providerCount int, price int64) (string, error) {
	localID := localNode.Identity.Pretty()

	kmKid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		return "", err
	}

	kmPid, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		return "", err
	}

	var newResponse, pids strings.Builder
	// fill self
	newResponse.WriteString(localID)
	pids.WriteString(localID)
	keeperCount--
	//fill other keepers
	localPeerInfo.keepersInfo.Range(func(k, v interface{}) bool {
		if keeperCount == 0 {
			return false
		}

		key := k.(string)
		if key == localID {
			return true
		}

		thisinfo := v.(*kInfo)
		if thisinfo.online == true {
			newResponse.WriteString(key)
			pids.WriteString(key)
			keeperCount--
		}
		return true
	})

	putKeyTo(kmKid.ToString(), pids.String(), "local")

	newResponse.WriteString(metainfo.DELIMITER)
	pids.Reset()
	// fill providers
	localPeerInfo.providersInfo.Range(func(k, v interface{}) bool {
		if providerCount == 0 {
			return false
		}
		key := k.(string)
		thisinfo := v.(*pInfo)
		if thisinfo.online == true {
			newResponse.WriteString(key)
			pids.WriteString(key)
			providerCount--
		}
		return true
	})

	putKeyTo(kmPid.ToString(), pids.String(), "local")

	return newResponse.String(), nil
}

func fillUserInfo(groupid string, keepers, providers []string) (*groupsInfo, error) {
	tempInfo := &groupsInfo{
		keepers:     keepers,
		providers:   providers,
		userID:      groupid,
		localKeeper: groupid,
	}

	saveUpkeepingToGP(groupid, tempInfo)

	localID := localNode.Identity.Pretty()
	for _, keeperID := range tempInfo.keepers {
		if localID == keeperID {
			tempInfo.localKeeper = localID
		}
	}

	// not my user
	if tempInfo.localKeeper == groupid {
		log.Println(groupid, "is not my user")
		return nil, errors.New("Not my user")
	}

	ukpInfo.Store(groupid, tempInfo)

	err := saveQuery(groupid, false)
	if err != nil {
		log.Println("Save ", groupid, "'s Query error: ", err)
	}

	_, err = getUserBLS12Config(groupid)
	if err != nil {
		log.Println("Save ", groupid, "'s BLS pubkey error: ", err)
	}

	return tempInfo, nil
}

func initUserInfo(groupid string, keepers, providers []string) (*groupsInfo, error) {
	tempInfo := &groupsInfo{
		keepers:     keepers,
		providers:   providers,
		userID:      groupid,
		localKeeper: groupid,
	}

	localID := localNode.Identity.Pretty()
	for _, keeperID := range tempInfo.keepers {
		if localID == keeperID {
			tempInfo.localKeeper = localID
		}
	}

	// not my user
	if tempInfo.localKeeper == groupid {
		log.Println(groupid, "is not my user")
		return nil, errors.New("Not my user")
	}

	ukpInfo.Store(groupid, tempInfo)
	return tempInfo, nil
}

// fillPinfo fill user's uInfo, groupsInfo in ukpMap
// not get upkeeping contract
func fillPinfo(groupid string, keepers, providers []string, from string) {
	tempInfo, err := initUserInfo(groupid, keepers, providers)
	if err != nil {
		return
	}

	kmKid, err := metainfo.NewKeyMeta(groupid, metainfo.Local, metainfo.SyncTypeKid)
	if err != nil {
		log.Println("handleNewUserNotif err: ", err)
		return
	}

	kmPid, err := metainfo.NewKeyMeta(groupid, metainfo.Local, metainfo.SyncTypePid)
	if err != nil {
		log.Println("handleNewUserNotif err: ", err)
		return
	}

	var pids strings.Builder
	for _, keeperID := range tempInfo.keepers {
		pids.WriteString(keeperID)
	}

	putKeyTo(kmKid.ToString(), pids.String(), "local")

	pids.Reset()
	for _, proID := range tempInfo.providers {
		pids.WriteString(proID)
		// replace ledgerinfo
		thisPU := puKey{
			uid: groupid,
			pid: proID,
		}
		newChal := &chalinfo{}
		ledgerInfo.Store(thisPU, newChal)
	}

	putKeyTo(kmPid.ToString(), pids.String(), "local")

	if !localPeerInfo.enableBft {
		kmRes, err := metainfo.NewKeyMeta(groupid, metainfo.Local, metainfo.SyncTypeBft)
		if err != nil {
			log.Println(err)
			return
		}
		resValue := "simple"
		putKeyTo(kmRes.ToString(), resValue, "local")
		kmRes.SetKeyType(metainfo.UserInitNotifRes)
		_, err = sendMetaRequest(kmRes, resValue, from)
		if err != nil {
			log.Println(err)
		}
		log.Println("use simple mode，userID:", groupid)
	}

	return
}
