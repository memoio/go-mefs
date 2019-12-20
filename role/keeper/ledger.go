package keeper

import (
	"log"
	"strconv"
	"sync"

	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//puKey 作为其他结构体的key，
//目前用PU做key的结构体有LedgerInfo，PayInfo,用pid和uid共同索引信息
type puKey struct {
	pid string
	uid string
}

//chalinfo 作为LedgerInfo的value key是PU对
type chalinfo struct {
	chalMap      sync.Map // key is challenge time(int64),value is *chalresult
	cidMap       sync.Map // key is bucketid_stripeid_blockid, value is *cidInfo
	chalCid      []string // value is []bucketid_stripeid_blockid
	inChallenge  bool     // during challenge or not
	maxlength    int64
	lastChalTime int64
	lastPay      *chalpay // stores result of last pay
}

type cidInfo struct {
	res       bool
	repair    int32 // need repair
	availtime int64
	offset    int    // length of this cid
	storedOn  string // stored on which provider
}

//chalresult 挑战结果在内存中的结构
//作为chalinfo.Time的value 记录单次挑战的各项信息
type chalresult struct {
	kid           string //挑战发起者
	pid           string //挑战对象
	uid           string //挑战的数据所属对象
	challengeTime int64  //挑战发起时间 使用unix时间戳
	totalSpace    int64  //the amount of this user's data on the provider
	sum           int64  //挑战总空间
	length        int64  //挑战成功空间
	h             int    //挑战的随机数
	res           bool   //挑战是否成功
	proof         string //挑战结果的证据
}

//key: PU，value: *chalinfo
var ledgerInfo sync.Map

func getChalinfo(thisPU puKey) (*chalinfo, bool) {
	thischalinfo, ok := ledgerInfo.Load(thisPU)
	if !ok {
		gp, ok := getGroupsInfo(thisPU.uid)
		if ok && gp != nil {
			newchalinfo := &chalinfo{}
			ledgerInfo.Store(thisPU, newchalinfo)
		}
	}

	thischalinfo, ok = ledgerInfo.Load(thisPU)
	if !ok {
		return nil, false
	}

	return thischalinfo.(*chalinfo), true
}

func getChalresult(thisPU puKey, time int64) (*chalresult, bool) {
	thischalinfo, ok := getChalinfo(thisPU)
	if !ok || thischalinfo == nil {
		return nil, false
	}

	thischalresult, ok := thischalinfo.chalMap.Load(time)
	if !ok {
		return nil, false
	}
	return thischalresult.(*chalresult), true
}

func addCidinfotoMem(uid, pid, blockid string, newCidinfo *cidInfo) error {
	cidString, err := metainfo.GetCidFromBlock(blockid)
	if err != nil {
		return err
	}

	newCidInfo, err := doAddCidinfotoLedger(uid, pid, cidString, newCidinfo)
	if err != nil {
		return err
	}
	return doAddBlocktoBucket(uid, cidString, newCidInfo)
}

func addBlocktoMem(uid, pid, blockid string, offset int) error {
	log.Println("add block: ", blockid, "to user: ", uid, " and provider: ", pid)
	cidString, err := metainfo.GetCidFromBlock(blockid)
	if err != nil {
		return err
	}

	newCidInfo, err := doAddBlocktoLedger(uid, pid, cidString, offset)
	if err != nil {
		return err
	}
	return doAddBlocktoBucket(uid, cidString, newCidInfo)
}

func doAddBlocktoLedger(uid, pid, cidString string, offset int) (*cidInfo, error) {
	newcidinfo := &cidInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
		storedOn:  pid,
	}

	return doAddCidinfotoLedger(uid, pid, cidString, newcidinfo)
}

func doAddCidinfotoLedger(uid, pid, cidString string, newCidinfo *cidInfo) (*cidInfo, error) {

	pu := puKey{
		pid: pid,
		uid: uid,
	}

	thisinfo, ok := ledgerInfo.Load(pu)
	oldOffset := 0
	offset := newCidinfo.offset
	if ok {
		thechalinfo := thisinfo.(*chalinfo)

		act, loaded := thechalinfo.cidMap.Load(cidString)
		if loaded {
			oldOffset = act.(*cidInfo).offset
			if oldOffset < offset {
				newCidinfo.offset = offset
			} else {
				// stored length is longer
				return act.(*cidInfo), nil
			}
		} else {
			oldOffset = -1
		}
		thechalinfo.maxlength += (int64(offset-oldOffset) * df.DefaultSegmentSize)

		thechalinfo.cidMap.Store(cidString, newCidinfo)
	} else {
		newchalinfo := &chalinfo{
			maxlength: int64(offset+1) * df.DefaultSegmentSize,
		}
		newchalinfo.cidMap.Store(cidString, newCidinfo)
		ledgerInfo.Store(pu, newchalinfo)
	}

	return newCidinfo, nil
}

func doAddBlocktoBucket(uid, cid string, newcidinfo *cidInfo) error {
	bid, sid, chunkID, err := metainfo.GetIDsFromBlock(cid)
	if err != nil {
		return err
	}

	thisBucketinfo, ok := getBucketInfo(uid, bid)
	if !ok {
		return nil
	}

	thisBucketinfo.stripes.Store(cid, newcidinfo)

	snum, err := strconv.Atoi(sid)
	if err != nil {
		return err
	}

	if int32(snum) > thisBucketinfo.largestStripes {
		thisBucketinfo.largestStripes = int32(snum)
	}

	cnum, err := strconv.Atoi(chunkID)
	if err != nil {
		return err
	}

	if int32(cnum) > thisBucketinfo.chunkNum {
		thisBucketinfo.chunkNum = int32(cnum)
	}

	return nil
}

func deleteBlockFromMem(uid, pid, cidString string) {
	cidString, err := metainfo.GetCidFromBlock(cidString)
	if err != nil {
		return
	}
	deleteBlockInLedger(uid, pid, cidString)
	deleteBlockFromBucket(uid, cidString)
}

func deleteBlockInLedger(uid, pid, cidString string) {
	pu := puKey{
		pid: pid,
		uid: uid,
	}

	if thischalinfo, ok := getChalinfo(pu); ok {
		thiscid, ok := thischalinfo.cidMap.Load(cidString)
		if ok {
			thischalinfo.maxlength -= (int64(thiscid.(*cidInfo).offset+1) * df.DefaultSegmentSize)
		}
		thischalinfo.cidMap.Delete(cidString)
	}
}

func deleteBlockFromBucket(uid, cidString string) {
	bid, _, _, err := metainfo.GetIDsFromBlock(cidString)
	if err != nil {
		return
	}

	thisgroup, ok := getGroupsInfo(uid)
	if !ok {
		return
	}

	thisbucket, ok := thisgroup.buckets.Load(bid)
	if ok {
		thisbucket.(*bucketInfo).stripes.Delete(cidString)
	}
}
