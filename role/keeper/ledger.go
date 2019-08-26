package keeper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/memoio/go-mefs/utils"
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

//chalinfo 作为LedgerInfo的value
type chalinfo struct {
	Time        sync.Map //某provider下user数据在某时刻发起挑战的结果，key为挑战发起时间的时间戳，格式为int64,value为chalresult
	Cid         sync.Map
	tmpCid      sync.Map
	inChallenge int
	maxlength   uint32
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
		}
		return nil
	}

	var Cid sync.Map
	var Time sync.Map
	Cid.Store(blockid, newcidinfo)
	newchalinfo := &chalinfo{
		Cid:  Cid,
		Time: Time,
	}
	LedgerInfo.Store(pu, newchalinfo)
	return nil

}

func checkLedger(ctx context.Context) {
	fmt.Println("CheckLedger() start!")
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
					thischalinfo := v.(*chalinfo)
					thischalinfo.Cid.Range(func(key, value interface{}) bool {
						//fmt.Println("avaltime :", value.(*cidInfo).availtime)
						if EXPIRETIME < (utils.GetUnixNow()-value.(*cidInfo).availtime) && value.(*cidInfo).repair < 3 {
							fmt.Println("need repair cid :", key.(string))
							value.(*cidInfo).repair += 1
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
