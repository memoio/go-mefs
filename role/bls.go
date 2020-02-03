package role

import (
	"errors"

	"github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
)

func BLS12KeysetToByte(mkey *mcl.KeySet, privKey []byte) ([]byte, error) {
	pubkey := mkey.Pk
	pubkeyBls := pubkey.BlsPk.Serialize()
	pubkeyG := pubkey.SignG2.Serialize()
	pubkeyU := make([][]byte, mcl.PDPCount)
	pubkeyW := make([][]byte, mcl.PDPCount)

	for i, u := range pubkey.ElemG1s {
		pubkeyU[i] = u.Serialize()
	}

	for i, w := range pubkey.ElemG2s {
		pubkeyW[i] = w.Serialize()
	}

	// 对BLS12的私钥进行加密
	c := btcec.S256()
	_, pubk := btcec.PrivKeyFromBytes(c, privKey)
	secrectKey := mkey.Sk
	blsSK := secrectKey.BlsSk.Serialize()
	blsSKByte, err := btcec.Encrypt(pubk, blsSK)
	if err != nil {
		return nil, err
	}
	x := secrectKey.ElemSk.Serialize()
	XByte, err := btcec.Encrypt(pubk, x)
	if err != nil {
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
		return nil, err
	}

	return userBLS12Config, nil
}

func BLS12ByteToKeyset(userBLS12config []byte, privKey []byte) (*mcl.KeySet, error) {
	if len(userBLS12config) == 0 {
		return nil, errors.New("empty blskey byte")
	}

	mkey := new(mcl.KeySet)

	userBLS12ConfigProto := new(pb.UserBLS12Config)
	err := proto.Unmarshal(userBLS12config, userBLS12ConfigProto)
	if err != nil {
		return mkey, err
	}

	pk := new(mcl.PublicKey)

	err = pk.BlsPk.Deserialize(userBLS12ConfigProto.PubkeyBls)
	if err != nil {
		return mkey, err
	}
	err = pk.SignG2.Deserialize(userBLS12ConfigProto.PubkeyG)
	if err != nil {
		return mkey, err
	}
	pk.ElemG1s = make([]mcl.G1, mcl.PDPCount)
	for i, u := range userBLS12ConfigProto.PubkeyU {
		var temp mcl.G1
		err = temp.Deserialize(u)
		if err != nil {
			utils.MLogger.Info("Deserialize failed: ", err)
		}
		pk.ElemG1s[i] = temp
	}
	pk.ElemG2s = make([]mcl.G2, mcl.PDPCount)
	for i, w := range userBLS12ConfigProto.PubkeyW {
		var temp mcl.G2
		err = temp.Deserialize(w)
		if err != nil {
			utils.MLogger.Info("Deserialize failed: ", err)
		}
		pk.ElemG2s[i] = temp
	}

	mkey.Pk = pk

	if len(privKey) > 0 {
		sk := new(mcl.SecretKey)
		c := btcec.S256()
		seck, _ := btcec.PrivKeyFromBytes(c, privKey)
		if seck == nil {
			return mkey, errors.New("get user's secrete key error")
		}
		blsk, err := btcec.Decrypt(seck, userBLS12ConfigProto.PrikeyBls)
		if err != nil {
			return mkey, err
		}
		err = sk.BlsSk.Deserialize(blsk)
		if err != nil {
			return mkey, err
		}

		x, err := btcec.Decrypt(seck, userBLS12ConfigProto.X)
		if err != nil {
			utils.MLogger.Info("decrypt private key err: ", err)
			return mkey, nil
		}
		err = sk.ElemSk.Deserialize(x)
		if err != nil {
			return mkey, err
		}

		sk.ElemPowerSk = make([]mcl.Fr, mcl.PDPCount)
		err = sk.CalculateXi()
		if err != nil {
			return mkey, err
		}
		mkey.Sk = sk
	}

	return mkey, nil
}
