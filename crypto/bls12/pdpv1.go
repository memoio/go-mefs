package mcl

import (
	"math/big"
	"math/rand"
	"strconv"

	"github.com/gogo/protobuf/proto"
	mpb "github.com/memoio/go-mefs/pb"
	"golang.org/x/crypto/blake2s"
)

// 自/data-format/common.go，目前segment的default size为4KB
const (
	AtomCount = 1024
)

// the data structures for the Proof of data possession

// SecretKeyV1 is bls secret key
type SecretKeyV1 struct {
	BlsSk     Fr
	Alpha     Fr
	ElemAlpha []Fr
}

// PublicKeyV1 is bls public key
type PublicKeyV1 struct {
	Count      int
	SignG1     G1   //g_1
	SignG2     G2   //g_2
	BlsPk      G2   //pk = g_2 * sk
	Zeta       G2   //zeta = g_2 * (alpha * sk)
	ElemAlphas []G1 //g_1 * alpha^0,g_1 * alpha^1...g_1 * alpha^count-1
}

// KeySetV1 is wrap
type KeySetV1 struct {
	Pk *PublicKeyV1
	Sk *SecretKeyV1
}

// ChallengeV1 gives
type ChallengeV1 struct {
	R       int64
	Indices []string
}

// ProofV1 is result
type ProofV1 struct {
	Delta []byte `json:"delta"`
	Psi   []byte `json:"psi"`
	Y     []byte `json:"y"`
}

// GenKeySetV1WithSeed create instance
func GenKeySetV1WithSeed(seed []byte, count int) (*KeySetV1, error) {
	sk := &SecretKeyV1{
		ElemAlpha: make([]Fr, count),
	}

	pk := &PublicKeyV1{
		Count:      count,
		ElemAlphas: make([]G1, count),
	}

	ks := &KeySetV1{pk, sk}

	// bls
	// private key
	seed1 := blake2s.Sum256(seed)
	sk.BlsSk.SetLittleEndian(seed1[:])

	seed2 := blake2s.Sum256(seed1[:])
	sk.Alpha.SetLittleEndian(seed2[:])

	seed3 := blake2s.Sum256(seed2[:])
	//g_1
	err := pk.ElemAlphas[0].HashAndMapTo(seed3[:])
	if err != nil {
		return nil, err
	}

	//g_1
	pk.SignG1.HashAndMapTo(seed3[:])

	seed4 := blake2s.Sum256(seed3[:])
	//g_2
	err = pk.SignG2.HashAndMapTo(seed4[:])
	if err != nil {
		return nil, err
	}

	ks.Calculate()

	// return instance
	return ks, nil
}

