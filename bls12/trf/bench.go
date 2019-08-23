package main

import (
	"crypto/sha512"
	"fmt"
	"strconv"
	"time"

	"github.com/memoio/go-mefs/bls12"
	big "github.com/ncw/gmp"
)

func main() {
	order, phi, err := mcl.GenParams()

	fmt.Println("order:", order.BitLen(), order.String(), "\nphi:", phi.BitLen(), phi.String())
	if err != nil {
		panic(err)
	}
	//调整难度值
	t := big.NewInt(2)
	tTemp := new(big.Int)
	i2 := big.NewInt(2)
	hi := new(big.Int)
	fasttemp := new(big.Int)
	fast := new(big.Int)
	slowtemp := new(big.Int)
	slow := new(big.Int)
	for i := 0; i < 12; i++ {
		if i <= 6 {
			t.Mul(t, tTemp.Exp(i2, big.NewInt(int64(i)), nil))
		} else {
			t.Mul(t, i2)
		}
		//先对index进行Hash，然后作为底数
		h := sha512.Sum512_224([]byte("1234563hdhdbdjjdkdkjaawd" + strconv.Itoa(i)))
		hi.SetBytes(h[:])
		//运用快速和慢速进行比较
		var avafast float64
		fasttemp.Exp(i2, t, phi)
		beginTime := time.Now()
		for j := 0; j < 100; j++ {
			fast.Exp(hi, fasttemp, order)
		}
		endTime := time.Now()
		avafast = float64(endTime.Sub(beginTime).Nanoseconds())
		avafast = avafast / 100.0

		var avaslow float64
		slowtemp.Exp(i2, t, nil)
		beginTime = time.Now()
		for j := 0; j < 5; j++ {
			slow.Exp(hi, slowtemp, order)
		}
		endTime = time.Now()
		avaslow = float64(endTime.Sub(beginTime).Nanoseconds())
		avaslow = avaslow / 5.0

		fmt.Println("i:", i, "\nh:", hi, "\nt/256:", new(big.Int).Rsh(t, 8), "\navalfast:", avafast, "\navalslow:", avaslow, "\navalslow/avalfast:", avaslow/avafast)
		fmt.Println("eij's byte:", fast.BitLen())
		if fast.Cmp(slow) != 0 {
			fmt.Println("failed")
		}
	}
}
