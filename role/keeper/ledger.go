package keeper

import (
	"errors"
	"log"
	"sync"

	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
)

type ledger struct {
	localID string
	ds      data.Service
	lMap    sync.Map //key:pqKey，value:*groupsInfo
}

type pqKey struct {
	pid string // providerID
	qid string // queryID
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

func (l *ledger) getChalinfo(thisPU pqKey) (*chalinfo, bool) {
	thischalinfo, ok := l.lMap.Load(thisPU)
	if !ok {
		return nil, false
	}

	return thischalinfo.(*chalinfo), true
}

// bid is bucketID_stripeID_chunkID
func (l *ledger) addBlockMeta(qid, pid, bid string, offset int) (*cidInfo, error) {
	log.Println("add block: ", bid, "for query: ", qid, " and provider: ", pid)

	pu := pqKey{
		pid: pid,
		qid: qid,
	}

	thisChal, ok := l.getChalinfo(pu)
	if !ok {
		return nil, errors.New("cannot create chalinfo")
	}

	newcidinfo := &cidInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
		storedOn:  pid,
	}

	oldOffset := -1
	v, ok := thisChal.cidMap.Load(bid)
	if ok {
		newcidinfo = v.(*cidInfo)
		oldOffset = newcidinfo.offset
		if offset > oldOffset {
			newcidinfo.offset = oldOffset
		}

		if newcidinfo.storedOn != pid {
			newcidinfo.storedOn = pid
		}
	} else {
		thisChal.cidMap.Store(bid, newcidinfo)
	}

	thisChal.maxlength += (int64(offset-oldOffset) * df.DefaultSegmentSize)

	return newcidinfo, nil
}

func (l *ledger) deleteBlockMeta(qid, pid, bid string) {
	pu := pqKey{
		pid: pid,
		qid: qid,
	}

	if thischalinfo, ok := l.lMap.Load(pu); ok {
		thischal := thischalinfo.(*chalinfo)
		thiscid, ok := thischal.cidMap.Load(bid)
		if ok {
			thischal.maxlength -= (int64(thiscid.(*cidInfo).offset+1) * df.DefaultSegmentSize)
			thischal.cidMap.Delete(bid)
		}
	}
}
