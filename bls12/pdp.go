package mcl

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"strings"

	mpb "github.com/memoio/go-mefs/proto"
	"golang.org/x/crypto/blake2b"
)

// customized errors
var (
	ErrSplitSegmentToAtoms   = errors.New("invalid segment")
	ErrKeyIsNil              = errors.New("the key is nil")
	ErrSetHashOf             = errors.New("SetHashOf is not true")
	ErrSetString             = errors.New("SetString is not true")
	ErrSetBigInt             = errors.New("SetBigInt is not true")
	ErrSetToBigInt           = errors.New("SetString (for big.Int) is not true")
	ErrNumOutOfRange         = errors.New("numOfAtoms is out of range")
	ErrSegmentSize           = errors.New("the size of the segment is wrong")
	ErrGenTag                = errors.New("GenTag failed")
	ErrOffsetIsNegative      = errors.New("offset is negative")
	ErrProofVerifyInProvider = errors.New("proof is wrong")
	ErrVerifyStepOne         = errors.New("verification failed in Step1")
	ErrVerifyStepTwo         = errors.New("verification failed in Step2")
)

// 自/data-format/common.go，目前segment的default size为4KB
const (
	// (PDPCount - TagAtomNum) * (48 + 96) >> len(segment)
	PDPCount    = 1024
	TagAtomSize = 32
	// (DefaultSegmentSize-1) / TagAtomSize + 1
	TagAtomNum   = 128
	TTagAtomSize = 24
)

// the data structures for the proof of data possession

// PublicKey is bls public key
type PublicKey struct {
	BlsPk   G2
	SignG2  G2
	ElemG1s []G1
	ElemG2s []G2
}

// SecretKey is bls secret key
type SecretKey struct {
	BlsSk       Fr
	ElemSk      Fr
	ElemPowerSk []Fr
}

// KeySet is wrap
type KeySet struct {
	Pk *PublicKey
	Sk *SecretKey
}

// Challenge gives
type Challenge struct {
	Seed    int
	Indices []string
}

// Proof is result
type Proof struct {
	Delta []byte `json:"delta"`
	Mu    []byte `json:"mu"`
	Nu    []byte `json:"nu"`
}

// GenKeySet create instance
func GenKeySet() (*KeySet, error) {
	pk := new(PublicKey)
	sk := new(SecretKey)

	// bls
	// private key
	sk.BlsSk.SetByCSPRNG()
	sk.ElemSk.SetByCSPRNG()
	err := sk.CalculateXi()
	if err != nil {
		fmt.Println(err)
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
	err = pk.SignG2.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}
	G2Mul(&pk.BlsPk, &pk.SignG2, &sk.BlsSk)
	// 各atom具有不同U
	// U = u^(x^i), i = 0, 1, ..., N-1
	// W = w^(x^i), i = 0, 1, ..., N-1
	pk.ElemG1s = make([]G1, PDPCount) // 须指定大小，否则出现"index out of range"错误
	pk.ElemG2s = make([]G2, PDPCount)
	for i := 0; i < PDPCount; i++ {
		G1Mul(&pk.ElemG1s[i], &u, &sk.ElemPowerSk[i])
		G2Mul(&pk.ElemG2s[i], &w, &sk.ElemPowerSk[i])
	}

	// return instance
	return &KeySet{pk, sk}, nil
}

// CalculateXi cals Xi = x^i, i = 0, 1, ..., N
func (s *SecretKey) CalculateXi() error {
	// var mid Fr
	if len(s.ElemPowerSk) != PDPCount {
		s.ElemPowerSk = make([]Fr, PDPCount)
	}
	err := s.ElemPowerSk[0].SetString("1", 10)
	if err != nil {
		return err
	}
	for i := 1; i < PDPCount; i++ {
		FrMul(&s.ElemPowerSk[i], &s.ElemPowerSk[i-1], &s.ElemSk)
	}
	return nil
}

// -------------------- proof related routines ------------------------ //
func splitSegmentToAtoms(data []byte, typ int) ([][]byte, error) {
	if len(data) == 0 {
		return nil, ErrSplitSegmentToAtoms
	}

	if typ > 32 || typ <= 0 {
		return nil, ErrSegmentSize
	}

	num := (len(data)-1)/typ + 1

	atom := make([][]byte, num)

	for i := 0; i < num-1; i++ {
		atom[i] = data[typ*i : typ*(i+1)]
	}

	// last one
	atom[num-1] = data[typ*(num-1):]

	return atom, nil
}

