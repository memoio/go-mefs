package keeper

import (
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
	"golang.org/x/crypto/sha3"
)

// user-keeper-provider map
type groupInfo struct {
	status       bool
	sessionID    uuid.UUID // for user
	sessionTime  int64
	clusterID    uint64 // raft clusterID
	nodeID       uint64
	groupID      string // is queryID
	userID       string // is userID
	localKeeper  string
	masterKeeper string
	bft          bool
	keepers      []string
	providers    []string
	rootID       string
	upkeeping    *role.UpKeepingItem
	query        *role.QueryItem
	bucketNum    int64    // largest bucketID
	buckets      sync.Map // key:bucketID(string); value: *bucketInfo
	ledgerMap    sync.Map // key:proID，value:*chalInfo
}

func newGroup(localID, uid, qid string, keepers, providers []string) (*groupInfo, error) {
	tempInfo := &groupInfo{
		groupID:      qid,
		userID:       uid,
		localKeeper:  localID,
		masterKeeper: qid,
		keepers:      keepers,
		providers:    providers,
		sessionID:    uuid.Nil,
		bucketNum:    -1,
	}

	if qid != uid {
		err := tempInfo.loadContracts(true)
		if err != nil {
			return nil, err
		}
	} else {
		flag := false
		for _, keeperID := range tempInfo.keepers {
			if localID == keeperID {
				flag = true
			}
		}

		// not my user
		if !flag {
			utils.MLogger.Warn(uid, " is not my user, keepers are: ", keepers)
			return nil, role.ErrNotMyUser
		}
	}

	nodeID, err := address.GetNodeIDFromID(localID)
	if err != nil {
		return nil, err
	}

	tempInfo.nodeID = nodeID

	clusterID, err := address.GetNodeIDFromID(qid)
	if err != nil {
		return nil, err
	}

	tempInfo.clusterID = clusterID

	tempInfo.masterKeeper = getMasterID(tempInfo.keepers, tempInfo.groupID)

	utils.MLogger.Debugf("%s has masterID %s, and localID %s", qid, tempInfo.masterKeeper, tempInfo.localKeeper)
	return tempInfo, nil
}

// if this provider belongs to this keeper, then this keeper is master
func (g *groupInfo) isMaster(pid string) bool {
	var mymaster []string
	mykids, ok := role.GetKeepersOfPro(pid)
	if ok {
		for _, keeperID := range g.keepers {
			for _, nkid := range mykids {
				if nkid == keeperID {
					mymaster = append(mymaster, keeperID)
					break
				}
			}
		}
		if len(mymaster) > 0 {
			utils.MLogger.Debugf(pid, " has master keepers: ", mymaster)
			return getMasterID(mymaster, g.groupID) == g.localKeeper
		}
	}

	return g.localKeeper == g.masterKeeper
}

//getMasterID choose the middle nodes
func getMasterID(kidlist []string, groupID string) string {
	if len(kidlist) == 0 {
		return groupID
	}

	if len(kidlist) == 1 {
		return kidlist[0]
	}

	largeKid := ""
	largeIndex := 0
	for i, kid := range kidlist {
		d := sha3.Sum256([]byte(groupID + kid))
		if strings.Compare(hex.EncodeToString(d[:]), largeKid) > 0 {
			largeKid = hex.EncodeToString(d[:])
			largeIndex = i
		}
	}

	return kidlist[largeIndex]
}

func (g *groupInfo) getLInfo(proID string, mode bool) *lInfo {
	if g == nil {
		return nil
	}

	thisIl, ok := g.ledgerMap.Load(proID)
	if !ok {
		if mode {
			templInfo := &lInfo{
				inChallenge:  false,
				lastChalTime: time.Now().Unix(),
			}
			g.ledgerMap.Store(proID, templInfo)
			return templInfo
		}
		return nil
	}

	return thisIl.(*lInfo)
}

func (g *groupInfo) getBucketInfo(bucketID string, mode bool) *bucketInfo {
	if g == nil {
		return nil
	}

	thisIb, ok := g.buckets.Load(bucketID)
	if !ok {
		if mode {
			bop := df.DefaultSuperBucketOptions()
			bid, err := strconv.ParseInt(bucketID, 10, 0)
			if err != nil {
				return nil
			}

			if g.userID == pos.GetPostId() {
				bop.DataCount = 1
				bop.ParityCount = pos.Reps - 1
				bop.SegmentCount = pos.SegCount
				bop.SegmentSize = pos.SegSize
			} else {
				if bid > 0 {
					bop = df.DefaultBucketOptions()
				}
			}

			tempInfo := &bucketInfo{
				bops:       bop,
				curStripes: -1,
				chunkNum:   int(bop.GetDataCount() + bop.GetParityCount()),
			}
			g.buckets.Store(bucketID, tempInfo)
			return tempInfo
		}
		return nil
	}

	return thisIb.(*bucketInfo)
}

