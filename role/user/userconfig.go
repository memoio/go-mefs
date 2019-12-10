package user

import (
	"log"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (gp *groupService) userBLS12ConfigInit() error {
	log.Printf("Generating BLS12 Sk and Pk for %s: \n", gp.userid)
	kset, err := mcl.GenKeySet()
	if err != nil {
		log.Println("Init BlS12 keyset error: ", err)
		return err
	}
	gp.keySet = kset
	return nil
}

func (gp *groupService) putUserConfig() {
	if gp.keySet == nil {
		log.Println("Need to init or load")
	}
	kmBls, err := metainfo.NewKeyMeta(gp.userid, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return
	}

	userBLS12Config, err := role.BLS12KeysetToByte(gp.keySet, gp.privateKey)
	if err != nil {
		log.Println("Marshal BlS12 config failed: ", err)
		return
	}

	// put to local first
	err = putKeyTo(kmBls.ToString(), string(userBLS12Config), "local")
	if err != nil {
		log.Println("CmdPutTo()err")
		return
	}

	//最后发送本节点的BLS12公钥到自己的keeper上保存
	for _, keeper := range gp.keepers {
		retry := 0
		for retry < 10 {
			err := putKeyTo(kmBls.ToString(), string(userBLS12Config), keeper.keeperID)
			if err != nil {
				retry++
				if retry >= 10 {
					log.Println("gp.localNode.Routing.CmdPut failed :", err)
				}
				time.Sleep(30 * time.Second)
			}
			break
		}
	}

}

func (gp *groupService) loadBLS12Config() error {
	kmBls, err := metainfo.NewKeyMeta(gp.userid, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return err
	}

	userBLS12ConfigKey := kmBls.ToString()
	userBLS12config, err := getKeyFrom(userBLS12ConfigKey, "local")
	if err == nil && len(userBLS12config) > 0 { //先从本地找，如果有就解析一下
		gp.keySet, _ = parseBLS12ConfigMeta(gp.privateKey, userBLS12config)
	}

	//get from remote
	found := false
	for _, keeper := range gp.keepers {
		userBLS12config, err = getKeyFrom(userBLS12ConfigKey, keeper.keeperID)
		if err == nil && len(userBLS12config) > 0 {
			gp.keySet, err = parseBLS12ConfigMeta(gp.privateKey, userBLS12config)
			if err == nil {
				found = true
				break
			}
		}
	}

	// get localconfig
	// remote has no config; resend
	if !found && len(userBLS12config) > 0 {
		// store local
		err = putKeyTo(userBLS12ConfigKey, string(userBLS12config), "local")
		if err != nil {
			log.Println("put blsconfig to lcoal failed: ", err)
		}

		for _, keeper := range gp.keepers {
			err := putKeyTo(userBLS12ConfigKey, string(userBLS12config), keeper.keeperID)
			if err != nil {
				log.Println("put blsconfig to keeper", keeper.keeperID, " failed: ", err)
			}
		}

	}

	log.Println("BlS12 SK and Pk is loaded for ", gp.userid)
	return nil
}

func parseBLS12ConfigMeta(privKey, userBLS12config []byte) (*mcl.KeySet, error) {
	mkey, err := role.BLS12ByteToKeyset(userBLS12config, privKey)
	if err != nil {
		return nil, err
	}

	return mkey, nil
}
