package keeper

import (
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// user-keeper-provider map
type groupInfo struct {
	sessionID    uuid.UUID // for user
	groupID      string    // is queryID
	userID       string    // is userID
	localKeeper  string
	masterKeeper string
	keepers      []string
	providers    []string
	upkeeping    *role.UpKeepingItem
	query        *role.QueryItem
	blsKey       *mcl.KeySet
	buckets      sync.Map // key:bucketID; value: *bucketInfo
	ledgerMap    sync.Map // key:proID，value:*chalInfo
}

func newGroup(localID, uid, qid string, keepers, providers []string) (*groupInfo, error) {
	tempInfo := &groupInfo{
		groupID:      qid,
		userID:       uid,
		localKeeper:  qid,
		masterKeeper: qid,
		keepers:      keepers,
		providers:    providers,
		sessionID:    uuid.New(),
	}

	if qid != uid {
		err := tempInfo.loadContracts(true)
		if err != nil {
			return nil, err
		}
	} else {
		for _, keeperID := range tempInfo.keepers {
			if localID == keeperID {
				tempInfo.localKeeper = localID
			}
		}

		// not my user
		if tempInfo.localKeeper != localID {
			utils.MLogger.Warn(uid, " is not my user, keepers are: ", keepers)
			return nil, errors.New("Not my user")
		}
	}

	tempInfo.masterKeeper = getMasterID(tempInfo.keepers)

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
			return getMasterID(mymaster) == g.localKeeper
		}
	}

	return g.localKeeper == g.masterKeeper
}

//getMasterID choose the middle nodes
func getMasterID(kidlist []string) string {
	sort.Strings(kidlist)
	return kidlist[len(kidlist)/2]
}

func (g *groupInfo) getLInfo(proID string, mode bool) *lInfo {
	if g == nil {
		return nil
	}

	thisIl, ok := g.ledgerMap.Load(proID)
	if !ok {
		if mode {
			templInfo := &lInfo{}
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
			tempInfo := &bucketInfo{}
			g.buckets.Store(bucketID, tempInfo)
			return tempInfo
		}
		return nil
	}

	return thisIb.(*bucketInfo)
}

// bid is bucketID_stripeID_chunkID
func (g *groupInfo) addBlockMeta(bid, pid string, offset int) error {

	thisLinfo := g.getLInfo(pid, true)
	if thisLinfo == nil {
		return errors.New("group addBlockMeta err")
	}

	newcidinfo := &blockInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
		storedOn:  pid,
	}

	// store in cidMap
	oldOffset := -1
	v, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		newcidinfo = v.(*blockInfo)
		oldOffset = newcidinfo.offset
		if offset > oldOffset {
			newcidinfo.offset = oldOffset
		}

		if newcidinfo.storedOn != pid {
			newcidinfo.storedOn = pid
		}
	} else {
		thisLinfo.blockMap.Store(bid, newcidinfo)
	}

	thisLinfo.maxlength += (int64(offset-oldOffset) * df.DefaultSegmentSize)

	// store in buckets
	bucketID, stripeID, chunkID, err := metainfo.GetIDsFromBlock(bid)
	if err != nil {
		return err
	}

	thisBucket := g.getBucketInfo(bucketID, true)
	if thisBucket == nil {
		return errors.New("cannot create bucket info")
	}

	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)
	// key: stripeID_chunkID
	thisBucket.stripes.Store(bids[1], newcidinfo)

	snum, err := strconv.Atoi(stripeID)
	if err != nil {
		return err
	}

	if snum > thisBucket.curStripes {
		thisBucket.curStripes = snum
	}

	cnum, err := strconv.Atoi(chunkID)
	if err != nil {
		return err
	}

	if cnum > thisBucket.chunkNum {
		thisBucket.chunkNum = cnum
	}

	return nil
}

func (g *groupInfo) deleteBlockMeta(bid, pid string) {
	thisLinfo := g.getLInfo(pid, false)
	if thisLinfo == nil {
		return
	}

	// delete from blockMap
	thisICid, ok := thisLinfo.blockMap.Load(bid)
	if ok {
		thisCid := thisICid.(*blockInfo)
		thisLinfo.maxlength -= (int64(thisCid.offset+1) * df.DefaultSegmentSize)
		thisLinfo.blockMap.Delete(bid)
	}

	// delete from buckets
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		bui.(*bucketInfo).stripes.Delete(bids[1])
	}

	return
}

// bucketID_stripeID_chunkID
func (g *groupInfo) getBlockPos(bid string) (string, error) {
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		sti, ok := bui.(*bucketInfo).stripes.Load(bids[1])
		if ok {
			return sti.(*blockInfo).storedOn, nil
		}
	}

	return "", errors.New("No such block")
}

func (g *groupInfo) getBlockAvail(bid string) (int64, error) {
	bids := strings.SplitN(bid, metainfo.BLOCK_DELIMITER, 2)

	bui, ok := g.buckets.Load(bids[0])
	if ok {
		sti, ok := bui.(*bucketInfo).stripes.Load(bids[1])
		if ok {
			return sti.(*blockInfo).availtime, nil
		}
	}

	return 0, errors.New("No such block")
}

type bucketInfo struct {
	bucketID    int
	dataCount   int
	parityCount int
	chunkNum    int      // = dataCount+parityCount; which is largest chunkID
	curStripes  int      // largest stripeID
	stripes     sync.Map // key is stripeID_chunkID, value is *cidInfo
}

//lInfo
type lInfo struct {
	chalMap      sync.Map // key:challenge time,value:*chalresult
	blockMap     sync.Map // key:bucketid_stripeid_blockid, value: *blockInfo
	chalCid      []string // value is []bucketid_stripeid_blockid
	inChallenge  bool     // during challenge or not
	maxlength    int64
	lastChalTime int64
	lastPay      *chalpay // stores result of last pay
}

type blockInfo struct {
	repair    int // need repair
	availtime int64
	offset    int    // length - 1
	storedOn  string // stored on which provider
}

//chalresult 挑战结果在内存中的结构
//作为chalInfo的value 记录单次挑战的各项信息
type chalresult struct {
	kid        string //挑战发起者
	pid        string //挑战对象
	qid        string //挑战的数据所属对象
	chalTime   int64  //挑战发起时间 使用unix时间戳
	totalSpace int64  //the amount of this user's data on the provider
	sum        int64  //挑战总空间
	length     int64  //挑战成功空间
	h          int    //挑战的随机数
	res        bool   //挑战是否成功
	proof      string //挑战结果的证据
}

//chalpay: for one pay informations
type chalpay struct {
	beginTime int64    // last end
	endTime   int64    // this end
	spacetime *big.Int // space time value
	signature string   // signature of spacetime
	proof     string
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
			//return true
		}

		if key == value.userID {
			//return true
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
	if !ok {
		if mode {
			ginfo, err := newGroup(k.localID, userID, groupID, []string{groupID}, []string{groupID})
			if err != nil {
				return nil
			}
			k.ukpGroup.Store(groupID, ginfo)
			return ginfo
		}
		return nil

	}

	return thisIgroup.(*groupInfo)
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
