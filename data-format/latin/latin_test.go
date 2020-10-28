package latin

import (
	"math/rand"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	rand.Seed(time.Now().Unix())
	coord, _ := Latin(13, 2)
	filecontent1 := make([]byte, 1<<23)
	fillRandom(filecontent1)
	filecontent2, err := Encode(filecontent1, coord, 13)
	if err != nil {
		t.Error(err)
	}
	filecontent3, err := Decode(filecontent2, coord, 13)
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

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Uint64()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}
