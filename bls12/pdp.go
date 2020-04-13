package mcl

import (
	"crypto/sha256"
	"errors"
	"math/big"
	"math/rand"
	"strconv"

	mpb "github.com/memoio/go-mefs/proto"
	"golang.org/x/crypto/blake2b"
)

// customized errors
var (
	ErrSplitSegmentToAtoms = errors.New("invalid segment")
	ErrKeyIsNil            = errors.New("the key is nil")
	ErrSetHashOf           = errors.New("SetHashOf is not true")
	ErrSetString           = errors.New("SetString is not true")
	ErrSetBigInt           = errors.New("SetBigInt is not true")
	ErrSetToBigInt         = errors.New("SetString (for big.Int) is not true")

	ErrInvalidSettings       = errors.New("setting is invalid")
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
	// (PDPCount - TagAtomNum) * 96 >> len(segment)
	PDPCount    = 1024
	PDPCountV1  = 2048
	AtomSize    = 32
	AtomTagSize = 24 // tag of tags
	TagG1Size   = 48
	TagG2Size   = 96
	// (DefaultSegmentSize-1) / TagAtomSize + 1
	TagAtomNum   = 128
	TagAtomNumV1 = 1024
)

// the data structures for the proof of data possession

// PublicKey is bls public key
type PublicKey struct {
	Count    int
	TagCount int
	BlsPk    G2
	SignG2   G2
	ElemG1s  []G1
	ElemG2s  []G2
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

// GenKeySetWithSeed create instance
func GenKeySetWithSeed(seed []byte, tagCount, count int) (*KeySet, error) {
	// preStored data should large than segmentSize
	if (count-tagCount)*TagG2Size <= tagCount*AtomSize {
		return nil, ErrInvalidSettings
	}

	pk := &PublicKey{
		Count:    count,
		TagCount: tagCount,
		ElemG1s:  make([]G1, tagCount),
		ElemG2s:  make([]G2, count),
	}
	sk := new(SecretKey)
	ks := &KeySet{pk, sk}

	// bls
	// private key
	seed1 := sha256.Sum256(seed)
	sk.BlsSk.SetHashOf(seed1[:])

	seed2 := sha256.Sum256(seed1[:])
	sk.ElemSk.SetHashOf(seed)

	var frSeed Fr
	seed3 := sha256.Sum256(seed2[:])
	frSeed.SetHashOf(seed3[:])
	err := pk.ElemG1s[0].HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	seed4 := sha256.Sum256(seed3[:])
	frSeed.SetHashOf(seed4[:])
	err = pk.ElemG2s[0].HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	seed5 := sha256.Sum256(seed4[:])
	frSeed.SetHashOf(seed5[:])
	err = pk.SignG2.HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	ks.Calculate()

	// return instance
	return ks, nil
}

// GenKeySet create instance
func GenKeySet() (*KeySet, error) {
	pk := &PublicKey{
		Count:    PDPCount,
		TagCount: TagAtomNum,
		ElemG1s:  make([]G1, TagAtomNum),
		ElemG2s:  make([]G2, PDPCount),
	}
	sk := new(SecretKey)
	ks := &KeySet{pk, sk}

	// bls
	// private key
	sk.BlsSk.SetByCSPRNG()
	sk.ElemSk.SetByCSPRNG()

	var seed Fr
	seed.SetByCSPRNG()
	err := pk.ElemG1s[0].HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}

	seed.SetByCSPRNG()
	err = pk.ElemG2s[0].HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}

	seed.SetByCSPRNG()
	err = pk.SignG2.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}

	ks.Calculate()

	// return instance
	return ks, nil
}

// Calculate cals Xi = x^i, Ui and Wi i = 0, 1, ..., N
func (k *KeySet) Calculate() {
	if len(k.Sk.ElemPowerSk) != k.Pk.Count {
		k.Sk.ElemPowerSk = make([]Fr, k.Pk.Count)
	}

	k.Sk.ElemPowerSk[1] = k.Sk.ElemSk

	for i := 2; i < k.Pk.Count; i++ {
		FrMul(&k.Sk.ElemPowerSk[i], &k.Sk.ElemPowerSk[i-1], &k.Sk.ElemSk)
	}

	G2Mul(&k.Pk.BlsPk, &k.Pk.SignG2, &k.Sk.BlsSk)
	// U = u^(x^i), i = 0, 1, ..., tagCount-1
	k.Pk.ElemG1s = make([]G1, k.Pk.TagCount)
	for i := 1; i < k.Pk.TagCount; i++ {
		G1Mul(&k.Pk.ElemG1s[i], &k.Pk.ElemG1s[0], &k.Sk.ElemPowerSk[i])
	}

	// W = w^(x^i), i = 0, 1, ..., count-1
	k.Pk.ElemG2s = make([]G2, k.Pk.Count)
	for i := 1; i < k.Pk.Count; i++ {
		G2Mul(&k.Pk.ElemG2s[i], &k.Pk.ElemG2s[0], &k.Sk.ElemPowerSk[i])
	}
	return
}

