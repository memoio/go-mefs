package keeper

import (
	"math/big"
	"sync"

	"github.com/google/uuid"
	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
)

// user-keeper-provider map
type groupInfo struct {
	sessionID    uuid.UUID // for user
	groupID      string    // is queryID
	owner        string    // is userID
	localKeeper  string
	masterKeeper string
	keepers      []string
	providers    []string
	upkeeping    *contracts.UpKeepingItem
	query        *contracts.QueryItem
	blsPubKey    *mcl.PublicKey
	buckets      sync.Map // key:bucketID; value: *bucketInfo
	ledgerMap    sync.Map // key:proID，value:*chalInfo
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
type uqKey struct {
	uid string
	qid string
}

func (k *Info) getUQKeys() []uqKey {
	var res []uqKey
	k.ukpGroup.Range(func(k, v interface{}) bool {
		key := k.(string)
		value := v.(*groupInfo)
		// filter test user
		if value.upkeeping == nil {
			return true
		}

		if key == value.owner {
			return true
		}

		tmpUQ := uqKey{
			uid: value.owner,
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
		if value.upkeeping == nil || key == value.owner {
			return true
		}

		res = append(res, key)
		return true
	})
	return res
}

// getGroupsInfo wrap get and create if "mode" is true
func (k *Info) getGroupInfo(groupID, userID string, mode bool) *groupInfo {
	thisIgroup, ok := k.ukpGroup.Load(groupID)
	if !ok {
		if mode {
			ginfo, err := newGroup(k.localID, groupID, userID, []string{groupID}, []string{groupID})
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
func (k *Info) getLInfo(groupID, userID, proID string, mode bool) *lInfo {
	thisGroup := k.getGroupInfo(groupID, userID, mode)
	if thisGroup != nil {
		return thisGroup.getLInfo(proID, mode)
	}
	return nil
}

// getBucketInfo wrap get and create if "mode" is true
func (k *Info) getBucketInfo(groupID, userID, bucketID string, mode bool) *bucketInfo {
	thisGroup := k.getGroupInfo(groupID, userID, mode)
	if thisGroup != nil {
		return thisGroup.getBucketInfo(bucketID, mode)
	}

	return nil
}
