package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"time"

	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
	shell "github.com/memoio/mefs-go-http-client"
)

//随机文件最大大小
const (
	randomDataSize = 1024 * 1024 * 10
	moneyTo        = 6000000000000000000
)

var ethEndPoint, qethEndPoint string

func main() {
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address for set;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address for query;")

	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	err := lfsTest()
	if err != nil {
		log.Fatal(err)
	}
}

func lfsTest() error {
	sh := shell.NewShell("localhost:5001")

	log.Println("1. test create user")

	testuser, err := sh.CreateUser()
	if err != nil {
		log.Println("Create user failed :", err)
		return err
	}

	addr := testuser.Address
	_, err = address.GetIDFromAddress(addr)
	if err != nil {
		log.Println("address to id failed")
		return err
	}

	log.Println("create user sk: ", testuser.Sk, ",addr: ", testuser.Address)

	err = test.TransferTo(big.NewInt(moneyTo), addr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Println("trnasfer fails: ", err)
		return err
	}

	log.Println("2. test start lfs")
	var startOpts []func(*shell.RequestBuilder) error
	//set option of bucket
	startOpts = append(startOpts, shell.SetOp("ks", "3"))
	startOpts = append(startOpts, shell.SetOp("ps", "6"))
	err = sh.StartUser(addr, startOpts...)
	if err != nil {
		log.Println("Start user failed :", err)
		return err
	}

	rand.Seed(time.Now().UnixNano())

	log.Println("3. test rs encrypt bucket")
	b1 := "enc-" + time.Now().Format("2006-01-02")
	r1 := rand.Int63n(randomDataSize)
	err = testPut(sh, b1, addr, 2, 3, df.RsPolicy, r1, true)
	if err != nil {
		return err
	}

	log.Println("4. test rs de encrypt bucket")
	b2 := "de-" + time.Now().Format("2006-01-02")
	r2 := rand.Int63n(randomDataSize)
	err = testPut(sh, b2, addr, 2, 3, df.RsPolicy, r2, false)
	if err != nil {
		return err
	}

	log.Println("5. test mul bucket")

	b3 := "mul-enc-" + time.Now().Format("2006-01-02")
	r3 := rand.Int63n(randomDataSize)
	err = testPut(sh, b3, addr, 1, 4, df.MulPolicy, r3, true)
	if err != nil {
		return err
	}

	b4 := "mul-de-" + time.Now().Format("2006-01-02")
	r4 := rand.Int63n(randomDataSize)
	err = testPut(sh, b4, addr, 1, 4, df.MulPolicy, r4, false)
	if err != nil {
		return err
	}

	log.Println("6. test showstorage")
	res, err := sh.ShowStorage(shell.SetAddress(addr))
	if err != nil {
		log.Fatal("show storage err : ", err)
		return err
	}

	log.Println("upload: ", r1+r2+r3+r4)
	log.Println("storage: ", res)
	return nil
}

func testPut(sh *shell.Shell, bucketName, addr string, dataCount, parityCount, policy int, length int64, crypto bool) error {
	log.Println("test create bucket")
	//set option of bucket
	var opts []func(*shell.RequestBuilder) error
	opts = append(opts, shell.SetAddress(addr))
	opts = append(opts, shell.SetDataCount(dataCount))
	opts = append(opts, shell.SetParityCount(parityCount))
	opts = append(opts, shell.SetPolicy(policy))
	opts = append(opts, shell.SetCrypto(crypto))
	_, err := sh.CreateBucket(bucketName, opts...)
	if err != nil {
		log.Fatal("create bucket err: ", err)
		return err
	}

	bk, err := sh.HeadBucket(bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("head bucket err: ", err)
		return err
	}

	log.Println(bk.Buckets)

	if bk.Buckets[0].Name != bucketName {
		log.Fatal("create bucket", bucketName, "fails, but got:", bk.Buckets[0].Name)
	}

	if bk.Buckets[0].Policy != int32(policy) {
		log.Fatal("create bucket fails policy")
	}

	if bk.Buckets[0].DataCount != int32(dataCount) {
		log.Fatal("create bucket fails datacount")
	}

	if bk.Buckets[0].ParityCount != int32(parityCount) {
		log.Fatal("create bucket fails paritycount")
	}

	if crypto {
		if bk.Buckets[0].Encryption == 0 {
			log.Fatal("create bucket fails to encrypt")
		}
	} else {
		if bk.Buckets[0].Encryption == 1 {
			log.Fatal("create bucket fails to no encrypt")
		}
	}

	log.Println("test put object")

	fileNum := 1

	//upload file
	data := make([]byte, length)
	fillRandom(data)
	buf := bytes.NewBuffer(data)
	objectName := addr + "_" + strconv.Itoa(fileNum)
	log.Println("Begin to upload file", fileNum, "，Filename is", objectName, "Size is", ToStorageSize(length), "addr", addr)
	uploadBeginTime := time.Now().Unix()
	_, err = sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("put object fails")
		return err
	}

	storagekb := float64(length) / 1024.0
	uploadEndTime := time.Now().Unix()
	speed := fmt.Sprintf("%.2f", storagekb/float64(uploadEndTime-uploadBeginTime))
	log.Println("Upload file", fileNum, "success，Filename is", objectName, "Size is", ToStorageSize(length), "speed is", speed, "KB/s", "addr", addr)

	log.Println("test head object")
	obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("head file ", objectName, " err:", err)
		return err
	}

	if obj.Objects[0].Name != objectName {
		log.Fatal("head file ", objectName, "but got: ", obj.Objects[0].Name)
	}

	if obj.Objects[0].Size != int64(length) {
		log.Fatal("head file size: ", length, "but got: ", obj.Objects[0].Size)
	}

	log.Println("test get object")
	outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("download file ", objectName, " err:", err)
		return err
	}

	obuf := new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(obj.Objects[0].Size) {
		log.Fatal("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", obj.Objects[0].Size)
	}

	gotTag := md5.Sum(obuf.Bytes())
	if hex.EncodeToString(gotTag[:]) != obj.Objects[0].MD5 {
		log.Fatal("download file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", obj.Objects[0].MD5)
	}
	return nil
}

func ToStorageSize(r int64) string {
	FloatStorage := float64(r)
	var OutStorage string
	if FloatStorage < 1024 && FloatStorage >= 0 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage) + "B"
	} else if FloatStorage < 1048576 && FloatStorage >= 1024 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1024) + "KB"
	} else if FloatStorage < 1073741824 && FloatStorage >= 1048576 {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1048576) + "MB"
	} else {
		OutStorage = fmt.Sprintf("%.2f", FloatStorage/1073741824) + "GB"
	}
	return OutStorage
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
