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

func putUserConfig(userID string, keepers []string, sk []byte, keySet *mcl.KeySet) {
	kmBls, err := metainfo.NewKeyMeta(userID, metainfo.Config)
	if err != nil {
		return
	}

	userBLS12Config, err := role.BLS12KeysetToByte(keySet, sk)
	if err != nil {
		log.Println("Marshal BlS12 config failed: ", err)
		return
	}

	// put to local first
	ctx := context.Background()
	err = localNode.Data.PutKey(ctx, kmBls.ToString(), userBLS12Config, "local")
	if err != nil {
		log.Println("CmdPutTo()err")
		return
	}

	//最后发送本节点的BLS12公钥到自己的keeper上保存
	for _, kid := range keepers {
		retry := 0
		for retry < 10 {
			err := localNode.Data.PutKey(ctx, kmBls.ToString(), userBLS12Config, kid)
			if err != nil {
				retry++
				if retry >= 10 {
					log.Println("put config failed :", err)
				}
				time.Sleep(30 * time.Second)
			}
			break
		}
	}

}

func loadBLS12Config(userID string, keepers []string, sk []byte) (keySet *mcl.KeySet, err error) {
	kmBls, err := metainfo.NewKeyMeta(userID, metainfo.Config)
	if err != nil {
		return nil, err
	}

	userBLS12ConfigKey := kmBls.ToString()
	ctx := context.Background()
	userBLS12config, err := localNode.Data.GetKey(ctx, userBLS12ConfigKey, "local")
	if err == nil && len(userBLS12config) > 0 { //先从本地找，如果有就解析一下
		mkey, err := parseBLS12ConfigMeta(sk, userBLS12config)
		if err == nil && keySet != nil {
			keySet = mkey
		}
	}

	//get from remote
	found := false
	bmap := make(map[string]bool)
	for _, kid := range keepers {
		userBLS12config, err = localNode.Data.GetKey(ctx, userBLS12ConfigKey, kid)
		if err == nil && len(userBLS12config) > 0 {
			mkey, err := parseBLS12ConfigMeta(sk, userBLS12config)
			if err == nil && mkey != nil {
				found = true
				if keySet == nil {
					keySet = mkey
				}
				break
			}
			bmap[kid] = true
		}
	}

	// get localconfig
	// remote has no config; resend
	if !found && len(userBLS12config) > 0 {
		// store local
		err = localNode.Data.PutKey(ctx, userBLS12ConfigKey, userBLS12config, "local")
		if err != nil {
			log.Println("put blsconfig to lcoal failed: ", err)
		}

		for kid, has := range bmap {
			if !has {
				continue
			}
			err := localNode.Data.PutKey(ctx, userBLS12ConfigKey, userBLS12config, kid)
			if err != nil {
				log.Println("put blsconfig to keeper", kid, " failed: ", err)
			}
		}
	}

	log.Println("BlS12 SK and Pk is loaded for ", userID)
	return keySet, nil
}

func parseBLS12ConfigMeta(privKey, userBLS12config []byte) (*mcl.KeySet, error) {
	mkey, err := role.BLS12ByteToKeyset(userBLS12config, privKey)
	if err != nil {
		return nil, err
	}

	return mkey, nil
}
