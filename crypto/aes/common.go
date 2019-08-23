package aes

import "errors"

const (
	KeySize   = 32 // 256bit，32B
	BlockSize = 16 // 128bit，16B
)

var (
	ErrCipher    = errors.New("NewCIpher error")
	ErrKeySize   = errors.New("Keysize error")
	ErrBlockSize = errors.New("Blocksize error,the blocksize must be an integer which can be divisible by 128")
)
