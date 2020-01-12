package aes

import (
	"fmt"
	blake2bsimd "github.com/minio/blake2b-simd"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"math/rand"
	"testing"
)

func BenchmarkBlake2s256(b *testing.B) {
	var d1 [32]byte
	fmt.Println((d1))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d1 = blake2s.Sum256([]byte("1234567"))
	}
}

func BenchmarkBlake2b512(b *testing.B) {
	var d1 [64]byte
	fmt.Println((d1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d1 = blake2b.Sum512([]byte("1234567sjjskdjkdlsllslagfgadddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddslkvjjk"))
	}
}
func BenchmarkBlake2b256(b *testing.B) {
	var d1 [32]byte
	fmt.Println((d1))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d1 = blake2b.Sum256([]byte("1234567sjjskdjkdlsllslagfgadddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddslkvjjk"))
	}
}

func BenchmarkBlake2bsimd512(b *testing.B) {
	var d1 [64]byte
	fmt.Println((d1))
	h := blake2bsimd.New512()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Reset()
		h.Write([]byte("1234567sjjskdjkdlsllslagfgadddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddslkvjjk"))
		h.Sum(nil)
	}
}

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
