package pdp

import (
	"encoding/binary"

	bls "github.com/herumi/bls-eth-go-binary/bls"
)

const MaxPkNumCount = 2048

// the data structures for the Proof of data possession

// SecretKeyV1 is bls secret key
type SecretKeyV1 struct {
	BlsSk     Fr
	Alpha     Fr
	ElemAlpha []Fr
}

func (sk *SecretKeyV1) Version() int {
	return 1
}

func (sk *SecretKeyV1) Serialize() []byte {
	buf := make([]byte, 2*FrSize)
	copy(buf[0:FrSize], sk.BlsSk.Serialize())
	copy(buf[FrSize:2*FrSize], sk.Alpha.Serialize())
	return buf
}

func (sk *SecretKeyV1) Deserialize(data []byte) error {
	if len(data) != 2*FrSize {
		return ErrNumOutOfRange
	}
	err := sk.BlsSk.Deserialize(data[0:FrSize])
	if err != nil {
		return err
	}
	err = sk.Alpha.Deserialize(data[FrSize : 2*FrSize])
	if err != nil {
		return err
	}
	return nil
}

// PublicKeyV1 is bls public key
type PublicKeyV1 struct {
	Count      int64
	BlsPk      G2   //pk = g_2 * sk
	Zeta       G2   //zeta = g_2 * (alpha * sk)
	ElemAlphas []G1 //g_1 * alpha^0,g_1 * alpha^1...g_1 * alpha^count-1
}

func (pk *PublicKeyV1) Version() int {
	return 1
}

func (pk *PublicKeyV1) GetCount() int64 {
	return pk.Count
}

func (pk *PublicKeyV1) VerifyKey() VerifyKey {
	return &VerifyKeyV1{
		BlsPk: pk.BlsPk,
		Zeta:  pk.Zeta,
	}
}

func (pk *PublicKeyV1) Serialize() []byte {
	if pk == nil {
		return nil
	}
	buf := make([]byte, 8+2*G2Size+int(pk.Count)*G1Size)
	binary.BigEndian.PutUint64(buf[:8], uint64(pk.Count))
	copy(buf[8:8+G2Size], pk.BlsPk.Serialize())
	copy(buf[8+G2Size:8+2*G2Size], pk.Zeta.Serialize())
	for i := 0; i < int(pk.Count); i++ {
		copy(buf[8+2*G2Size+i*G1Size:8+2*G2Size+(i+1)*G1Size], pk.ElemAlphas[i].Serialize())
	}
	return buf
}

func (pk *PublicKeyV1) Deserialize(data []byte) error {
	if pk == nil {
		return ErrKeyIsNil
	}
	if len(data) <= 8+2*G2Size {
		return ErrDeserializeFailed
	}
	pk.Count = int64(binary.BigEndian.Uint64(data[:8]))
	if pk.Count > MaxPkNumCount {
		return ErrDeserializeFailed
	}
	pk.BlsPk.Deserialize(data[8 : 8+G2Size])
	pk.Zeta.Deserialize(data[8+G2Size : 8+2*G2Size])
	if (len(data)-(8+2*G2Size))/G1Size != int(pk.Count) {
		return ErrNumOutOfRange
	}
	if int64(len(pk.ElemAlphas)) != pk.Count {
		pk.ElemAlphas = make([]bls.G1, pk.Count)
	}
	for i := 0; i < int(pk.Count); i++ {
		pk.ElemAlphas[i].Deserialize(data[8+2*G2Size+i*G1Size : 8+2*G2Size+(i+1)*G1Size])
	}
	return nil
}

type VerifyKeyV1 struct {
	BlsPk G2 //pk = g_2 * sk
	Zeta  G2 //zeta = g_2 * (alpha * sk)
}

func (vk *VerifyKeyV1) Version() int {
	return 1
}

func (vk *VerifyKeyV1) Serialize() []byte {
	if vk == nil {
		return nil
	}
	buf := make([]byte, 2*G2Size)
	copy(buf[0:G2Size], vk.BlsPk.Serialize())
	copy(buf[G2Size:2*G2Size], vk.Zeta.Serialize())
	return buf
}

func (vk *VerifyKeyV1) Deserialize(data []byte) error {
	if vk == nil {
		return ErrKeyIsNil
	}
	if len(data) != 2*G2Size {
		return ErrNumOutOfRange
	}
	vk.BlsPk.Deserialize(data[:G2Size])
	vk.Zeta.Deserialize(data[G2Size : 2*G2Size])
	return nil
}

// KeySetV1 is wrap
type KeySetV1 struct {
	Pk *PublicKeyV1
	Sk *SecretKeyV1
}

func (k *KeySetV1) PublicKey() PublicKey {
	return k.Pk
}

func (k *KeySetV1) VerifyKey() VerifyKey {
	return &VerifyKeyV1{
		BlsPk: k.Pk.BlsPk,
		Zeta:  k.Pk.Zeta,
	}
}

func (k *KeySetV1) SecreteKey() SecretKey {
	return k.Sk
}

func (k *KeySetV1) Version() int {
	return 1
}
