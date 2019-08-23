package code

const (
	CodeTypeOK            uint32 = 0
	CodeTypeEncodingError uint32 = 1
	CodeTypeBadNonce      uint32 = 2
	CodeTypeUnauthorized  uint32 = 3
	CodeTypeUnknownError  uint32 = 4
)

var (
	BlockMetaPrefix = []byte("/block")
	KvPairPrefixKey = []byte("/kv")
	ChalReqKey      = []byte("/chalReq")
	ChalResKey      = []byte("/chalRes")
	PrefixMatchKey  = []byte("prefix")
)
