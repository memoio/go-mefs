package user

import (
	"github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"log"

	"github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	pb "github.com/memoio/go-mefs/role/pb"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (gp *GroupService) userBLS12ConfigInit(password string) ([]byte, error) {
	log.Printf("Generating BLS12 Sk and Pk for %s: \n", gp.Userid)
	var err error
	gp.KeySet, err = mcl.GenKeySet()
	if err != nil {
		gp.KeySet = nil
		log.Println("Init BlS12 keyset error")
		return nil, err
	}
	pubkey := gp.KeySet.Pk
	pubkeyBls := pubkey.BlsPK.Serialize()
	pubkeyG := pubkey.G.Serialize()
	pubkeyU := make([][]byte, mcl.N)
	pubkeyW := make([][]byte, mcl.N)

	for i, u := range pubkey.U {
		if i >= mcl.N {
			break
		}
		pubkeyU[i] = u.Serialize()
	}

	for i, w := range pubkey.W {
		if i >= mcl.N {
			break
		}
		pubkeyW[i] = w.Serialize()
	}

	// 对BLS12的私钥进行加密
	c := btcec.S256()
	_, pubk := btcec.PrivKeyFromBytes(c, gp.PrivateKey)
	secrectKey := gp.KeySet.Sk
	blsSK := secrectKey.BlsSK.Serialize()
	blsSKByte, err := btcec.Encrypt(pubk, blsSK)
	if err != nil {
		gp.KeySet = nil
		return nil, err
	}
	x := secrectKey.X.Serialize()
	XByte, err := btcec.Encrypt(pubk, x)
	if err != nil {
		gp.KeySet = nil
		return nil, err
	}

	userBLS12ConfigProto := &pb.UserBLS12Config{
		PubkeyBls: pubkeyBls,
		PubkeyG:   pubkeyG,
		PubkeyU:   pubkeyU,
		PubkeyW:   pubkeyW,
		PrikeyBls: blsSKByte,
		X:         XByte,
	}

	userBLS12Config, err := proto.Marshal(userBLS12ConfigProto) //将user公私参数通过protobuf序列化以便存储到本地达到持久化的目的
	if err != nil {
		gp.KeySet = nil
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
	userBLS12ConfigProto := new(pb.UserBLS12Config)
	err := proto.Unmarshal(userBLS12config, userBLS12ConfigProto)
	if err != nil {
		return err
	}

	gp.KeySet = new(mcl.KeySet)

	pk := new(mcl.PublicKey)
	sk := new(mcl.SecretKey)

	c := btcec.S256()
	seck, _ := btcec.PrivKeyFromBytes(c, gp.PrivateKey)
	if seck == nil {
		gp.KeySet = nil
		return ErrGetSecreteKey
	}
	blsk, err := btcec.Decrypt(seck, userBLS12ConfigProto.PrikeyBls)
	if err != nil {
		gp.KeySet = nil
		return err
	}
	err = sk.BlsSK.Deserialize(blsk)
	if err != nil {
		gp.KeySet = nil
		return err
	}

	x, err := btcec.Decrypt(seck, userBLS12ConfigProto.X)
	if err != nil {
		gp.KeySet = nil
		return err
	}
	err = sk.X.Deserialize(x)
	if err != nil {
		gp.KeySet = nil
		return err
	}

	sk.XI = make([]mcl.Fr, mcl.N)
	err = sk.CalculateXi()
	if err != nil {
		gp.KeySet = nil
		return err
	}
	err = pk.BlsPK.Deserialize(userBLS12ConfigProto.PubkeyBls)
	if err != nil {
		gp.KeySet = nil
		return err
	}
	err = pk.G.Deserialize(userBLS12ConfigProto.PubkeyG)
	if err != nil {
		gp.KeySet = nil
		return err
	}
	pk.U = make([]mcl.G1, mcl.N)
	for i, u := range userBLS12ConfigProto.PubkeyU {
		if i >= mcl.N {
			break
		}
		var temp mcl.G1
		err = temp.Deserialize(u)
		if err != nil {
			log.Println("temp.Deserialize(u) failed :", err)
		}
		pk.U[i] = temp
	}
	pk.W = make([]mcl.G2, mcl.N)
	for i, w := range userBLS12ConfigProto.PubkeyW {
		if i >= mcl.N {
			break
		}
		var temp mcl.G2
		err = temp.Deserialize(w)
		if err != nil {
			log.Println("temp.Deserialize(u) failed :", err)
		}
		pk.W[i] = temp
	}

	gp.KeySet.Sk = sk
	gp.KeySet.Pk = pk
	return nil
}
