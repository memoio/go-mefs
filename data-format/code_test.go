package dataformat

import (
	"crypto/md5"
	"crypto/sha256"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/crypto/aes"
)

// 全局配置
var skbyte = []byte{179, 233, 48, 97, 94, 148, 140, 7, 78, 102, 169, 48, 136, 124, 152, 101, 76, 69, 210, 14, 38, 15, 176, 227, 73, 41, 135, 17, 170, 138, 242, 69}
var buckid = 1

// 因只考虑生成3+2个stripe，故测试Rs时，文件长度不超过3M；测试Mul时，文件长度不超过1M
var Rslen = 2 * 1024 * 1024
var Mullen = 1 * 1024 * 1024

var userID = "8MGxCuiT75bje883b7uFb6eMrJt5cP"

func BenchmarkEncode(b *testing.B) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, err := mcl.GenKeySet()
	if err != nil {
		log.Fatal(err)
	}

	opt, err := NewDataCoderWithDefault(keyset, RsPolicy, 3, 2, userID, userID)
	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 4096*3)
	fillRandom(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 构建加密秘钥icy
		tmpkey := append(skbyte, byte(buckid))
		skey := sha256.Sum256(tmpkey)

		// 加密、Encode
		data, _ = aes.AesEncrypt(data, skey[:])

		// 多副本含前缀
		opt.Encode(data, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	}
}

func CodeAndRepair(policy, dc, pc, size int) {
	keyset, err := mcl.GenKeySet()
	if err != nil {
		log.Fatal(err)
	}

	opt, err := NewDataCoderWithDefault(keyset, policy, dc, pc, userID, userID)
	if err != nil {
		log.Fatal(err)
	}

	dlen := size

	data := make([]byte, dlen)
	fillRandom(data)

	oldMd5 := md5.Sum(data)
	log.Println("data md5 is: ", oldMd5)

	tmpkey := append(skbyte, byte(buckid))
	skey := sha256.Sum256(tmpkey)
	// 加密、Encode

	if len(data)%aes.BlockSize != 0 {
		data = aes.PKCS5Padding(data)
	}

	data, _ = aes.AesEncrypt(data, skey[:])
	// 多副本含前缀
	datas, offset, err := opt.Encode(data, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("encode offset is: ", offset)

	gotdata, err := opt.Decode(datas, 0, dlen)
	if err != nil {
		log.Fatal(err)
	}

	gotdata, _ = aes.AesDecrypt(gotdata, skey[:])

	gotMd5 := md5.Sum(gotdata[:dlen])
	log.Println("data md5 is: ", gotMd5)
	if gotMd5 != oldMd5 {
		log.Fatal("Md5 is not right after decode")
	}

	log.Println("decode success")

	datas[0] = nil
	opt.Repair = true

	datas, len, err := Repair(datas)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("repair offset is: ", len)

	for i := 0; i < opt.blockCount; i++ {
		bid := "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0_" + strconv.Itoa(i)
		_, _, _, ok := opt.VerifyBlock(datas[i], bid)
		if !ok {
			log.Fatal("tag is wrong for: ", bid)
		}
	}

	gotdata, err = opt.Decode(datas, 0, dlen)
	if err != nil {
		log.Fatal(err)
	}

	gotdata, _ = aes.AesDecrypt(gotdata, skey[:])

	gotMd5 = md5.Sum(gotdata[:dlen])
	log.Println("data md5 is: ", gotMd5)
	if gotMd5 != oldMd5 {
		log.Fatal("Md5 is not right after decode 2")
	}

	log.Println("repair one success")

	datas[0] = nil
	datas[dc] = nil

	datas, len, err = Repair(datas)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("repair offset is: ", len)

	for i := 0; i < opt.blockCount; i++ {
		bid := "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0_" + strconv.Itoa(i)
		ok := VerifyBlock(datas[i], bid, keyset)
		if !ok {
			log.Fatal("tag is wrong for: ", bid)
		}
	}

	gotdata, err = opt.Decode(datas, 0, dlen)
	if err != nil {
		log.Fatal(err)
	}

	gotdata, _ = aes.AesDecrypt(gotdata, skey[:])

	gotMd5 = md5.Sum(gotdata[:dlen])
	log.Println("data md5 is: ", gotMd5)
	if gotMd5 != oldMd5 {
		log.Fatal("Md5 is not right after repair")
	}
	log.Println("repair two success")

	if policy == RsPolicy {
		datas[0] = nil
		datas[dc] = nil

		opt.Repair = true

		log.Println("repair offset is: ", len)

		gotdata, err = opt.Decode(datas, 0, dlen)
		if err != nil {
			log.Fatal(err)
		}

		gotdata, _ = aes.AesDecrypt(gotdata, skey[:])

		gotMd5 = md5.Sum(gotdata[:dlen])
		log.Println("data md5 is: ", gotMd5)
		if gotMd5 != oldMd5 {
			log.Fatal("Md5 is not right after repair")
		}
		log.Println("rs decode success")
	}
}

func TestCode(t *testing.T) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}

	size := []int{4095, 4096, 4097, 3*4096 - 1, 3 * 4096, 3*4096 + 1}
	for _, s := range size {
		log.Println("test size: ", s)
		CodeAndRepair(RsPolicy, 3, 2, s)
		CodeAndRepair(MulPolicy, 3, 2, s)
	}

	return
}

func fillRandom(p []byte) {
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}
