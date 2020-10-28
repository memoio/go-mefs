package mcl

import (
	"crypto/sha256"
)

// the data structures for the proof of data possession

// PublicKey is bls public key
type NPublicKey struct {
	Count   int
	BlsPk   G2
	G1Base  G1
	G2Base  G2
	ElemGt  GT
	ElemG1s []G1
	ElemG2s []G2
}

// SecretKey is bls secret key
type NSecretKey struct {
	BlsSk       Fr
	ElemSk      Fr
	ElemPowerSk []Fr
}

// KeySet is wrap
type NKeySet struct {
	Pk *NPublicKey
	Sk *NSecretKey
}

type NProof struct {
	Pi    []byte `json:"pi"`
	M     []byte `json:"m"`
	C     []byte `json:"c"`
	Delta []byte `json:"delta"`
}

// GenKeySetWithSeed create instance
func GenKeySetWithSeedForNPDP(seed []byte, count int) (*NKeySet, error) {
	pk := &NPublicKey{
		Count:   count,
		ElemG1s: make([]G1, 2*count),
		ElemG2s: make([]G2, count),
	}
	sk := &NSecretKey{
		ElemPowerSk: make([]Fr, count),
	}
	ks := &NKeySet{pk, sk}
	var oneFr Fr
	oneFr.SetInt64(1)
	// bls
	// private key
	seed1 := sha256.Sum256(seed)
	sk.BlsSk.SetLittleEndian(seed1[:])

	seed2 := sha256.Sum256(seed1[:])
	sk.ElemSk.SetLittleEndian(seed2[:])

	var frSeed Fr
	seed3 := sha256.Sum256(seed2[:])
	frSeed.SetLittleEndian(seed3[:])
	err := pk.G1Base.HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	seed4 := sha256.Sum256(seed3[:])
	frSeed.SetLittleEndian(seed4[:])
	err = pk.G2Base.HashAndMapTo(frSeed.Serialize())
	if err != nil {
		return nil, err
	}

	ks.Calculate()

	// return instance
	return ks, nil
}

// Calculate cals Xi = x^i, Ui and Wi i = 1, ..., N
func (k *NKeySet) Calculate() {
	var oneFr Fr
	oneFr.SetInt64(1)

	G2Mul(&k.Pk.BlsPk, &k.Pk.G2Base, &k.Sk.BlsSk)

	FrMul(&k.Sk.ElemPowerSk[0], &k.Sk.ElemSk, &oneFr)
	for i := 1; i < k.Pk.Count; i++ {
		FrMul(&k.Sk.ElemPowerSk[i], &k.Sk.ElemPowerSk[i-1], &k.Sk.ElemSk)
	}

	// U = u^(x^i), i = 1, ..., count
	for i := 0; i < k.Pk.Count; i++ {
		G1Mul(&k.Pk.ElemG1s[i], &k.Pk.G1Base, &k.Sk.ElemPowerSk[i])
	}
	// U = u^(x^i), i = count+2, ..., 2*count
	for i := k.Pk.Count; i < 2*k.Pk.Count; i++ {
		G1Mul(&k.Pk.ElemG1s[i], &k.Pk.ElemG1s[k.Pk.Count-1], &k.Sk.ElemPowerSk[i-k.Pk.Count])
	}

	// W = w^(x^i), i =  1, ..., count
	for i := 0; i < k.Pk.Count; i++ {
		G2Mul(&k.Pk.ElemG2s[i], &k.Pk.G2Base, &k.Sk.ElemPowerSk[i])
	}

	Pairing(&k.Pk.ElemGt, &k.Pk.ElemG1s[k.Pk.Count], &k.Pk.G2Base)

	// U = u^(x^(N+1)) should be secret
	k.Pk.ElemG1s[k.Pk.Count].Clear()

	return
}

// GenTag create tag for *SINGLE* segment
// typ: 32B atom or 24B atom
// mode: sign or not
func (k *NKeySet) GenTag(index []byte, segment []byte, start, typ int, mode bool) ([]byte, error) {
	if k == nil || k.Pk == nil {
		return nil, ErrKeyIsNil
	}

	var uMiDel G1

	atoms, err := splitSegmentToAtoms(segment, typ)
	if err != nil {
		return nil, err
	}

	if len(atoms)+start-1 > k.Pk.Count || start < 0 {
		return nil, ErrNumOutOfRange
	}

	// Prod(u_j^M_ij)，即Prod(u^Sigma(x^j*M_ij))
	if k.Sk != nil {
		var power Fr
		if start == 0 {
			//需要补一个零
			var frZero Fr
			frZero.Clear()
			atoms = append([]Fr{frZero}, atoms...)
			FrEvaluatePolynomial(&power, atoms, &(k.Sk.ElemSk))
		} else {
			power.Clear() // Set0
			for j, atom := range atoms {
				var mid Fr
				i := j + start
				FrMul(&mid, &(k.Sk.ElemPowerSk[i]), &atom) // Xi * Mi
				FrAdd(&power, &power, &mid)                // power = Sigma(Xi*Mi)
			}
		}
		G1Mul(&uMiDel, &(k.Pk.G1Base), &power) // uMiDel = u ^ Sigma(Xi*Mi)
	} else {
		//FrEvaluatePolynomial
		for j, atom := range atoms {
			// var Mi Fr
			var mid G1
			i := j + start
			G1Mul(&mid, &k.Pk.ElemG1s[i], &atom) // uMiDel = ui ^ Mi)
			G1Add(&uMiDel, &uMiDel, &mid)
		}
	}
	if start == 0 {
		// H(Wi)
		var HWi G1
		err = HWi.HashAndMapTo(index)
		if err != nil {
			return nil, err
		}
		// r = HWi * (u^Sgima(Xi*Mi))
		G1Add(&uMiDel, &HWi, &uMiDel)
	}

	var delta G1
	// sign
	if mode {
		// tag = (HWi * (u^Sgima(Xi*Mi))) ^ blsSK
		G1Mul(&delta, &uMiDel, &(k.Sk.BlsSk))
	}

	tag := append(uMiDel.Serialize(), delta.Serialize()...)
	return tag, nil
}

