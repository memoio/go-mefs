package role

import (
	"github.com/btcsuite/btcd/btcec"
	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/pdp"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/utils"
)

func BLS12KeysetToByte(mkey pdp.KeySet, privKey []byte) ([]byte, error) {
	if mkey == nil || mkey.PublicKey() == nil || mkey.SecreteKey() == nil {
		return nil, ErrEmptyBlsKey
	}

	pubkey := mkey.PublicKey().Serialize()

	// 对BLS12的私钥进行加密
	c := btcec.S256()
	_, pubk := btcec.PrivKeyFromBytes(c, privKey)
	secreteKey := mkey.SecreteKey().Serialize()
	blsSKByte, err := btcec.Encrypt(pubk, secreteKey)
	if err != nil {
		return nil, err
	}

	BLSKeyProto := &mpb.BLSKey{
		Version:   1,
		PubKey:    pubkey,
		SecretKey: blsSKByte,
	}

	BLS12Config, err := proto.Marshal(BLSKeyProto) //将user公私参数通过protobuf序列化以便存储到本地达到持久化的目的
	if err != nil {
		return nil, err
	}

	return BLS12Config, nil
}

func BLS12ByteToKeyset(BLS12config []byte, privKey []byte) (pdp.KeySet, error) {
	if len(BLS12config) == 0 {
		return nil, ErrEmptyBlsKey
	}

	BLSKey := new(mpb.BLSKey)
	err := proto.Unmarshal(BLS12config, BLSKey)
	if err != nil {
		return nil, err
	}
	switch BLSKey.Version {
	case 1:
		return deserializeKeyV1(BLSKey, privKey)
	default:
		return nil, ErrEmptyBlsKey
	}
}

func deserializeKeyV1(BLSKey *mpb.BLSKey, privKey []byte) (pdp.KeySet, error) {
	mkey := new(pdp.KeySetV1)
	pk := new(pdp.PublicKeyV1)
	err := pk.Deserialize(BLSKey.PubKey)
	if err != nil {
		return mkey, err
	}

	mkey.Pk = pk

	if len(privKey) > 0 {
		sk := &pdp.SecretKeyV1{
			ElemAlpha: make([]pdp.Fr, pk.Count),
		}
		c := btcec.S256()
		seck, _ := btcec.PrivKeyFromBytes(c, privKey)
		if seck == nil {
			utils.MLogger.Info("calculate private key fails")
			return mkey, nil
		}

		blsk, err := btcec.Decrypt(seck, BLSKey.SecretKey)
		if err != nil {
			utils.MLogger.Info("decrypt private bls key err: ", err)
			return mkey, nil
		}
		err = sk.Deserialize(blsk)
		if err != nil {
			return mkey, err
		}

		sk.Calculate(pk.Count)
		mkey.Sk = sk

		mkey.Calculate()
	}

	return mkey, nil
}