// GenKeySetV1 create instance
func GenKeySetV1() (*KeySetV1, error) {
	pk := &PublicKeyV1{
		Count:      AtomCount,
		ElemAlphas: make([]G1, AtomCount),
	}
	sk := &SecretKeyV1{
		ElemAlpha: make([]Fr, AtomCount),
	}
	ks := &KeySetV1{pk, sk}

	// bls
	// private key
	sk.BlsSk.SetByCSPRNG()
	sk.Alpha.SetByCSPRNG()

	var seed Fr
	seed.SetByCSPRNG()
	err := pk.ElemAlphas[0].HashAndMapTo(seed.Serialize())
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
func (k *KeySetV1) Calculate() {
	var oneFr Fr
	oneFr.SetInt64(1)
	k.Sk.ElemAlpha[0] = oneFr

	// alpha_i = alpha ^ i
	for i := 1; i < k.Pk.Count; i++ {
		FrMul(&k.Sk.ElemAlpha[i], &k.Sk.ElemAlpha[i-1], &k.Sk.Alpha)
	}

	var tempFr Fr
	FrMul(&tempFr, &k.Sk.Alpha, &k.Sk.BlsSk)

	//zeta = g_2 * (sk*alpha)
	G2Mul(&k.Pk.Zeta, &k.Pk.SignG2, &tempFr)

	//pk = g_2 * sk
	G2Mul(&k.Pk.BlsPk, &k.Pk.SignG2, &k.Sk.BlsSk)

	// U = u^(x^i), i = 0, 1, ..., tagCount-1
	for i := 1; i < k.Pk.Count; i++ {
		G1Mul(&k.Pk.ElemAlphas[i], &k.Pk.ElemAlphas[0], &k.Sk.ElemAlpha[i])
	}

	return
}

// GenTag create tag for *SINGLE* segment
// typ: 32B atom or 24B atom
// mode: sign or not
func (k *KeySetV1) GenTag(index []byte, segment []byte, start, typ int, mode bool) ([]byte, error) {
	if k == nil || k.Pk == nil {
		return nil, ErrKeyIsNil
	}

	var uMiDel G1
	atoms, err := splitSegmentToAtoms(segment, typ)
	if err != nil {
		return nil, err
	}

	// Prod(alpha_j^M_ij)，即Prod(g_1^Sigma(alpha^j*M_ij))
	if k.Sk != nil {
		var power Fr
		if start == 0 {
			FrEvaluatePolynomial(&power, atoms, &(k.Sk.ElemAlpha[1]))
		} else {
			power.Clear() // Set0
			for j, atom := range atoms {
				var mid Fr
				i := j + start
				FrMul(&mid, &(k.Sk.ElemAlpha[i]), &atom) // Xi * Mi
				FrAdd(&power, &power, &mid)              // power = Sigma(Xi*Mi)
			}
		}
		G1Mul(&uMiDel, &(k.Pk.SignG1), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		//FrEvaluatePolynomial
		//FrEvaluatePolynomial
		for j, atom := range atoms {
			// var Mi Fr
			var mid G1
			G1Mul(&mid, &k.Pk.ElemAlphas[j+start], &atom) // uMiDel = ui ^ Mi)
			G1Add(&uMiDel, &uMiDel, &mid)
		}
	}
	if start == 0 {
		// H(Wi)
		var HWi Fr
		h := blake2s.Sum256(index)
		HWi.SetLittleEndian(h[:])
		var HWiG1 G1
		G1Mul(&HWiG1, &k.Pk.SignG1, &HWi)

		// r = HWi * (u^Sgima(Xi*Mi))
		G1Add(&uMiDel, &HWiG1, &uMiDel)
	}

	// sign
	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		G1Mul(&uMiDel, &uMiDel, &(k.Sk.BlsSk))
	}

	return uMiDel.Serialize(), nil
}

// GenChallengeV1 根据时间随机选取的待挑战segments由随机数c对各offset取模而得
func GenChallengeV1(chal *mpb.ChalInfo) int64 {
	newHash, err := blake2s.New256(nil)
	if err != nil {
		return 0
	}
	newHash.Write([]byte(chal.GetPolicy()))
	newHash.Write([]byte(chal.GetQueryID()))
	newHash.Write([]byte(chal.GetUserID()))
	newHash.Write([]byte(chal.GetProviderID()))
	newHash.Write([]byte(chal.GetKeeperID()))
	newHash.Write([]byte(strconv.FormatInt(chal.GetRootTime(), 10)))
	newHash.Write([]byte(strconv.FormatInt(chal.GetChalTime(), 10)))
	newHash.Write([]byte(strconv.FormatInt(chal.GetTotalLength(), 10)))

	for _, bu := range chal.GetBuckets() {
		budata, err := proto.Marshal(bu)
		if err != nil {
			continue
		}
		newHash.Write(budata)
	}

	hashValue := newHash.Sum(chal.GetChunkMap())

	k := new(big.Int).SetBytes(hashValue[:])
	rand.Seed(k.Int64())
	var c int64
	for {
		c = rand.Int63()
		if c != 0 {
			break
		}
	}
	return c
}

// VerifyTag check segment和tag是否对应
func (k *KeySetV1) VerifyTag(index, segment, tag []byte) bool {
	if k == nil || k.Pk == nil {
		return false
	}

	var HWiG1, mido, midt, formula, t G1
	var left, right GT
	var HWi Fr
	formula.Clear()

	//H(W_i) * g_1
	h := blake2s.Sum256(index)
	HWi.SetLittleEndian(h[:])
	G1Mul(&HWiG1, &k.Pk.SignG1, &HWi)

	err := t.Deserialize(tag)
	if err != nil {
		return false
	}

	if t.IsZero() {
		return false
	}

	atoms, err := splitSegmentToAtoms(segment, 32)
	if err != nil {
		return false
	}

	for j, atom := range atoms {
		G1Mul(&mido, &k.Pk.ElemAlphas[j], &atom) // mido = uj ^ mij
		G1Add(&midt, &midt, &mido)               // midt = Prod(uj^mij)
	}

	G1Add(&formula, &HWiG1, &midt) // formula = H(Wi) * Prod(uj^mij)

	Pairing(&left, &t, &(k.Pk.SignG2))       // left = e(tag, g)
	Pairing(&right, &formula, &(k.Pk.BlsPk)) // right = e(H(Wi) * Prod(uj^mij), pk)

	return left.IsEqual(&right)
}

// GenProofV1 gens
func (pk *PublicKeyV1) GenProof(chal ChallengeV1, segments, tags [][]byte, typ int) (*ProofV1, error) {
	if pk == nil || typ <= 0 {
		return nil, ErrKeyIsNil
	}
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	if len(segments) == 0 || len(segments[0]) == 0 {
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
			return nil, err
		}

		for j, atom := range atoms { // 扫描各segment
			FrAdd(&sums[j], &sums[j], &atom)
		}
	}
	if len(pk.ElemAlphas) < tagNum {
		return nil, ErrNumOutOfRange
	}
	var pk_r Fr //P_k(r)
	var fr_r Fr
	fr_r.SetInt64(chal.R)
	FrEvaluatePolynomial(&pk_r, sums, &fr_r)

	// poly(x) - poly(r) always divides (x - r) since the latter is a root of the former.
	// With that infomation we can jump into the division.
	quotient := make([]Fr, len(sums)-1)
	if len(quotient) > 0 {
		// q_(n - 1) = p_n
		quotient[len(quotient)-1] = sums[len(quotient)]
		for j := len(quotient) - 2; j >= 0; j-- {
			// q_j = p_(j + 1) + q_(j + 1) * r
			FrMul(&quotient[j], &quotient[j+1], &fr_r)
			FrAdd(&quotient[j], &quotient[j], &sums[j+1])
		}
	}
	var psi G1
	var tmpG1 G1
	psi.Clear()
	tmpG1.Clear()

	for i := 0; i < len(quotient); i++ {
		G1Mul(&tmpG1, &pk.ElemAlphas[i], &quotient[i])
		G1Add(&psi, &psi, &tmpG1)
	}

	// delta = Prod(tag_i)
	var delta G1
	delta.Clear()
	var t G1
	for _, tag := range tags {
		err := t.Deserialize(tag)
		if err != nil {
			return nil, err
		}
		G1Add(&delta, &delta, &t)
	}

	return &ProofV1{
		Delta: delta.Serialize(),
		Psi:   psi.Serialize(),
		Y:     pk_r.Serialize(),
	}, nil
}

