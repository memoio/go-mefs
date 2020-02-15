package provider

import (
	"context"
	"errors"

	mcl "github.com/memoio/go-mefs/bls12"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	DefaultCapacity int64 = 100000 //单位：MB
	DefaultDuration int64 = 365    //单位：天
	EXPIRETIME            = int64(30 * 60)
)

var (
	errUnmatchedPeerID         = errors.New("Peer ID is not match")
	errProviderServiceNotReady = errors.New("Provider service is not ready")
	errGetContractItem         = errors.New("Can't get contract Item")
)

func (p *Info) getNewUserConfig(userID, groupID string) (*mcl.KeySet, error) {
	gp := p.getGroupInfo(userID, groupID, true)
	if gp == nil {
		return nil, errors.New("No user")
	}

	if gp.blsKey != nil {
		return gp.blsKey, nil
	}

	kmBls12, err := metainfo.NewKey(groupID, mpb.KeyType_Config, userID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	userconfigkey := kmBls12.ToString()
	userconfigbyte, _ := p.ds.GetKey(ctx, userconfigkey, "local")
	if len(userconfigbyte) > 0 {
		mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
		if err == nil && mkey != nil {
			gp.blsKey = mkey
			return mkey, nil
		}
	}

	for _, kid := range gp.keepers {
		userconfigbyte, _ := p.ds.GetKey(ctx, userconfigkey, kid)
		if len(userconfigbyte) > 0 {
			mkey, err := role.BLS12ByteToKeyset(userconfigbyte, nil)
			if err != nil {
				return nil, err
			}

			p.ds.PutKey(ctx, userconfigkey, userconfigbyte, nil, "local")

			gp.blsKey = mkey
			return mkey, nil
		}
	}

	return nil, errors.New("No bls config")
}

func (p *Info) getUserPrivateKey(userID, groupID string) (*mcl.SecretKey, error) {
	gp := p.getGroupInfo(userID, groupID, true)
	if gp == nil {
		return nil, errors.New("No user")
	}

	if gp.blsKey != nil && gp.blsKey.Sk != nil {
		return gp.blsKey.Sk, nil
	}

	kmBls12, err := metainfo.NewKey(groupID, mpb.KeyType_Role, userID)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	userconfigkey := kmBls12.ToString()
	userconfigbyte, err := p.ds.GetKey(ctx, userconfigkey, "local")
	if err != nil {
		return nil, err
	}

	mkey, err := role.BLS12ByteToKeyset(userconfigbyte, posSkByte)
	if err == nil && mkey != nil {
		gp.blsKey = mkey
		return mkey.Sk, nil
	}

	for _, kid := range gp.keepers {
		userconfigbyte, err := p.ds.GetKey(ctx, userconfigkey, kid)
		if err != nil {
			return nil, err
		}
		mkey, err := role.BLS12ByteToKeyset(userconfigbyte, posSkByte)
		if err != nil {
			return nil, err
		}

		p.ds.PutKey(ctx, userconfigkey, userconfigbyte, nil, "local")

		return mkey.Sk, nil
	}

	return nil, errors.New("No bls config")
}

// getDiskUsage gets the disk usage
func (p *Info) getDiskUsage() (uint64, error) {
	return 0, nil
}

// getDiskTotal gets the disk total space which is set in config
// default is 10TB
func (p *Info) getDiskTotal() uint64 {
	maxSpaceInByte := uint64(1024 * 1024 * 1024 * 1024)
	if p.proContract != nil {
		if p.proContract.Capacity != 0 {
			maxSpaceInByte = uint64(p.proContract.Capacity) * 1024 * 1024
		}
	}
	return maxSpaceInByte
}

// getDiskUsage gets the disk total space which is set in config
func getFreeSpace() {
	return
}
