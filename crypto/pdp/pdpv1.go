package pdp

import (
	"math/big"
	"math/rand"
	"strconv"

	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/crypto/bls12"
	mpb "github.com/memoio/go-mefs/pb"
	"golang.org/x/crypto/blake2s"
)

// 自/data-format/common.go，目前segment的default size为4KB
const (
	SCount = 1024
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
	SignG1     G1   //g_1, 考虑将G1，G2做成全局公用的。
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

type ProofAggregatorV1 struct {
	pk    *PublicKeyV1
	r     int64
	typ   int
	sums  []Fr
	delta G1
}

type DataVerifierV1 struct {
	pk    *PublicKeyV1
	sk    *SecretKeyV1
	typ   int
	sums  []Fr
	delta G1
	HWi   Fr
}

func (k *KeySetV1) PublicKey() PublicKey {
	return k.Pk
}

func (k *KeySetV1) SecreteKey() SecretKey {
	return k.Sk
}

func (chal *ChallengeV1) GetSeed() int64 {
	return chal.R
}

func (chal *ChallengeV1) GetIndices() []string {
	return chal.Indices
}

func (pk *PublicKeyV1) GetTagCount() int64 {
	return int64(pk.Count)
}

func (pf *ProofV1) Serialize() []byte {
	buf := make([]byte, 0, len(pf.Delta)+len(pf.Psi)+len(pf.Y))
	buf = append(buf, pf.Delta...)
	buf = append(buf, pf.Psi...)
	buf = append(buf, pf.Y...)

	return buf
}
func (pf *ProofV1) Deserialize(buf []byte) error {
	if len(buf) != 128 {
		return ErrNumOutOfRange
	}
	pf.Delta = make([]byte, 48)
	pf.Psi = make([]byte, 48)
	pf.Y = make([]byte, 32)
	copy(pf.Delta, buf[0:48])
	copy(pf.Psi, buf[48:96])
	copy(pf.Y, buf[96:128])
	return nil
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
		Count:      SCount,
		ElemAlphas: make([]G1, SCount),
	}
	sk := &SecretKeyV1{
		ElemAlpha: make([]Fr, SCount),
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

// How to verify G2?
func (k *PublicKeyV1) Validate() bool {
	var lhs, rhs GT

	if k == nil || len(k.ElemAlphas) == 0 {
		return false
	}
	if len(k.ElemAlphas) != k.Count {
		return false
	}
	if !k.SignG1.IsEqual(&k.ElemAlphas[0]) {
		return false
	}
	for i := 0; i < k.Count-1; i++ {
		mcl.Pairing(&lhs, &k.ElemAlphas[i], &k.Zeta)
		mcl.Pairing(&rhs, &k.ElemAlphas[i], &k.BlsPk)
		if !lhs.IsEqual(&rhs) {
			return false
		}
	}
	return true
}

// Calculate cals Xi = x^i, Ui and Wi i = 0, 1, ..., N
func (k *KeySetV1) Calculate() {
	var oneFr Fr
	oneFr.SetInt64(1)
	k.Sk.ElemAlpha[0] = oneFr

	// alpha_i = alpha ^ i
	for i := 1; i < k.Pk.Count; i++ {
		mcl.FrMul(&k.Sk.ElemAlpha[i], &k.Sk.ElemAlpha[i-1], &k.Sk.Alpha)
	}

	var tempFr Fr
	mcl.FrMul(&tempFr, &k.Sk.Alpha, &k.Sk.BlsSk)

	//zeta = g_2 * (sk*alpha)
	mcl.G2Mul(&k.Pk.Zeta, &k.Pk.SignG2, &tempFr)

	//pk = g_2 * sk
	mcl.G2Mul(&k.Pk.BlsPk, &k.Pk.SignG2, &k.Sk.BlsSk)

	// U = u^(x^i), i = 0, 1, ..., tagCount-1
	for i := 1; i < k.Pk.Count; i++ {
		mcl.G1Mul(&k.Pk.ElemAlphas[i], &k.Pk.ElemAlphas[0], &k.Sk.ElemAlpha[i])
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
			mcl.FrEvaluatePolynomial(&power, atoms, &(k.Sk.ElemAlpha[1]))
		} else {
			power.Clear() // Set0
			for j, atom := range atoms {
				var mid Fr
				i := j + start
				mcl.FrMul(&mid, &(k.Sk.ElemAlpha[i]), &atom) // Xi * Mi
				mcl.FrAdd(&power, &power, &mid)              // power = Sigma(Xi*Mi)
			}
		}
		mcl.G1Mul(&uMiDel, &(k.Pk.SignG1), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		//FrEvaluatePolynomial
		//FrEvaluatePolynomial
		for j, atom := range atoms {
			// var Mi Fr
			var mid G1
			mcl.G1Mul(&mid, &k.Pk.ElemAlphas[j+start], &atom) // uMiDel = ui ^ Mi)
			mcl.G1Add(&uMiDel, &uMiDel, &mid)
		}
	}
	if start == 0 {
		// H(Wi)
		var HWi Fr
		h := blake2s.Sum256(index)
		HWi.SetLittleEndian(h[:])
		var HWiG1 G1
		mcl.G1Mul(&HWiG1, &k.Pk.SignG1, &HWi)

		// r = HWi * (u^Sgima(Xi*Mi))
		mcl.G1Add(&uMiDel, &HWiG1, &uMiDel)
	}

	// sign
	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		mcl.G1Mul(&uMiDel, &uMiDel, &(k.Sk.BlsSk))
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
func (k *PublicKeyV1) VerifyTag(index, segment, tag []byte) bool {
	if k == nil {
		return false
	}

	var HWiG1, mido, midt, formula, t G1
	var left, right GT
	var HWi Fr
	formula.Clear()

	//H(W_i) * g_1
	h := blake2s.Sum256(index)
	HWi.SetLittleEndian(h[:])
	mcl.G1Mul(&HWiG1, &k.SignG1, &HWi)

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
		mcl.G1Mul(&mido, &k.ElemAlphas[j], &atom) // mido = uj ^ mij
		mcl.G1Add(&midt, &midt, &mido)            // midt = Prod(uj^mij)
	}

	mcl.G1Add(&formula, &HWiG1, &midt) // formula = H(Wi) * Prod(uj^mij)

	mcl.Pairing(&left, &t, &(k.SignG2))       // left = e(tag, g)
	mcl.Pairing(&right, &formula, &(k.BlsPk)) // right = e(H(Wi) * Prod(uj^mij), pk)

	return left.IsEqual(&right)
}

// GenProofV1 gens
func (pk *PublicKeyV1) GenProof(chal Challenge, segments, tags [][]byte, typ int) (Proof, error) {
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
			mcl.FrAdd(&sums[j], &sums[j], &atom)
		}
	}
	if len(pk.ElemAlphas) < tagNum {
		return nil, ErrNumOutOfRange
	}
	var pk_r Fr //P_k(r)
	var fr_r Fr
	fr_r.SetInt64(chal.GetSeed())
	mcl.FrEvaluatePolynomial(&pk_r, sums, &fr_r)

	// poly(x) - poly(r) always divides (x - r) since the latter is a root of the former.
	// With that infomation we can jump into the division.
	quotient := make([]Fr, len(sums)-1)
	if len(quotient) > 0 {
		// q_(n - 1) = p_n
		quotient[len(quotient)-1] = sums[len(quotient)]
		for j := len(quotient) - 2; j >= 0; j-- {
			// q_j = p_(j + 1) + q_(j + 1) * r
			mcl.FrMul(&quotient[j], &quotient[j+1], &fr_r)
			mcl.FrAdd(&quotient[j], &quotient[j], &sums[j+1])
		}
	}
	var psi G1
	var tmpG1 G1
	psi.Clear()
	tmpG1.Clear()

	for i := 0; i < len(quotient); i++ {
		mcl.G1Mul(&tmpG1, &pk.ElemAlphas[i], &quotient[i])
		mcl.G1Add(&psi, &psi, &tmpG1)
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
		mcl.G1Add(&delta, &delta, &t)
	}

	return &ProofV1{
			Delta: delta.Serialize(),
			Psi:   psi.Serialize(),
			Y:     pk_r.Serialize(),
		},
		nil
}

