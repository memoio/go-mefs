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
	"time"

	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils"
	shell "github.com/memoio/mefs-go-http-client"
)

const (
	moneyTo        = 6000000000000000000
	randomDataSize = 1024 * 1024 * 10
)

var ethEndPoint, qethEndPoint string

func main() {
	utils.StartLogger()
	eth := flag.String("eth", "http://119.147.213.220:8193", "eth api address for set;")
	qeth := flag.String("qeth", "http://119.147.213.220:8196", "eth api address for query;")

	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	sh := shell.NewShell("localhost:5001")

	log.Println("======1. begin to create account======")
	testuser, err := sh.CreateUser()
	if err != nil {
		log.Fatal("Create user failed :", err)
	}
	addr := testuser.Address

	log.Println("======2. begin to transfer money======")
	err = test.TransferTo(big.NewInt(moneyTo), addr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Fatal("transfer failed ", err)
	}

	log.Println("======3. begin to start lfs======")
	var startOpts []func(*shell.RequestBuilder) error
	startOpts = append(startOpts, shell.SetOp("ks", "3"))
	startOpts = append(startOpts, shell.SetOp("ps", "6"))
	err = sh.StartUser(addr, startOpts...)
	if err != nil {
		log.Fatal("Start lfs failed :", err)
	}

	log.Println("======4. begin to creat bucket======")
	bucketName1 := "bucket1"
	var opts []func(*shell.RequestBuilder) error
	opts = append(opts, shell.SetAddress(addr))
	_, err = sh.CreateBucket(bucketName1, opts...)
	if err != nil {
		log.Fatal("create bucket err: ", err)
	}

	log.Println("======5. begin to head bucket======")
	bk, err := sh.HeadBucket(bucketName1, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("head bucket err: ", err)
	}
	log.Println(bk.Buckets)

	log.Println("======6. begin to put object======")
	rand.Seed(time.Now().UnixNano())
	length := rand.Int63n(randomDataSize)
	data := make([]byte, length)
	fillRandom(data)
	buf := bytes.NewBuffer(data)
	objectName := "a.txt"
	log.Println("Begin to upload file: ", objectName, "; addr: ", addr)
	_, err = sh.PutObject(buf, objectName, bucketName1, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("put object fails")
	}

	log.Println("======7. begin to get object======")
	outer, err := sh.GetObject(objectName, bucketName1, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("download file ", objectName, " err:", err)
	}
	obuf := new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(length) {
		log.Fatal("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", length)
	}
	tag := md5.Sum(data)
	gotTag := md5.Sum(obuf.Bytes())
	fmt.Println("tag is ", hex.EncodeToString(tag[:]), " gotTag is ", hex.EncodeToString(gotTag[:]))
	if hex.EncodeToString(gotTag[:]) != hex.EncodeToString(tag[:]) {
		log.Fatal("download file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", hex.EncodeToString(tag[:]))
	}

	log.Println("======8. begin to generate shareLink======")
	slink, err := sh.GenShare(objectName, bucketName1, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("generate link fails ", err)
	}
	log.Println("share link is ", slink)

	log.Println("======9. begin to kill user======")

	retry := 0
	for {
		retry++
		res, err := sh.Kill(addr)
		if err != nil {
			if retry > 5 {
				log.Fatal("kill user ", addr, " fails ", err)
			}
			time.Sleep(27 * time.Second)
			continue
		}
		log.Println(res.ChildLists[0])
		break
	}

	log.Println("======10. begin to create another account======")
	testuser2, err := sh.CreateUser()
	if err != nil {
		log.Fatal("Create user failed :", err)
	}
	addr2 := testuser2.Address

	log.Println("======11. begin to transfer money======")
	err = test.TransferTo(big.NewInt(moneyTo), addr2, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Fatal("transfer failed ", err)
	}

	log.Println("======12. begin to get file by shareLink======")
	outer2, err := sh.GetShare(slink, "getSlinkFile", shell.SetAddress(addr2))
	if err != nil {
		log.Fatal("get slink fails ", err)
	}
	obuf2 := new(bytes.Buffer)
	obuf2.ReadFrom(outer2)
	gotTag2 := md5.Sum(obuf2.Bytes())
	fmt.Println("gotTag2 is ", hex.EncodeToString(gotTag2[:]))
	if hex.EncodeToString(gotTag2[:]) != hex.EncodeToString(tag[:]) {
		log.Fatal("get share file failed, got md5: ", hex.EncodeToString(gotTag2[:]), "expected: ", hex.EncodeToString(tag[:]))
	}
	log.Println("******test success!******")
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
