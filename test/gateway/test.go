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
	minioClient, err := minio.New("127.0.0.1:5080", "0xe95c4F0eb00256a9Ffac626f135D466CA28586ba", "123456789", ssl)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("Successfully link client.")

	bucketName := "test"
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

	err = minioClient.RemoveBucket(bucketName)
	if err != nil {
		log.Println(err)
		return
	}

	found, err = minioClient.BucketExists(bucketName)
	if err != nil {
		log.Println("head bucket fails: ", err)
		return
	}

	if found {
		log.Println("remove fails")
	}

	log.Println("Successfully remove mybucket.")

	err = minioClient.MakeBucket(bucketName, "us-east-1")
	if err != nil {
		log.Println(err)
		return
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
