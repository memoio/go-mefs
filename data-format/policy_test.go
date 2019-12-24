package dataformat

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"

	"github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/crypto/aes"
)

// 全局配置
var skbyte = []byte{179, 233, 48, 97, 94, 148, 140, 7, 78, 102, 169, 48, 136, 124, 152, 101, 76, 69, 210, 14, 38, 15, 176, 227, 73, 41, 135, 17, 170, 138, 242, 69}
var buckid = 1

// 因只考虑生成3+2个stripe，故测试Rs时，文件长度不超过3M；测试Mul时，文件长度不超过1M
var Rslen = 2 * 1024 * 1024
var Mullen = 1 * 1024 * 1024
var opt = &DataCoder{
	DataCount:   3,
	ParityCount: 2,
	TagFlag:     BLS12,
	SegmentSize: DefaultSegmentSize,
}

func UploadRspolicy(data []byte) ([][]byte, error) {
	opt.Policy = RsPolicy
	// 构建加密秘钥
	tmpkey := append(skbyte, byte(buckid))
	skey := sha256.Sum256(tmpkey)
	// 加密、Encode
	data, err := aes.AesEncrypt(data, skey[:])
	if err != nil {
		return nil, err
	}
	encodeData, _, err := opt.Encode(data, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		return nil, err
	}
	return encodeData, nil
}

func UploadMulpolicy(data []byte) ([][]byte, error) {
	opt.Policy = MulPolicy
	// 构建加密秘钥
	tmpkey := append(skbyte, byte(buckid))
	skey := sha256.Sum256(tmpkey)
	// 加密、Encode
	data, err := aes.AesEncrypt(data, skey[:])
	if err != nil {
		return nil, err
	}
	encodeData, _, err := opt.Encode(data, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		return nil, err
	}
	return encodeData, nil
}

func Download(data [][]byte) ([]byte, error) {
	tmpkey := append(skbyte, byte(buckid))
	skey := sha256.Sum256(tmpkey)
	var filedata []byte
	var err error
	switch opt.Policy {
	case RsPolicy:
		filedata, err = GetFileDataFromSripe(data, int(opt.DataCount), 0, -1)
		if err != nil {
			fmt.Println("error is : ", err)
		}
	case MulPolicy:
		filedata, err = GetSegsFromData(data[0], 0, -1)
		if err != nil {
			fmt.Println("error is : ", err)
		}
	default:
		return nil, ErrWrongPolicy
	}
	filedata, err = aes.AesDecrypt(filedata, skey[:])
	if err != nil {
		return nil, err
	}
	return filedata, nil
}
func TestEncode(t *testing.T) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)

	// 多副本含前缀
	opt.Policy = MulPolicy
	stripe, offset, err := opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	fmt.Println(len(stripe))

	// 多副本不含前缀
	stripe, offset, err = opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	fmt.Println(len(stripe))

	// RS含前缀
	opt.Policy = RsPolicy
	stripe, offset, err = opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	fmt.Println(len(stripe))

	// RS不含前缀
	stripe, offset, err = opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	fmt.Println(len(stripe))
}

func TestRepair(t *testing.T) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	tmpData := make([]byte, 150)
	rand.Seed(0)
	fillRandom(tmpData)
	// 副本修复
	opt.Policy = MulPolicy
	stripe, offset, err := opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	fmt.Println(len(stripe))
	newStripe, err := Repair(stripe)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(newStripe))
	newStripe[1] = nil
	newStripe[3] = nil

	newStripe, err = Repair(newStripe)
	if err != nil {
		log.Fatal(err)
	}
	for i := range newStripe {
		if !bytes.Equal(newStripe[i], stripe[i]) {
			t.Error("error")
		}
	}
	fmt.Println(len(newStripe))

	// RS修复
	opt.Policy = RsPolicy
	stripe, offset, err = opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	fmt.Println(len(stripe))
	newStripe, err = Repair(stripe)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(newStripe))
	newStripe[1] = nil
	newStripe[3] = nil

	newStripe, err = Repair(newStripe)
	if err != nil {
		log.Fatal(err)
	}
	for i := range newStripe {
		if !bytes.Equal(newStripe[i], stripe[i]) {
			t.Error("error")
		}
	}
	fmt.Println(len(newStripe))
	secStripe := creatGroup(3, uint64(len(newStripe[0])))
	for i := 0; i < 3; i++ {
		copy(secStripe[i], newStripe[i])
	}
	secStripe, err = Repair(secStripe)
	if err != nil {
		log.Fatal(err)
	}
	for i := range secStripe {
		if !bytes.Equal(secStripe[i], stripe[i]) {
			t.Error("error")
		}
	}
	fmt.Println(len(stripe))
}

