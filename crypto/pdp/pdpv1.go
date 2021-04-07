package pdp

import (
	"math/big"
	"math/rand"
	"strconv"

	"github.com/gogo/protobuf/proto"
	bls "github.com/herumi/bls-eth-go-binary/bls"

	//bls "github.com/herumi/bls-eth-go-binary/bls"
	mpb "github.com/memoio/go-mefs/pb"
	"golang.org/x/crypto/blake2s"
)

// GenKeySetV1WithSeed create instance
func GenKeySetV1WithSeed(seed []byte, count int64) (*KeySetV1, error) {
	sk := &SecretKeyV1{
		ElemAlpha: make([]bls.Fr, count),
	}

	pk := &PublicKeyV1{
		Count:      count,
		ElemAlphas: make([]bls.G1, count),
	}

	ks := &KeySetV1{pk, sk}

	// bls
	// private key
	seed1 := blake2s.Sum256(seed)
	sk.BlsSk.SetLittleEndian(seed1[:])

	seed2 := blake2s.Sum256(seed1[:])
	sk.Alpha.SetLittleEndian(seed2[:])

	//g_1
	pk.ElemAlphas[0] = GenG1

	ks.Calculate()

	// return instance
	return ks, nil
}

// GenKeySetV1 create instance
func GenKeySetV1() (*KeySetV1, error) {
	pk := &PublicKeyV1{
		Count:      SCount,
		ElemAlphas: make([]bls.G1, SCount),
	}
	sk := &SecretKeyV1{
		ElemAlpha: make([]bls.Fr, SCount),
	}
	ks := &KeySetV1{pk, sk}

	// bls
	// private key
	sk.BlsSk.SetByCSPRNG()
	sk.Alpha.SetByCSPRNG()

	//g_1
	pk.ElemAlphas[0] = GenG1

	ks.Calculate()

	// return instance
	return ks, nil
}

// How to verify G2?
func (k *PublicKeyV1) Validate() bool {
	var lhs, rhs bls.GT

	if k == nil || len(k.ElemAlphas) == 0 {
		return false
	}
	if len(k.ElemAlphas) != int(k.Count) {
		return false
	}
	if !GenG1.IsEqual(&k.ElemAlphas[0]) {
		return false
	}
	var i int64
	for i = 0; i < k.Count-1; i++ {
		bls.Pairing(&lhs, &k.ElemAlphas[i], &k.Zeta)
		bls.Pairing(&rhs, &k.ElemAlphas[i], &k.BlsPk)
		if !lhs.IsEqual(&rhs) {
			return false
		}
	}
	return true
}

// Calculate cals Xi = x^i, Ui and Wi i = 0, 1, ..., N
func (k *KeySetV1) Calculate() {
	var oneFr bls.Fr
	oneFr.SetInt64(1)
	k.Sk.ElemAlpha[0] = oneFr

	var i int64
	// alpha_i = alpha ^ i
	for i = 1; i < k.Pk.Count; i++ {
		bls.FrMul(&k.Sk.ElemAlpha[i], &k.Sk.ElemAlpha[i-1], &k.Sk.Alpha)
	}

	var tempFr bls.Fr
	bls.FrMul(&tempFr, &k.Sk.Alpha, &k.Sk.BlsSk)

	//zeta = g_2 * (sk*alpha)
	bls.G2Mul(&k.Pk.Zeta, &GenG2, &tempFr)

	//pk = g_2 * sk
	bls.G2Mul(&k.Pk.BlsPk, &GenG2, &k.Sk.BlsSk)

	// U = u^(x^i), i = 0, 1, ..., tagCount-1
	for i = 1; i < k.Pk.Count; i++ {
		bls.G1Mul(&k.Pk.ElemAlphas[i], &k.Pk.ElemAlphas[0], &k.Sk.ElemAlpha[i])
	}

	return
}

