package main

import (
	"fmt"

	"github.com/minio/minio-go/v6"
)

func main() {
	// Use a secure connection.
	ssl := false

	// Initialize minio client object.
	minioClient, err := minio.New("127.0.0.1:5080", "user", "12345678", ssl)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully link client.")

	found, err := minioClient.BucketExists("mybucket")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(found)
	err = minioClient.MakeBucket("mybucket", "us-east-1")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully created mybucket.")
	return
}
