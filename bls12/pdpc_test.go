package mcl

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

const SegSize = 4 * 1024
const FileSize = 10 * 1024 * 1024
const SegNum = FileSize / SegSize

// 测试tag的形成
func BenchmarkGenTag(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		panic(err)
	}

	// sample data
	// 10MB
	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	segmentCount := (len(data) - 1) / SegSize
	segments := make([][]byte, segmentCount)
	for i := 0; i < segmentCount; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
	}

	// ------------- the data owner --------------- //
	tagTable := make(map[string][]byte)
	// add index/data pair
	// index := fmt.Sprintf("%s", sampleIdPrefix)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, segment := range segments {
			// generate the data tag
			tagTable[strconv.Itoa(j)], err = keySet.GenTag([]byte(strconv.Itoa(j)), segment, 0, 32, true)
			if err != nil {
				println(err.Error())
			}
		}
	}
}

func BenchmarkGenChallenge(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		panic(err)
	}

	// sample data
	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	segments := make([][]byte, SegNum)
	blocks := make([]string, SegNum)
	for i := 0; i < SegNum; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
		blocks[i] = strconv.Itoa(i)
	}

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum)
	for j, segment := range segments {
		// generate the data tag
		tags[j], err = keySet.GenTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, 0, 32, true)
		if err != nil {
			panic("Error")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//_ = GenChallenge(int64(i), blocks)
	}
}

func BenchmarkGenProof(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	segments := make([][]byte, SegNum)
	blocks := make([]string, SegNum)
	for i := 0; i < SegNum; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
		blocks[i] = strconv.Itoa(i)
	}

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum)
	for j, segment := range segments {
		// generate the data tag
		tags[j], err = keySet.GenTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, 0, 32, true)
		if err != nil {
			panic("Error")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := Challenge{
		Seed:    0,
		Indices: blocks,
	}

	// ------------- the storage provider ---------------- //
	// fetch the tag & challenge
	for j, segment := range segments {
		index := strconv.Itoa(j) + "_" + "0"
		boo := keySet.VerifyTag([]byte(index), segment, tags[j])
		if boo == false {
			println("VerifyTag: ", boo)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// generate the proof
		_, err = keySet.GenProof(chal, segments, tags, 32)
		if err != nil {
			panic("Error")
		}
	}
}

func BenchmarkVerifyProof(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// sample data
	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetWithSeed(data[:32], 1024, 2048)
	if err != nil {
		panic(err)
	}

	segments := make([][]byte, SegNum)
	blocks := make([]string, SegNum)
	for i := 0; i < SegNum; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
		blocks[i] = strconv.Itoa(i)
	}

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum)
	for j, segment := range segments {
		// generate the data tag
		tags[j], err = keySet.GenTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, 0, 32, true)
		if err != nil {
			panic(err)
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := Challenge{
		Seed:    time.Now().Unix(),
		Indices: blocks,
	}

	// ------------- the storage provider ---------------- //
	// fetch the tag & challenge
	for j, segment := range segments {
		index := strconv.Itoa(j) + "_" + "0"
		boo := keySet.VerifyTag([]byte(index), segment, tags[j])
		if boo == false {
			println("VerifyTag: ", boo)
		}
	}
	// generate the proof
	proof, err := keySet.GenProof(chal, segments, tags, 32)
	if err != nil {
		panic("Error")
	}

	// -------------- TPA --------------- //
	// Verify the proof
	h := Challenge{}
	h.Seed = chal.Seed
	h.Indices = make([]string, SegNum)
	for i := range chal.Indices {
		bid := strconv.Itoa(i)
		h.Indices[i] = bid + "_" + "0"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := keySet.VerifyProof(h, proof, true)
		if err != nil {
			panic("Error")
		}
		if !result {
			b.Errorf("Verificaition failed!")
		}
	}
}

func TestVerifyProof(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetWithSeed(data[:32], 1024, 2048)
	if err != nil {
		panic(err)
	}

	segments := make([][]byte, SegNum/2)
	blocks := make([]string, SegNum/2)
	for i := 0; i < SegNum/4; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
		blocks[i] = strconv.Itoa(i)
	}

	for i := 0; i < SegNum/4; i++ {
		segments[SegNum/4+i] = data[SegSize*(SegNum/4+2*i) : SegSize*(SegNum/4+2*i+2)]
		blocks[SegNum/4+i] = strconv.Itoa(i)
	}

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum/2)
	for j, segment := range segments {
		// generate the data tag
		tags[j], err = keySet.GenTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, 0, 32, true)
		if err != nil {
			panic("Error")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := Challenge{
		Seed:    time.Now().Unix(),
		Indices: blocks,
	}

	// ------------- the storage provider ---------------- //
	// fetch the tag & challenge
	for j, segment := range segments {
		index := strconv.Itoa(j) + "_" + "0"
		boo := keySet.VerifyTag([]byte(index), segment, tags[j])
		if boo == false {
			println("VerifyTag: ", boo)
		}
	}
	// generate the proof
	proof, err := keySet.GenProof(chal, segments, tags, 32)
	if err != nil {
		panic(err)
	}

	// -------------- TPA --------------- //
	// Verify the proof
	h := Challenge{}
	h.Seed = chal.Seed
	h.Indices = make([]string, SegNum/2)
	for i := range chal.Indices {
		bid := strconv.Itoa(i)
		h.Indices[i] = bid + "_" + "0"
	}

	result, err := keySet.VerifyProof(h, proof, true)
	if err != nil {
		panic("Error")
	}
	if !result {
		t.Errorf("Verificaition failed!")
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
