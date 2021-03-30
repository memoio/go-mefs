package pdp

import (
	"encoding/binary"
	"io"

	mcl "github.com/memoio/go-mefs/crypto/bls12"
)

var Init = mcl.Init
var BLS12_381 = mcl.BLS12_381

type G1 = mcl.G1
type G2 = mcl.G2
type GT = mcl.GT
type Fr = mcl.Fr

// Tag constants
const (
	CRC32 = 1
	BLS   = 2
	PDPV0 = 3
	PDPV1 = 4
)

// TagMap maps a hash code to it's default length
var TagMap = map[int]int{
	CRC32: 4,
	BLS:   32,
	PDPV0: 48,
	PDPV1: 48,
}

type FaultBlocks struct {
	ID       string
	BucketID uint64
	CID      map[uint64][]uint64 //stripe指向一个[]uint64，这是一个位图，第几位标记上了，就代表stripe内的该块损坏
}

//将proof序列化后，加个版本号
type ProofWithVersion struct {
	Version uint64
	Proof   Proof
}

func (pfv *ProofWithVersion) Serialize() ([]byte, error) {
	if pfv == nil {
		return nil, ErrKeyIsNil
	}
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, pfv.Version)
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
	if v == PDPV0 {
		proof = new(ProofV0)
	} else if v == PDPV1 {
		proof = new(ProofV1)
	}
	err := proof.Deserialize(data[n:])
	pfv.Proof = proof
	pfv.Version = v
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
	Version uint64
	Chal    ChallengeSeed
}

func (ch *ChallengeWithVersion) Serialize() ([]byte, error) {
	if ch == nil {
		return nil, ErrKeyIsNil
	}
	lenBuf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(lenBuf, ch.Version)
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
	// if v == PDPV0 {
	// 	chal = new(ChallengeV0)
	// } else if v == PDPV1 {
	// 	chal = new(ChallengeV1)
	// }
	err := chal.Deserialize(data[n:])
	ch.Chal = chal
	ch.Version = v
	return err
}

//当然，其实这里只记录stripe号，因为按理说，一个节点只能存储该stripe内的一个块，不可能冲突
