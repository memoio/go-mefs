package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// 目前key打算为user的privatekey+bucketid后的32字节
func AesEncrypt(origData, key []byte) ([]byte, error) {
	if len(origData)%BlockSize != 0 {
		return nil, ErrBlockSize
	}
	if len(key) != KeySize {
		return nil, ErrKeySize
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrCipher
	}
	blockSize := block.BlockSize()
	// 目前初始向量vi为key的前blocksize个字节
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

// 目前是对于不是16字节整数倍的数据，向最后补该数据最后一个字节的数据，直到新数据为16字节的整数倍
func PKCS5Padding(ciphertext []byte) []byte {
	padding := BlockSize - len(ciphertext)%BlockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
