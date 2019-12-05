package keeper

import (
	"log"
	"sort"
	"sync"

	"github.com/memoio/go-mefs/contracts"
)

// user-keeper-provider map
type groupsInfo struct {
	userID       string
	keepers      []string
	providers    []string
	localKeeper  string
	masterKeeper string
	buckets      sync.Map // key is bucket id; value is *bucketInfo
	upkeeping    *contracts.UpKeepingItem
}

type bucketInfo struct {
	bucketID       int32
	dataCount      int32
	parityCount    int32
	largestStripes int32
	stripes        sync.Map // key is cid, value is *cidInfo
}

//key:userid，value:*groupsInfo
var ukpInfo sync.Map

func getPUKeysFromukpInfo() []puKey {
	var res []puKey
	localKeeper := localNode.Identity.Pretty()
	ukpInfo.Range(func(k, v interface{}) bool {
		key := k.(string)
		if key == localKeeper {
			return true
		}
		value := v.(*groupsInfo)
		if value.upkeeping == nil {
			return true
		}
		for _, proID := range value.providers {
			tmpPU := puKey{
				uid: key,
				pid: proID,
			}
			res = append(res, tmpPU)
		}
		return true
	})
	return res
}

func getUnpaidUsers() []string {
	var res []string
	ukpInfo.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*groupsInfo)
		if value.upkeeping != nil {
			return true
		}
		res = append(res, key)
		return true
	})
	return res
}

//getGroupsInfo wrap get and set if not exist
func getGroupsInfo(groupid string) (*groupsInfo, bool) {
	thisGroupinfo, ok := ukpInfo.Load(groupid)
	if !ok {
		tempInfo := &groupsInfo{
			userID:       groupid,
			localKeeper:  groupid,
			masterKeeper: groupid,
		}
		err := saveUpkeepingToGP(groupid, tempInfo)
		if err != nil {
			return nil, false
		}

		localID := localNode.Identity.Pretty()
		for _, keeperID := range tempInfo.keepers {
			if localID == keeperID {
				tempInfo.localKeeper = localID
			}
		}
		if tempInfo.localKeeper == localID {
			ukpInfo.Store(groupid, tempInfo)
			return tempInfo, true
		}

		return nil, false
	}

	out, ok := thisGroupinfo.(*groupsInfo) //做类型断言的检查，接口的类型转换出错说明数据有问题，报错
	if !ok {
		log.Println("thisGroupinfo.(*groupsInfo) err！", thisGroupinfo)
		return nil, false
	}
	return out, true
}

func getBucketInfo(userID, bucketID string) (*bucketInfo, bool) {
	thisgroup, ok := getGroupsInfo(userID)
	if !ok {
		return nil, false
	}

	thisbucket, ok := thisgroup.buckets.Load(bucketID)
	if !ok {
		newBucket := &bucketInfo{}
		thisgroup.buckets.Store(bucketID, newBucket)
		return newBucket, true
	}
	return thisbucket.(*bucketInfo), true
}

func getLocalKeeperInGroup(groupid string) (string, error) {
	thisGroupInfo, ok := getGroupsInfo(groupid)
	if !ok {
		log.Println("getGroupsInfo err! groupid:", groupid)
		return "", errNoGroupsInfo
	}
	if thisGroupInfo.localKeeper == groupid {
		localID := localNode.Identity.Pretty()
		for _, keeperID := range thisGroupInfo.keepers {
			if keeperID == localID {
				thisGroupInfo.localKeeper = keeperID
			}
		}
	}
	if thisGroupInfo.localKeeper == groupid {
		// not my user
		ukpInfo.Delete(groupid)
		return "", errNotKeeperInThisGroup
	}
	return thisGroupInfo.localKeeper, nil
}

func getMasterKeeperInGroup(groupid string) (string, error) {
	thisGroupInfo, ok := getGroupsInfo(groupid)
	if !ok {
		log.Println("getGroupsInfo err! groupid:", groupid)
		return "", errNoGroupsInfo
	}

	if thisGroupInfo.masterKeeper == groupid {
		thisGroupInfo.masterKeeper = getMasterID(thisGroupInfo.keepers)
	}

	if thisGroupInfo.masterKeeper == groupid {
		return "", errNotKeeperInThisGroup
	}
	return thisGroupInfo.masterKeeper, nil
}

func localKeeperIsMaster(groupid string) bool {
	masterID, err := getMasterKeeperInGroup(groupid)
	if err != nil {
		log.Println("getMasterKeeperInGroup err.", err)
		return false
	}

	localID, err := getLocalKeeperInGroup(groupid)
	if err != nil {
		log.Println("getLocalKeeperInGroup err.", err)
		return false
	}

	return masterID == localID
}

// if this provider belongs to this keeper, then this keeper is master
// else call localKeeperIsMaster
func isMasterKeeper(groupid string, pid string) bool {
	thisGroupsInfo, ok := getGroupsInfo(groupid)
	if !ok {
		log.Println("localkeeperIsMaster err! There is no information in Pinfo,groupid:", groupid)
		return false
	}
	var mymaster []string
	mykids, ok := contracts.GetKeepersOfPro(pid)
	if ok {
		for _, keeperID := range thisGroupsInfo.keepers {
			for _, nkid := range mykids {
				if nkid == keeperID {
					mymaster = append(mymaster, keeperID)
					break
				}
			}
		}
		if len(mymaster) > 0 {
			return getMasterID(mymaster) == localNode.Identity.Pretty()
		}
	}

	return localKeeperIsMaster(groupid)
}

//getMasterID choose the middle nodes
func getMasterID(kidlist []string) string {
	sort.Strings(kidlist)
	return kidlist[len(kidlist)/2]
}
