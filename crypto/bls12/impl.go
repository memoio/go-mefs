package mcl

import "math/rand"

//TODO: incomlete

//TagGenerator a type for generating tags.
type MclTagGenerator struct {
	keyset *KeySet
	typ    int  //atom szie
	mode   bool //sign or not
}

//ChalGenerator a type for generating challenges.
type MclChalGenerator struct {
	seed      int64
	rander    rand.Rand
	chalCount int
}

//ProofGenerator a type for generating proofs.
type MclProofGenerator struct {
	pk           *PublicKey
	segCount     int
	tagCount     int
	segmentState []Fr
	tagState     G1
	typ          int //atom szie
}

//ProofVerifier a type for verifing proofs.
type MclProofVerifier struct {
	pk   *PublicKey
	seed []byte
}

//ProofVerifier a type for verifing multi proofs.
//可以add很多proof统一验证
type MclMultiProofVerifier struct {
	pk         *PublicKey
	seed       []byte
	indexState G1
	muState    G1
	nuState    G2
	deltaState G1
}

//TagVerifier a type for verifing the consistency of segments, tags and indice.
type MclTagVerifier struct {
	pk           *PublicKey
	indexState   []byte
	segmentState []byte
	tagState     []byte
}

//ProofAggregator a type for aggregating proofs.
type MclProofAggregator struct {
	deltaState []byte
	muState    []byte
	nuState    []byte
}

func NewTagGenerator(keyset *KeySet, atomCount, typ int, mode bool) (res *MclTagGenerator, err error) {
	if typ > 32 || typ <= 0 {
		return nil, ErrSegmentSize
	}
	res = &MclTagGenerator{
		keyset: keyset,
		typ:    typ,
		mode:   mode,
	}
	return res, nil
}

func (tg *MclTagGenerator) GenTag(index []byte, segment []byte, start int) ([]byte, error) {
	var uMiDel G1

	atoms, err := splitSegmentToAtoms(segment, tg.typ)
	if err != nil {
		return nil, err
	}

	if len(atoms)+start > tg.keyset.Pk.TagCount || start < 0 {
		return nil, ErrNumOutOfRange
	}

	// Prod(u_j^M_ij)，即Prod(u^Sigma(x^j*M_ij))
	if tg.keyset.Sk != nil {
		var power Fr
		if start == 0 {
			FrEvaluatePolynomial(&power, atoms, &(tg.keyset.Sk.ElemPowerSk[1]))
		} else {
			power.Clear() // Set0
			for j, atom := range atoms {
				var mid Fr
				i := j + start
				FrMul(&mid, &(tg.keyset.Sk.ElemPowerSk[i]), &atom) // Xi * Mi
				FrAdd(&power, &power, &mid)                        // power = Sigma(Xi*Mi)
			}
		}
		G1Mul(&uMiDel, &(tg.keyset.Pk.ElemG1s[0]), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		//FrEvaluatePolynomial
		for j, atom := range atoms {
			// var Mi Fr
			var mid G1
			i := j + start
			G1Mul(&mid, &tg.keyset.Pk.ElemG1s[i], &atom) // uMiDel = ui ^ Mi)
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
	if tg.mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		G1Mul(&uMiDel, &uMiDel, &(tg.keyset.Sk.BlsSk))
	}

	return uMiDel.Serialize(), nil
}

func NewChalGenerator(seed int64, chalCount int, rander rand.Rand) *MclChalGenerator {
	rander.Seed(seed)
	return &MclChalGenerator{
		seed:      seed,
		rander:    rander,
		chalCount: chalCount,
	}
}

func (cg *MclChalGenerator) GenChalNums(UserID string, bucketID, startStripeID, endStripeID int64) []int64 {
	res := make([]int64, 0, cg.chalCount)
	ran := endStripeID - startStripeID
	for i := 0; i < cg.chalCount; i++ {
		res = append(res, cg.rander.Int63n(ran))
	}
	return res
}

func (cg *MclChalGenerator) GenChalIndices(UserID string, bucketID, startStripeID, endStripeID, blockCount int64) (res []string) {
	return nil
}

func (cg *MclChalGenerator) GenChalAggIndicesWithFaults(UserID string, bucketID, startStripeID, endStripeID, blockCount int64, FaultsBlocks map[int][]int) (res G1) {
	return res
}

func NewProofGenerator(pk *PublicKey, typ int) (res *MclProofGenerator, err error) {
	if typ > 32 || typ <= 0 {
		return nil, ErrSegmentSize
	}
	if pk == nil {
		return nil, ErrKeyIsNil
	}
	res = &MclProofGenerator{
		pk:       pk,
		typ:      typ,
		segCount: 0,
		tagCount: 0,
	}
	return res, nil
}

func (pg *MclProofGenerator) AddSegmentAndTag(segment []byte, tag []byte) error {

	return nil
}

func (pg *MclProofGenerator) AddSegmentsAndTags(segment [][]byte, tag [][]byte) error {

	return nil
}

func (pg *MclProofGenerator) Result() (Proof, error) {
	return Proof{}, nil
}

func (pg *MclProofGenerator) ResultWith(segment []byte, tag []byte) (Proof, error) {
	return Proof{}, nil
}

func NewProofVerifier(pk *PublicKey, seed []byte) (res *MclProofVerifier, err error) {
	if pk == nil {
		return nil, ErrKeyIsNil
	}
	res = &MclProofVerifier{
		pk:   pk,
		seed: seed,
	}
	return res, nil
}

func (pv *MclProofVerifier) VerityOne() bool {
	return false
}

func (pv *MclProofVerifier) AddProofs(proofs []Proof, index [][]byte) error {
	return nil
}

func (pv *MclProofVerifier) Result() bool {
	return false
}

func NewMclTagVerifier(pk *PublicKey) (res *MclTagVerifier, err error) {
	if pk == nil {
		return nil, ErrKeyIsNil
	}
	res = &MclTagVerifier{
		pk: pk,
	}
	return res, nil
}

func (tv *MclTagVerifier) AddTag(index, seg, tag []byte) {

}

func (tv *MclTagVerifier) Result() bool {
	return false
}

func NewMclProofAggregator() (res *MclProofAggregator, err error) {
	res = &MclProofAggregator{}
	return res, nil
}
