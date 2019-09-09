package keeper

import (
	"bytes"
	"log"
	"math/rand"
	"testing"

	mcl "github.com/memoio/go-mefs/bls12"
	dataformat "github.com/memoio/go-mefs/data-format"
)

func TestRepair(t *testing.T) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Println(err)
	}
	keySet, _ := mcl.GenKeySet()
	//providerID := "17nCkDaiLjr2guQG1aYxkdXZsUg"
	//userID := "17nCkDaiM31QP53ddaxwT8pSyzK"
	// blockID := "17nCkDaiLjeKXvQVDBtieLTasVP_1_0_0"
	DataCount, ParityCount, SegmentSize := 3, 4, 4096

	tmpData := make([]byte, 1024*1024)
	rand.Seed(0)
	fillRandom(tmpData)
	blocks, _, err := dataformat.EncodeDataToPreStripe(tmpData, "17nCkDaiLjeKXvQVDBtieLTasVP_1_0", DataCount, ParityCount, dataformat.BLS12, uint64(SegmentSize), keySet)
	if err != nil {
		t.Fatal(err)
	}

	// splitedID := strings.Split(blockID, "_")
	// //gid := splitedID[1]
	// //sid := splitedID[2]
	// bid := splitedID[3]
	// //nbid, _ := strconv.Atoi(bid)
	// outs := ""
	// out := strings.Split(outs, "\n")

	// for _, kvs := range out {
	// 	kv := strings.Split(kvs, "\t")
	// 	//cid := (kv[0])[11:]
	// 	subs := strings.Replace(kv[0], "/pid/offset", "", 1)
	// 	cid := (strings.Split(subs, "_"))[3]
	// 	//i, _ := strconv.ParseUint(cid, 10, 64)
	// 	i, _ := strconv.Atoi(cid)
	// 	if bid == cid {
	// 		blocks[i] = nil
	// 	}
	// }
	blocks[1] = nil
	blocks[3] = nil
	blocks[5] = nil
	newstripe, _ := dataformat.RecoverStripe(blocks)
	if bytes.Equal(newstripe[1], blocks[1]) {
		log.Println("no1 equal")
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

// func TestSearchNewProvider(t *testing.T) {
// 	uid := "17nCkDaiM31QP53ddaxwT8pSyzK"
// 	UID, _ := peer.IDB58Decode(uid)
// 	groupid := "0"
// 	s := SearchNewProvider(UID, groupid)
// 	log.Println("s :", s)
// }
