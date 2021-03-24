package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"

	"github.com/memoio/go-mefs/utils"
	shell "github.com/memoio/mefs-go-http-client"
)

var ethEndPoint, qethEndPoint string

func main() {
	utils.StartLogger()
	count := flag.Int("count", 20, "count of files downloaded each time")
	eth := flag.String("eth", "http://119.147.213.219:8101", "eth api address for set")
	qeth := flag.String("qeth", "http://119.147.213.219:8101", "eth api address for query")

	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	err := downloadTest(*count)
	if err != nil {
		log.Fatal(err)
	}
}

//指定user节点,重复下载文件
func downloadTest(count int) error {
	sh := shell.NewShell("http://121.37.158.192:5001")

	//get user's address
	idOutPut, err := sh.ID()
	if err != nil {
		log.Println("get user's address err:", err)
		return err
	}

	addr := idOutPut.AccountAddr
	if addr == "" {
		log.Println("get user's address is nil")
		return errors.New("get user's address is nil")
	}
	fmt.Println("addr:", addr)

	bucketName := "bucket1"
	var objectName string
	var fileDownloadSuccessNum int
	for i := 1; i <= count; i++ {
		objectName = "file" + strconv.Itoa(i)

		outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println("download file ", objectName, " err:", err)
			continue
		}

		obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println("head file ", objectName, " err:", err)
			continue
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(outer)

		if buf.Len() != int(obj.Objects[0].Size) {
			log.Println("download file ", objectName, "failed, got: ", buf.Len(), " expected: ", obj.Objects[0].Size)
			continue
		}

		gotTag := md5.Sum(buf.Bytes())
		if hex.EncodeToString(gotTag[:]) != obj.Objects[0].MD5 {
			log.Println("download file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", obj.Objects[0].MD5)
			continue
		}

		log.Println("download file ", objectName, " success, got: ", ToStorageSize(int64(buf.Len())))
		fileDownloadSuccessNum++
	}

	if fileDownloadSuccessNum == count {
		log.Println("download test complete")
		log.Println("download ", fileDownloadSuccessNum, " files")
		return nil
	}

	log.Println("download test failed")
	log.Println("download ", fileDownloadSuccessNum, " files")
	return errors.New("download test failed")
}

//ToStorageSize transfer length to human-friendly file size
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