// GenTag create tag for *SINGLE* segment
// typ: 32B atom or 24B atom
// mode: sign or not
func (k *KeySet) GenTag(index []byte, segments []byte, start, typ int, mode bool) ([]byte, error) {
	if k == nil || k.Pk == nil {
		return nil, ErrKeyIsNil
	}

	var uMiDel G1

	atoms, err := splitSegmentToAtoms(segments, typ)
	if err != nil {
		return nil, err
	}

	if len(atoms)+start > PDPCount {
		return nil, ErrGenTag
	}

	// Prod(u_j^M_ij)，即Prod(u^Sigma(x^j*M_ij))
	if k.Sk != nil {
		var power Fr
		power.Clear() // Set0
		for j, atom := range atoms {
			var mid, Mi Fr
			i := j + start
			judge := Mi.SetHashOf(atom)
			if !judge {
				return nil, ErrSetHashOf
			}

			FrMul(&mid, &(k.Sk.ElemPowerSk[i]), &Mi) // Xi * Mi
			FrAdd(&power, &power, &mid)              // power = Sigma(Xi*Mi)
		}

		G1Mul(&uMiDel, &(k.Pk.ElemG1s[0]), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		for j, atom := range atoms {
			var Mi Fr
			var mid G1
			i := j + start
			// Mi为atom而非block或segment
			judge := Mi.SetHashOf(atom)
			if !judge {
				return nil, ErrSetHashOf
			}

			G1Mul(&mid, &(k.Pk.ElemG1s[i]), &Mi) // uMiDel = ui ^ Mi)
			G1Add(&uMiDel, &uMiDel, &mid)
		}
	}

	if start == 0 {
		// H(Wi)
		var HWi G1
		err = HWi.HashAndMapTo(index)
		if err != nil {
			return nil, err
		}
		// r = HWi * (u^Sgima(Xi*Mi))
		G1Add(&uMiDel, &HWi, &uMiDel)
	}

	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		G1Mul(&uMiDel, &uMiDel, &(k.Sk.BlsSk))
	}

	return uMiDel.Serialize(), nil
}

// GenChallenge 根据时间随机选取的待挑战segments由随机数c对各offset取模而得
func GenChallenge(chal *mpb.ChalInfo) int {
	var nb strings.Builder
	nb.WriteString(chal.QueryID)
	nb.WriteString(chal.UserID)
	nb.WriteString(chal.KeeperID)
	nb.WriteString(chal.ProviderID)
	nb.WriteString(strconv.FormatInt(chal.ChalTime, 10))
	nb.WriteString(strconv.FormatInt(chal.TotalLength, 10))
	nb.WriteString(strconv.FormatInt(chal.ChalLength, 10))

	for i := 0; i < len(chal.Blocks); i++ {
		nb.WriteString(chal.Blocks[i])
	}

	hashValue := blake2b.Sum256([]byte(nb.String()))
	k := new(big.Int).SetBytes(hashValue[:])
	rand.Seed(k.Int64())
	var c int
	for {
		c = rand.Int()
		if c != 0 {
			break
		}
	}
	return c
}

// VerifyTag check segment和tag是否对应
func (k *KeySet) VerifyTag(segment, tag []byte, index string) bool {
	if k == nil || k.Pk == nil {
		return false
	}

	var HWi, mido, midt, formula, t G1
	var left, right GT
	formula.Clear()

	err := HWi.HashAndMapTo([]byte(index))
	if err != nil {
		return false
	}

	err = t.Deserialize(tag)
	if err != nil {
		return false
	}

	atoms, err := splitSegmentToAtoms(segment, 32)
	if err != nil {
		return false
	}

	for j, atom := range atoms {
		var Mi Fr
		judge := Mi.SetHashOf(atom)
		if !judge {
			return false
		}
		G1Mul(&mido, &(k.Pk.ElemG1s[j]), &Mi) // mido = uj ^ mij
		G1Add(&midt, &midt, &mido)            // midt = Prod(uj^mij)
	}
	G1Add(&formula, &HWi, &midt) // formula = H(Wi) * Prod(uj^mij)

	Pairing(&left, &t, &(k.Pk.SignG2))       // left = e(tag, g)
	Pairing(&right, &formula, &(k.Pk.BlsPk)) // right = e(H(Wi) * Prod(uj^mij), pk)

	return left.IsEqual(&right)
}

