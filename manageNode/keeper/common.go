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
	expireTime       = int64(60 * 60) //超过这个时间，触发修复，单位：秒
	rafiTime         = int64(10 * 60)
	chalTime         = 5 * time.Minute
	chalRepairTime   = 7 * time.Minute
	persistTime      = 3 * time.Minute
	spaceTimePayTime = 61 * time.Minute
	checkConnTime    = 5 * time.Minute
	kpMapTime        = 11 * time.Minute
)

// MarketingMoney is used to
var MarketingMoney int64 = 1

//---config----
func (k *Info) getUserBLS12Config(userID, groupID string) (*mcl.KeySet, error) {
	value, ok := k.userConfigs.Get(groupID)
	if ok {
		return value.(*mcl.KeySet), nil
	}

	userconfigbyte, err := k.getUserBLS12ConfigByte(userID, groupID)
	if err != nil {
		if userID == pos.GetPosId() {
			mkey, err := mcl.GenKeySetWithSeed(pos.GetPosSeed(), mcl.TagAtomNum, mcl.PDPCount)
			if err != nil {
				return nil, err
			}

			k.userConfigs.Add(groupID, mkey)
			return mkey, nil
		}
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
	if err != nil {
		return nil, err
	}

	k.userConfigs.Add(groupID, mkey)
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