func (sk *SecretKeyV1) Calculate(count int64) {
	if sk == nil {
		return
	}
	var oneFr bls.Fr
	oneFr.SetInt64(1)
	sk.ElemAlpha[0] = oneFr
	if int64(len(sk.ElemAlpha)) != count {
		sk.ElemAlpha = make([]bls.Fr, count)
	}

	var i int64
	// alpha_i = alpha ^ i
	for i = 1; i < count; i++ {
		bls.FrMul(&sk.ElemAlpha[i], &sk.ElemAlpha[i-1], &sk.Alpha)
	}
}

// GenTag create tag for *SINGLE* segment
// typ: 32B atom or 24B atom
// mode: sign or not
func (k *KeySetV1) GenTag(index []byte, segment []byte, start, typ int, mode bool) ([]byte, error) {
	if k == nil || k.Pk == nil {
		return nil, ErrKeyIsNil
	}

	var uMiDel bls.G1
	atoms, err := splitSegmentToAtomsForBLS(segment, typ)
	if err != nil {
		return nil, err
	}

	// Prod(alpha_j^M_ij)，即Prod(g_1^Sigma(alpha^j*M_ij))
	if k.Sk != nil {
		var power bls.Fr
		if start == 0 {
			bls.FrEvaluatePolynomial(&power, atoms, &(k.Sk.ElemAlpha[1]))
		} else {
			power.Clear() // Set0
			for j, atom := range atoms {
				var mid bls.Fr
				i := j + start
				bls.FrMul(&mid, &(k.Sk.ElemAlpha[i]), &atom) // Xi * Mi
				bls.FrAdd(&power, &power, &mid)              // power = Sigma(Xi*Mi)
			}
		}
		bls.G1Mul(&uMiDel, &(GenG1), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		//FrEvaluatePolynomial
		for j, atom := range atoms {
			// var Mi Fr
			var mid bls.G1
			bls.G1Mul(&mid, &k.Pk.ElemAlphas[j+start], &atom) // uMiDel = ui ^ Mi)
			bls.G1Add(&uMiDel, &uMiDel, &mid)
		}
	}
	if start == 0 {
		// H(Wi)
		var HWi bls.Fr
		h := blake2s.Sum256(index)
		HWi.SetLittleEndian(h[:])
		var HWiG1 bls.G1
		bls.G1Mul(&HWiG1, &GenG1, &HWi)

		// r = HWi * (u^Sgima(Xi*Mi))
		bls.G1Add(&uMiDel, &HWiG1, &uMiDel)
	}

	// sign
	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		bls.G1Mul(&uMiDel, &uMiDel, &(k.Sk.BlsSk))
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

	var HWiG1, mido, midt, formula, t bls.G1
	var left, right bls.GT
	var HWi bls.Fr
	formula.Clear()

	//H(W_i) * g_1
	h := blake2s.Sum256(index)
	HWi.SetLittleEndian(h[:])
	bls.G1Mul(&HWiG1, &GenG1, &HWi)

	err := t.Deserialize(tag)
	if err != nil {
		return false
	}

	if t.IsZero() {
		return false
	}

	atoms, err := splitSegmentToAtomsForBLS(segment, 32)
	if err != nil {
		return false
	}

	for j, atom := range atoms {
		bls.G1Mul(&mido, &k.ElemAlphas[j], &atom) // mido = uj ^ mij
		bls.G1Add(&midt, &midt, &mido)            // midt = Prod(uj^mij)
	}

	bls.G1Add(&formula, &HWiG1, &midt) // formula = H(Wi) * Prod(uj^mij)

	bls.Pairing(&left, &t, &(GenG2))          // left = e(tag, g)
	bls.Pairing(&right, &formula, &(k.BlsPk)) // right = e(H(Wi) * Prod(uj^mij), pk)

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
	sums := make([]bls.Fr, tagNum)
	for _, segment := range segments {
		atoms, err := splitSegmentToAtomsForBLS(segment, typ)
		if err != nil {
			return nil, err
		}

		for j, atom := range atoms { // 扫描各segment
			bls.FrAdd(&sums[j], &sums[j], &atom)
		}
	}
	if len(pk.ElemAlphas) < tagNum {
		return nil, ErrNumOutOfRange
	}
	var pk_r bls.Fr //P_k(r)
	var fr_r bls.Fr
	fr_r.SetInt64(chal.GetSeed())
	bls.FrEvaluatePolynomial(&pk_r, sums, &fr_r)

	// poly(x) - poly(r) always divides (x - r) since the latter is a root of the former.
	// With that infomation we can jump into the division.
	quotient := make([]bls.Fr, len(sums)-1)
	if len(quotient) > 0 {
		// q_(n - 1) = p_n
		quotient[len(quotient)-1] = sums[len(quotient)]
		for j := len(quotient) - 2; j >= 0; j-- {
			// q_j = p_(j + 1) + q_(j + 1) * r
			bls.FrMul(&quotient[j], &quotient[j+1], &fr_r)
			bls.FrAdd(&quotient[j], &quotient[j], &sums[j+1])
		}
	}
	var psi bls.G1
	var tmpG1 bls.G1
	psi.Clear()
	tmpG1.Clear()

	for i := 0; i < len(quotient); i++ {
		bls.G1Mul(&tmpG1, &pk.ElemAlphas[i], &quotient[i])
		bls.G1Add(&psi, &psi, &tmpG1)
	}

	// delta = Prod(tag_i)
	var delta bls.G1
	delta.Clear()
	var t bls.G1
	for _, tag := range tags {
		err := t.Deserialize(tag)
		if err != nil {
			return nil, err
		}
		bls.G1Add(&delta, &delta, &t)
	}

	return &ProofV1{
			Delta: delta.Serialize(),
			Psi:   psi.Serialize(),
			Y:     pk_r.Serialize(),
		},
		nil
}

// VerifyProof verify proof
func (vk *VerifyKeyV1) VerifyProof(chal Challenge, proof Proof) (bool, error) {
	if vk == nil {
		return false, ErrKeyIsNil
	}

	pf, ok := proof.(*ProofV1)
	if !ok {
		return false, nil
	}
	var psi, delta bls.G1
	var y bls.Fr
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
	var ProdHWi bls.G1
	var tempG2 bls.G2
	var tempFr, HWi bls.Fr
	var lhs1, lhs2, rhs1, rhs2 bls.GT

	ProdHWi.Clear()
	indices := chal.GetIndices()
	for _, index := range indices {
		h := blake2s.Sum256([]byte(index))
		tempFr.SetLittleEndian(h[:])
		bls.FrAdd(&HWi, &HWi, &tempFr)
	}
	bls.G1Mul(&ProdHWi, &GenG1, &HWi)
	bls.Pairing(&lhs1, &ProdHWi, &vk.BlsPk)

	tempFr.SetInt64(chal.GetSeed())
	bls.FrNeg(&tempFr, &tempFr)
	bls.G2Mul(&tempG2, &vk.BlsPk, &tempFr)
	bls.G2Add(&tempG2, &tempG2, &vk.Zeta)

	bls.Pairing(&lhs2, &psi, &tempG2)
	bls.GTMul(&lhs1, &lhs1, &lhs2)

	bls.Pairing(&rhs1, &delta, &GenG2)

	bls.FrNeg(&tempFr, &y)
	bls.G2Mul(&tempG2, &vk.BlsPk, &tempFr)

	bls.Pairing(&rhs2, &GenG1, &tempG2)
	bls.GTMul(&rhs1, &rhs1, &rhs2)

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
	sums := make([]bls.Fr, tagNum)
	for _, segment := range segments {
		atoms, err := splitSegmentToAtomsForBLS(segment, typ)
		if err != nil {
			return false, err
		}

		for j, atom := range atoms { // 扫描各segment
			if len(atoms) < tagNum {
				return false, ErrNumOutOfRange
			}
			bls.FrAdd(&sums[j], &sums[j], &atom)
		}
	}

	if len(k.Pk.ElemAlphas) < tagNum {
		return false, ErrNumOutOfRange
	}

	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]bls.G1, tagNum)
	var muProd bls.G1
	muProd.Clear()
	if k.Sk != nil {

		var power bls.Fr
		bls.FrEvaluatePolynomial(&power, sums, &(k.Sk.Alpha))
		bls.G1Mul(&muProd, &(GenG1), &power)
	} else {
		for j, sum := range sums {
			bls.G1Mul(&mu[j], &(k.Pk.ElemAlphas[j]), &sum) // mu_j = U_j ^ sum_j
			bls.G1Add(&muProd, &muProd, &mu[j])            // mu = Prod(U_j^sum_j)
		}
	}
	// delta = Prod(tag_i)
	var delta bls.G1
	delta.Clear()
	for _, tag := range tags {
		var t bls.G1
		err := t.Deserialize(tag)
		if err != nil {
			return false, err
		}

		bls.G1Add(&delta, &delta, &t)
	}

	if delta.IsZero() {
		return false, nil
	}

	var ProdHWi, ProdHWimu bls.G1
	var lhs1, rhs1 bls.GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	bls.Pairing(&lhs1, &delta, &GenG2)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	var tempFr, HWi bls.Fr
	ProdHWi.Clear()
	for _, index := range indices {
		h := blake2s.Sum256([]byte(index))
		tempFr.SetLittleEndian(h[:])
		bls.FrAdd(&HWi, &HWi, &tempFr)
	}
	bls.G1Mul(&ProdHWi, &GenG1, &HWi)
	bls.G1Add(&ProdHWimu, &ProdHWi, &muProd)
	bls.Pairing(&rhs1, &ProdHWimu, &k.Pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}

