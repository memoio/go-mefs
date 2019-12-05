package mcl

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

// customized errors
var (
	ErrSplitSegmentToAtoms   = errors.New("invalid segment")
	ErrKeyIsNil              = errors.New("the key is nil")
	ErrSetHashOf             = errors.New("SetHashOf is not true")
	ErrSetString             = errors.New("SetString is not true")
	ErrSetBigInt             = errors.New("SetBigInt is not true")
	ErrSetToBigInt           = errors.New("SetString (for big.Int) is not true")
	ErrNumOutOfRange         = errors.New("numOfAtoms is out of range")
	ErrSegmentSize           = errors.New("the size of the segment is wrong")
	ErrGenTag                = errors.New("GenTag failed")
	ErrOffsetIsNegative      = errors.New("offset is negative")
	ErrProofVerifyInProvider = errors.New("proof is wrong")
	ErrVerifyStepOne         = errors.New("verification failed in Step1")
	ErrVerifyStepTwo         = errors.New("verification failed in Step2")
)

// 待挑战的segment按32B进行分割，得到的切片称作atom，每个segment含numOfAtoms个atom
// 自/data-format/common.go，目前segment的default size为4KB
// numOfAtoms = (DefaultSegmentSize-1) / LengthOfAtom + 1
const LengthOfAtom = 32
const numOfAtoms = 128

// 应使(N - numOfAtoms) * (48 + 96) >> len(segment)
// 暂定为1024
const N = 1024

// the data structures for the proof of data possession
type (
	PublicKey struct {
		BlsPK G2
		G     G2
		U     []G1
		W     []G2
	}

	SecretKey struct {
		BlsSK Fr
		X     Fr
		XI    []Fr
	}

	KeySet struct {
		Pk *PublicKey
		Sk *SecretKey
	}

	Challenge struct {
		C       int
		Indices []string
	}

	// Proof由provider发送给keeper
	// 元素开头大写便于序列化导出
	Proof struct {
		Mu    []byte
		Nu    []byte
		Delta []byte
	}
)

// create instance
func GenKeySet() (*KeySet, error) {
	pk := new(PublicKey)
	sk := new(SecretKey)

	// bls
	// private key
	sk.BlsSK.SetByCSPRNG()
	sk.X.SetByCSPRNG()
	err := sk.CalculateXi()
	if err != nil {
		fmt.Println(err)
	}

	//public key
	var u G1
	var w G2
	// 借助Fr的随机方法为G1和G2产生随机元素
	var seed Fr
	seed.SetByCSPRNG()
	err = u.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}
	err = w.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}

	seed.SetByCSPRNG()
	err = pk.G.HashAndMapTo(seed.Serialize())
	if err != nil {
		return nil, err
	}
	G2Mul(&pk.BlsPK, &pk.G, &sk.BlsSK)
	// 各atom具有不同U
	// U = u^(x^i), i = 0, 1, ..., N-1
	// W = w^(x^i), i = 0, 1, ..., N-1
	pk.U = make([]G1, N) // 须指定大小，否则出现"index out of range"错误
	pk.W = make([]G2, N)
	for i := 0; i < N; i++ {
		G1Mul(&pk.U[i], &u, &sk.XI[i])
		G2Mul(&pk.W[i], &w, &sk.XI[i])
	}

	// return instance
	return &KeySet{pk, sk}, nil
}

// Xi = x^i, i = 0, 1, ..., N
func (sk *SecretKey) CalculateXi() error {
	// var mid Fr
	if len(sk.XI) != N {
		sk.XI = make([]Fr, N)
	}
	err := sk.XI[0].SetString("1", 10)
	if err != nil {
		return err
	}
	for i := 1; i < N; i++ {
		FrMul(&sk.XI[i], &sk.XI[i-1], &sk.X)
	}
	return nil
}

// -------------------- proof related routines ------------------------ //
func splitSegmentToAtoms(segment []byte) ([][]byte, error) {
	if len(segment) == 0 {
		return nil, ErrSplitSegmentToAtoms
	}

	num := (len(segment)-1)/LengthOfAtom + 1
	if num != numOfAtoms {
		return nil, ErrSegmentSize
	}

	atom := make([][]byte, numOfAtoms)

	for i := 0; i < numOfAtoms; i++ {
		atom[i] = segment[LengthOfAtom*i : LengthOfAtom*(i+1)]
	}

	return atom, nil
}

