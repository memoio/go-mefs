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
	errKeeperServiceNotReady = errors.New("keeper service is not ready")
	errUnmatchedPeerID       = errors.New("peer ID is not match")
	errBlockNotExist         = errors.New("block does not exist")
	errNoGroupsInfo          = errors.New("can not find groupsInfo")
	errParaseMetaFailed      = errors.New("no enough data in metainfo")
	errNotKeeperInThisGroup  = errors.New("local node is not keeper in this group")
	errPInfoTypeAssert       = errors.New("type asserts err in ukpInfo")
	errNoChalInfo            = errors.New("can not find this chalinfo")
	errGetContractItem       = errors.New("Can't get contract Item")
	errIncorrectParams       = errors.New("Input incorrect params")
)

//---config----
func getUserBLS12Config(userID string) (*mcl.PublicKey, error) {
	thisInfo, err := getUInfo(userID)
	if err != nil {
		return nil, err
	}

	if thisInfo.pubKey != nil {
		return thisInfo.pubKey, nil
	}

	userconfigbyte, err := getUserBLS12ConfigByte(userID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	thisInfo.pubKey = mkey.Pk

	return mkey.Pk, nil
}

func getUserBLS12ConfigByte(userID string) ([]byte, error) {
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return nil, err
	}
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := getKeyFrom(userconfigkey, "local")
	if err == nil {
		return userconfigbyte, nil
	}
	gp, ok := getGroupsInfo(userID)
	if !ok {
		return nil, errors.New("no groupinfo")
	}

	for _, keeperID := range gp.keepers {
		userconfigbyte, err = getKeyFrom(userconfigkey, keeperID)
		if err == nil {
			putKeyTo(userconfigkey, string(userconfigbyte), "local")
			return userconfigbyte, nil
		}
	}

	return nil, errors.New("no user configkey")
}

//---network---
func putKeyTo(key, value, node string) error {
	return localNode.Routing.(*dht.IpfsDHT).CmdPutTo(key, value, node)
}

func getKeyFrom(key, node string) ([]byte, error) {
	return localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(key, node)
}

func deleteFrom(key, node string) error {
	if node == "local" {
		return localNode.Routing.(*dht.IpfsDHT).DeleteLocal(key)
	}
	return nil
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
