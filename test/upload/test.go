package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/utils"
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
	err := UploadTest(*count)
	if err != nil {
		log.Fatal(err)
	}
}

func UploadTest(count int) error {
	sh := shell.NewShell("localhost:5001")
	testuser, err := sh.CreateUser()
	if err != nil {
		fmt.Println("Create user failed :", err)
		return err
	}
	addr := testuser.Address
	uid, err := address.GetIDFromAddress(addr)
	if err != nil {
		fmt.Println("address to id failed")
		return err
	}

	fmt.Println("GetIDFromAddress success,uid is", uid)
	transferTo(big.NewInt(moneyTo), addr)
	time.Sleep(90 * time.Second)
	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(addr)
		if balance.Cmp(big.NewInt(moneyTo)) >= 0 {
			break
		}
		fmt.Println(addr, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}
	err = sh.StartUser(addr)
	if err != nil {
		fmt.Println("Start user failed :", err)
		return err
	}

	bucketName := "Bucket0"

	for {
		_, err := sh.ShowStorage(shell.SetAddress(addr))
		if err != nil {
			time.Sleep(20 * time.Second)
			fmt.Println(addr, " not start, waiting..., err : ", err)
			continue
		}
		var opts []func(*shell.RequestBuilder) error
		//set option of bucket
		opts = append(opts, shell.SetAddress(addr))
		opts = append(opts, shell.SetDataCount(dataCount))
		opts = append(opts, shell.SetParityCount(parityCount))
		opts = append(opts, shell.SetPolicy(df.RsPolicy))
		bk, err := sh.CreateBucket(bucketName, opts...)
		if err != nil {
			time.Sleep(20 * time.Second)
			fmt.Println(addr, " not start, waiting, err : ", err)
			continue
		}
		fmt.Println(bk, "addr:", addr)
		fmt.Println(addr, "started, begin to upload")
		break
	}

	bucketNum := 0
	errNum := 0
	fileNum := 1
	fileUploadSuccessNum := 0

	//upload file
	for {
		r := rand.Int63n(randomDataSize)
		data := make([]byte, r)
		fillRandom(data)
		buf := bytes.NewBuffer(data)
		objectName := addr + "_" + strconv.Itoa(fileNum)
		fmt.Println("\n  Begin to upload file", fileNum, "，Filename is", objectName, "Size is", ToStorageSize(r), "addr", addr)
		uploadBeginTime := time.Now().Unix()
		ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			errNum++
			fmt.Println(addr, "Upload failed in file ", fileNum, ",", err)
			if errNum < 2 {
				bucketNum++
				var opts []func(*shell.RequestBuilder) error
				opts = append(opts, shell.SetAddress(addr))
				opts = append(opts, shell.SetDataCount(dataCount))
				opts = append(opts, shell.SetParityCount(parityCount))
				opts = append(opts, shell.SetPolicy(df.RsPolicy))
				bucketName = "Bucket" + string(bucketNum)
				_, errBucket := sh.CreateBucket(bucketName, opts...)
				if errBucket != nil {
					fmt.Println("create bucket err: ", err)
					time.Sleep(2 * time.Minute)
					continue
				}
			} else {
				fmt.Println("\n连续两次更换bucket后依然上传失败，可能是网络故障，停止上传")
				fmt.Println("upload ", fileNum, " files,", fileUploadSuccessNum, " files uploaded success.fileUploadSuccess rate is", fileUploadSuccessNum/fileNum)
				break
			}
		} else {
			errNum = 0
			objsInBucket.Store(objectName, bucketName)

			storagekb := float64(r) / 1024.0
			uploadEndTime := time.Now().Unix()
			speed := fmt.Sprintf("%.2f", storagekb/float64(uploadEndTime-uploadBeginTime))
			fmt.Println("  Upload file", fileNum, "success，Filename is", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", "addr", addr)
			fmt.Println(ob.String() + "address: " + addr)
			fileUploadSuccessNum++
			if fileNum == count {
				fmt.Println("upload test complete")
				fmt.Println("upload ", fileNum, " files,", fileUploadSuccessNum, " files success.fileUploadSuccess rate is", fileUploadSuccessNum/count)
				break
			}
		}
		fileNum++
	}
	//download file
	fileDownloadSuccessNum := 0

	objsInBucket.Range(func(k, v interface{}) bool {
		objectName := k.(string)
		bucketName := v.(string)
		outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			fmt.Println("download file ", objectName, " err:", err)
			return true
		}

		obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			fmt.Println("head file ", objectName, " err:", err)
			return true
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(outer)
		if buf.Len() != int(obj.Objects[0].Size) {
			fmt.Println("download file ", objectName, "failed, got: ", buf.Len(), "expected: ", obj.Objects[0].Size)
			return true
		}

		fileDownloadSuccessNum++
		return true
	})

	fmt.Println("download test complete")
	fmt.Println("download ", fileDownloadSuccessNum, " files,")

	fmt.Println("upload: ", fileNum, "; success:", fileUploadSuccessNum, " rate is", fileUploadSuccessNum/count)
	fmt.Println("downlaod: ", fileNum, "; success:", fileDownloadSuccessNum, " rate is", fileDownloadSuccessNum/fileUploadSuccessNum)
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

func transferTo(value *big.Int, addr string) {
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		fmt.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	fmt.Println("ethclient.Dial success")

	privateKey, err := crypto.HexToECDSA("928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("crypto.HexToECDSA success")

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	fmt.Println("cast public key to ECDSA success")

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("client.PendingNonceAt success")
	gasLimit := uint64(21000) // in units

	gasPrice := big.NewInt(30000000000) // in wei (30 gwei)
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("client.SuggestGasPrice success")

	toAddress := common.HexToAddress(addr[2:])
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println("client.NetworkID error,use the default chainID")
		chainID = big.NewInt(666)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("types.SignTx success")

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("transfer ", value.String(), "to", addr)
	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
}

func queryBalance(addr string) *big.Int {
	var result string
	client, err := rpc.Dial(ethEndPoint)
	if err != nil {
		log.Fatal("rpc.dial err:", err)
	}
	err = client.Call(&result, "eth_getBalance", addr, "latest")
	if err != nil {
		log.Fatal("client.call err:", err)
	}
	return utils.HexToBigInt(result)
}
