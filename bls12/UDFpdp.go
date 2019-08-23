package mcl

import (
	"math/rand"
	"crypto/sha512"
	"time"

	big "github.com/ncw/gmp"
)

// 难度值
const T = 10

// the data structures for the proof of data possession(contain UDF)
type (
	PublicKeyUDF struct {
		BlsPK G2
		G     G2
		U     []G1
		W     []G2
		Y     []*big.Int
		// For trf
		NConv *big.Int
	}

	SecretKeyUDF struct {
		BlsSK Fr
		BiX   *big.Int
		BiXI  []*big.Int
		FrXI  []Fr
		// For trf
		phiConv *big.Int
	}

	KeySetUDF struct {
		Pk *PublicKeyUDF
		Sk *SecretKeyUDF
	}
)

// create instance
func GenKeySetUDF() (*KeySetUDF, error) {
	pk := new(PublicKeyUDF)
	sk := new(SecretKeyUDF)

	// bls
	// private key
	sk.BlsSK.SetByCSPRNG()
	rand.Seed(time.Now().UnixNano())
	sk.BiX = new(big.Int).Rand(rand.New(rand.NewSource(time.Now().UnixNano())), phi)
	err := sk.CalculateXi()
	if err != nil {
		return nil, err
	}

	//public key
	var u G1
	var w G2
	// 借助Fr的随机方法为G1和G2产生随机元素
	var seed Fr
	seed.SetByCSPRNG()
	err = u.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}
	err = w.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}

	seed.SetByCSPRNG()
	err = pk.G.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}
	G2Mul(&pk.BlsPK, &pk.G, &sk.BlsSK)
	// 各atom具有不同U
	// U = u^(x^i), i = 0, 1, ..., N-1
	// W = w^(x^i), i = 0, 1, ..., N-1
	pk.U = make([]G1, N) // 须指定大小，否则出现"index out of range"错误
	pk.W = make([]G2, N)
	for i := 0; i < N; i++ {
		G1Mul(&pk.U[i], &u, &sk.FrXI[i])
		G2Mul(&pk.W[i], &w, &sk.FrXI[i])
	}

	y := new(big.Int)
	rand.Seed(time.Now().UnixNano())
	y = y.Rand(rand.New(rand.NewSource(time.Now().UnixNano())), order)
	// y只能为奇数
	if y.Bit(0) == 0 {
		y.SubUint32(y, 1)
	}

	pk.Y = make([]*big.Int, T+numOfAtoms) // TODO：目前T不能大于N
	for i := 0; i < (T + numOfAtoms); i++ {
		pk.Y[i] = new(big.Int).Exp(y, sk.BiXI[i], order) // 用于产生eij
	}

	pk.NConv, sk.phiConv, err = GenParams()
	if err != nil {
		return nil, err
	}

	// return instance
	return &KeySetUDF{pk, sk}, nil
}

// Xi = x^i, i = 0, 1, ..., N
func (sk *SecretKeyUDF) CalculateXi() error {
	var FrX Fr
	ok := FrX.SetHashOf(sk.BiX.Bytes())
	if !ok {
		return ErrSetHashOf
	}
	if len(sk.BiXI) != T && len(sk.FrXI) != N {
		sk.FrXI = make([]Fr, N)
		sk.BiXI = make([]*big.Int, T+numOfAtoms)
	}
	err := sk.FrXI[0].SetString("1", 10)
	if err != nil {
		return ErrSetString
	}
	sk.BiXI[0] = new(big.Int).SetInt64(1)

	for i := 1; i < N; i++ {
		FrMul(&sk.FrXI[i], &sk.FrXI[i-1], &FrX)
	}

	for j := 1; j < (T + numOfAtoms); j++ {
		sk.BiXI[j] = new(big.Int)
		sk.BiXI[j] = sk.BiXI[j].Mul(sk.BiXI[j-1], sk.BiX).Mod(sk.BiXI[j], phi)
	}
	return nil
}

// create tag for *SINGLE* segment
// GenTag + UDF(User-defined Function)
// user应同时将dij发送给provider
func GenTagUDF(keys *KeySetUDF, segment []byte, index []byte) ([]byte, error) {
	var uMiDel, HWi, res G1
	var power, dij, tij Fr
	power.Clear() // Set0

	atoms, err := splitSegmentToAtoms(segment)
	if err != nil {
		return nil, ErrGenTag
	}

	ei := produceEiForEachSegmentFast(keys, index)

	// Prod(u_j^d_ij)，即Prod(u^Sigma(x^j*d_ij))
	for j, atom := range atoms {
		for k, alpha := range ei[j*32 : (j+1)*32] {
			atom[k] = atom[k] ^ alpha // dij = mij + eij，此处的+为异或运算
		}
		judge := dij.SetHashOf(atom)
		if !judge {
			return nil, ErrSetHashOf
		}

		FrMul(&tij, &keys.Sk.FrXI[j], &dij)
		FrAdd(&power, &power, &tij) // power = Sigma(x^j*(mij+eij))
	}
	G1Mul(&uMiDel, &keys.Pk.U[0], &power) // uMiDel = u ^ Sigma(x^j*(mij+eij))

	// H(Wi)
	err = HWi.HashAndMapTo(index)
	if err != nil {
		return nil, err
	}

	G1Add(&res, &HWi, &uMiDel)        // r = HWi * (u^Sigma(x^j*(mij+eij)))
	G1Mul(&res, &res, &keys.Sk.BlsSK) // tag = (HWi * (u^Sigma(x^j*(mij+eij))) ^ blsSK

	return res.Serialize(), nil
}

