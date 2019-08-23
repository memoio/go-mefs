package vdf

import (
	"fmt"
	"math/big"
)

/*
const (
	T uint = 3201
)
var zero *big.Int = big.NewInt(0)
var one *big.Int = big.NewInt(1)
var five *big.Int = big.NewInt(5)
var r *big.Int = big.NewInt(1).Lsh(big.NewInt(1), T-1)
var MOD *big.Int = big.NewInt(1).Lsh(big.NewInt(1), T+1) //取模的模数
var v *big.Int = new(big.Int)
*/

var zero *big.Int = big.NewInt(0)
var one *big.Int = big.NewInt(1)
var five *big.Int = big.NewInt(5)

type VDFCodec struct {
	T   uint
	Mod *big.Int
	V   *big.Int
}

//初始化编解码器，传入参数T，返回编解码器结构
func NewVDF(t uint) *VDFCodec {
	vdfCodec := new(VDFCodec)
	vdfCodec.T = t
	vdfCodec.Mod = big.NewInt(1).Lsh(big.NewInt(1), t+1)
	r := big.NewInt(1).Lsh(big.NewInt(1), t-1)
	vdfCodec.V = setr(r)
	fmt.Println("vdfCodec.V=", vdfCodec.V)
	return vdfCodec

}

func setr(r *big.Int) *big.Int {
	v := new(big.Int)
	m := new(big.Int)
	tmp := new(big.Int)
	for i := 1; i <= 4; i++ {
		m.Mul(r, big.NewInt(int64(i)))
		m.Add(m, one)
		tmp.Mod(m, five)
		if tmp.Cmp(zero) == 0 {
			return v.Div(m, five)
		}
	}
	return nil
}

//每16byte encode一次然后拼接，decode同理
func (codec *VDFCodec) encode16(endata []byte, key string, round int) []byte {

	if round == 0 {
		return endata
	} else {
		length := len(endata)
		k := new(big.Int).SetBytes([]byte(key))
		var out []byte
		var temp []byte

		for i := 0; i < length/16; i++ {
			temp = endata[i*16 : (i+1)*16]
			temp = codec.encode(temp, k, round)
			out = append(out, temp...)
		}
		d := length % 16
		if d == 0 {
			return out
		} else {
			temp = endata[length-d : length]
			temp = codec.encode(temp, k, round)
			out = append(out, temp...)
			return out
		}
	}
}

func (codec *VDFCodec) decode16(dedata []byte, key string, round int) []byte {
	if round == 0 {
		return dedata
	} else {
		length := len(dedata)
		k := new(big.Int).SetBytes([]byte(key))
		var out []byte
		var temp []byte

		for i := 0; i < length/16; i++ {
			temp = dedata[i*16 : (i+1)*16]
			temp = codec.decode(temp, k, round)
			out = append(out, temp...)
		}
		d := length % 16
		if d == 0 {
			return out
		} else {
			temp = dedata[length-d : length]
			temp = codec.decode(temp, k, round)
			out = append(out, temp...)
			return out
		}
	}

}

//编码函数，循环  {plain = (plain + key + flag) ** v mod MOD + flag}
//传入参数为 plain 原数据，key 加密用key， round 加密次数
func (codec *VDFCodec) encode(plain []byte, key *big.Int, round int) []byte {

	flag := big.NewInt(0)
	out := new(big.Int).SetBytes(plain)

	if round == 0 {
		return out.Bytes()
	} else {
		for ; round > 0; round-- {
			out.Add(out, key)
			if out.Bit(0) == 0 { //保证结果为奇数
				out.Add(out, one)
				flag.SetInt64(1)
			} else {
				flag.SetInt64(0)
			}
			out.Exp(out, codec.V, codec.Mod)
			out.Add(out, flag)
		}
	}
	out.Mod(out, codec.Mod)
	return out.Bytes()
}

//解码函数，循环{crypto = (crypto - flag) ** v mod MOD - key - flag}
func (codec *VDFCodec) decode(crypto []byte, key *big.Int, round int) []byte {

	flag := big.NewInt(0)
	out := new(big.Int).SetBytes(crypto)

	if round == 0 {
		return out.Bytes()
	} else {
		for i := 0; i < round; i++ {
			if out.Bit(0) == 0 {
				out.Sub(out, one)
				flag.SetInt64(1)
			} else {
				flag.SetInt64(0)
			}
			out.Exp(out, five, codec.Mod)
			out.Sub(out, key)
			out.Sub(out, flag)
		}
	}
	return out.Bytes()
}
