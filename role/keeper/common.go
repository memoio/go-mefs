package keeper

import (
	"errors"
	"runtime"
	"time"

	"github.com/memoio/go-mefs/utils/metainfo"

	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
)

var (
	ErrKeeperServiceNotReady = errors.New("keeper service is not ready")
	ErrUnmatchedPeerID       = errors.New("peer ID is not match")
	ErrBlockNotExist         = errors.New("block does not exist")
	ErrNoGroupsInfo          = errors.New("can not find groupsInfo")
	ErrParaseMetaFailed      = errors.New("no enough data in metainfo")
	ErrNotKeeperInThisGroup  = errors.New("local node is not keeper in this group")
	ErrPInfoTypeAssert       = errors.New("type asserts err in PInfo")
	ErrNoChalInfo            = errors.New("can not find this chalinfo")
)

//tendermint启动时，获取启动信息中所需的时间参数。同组内时间参数需一致，目前用该函数获取
func getTendermintTime() time.Time {
	tendermintTime, _ := time.Parse("2006-01-02 15:04:05", "2019-03-30 13:00:00")
	return tendermintTime
}

func addCredit(provider string) {
	val, ok := localPeerInfo.Credit.Load(provider)
	if !ok {
		localPeerInfo.Credit.Store(provider, 100)
	} else {
		cre := val.(int)
		cre++
		if cre >= 100 {
			cre = 100
		}
		localPeerInfo.Credit.Store(provider, cre)
	}
}

func reduceCredit(provider string) {
	val, ok := localPeerInfo.Credit.Load(provider)
	if !ok {
		localPeerInfo.Credit.Store(provider, 100)
	} else {
		cre := val.(int)
		cre--
		if cre <= 0 {
			cre = 0
		}
		localPeerInfo.Credit.Store(provider, cre)
	}
}

//=============v2版本信息结构,上面的信息修改后逐渐删除===============
func sendMetaMessage(km *metainfo.KeyMeta, metaValue, to string) error {
	caller := ""
	for _, i := range []int{0, 1, 2, 3, 4} {
		pc, _, _, _ := runtime.Caller(i)
		caller += string(i) + ":" + runtime.FuncForPC(pc).Name() + "\n"
	}
	return localNode.Routing.(*dht.IpfsDHT).SendMetaMessage(km.ToString(), metaValue, to, caller)
}

func sendMetaRequest(km *metainfo.KeyMeta, metaValue, to string) (string, error) {
	caller := ""
	for _, i := range []int{0, 1, 2, 3, 4} {
		pc, _, _, _ := runtime.Caller(i)
		caller += string(i) + ":" + runtime.FuncForPC(pc).Name() + "\n"
	}
	return localNode.Routing.(*dht.IpfsDHT).SendMetaRequest(km.ToString(), metaValue, to, caller)
}