type ProofAggregatorV1 struct {
	pk    *PublicKeyV1
	r     int64
	typ   int
	sums  []bls.Fr
	delta bls.G1
}

func NewProofAggregatorV1(pk *PublicKeyV1, r int64, typ int) ProofAggregatorV1 {
	sums := make([]bls.Fr, pk.Count)
	var delta bls.G1
	delta.Clear()
	return ProofAggregatorV1{pk, r, typ, sums, delta}
}

func (pa *ProofAggregatorV1) Input(segment []byte, tag []byte) error {
	atoms, err := splitSegmentToAtomsForBLS(segment, pa.typ)
	if err != nil {
		return err
	}

	for j, atom := range atoms { // 扫描各segment
		bls.FrAdd(&pa.sums[j], &pa.sums[j], &atom)
	}

	var t bls.G1
	err = t.Deserialize(tag)
	if err != nil {
		return err
	}
	bls.G1Add(&pa.delta, &pa.delta, &t)
	return nil
}

func (pa *ProofAggregatorV1) InputMulti(segments [][]byte, tags [][]byte) error {
	if len(tags) != len(segments) || len(segments) < 0 {
		return ErrNumOutOfRange
	}
	le := len(segments[0])
	var t bls.G1
	for i, segment := range segments {
		if len(segment) != le {
			return ErrNumOutOfRange
		}
		atoms, err := splitSegmentToAtomsForBLS(segment, pa.typ)
		if err != nil {
			return err
		}

		for j, atom := range atoms { // 扫描各segment
			bls.FrAdd(&pa.sums[j], &pa.sums[j], &atom)
		}
		//tag aggregation
		err = t.Deserialize(tags[i])
		if err != nil {
			return err
		}
		bls.G1Add(&pa.delta, &pa.delta, &t)
	}

	return nil
}