// VerifyProof verify proof
func (vk *PublicKeyV1) VerifyProof(chal Challenge, proof Proof, _ bool) (bool, error) {
	if vk == nil {
		return false, ErrKeyIsNil
	}

	pf, ok := proof.(*ProofV1)
	if !ok {
		return false, nil
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
	indices := chal.GetIndices()
	for _, index := range indices {
		h := blake2s.Sum256([]byte(index))
		tempFr.SetLittleEndian(h[:])
		mcl.FrAdd(&HWi, &HWi, &tempFr)
	}
	mcl.G1Mul(&ProdHWi, &vk.SignG1, &HWi)
	mcl.Pairing(&lhs1, &ProdHWi, &vk.BlsPk)

	tempFr.SetInt64(chal.GetSeed())
	mcl.FrNeg(&tempFr, &tempFr)
	mcl.G2Mul(&tempG2, &vk.BlsPk, &tempFr)
	mcl.G2Add(&tempG2, &tempG2, &vk.Zeta)

	mcl.Pairing(&lhs2, &psi, &tempG2)
	mcl.GTMul(&lhs1, &lhs1, &lhs2)

	mcl.Pairing(&rhs1, &delta, &vk.SignG2)

	mcl.FrNeg(&tempFr, &y)
	mcl.G2Mul(&tempG2, &vk.BlsPk, &tempFr)

	mcl.Pairing(&rhs2, &vk.SignG1, &tempG2)
	mcl.GTMul(&rhs1, &rhs1, &rhs2)

	return lhs1.IsEqual(&rhs1), nil
}

// VerifyData User or Provider用于聚合验证数据完整性
func (k *KeySetV1) VerifyData(indices []string, segments, tags [][]byte, typ int) (bool, error) {
	if (len(indices) != len(segments)) || (len(indices) != len(tags)) {
		return false, ErrNumOutOfRange
	}
	if k.Pk == nil {
		return false, ErrKeyIsNil
	}

	if len(segments) == 0 || len(segments[0]) == 0 {
		return false, ErrSegmentSize
	}

	tagNum := len(segments[0]) / typ
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	sums := make([]Fr, tagNum)
	for _, segment := range segments {
		atoms, err := splitSegmentToAtoms(segment, typ)
		if err != nil {
			return false, err
		}

		for j, atom := range atoms { // 扫描各segment
			if len(atoms) < tagNum {
				return false, ErrNumOutOfRange
			}
			mcl.FrAdd(&sums[j], &sums[j], &atom)
		}
	}

	if len(k.Pk.ElemAlphas) < tagNum {
		return false, ErrNumOutOfRange
	}

	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, tagNum)
	var muProd G1
	muProd.Clear()
	if k.Sk != nil {

		var power Fr
		mcl.FrEvaluatePolynomial(&power, sums, &(k.Sk.Alpha))
		mcl.G1Mul(&muProd, &(k.Pk.SignG1), &power)
	} else {
		for j, sum := range sums {
			mcl.G1Mul(&mu[j], &(k.Pk.ElemAlphas[j]), &sum) // mu_j = U_j ^ sum_j
			mcl.G1Add(&muProd, &muProd, &mu[j])            // mu = Prod(U_j^sum_j)
		}
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

		mcl.G1Add(&delta, &delta, &t)
	}

	if delta.IsZero() {
		return false, nil
	}

	var ProdHWi, ProdHWimu G1
	var lhs1, rhs1 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	mcl.Pairing(&lhs1, &delta, &k.Pk.SignG2)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	var tempFr, HWi Fr
	ProdHWi.Clear()
	for _, index := range indices {
		h := blake2s.Sum256([]byte(index))
		tempFr.SetLittleEndian(h[:])
		mcl.FrAdd(&HWi, &HWi, &tempFr)
	}
	mcl.G1Mul(&ProdHWi, &k.Pk.SignG1, &HWi)
	mcl.G1Add(&ProdHWimu, &ProdHWi, &muProd)
	mcl.Pairing(&rhs1, &ProdHWimu, &k.Pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}

func NewProofAggregatorV1(pk *PublicKeyV1, r int64, typ int) ProofAggregatorV1 {
	sums := make([]Fr, pk.Count)
	var delta G1
	delta.Clear()
	return ProofAggregatorV1{pk, r, typ, sums, delta}
}

func (pa *ProofAggregatorV1) Input(segment []byte, tag []byte) error {
	atoms, err := splitSegmentToAtoms(segment, pa.typ)
	if err != nil {
		return err
	}

	for j, atom := range atoms { // 扫描各segment
		mcl.FrAdd(&pa.sums[j], &pa.sums[j], &atom)
	}

	var t G1
	err = t.Deserialize(tag)
	if err != nil {
		return err
	}
	mcl.G1Add(&pa.delta, &pa.delta, &t)
	return nil
}

func (pa *ProofAggregatorV1) InputMulti(segments [][]byte, tags [][]byte) error {
	if len(tags) != len(segments) || len(segments) < 0 {
		return ErrNumOutOfRange
	}
	le := len(segments[0])
	var t G1
	for i, segment := range segments {
		if len(segment) != le {
			return ErrNumOutOfRange
		}
		atoms, err := splitSegmentToAtoms(segment, pa.typ)
		if err != nil {
			return err
		}

		for j, atom := range atoms { // 扫描各segment
			mcl.FrAdd(&pa.sums[j], &pa.sums[j], &atom)
		}
		//tag aggregation
		err = t.Deserialize(tags[i])
		if err != nil {
			return err
		}
		mcl.G1Add(&pa.delta, &pa.delta, &t)
	}

	return nil
}

func (pa *ProofAggregatorV1) Result() (Proof, error) {
	var pk_r Fr //P_k(r)
	var fr_r Fr
	fr_r.SetInt64(pa.r)
	mcl.FrEvaluatePolynomial(&pk_r, pa.sums, &fr_r)

	// poly(x) - poly(r) always divides (x - r) since the latter is a root of the former.
	// With that infomation we can jump into the division.
	quotient := make([]Fr, len(pa.sums)-1)
	if len(quotient) > 0 {
		// q_(n - 1) = p_n
		quotient[len(quotient)-1] = pa.sums[len(quotient)]
		for j := len(quotient) - 2; j >= 0; j-- {
			// q_j = p_(j + 1) + q_(j + 1) * r
			mcl.FrMul(&quotient[j], &quotient[j+1], &fr_r)
			mcl.FrAdd(&quotient[j], &quotient[j], &pa.sums[j+1])
		}
	}
	var psi G1
	var tmpG1 G1
	psi.Clear()
	tmpG1.Clear()

	for i := 0; i < len(quotient); i++ {
		mcl.G1Mul(&tmpG1, &pa.pk.ElemAlphas[i], &quotient[i])
		mcl.G1Add(&psi, &psi, &tmpG1)
	}

	return &ProofV1{
			Delta: pa.delta.Serialize(),
			Psi:   psi.Serialize(),
			Y:     pk_r.Serialize()},
		nil
}

func NewDataVerifierV1(pk *PublicKeyV1, sk *SecretKeyV1, typ int) DataVerifierV1 {
	sums := make([]Fr, pk.Count)
	var delta G1
	delta.Clear()
	var HWi Fr
	HWi.Clear()
	return DataVerifierV1{pk, sk, typ, sums, delta, HWi}
}

func (dv *DataVerifierV1) Input(index []byte, segment []byte, tag []byte) error {
	atoms, err := splitSegmentToAtoms(segment, dv.typ)
	if err != nil {
		return err
	}

	for j, atom := range atoms { // 扫描各segment
		mcl.FrAdd(&dv.sums[j], &dv.sums[j], &atom)
	}

	var t G1
	err = t.Deserialize(tag)
	if err != nil {
		return err
	}
	mcl.G1Add(&dv.delta, &dv.delta, &t)

	var tempFr Fr
	h := blake2s.Sum256([]byte(index))
	tempFr.SetLittleEndian(h[:])
	mcl.FrAdd(&dv.HWi, &dv.HWi, &tempFr)

	return nil
}

func (dv *DataVerifierV1) InputMulti(indices [][]byte, segments [][]byte, tags [][]byte) error {
	if len(tags) != len(segments) || len(indices) != len(segments) || len(segments) < 0 {
		return ErrNumOutOfRange
	}
	le := len(segments[0])
	var t G1
	var tempFr Fr
	for i, segment := range segments {
		if len(segment) != le {
			return ErrNumOutOfRange
		}
		atoms, err := splitSegmentToAtoms(segment, dv.typ)
		if err != nil {
			return err
		}

		for j, atom := range atoms { // 扫描各segment
			mcl.FrAdd(&dv.sums[j], &dv.sums[j], &atom)
		}
		//tag aggregation
		err = t.Deserialize(tags[i])
		if err != nil {
			return err
		}
		mcl.G1Add(&dv.delta, &dv.delta, &t)

		h := blake2s.Sum256([]byte(indices[i]))
		tempFr.SetLittleEndian(h[:])
		mcl.FrAdd(&dv.HWi, &dv.HWi, &tempFr)
	}

	return nil
}

func (dv *DataVerifierV1) Result() (bool, error) {
	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, dv.pk.Count)
	var muProd G1
	muProd.Clear()
	if dv.sk != nil {
		var power Fr
		mcl.FrEvaluatePolynomial(&power, dv.sums, &(dv.sk.Alpha))
		mcl.G1Mul(&muProd, &(dv.pk.SignG1), &power)
	} else {
		for j, sum := range dv.sums {
			mcl.G1Mul(&mu[j], &(dv.pk.ElemAlphas[j]), &sum) // mu_j = U_j ^ sum_j
			mcl.G1Add(&muProd, &muProd, &mu[j])             // mu = Prod(U_j^sum_j)
		}
	}
	if dv.delta.IsZero() {
		return false, nil
	}

	var ProdHWimu G1
	var lhs1, rhs1 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	mcl.Pairing(&lhs1, &dv.delta, &dv.pk.SignG2)
	mcl.G1Mul(&ProdHWimu, &dv.pk.SignG1, &dv.HWi)
	mcl.G1Add(&ProdHWimu, &ProdHWimu, &muProd)
	mcl.Pairing(&rhs1, &ProdHWimu, &dv.pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}
