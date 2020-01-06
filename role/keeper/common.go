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
func (k *Info) getUserBLS12Config(userID, groupID string) (*mcl.KeySet, error) {
	thisGroup := k.getGroupInfo(userID, groupID, false)
	if thisGroup == nil {
		return nil, errors.New("No Bls Key")
	}

	if thisGroup.blsKey != nil {
		return thisGroup.blsKey, nil
	}

	userconfigbyte, err := k.getUserBLS12ConfigByte(userID, groupID)
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	thisGroup.blsKey = mkey

	return mkey, nil
}

func (k *Info) getUserBLS12ConfigByte(uid, qid string) ([]byte, error) {
	kmBls12, err := metainfo.NewKeyMeta(qid, metainfo.Config, uid)
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
	gp := k.getGroupInfo(uid, qid, false)
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