// GenProof gens
func (k *KeySet) GenProof(chal Challenge, segments, tags [][]byte, typ int) (*Proof, error) {
	if k == nil || k.Pk == nil {
		return nil, ErrKeyIsNil
	}
	var m Fr
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	sums := make([]Fr, TagAtomNum)
	for _, segment := range segments {
		if len(segment) > 0 {
			atoms, err := splitSegmentToAtoms(segment, typ)
			if err != nil {
				return nil, ErrSplitSegmentToAtoms
			}

			for j, atom := range atoms { // 扫描各segment
				if len(atoms) < TagAtomNum {
					return nil, ErrNumOutOfRange
				}
				judge := m.SetHashOf(atom)
				if !judge {
					return nil, ErrSetHashOf
				}
				FrAdd(&sums[j], &sums[j], &m)
			}
		}
	}

	c := chal.Seed % (PDPCount - TagAtomNum)

	if len(k.Pk.ElemG2s) < TagAtomNum+c || len(k.Pk.ElemG1s) < TagAtomNum {
		return nil, ErrNumOutOfRange
	}
	// 计算h_j = u_(c+j), j = 0, 1, ..., k-1
	// 对于BLS12_381,h_j = w_(c+j)

	h := make([]G2, TagAtomNum)
	for j := 0; j < TagAtomNum; j++ {
		h[j] = k.Pk.ElemG2s[c+j]
	}
	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, TagAtomNum)
	nu := make([]G2, TagAtomNum)
	var muProd G1
	var nuProd G2
	muProd.Clear()
	nuProd.Clear()
	for j, sum := range sums {
		G1Mul(&mu[j], &(k.Pk.ElemG1s[j]), &sum) // mu_j = U_j ^ sum_j
		G1Add(&muProd, &muProd, &mu[j])         // mu = Prod(U_j^sum_j)

		G2Mul(&nu[j], &h[j], &sum)      // nu_j = h_j ^ sum_j
		G2Add(&nuProd, &nuProd, &nu[j]) // nu = Prod(h_j^sum_j)
	}
	// delta = Prod(tag_i)
	var delta G1
	delta.Clear()
	for _, tag := range tags {
		var t G1
		err := t.Deserialize(tag)
		if err != nil {
			return nil, err
		}
		G1Add(&delta, &delta, &t)
	}

	return &Proof{
		Mu:    muProd.Serialize(),
		Nu:    nuProd.Serialize(),
		Delta: delta.Serialize(),
	}, nil
}

// VerifyProof verify proof
func (k *KeySet) VerifyProof(chal Challenge, pf *Proof) (bool, error) {
	if k == nil || k.Pk == nil {
		return false, ErrKeyIsNil
	}
	var mu, delta G1
	var nu G2
	err := mu.Deserialize(pf.Mu)
	if err != nil {
		return false, err
	}

	err = nu.Deserialize(pf.Nu)
	if err != nil {
		return false, err
	}

	err = delta.Deserialize(pf.Delta)
	if err != nil {
		return false, err
	}

	var ProdHWi, ProdHWimu, HWi G1
	var lhs1, lhs2, rhs1, rhs2 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	Pairing(&lhs1, &delta, &k.Pk.SignG2)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	ProdHWi.Clear()
	for _, index := range chal.Indices {
		err := HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		G1Add(&ProdHWi, &ProdHWi, &HWi)
	}
	G1Add(&ProdHWimu, &ProdHWi, &mu)
	Pairing(&rhs1, &ProdHWimu, &k.Pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	// 第二步：验证mu与nu是对应的
	// lhs = e(mu, h0)
	c := chal.Seed % (PDPCount - TagAtomNum)
	Pairing(&lhs2, &mu, &k.Pk.ElemG2s[c])
	// rhs = e(u, nu)
	Pairing(&rhs2, &k.Pk.ElemG1s[0], &nu)
	// check
	if !lhs2.IsEqual(&rhs2) {
		return false, ErrVerifyStepTwo
	}

	return true, nil
}

// VerifyDataForUser User用于聚合验证数据完整性
func (k *KeySet) VerifyDataForUser(indices []string, segments, tags [][]byte, typ int) (bool, error) {
	if (len(indices) != len(segments)) || (len(indices) != len(tags)) {
		return false, ErrNumOutOfRange
	}
	if k.Pk == nil {
		return false, ErrKeyIsNil
	}
	var m Fr
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	sums := make([]Fr, TagAtomNum)
	for _, segment := range segments {
		if len(segment) > 0 {
			atoms, err := splitSegmentToAtoms(segment, typ)
			if err != nil {
				return false, ErrSplitSegmentToAtoms
			}

			for j, atom := range atoms { // 扫描各segment
				if len(atoms) < TagAtomNum {
					return false, ErrNumOutOfRange
				}
				judge := m.SetHashOf(atom)
				if !judge {
					return false, ErrSetHashOf
				}
				FrAdd(&sums[j], &sums[j], &m)
			}
		}
	}

	if len(k.Pk.ElemG1s) < TagAtomNum {
		return false, ErrNumOutOfRange
	}

	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, TagAtomNum)
	var muProd G1
	muProd.Clear()
	for j, sum := range sums {
		G1Mul(&mu[j], &(k.Pk.ElemG1s[j]), &sum) // mu_j = U_j ^ sum_j
		G1Add(&muProd, &muProd, &mu[j])         // mu = Prod(U_j^sum_j)
	}
	// delta = Prod(tag_i)
	var delta G1
	delta.Clear()
	for _, tag := range tags {
		var t G1
		err := t.Deserialize(tag)
		if err != nil {
			return false, err
		}
		G1Add(&delta, &delta, &t)
	}

	var ProdHWi, ProdHWimu, HWi G1
	var lhs1, rhs1 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	Pairing(&lhs1, &delta, &k.Pk.SignG2)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	ProdHWi.Clear()
	for _, index := range indices {
		err := HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		G1Add(&ProdHWi, &ProdHWi, &HWi)
	}
	G1Add(&ProdHWimu, &ProdHWi, &muProd)
	Pairing(&rhs1, &ProdHWimu, &k.Pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}