func (g *groupInfo) addBucket(bucketID string, binfo *mpb.BucketInfo) error {
	bucketNum, err := strconv.ParseInt(bucketID, 10, 0)
	if err != nil {
		return err
	}

	thisBucket := g.getBucketInfo(bucketID, true)
	if thisBucket == nil {
		return nil
	}

	if g.bucketNum < bucketNum {
		g.bucketNum = bucketNum
	}

	thisBucket.bops = binfo.GetBOpts()
	thisBucket.chunkNum = int(binfo.GetBOpts().GetDataCount() + binfo.GetBOpts().GetParityCount())

	return nil
}

// bid is bucketID_stripeID_chunkID
func (g *groupInfo) addBlockMeta(bid, pid string, offset int) error {
	// store in buckets
	bucketID, stripeID, chunkID, err := metainfo.GetIDsFromBlock(bid)
	if err != nil {
		return err
	}

	thisBucket := g.getBucketInfo(bucketID, true)
	if thisBucket == nil {
		return nil
	}

	snum, err := strconv.Atoi(stripeID)
	if err != nil {
		return err
	}

	if snum < thisBucket.curStripes && offset < int(thisBucket.bops.GetSegmentCount()) {
		utils.MLogger.Infof("group %s add chunk: %s and its offset %d, is short", g.groupID, bid, offset)
	}

	for _, proID := range g.providers {
		if proID != pid {
			thisLinfo := g.getLInfo(proID, false)
			if thisLinfo == nil {
				continue
			}

			binfo, ok := thisLinfo.blockMap.Load(bid)
			if ok {
				thisLinfo.maxlength -= int64((binfo.(*blockInfo).offset) * int(thisBucket.bops.GetSegmentSize()))
				thisLinfo.blockMap.Delete(bid)
			}
		}

	}

	thisLinfo := g.getLInfo(pid, true)
	if thisLinfo == nil {
		return nil
	}

	newcidinfo := &blockInfo{
		availtime: time.Now().Unix(),
		offset:    offset,
		repair:    0,
		storedOn:  pid,
	}

	// store in cidMap
	oldOffset := 0
	v, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		newcidinfo = v.(*blockInfo)
		oldPid := newcidinfo.storedOn
		// verify again
		// delete from old provider
		if oldPid != pid {
			oldLinfo := g.getLInfo(oldPid, true)
			if oldLinfo != nil {
				binfo, ok := thisLinfo.blockMap.Load(bid)
				if ok {
					oldLinfo.maxlength -= int64((binfo.(*blockInfo).offset) * int(thisBucket.bops.GetSegmentSize()))
					oldLinfo.blockMap.Delete(bid)
				}
			}
			newcidinfo.storedOn = pid
			newcidinfo.repair = 0
		}

		// update offset
		oldOffset = newcidinfo.offset
		if offset > oldOffset {
			newcidinfo.offset = offset
			thisLinfo.maxlength += int64((offset - oldOffset) * int(thisBucket.bops.GetSegmentSize()))
		}
	} else {
		thisLinfo.blockMap.Store(bid, newcidinfo)
		// change length;store or calculate at startup
		thisLinfo.maxlength += int64((offset) * int(thisBucket.bops.GetSegmentSize()))
	}

	bids := strings.SplitN(bid, metainfo.BlockDelimiter, 2)
	if len(bids) < 2 {
		return role.ErrWrongKey
	}
	// key: stripeID_chunkID
	thisBucket.stripes.Store(bids[1], newcidinfo)

	if snum > thisBucket.curStripes {
		thisBucket.curStripes = snum
	}

	cnum, err := strconv.Atoi(chunkID)
	if err != nil {
		return err
	}

	if cnum > thisBucket.chunkNum {
		utils.MLogger.Warn("chunkID %d is larger than bucket ops %d", cnum, thisBucket.chunkNum)
		//thisBucket.chunkNum = cnum
	}

	return nil
}