// VerifyProof verify proof
func (vk *PublicKeyV1) VerifyProof(chal ChallengeV1, pf *ProofV1) (bool, error) {
	if vk == nil {
		return false, ErrKeyIsNil
	}
	var psi, delta G1
	var y Fr
	err := delta.Deserialize(pf.Delta)
	if err != nil {
		return false, err
	}

	if delta.IsZero() {
		return false, nil
	}

	err = psi.Deserialize(pf.Psi)
	if err != nil {
		return false, err
	}

	if psi.IsZero() {
		return false, nil
	}

	err = delta.Deserialize(pf.Delta)
	if err != nil {
		return false, err
	}

	if delta.IsZero() {
		return false, nil
	}

	err = y.Deserialize(pf.Y)
	if err != nil {
		return false, err
	}
	var ProdHWi G1
	var tempG2 G2
	var tempFr, HWi Fr
	var lhs1, lhs2, rhs1, rhs2 GT

	ProdHWi.Clear()
	for _, index := range chal.Indices {
		h := blake2s.Sum256([]byte(index))
		tempFr.SetLittleEndian(h[:])
		FrAdd(&HWi, &HWi, &tempFr)
	}
	G1Mul(&ProdHWi, &vk.SignG1, &HWi)
	Pairing(&lhs1, &ProdHWi, &vk.BlsPk)

	tempFr.SetInt64(chal.R)
	FrNeg(&tempFr, &tempFr)
	G2Mul(&tempG2, &vk.BlsPk, &tempFr)
	G2Add(&tempG2, &tempG2, &vk.Zeta)

	Pairing(&lhs2, &psi, &tempG2)
	GTMul(&lhs1, &lhs1, &lhs2)

	Pairing(&rhs1, &delta, &vk.SignG2)

	FrNeg(&tempFr, &y)
	G2Mul(&tempG2, &vk.BlsPk, &tempFr)

	Pairing(&rhs2, &vk.SignG1, &tempG2)
	GTMul(&rhs1, &rhs1, &rhs2)

	return lhs1.IsEqual(&rhs1), nil
}
