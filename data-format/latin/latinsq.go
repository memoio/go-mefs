package latin

const n uint64 = 1 << 16

// var F [65536]uint32

//unsigned short A[n][n*n];
func xor(a, b uint64) uint64 {
	return a ^ b
}

//n=8:x & 0x0004) ? 0x0b : 0x0000
//n=16:x & 0x0008) ? 0x13 : 0x0000
//n=32:x & 0x0010) ? 0x25 : 0x0000
//n=64:x & 0x0020) ? 0x43 : 0x0000
//n=128:x & 0x0040) ? 0x49 : 0x0000
//n=256:x & 0x0080) ? 0x011d : 0x0000
//n=512:x & 0x0100) ? 0x0211 : 0x0000
//n=1024:x & 0x0200) ? 0x0409 : 0x0000
//n=65536:x & 0x8000) ? 0x100b : 0x0000
func XTIME(x uint32, n int) uint32 {
	var temp uint32
	switch n {
	case 16:
		if (x & 0x8000) != 0 {
			temp = 0x100b
		} else {
			temp = 0x0000
		}
	case 10:
		if (x & 0x0200) != 0 {
			temp = 0x0409
		} else {
			temp = 0x0000
		}
	case 9:
		if (x & 0x0100) != 0 {
			temp = 0x0211
		} else {
			temp = 0x0000
		}
	case 8:
		if (x & 0x0080) != 0 {
			temp = 0x011d
		} else {
			temp = 0x0000
		}
	case 7:
		if (x & 0x0040) != 0 {
			temp = 0x49
		} else {
			temp = 0x0000
		}
	case 6:
		if (x & 0x0020) != 0 {
			temp = 0x43
		} else {
			temp = 0x0000
		}
	case 5:
		if (x & 0x0010) != 0 {
			temp = 0x25
		} else {
			temp = 0x0000
		}
	case 4:
		if (x & 0x0008) != 0 {
			temp = 0x13
		} else {
			temp = 0x0000
		}
	case 3:
		if (x & 0x0004) != 0 {
			temp = 0x0b
		} else {
			temp = 0x0000
		}
	}

	return (x << 1) ^ temp
}

func mul(a, b uint32, n int) uint32 {
	var c uint32 = 0x0000
	var temp = make([]uint32, n)
	temp[0] = a

	var i int
	for i = 1; i < n; i++ {
		temp[i] = XTIME(temp[i-1], n)
	}
	c = (b & 0x0001) * a
	for i = 1; i < n; i++ {
		c ^= (((b >> uint(i)) & 0x0001) * temp[i])
	}
	return c
}

//[65536]uint32
func Latin(n, count int) ([][]uint32, error) {
	var length uint32 = 1 << uint(n)
	latin := make([]uint32, length)

	var i uint32
	for i = 0; i < length; i++ {
		latin[i] = (i + 1) % length
	}

	res := make([][]uint32, count)
	for i := 0; i < count; i++ {
		res[i] = make([]uint32, length*length)
	}
	//否则n为2的幂时
	for i := 1; i <= count; i++ { //输出第1个和第2个拉丁方，理论上一共可以生成n-1个拉丁方
		var j uint32
		for j = 0; j < length*length; j++ {
			temp := mul(latin[i-1], latin[j/length], n)
			res[i-1][j] = (temp ^ latin[j%length]) % length
		}
	}
	return res, nil
}
