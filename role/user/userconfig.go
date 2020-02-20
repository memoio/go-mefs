package user

import (
	"context"

	mcl "github.com/memoio/go-mefs/bls12"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func initBLS12Config(seed []byte) (*mcl.KeySet, error) {
	utils.MLogger.Info("Generating BLS12 Sk and Pk")
	kset, err := mcl.GenKeySetWithSeed(seed)
	if err != nil {
		utils.MLogger.Error("Init BlS12 keyset error: ", err)
		return nil, err
	}
	return kset, nil
}

func parseBLS12ConfigMeta(privKey, userBLS12config []byte) (*mcl.KeySet, error) {
	mkey, err := role.BLS12ByteToKeyset(userBLS12config, privKey)
	if err != nil {
		return nil, err
	}

	return mkey, nil
}

func (l *LfsInfo) putUserConfig(ctx context.Context) {
	kmBls, err := metainfo.NewKey(l.fsID, mpb.KeyType_Config, l.userID)
	if err != nil {
		return
	}

	userBLS12Config, err := role.BLS12KeysetToByte(l.keySet, []byte(l.privateKey))
	if err != nil {
		return
	}

	blskey := kmBls.ToString()

	l.gInfo.putToAll(ctx, blskey, userBLS12Config)
}

func (l *LfsInfo) loadBLS12Config() error {
	kmBls, err := metainfo.NewKey(l.fsID, mpb.KeyType_Config, l.userID)
	if err != nil {
		return err
	}

	blskey := kmBls.ToString()
	ctx := context.Background()

	has := false
	userBLS12config, err := l.ds.GetKey(ctx, blskey, "local")
	if err == nil && len(userBLS12config) > 0 { //先从本地找，如果有就解析一下
		mkey, err := parseBLS12ConfigMeta([]byte(l.privateKey), userBLS12config)
		if err == nil && mkey != nil {
			l.keySet = mkey
			has = true
		}
	}

	//get from remote
	for _, kid := range l.gInfo.tempKeepers {
		res, err := l.ds.GetKey(ctx, blskey, kid)
		if err == nil && len(res) > 0 {
			mkey, err := parseBLS12ConfigMeta([]byte(l.privateKey), res)
			if err == nil && mkey != nil {
				if l.keySet == nil {
					userBLS12config = res
					l.keySet = mkey
				}
				break
			}
		}
	}

	// get localconfig
	// remote has no config; resend
	if !has && len(userBLS12config) > 0 {
		// store local
		l.gInfo.putToAll(ctx, blskey, userBLS12config)
	}

	if l.keySet != nil {
		utils.MLogger.Info("BlS12 SK and Pk is loaded for ", l.fsID)
		return nil
	}

	return role.ErrEmptyBlsKey
}
