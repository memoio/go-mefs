package keeper

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"

	df "github.com/memoio/go-mefs/data-format"
)

//LedgerInfo 存放挑战信息的内存结构体
//key为结构体PU，value为结构体*chalinfo
var LedgerInfo sync.Map //key为PU结构体，value为chalinfo结构体

//PU 作为其他结构体的key，
//目前用PU做key的结构体有LedgerInfo，PayInfo,用pid和uid共同索引信息
type PU struct {
	pid string
	uid string
}

//chalinfo 作为LedgerInfo的value key是PU对
type chalinfo struct {
	Time        sync.Map //某provider下user数据在某时刻发起挑战的结果，key为挑战发起时间的时间戳，格式为int64,value为chalresult
	Cid         sync.Map
	tmpCid      sync.Map
	chalCid     []string //当前挑战的块
	inChallenge int
	maxlength   int64
	testuser    bool // 记录是否是test用户，用于不发起挑战
}

type cidInfo struct {
	res       bool
	repair    int32
	availtime int64
	offset    int
}

//getChalinfo 输入pid和uid 获取LedgerInfo中的chalinfo指针
func getChalinfo(thisPU PU) (*chalinfo, bool) {
	thischalinfo, ok := LedgerInfo.Load(thisPU)
	if !ok {
		return nil, false
	}
	return thischalinfo.(*chalinfo), true
}

//doAddBlocktoLedger 将block信息加入本地LedgerInfo结构体里的Cid字段，用于挑战
func doAddBlocktoLedger(pid string, uid string, blockid string, offset int) error {
	pu := PU{
		pid: pid,
		uid: uid,
	}

	newcidinfo := &cidInfo{
		availtime: utils.GetUnixNow(),
		offset:    offset,
		repair:    0,
	}

	if thischalinfo, ok := getChalinfo(pu); ok {
		if thischalinfo.inChallenge == 1 {
			thischalinfo.tmpCid.Store(blockid, newcidinfo)
		} else if thischalinfo.inChallenge == 0 {
			thischalinfo.Cid.Store(blockid, newcidinfo)
			thischalinfo.maxlength += int64(offset * df.DefaultSegmentSize)
		}
		return nil
	}

	var Cid sync.Map
	var Time sync.Map
	Cid.Store(blockid, newcidinfo)

	isTestUser := false
	addr, err := address.GetAddressFromID(pu.uid)
	if err == nil {
		_, _, err = contracts.GetUKFromResolver(addr)
		if err != nil {
			isTestUser = true
		}
	}

	newchalinfo := &chalinfo{
		Cid:      Cid,
		Time:     Time,
		testuser: isTestUser,
	}
	LedgerInfo.Store(pu, newchalinfo)
	return nil

}

//deleteBlockInLedger 删除LedgerInfo中的块信息，传入保存这个块的pid和块信息结构体
func deleteBlockInLedger(pid string, bm *metainfo.BlockMeta) {
	pu := PU{
		pid: pid,
		uid: bm.GetUid(),
	}
	if thischalinfo, ok := getChalinfo(pu); ok {
		thischalinfo.Cid.Delete(bm.ToString())
		thischalinfo.tmpCid.Delete(bm.ToString())
	}
}

func checkLedger(ctx context.Context) {
	log.Println("Check Ledger Start!")
	time.Sleep(2 * CHALTIME)
	ticker := time.NewTicker(CHECKTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			go func() {
				LedgerInfo.Range(func(k, v interface{}) bool {
					pu := k.(PU)
					if pu.uid == pos.GetPosId() {
						return true
					}

					thischalinfo := v.(*chalinfo)
					thischalinfo.Cid.Range(func(key, value interface{}) bool {
						//log.Println("avaltime :", value.(*cidInfo).availtime)
						if EXPIRETIME < (utils.GetUnixNow()-value.(*cidInfo).availtime) && value.(*cidInfo).repair < 3 {
							log.Println("Need repair cid: ", key.(string))
							value.(*cidInfo).repair++
							repch <- key.(string)
						}
						return true
					})
					return true
				})
			}()
		}
	}
}
