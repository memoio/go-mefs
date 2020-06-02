package role

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/crypto/bls12"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/utils"
)

func BLS12KeysetToByte(mkey *mcl.KeySet, privKey []byte) ([]byte, error) {
	if mkey == nil || mkey.Pk == nil || mkey.Sk == nil {
		return nil, ErrEmptyBlsKey
	}

	pubkey := mkey.Pk
	pubkeyBls := pubkey.BlsPk.Serialize()
	pubkeyG := pubkey.SignG2.Serialize()
	pubkeyU := make([][]byte, mkey.Pk.TagCount)
	pubkeyW := make([][]byte, mkey.Pk.Count)

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

	userBLS12ConfigProto := &mpb.UserBLS12Config{
		PubkeyBls: pubkeyBls,
		PubkeyG:   pubkeyG,
		PubkeyU:   pubkeyU,
		PubkeyW:   pubkeyW,
		PrikeyBls: blsSKByte,
		X:         XByte,
		Count:     int32(mkey.Pk.Count),
		TagCount:  int32(mkey.Pk.TagCount),
	}

	userBLS12Config, err := proto.Marshal(userBLS12ConfigProto) //将user公私参数通过protobuf序列化以便存储到本地达到持久化的目的
	if err != nil {
		return nil, err
	}

	return userBLS12Config, nil
}

func BLS12ByteToKeyset(userBLS12config []byte, privKey []byte) (*mcl.KeySet, error) {
	if len(userBLS12config) == 0 {
		return nil, ErrEmptyBlsKey
	}

	mkey := new(mcl.KeySet)

	userBLS12ConfigProto := new(mpb.UserBLS12Config)
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

	if userBLS12ConfigProto.GetTagCount() == 0 {
		pk.TagCount = mcl.TagAtomNum
	} else {
		pk.TagCount = int(userBLS12ConfigProto.GetTagCount())
	}

	pk.TagCount = len(userBLS12ConfigProto.PubkeyU)

	pk.ElemG1s = make([]mcl.G1, pk.TagCount)
	for i, u := range userBLS12ConfigProto.PubkeyU {
		var temp mcl.G1
		err = temp.Deserialize(u)
		if err != nil {
			utils.MLogger.Info("Deserialize failed: ", err)
		}
		pk.ElemG1s[i] = temp
	}

	// version is user for different default
	if userBLS12ConfigProto.GetCount() == 0 {
		pk.Count = mcl.PDPCount
	} else {
		pk.Count = int(userBLS12ConfigProto.GetCount())
	}

	pk.Count = len(userBLS12ConfigProto.PubkeyW)

	pk.ElemG2s = make([]mcl.G2, pk.Count)
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
			utils.MLogger.Info("calculate private key fails")
			return mkey, nil
		}

		blsk, err := btcec.Decrypt(seck, userBLS12ConfigProto.PrikeyBls)
		if err != nil {
			utils.MLogger.Info("decrypt private bls key err: ", err)
			return mkey, nil
		}
		err = sk.BlsSk.Deserialize(blsk)
		if err != nil {
			return mkey, err
		}

		x, err := btcec.Decrypt(seck, userBLS12ConfigProto.X)
		if err != nil {
			utils.MLogger.Info("decrypt private bls x err: ", err)
			return mkey, nil
		}
		err = sk.ElemSk.Deserialize(x)
		if err != nil {
			return mkey, err
		}

		mkey.Sk = sk

		mkey.Calculate()
	}

	return mkey, nil
}