// -------------------- proof related routines ------------------- //
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

	if len(atoms)+start > k.Pk.Count {
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

	// sign
	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		G1Mul(&uMiDel, &uMiDel, &(k.Sk.BlsSk))
	}

	return uMiDel.Serialize(), nil
}

// GenChallenge 根据时间随机选取的待挑战segments由随机数c对各offset取模而得
func GenChallenge(chal *mpb.ChalInfo) int {
	newHash, err := blake2b.New256(nil)
	if err != nil {
		return 0
	}
	newHash.Write([]byte(chal.QueryID))
	newHash.Write([]byte(chal.UserID))
	newHash.Write([]byte(chal.ProviderID))
	newHash.Write([]byte(chal.KeeperID))

	newHash.Write([]byte(strconv.FormatInt(chal.ChalTime, 10)))
	newHash.Write([]byte(strconv.FormatInt(chal.TotalLength, 10)))
	newHash.Write([]byte(strconv.FormatInt(chal.BucketNum, 10)))

	for _, stripeNum := range chal.StripeNum {
		newHash.Write([]byte(strconv.FormatInt(stripeNum, 10)))
	}

	hashValue := newHash.Sum(chal.ChunkMap)

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
func (k *KeySet) VerifyTag(index, segment, tag []byte) bool {
	if k == nil || k.Pk == nil {
		return false
	}

	var HWi, mido, midt, formula, t G1
	var left, right GT
	formula.Clear()

	err := HWi.HashAndMapTo(index)
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
	if k == nil || k.Pk == nil || typ <= 0 {
		return nil, ErrKeyIsNil
	}
	var m Fr
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	if len(segments) == 0 {
		return nil, ErrSegmentSize
	}

	tagNum := len(segments[0])
	for _, segment := range segments {
		if len(segment) > tagNum {
			tagNum = len(segment)
		}
	}

	tagNum = tagNum / typ
	sums := make([]Fr, tagNum)
	for _, segment := range segments {
		atoms, err := splitSegmentToAtoms(segment, typ)
		if err != nil {
			return nil, ErrSplitSegmentToAtoms
		}

		for j, atom := range atoms { // 扫描各segment
			judge := m.SetHashOf(atom)
			if !judge {
				return nil, ErrSetHashOf
			}
			FrAdd(&sums[j], &sums[j], &m)
		}
	}

	c := chal.Seed % (k.Pk.Count - k.Pk.TagCount)

	if len(k.Pk.ElemG2s) < tagNum+c || len(k.Pk.ElemG1s) < tagNum {
		return nil, ErrNumOutOfRange
	}
	// 计算h_j = u_(c+j), j = 0, 1, ..., k-1
	// 对于BLS12_381,h_j = w_(c+j)

	h := make([]G2, tagNum)
	for j := 0; j < tagNum; j++ {
		h[j] = k.Pk.ElemG2s[c+j]
	}
	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, tagNum)
	nu := make([]G2, tagNum)
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
func (k *KeySet) VerifyProof(chal Challenge, pf *Proof, mode bool) (bool, error) {
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

	if mode {
		// 第二步：验证mu与nu是对应的
		// lhs = e(mu, h0)
		c := chal.Seed % (k.Pk.Count - k.Pk.TagCount)
		Pairing(&lhs2, &mu, &k.Pk.ElemG2s[c])
		// rhs = e(u, nu)
		Pairing(&rhs2, &k.Pk.ElemG1s[0], &nu)
		// check
		if !lhs2.IsEqual(&rhs2) {
			return false, ErrVerifyStepTwo
		}
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

	if len(segments) == 0 {
		return false, ErrSegmentSize
	}

	tagNum := len(segments[0]) / typ
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	sums := make([]Fr, tagNum)
	for _, segment := range segments {
		atoms, err := splitSegmentToAtoms(segment, typ)
		if err != nil {
			return false, ErrSplitSegmentToAtoms
		}

		for j, atom := range atoms { // 扫描各segment
			if len(atoms) < tagNum {
				return false, ErrNumOutOfRange
			}
			judge := m.SetHashOf(atom)
			if !judge {
				return false, ErrSetHashOf
			}
			FrAdd(&sums[j], &sums[j], &m)
		}
	}

	if len(k.Pk.ElemG1s) < tagNum {
		return false, ErrNumOutOfRange
	}

	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, tagNum)
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
