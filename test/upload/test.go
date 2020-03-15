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
	"sync"
	"time"

	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
	shell "github.com/memoio/mefs-go-http-client"
)

//随机文件最大大小
const randomDataSize = 1024 * 1024 * 10
const dataCount = 3
const parityCount = 2

const moneyTo = 1000000000000000000

var objsInBucket sync.Map

var ethEndPoint string

func main() {
	count := flag.Int("count", 100, "count of file we want to upload")
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address")
	flag.Parse()
	ethEndPoint = *eth
	err := uploadTest(*count)
	if err != nil {
		log.Fatal(err)
	}
}

func uploadTest(count int) error {
	sh := shell.NewShell("localhost:5001")
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

	err = test.TransferTo(big.NewInt(moneyTo), addr, ethEndPoint, ethEndPoint)
	if err != nil {
		log.Println("transfer fails: ", err)
		return err
	}

	err = sh.StartUser(addr)
	if err != nil {
		log.Println("Start user failed :", err)
		return err
	}

	log.Println(addr, "started, begin to upload")

	rand.Seed(time.Now().Unix())
	bucketName := time.Now().Format("2006-01-02")
	_, err = sh.ShowStorage(shell.SetAddress(addr))
	if err != nil {
		log.Println("show storage fails: ", err)
		return err
	}
	var opts []func(*shell.RequestBuilder) error
	//set option of bucket
	opts = append(opts, shell.SetAddress(addr))
	opts = append(opts, shell.SetDataCount(dataCount))
	opts = append(opts, shell.SetParityCount(parityCount))
	opts = append(opts, shell.SetPolicy(df.RsPolicy))
	bk, err := sh.CreateBucket(bucketName, opts...)
	if err != nil {
		log.Println("create bucket fails: ", err)
		return err
	}
	log.Println(bk, "addr:", addr)

	fileNum := 0
	fileUploadSuccessNum := 0

	rand.Seed(time.Now().Unix())
	//upload file
	for fileNum < count {
		fileNum++
		r := rand.Int63n(randomDataSize)
		data := make([]byte, r)
		fillRandom(data)
		buf := bytes.NewBuffer(data)
		objectName := addr + "_" + strconv.Itoa(fileNum)
		log.Println("Begin to upload file", fileNum, "，Filename is", objectName, "Size is", ToStorageSize(r), "addr", addr)
		uploadBeginTime := time.Now().Unix()
		ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println("put object fails:", err)
			continue
		}
		objsInBucket.Store(objectName, bucketName)

		storagekb := float64(r) / 1024.0
		uploadEndTime := time.Now().Unix()
		speed := fmt.Sprintf("%.2f", storagekb/float64(uploadEndTime-uploadBeginTime))
		log.Println("Upload file: ", fileNum, "success，name is:", objectName, "size is:", ToStorageSize(r), "speed is:", speed, " KB/s")
		log.Println(ob.String())
		fileUploadSuccessNum++
	}

	log.Println("Upload test complete")
	log.Println("uUpload ", fileNum, " files, ", fileUploadSuccessNum, " files success; rate is: ", fileUploadSuccessNum/count)

	//download file
	fileDownloadSuccessNum := 0

	objsInBucket.Range(func(k, v interface{}) bool {
		objectName := k.(string)
		bucketName := v.(string)
		outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println("download file ", objectName, " err:", err)
			return true
		}

		obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println("head file ", objectName, " err:", err)
			return true
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(outer)
		if buf.Len() != int(obj.Objects[0].Size) {
			log.Println("download file ", objectName, "failed, got: ", buf.Len(), "expected: ", obj.Objects[0].Size)
			return true
		}

		gotTag := md5.Sum(buf.Bytes())
		if hex.EncodeToString(gotTag[:]) != obj.Objects[0].MD5 {
			log.Println("download file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", obj.Objects[0].MD5)
			return true
		}

		log.Println("download file ", objectName, "success, got: ", buf.Len())
		fileDownloadSuccessNum++
		return true
	})

	log.Println("download test complete")
	log.Println("download ", fileDownloadSuccessNum, " files")

	log.Println("upload: ", fileNum, "; success:", fileUploadSuccessNum, " rate is", 100*fileUploadSuccessNum/count)
	log.Println("downlaod: ", fileNum, "; success:", fileDownloadSuccessNum, " rate is", 100*fileDownloadSuccessNum/count)

	if 100*fileUploadSuccessNum/count < 90 {
		log.Fatal("upload rate is too low")
	}

	if 100*fileDownloadSuccessNum/count < 90 {
		log.Fatal("download rate is too low")
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
