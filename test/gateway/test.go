package main

import (
	"log"
	"os"

	"github.com/minio/minio-go/v6"
)

func main() {
	// Use a secure connection.
	ssl := false

	// Initialize minio client object.
	minioClient, err := minio.New("127.0.0.1:5080", "0x7aD8AA67aFEE05Fd539B938db69957050dDDA1c3", "123456789", ssl)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Successfully link client.")

	found, err := minioClient.BucketExists("mybucket")
	if err != nil {
		log.Println(err)
		return
	}

	if !found {
		err = minioClient.MakeBucket("mybucket", "us-east-1")
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

	n, err := minioClient.PutObject("mybucket", "/test/myobject", file, fileStat.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Successfully uploaded bytes: ", n)

	return
}
