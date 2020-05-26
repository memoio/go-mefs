package provider

import (
	mcl "github.com/memoio/go-mefs/bls12"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	datastore "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	expireTime = int64(30 * 60)
)

func (p *Info) getNewUserConfig(userID, groupID string) (*mcl.KeySet, error) {
	value, ok := p.userConfigs.Get(groupID)
	if ok {
		return value.(*mcl.KeySet), nil
	}

	kmBls12, err := metainfo.NewKey(groupID, mpb.KeyType_Config, userID)
	if err != nil {
		return nil, err
	}

	ctx := p.context
	userconfigkey := kmBls12.ToString()
	userconfigbyte, _ := p.ds.GetKey(ctx, userconfigkey, "local")
	if len(userconfigbyte) > 0 {
		mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
		if err == nil && mkey != nil {
			p.userConfigs.Add(groupID, mkey)
			return mkey, nil
		}
	}

	gp := p.getGroupInfo(userID, groupID, true)
	if gp == nil {
		return nil, role.ErrNotMyUser
	}

	for _, kid := range gp.keepers {
		userconfigbyte, _ := p.ds.GetKey(ctx, userconfigkey, kid)
		if len(userconfigbyte) > 0 {
			mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
			if err != nil {
				return nil, err
			}

			p.ds.PutKey(ctx, userconfigkey, userconfigbyte, nil, "local")

			p.userConfigs.Add(groupID, mkey)
			return mkey, nil
		}
	}

	return nil, role.ErrEmptyBlsKey
}

// getDiskUsage gets the disk usage
func (p *Info) getDiskUsage() (uint64, error) {
	used, err := datastore.DiskUsage(p.ds.DataStore())
	if err != nil {
		return 0, err
	}

	rootpath, err := fsrepo.BestKnownPath()
	if err != nil {
		return 0, err
	}

	localUsed, err := utils.GetDirSize(rootpath)
	if err != nil {
		return 0, err
	}

	if used >= localUsed {
		return localUsed, nil
	}

	return used, nil
}

// getDiskTotal gets the disk total space which is set in config
// default is 1TB
func (p *Info) getDiskTotal() uint64 {
	maxSpaceInByte := uint64(1024 * 1024 * 1024 * 1024)
	if p.proContract != nil && p.proContract.Capacity != 0 {
		maxSpaceInByte = uint64(p.proContract.Capacity) * 1024 * 1024
	}
	return maxSpaceInByte
}
