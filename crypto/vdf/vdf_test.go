package vdf

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

//生成随机[]byte
func RandBytesMaskImprSrc(n int) []byte {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return b
}

func BenchmarkVdf(b *testing.B) {
	//start := time.Now()

	vdf := NewVDF(128)
	//	fmt.Println("初始化编解码器")

	sourceData := RandBytesMaskImprSrc(16)
	key := big.NewInt(2433141)
	round := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 256; j++ {
			en := vdf.encode(sourceData, key, round)
			de := vdf.decode(en, key, round)
			de[0] = 0
		}
	}
}

func BenchmarkVdfEncode(b *testing.B) {
	//start := time.Now()

	vdf := NewVDF(128)
	//	fmt.Println("初始化编解码器")

	sourceData := RandBytesMaskImprSrc(16)
	key := big.NewInt(2433141)
	round := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 256; j++ {
			en := vdf.encode(sourceData, key, round)
			en[0] = 0
		}
	}
}

func BenchmarkVdfDecode(b *testing.B) {
	//start := time.Now()

	vdf := NewVDF(128)
	//	fmt.Println("初始化编解码器")

	sourceData := RandBytesMaskImprSrc(16)
	key := big.NewInt(2433141)
	round := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 256; j++ {
			de := vdf.decode(sourceData, key, round)
			de[0] = 0
		}
	}
}

func TestVdf(t *testing.T) {
	vdf := NewVDF(127)
	fmt.Println("初始化编解码器")

	//随机生成data
	sourceData := RandBytesMaskImprSrc(70)
	key := string(RandBytesMaskImprSrc(10))
	round := 1

	fmt.Println("sourceData =", sourceData)
	fmt.Println("key =", key)
	fmt.Println("round =", round)

	start := time.Now()

	en16 := vdf.encode16(sourceData, key, round)

	t1 := time.Now()
	enTime := t1.Sub(start)
	fmt.Println("en16Time =", enTime)
	fmt.Println("Encoded16 Data =", en16)

	de16 := vdf.decode16(en16, key, round)
	t2 := time.Now()
	deTime := t2.Sub(t1)
	fmt.Println("de16Time =", deTime)
	fmt.Println("Decoded16 Data =", de16)

	k := new(big.Int).SetBytes([]byte(key))

	en := vdf.encode(sourceData, k, round)

	fmt.Println("Encoded Data =", en)

	de := vdf.decode(en, k, round)

	fmt.Println("Decoded Data =", de)
}
