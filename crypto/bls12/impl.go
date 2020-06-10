package mcl

//TODO: incomlete

//TagGenerator a type for accumulating data to generate tag.
type MclTagAccumulator struct {
	keyset     *KeySet
	typ        int    //atom szie
	mode       bool   //sign or not
	start      int    //start index
	Cur        int    //Current index in atomsState
	atomCount  int    //how much atomCount, 0 mean unlimited
	atomsState []Fr   //tempState
	buf        []byte // buf
}

//TagGenerator a type for generating tags.
type MclTagGenerator struct {
	keyset *KeySet
	typ    int  //atom szie
	mode   bool //sign or not
}

//ChalGenerator a type for generating challenges.
type MclChalGenerator struct {
	seed      []byte
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
	seed []byte
	pk   *PublicKey
}

//ProofVerifier a type for verifing multi proofs.
//可以add很多proof统一验证
type MclMultiProofVerifier struct {
	seed       []byte
	pk         *PublicKey
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

func NewTagAccumulator(keyset *KeySet, atomCount, typ int, mode bool) (res *MclTagAccumulator, err error) {
	if typ > 32 || typ <= 0 {
		return nil, ErrSegmentSize
	}
	res = &MclTagAccumulator{
		keyset:    keyset,
		atomCount: atomCount,
		typ:       typ,
		mode:      mode,
		Cur:       0,
		buf:       make([]byte, typ),
	}
	if atomCount > 0 {
		res.atomsState = make([]Fr, atomCount)
	}
	return res, err
}

func (ta *MclTagAccumulator) AddAtoms([]byte) error {
	return nil
}

func (ta *MclTagAccumulator) Result() ([]byte, error) {
	return nil, nil
}

func (ta *MclTagAccumulator) ResultWith([]byte) ([]byte, error) {
	return nil, nil
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

func (tg *MclTagGenerator) GenTag(index []byte, segment []byte) ([]byte, error) {
	return nil, nil
}

func NewChalGenerator(seed []byte, chalCount int) *MclChalGenerator {
	return &MclChalGenerator{
		seed:      seed,
		chalCount: chalCount,
	}
}

func (cg *MclChalGenerator) GenChalNums(UserID string, bucketID, startStripeID, endStripeID int64) (res []int) {
	return nil
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

func (pg *MclProofGenerator) AddSegment(segment []byte) error {
	return nil
}

func (pg *MclProofGenerator) AddTag(tag []byte) error {
	return nil
}

func (pg *MclProofGenerator) AddSegmentAndTag(segment []byte, tag []byte) error {
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
