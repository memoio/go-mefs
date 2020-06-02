package mcl

import (
	"crypto/rand"
	"crypto/rsa"
	mbig "math/big"

	big "github.com/ncw/gmp"
)

// 带有难度值的门限随机转换函数
//trap door random funtion

//GenParams 利用生成参数，返回一个阶N以及对应的phi(N)
func GenParams() (*big.Int, *big.Int, error) {
	key, err := rsa.GenerateKey(rand.Reader, 256)
	if err != nil {
		return nil, nil, err
	}
	phi := mbig.NewInt(1)
	temp := new(mbig.Int)
	int1 := mbig.NewInt(1)
	for _, prime := range key.Primes {
		phi.Mul(phi, temp.Div(key.N, prime).Sub(temp, int1))
	}
	phiConv, ok := new(big.Int).SetString(phi.String(), 10)
	if !ok {
		return nil, nil, ErrSetBigInt
	}
	NConv, ok := new(big.Int).SetString(key.N.String(), 10)
	if !ok {
		return nil, nil, ErrSetBigInt
	}
	return NConv, phiConv, nil
}
