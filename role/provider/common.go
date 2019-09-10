package provider

import (
	"errors"
	"runtime"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	//ReDeployOffer redeploy offer-contract,default is false
	ReDeployOffer = false

	DefaultCapacity int64 = 2000   //单位：MB
	DefaultDuration int64 = 200    //单位：天
	DefaultPrice    int64 = 100000 //单位：wei
)

var (
	ErrUnmatchedPeerID         = errors.New("Peer ID is not match")
	ErrProviderServiceNotReady = errors.New("Provider service is not ready")
	ErrGetContractItem         = errors.New("Can't get contract Item")
)

type ProviderContracts struct {
	upKeepingBook sync.Map // K-user的id, V-upkeeping
	channelBook   sync.Map // K-user的id, V-Channel
	queryBook     sync.Map // K-user的id, V-Query
	offer         contracts.OfferItem
}

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

func getNewUserConfig(userID, keeperID string) (*mcl.PublicKey, error) {
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, keeperID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	return mkey.Pk, nil
}

func getUserPrivateKey(userID, keeperID string) (*mcl.SecretKey, error) {
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userconfigkey, keeperID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, posSkByte)
	if err != nil {
		return nil, err
	}

	return mkey.Sk, nil
}
