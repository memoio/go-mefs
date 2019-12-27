package keeper

import (
	"context"
	"errors"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/metainfo"
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
	kmBls12, err := metainfo.NewKeyMeta(userID, metainfo.Config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := localNode.Data.GetKey(ctx, userconfigkey, "local")
	if err == nil && userconfigbyte != nil {
		return userconfigbyte, nil
	}
	gp, ok := getGroupsInfo(userID)
	if !ok {
		return nil, errors.New("no groupinfo")
	}

	for _, keeperID := range gp.keepers {
		if keeperID == localNode.Identity.Pretty() {
			continue
		}
		userconfigbyte, err = localNode.Data.GetKey(ctx, userconfigkey, keeperID)
		if err == nil && userconfigbyte != nil {
			return userconfigbyte, nil
		}
	}

	return nil, errors.New("no user configkey")
}
