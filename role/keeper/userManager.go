package keeper

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils/metainfo"
)

type ukp struct {
	localID string
	ds      data.Service
	gMap    sync.Map //key:queryid，value:*groupsInfo
}

// user-keeper-provider map
type groupsInfo struct {
	groupID      string // is queryID
	owner        string // is userID
	localKeeper  string
	masterKeeper string
	keepers      []string
	providers    []string
	upkeeping    *contracts.UpKeepingItem
	query        *contracts.QueryItem
	blsPubKey    *mcl.PublicKey
	buckets      sync.Map // key is bucketID; value is *bucketInfo
}

type bucketInfo struct {
	bucketID    int
	dataCount   int
	parityCount int
	chunkNum    int      // = dataCount+parityCount; which is largest chunkID
	curStripes  int      // largest stripeID
	stripes     sync.Map // key is stripeID_chunkID, value is *cidInfo
}

func (u *ukp) getPUKeys() []pqKey {
	var res []pqKey
	localKeeper := u.localID
	u.gMap.Range(func(k, v interface{}) bool {
		key := k.(string)
		if key == localKeeper {
			return true
		}
		value := v.(*groupsInfo)
		if value.upkeeping == nil {
			return true
		}
		for _, proID := range value.providers {
			tmpPU := pqKey{
				qid: key,
				pid: proID,
			}
			res = append(res, tmpPU)
		}
		return true
	})
	return res
}

func (u *ukp) getUnpaidUsers() []string {
	var res []string
	u.gMap.Range(func(k, v interface{}) bool {
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
func (u *ukp) getGroupsInfo(groupid string) (*groupsInfo, bool) {
	thisGroupinfo, ok := u.gMap.Load(groupid)
	if !ok {
		tempInfo := &groupsInfo{
			groupID:      groupid,
			owner:        groupid,
			localKeeper:  u.localID,
			masterKeeper: groupid,
		}
		err := saveUpkeepingToGP(groupid, tempInfo)
		if err != nil {
			return nil, false
		}

		u.gMap.Store(groupid, tempInfo)
		return tempInfo, true
	}

	return thisGroupinfo.(*groupsInfo), true
}

func (u *ukp) getBucketInfo(groupID, bucketID string) (*bucketInfo, bool) {
	thisGroupinfo, ok := u.gMap.Load(groupID)
	if !ok {
		return nil, false
	}

	thisgroup := thisGroupinfo.(*groupsInfo)

	thisbucket, ok := thisgroup.buckets.Load(bucketID)
	if !ok {
		newBucket := &bucketInfo{}
		thisgroup.buckets.Store(bucketID, newBucket)
		return newBucket, true
	}
	return thisbucket.(*bucketInfo), true
}

func (u *ukp) addBlockMeta(gid, bid string, ci *cidInfo) error {
	bucketID, stripeID, chunkID, err := metainfo.GetIDsFromBlock(bid)
	if err != nil {
		return err
	}

	thisBucketinfo, ok := u.getBucketInfo(gid, bucketID)
	if !ok {
		return errors.New("cannot create bucket info")
	}

	thisBucketinfo.stripes.Store(bid, ci)

	snum, err := strconv.Atoi(stripeID)
	if err != nil {
		return err
	}

	if snum > thisBucketinfo.curStripes {
		thisBucketinfo.curStripes = snum
	}

	cnum, err := strconv.Atoi(chunkID)
	if err != nil {
		return err
	}

	if cnum > thisBucketinfo.chunkNum {
		thisBucketinfo.chunkNum = cnum
	}

	return nil
}

func (u *ukp) deleteBlockMeta(gid, bid string) {
	gri, ok := u.gMap.Load(gid)
	if !ok {
		return
	}

	bis := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := gri.(*groupsInfo).buckets.Load(bis[0])
	if !ok {
		return
	}

	bui.(*bucketInfo).stripes.Delete(bis[1])
}

func (u *ukp) getLocalKeeperInGroup(groupid string) (string, error) {
	thisGroupInfo, ok := u.getGroupsInfo(groupid)
	if !ok {
		log.Println("getGroupsInfo err! groupid:", groupid)
		return "", errNoGroupsInfo
	}
	if thisGroupInfo.localKeeper == groupid {
		for _, keeperID := range thisGroupInfo.keepers {
			if keeperID == u.localID {
				thisGroupInfo.localKeeper = keeperID
			}
		}
	}
	if thisGroupInfo.localKeeper == groupid {
		// not my user
		u.gMap.Delete(groupid)
		return "", errNotKeeperInThisGroup
	}
	return thisGroupInfo.localKeeper, nil
}

func (u *ukp) getMasterKeeperInGroup(groupid string) (string, error) {
	thisGroupInfo, ok := u.getGroupsInfo(groupid)
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

func (u *ukp) localKeeperIsMaster(groupid string) bool {
	masterID, err := u.getMasterKeeperInGroup(groupid)
	if err != nil {
		log.Println("getMasterKeeperInGroup err.", err)
		return false
	}

	localID, err := u.getLocalKeeperInGroup(groupid)
	if err != nil {
		log.Println("getLocalKeeperInGroup err.", err)
		return false
	}

	return masterID == localID
}

// if this provider belongs to this keeper, then this keeper is master
// else call localKeeperIsMaster
func (u *ukp) isMasterKeeper(groupid string, pid string) bool {
	thisGroupsInfo, ok := u.getGroupsInfo(groupid)
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
			return getMasterID(mymaster) == u.localID
		}
	}

	return u.localKeeperIsMaster(groupid)
}

//getMasterID choose the middle nodes
func getMasterID(kidlist []string) string {
	sort.Strings(kidlist)
	return kidlist[len(kidlist)/2]
}