// VerifyTag check segment和tag是否对应
func (k *NKeySet) VerifyTag(index, segment, tag []byte) bool {
	if k == nil || k.Pk == nil {
		return false
	}

	var HWi, mido, midt, formula, t, c G1
	var left, right GT
	formula.Clear()

	err := HWi.HashAndMapTo(index)
	if err != nil {
		return false
	}

	err = c.Deserialize(tag[0:48])
	if err != nil {
		return false
	}

	if c.IsZero() {
		return false
	}

	err = t.Deserialize(tag[48:96])
	if err != nil {
		return false
	}

	if t.IsZero() {
		return false
	}

	atoms, err := splitSegmentToAtoms(segment, 32)
	if err != nil {
		return false
	}

	for j, atom := range atoms {
		G1Mul(&mido, &k.Pk.ElemG1s[j], &atom) // mido = uj ^ mij
		G1Add(&midt, &midt, &mido)            // midt = Prod(uj^mij)
	}

	G1Add(&formula, &HWi, &midt) // formula = H(Wi) * Prod(uj^mij)

	Pairing(&left, &c, &(k.Pk.BlsPk))
	Pairing(&right, &t, &(k.Pk.G2Base))
	if !left.IsEqual(&right) {
		return false
	}

	Pairing(&left, &t, &(k.Pk.G2Base))       // left = e(tag, g)
	Pairing(&right, &formula, &(k.Pk.BlsPk)) // right = e(H(Wi) * Prod(uj^mij), pk)

	if !left.IsEqual(&right) {
		return false
	}
	return true
}

// GenProof gens
func (k *NKeySet) GenProof(chal Challenge, segments, tags [][]byte, typ int) (*NProof, error) {
	if k == nil || k.Pk == nil || typ <= 0 {
		return nil, ErrKeyIsNil
	}
	// sums_j为待挑战的各segments位于同一位置(即j)上的atom的和
	if len(segments) == 0 || len(segments[0]) == 0 {
		return nil, ErrSegmentSize
	}

	tagNum := len(segments[0])
	for _, segment := range segments {
		if len(segment) > tagNum {
			tagNum = len(segment)
		}
	}

	var m Fr
	m.Clear()
	tagNum = tagNum / typ

	sums := make([]Fr, tagNum)

	for _, segment := range segments {
		atoms, err := splitSegmentToAtoms(segment, typ)
		if err != nil {
			return nil, ErrSplitSegmentToAtoms
		}

		for j, atom := range atoms { // 扫描各segment
			if j == int(chal.Seed) {
				FrAdd(&m, &m, &atom)
			} else {
				FrAdd(&sums[j], &sums[j], &atom)
			}
		}
	}
	if len(k.Pk.ElemG1s) < tagNum {
		return nil, ErrNumOutOfRange
	}
	// muProd = Prod(u_j^sums_j)
	mu := make([]G1, tagNum)
	var piProd G1
	piProd.Clear()
	for j, sum := range sums {
		if j != int(chal.Seed) {
			G1Mul(&mu[j], &k.Pk.ElemG1s[k.Pk.Count-int(chal.Seed)+j], &sum) // mu_j = U_j ^ sum_j
			G1Add(&piProd, &piProd, &mu[j])                                 // mu = Prod(U_j^sum_j)
		}
	}

	// c_s = Prod(c_i)
	var c G1
	c.Clear()
	for _, tag := range tags {
		var t G1
		err := t.Deserialize(tag[0:48])
		if err != nil {
			return nil, err
		}
		G1Add(&c, &c, &t)
	}

	// delta_s = Prod(delta_i)
	var delta G1
	delta.Clear()
	for _, tag := range tags {
		var t G1
		err := t.Deserialize(tag[48:96])
		if err != nil {
			return nil, err
		}
		G1Add(&delta, &delta, &t)
	}

	return &NProof{
		Pi:    piProd.Serialize(),
		M:     m.Serialize(),
		C:     c.Serialize(),
		Delta: delta.Serialize(),
	}, nil
}

