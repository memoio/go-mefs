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
	errNoGroupsInfo          = errors.New("can not find groupInfo")
	errParaseMetaFailed      = errors.New("no enough data in metainfo")
	errNotKeeperInThisGroup  = errors.New("local node is not keeper in this group")
	errPInfoTypeAssert       = errors.New("type asserts err in ukpInfo")
	errNoChalInfo            = errors.New("can not find this chalinfo")
	errGetContractItem       = errors.New("Can't get contract Item")
	errIncorrectParams       = errors.New("Input incorrect params")
)

//---config----
func (k *Info) getUserBLS12Config(groupID, userID string) (*mcl.PublicKey, error) {
	thisGroup := k.getGroupInfo(groupID, userID, false)
	if thisGroup == nil {
		return nil, errors.New("No Bls Key")
	}

	if thisGroup.blsPubKey != nil {
		return thisGroup.blsPubKey, nil
	}

	userconfigbyte, err := k.getUserBLS12ConfigByte(groupID, userID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	thisGroup.blsPubKey = mkey.Pk

	return mkey.Pk, nil
}

func (k *Info) getUserBLS12ConfigByte(qid, uid string) ([]byte, error) {
	kmBls12, err := metainfo.NewKeyMeta(qid, metainfo.Config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := k.ds.GetKey(ctx, userconfigkey, "local")
	if err == nil && userconfigbyte != nil {
		return userconfigbyte, nil
	}
	gp := k.getGroupInfo(qid, uid, false)
	if gp == nil {
		return nil, errors.New("no groupinfo")
	}

	for _, keeperID := range gp.keepers {
		userconfigbyte, err = k.ds.GetKey(ctx, userconfigkey, keeperID)
		if err == nil && userconfigbyte != nil {
			return userconfigbyte, nil
		}
	}

	return nil, errors.New("no user configkey")
}
