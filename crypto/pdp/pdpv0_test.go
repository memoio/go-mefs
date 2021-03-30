package pdp

import (
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	mcl "github.com/memoio/go-mefs/crypto/bls12"
)

var SegSize = 32 * 1024
var FileSize = 1 * 1024 * 1024
var SegNum = FileSize / SegSize

// 测试tag的形成
func BenchmarkGenTag(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	// generate the key set for proof of data possession
	keySet, err := GenKeySetV0()
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
	//tagTable := make(map[string][]byte)
	// add index/data pair
	// index := fmt.Sprintf("%s", sampleIdPrefix)
	b.SetBytes(int64(FileSize))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j, segment := range segments {
			// generate the data tag
			_, err = keySet.GenTag([]byte("123456"+strconv.Itoa(j)), segment, 0, 32, true)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func BenchmarkGenOneTag(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	keySet, err := GenKeySetV0()
	if err != nil {
		panic(err)
	}
	// sample data
	data := make([]byte, SegSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// ------------- the data owner --------------- //
	//tagTable := make(map[string][]byte)
	// add index/data pair
	// index := fmt.Sprintf("%s", sampleIdPrefix)
	b.SetBytes(int64(SegSize))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// generate the data tag
		_, err := keySet.GenTag([]byte("123456"+strconv.Itoa(i)), data, 0, 32, true)
		if err != nil {
			b.Error(err)
		}
	}
}

func benchmarkGenOneTag(keySet *KeySetV0) func(b *testing.B) {
	return func(b *testing.B) {
		// sample data
		data := make([]byte, SegSize)
		rand.Seed(time.Now().UnixNano())
		fillRandom(data)

		// ------------- the data owner --------------- //
		//tagTable := make(map[string][]byte)
		// add index/data pair
		// index := fmt.Sprintf("%s", sampleIdPrefix)
		b.SetBytes(int64(SegSize))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// generate the data tag
			_, err := keySet.GenTag([]byte("123456"+strconv.Itoa(i)), data, 0, 32, true)
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func BenchmarkMultiGenTag(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	keySet, err := GenKeySetV0()
	if err != nil {
		panic(err)
	}
	SegSize = 4 * 1024
	for i := 0; i < 8; i++ {
		b.Run("SegSize:"+strconv.Itoa(SegSize/1024)+"Kb", benchmarkGenOneTag(keySet))
		SegSize = SegSize * 2
	}
}
func BenchmarkGenChallenge(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV0()
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

		boo := keySet.Pk.VerifyTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, tags[j])
		if boo == false {
			panic("VerifyTag false")
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
	keySet, err := GenKeySetV0()
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
		blocks[i] = strconv.Itoa(i) + "_" + "0"
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
	chal := ChallengeV0{
		Seed:    0,
		Indices: blocks,
	}

	b.ResetTimer()
	b.SetBytes(int64(FileSize))
	for i := 0; i < b.N; i++ {
		// generate the proof
		_, err = keySet.Pk.GenProof(&chal, segments, tags, 32)
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
	keySet, err := GenKeySetV0WithSeed(data[:32], 1024, 2048)
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
	for i, segment := range segments {
		// generate the data tag
		tags[i], err = keySet.GenTag([]byte(blocks[i]), segment, 0, 32, true)
		if err != nil {
			panic(err)
		}

		boo := keySet.Pk.VerifyTag([]byte(blocks[i]), segment, tags[i])
		if boo == false {
			panic("VerifyTag false")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := ChallengeV0{
		Seed:    time.Now().Unix(),
		Indices: blocks,
	}

	// generate the proof
	proof, err := keySet.Pk.GenProof(&chal, segments, tags, 32)
	if err != nil {
		panic("Error")
	}

	// -------------- TPA --------------- //
	// Verify the proof
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := keySet.Pk.VerifyProof(&chal, proof, true)
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
	keySet, err := GenKeySetV0WithSeed(data[:32], 1024, 2048)
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
	for i, segment := range segments {
		// generate the data tag
		tag, err := keySet.GenTag([]byte(blocks[i]), segment, 0, 32, true)
		if err != nil {
			panic("gentag Error" + err.Error())
		}

		boo := keySet.Pk.VerifyTag([]byte(blocks[i]), segment, tag)
		if boo == false {
			panic("VerifyTag1")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := ChallengeV0{
		Seed:    time.Now().Unix(),
		Indices: blocks,
	}

	// generate the proof
	proof, err := keySet.Pk.GenProof(&chal, segments, tags, 32)
	if err != nil {
		panic(err)
	}

	t.Log(proof)

	// -------------- TPA --------------- //
	// Verify the proof

	result, err := keySet.Pk.VerifyProof(&chal, proof, true)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Errorf("Verificaition failed!")
	}
}

func TestProofAggregatorV0(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV0WithSeed(data[:32], 1024, 2048)
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
	for i, segment := range segments {
		// generate the data tag
		tag, err := keySet.GenTag([]byte(blocks[i]), segment, 0, 32, true)
		if err != nil {
			panic("gentag Error")
		}

		res := keySet.Pk.VerifyTag([]byte(blocks[i]), segment, tag)
		if !res {
			panic("VerifyTag failed")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := ChallengeV0{
		Seed:    time.Now().Unix(),
		Indices: blocks,
	}

	proofAggregator := NewProofAggregatorV0(keySet.Pk, chal.Seed, 32)
	err = proofAggregator.Input(segments[0], tags[0])
	err = proofAggregator.InputMulti(segments[1:], tags[1:])
	if err != nil {
		panic(err.Error())
	}

	proof, err := proofAggregator.Result()
	if err != nil {
		panic(err.Error())
	}
	// -------------- TPA --------------- //
	// Verify the proof
	result, err := keySet.Pk.VerifyProof(&chal, proof, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Errorf("Verificaition failed!")
	}
}

func TestDataVerifierV0(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV0WithSeed(data[:32], 1024, 2048)
	if err != nil {
		panic(err)
	}

	segments := make([][]byte, SegNum)
	blocks := make([][]byte, SegNum)
	for i := 0; i < SegNum; i++ {
		segments[i] = data[SegSize*i : SegSize*(i+1)]
		blocks[i] = []byte(strconv.Itoa(i))
	}

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum)
	for i, segment := range segments {
		// generate the data tag
		tag, err := keySet.GenTag([]byte(blocks[i]), segment, 0, 32, true)
		if err != nil {
			panic("gentag Error")
		}

		res := keySet.Pk.VerifyTag([]byte(blocks[i]), segment, tag)
		if !res {
			panic("VerifyTag failed")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation

	dataVerifier := NewDataVerifierV0(keySet.Pk, keySet.Sk, 32)
	err = dataVerifier.Input(blocks[0], segments[0], tags[0])
	err = dataVerifier.InputMulti(blocks[1:], segments[1:], tags[1:])
	if err != nil {
		panic(err.Error())
	}

	result, err := dataVerifier.Result()
	if err != nil {
		panic(err.Error())
	}
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Errorf("Verificaition failed!")
	}
}

func TestEvaluatePolynomial(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	k, err := GenKeySetV0()
	if err != nil {
		panic(err)
	}
	SegSize = 32 * 1024
	// sample data
	segment := make([]byte, SegSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(segment)
	atoms, err := splitSegmentToAtoms(segment, 32)
	if err != nil {
		t.Error(err)
	}
	var power1 Fr
	power1.Clear() // Set0
	mcl.FrEvaluatePolynomial(&power1, atoms, &(k.Sk.ElemPowerSk[1]))
	var power2 Fr
	power2.Clear() // Set0
	for j, atom := range atoms {
		var mid Fr
		mcl.FrMul(&mid, &(k.Sk.ElemPowerSk[j]), &atom) // Xi * Mi
		mcl.FrAdd(&power2, &power2, &mid)              // power = Sigma(Xi*Mi)
	}
	ok := power1.IsEqual(&power2)
	if !ok {
		t.Error("Not Equal")
	}
}

func benchmarkEvaluatePolynomial(k *KeySetV0) func(b *testing.B) {
	return func(b *testing.B) {
		// sample data
		segment := make([]byte, SegSize)
		rand.Seed(time.Now().UnixNano())
		fillRandom(segment)
		b.SetBytes(int64(SegSize))
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			atoms, err := splitSegmentToAtoms(segment, 32)
			if err != nil {
				b.Error(err)
			}
			var power1 Fr
			power1.Clear() // Set0
			mcl.FrEvaluatePolynomial(&power1, atoms, &(k.Sk.ElemPowerSk[1]))
		}
	}
}

func BenchmarkMultiEP(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	keySet, err := GenKeySetV0()
	if err != nil {
		panic(err)
	}
	SegSize = 4 * 1024
	for i := 0; i < 8; i++ {
		b.Run("SegSize:"+strconv.Itoa(SegSize/1024)+"KB", benchmarkEvaluatePolynomial(keySet))
		SegSize = SegSize * 2
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
