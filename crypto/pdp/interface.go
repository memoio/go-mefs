package pdp

type KeySet interface {
	GenTag(index []byte, segment []byte, start, typ int, mode bool) ([]byte, error)
	PublicKey() PublicKey
	SecreteKey() SecretKey
	VerifyData(indices []string, segments, tags [][]byte, typ int) (bool, error)
}

type Challenge interface {
	GetSeed() int64
	GetIndices() []string
}

type SecretKey interface {
}

type PublicKey interface {
	VerifyTag(index, segment, tag []byte) bool
	GenProof(chal Challenge, segments, tags [][]byte, typ int) (Proof, error)
	VerifyProof(chal Challenge, proof Proof, mode bool) (bool, error)
	GetTagCount() int64
}

type Proof interface {
	Serialize() []byte
	Deserialize(buf []byte) error
}

//往里面塞segment, tag，最后返回一个证明。
type ProofAggregator interface {
	Input(segment []byte, tag []byte, typ int) error
	InputMulti(segments [][]byte, tags [][]byte, typ int) error
	Result() (Proof, error)
}

//往里面加入index，segment，tag，最后返回是否能通过证明。
type DataVerifier interface {
	Input(index []byte, segment []byte, tag []byte, typ int) error
	InputMulti(indices [][]byte, segments [][]byte, tags [][]byte, typ int) error
	Result() (bool, error)
}

// //聚合Indices，用于最终验证证明
// type IndicesAggregator interface {
// 	Input(index []byte) error
// 	InputMulti(indices [][]byte) error
// 	Result() ([]byte, error)
// }
