package pdp

import (
	"crypto/sha256"
	"encoding/binary"
	"math/big"
	"math/rand"
	"strconv"

	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/crypto/bls12"
	mpb "github.com/memoio/go-mefs/pb"
	"golang.org/x/crypto/blake2b"
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

// SecretKeyV0 is bls secret key
type SecretKeyV0 struct {
	BlsSk       Fr
	ElemSk      Fr
	ElemPowerSk []Fr
}

// PublicKeyV0 is bls public key
type PublicKeyV0 struct {
	Count    int64
	TagCount int64
	BlsPk    G2
	SignG2   G2
	ElemG1s  []G1
	ElemG2s  []G2
}

// KeySetV0 is wrap
type KeySetV0 struct {
	Pk *PublicKeyV0
	Sk *SecretKeyV0
}

// ChallengeV0 gives
type ChallengeV0 struct {
	Seed    int64
	Indices []string
}

// ProofV0 is result
type ProofV0 struct {
	Delta []byte `json:"delta"` //G1 48Byte
	Mu    []byte `json:"mu"`    //G1 48Byte
	Nu    []byte `json:"nu"`    //G2 96Byte
}

type ProofAggregatorV0 struct {
	pk    *PublicKeyV0
	seed  int64
	typ   int
	sums  []Fr
	delta G1
}

type DataVerifierV0 struct {
	pk      *PublicKeyV0
	sk      *SecretKeyV0
	typ     int
	sums    []Fr
	delta   G1
	prodHWi G1
}

func (k *KeySetV0) PublicKey() PublicKey {
	return k.Pk
}

func (k *KeySetV0) SecreteKey() SecretKey {
	return k.Sk
}

func (pk *PublicKeyV0) GetTagCount() int64 {
	return pk.TagCount
}

func (chal *ChallengeV0) GetSeed() int64 {
	return chal.Seed
}

func (chal *ChallengeV0) GetIndices() []string {
	return chal.Indices
}

func (pf *ProofV0) Serialize() []byte {
	buf := make([]byte, 0, 192)
	buf = append(buf, pf.Delta...)
	buf = append(buf, pf.Mu...)
	buf = append(buf, pf.Nu...)
	return buf
}

func (pf *ProofV0) Deserialize(buf []byte) error {
	if len(buf) != 192 {
		return ErrNumOutOfRange
	}
	pf.Delta = make([]byte, 48)
	pf.Mu = make([]byte, 48)
	pf.Nu = make([]byte, 96)
	copy(pf.Delta, buf[0:48])
	copy(pf.Mu, buf[48:96])
	copy(pf.Nu, buf[96:192])
	return nil
}

// GenKeySetV0WithSeed create instance
func GenKeySetV0WithSeed(seed []byte, tagCount, count int64) (*KeySetV0, error) {
	// preStored data should large than segmentSize
	if (count-tagCount)*TagG2Size <= tagCount*AtomSize {
		return nil, ErrInvalidSettings
	}

	pk := &PublicKeyV0{
		Count:    count,
		TagCount: tagCount,
		ElemG1s:  make([]G1, tagCount),
		ElemG2s:  make([]G2, count),
	}
	sk := &SecretKeyV0{
		ElemPowerSk: make([]Fr, count),
	}
	ks := &KeySetV0{pk, sk}

	// bls
	// private key
	seed1 := sha256.Sum256(seed)
	sk.BlsSk.SetLittleEndian(seed1[:])

	seed2 := sha256.Sum256(seed1[:])
	sk.ElemSk.SetLittleEndian(seed2[:])

	var frSeed Fr
	seed3 := sha256.Sum256(seed2[:])
	frSeed.SetLittleEndian(seed3[:])

	err := pk.ElemG1s[0].HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	seed4 := sha256.Sum256(seed3[:])
	frSeed.SetLittleEndian(seed4[:])
	err = pk.ElemG2s[0].HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	seed5 := sha256.Sum256(seed4[:])
	frSeed.SetLittleEndian(seed5[:])
	err = pk.SignG2.HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	ks.Calculate()

	// return instance
	return ks, nil
}

// GenKeySetV0 create instance
func GenKeySetV0() (*KeySetV0, error) {
	pk := &PublicKeyV0{
		Count:    PDPCountV1,
		TagCount: TagAtomNumV1,
		ElemG1s:  make([]G1, TagAtomNumV1),
		ElemG2s:  make([]G2, PDPCountV1),
	}
	sk := &SecretKeyV0{
		ElemPowerSk: make([]Fr, PDPCountV1),
	}
	ks := &KeySetV0{pk, sk}

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
func (k *KeySetV0) Calculate() {
	var oneFr Fr
	oneFr.SetInt64(1)
	k.Sk.ElemPowerSk[0] = oneFr
	var i int64
	for i = 1; i < k.Pk.Count; i++ {
		mcl.FrMul(&k.Sk.ElemPowerSk[i], &k.Sk.ElemPowerSk[i-1], &k.Sk.ElemSk)
	}

	mcl.G2Mul(&k.Pk.BlsPk, &k.Pk.SignG2, &k.Sk.BlsSk)
	// U = u^(x^i), i = 0, 1, ..., tagCount-1
	for i = 1; i < k.Pk.TagCount; i++ {
		mcl.G1Mul(&k.Pk.ElemG1s[i], &k.Pk.ElemG1s[0], &k.Sk.ElemPowerSk[i])
	}

	// W = w^(x^i), i = 0, 1, ..., count-1
	for i = 1; i < k.Pk.Count; i++ {
		mcl.G2Mul(&k.Pk.ElemG2s[i], &k.Pk.ElemG2s[0], &k.Sk.ElemPowerSk[i])
	}
	return
}

// GenTag create tag for *SINGLE* segment
// typ: 32B atom or 24B atom
// mode: sign or not
func (k *KeySetV0) GenTag(index []byte, segment []byte, start, typ int, mode bool) ([]byte, error) {
	if k == nil || k.Pk == nil {
		return nil, ErrKeyIsNil
	}

	var uMiDel G1

	atoms, err := splitSegmentToAtoms(segment, typ)
	if err != nil {
		return nil, err
	}

	if int64(len(atoms)+start) > k.Pk.TagCount || start < 0 {
		return nil, ErrNumOutOfRange
	}

	// Prod(u_j^M_ij)，即Prod(u^Sigma(x^j*M_ij))
	if k.Sk != nil {
		var power Fr
		if start == 0 {
			mcl.FrEvaluatePolynomial(&power, atoms, &(k.Sk.ElemPowerSk[1]))
		} else {
			power.Clear() // Set0
			for j, atom := range atoms {
				var mid Fr
				i := j + start
				mcl.FrMul(&mid, &(k.Sk.ElemPowerSk[i]), &atom) // Xi * Mi
				mcl.FrAdd(&power, &power, &mid)                // power = Sigma(Xi*Mi)
			}
		}
		mcl.G1Mul(&uMiDel, &(k.Pk.ElemG1s[0]), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		//FrEvaluatePolynomial
		for j, atom := range atoms {
			// var Mi Fr
			var mid G1
			i := j + start
			mcl.G1Mul(&mid, &k.Pk.ElemG1s[i], &atom) // uMiDel = ui ^ Mi)
			mcl.G1Add(&uMiDel, &uMiDel, &mid)
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
		mcl.G1Add(&uMiDel, &HWi, &uMiDel)
	}

	// sign
	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		mcl.G1Mul(&uMiDel, &uMiDel, &(k.Sk.BlsSk))
	}

	return uMiDel.Serialize(), nil
}

// GenChallengeV0 根据时间随机选取的待挑战segments由随机数c对各offset取模而得
func GenChallengeV0(chal *mpb.ChalInfo) int64 {
	newHash, err := blake2b.New256(nil)
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
func (pk *PublicKeyV0) VerifyTag(index, segment, tag []byte) bool {
	if pk == nil {
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

	if t.IsZero() {
		return false
	}

	atoms, err := splitSegmentToAtoms(segment, 32)
	if err != nil {
		return false
	}

	for j, atom := range atoms {
		mcl.G1Mul(&mido, &pk.ElemG1s[j], &atom) // mido = uj ^ mij
		mcl.G1Add(&midt, &midt, &mido)          // midt = Prod(uj^mij)
	}

	mcl.G1Add(&formula, &HWi, &midt) // formula = H(Wi) * Prod(uj^mij)

	mcl.Pairing(&left, &t, &(pk.SignG2))       // left = e(tag, g)
	mcl.Pairing(&right, &formula, &(pk.BlsPk)) // right = e(H(Wi) * Prod(uj^mij), pk)

	return left.IsEqual(&right)
}

// GenProofV0 gens
func (pk *PublicKeyV0) GenProof(chal Challenge, segments, tags [][]byte, typ int) (Proof, error) {
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
	if len(pk.ElemG1s) < tagNum {
		return nil, ErrNumOutOfRange
	}
	// muProd = Prod(u_j^sums_j)
	mu := make([]G1, tagNum)
	var muProd G1
	muProd.Clear()
	for j, sum := range sums {
		mcl.G1Mul(&mu[j], &pk.ElemG1s[j], &sum) // mu_j = U_j ^ sum_j
		mcl.G1Add(&muProd, &muProd, &mu[j])     // mu = Prod(U_j^sum_j)
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
		mcl.G1Add(&delta, &delta, &t)
	}

	// need modify according to (mu and delta), make sure c is unpredictable
	newHash := sha256.New()
	newHash.Write(delta.Serialize())
	newHash.Write(muProd.Serialize())

	hashValue := newHash.Sum(nil)
	cmix := binary.LittleEndian.Uint64(hashValue[:])
	rand.Seed(chal.GetSeed() + int64(cmix))
	c := rand.Int63n(pk.Count - pk.TagCount)

	if int64(len(pk.ElemG2s)) < int64(tagNum)+c {
		return nil, ErrChalOutOfRange
	}
	// 计算h_j = u_(c+j), j = 0, 1, ..., k-1
	// 对于BLS12_381,h_j = w_(c+j)
	// nuProd = Prod(h_j^sums_j)
	nu := make([]G2, tagNum)
	var nuProd G2
	nuProd.Clear()
	for j, sum := range sums {
		mcl.G2Mul(&nu[j], &pk.ElemG2s[c+int64(j)], &sum) // nu_j = h_j ^ sum_j
		mcl.G2Add(&nuProd, &nuProd, &nu[j])              // nu = Prod(h_j^sum_j)
	}

	return &ProofV0{
		Mu:    muProd.Serialize(),
		Nu:    nuProd.Serialize(),
		Delta: delta.Serialize(),
	}, nil
}

// VerifyProofV0 verify proof
func (vk *PublicKeyV0) VerifyProof(chal Challenge, proof Proof, mode bool) (bool, error) {
	if vk == nil {
		return false, ErrKeyIsNil
	}

	pf := proof.(*ProofV0)
	var mu, delta G1
	var nu G2
	err := mu.Deserialize(pf.Mu)
	if err != nil {
		return false, err
	}

	if mu.IsZero() {
		return false, nil
	}

	err = nu.Deserialize(pf.Nu)
	if err != nil {
		return false, err
	}

	if nu.IsZero() {
		return false, nil
	}

	err = delta.Deserialize(pf.Delta)
	if err != nil {
		return false, err
	}

	if delta.IsZero() {
		return false, nil
	}

	var ProdHWi, ProdHWimu, HWi G1
	var lhs1, lhs2, rhs1, rhs2 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	mcl.Pairing(&lhs1, &delta, &vk.SignG2)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	ProdHWi.Clear()
	indices := chal.GetIndices()
	for _, index := range indices {
		err := HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		mcl.G1Add(&ProdHWi, &ProdHWi, &HWi)
	}
	mcl.G1Add(&ProdHWimu, &ProdHWi, &mu)
	mcl.Pairing(&rhs1, &ProdHWimu, &vk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	if mode {
		// 第二步：验证mu与nu是对应的
		// lhs = e(mu, h0)
		newHash := sha256.New()
		newHash.Write(pf.Delta)
		newHash.Write(pf.Mu)

		hashValue := newHash.Sum(nil)
		cmix := binary.LittleEndian.Uint64(hashValue[:])
		rand.Seed(chal.GetSeed() + int64(cmix))
		c := rand.Int63n(vk.Count - vk.TagCount)

		mcl.Pairing(&lhs2, &mu, &vk.ElemG2s[c])
		// rhs = e(u, nu)
		mcl.Pairing(&rhs2, &vk.ElemG1s[0], &nu)
		// check
		if !lhs2.IsEqual(&rhs2) {
			return false, ErrVerifyStepTwo
		}
	}

	return true, nil
}

// VerifyData User or provider用于聚合验证数据完整性
func (k *KeySetV0) VerifyData(indices []string, segments, tags [][]byte, typ int) (bool, error) {
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

	if len(k.Pk.ElemG1s) < tagNum {
		return false, ErrNumOutOfRange
	}

	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, tagNum)
	var muProd G1
	muProd.Clear()
	if k.Sk != nil {
		var power Fr
		mcl.FrEvaluatePolynomial(&power, sums, &(k.Sk.ElemSk))
		mcl.G1Mul(&muProd, &(k.Pk.ElemG1s[0]), &power)
	} else {
		for j, sum := range sums {
			mcl.G1Mul(&mu[j], &(k.Pk.ElemG1s[j]), &sum) // mu_j = U_j ^ sum_j
			mcl.G1Add(&muProd, &muProd, &mu[j])         // mu = Prod(U_j^sum_j)
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

	var ProdHWi, ProdHWimu, HWi G1
	var lhs1, rhs1 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	mcl.Pairing(&lhs1, &delta, &k.Pk.SignG2)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	ProdHWi.Clear()
	for _, index := range indices {
		err := HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		mcl.G1Add(&ProdHWi, &ProdHWi, &HWi)
	}
	mcl.G1Add(&ProdHWimu, &ProdHWi, &muProd)
	mcl.Pairing(&rhs1, &ProdHWimu, &k.Pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}

func NewProofAggregatorV0(pk *PublicKeyV0, seed int64, typ int) ProofAggregatorV0 {
	sums := make([]Fr, pk.TagCount)
	var delta G1
	delta.Clear()
	return ProofAggregatorV0{pk, seed, typ, sums, delta}
}

func (pa *ProofAggregatorV0) Input(segment []byte, tag []byte) error {
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

func (pa *ProofAggregatorV0) InputMulti(segments [][]byte, tags [][]byte) error {
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

func (pa *ProofAggregatorV0) Result() (Proof, error) {
	// muProd = Prod(u_j^sums_j)
	mu := make([]G1, pa.pk.Count)
	var muProd G1
	muProd.Clear()
	for j, sum := range pa.sums {
		mcl.G1Mul(&mu[j], &pa.pk.ElemG1s[j], &sum) // mu_j = U_j ^ sum_j
		mcl.G1Add(&muProd, &muProd, &mu[j])        // mu = Prod(U_j^sum_j)
	}

	// need modify according to (mu and delta), make sure c is unpredictable
	newHash := sha256.New()
	newHash.Write(pa.delta.Serialize())
	newHash.Write(muProd.Serialize())

	hashValue := newHash.Sum(nil)
	cmix := binary.LittleEndian.Uint64(hashValue[:])
	rand.Seed(pa.seed + int64(cmix))
	c := rand.Int63n(pa.pk.Count - pa.pk.TagCount)

	if int64(len(pa.pk.ElemG2s)) < int64(pa.pk.TagCount)+c {
		return nil, ErrChalOutOfRange
	}
	// 计算h_j = u_(c+j), j = 0, 1, ..., k-1
	// 对于BLS12_381,h_j = w_(c+j)
	// nuProd = Prod(h_j^sums_j)
	nu := make([]G2, pa.pk.TagCount)
	var nuProd G2
	nuProd.Clear()
	for j, sum := range pa.sums {
		mcl.G2Mul(&nu[j], &pa.pk.ElemG2s[c+int64(j)], &sum) // nu_j = h_j ^ sum_j
		mcl.G2Add(&nuProd, &nuProd, &nu[j])                 // nu = Prod(h_j^sum_j)
	}

	return &ProofV0{
		Mu:    muProd.Serialize(),
		Nu:    nuProd.Serialize(),
		Delta: pa.delta.Serialize(),
	}, nil
}

func NewDataVerifierV0(pk *PublicKeyV0, sk *SecretKeyV0, typ int) DataVerifierV0 {
	sums := make([]Fr, pk.TagCount)
	var delta G1
	delta.Clear()
	var prodHWi G1
	prodHWi.Clear()
	return DataVerifierV0{pk, sk, typ, sums, delta, prodHWi}
}

func (dv *DataVerifierV0) Input(index []byte, segment []byte, tag []byte) error {
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

	var HWi G1
	err = HWi.HashAndMapTo([]byte(index))
	if err != nil {
		return err
	}
	mcl.G1Add(&dv.prodHWi, &dv.prodHWi, &HWi)

	return nil
}

func (dv *DataVerifierV0) InputMulti(indices [][]byte, segments [][]byte, tags [][]byte) error {
	if len(tags) != len(segments) || len(indices) != len(segments) || len(segments) < 0 {
		return ErrNumOutOfRange
	}
	le := len(segments[0])
	var t G1
	var HWi G1
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

		err = HWi.HashAndMapTo([]byte(indices[i]))
		if err != nil {
			return err
		}
		mcl.G1Add(&dv.prodHWi, &dv.prodHWi, &HWi)
	}

	return nil
}

func (dv *DataVerifierV0) Result() (bool, error) {
	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, dv.pk.TagCount)
	var muProd G1
	muProd.Clear()
	if dv.sk != nil {
		var power Fr
		mcl.FrEvaluatePolynomial(&power, dv.sums, &(dv.sk.ElemPowerSk[1]))
		mcl.G1Mul(&muProd, &(dv.pk.ElemG1s[0]), &power)
	} else {
		for j, sum := range dv.sums {
			mcl.G1Mul(&mu[j], &(dv.pk.ElemG1s[j]), &sum) // mu_j = U_j ^ sum_j
			mcl.G1Add(&muProd, &muProd, &mu[j])          // mu = Prod(U_j^sum_j)
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

	mcl.G1Add(&ProdHWimu, &dv.prodHWi, &muProd)
	mcl.Pairing(&rhs1, &ProdHWimu, &dv.pk.BlsPk)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}
