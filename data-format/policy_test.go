package dataformat

import (
	"crypto/sha256"
	"log"
	"math/rand"
	"testing"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/crypto/aes"
)

// 全局配置
var skbyte = []byte{179, 233, 48, 97, 94, 148, 140, 7, 78, 102, 169, 48, 136, 124, 152, 101, 76, 69, 210, 14, 38, 15, 176, 227, 73, 41, 135, 17, 170, 138, 242, 69}
var buckid = 1

// 因只考虑生成3+2个stripe，故测试Rs时，文件长度不超过3M；测试Mul时，文件长度不超过1M
var Rslen = 2 * 1024 * 1024
var Mullen = 1 * 1024 * 1024

func BenchmarkEncode(b *testing.B) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()

	opt := NewDefaultDataCoder(RsPolicy, 3, 2, keyset)

	data := make([]byte, 4096*3)
	rand.Seed(0)
	fillRandom(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 构建加密秘钥icy
		tmpkey := append(skbyte, byte(buckid))
		skey := sha256.Sum256(tmpkey)

		// 加密、Encode
		data, _ = aes.AesEncrypt(data, skey[:])

		// 多副本含前缀
		opt.Encode(data, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	}
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
