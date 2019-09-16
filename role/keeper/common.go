package keeper

import (
	"errors"
	"runtime"

	"github.com/memoio/go-mefs/utils/metainfo"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
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
	ErrGetContractItem       = errors.New("Can't get contract Item")
	ErrIncorrectParams       = errors.New("Input incorrect params.")
)

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

//获得用于证明的user的公用参数
func getUserBLS12Config(userID string) (*mcl.PublicKey, error) {
	pubKeyI, ok := usersConfigs.Load(userID)
	if ok {
		return pubKeyI.(*mcl.PublicKey), nil
	}

	userconfigbyte, err := getUserBLS12ConfigByte(userID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	usersConfigs.Store(userID, mkey.Pk)

	return mkey.Pk, nil
}

func getUserBLS12ConfigByte(userID string) ([]byte, error) {
	if !IsKeeperServiceRunning() {
		return nil, ErrKeeperServiceNotReady
	}
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, "local")
	if err != nil {
		return nil, err
	}
	return userconfigbyte, nil
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
