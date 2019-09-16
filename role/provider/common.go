package provider

import (
	"errors"
	"log"
	"runtime"
	"sync"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	//ReDeployOffer redeploy offer-contract,default is false
	ReDeployOffer = false

	DefaultCapacity int64 = 100000 //单位：MB
	DefaultDuration int64 = 365    //单位：天
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
	pubKeyI, ok := usersConfigs.Load(userID)
	if ok {
		return pubKeyI.(*mcl.PublicKey), nil
	}

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

	usersConfigs.Store(userID, mkey.Pk)

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

// getDiskUsage gets the disk usage
func getDiskUsage() (uint64, error) {
	dataStore := localNode.Repo.Datastore()
	DataSpace, err := ds.DiskUsage(dataStore)
	if err != nil {
		log.Println("get disk usage failed :", err)
		return 0, err
	}
	return DataSpace, nil
}

// getDiskTotal gets the disk total space which is set in config
func getDiskTotal() uint64 {
	var maxSpaceInByte uint64
	offerItem, err := GetOffer()
	if err != nil {
		maxSpaceInByte = 10 * 1024 * 1024 * 1024
	} else {
		if offerItem.Capacity == 0 {
			maxSpaceInByte = 10 * 1024 * 1024 * 1024
		}
		maxSpaceInByte = uint64(offerItem.Capacity) * 1024 * 1024
	}
	return maxSpaceInByte
}

// getDiskUsage gets the disk total space which is set in config
func getFreeSpace() {
	return
}