// GenTag create tag for *SINGLE* segment
func GenTag(keys *KeySet, segment []byte, index []byte) ([]byte, error) {
	var uMiDel, HWi, r G1
	var power Fr
	power.Clear() // Set0

	atoms, err := splitSegmentToAtoms(segment)
	if err != nil {
		return nil, ErrGenTag
	}

	// Prod(u_j^M_ij)，即Prod(u^Sigma(x^j*M_ij))
	for j, atom := range atoms {
		var mid, Mi Fr
		// Mi为atom而非block或segment
		judge := Mi.SetHashOf(atom)
		if !judge {
			return nil, ErrSetHashOf
		}

		FrMul(&mid, &keys.Sk.XI[j], &Mi) // Xi * Mi
		FrAdd(&power, &power, &mid)      // power = Sigma(Xi*Mi)
	}
	G1Mul(&uMiDel, &keys.Pk.U[0], &power) // uMiDel = u ^ Sigma(Xi*Mi)

	// H(Wi)
	err = HWi.HashAndMapTo(index)
	if err != nil {
		return nil, err
	}

	G1Add(&r, &HWi, &uMiDel)      // r = HWi * (u^Sgima(Xi*Mi))
	G1Mul(&r, &r, &keys.Sk.BlsSK) // tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK

	return r.Serialize(), nil
}

// GenChallenge 根据时间随机选取的待挑战segments由随机数c对各offset取模而得
func GenChallenge(src int64, blocks []string) Challenge {
	// 在[1, N-numOfAtoms]间随机选出一个整数C
	rand.Seed(src)
	var c int
	for {
		c = rand.Intn(N - numOfAtoms)
		if c != 0 {
			break
		}
	}

	return Challenge{c, blocks}
}

// VerifyChalNum 确认challenge num的正确性
func VerifyChalNum(src int64, chalNum int) bool {
	// 在[1, N-numOfAtoms]间随机选出一个整数C
	rand.Seed(src)
	var c int
	for {
		c = rand.Intn(N - numOfAtoms)
		if c != 0 {
			break
		}
	}

	return chalNum == c
}

// 检查provider取到的segment和tag是否对应
func VerifyTag(segment, tag []byte, index string, key *PublicKey) bool {
	var HWi, mido, midt, formula, t G1
	var left, right GT
	formula.Clear()

	err := HWi.HashAndMapTo([]byte(index))
	if err != nil {
		return false
	}

	err = t.Deserialize(tag)
	if err != nil {
		return false
	}

	atoms, err := splitSegmentToAtoms(segment)
	if err != nil {
		return false
	}

	for j, atom := range atoms {
		var Mi Fr
		judge := Mi.SetHashOf(atom)
		if !judge {
			return false
		}
		G1Mul(&mido, &key.U[j], &Mi) // mido = uj ^ mij
		G1Add(&midt, &midt, &mido)   // midt = Prod(uj^mij)
	}
	G1Add(&formula, &HWi, &midt) // formula = H(Wi) * Prod(uj^mij)

	Pairing(&left, &t, &key.G)            // left = e(tag, g)
	Pairing(&right, &formula, &key.BlsPK) // right = e(H(Wi) * Prod(uj^mij), pk)

	return left.IsEqual(&right)
}

// 对所有被挑战的segments生成一个proof
func GenProof(key *PublicKey, chal Challenge, segments, tags [][]byte) (string, error) {
	// GenChallenge在challenge.go的DoChallenge中完成，chal会在序列化后成为metaKey的一部分
	var err error
	//指明atoms所归属的segment
	atomsForSegment := make(map[int][][]byte)
	for i, segment := range segments {
		if len(segment) > 0 {
			atomsForSegment[i], err = splitSegmentToAtoms(segment)
			if err != nil {
				return "", ErrSplitSegmentToAtoms
			}
		}
	}

	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	var m Fr
	sums := make([]Fr, numOfAtoms)
	for j := 0; j < numOfAtoms; j++ { // 定位atom
		sums[j].Clear()
		for _, atoms := range atomsForSegment { // 扫描各segment
			if len(atoms) < numOfAtoms {
				return "", ErrNumOutOfRange
			}
			judge := m.SetHashOf(atoms[j])
			if !judge {
				return "", ErrSetHashOf
			}
			FrAdd(&sums[j], &sums[j], &m)
		}
	}
	if len(key.W) < numOfAtoms+chal.C || len(key.U) < numOfAtoms {
		return "", ErrNumOutOfRange
	}
	// 计算h_j = u_(c+j), j = 0, 1, ..., k-1
	// 对于BLS12_381,h_j = w_(c+j)
	h := make([]G2, numOfAtoms)
	for j := 0; j < numOfAtoms; j++ {
		h[j] = key.W[chal.C+j]
	}
	// muProd = Prod(u_j^sums_j)
	// nuProd = Prod(h_j^sums_j)
	mu := make([]G1, numOfAtoms)
	nu := make([]G2, numOfAtoms)
	var muProd G1
	var nuProd G2
	muProd.Clear()
	nuProd.Clear()
	for j, sum := range sums {
		G1Mul(&mu[j], &key.U[j], &sum)  // mu_j = U_j ^ sum_j
		G1Add(&muProd, &muProd, &mu[j]) // mu = Prod(U_j^sum_j)

		G2Mul(&nu[j], &h[j], &sum)      // nu_j = h_j ^ sum_j
		G2Add(&nuProd, &nuProd, &nu[j]) // nu = Prod(h_j^sum_j)
	}
	// delta = Prod(tag_i)
	var delta G1
	delta.Clear()
	for _, tag := range tags {
		var t G1
		err = t.Deserialize(tag)
		if err != nil {
			return "", err
		}
		G1Add(&delta, &delta, &t)
	}
	// 序列化
	mustr := b58.Encode(muProd.Serialize())
	nustr := b58.Encode(nuProd.Serialize())
	deltastr := b58.Encode(delta.Serialize())
	res := mustr + metainfo.DELIMITER + nustr + metainfo.DELIMITER + deltastr

	return res, nil
}