func (pa *ProofAggregatorV1) Result() (Proof, error) {
	var pk_r bls.Fr //P_k(r)
	var fr_r bls.Fr
	fr_r.SetInt64(pa.r)
	bls.FrEvaluatePolynomial(&pk_r, pa.sums, &fr_r)

	// poly(x) - poly(r) always divides (x - r) since the latter is a root of the former.
	// With that infomation we can jump into the division.
	quotient := make([]bls.Fr, len(pa.sums)-1)
	if len(quotient) > 0 {
		// q_(n - 1) = p_n
		quotient[len(quotient)-1] = pa.sums[len(quotient)]
		for j := len(quotient) - 2; j >= 0; j-- {
			// q_j = p_(j + 1) + q_(j + 1) * r
			bls.FrMul(&quotient[j], &quotient[j+1], &fr_r)
			bls.FrAdd(&quotient[j], &quotient[j], &pa.sums[j+1])
		}
	}
	var psi bls.G1
	var tmpG1 bls.G1
	psi.Clear()
	tmpG1.Clear()

	for i := 0; i < len(quotient); i++ {
		bls.G1Mul(&tmpG1, &pa.pk.ElemAlphas[i], &quotient[i])
		bls.G1Add(&psi, &psi, &tmpG1)
	}

	return &ProofV1{
			Delta: pa.delta.Serialize(),
			Psi:   psi.Serialize(),
			Y:     pk_r.Serialize()},
		nil
}