// For User
func produceEiForEachSegmentFast(keys *KeySetUDF, index []byte) []byte {
	var sigma *big.Int
	H := make([]*big.Int, T)
	eij := new(big.Int)
	ei := make([]byte, numOfAtoms*32)

	// 对各segment有hs = h(index_s), s = 0, 1, …, T-1
	sigma = big.NewInt(0)
	hs := new(big.Int)
	for s := 0; s < T; s++ {
		r := new(big.Int)
		str := string(index) + "_" + string(s)
		H[s] = r.SetBytes([]byte(str)).Mod(r, phi)                              // 直接将index_s由string转化为big.Int
		sigma = sigma.Add(sigma, hs.Mul(keys.Sk.BiXI[s], H[s])).Mod(sigma, phi) // sigma = Sigma((x^s) * hs)
	}

	for j := 0; j < numOfAtoms; j++ {
		// 递归，eij = ei(j-1) ^ x, ei0 = y ^ Sigma((x^s) * hs)
		if j == 0 {
			eij.Exp(keys.Pk.Y[0], sigma, order)
			copy(ei[0:32], eij.Bytes())
		} else {
			eij.Exp(eij, keys.Sk.BiX, order)
			copy(ei[j*32:(j+1)*32], eij.Bytes())
		}
	}

	return ei
}

// For Provider
func produceEiForEachSegmentSlow(key *PublicKeyUDF, index []byte) []byte {
	var prod *big.Int
	H := make([]*big.Int, T)
	eij := new(big.Int)
	ei := make([]byte, numOfAtoms*32)

	// 对各segment有hs = h(index_s), s = 0, 1, …, T-1
	prod = big.NewInt(1)
	for s := 0; s < T; s++ {
		r := new(big.Int)
		str := string(index) + "_" + string(s)
		H[s] = r.SetBytes([]byte(str)).Mod(r, phi) // 直接将index_s由string转化为big.Int
	}

	// 对同个segment上的每个atom,有eij = Prod(Y[j+s]^hs), s = 0, 1, …, T-1
	part := new(big.Int)
	for j := 0; j < numOfAtoms; j++ {
		for s := 0; s < T; s++ {
			eij.Mul(prod, part.Exp(key.Y[j+s], H[s], order)).Mod(eij, order)
		}
		copy(ei[j*32:(j+1)*32], eij.Bytes())
	}

	return ei
}

func produceEiForEachSegmentFastTRF(keys *KeySetUDF, index []byte) []byte {
	temp := big.NewInt(2)
	tempT := big.NewInt(T)
	fastPow := new(big.Int)
	hij := new(big.Int)
	eij := new(big.Int)
	ei := make([]byte, numOfAtoms*32)

	for j := 0; j < numOfAtoms; j++ {
		fastPow.Exp(temp, tempT, keys.Sk.phiConv)   // 知道phi(N)，速度快
		h := sha512.Sum512_224([]byte(string(index) + string(j)))
        hij.SetBytes(h[:])
		eij.Exp(hij, fastPow, keys.Pk.NConv)        // eij = h(index_j)^(2^T)
		copy(ei[j*32:(j+1)*32], eij.Bytes())
	}

	return ei
}

func produceEiForEachSegmentSlowTRF(key *PublicKeyUDF, index []byte) []byte {
	temp := big.NewInt(2)
	tempT := big.NewInt(T)
	slowPow := new(big.Int)
	hij := new(big.Int)
	eij := new(big.Int)
	ei := make([]byte, numOfAtoms*32)

	for j := 0; j < numOfAtoms; j++ {
		slowPow.Exp(temp, tempT, nil)               // phi(N)未知，速度慢
		h := sha512.Sum512_224([]byte(string(index) + string(j)))
        hij.SetBytes(h[:])
		eij.Exp(hij, slowPow, key.NConv)            // eij = h(index_j)^(2^T)
		copy(ei[j*32:(j+1)*32], eij.Bytes())
	}

	return ei
}
