package user

import (
	"log"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (gp *GroupService) userBLS12ConfigInit() ([]byte, error) {
	log.Printf("Generating BLS12 Sk and Pk for %s: \n", gp.Userid)
	var err error
	gp.KeySet, err = mcl.GenKeySet()
	if err != nil {
		gp.KeySet = nil
		log.Println("Init BlS12 keyset error: ", err)
		return nil, err
	}

	userBLS12Config, err := role.BLS12KeysetToByte(gp.KeySet, gp.PrivateKey)
	if err != nil {
		gp.KeySet = nil
		log.Println("Marshal BlS12 config failed: ", err)
		return nil, err
	}
	log.Println(gp.Userid, "'s BlS12 SK and PK init success")
	return userBLS12Config, nil
}

func (gp *GroupService) loadBLS12Config() error {
	log.Printf("Loading BLS12 Sk and Pk for %s: \n", gp.Userid)
	var userBLS12config []byte
	var err error
	kmBls, err := metainfo.NewKeyMeta(gp.Userid, metainfo.Local, metainfo.SyncTypeCfg, metainfo.CfgTypeBls12)
	if err != nil {
		return err
	}

	found := false

	userBLS12ConfigKey := kmBls.ToString()
	userBLS12config, err = localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userBLS12ConfigKey, "local")
	if err == nil && len(userBLS12config) > 0 { //先从本地找，如果有就解析一下
		err = gp.parseBLS12ConfigMeta(userBLS12config)
		if err == nil {
			found = true
		}
	}

	//本地没有，然后去找Keeper要
	if !found && len(gp.localPeersInfo.Keepers) > 0 {
		for _, keeper := range gp.localPeersInfo.Keepers {
			userBLS12config, err = localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(userBLS12ConfigKey, keeper.KeeperID)
			if err == nil && len(userBLS12config) > 0 {
				err = gp.parseBLS12ConfigMeta(userBLS12config)
				if err == nil {
					found = true
					break
				}
			}
		}
	}

	// get localconfig
	if found && len(userBLS12config) > 0 {
		// store local
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(userBLS12ConfigKey, string(userBLS12config), "local")
		if err != nil {
			log.Println("put blsconfig to lcoal failed: ", err)
		}

		if len(gp.localPeersInfo.Keepers) > 0 {
			for _, keeper := range gp.localPeersInfo.Keepers {
				err := localNode.Routing.(*dht.IpfsDHT).CmdPutTo(userBLS12ConfigKey, string(userBLS12config), keeper.KeeperID)
				if err != nil {
					log.Println("put blsconfig to keeper", keeper.KeeperID, " failed: ", err)
				}
			}
		}
	}

	log.Println("BlS12 SK and Pk is loaded for ", gp.Userid)
	return nil
}

func (gp *GroupService) parseBLS12ConfigMeta(userBLS12config []byte) error {
	mkey, err := role.BLS12ByteToKeyset(userBLS12config, gp.PrivateKey)
	if err != nil {
		return err
	}
	gp.KeySet = mkey
	return nil
}
