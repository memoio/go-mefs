package pdp

import (
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// 测试tag的形成
func BenchmarkGenTagV1(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1()
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
			_, err = keySet.GenTag([]byte("123456"+strconv.Itoa(j)), segment, 0, 30, true)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func BenchmarkGenOneTagV1(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	keySet, err := GenKeySetV1()
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
		_, err := keySet.GenTag([]byte("123456"+strconv.Itoa(i)), data, 0, 30, true)
		if err != nil {
			b.Error(err)
		}
	}
}

func benchmarkGenOneTagV1(keySet *KeySetV1) func(b *testing.B) {
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

func BenchmarkMultiGenTagV1(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}
	keySet, err := GenKeySetV1()
	if err != nil {
		panic(err)
	}
	SegSize = 4 * 1024
	for i := 0; i < 4; i++ {
		b.Run("SegSize:"+strconv.Itoa(SegSize/1024)+"Kb", benchmarkGenOneTagV1(keySet))
		SegSize = SegSize * 2
	}
}
func BenchmarkGenChallengeV1(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1()
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

	pk := keySet.PublicKey()

	// ------------- the data owner --------------- //
	tags := make([][]byte, SegNum)
	for j, segment := range segments {
		// generate the data tag
		tags[j], err = keySet.GenTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, 0, 32, true)
		if err != nil {
			panic("Error")
		}

		boo := pk.VerifyTag([]byte(strconv.Itoa(j)+"_"+"0"), segment, tags[j])
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

func BenchmarkGenProofV1(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1()
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
	chal := ChallengeV1{
		R:       0,
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

func BenchmarkVerifyProofV1(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	// sample data
	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1WithSeed(data[:32], SCount)
	if err != nil {
		panic(err)
	}

	vk := keySet.VerifyKey()
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

		// boo := keySet.VerifyTag([]byte(blocks[i]), segment, tags[i])
		// if boo == false {
		// 	panic("VerifyTag false")
		// }
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := ChallengeV1{
		R:       time.Now().Unix(),
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
	b.SetBytes(int64(FileSize))
	for i := 0; i < b.N; i++ {
		result, err := vk.VerifyProof(&chal, proof)
		if err != nil {
			panic("Error")
		}
		if !result {
			b.Errorf("Verificaition failed!")
		}
	}
}

func TestVerifyProofV1(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1WithSeed(data[:32], SCount)
	if err != nil {
		panic(err)
	}

	pk := keySet.PublicKey()
	vk := keySet.VerifyKey()
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

		res := pk.VerifyTag([]byte(blocks[i]), segment, tag)
		if !res {
			panic("VerifyTag failed")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := ChallengeV1{
		R:       time.Now().Unix(),
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
	result, err := vk.VerifyProof(&chal, proof)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Errorf("Verificaition failed!")
	}
}

func TestProofAggregatorV1(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1WithSeed(data[:32], SCount)
	if err != nil {
		panic(err)
	}

	pk := keySet.PublicKey()
	vk := keySet.VerifyKey()
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

		res := pk.VerifyTag([]byte(blocks[i]), segment, tag)
		if !res {
			panic("VerifyTag failed")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := ChallengeV1{
		R:       time.Now().Unix(),
		Indices: blocks,
	}

	proofAggregator := NewProofAggregatorV1(keySet.Pk, chal.R, 32)
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
	result, err := vk.VerifyProof(&chal, proof)
	if err != nil {
		t.Fatal(err)
	}
	if !result {
		t.Errorf("Verificaition failed!")
	}
}

func TestDataVerifierV1(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1WithSeed(data[:32], SCount)
	if err != nil {
		panic(err)
	}

	pk := keySet.PublicKey()
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

		res := pk.VerifyTag([]byte(blocks[i]), segment, tag)
		if !res {
			panic("VerifyTag failed")
		}
		tags[i] = tag
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation

	dataVerifier := NewDataVerifierV1(keySet.Pk, keySet.Sk, 32)
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

func TestKeyDeserialize(t *testing.T) {
	err := Init(BLS12_381)
	if err != nil {
		panic(err)
	}

	data := make([]byte, FileSize)
	rand.Seed(time.Now().UnixNano())
	fillRandom(data)

	// generate the key set for proof of data possession
	keySet, err := GenKeySetV1WithSeed(data[:32], SCount)
	if err != nil {
		panic(err)
	}
	skBytes := keySet.SecreteKey().Serialize()
	skDes := new(SecretKeyV1)
	err = skDes.Deserialize(skBytes)
	if err != nil {
		t.Fatal(err)
	}

	pk := keySet.Pk
	pkBytes := pk.Serialize()
	pkDes := new(PublicKeyV1)
	err = pkDes.Deserialize(pkBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !pkDes.BlsPk.IsEqual(&pk.BlsPk) {
		t.Fatal("not equal")
	}

}