func (k *NKeySet) GenSingleProof(chal Challenge, segment, tag []byte, typ int) (*NProof, error) {
	if k == nil || k.Pk == nil || typ <= 0 {
		return nil, ErrKeyIsNil
	}

	if len(segment) == 0 {
		return nil, ErrSegmentSize
	}

	tagNum := len(segment) / typ

	atoms, err := splitSegmentToAtoms(segment, typ)
	if err != nil {
		return nil, ErrSplitSegmentToAtoms
	}
	var m Fr
	m.SetInt64(0)
	var mu G1
	var piProd G1
	// muProd = Prod(u_j^sums_j)
	piProd.Clear()
	mu.Clear()
	for j := 0; j < k.Pk.Count; j++ { // 扫描各segment
		if j == int(chal.Seed) {
			m = atoms[j]
		} else {
			G1Mul(&mu, &k.Pk.ElemG1s[k.Pk.Count-int(chal.Seed)+j], &atoms[j])
			G1Add(&piProd, &piProd, &mu)
		}
	}

	if len(k.Pk.ElemG1s) < tagNum {
		return nil, ErrNumOutOfRange
	}

	// c_s = Prod(c_i)
	var c G1
	c.Clear()

	err = c.Deserialize(tag[0:48])
	if err != nil {
		return nil, err
	}

	// delta_s = Prod(delta_i)
	var delta G1
	delta.Clear()

	err = delta.Deserialize(tag[48:96])
	if err != nil {
		return nil, err
	}

	return &NProof{
		Pi:    piProd.Serialize(),
		M:     m.Serialize(),
		C:     c.Serialize(),
		Delta: delta.Serialize(),
	}, nil
}

// VerifyProof verify proof
func (k *NKeySet) VerifySingleProof(chal Challenge, pf *NProof, mode bool) (bool, error) {
	if k == nil || k.Pk == nil {
		return false, ErrKeyIsNil
	}
	var pi, c_s, delta G1
	var m Fr
	err := pi.Deserialize(pf.Pi)
	if err != nil {
		return false, err
	}

	if pi.IsZero() {
		return false, nil
	}

	err = c_s.Deserialize(pf.C)
	if err != nil {
		return false, err
	}

	if c_s.IsZero() {
		return false, nil
	}

	err = delta.Deserialize(pf.Delta)
	if err != nil {
		return false, err
	}

	if delta.IsZero() {
		return false, nil
	}

	err = m.Deserialize(pf.M)
	if err != nil {
		return false, err
	}

	if m.IsZero() {
		return false, nil
	}
	var ProdHWi, HWi G1
	ProdHWi.Clear()
	index := chal.Indices[0]
	err = HWi.HashAndMapTo([]byte(index))
	if err != nil {
		return false, err
	}
	G1Add(&ProdHWi, &ProdHWi, &HWi)

	var left1, left2, right GT
	Pairing(&left1, &c_s, &k.Pk.BlsPk)

	G1Sub(&c_s, &c_s, &ProdHWi)
	Pairing(&left2, &c_s, &k.Pk.ElemG2s[k.Pk.Count-1-int(chal.Seed)])

	G1Add(&pi, &pi, &delta)
	Pairing(&right, &pi, &k.Pk.G2Base)

	var gttmp GT
	GTPow(&gttmp, &k.Pk.ElemGt, &m)
	GTMul(&right, &right, &gttmp)

	GTMul(&left1, &left1, &left2)
	return left1.IsEqual(&right), nil
}

// VerifyProof verify proof
func (k *NKeySet) VerifyProof(chal Challenge, pf *NProof, mode bool) (bool, error) {
	if k == nil || k.Pk == nil {
		return false, ErrKeyIsNil
	}
	var pi, c_s, delta G1
	var m Fr
	err := pi.Deserialize(pf.Pi)
	if err != nil {
		return false, err
	}

	if pi.IsZero() {
		return false, nil
	}

	err = c_s.Deserialize(pf.C)
	if err != nil {
		return false, err
	}

	if c_s.IsZero() {
		return false, nil
	}

	err = delta.Deserialize(pf.Delta)
	if err != nil {
		return false, err
	}

	if delta.IsZero() {
		return false, nil
	}

	err = m.Deserialize(pf.M)
	if err != nil {
		return false, err
	}

	if m.IsZero() {
		return false, nil
	}
	var ProdHWi, HWi G1
	ProdHWi.Clear()
	for _, index := range chal.Indices {
		err := HWi.HashAndMapTo([]byte(index))
		if err != nil {
			return false, err
		}
		G1Add(&ProdHWi, &ProdHWi, &HWi)
	}

	var left1, left2, right GT
	Pairing(&left1, &c_s, &k.Pk.BlsPk)
	G1Sub(&c_s, &c_s, &ProdHWi)
	Pairing(&left2, &c_s, &k.Pk.ElemG2s[k.Pk.Count-1-int(chal.Seed)])
	GTMul(&left1, &left1, &left2)

	G1Add(&pi, &pi, &delta)
	Pairing(&right, &pi, &k.Pk.G2Base)

	var gttmp GT
	GTPow(&gttmp, &k.Pk.ElemGt, &m)
	GTMul(&right, &right, &gttmp)

	return left1.IsEqual(&right), nil
}