type DataVerifierV1 struct {
	pk    *PublicKeyV1
	sk    *SecretKeyV1
	typ   int
	sums  []bls.Fr
	delta bls.G1
	HWi   bls.Fr
}

func NewDataVerifierV1(pk *PublicKeyV1, sk *SecretKeyV1, typ int) DataVerifierV1 {
	sums := make([]bls.Fr, pk.Count)
	var delta bls.G1
	delta.Clear()
	var HWi bls.Fr
	HWi.Clear()
	return DataVerifierV1{pk, sk, typ, sums, delta, HWi}
}

func (dv *DataVerifierV1) Input(index []byte, segment []byte, tag []byte) error {
	atoms, err := splitSegmentToAtomsForBLS(segment, dv.typ)
	if err != nil {
		return err
	}

	for j, atom := range atoms { // 扫描各segment
		bls.FrAdd(&dv.sums[j], &dv.sums[j], &atom)
	}

	var t bls.G1
	err = t.Deserialize(tag)
	if err != nil {
		return err
	}
	bls.G1Add(&dv.delta, &dv.delta, &t)

	var tempFr bls.Fr
	h := blake2s.Sum256([]byte(index))
	tempFr.SetLittleEndian(h[:])
	bls.FrAdd(&dv.HWi, &dv.HWi, &tempFr)

	return nil
}

func (dv *DataVerifierV1) InputMulti(indices [][]byte, segments [][]byte, tags [][]byte) error {
	if len(tags) != len(segments) || len(indices) != len(segments) || len(segments) < 0 {
		return ErrNumOutOfRange
	}
	le := len(segments[0])
	var t bls.G1
	var tempFr bls.Fr
	for i, segment := range segments {
		if len(segment) != le {
			return ErrNumOutOfRange
		}
		atoms, err := splitSegmentToAtomsForBLS(segment, dv.typ)
		if err != nil {
			return err
		}

		for j, atom := range atoms { // 扫描各segment
			bls.FrAdd(&dv.sums[j], &dv.sums[j], &atom)
		}
		//tag aggregation
		err = t.Deserialize(tags[i])
		if err != nil {
			return err
		}
		bls.G1Add(&dv.delta, &dv.delta, &t)

		h := blake2s.Sum256([]byte(indices[i]))
		tempFr.SetLittleEndian(h[:])
		bls.FrAdd(&dv.HWi, &dv.HWi, &tempFr)
	}

	return nil
}

func (dv *DataVerifierV1) Result() (bool, error) {
	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]bls.G1, dv.pk.Count)
	var muProd bls.G1
	muProd.Clear()
	if dv.sk != nil {
		var power bls.Fr
		bls.FrEvaluatePolynomial(&power, dv.sums, &(dv.sk.Alpha))
		bls.G1Mul(&muProd, &(GenG1), &power)
	} else {
		for j, sum := range dv.sums {
			bls.G1Mul(&mu[j], &(dv.pk.ElemAlphas[j]), &sum) // mu_j = U_j ^ sum_j
			bls.G1Add(&muProd, &muProd, &mu[j])             // mu = Prod(U_j^sum_j)
		}
	}
	if dv.delta.IsZero() {
		return false, nil
	}

	var ProdHWimu bls.G1
	var lhs1, rhs1 bls.GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	bls.Pairing(&lhs1, &dv.delta, &GenG2)
	bls.G1Mul(&ProdHWimu, &GenG1, &dv.HWi)
	bls.G1Add(&ProdHWimu, &ProdHWimu, &muProd)
	bls.Pairing(&rhs1, &ProdHWimu, &dv.pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}
