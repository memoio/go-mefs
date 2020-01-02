package user

import (
	"context"
	"log"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func initBLS12Config() (*mcl.KeySet, error) {
	log.Println("Generating BLS12 Sk and Pk")
	kset, err := mcl.GenKeySet()
	if err != nil {
		log.Println("Init BlS12 keyset error: ", err)
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

func (l *LfsInfo) putUserConfig() {
	kmBls, err := metainfo.NewKeyMeta(l.fsID, metainfo.Config, l.userID)
	if err != nil {
		return
	}

	userBLS12Config, err := role.BLS12KeysetToByte(l.keySet, l.privateKey)
	if err != nil {
		log.Println("Marshal BlS12 config failed: ", err)
		return
	}

	blskey := kmBls.ToString()
	ctx := context.Background()
	// put to local first
	err = l.ds.PutKey(ctx, blskey, userBLS12Config, "local")
	if err != nil {
		log.Println("CmdPutTo()err")
		return
	}

	//最后发送本节点的BLS12公钥到自己的keeper上保存
	for _, kid := range l.gInfo.tempKeepers {
		retry := 0
		for retry < 10 {
			err := l.ds.PutKey(ctx, blskey, userBLS12Config, kid)
			if err != nil {
				retry++
				if retry >= 10 {
					log.Println("put config failed :", err)
				}
				time.Sleep(60 * time.Second)
			}
			break
		}
	}
}

func (l *LfsInfo) loadBLS12Config() error {
	kmBls, err := metainfo.NewKeyMeta(l.fsID, metainfo.Config, l.userID)
	if err != nil {
		return err
	}

	blskey := kmBls.ToString()
	ctx := context.Background()

	has := false
	userBLS12config, err := l.ds.GetKey(ctx, blskey, "local")
	if err == nil && len(userBLS12config) > 0 { //先从本地找，如果有就解析一下
		mkey, err := parseBLS12ConfigMeta(l.privateKey, userBLS12config)
		if err == nil && mkey != nil {
			l.keySet = mkey
			has = true
		}
	}

	//get from remote
	for _, kid := range l.gInfo.tempKeepers {
		res, err := l.ds.GetKey(ctx, blskey, kid)
		if err == nil && len(res) > 0 {
			mkey, err := parseBLS12ConfigMeta(l.privateKey, res)
			if err == nil && mkey != nil {
				if l.keySet == nil {
					userBLS12config = res
					l.keySet = mkey
				}
				break
			}
			// send to keeper
			if len(userBLS12config) > 0 {
				err = l.ds.PutKey(ctx, blskey, userBLS12config, kid)
				if err != nil {
					log.Println("put blsconfig to keeper", kid, " failed: ", err)
				}
			}
		}
	}

	// get localconfig
	// remote has no config; resend
	if !has && len(userBLS12config) > 0 {
		// store local
		err = l.ds.PutKey(ctx, blskey, userBLS12config, "local")
		if err != nil {
			log.Println("put blsconfig to lcoal failed: ", err)
		}
	}

	log.Println("BlS12 SK and Pk is loaded for ", l.userID)
	return nil
}
