package aes

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestAes(t *testing.T) {
	key := make([]byte, KeySize)
	data := make([]byte, BlockSize*2)
	rand.Seed(0)
	fillRandom(key)
	fillRandom(data)
	fmt.Println(data)
	fmt.Println(key)

	tmpdata, err := AesEncrypt(data, key)
	if err != nil {
		t.Fatal("AesEncrypt error")
	}
	fmt.Println(tmpdata)

	newdata, err := AesDecrypt(tmpdata, key)
	if err != nil {
		t.Fatal("AesEncrypt error")
	}
	fmt.Println(newdata)
}

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}