func (g *groupInfo) deleteBlockMeta(bid, pid string) {
	segSize := df.DefaultSegmentSize
	// delete from buckets
	bids := strings.Split(bid, metainfo.BlockDelimiter)
	if len(bids) != 3 {
		return
	}

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		binfo := bui.(*bucketInfo)
		binfo.stripes.Delete(bids[1] + metainfo.BlockDelimiter + bids[2])
		segSize = int(bui.(*bucketInfo).bops.GetSegmentSize())
		if g.userID == pos.GetPostId() {
			strpeNum, err := strconv.Atoi(bids[1])
			if err == nil {
				binfo.curStripes = strpeNum - 1
			}
		}
	}

	thisLinfo := g.getLInfo(pid, false)
	if thisLinfo == nil {
		return
	}

	// delete from blockMap
	thisICid, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		thisCid := thisICid.(*blockInfo)
		thisLinfo.maxlength -= int64(thisCid.offset * segSize)
		thisLinfo.blockMap.Delete(bid)
		thisLinfo.faultCid.Delete(bid)
	}

	return
}

// bucketID_stripeID_chunkID
func (g *groupInfo) getBlockPost(bid string) (string, error) {
	bids := strings.SplitN(bid, metainfo.BlockDelimiter, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		sti, ok := bui.(*bucketInfo).stripes.Load(bids[1])
		if ok {
			return sti.(*blockInfo).storedOn, nil
		}
	}

	return "", role.ErrNoBlock
}

func (g *groupInfo) getBlockAvail(bid string) (int64, error) {
	bids := strings.SplitN(bid, metainfo.BlockDelimiter, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		sti, ok := bui.(*bucketInfo).stripes.Load(bids[1])
		if ok {
			return sti.(*blockInfo).availtime, nil
		}
	}

	return 0, role.ErrNoBlock
}

type bucketInfo struct {
	bops       *mpb.BucketOptions
	chunkNum   int      // = dataCount+parityCount; which is largest chunkID
	curStripes int      // largest stripeID
	stripes    sync.Map // key is stripeID_chunkID, value is *cidInfo
}

//lInfo
type lInfo struct {
	chalMap      sync.Map       // key:challenge time,value:*chalresult
	blockMap     sync.Map       // key:bucketid_stripeid_blockid, value: *blockInfo
	chalCid      map[string]int // value is []bucketid_stripeid_blockid
	faultCid     sync.Map
	inChallenge  bool // during challenge or not
	maxlength    int64
	lastChalTime int64
	lastPay      *chalpay // stores result of last pay
	currentPay   *chalpay // current
	stopSign     map[string][]byte
}

type blockInfo struct {
	repair    int // need repair
	availtime int64
	offset    int    // length
	storedOn  string // stored on which provider
}

//chalpay: for one pay informations
type chalpay struct {
	sync.RWMutex
	checkNum int
	stop     bool
	mpb.STValue
	hash []byte
}

type quKey struct {
	uid string
	qid string
}

func (k *Info) getQUKeys() []quKey {
	var res []quKey
	k.ukpGroup.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*groupInfo)
		// filter test user
		if value.upkeeping == nil {
			return true
		}

		if key == value.userID {
			return true
		}

		tmpUQ := quKey{
			uid: value.userID,
			qid: key,
		}
		res = append(res, tmpUQ)

		return true
	})
	return res
}

func (k *Info) getUnpaidUsers() []string {
	var res []string
	k.ukpGroup.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*groupInfo)
		// filter test user
		if value.upkeeping == nil || key == value.userID {
			return true
		}

		res = append(res, key)
		return true
	})
	return res
}

// getGroupsInfo wrap get and create if "mode" is true
func (k *Info) getGroupInfo(userID, groupID string, mode bool) *groupInfo {
	thisIgroup, ok := k.ukpGroup.Load(groupID)
	if ok {
		return thisIgroup.(*groupInfo)
	}

	if mode {
		ginfo, err := k.createGroup(userID, groupID, []string{groupID}, []string{groupID})
		if err != nil {
			return nil
		}

		return ginfo
	}
	return nil
}

// getLInfo wrap get and create if "mode" is true
func (k *Info) getLInfo(userID, groupID, proID string, mode bool) *lInfo {
	thisGroup := k.getGroupInfo(userID, groupID, mode)
	if thisGroup != nil {
		return thisGroup.getLInfo(proID, mode)
	}
	return nil
}

// getBucketInfo wrap get and create if "mode" is true
func (k *Info) getBucketInfo(userID, groupID, bucketID string, mode bool) *bucketInfo {
	thisGroup := k.getGroupInfo(userID, groupID, mode)
	if thisGroup != nil {
		return thisGroup.getBucketInfo(bucketID, mode)
	}

	return nil
}
