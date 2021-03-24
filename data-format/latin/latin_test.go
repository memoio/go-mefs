package latin

import (
	"math/rand"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	rand.Seed(time.Now().Unix())
	var size int64 = 1 << 25 //4M
	n, ok := GetN(size)
	if !ok {
		t.Fatal("")
	}
	coord, _ := Latin(n, 2)
	filecontent1 := make([]byte, size)
	fillRandom(filecontent1)
	filecontent2, err := Encode(filecontent1, coord, n)
	if err != nil {
		t.Error(err)
	}
	filecontent3, err := Decode(filecontent2, coord, n)
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 20; i++ {
		if filecontent1[i] != filecontent3[i] {
			t.Error("not equal")
			return
		}
	}
}

func BenchmarkLatin(b *testing.B) {
	rand.Seed(time.Now().Unix())
	var size int64 = 1 << 25 //32M
	n, ok := GetN(size)
	if !ok {
		b.Fatal("")
	}

	b.ResetTimer()
	b.SetBytes(size)
	for i := 0; i < b.N; i++ {
		_, _ = Latin(n, 2)
	}
}

func BenchmarkEncode(b *testing.B) {
	rand.Seed(time.Now().Unix())
	var size int64 = 1 << 25 //32M
	n, ok := GetN(size)
	if !ok {
		b.Fatal("")
	}
	coord, _ := Latin(n, 2)
	b.ResetTimer()
	b.SetBytes(size)
	for i := 0; i < b.N; i++ {
		filecontent1 := make([]byte, size)
		fillRandom(filecontent1)
		filecontent2, err := Encode(filecontent1, coord, n)
		if err != nil {
			b.Error(err)
		}
		filecontent3, err := Decode(filecontent2, coord, n)
		if err != nil {
			b.Error(err)
		}

		for i := 0; i < 20; i++ {
			if filecontent1[i] != filecontent3[i] {
				b.Error("not equal")
				return
			}
		}
	}
}

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Uint64()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}
