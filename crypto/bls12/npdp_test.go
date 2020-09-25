package mcl

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// var SegSize = 32 * 1024
// var FileSize = 8 * 1024 * 1024
// var SegNum = FileSize / SegSize

func TestVerifyProofN(t *testing.T) {
	var SegSize = 32 * 1024
	var FileSize = 128 * 1024
	var SegNum = FileSize / SegSize

	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetWithSeedForNPDP(data[:32], 1024)
	if err != nil {
		panic(err)
	}

	segments := make([][]byte, SegNum)
	blocks := make([]string, SegNum)
	for i := 0; i < SegNum; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
		blocks[i] = strconv.Itoa(i) + "x"
	}

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum)
	for i, segment := range segments {
		// generate the data tag
		tag, err := keySet.GenTag([]byte(blocks[i]), segment, 0, 32, true)
		if err != nil {
			panic("gentag Error: " + err.Error())
		}

		boo := keySet.VerifyTag([]byte(blocks[i]), segment, tag)
		if boo == false {
			panic("VerifyTag1")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := Challenge{
		Seed:    time.Now().Unix() % 1024,
		Indices: blocks,
	}

	// generate the proof
	proof, err := keySet.GenProof(chal, segments, tags, 32)
	if err != nil {
		panic(err)
	}

	fmt.Println("proof:", proof, "\nchal:", chal.Seed)

	// -------------- TPA --------------- //
	// Verify the proof

	result, err := keySet.VerifyProof(chal, proof, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Errorf("Verificaition failed!")
	}
}

func TestVerifySingleProofN(t *testing.T) {
	var SegSize = 32 * 1024

	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	segment := make([]byte, SegSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(segment)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetWithSeedForNPDP(segment[:32], 1024)
	if err != nil {
		panic(err)
	}

	index := "12345"
	// ------------- the data owner --------------- //

	// generate the data tag
	tag, err := keySet.GenTag([]byte(index), segment, 0, 32, true)
	if err != nil {
		panic("gentag Error: " + err.Error())
	}

	boo := keySet.VerifyTag([]byte(index), segment, tag)
	if boo == false {
		panic("VerifyTag1")
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := Challenge{
		Seed:    time.Now().Unix() % 1024,
		Indices: []string{index},
	}
	fmt.Println("chal:", chal.Seed)
	// generate the proof
	proof, err := keySet.GenSingleProof(chal, segment, tag, 32)
	if err != nil {
		panic(err)
	}

	fmt.Println("proof:", proof)

	// -------------- TPA --------------- //
	// Verify the proof

	result, err := keySet.VerifySingleProof(chal, proof, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Fatal("Verificaition failed!")
	}
}