// check
func VerifyProof(key *PublicKey, chal Challenge, proofres string) (bool, error) {
	var mu, delta G1
	var nu G2
	// 反序列化
	proofs := strings.Split(proofres, metainfo.DELIMITER)
	muByte, _ := b58.Decode(proofs[0])
	err := mu.Deserialize(muByte)
	if err != nil {
		return false, err
	}
	nuByte, _ := b58.Decode(proofs[1])
	err = nu.Deserialize(nuByte)
	if err != nil {
		return false, err
	}
	delByte, _ := b58.Decode(proofs[2])
	err = delta.Deserialize(delByte)
	if err != nil {
		return false, err
	}

	var ProdHWi, ProdHWimu, HWi G1
	var lhs1, lhs2, rhs1, rhs2 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	Pairing(&lhs1, &delta, &key.G)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	ProdHWi.Clear()
	for _, index := range chal.Indices {
		err = HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		G1Add(&ProdHWi, &ProdHWi, &HWi)
	}
	G1Add(&ProdHWimu, &ProdHWi, &mu)
	Pairing(&rhs1, &ProdHWimu, &key.BlsPK)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	// 第二步：验证mu与nu是对应的
	// lhs = e(mu, h0)
	Pairing(&lhs2, &mu, &key.W[chal.C])
	// rhs = e(u, nu)
	Pairing(&rhs2, &key.U[0], &nu)
	// check
	if !lhs2.IsEqual(&rhs2) {
		return false, ErrVerifyStepTwo
	}

	return true, nil
}

// User用于聚合验证数据完整性
func VerifyDataForUser(key *PublicKey, indices []string, segments, tags [][]byte) (bool, error) {
	if (len(indices) != len(segments)) || (len(indices) != len(tags)) {
		return false, ErrNumOutOfRange
	}
	if key == nil {
		return false, ErrKeyIsNil
	}
	// TODO:根据chal.Indices获取segment和tag，并聚合
	var err error
	//指明atoms所归属的segment
	atomsForSegment := make(map[int][][]byte)
	for i, segment := range segments {
		if len(segment) > 0 {
			atomsForSegment[i], err = splitSegmentToAtoms(segment)
			if err != nil {
				return false, ErrSplitSegmentToAtoms
			}
		}
	}

	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	var m Fr
	sums := make([]Fr, numOfAtoms)
	for j := 0; j < numOfAtoms; j++ { // 定位atom
		sums[j].Clear()
		for _, atoms := range atomsForSegment { // 扫描各segment
			if len(atoms) < numOfAtoms {
				return false, ErrNumOutOfRange
			}
			judge := m.SetHashOf(atoms[j])
			if !judge {
				return false, ErrSetHashOf
			}
			FrAdd(&sums[j], &sums[j], &m)
		}
	}
	if len(key.W) < numOfAtoms || len(key.U) < numOfAtoms {
		return false, ErrNumOutOfRange
	}

	// muProd = Prod(u_j^sums_j)
	mu := make([]G1, numOfAtoms)
	var muProd G1
	muProd.Clear()
	for j, sum := range sums {
		G1Mul(&mu[j], &key.U[j], &sum)  // mu_j = U_j ^ sum_j
		G1Add(&muProd, &muProd, &mu[j]) // mu = Prod(U_j^sum_j)
	}
	// delta = Prod(tag_i)
	var delta G1
	delta.Clear()
	for _, tag := range tags {
		var t G1
		err = t.Deserialize(tag)
		if err != nil {
			return false, err
		}
		G1Add(&delta, &delta, &t)
	}

	var ProdHWi, ProdHWimu, HWi G1
	var lhs1, rhs1 GT
	// var index string
	// var offset int
	// 第一步：验证tag和mu是对应的
	// lhs = e(delta, g)
	Pairing(&lhs1, &delta, &key.G)
	// rhs = e(Prod(H(Wi)) * mu, pk)
	ProdHWi.Clear()
	for _, index := range indices {
		err = HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		G1Add(&ProdHWi, &ProdHWi, &HWi)
	}
	G1Add(&ProdHWimu, &ProdHWi, &muProd)
	Pairing(&rhs1, &ProdHWimu, &key.BlsPK)
	// check
	if !lhs1.IsEqual(&rhs1) {
		return false, ErrVerifyStepOne
	}

	return true, nil
}
