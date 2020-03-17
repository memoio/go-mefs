package keeper

import (
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

const (
	expireTime       = int64(30 * 60) //超过这个时间，触发修复，单位：秒
	chalTime         = 5 * time.Minute
	chalRepairTime   = 7 * time.Minute
	persistTime      = 3 * time.Minute
	spaceTimePayTime = 61 * time.Minute
	checkConnTime    = 5 * time.Minute
	kpMapTime        = 11 * time.Minute
)

//---config----
func (k *Info) getUserBLS12Config(userID, groupID string) (*mcl.KeySet, error) {
	thisGroup := k.getGroupInfo(userID, groupID, false)
	if thisGroup == nil {
		return nil, role.ErrNotMyUser
	}

	if thisGroup.blsKey != nil {
		return thisGroup.blsKey, nil
	}

	userconfigbyte, err := k.getUserBLS12ConfigByte(userID, groupID)
	if err != nil {
		if userID == pos.GetPosId() {
			mkey, err := mcl.GenKeySetWithSeed(pos.GetPosSeed(), mcl.TagAtomNum, mcl.PDPCount)
			if err != nil {
				return nil, err
			}

			thisGroup.blsKey = mkey

			return mkey, nil
		}
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
	kmBls12, err := metainfo.NewKey(qid, mpb.KeyType_Config, uid)
	if err != nil {
		return nil, err
	}

	ctx := k.context

	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := k.ds.GetKey(ctx, userconfigkey, "local")
	if err == nil && userconfigbyte != nil {
		return userconfigbyte, nil
	}
	gp := k.getGroupInfo(uid, qid, false)
	if gp == nil {
		return nil, role.ErrNotMyUser
	}

	for _, keeperID := range gp.keepers {
		if keeperID != k.localID {
			userconfigbyte, err = k.ds.GetKey(ctx, userconfigkey, keeperID)
			if err == nil && userconfigbyte != nil {
				return userconfigbyte, nil
			}
		}
	}

	return nil, role.ErrEmptyBlsKey
}
