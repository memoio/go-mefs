package aes

import (
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/blake2b"
)

const (
	KeySize   = 32 // 256bit，32B
	BlockSize = 16 // 128bit，16B
)

var (
	ErrCipher    = errors.New("NewCIpher error")
	ErrKeySize   = errors.New("Keysize error")
	ErrBlockSize = errors.New("Blocksize error,the blocksize must be an integer which can be divisible by 128")
)

// CreateAesKey creates
func CreateAesKey(privateKey, queryID []byte, bucketID, objectStart int64) [32]byte {
	tmpkey := make([]byte, len(privateKey)+len(queryID)+16)
	copy(tmpkey, privateKey)
	copy(tmpkey[len(privateKey):], queryID)
	binary.LittleEndian.PutUint64(tmpkey[len(privateKey)+len(queryID):], uint64(bucketID))
	binary.LittleEndian.PutUint64(tmpkey[len(privateKey)+len(queryID)+8:], uint64(objectStart))
	return blake2b.Sum256(tmpkey)
}