func TestRepairCorrect(t *testing.T) {
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	tmpData := make([]byte, 2*1024)
	rand.Seed(0)
	fillRandom(tmpData)
	Md5 := md5.Sum(tmpData)
	// 副本检验
	opt.Policy = MulPolicy
	stripe, offset, err := opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	stripe[4] = nil
	stripe[1] = nil
	newStripe, err := Repair(stripe)
	if err != nil {
		log.Fatal(err)
	}
	// 恢复后的数据验证
	newData, err := GetSegsFromData(newStripe[0], 0, -1)
	if err != nil {
		log.Fatal(err)
	}
	newMd5 := md5.Sum(newData[:2*1024])
	for i := range newMd5 {
		if Md5[i] != newMd5[i] {
			t.Error("error")
		}
	}
	// tag与data对应验证
	for i := range stripe {
		prefix, data, err := PrefixDecode(newStripe[i])
		if err != nil {
			log.Fatal(err)
		}
		segment, tag, _, err := prefix.GetSegAndTagFromRawdata(data, 0)
		if err != nil {
			log.Fatal(err)
		}
		newTag, err := GenTagForSegment(segment, []byte("8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_0"), uint64(opt.TagFlag), opt.SegmentSize, opt.KeySet)
		if err != nil {
			log.Fatal(err)
		}
		boo := mcl.VerifyTag(segment, newTag, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_0", opt.KeySet.Pk)
		if boo == false {
			t.Error("error")
		}
		boo = mcl.VerifyTag(segment, tag, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_0", opt.KeySet.Pk)
		if boo == false {
			t.Error("error")
		}
	}

	// RS检验
	opt.Policy = RsPolicy
	stripe, offset, err = opt.Encode(tmpData, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0", 0)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(offset)
	stripe[2] = nil
	stripe[3] = nil
	newStripe, err = Repair(stripe)
	if err != nil {
		log.Fatal(err)
	}
	// 恢复后的数据验证
	newData, err = GetFileDataFromSripe(newStripe, int(opt.DataCount), 0, -1)
	if err != nil {
		log.Fatal(err)
	}
	newMd5 = md5.Sum(newData[:2*1024])
	for i := range newMd5 {
		if Md5[i] != newMd5[i] {
			t.Error("error")
		}
	}
	// tag与data对应验证
	for i := range stripe {
		prefix, data, err := PrefixDecode(newStripe[i])
		if err != nil {
			log.Fatal(err)
		}
		segment, tag, _, err := prefix.GetSegAndTagFromRawdata(data, 0)
		if err != nil {
			log.Fatal(err)
		}
		newTag, err := GenTagForSegment(segment, []byte("8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_0"), uint64(opt.TagFlag), opt.SegmentSize, opt.KeySet)
		if err != nil {
			log.Fatal(err)
		}
		boo := mcl.VerifyTag(segment, newTag, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_0", opt.KeySet.Pk)
		if boo == false {
			t.Error("error")
		}
		boo = mcl.VerifyTag(segment, tag, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_0", opt.KeySet.Pk)
		if boo == false {
			t.Error("error")
		}
	}
}

func TestUploadRspolicy(t *testing.T) {
	// 1MB文件
	tmpData := make([]byte, Rslen)
	rand.Seed(0)
	fillRandom(tmpData)
	// 配置部分
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	data, err := UploadRspolicy(tmpData)
	if err != nil {
		log.Fatal(err)
	}
	// 将编码的数据存入指定目录
	dir := path.Join(os.Getenv("HOME"), "encodetest")
	if err := os.MkdirAll(dir, 0777); err != nil && !os.IsExist(err) {
		fmt.Println("error is : ", err)
	}
	for i := 0; i < int(opt.DataCount+opt.ParityCount); i++ {
		tmpdir := path.Join(dir, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_Rspolicy")
		f, err := os.Create(tmpdir)
		if err != nil {
			fmt.Println("error is : ", err)
		}
		_, err = f.Write(data[i])
		if err != nil {
			fmt.Println("error is : ", err)
		}
		err = f.Close()
		if err != nil {
			fmt.Println("error is : ", err)
		}
	}
}

func TestDownloadRspolicy(t *testing.T) {
	// 读取文件夹下文件
	data := make([][]byte, opt.DataCount+opt.ParityCount)
	dir := path.Join(os.Getenv("HOME"), "encodetest")
	for i := 0; i < int(opt.DataCount+opt.ParityCount); i++ {
		f, err := os.Open(path.Join(dir, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)) + "_Rspolicy")
		if err != nil {
			log.Fatal(err)
		}
		finfo, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		tmpdata := make([]byte, finfo.Size())
		_, err = f.Read(tmpdata)
		if err != nil {
			fmt.Println("error is : ", err)
		}
		data[i] = tmpdata
	}
	opt.Policy = RsPolicy
	filedata, err := Download(data)
	if err != nil {
		log.Fatal(err)
	}
	// 数据比对
	rawData := make([]byte, Rslen)
	rand.Seed(0)
	fillRandom(rawData)
	rawmd5 := md5.Sum(rawData)
	Datamd5 := md5.Sum(filedata[:Rslen])
	for i := range rawmd5 {
		if rawmd5[i] != Datamd5[i] {
			log.Fatal("md5 mismatch")
		}
	}
}

func TestUploadMulpolicy(t *testing.T) {
	tmpData := make([]byte, Mullen)
	rand.Seed(0)
	fillRandom(tmpData)
	// 配置部分
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	data, err := UploadMulpolicy(tmpData)
	if err != nil {
		log.Fatal(err)
	}
	// 将编码的数据存入指定目录
	dir := path.Join(os.Getenv("HOME"), "encodetest")
	if err := os.MkdirAll(dir, 0777); err != nil && !os.IsExist(err) {
		fmt.Println("error is : ", err)
	}
	for i := 0; i < int(opt.DataCount+opt.ParityCount); i++ {
		tmpdir := path.Join(dir, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)+"_Mulpolicy")
		f, err := os.Create(tmpdir)
		if err != nil {
			fmt.Println("error is : ", err)
		}
		_, err = f.Write(data[i])
		if err != nil {
			log.Fatal(err)
		}
		err = f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestDownloadMulpolicy(t *testing.T) {
	// 读取文件夹下文件
	data := make([][]byte, opt.DataCount+opt.ParityCount)
	dir := path.Join(os.Getenv("HOME"), "encodetest")
	for i := 0; i < int(opt.DataCount+opt.ParityCount); i++ {
		f, err := os.Open(path.Join(dir, "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0"+"_"+strconv.Itoa(i)) + "_Mulpolicy")
		if err != nil {
			log.Fatal(err)
		}
		finfo, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		tmpdata := make([]byte, finfo.Size())
		_, err = f.Read(tmpdata)
		if err != nil {
			log.Fatal(err)
		}
		data[i] = tmpdata
	}
	opt.Policy = MulPolicy
	filedata, err := Download(data)
	if err != nil {
		log.Fatal(err)
	}
	// 数据比对
	rawData := make([]byte, Mullen)
	rand.Seed(0)
	fillRandom(rawData)
	rawmd5 := md5.Sum(rawData)
	datamd5 := md5.Sum(filedata[:Mullen])
	for i := range rawmd5 {
		if rawmd5[i] != datamd5[i] {
			log.Fatal("md5 mismatch")
		}
	}
}

func BenchmarkUploadRspolicy(b *testing.B) {
	rawData := make([]byte, Rslen)
	rand.Seed(0)
	fillRandom(rawData)
	// 配置部分
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			_, err := UploadRspolicy(rawData)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func BenchmarkUploadMulpolicy(b *testing.B) {
	rawData := make([]byte, Mullen)
	rand.Seed(0)
	fillRandom(rawData)
	// 配置部分
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func() {
			_, err := UploadMulpolicy(rawData)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func BenchmarkDownloadRspolicy(b *testing.B) {
	rawData := make([]byte, Rslen)
	rand.Seed(0)
	fillRandom(rawData)
	// 配置部分
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	data, err := UploadRspolicy(rawData)
	if err != nil {
		log.Fatal(err)
	}
	// decode、解密
	for i := 0; i < b.N; i++ {
		func() {
			_, err := Download(data)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}

func BenchmarkDownloadMulpolicy(b *testing.B) {
	rawData := make([]byte, Mullen)
	rand.Seed(0)
	fillRandom(rawData)
	// 配置部分
	err := mcl.Init(mcl.BLS12_381)
	if err != nil {
		log.Fatal(err)
	}
	keyset, _ := mcl.GenKeySet()
	opt.KeySet = keyset
	data, err := UploadMulpolicy(rawData)
	if err != nil {
		log.Fatal(err)
	}
	// decode、解密
	for i := 0; i < b.N; i++ {
		func() {
			_, err := Download(data)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}
}
