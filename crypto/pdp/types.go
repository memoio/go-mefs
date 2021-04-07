package pdp

import (
	"encoding/binary"
	"io"

	bls "github.com/herumi/bls-eth-go-binary/bls"
)

var Init = bls.Init
var BLS12_381 = bls.BLS12_381

// 自/data-format/common.go，目前segment的default size为32KB
const (
	SCount = 1024
)

type G1 = bls.G1
type G2 = bls.G2
type GT = bls.GT
type Fr = bls.Fr

// Tag constants
const (
	CRC32 = 1
	BLS   = 2
	PDPV0 = 3
	PDPV1 = 4
)

var GenG1 G1
var GenG2 G2

var ZeroG1 G1
var ZeroG2 G2

var G1Size int
var G2Size int
var FrSize int
var GtSize int

// TagMap maps a hash code to it's default length
var TagMap = map[int]int{
	CRC32: 4,
	BLS:   32,
	PDPV0: 48,
	PDPV1: 48,
}

// ChallengeV1 gives
type ChallengeV1 struct {
	R       int64
	Indices []string
}

func (chal *ChallengeV1) Version() int {
	return 1
}

func (chal *ChallengeV1) GetSeed() int64 {
	return chal.R
}

func (chal *ChallengeV1) GetIndices() []string {
	return chal.Indices
}

// ProofV1 is result
type ProofV1 struct {
	Delta []byte `json:"delta"`
	Psi   []byte `json:"psi"`
	Y     []byte `json:"y"`
}

func (pf *ProofV1) Version() int {
	return 1
}

func (pf *ProofV1) Serialize() []byte {
	buf := make([]byte, 0, len(pf.Delta)+len(pf.Psi)+len(pf.Y))
	buf = append(buf, pf.Delta...)
	buf = append(buf, pf.Psi...)
	buf = append(buf, pf.Y...)

	return buf
}
func (pf *ProofV1) Deserialize(buf []byte) error {
	if len(buf) != 2*G1Size+FrSize {
		return ErrNumOutOfRange
	}
	pf.Delta = make([]byte, G1Size)
	pf.Psi = make([]byte, G1Size)
	pf.Y = make([]byte, FrSize)
	copy(pf.Delta, buf[0:G1Size])
	copy(pf.Psi, buf[G1Size:2*G1Size])
	copy(pf.Y, buf[2*G1Size:2*G1Size+FrSize])
	return nil
}

type FaultBlocks struct {
	ID       string
	BucketID uint64
	CID      map[uint64][]uint64 //stripe指向一个[]uint64，这是一个位图，第几位标记上了，就代表stripe内的该块损坏
	//当然，其实这里只记录stripe号，因为按理说，一个节点只能存储该stripe内的一个块，不可能冲突
}

//将proof序列化后，加个版本号
type ProofWithVersion struct {
	Ver   uint64
	Proof Proof
}

func (pfv *ProofWithVersion) Version() int {
	return int(pfv.Ver)
}

func (pfv *ProofWithVersion) Serialize() ([]byte, error) {
	if pfv == nil {
		return nil, ErrKeyIsNil
	}
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, pfv.Ver)
	pf := pfv.Proof.Serialize()
	buf := make([]byte, 0, n+len(pf))
	buf = append(buf, lenBuf[:n]...)
	buf = append(buf, pf...)
	return buf, nil
}

func (pfv *ProofWithVersion) Deserialize(data []byte) error {
	v, n := binary.Uvarint(data)
	length := int(n)
	if length < 0 || length > binary.MaxVarintLen64 {
		return io.ErrShortBuffer
	}
	var proof Proof
	switch v {
	case PDPV1:
		proof = new(ProofV1)
	default:
		return ErrInvalidSettings
	}

	err := proof.Deserialize(data[n:])
	pfv.Proof = proof
	pfv.Ver = v
	return err
}

type ChallengeSeed struct {
	UserID      []byte
	BucketID    int64
	Seed        int64
	StripeStart int64
	StripeEnd   int64
	SegEnd      int64
}

//TODO:
func (chal *ChallengeSeed) Serialize() ([]byte, error) {
	return nil, nil
}

func (chal *ChallengeSeed) Deserialize(data []byte) error {
	return nil
}

//将Challenge序列化后，加个版本号
type ChallengeWithVersion struct {
	Ver  uint64
	Chal ChallengeSeed
}

func (ch *ChallengeWithVersion) Version() int {
	return int(ch.Ver)
}

func (ch *ChallengeWithVersion) Serialize() ([]byte, error) {
	if ch == nil {
		return nil, ErrKeyIsNil
	}
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, ch.Ver)
	chal, err := ch.Chal.Serialize()
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0, n+len(chal))
	buf = append(buf, lenBuf[:n]...)
	buf = append(buf, chal...)
	return buf, nil
}

func (ch *ChallengeWithVersion) Deserialize(data []byte) error {
	v, n := binary.Uvarint(data)
	length := int(n)
	if length < 0 || length > binary.MaxVarintLen64 {
		return io.ErrShortBuffer
	}
	var chal ChallengeSeed
	err := chal.Deserialize(data[n:])
	ch.Chal = chal
	ch.Ver = v
	return err
}
