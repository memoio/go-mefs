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
const randomDataSize = 1024 * 1024 * 10
const dataCount = 2
const parityCount = 3
const moneyTo = 1000000000000000000

var ethEndPoint string

func main() {
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address")
	flag.Parse()
	ethEndPoint = *eth

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

	flag := true
	if flag {
		err := test.TransferTo(big.NewInt(moneyTo), addr, ethEndPoint, ethEndPoint)
		if err != nil {
			log.Println("trnasfer fails: ", err)
			return err
		}
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

	log.Println("3. test rs create bucket")

	bucketName := time.Now().Format("2006-01-02")
	var opts []func(*shell.RequestBuilder) error
	//set option of bucket
	opts = append(opts, shell.SetAddress(addr))
	opts = append(opts, shell.SetDataCount(dataCount))
	opts = append(opts, shell.SetParityCount(parityCount))
	opts = append(opts, shell.SetPolicy(df.RsPolicy))
	_, err = sh.CreateBucket(bucketName, opts...)
	if err != nil {
		log.Println("create bucket err: ", err)
		return err
	}

	bk, err := sh.HeadBucket(bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("create bucket err: ", err)
		return err
	}

	log.Println(bk.Buckets)

	if bk.Buckets[0].Name != bucketName {
		log.Fatal("create bucket", bucketName, "fails, but got:", bk.Buckets[0].Name)
	}

	if bk.Buckets[0].Policy != df.RsPolicy {
		log.Fatal("create bucket fails Policy")
	}

	if bk.Buckets[0].DataCount != dataCount {
		log.Fatal("create bucket fails datacount")
	}

	if bk.Buckets[0].ParityCount != parityCount {
		log.Fatal("create bucket fails paritycount")
	}

	log.Println("4. test rs put bucket")

	fileNum := 1

	//upload file
	r := rand.Int63n(randomDataSize)
	data := make([]byte, r)
	fillRandom(data)
	buf := bytes.NewBuffer(data)
	objectName := addr + "_" + strconv.Itoa(fileNum)
	log.Println("Begin to upload file", fileNum, "，Filename is", objectName, "Size is", ToStorageSize(r), "addr", addr)
	uploadBeginTime := time.Now().Unix()
	ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("put object fails")
		return err
	}

	storagekb := float64(r) / 1024.0
	uploadEndTime := time.Now().Unix()
	speed := fmt.Sprintf("%.2f", storagekb/float64(uploadEndTime-uploadBeginTime))
	log.Println(" Upload file", fileNum, "success，Filename is", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", "addr", addr)
	log.Println(ob.String() + "address: " + addr)

	obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("head file ", objectName, " err:", err)
		return err
	}

	if obj.Objects[0].Name != objectName {
		log.Println("head file ", objectName, "but got: ", obj.Objects[0].Name)
	}

	if int64(obj.Objects[0].Size) != r {
		log.Println("head file siez: ", r, "but got: ", obj.Objects[0].Size)
	}

	log.Println("5. test rs get object")

	outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("download file ", objectName, " err:", err)
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

	log.Println("6. test mul create bucket")

	mbucketName := "mtest" + time.Now().Format("2006-01-02")
	var mopts []func(*shell.RequestBuilder) error
	mopts = append(mopts, shell.SetAddress(addr))
	mopts = append(mopts, shell.SetDataCount(dataCount))
	mopts = append(mopts, shell.SetParityCount(parityCount))
	mopts = append(mopts, shell.SetPolicy(df.MulPolicy))

	_, errBucket := sh.CreateBucket(mbucketName, mopts...)
	if errBucket != nil {
		log.Println("create mbucket err: ", err)
	}

	bk, err = sh.HeadBucket(mbucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("create mbucket err: ", err)
	}

	log.Println(bk.Buckets)

	if bk.Buckets[0].Name != mbucketName {
		log.Fatal("create mbucket", mbucketName, "fails")
	}

	if bk.Buckets[0].Policy != df.MulPolicy {
		log.Fatal("create mbucket fails Policy")
	}

	if bk.Buckets[0].DataCount != 1 {
		log.Fatal("create mbucket fails datacount")
	}

	if bk.Buckets[0].ParityCount != parityCount+dataCount-1 {
		log.Fatal("create mbucket fails paritycount")
	}

	log.Println("7. test mul put object")

	fileNum = 20

	r1 := rand.Int63n(randomDataSize)
	data = make([]byte, r1)
	fillRandom(data)
	buf = bytes.NewBuffer(data)
	objectName = addr + "_" + strconv.Itoa(fileNum)
	log.Println("Begin to upload file", fileNum, "，Filename is", objectName, "Size is", ToStorageSize(r1), "addr", addr)
	uploadBeginTime = time.Now().Unix()
	ob, err = sh.PutObject(buf, objectName, mbucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("put object fails")
	} else {
		storagekb := float64(r1) / 1024.0
		uploadEndTime := time.Now().Unix()
		speed := fmt.Sprintf("%.2f", storagekb/float64(uploadEndTime-uploadBeginTime))
		log.Println(" Upload file", fileNum, "success，Filename is", objectName, "Size is", ToStorageSize(r1), "speed is", speed, "KB/s", "addr", addr)
		log.Println(ob.String() + "address: " + addr)
	}

	obj, err = sh.HeadObject(objectName, mbucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("head file ", objectName, " err:", err)
		return err
	}

	if obj.Objects[0].Name != objectName {
		log.Fatal("head file ", objectName, "but got: ", obj.Objects[0].Name)
	}

	if int64(obj.Objects[0].Size) != r1 {
		log.Fatal("head file siez: ", r1, "but got: ", obj.Objects[0].Size)
	}

	log.Println("8. test mul get object")

	outer, err = sh.GetObject(objectName, mbucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("download file ", objectName, " err:", err)
		return err
	}

	obuf = new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(obj.Objects[0].Size) {
		log.Println("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", obj.Objects[0].Size)
	}

	gotTag = md5.Sum(obuf.Bytes())
	if hex.EncodeToString(gotTag[:]) != obj.Objects[0].MD5 {
		log.Fatal("download mul file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", obj.Objects[0].MD5)
	}

	log.Println("9. test showstorage")

	res, err := sh.ShowStorage(shell.SetAddress(addr))
	if err != nil {
		log.Println("show storage err : ", err)
		return err
	}

	log.Println("storage: ", res)
	log.Println("upload: ", r+r1)
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
