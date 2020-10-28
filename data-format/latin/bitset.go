package latin

// var length uint32 = 65536        //2 ^ 16
// const csize_n uint64 = 268435456 // 2^32 / 16
// var coord [2][4294967296]uint32  //坐标数组
// var filecontent1 [64 * 1024 * 1024]uint64
// var filecontent2 [64 * 1024 * 1024]uint64

func Encode(origin []byte, coord [][]uint32, n int) ([]byte, error) {
	encoded := make([]byte, len(origin))
	var length uint32 = 1 << uint(n)
	var csize_n uint64 = 1 << uint(n*2)
	var k1, k2 uint32
	var j uint32
	var i uint64
	var temp byte
	var c, r byte
	for i = 0; i < csize_n; i++ {
		j = coord[0][i]*length + coord[1][i]
		k1 = j >> 3
		k2 = j & 0x7
		temp = (origin[k1] >> (7 - k2)) & 0x1 //每次取出1bit
		c = (c << 1) | temp                   //依次存入1bit
		r++
		if r == 8 { //每8个字节写入一次
			//fwrite(&c2, 1, 1, fp4);
			encoded[i>>3] = c
			c = 0
			r = 0
		}
	}
	return encoded, nil
}

func Decode(encoded []byte, coord [][]uint32, n int) ([]byte, error) {
	origin := make([]byte, len(encoded))
	var length uint32 = 1 << uint(n)
	var csize_n uint64 = 1 << (uint(n) * 2)
	var i, j uint64
	var k1, k2 uint64
	var c, temp byte
	var r uint64 = 8
	// var p1, p2 uint32
	for i = 0; i < csize_n; i++ {
		if r == 8 {
			c = encoded[i>>3]
			r = 0
		}
		temp = (c >> (7 - r)) & 0x1
		r++
		//temp = c & kx[r];
		j = uint64(coord[0][i]*length + coord[1][i])
		k1 = j >> 3
		k2 = j & 0x7
		//filecontent1[k1] |= (temp == 0? 0:kx[k2]);
		origin[k1] = origin[k1] | (temp << (7 - k2))
	}
	return origin, nil
}
