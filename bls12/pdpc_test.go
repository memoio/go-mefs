package mcl

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"

	big "github.com/ncw/gmp"
)

const SegSize = 4 * 1024
const FileSize = 10 * 1024 * 1024
const SegNum = FileSize / SegSize

func TestGenFile(t *testing.T) {
	testpath := path.Join(os.Getenv("HOME"), "test.data")
	data := make([]byte, FileSize)
	fillRandom(data)
	f, err := os.Create(testpath)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		t.Error(err)
	}
}

// 测试tag的形成
func BenchmarkGenTag(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		fmt.Println("Initialization panic.")
	}
	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		println(err.Error())
	}

	// sample data
	// 10MB
	sampleFile := path.Join(os.Getenv("HOME"), "test.data")

	data, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		b.Fatalf("Can't read the file.")
	}
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
			tagTable[strconv.Itoa(j)], err = GenTag(keySet, segment, []byte(strconv.Itoa(j)))
			if err != nil {
				println(err.Error())
			}
		}
	}
}

func BenchmarkGenChallenge(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		fmt.Println("Initialization panic.")
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		panic("Error")
	}

	// sample data
	sampleFile := path.Join(os.Getenv("HOME"), "test.data")

	data, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		b.Fatalf("Can't read the file.")
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
		tags[j], err = GenTag(keySet, segment, []byte(strconv.Itoa(j)+"_"+"0"))
		if err != nil {
			panic("Error")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenChallenge(blocks)
	}
}

func BenchmarkGenProof(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		fmt.Println("Initialization panic.")
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		panic("Error")
	}

	// sample data
	sampleFile := path.Join(os.Getenv("HOME"), "test.data")

	data, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		b.Fatalf("Can't read the file.")
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
		tags[j], err = GenTag(keySet, segment, []byte(strconv.Itoa(j)+"_"+"0"))
		if err != nil {
			panic("Error")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := GenChallenge(blocks)

	// ------------- the storage provider ---------------- //
	// fetch the tag & challenge
	for j, segment := range segments {
		index := strconv.Itoa(j) + "_" + "0"
		boo := VerifyTag(segment, tags[j], index, keySet.Pk)
		if boo == false {
			println("VerifyTag: ", boo)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// generate the proof
		_, err = GenProof(keySet.Pk, chal, segments, tags)
		if err != nil {
			panic("Error")
		}
	}
}

func BenchmarkVerifyProof(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		fmt.Println("Initialization panic.")
	}

	// generate the key set for proof of data possession
	keySet, err := GenKeySet()
	if err != nil {
		panic("Error")
	}

	// sample data
	sampleFile := path.Join(os.Getenv("HOME"), "test.data")

	data, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		b.Fatalf("Can't read the file.")
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
		tags[j], err = GenTag(keySet, segment, []byte(strconv.Itoa(j)+"_"+"0"))
		if err != nil {
			panic("Error")
		}
	}

	// -------------- TPA --------------- //
	// generate the challenge for data possession validation
	chal := GenChallenge(blocks)

	// ------------- the storage provider ---------------- //
	// fetch the tag & challenge
	for j, segment := range segments {
		index := strconv.Itoa(j) + "_" + "0"
		boo := VerifyTag(segment, tags[j], index, keySet.Pk)
		if boo == false {
			println("VerifyTag: ", boo)
		}
	}
	// generate the proof
	proof, err := GenProof(keySet.Pk, chal, segments, tags)
	if err != nil {
		panic("Error")
	}

	// -------------- TPA --------------- //
	// Verify the proof
	h := Challenge{}
	h.C = chal.C
	h.Indices = make([]string, SegNum)
	for i := range chal.Indices {
		bid := strconv.Itoa(i)
		h.Indices[i] = bid + "_" + "0"
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := VerifyProof(keySet.Pk, h, proof)
		if err != nil {
			panic("Error")
		}
		if !result {
			b.Errorf("Verificaition failed!")
		}
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

// 测试加入UDF后生成tag的性能
func BenchmarkGenTagUDF(b *testing.B) {
	err := Init(BLS12_381)
	if err != nil {
		fmt.Println("Initialization panic.")
	}
	// generate the key set for proof of data possession
	keySet, err := GenKeySetUDF()
	if err != nil {
		println(err.Error())
	}

	// sample data
	// 10MB
	sampleFile := path.Join(os.Getenv("HOME"), "test.data")

	data, err := ioutil.ReadFile(sampleFile)
	if err != nil {
		b.Fatalf("Can't read the file.")
	}
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
			tagTable[strconv.Itoa(j)], err = GenTagUDF(keySet, segment, []byte(strconv.Itoa(j)))
			if err != nil {
				println(err.Error())
			}
		}
	}
}

// 测试加入UDF后验证是否能通过
func TestUDF(t *testing.T) {
	var left, right Fr
	var sigma, prod *big.Int
	H := make([]*big.Int, T)

	err := Init(BLS12_381)
	if err != nil {
		fmt.Println("Initialization panic.")
	}

	keySet, err := GenKeySetUDF()
	if err != nil {
		panic("Error")
	}

	sigma = big.NewInt(0)
	prod = big.NewInt(1)
	eij := new(big.Int)
	ei := make([]byte, numOfAtoms*32)
	hs := new(big.Int)
	yh := new(big.Int)
	// 对各atom有hs = h(index_j_s), s = 0, 1, …, T-1
	for s := 0; s < T; s++ {
		r := new(big.Int)
		H[s] = r.SetBytes([]byte("ABC_"+strconv.Itoa(s))).Mod(r, phi) // 直接将index_j_s由string转化为big.Int
		sigma = sigma.Add(sigma, hs.Mul(keySet.Sk.BiXI[s], H[s])).Mod(sigma, phi)

		// calculate right
		prod = prod.Mul(prod, yh.Exp(keySet.Pk.Y[s+5], H[s], order)).Mod(prod, order)
	}

	// calculate left
	for k := 0; k < T; k++ {
		// 递归，eij = ei(j-1) ^ x, ei0 = y ^ Sigma((x^s) * hs)
		if k == 0 {
			eij.Exp(keySet.Pk.Y[0], sigma, order)
			copy(ei[0:32], eij.Bytes())
		} else {
			eij.Exp(eij, keySet.Sk.BiX, order)
			copy(ei[k*32:(k+1)*32], eij.Bytes())
		}
	}

	left.SetBigInt(new(big.Int).SetBytes(ei[5*32 : 6*32]))
	right.SetBigInt(prod)
	fmt.Println("left:", left.GetString(10), "\nright:", right.GetString(10))
	judge := left.IsEqual(&right)
	if !judge {
		t.Error("Verify failed.")
	}
}
