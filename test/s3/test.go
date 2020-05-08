package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"

	"github.com/minio/minio-go/v6"
)

// ssl not supported now.
var ssl = false

// "39.100.146.21:5080"
// "39.100.146.165:5080"
// "39.98.240.7:5080"
var readurl = "39.100.146.21:5080"

// only one
var writeurl = "39.100.146.21:5080"
var account = "0xC2b27Aa18A1930D5b403b2021D8f52044C7B092B"
var pwd = "memoriae"
var bucketName = "test"
var objectName = "vlc-3.0.8-win64.exe"

func main() {
	list()
}

func list() {
	// Initialize minio client object.
	mc, err := minio.New(readurl, account, pwd, ssl)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Successfully link client.")

	found, err := mc.BucketExists(bucketName)
	if err != nil {
		log.Println(err)
		return
	}

	if !found {
		log.Printf("bucket %s not exist", bucketName)
		return
	}

	// Create a done channel to control 'ListObjects' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := true
	objectCh := mc.ListObjects(bucketName, "", isRecursive, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			log.Println(object.Err)
			return
		}
		log.Println(object)
	}

	return
}

func head() (minio.ObjectInfo, error) {
	var obj minio.ObjectInfo

	// Initialize minio client object.
	mc, err := minio.New(readurl, account, pwd, ssl)
	if err != nil {
		log.Fatal(err)
		return obj, err
	}

	log.Println("Successfully link client.")

	objInfo, err := mc.StatObject(bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		log.Println(err)
		return objInfo, err
	}
	log.Println(objInfo)

	return objInfo, err
}

func read() error {
	// Initialize minio client object.
	mc, err := minio.New(readurl, account, pwd, ssl)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println("Successfully link client.")

	object, err := mc.GetObject(bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	objInfo, err := object.Stat()
	if err != nil {
		log.Println(err)
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(object)
	if buf.Len() != int(objInfo.Size) {
		log.Println("read file ", objectName, "failed, got: ", buf.Len(), "expected: ", objInfo.Size)
		return err
	}

	gotTag := md5.Sum(buf.Bytes())
	if hex.EncodeToString(gotTag[:]) != objInfo.ETag {
		log.Println("download file ", objectName, "failed, got md5: ", hex.EncodeToString(gotTag[:]), "expected: ", objInfo.ETag)
		return err
	}

	return nil
}

func write() {

	// Initialize minio client object.
	minioClient, err := minio.New(readurl, account, pwd, ssl)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Successfully link client.")

	found, err := minioClient.BucketExists(bucketName)
	if err != nil {
		log.Println(err)
		return
	}

	if !found {
		err = minioClient.MakeBucket(bucketName, "us-east-1")
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Successfully created mybucket.")
	}

	file, err := os.Open("./test.go")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		log.Println(err)
		return
	}

	n, err := minioClient.PutObject("mybucket", "myobject", file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Successfully uploaded bytes: ", n)

	return
}
