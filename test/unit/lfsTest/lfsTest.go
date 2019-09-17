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
		log.Fatal("Create user failed :", err)
		return err
	}
	addr := testuser.Address
	_, err = address.GetIDFromAddress(addr)
	if err != nil {
		log.Fatal("address to id failed")
		return err
	}

	transferTo(big.NewInt(moneyTo), addr)
	time.Sleep(90 * time.Second)
	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(addr)
		if balance.Cmp(big.NewInt(moneyTo)) >= 0 {
			break
		}
		log.Println(addr, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}

	log.Println("2. test start lfs")

	var startOpts []func(*shell.RequestBuilder) error
	//set option of bucket
	startOpts = append(startOpts, shell.SetOp("ks", "2"))
	startOpts = append(startOpts, shell.SetOp("ps", "6"))
	err = sh.StartUser(addr, startOpts...)
	if err != nil {
		log.Println("Start user failed :", err)
		return err
	}

	log.Println("3. test rs create bucket")

	bucketName := "Bucket0"
	var opts []func(*shell.RequestBuilder) error
	//set option of bucket
	opts = append(opts, shell.SetAddress(addr))
	opts = append(opts, shell.SetDataCount(dataCount))
	opts = append(opts, shell.SetParityCount(parityCount))
	opts = append(opts, shell.SetPolicy(df.RsPolicy))
	_, err = sh.CreateBucket(bucketName, opts...)
	if err != nil {
		log.Println("create bucket err: ", err)
	}

	bk, err := sh.HeadBucket(bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("create bucket err: ", err)
	}

	fmt.Println(bk.Buckets)

	if bk.Buckets[0].BucketName != bucketName {
		log.Println("create bucket", bucketName, "fails")
	}

	if bk.Buckets[0].Policy != df.RsPolicy {
		log.Println("create bucket fails Policy")
	}

	if bk.Buckets[0].DataCount != dataCount {
		log.Println("create bucket fails datacount")
	}

	if bk.Buckets[0].ParityCount != parityCount {
		log.Println("create bucket fails paritycount")
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
		log.Fatal("put object fails")
	} else {
		storagekb := float64(r) / 1024.0
		uploadEndTime := time.Now().Unix()
		speed := fmt.Sprintf("%.2f", storagekb/float64(uploadEndTime-uploadBeginTime))
		log.Println(" Upload file", fileNum, "success，Filename is", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", "addr", addr)
		log.Println(ob.String() + "address: " + addr)
	}

	obj, err := sh.HeadObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println("head file ", objectName, " err:", err)
		return err
	}

	if obj.Objects[0].ObjectName != objectName {
		log.Println("head file ", objectName, "but got: ", obj.Objects[0].ObjectName)
	}

	if int64(obj.Objects[0].ObjectSize) != r {
		log.Println("head file siez: ", r, "but got: ", obj.Objects[0].ObjectSize)
	}

	log.Println("5. test rs get object")

	outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("download file ", objectName, " err:", err)
		return err
	}

	obuf := new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(obj.Objects[0].ObjectSize) {
		log.Fatal("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", obj.Objects[0].ObjectSize)
	}

	log.Println("6. test mul create bucket")

	mbucketName := "Bucket1"
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
		log.Println("create mbucket err: ", err)
	}

	fmt.Println(bk.Buckets)

	if bk.Buckets[0].BucketName != mbucketName {
		log.Println("create mbucket", mbucketName, "fails")
	}

	if bk.Buckets[0].Policy != df.MulPolicy {
		log.Println("create mbucket fails Policy")
	}

	if bk.Buckets[0].DataCount != 1 {
		log.Println("create mbucket fails datacount")
	}

	if bk.Buckets[0].ParityCount != parityCount+dataCount-1 {
		log.Println("create mbucket fails paritycount")
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
		log.Fatal("put object fails")
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

	if obj.Objects[0].ObjectName != objectName {
		log.Println("head file ", objectName, "but got: ", obj.Objects[0].ObjectName)
	}

	if int64(obj.Objects[0].ObjectSize) != r1 {
		log.Println("head file siez: ", r1, "but got: ", obj.Objects[0].ObjectSize)
	}

	log.Println("8. test mul get object")

	outer, err = sh.GetObject(objectName, mbucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("download file ", objectName, " err:", err)
		return err
	}

	obuf = new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(obj.Objects[0].ObjectSize) {
		log.Fatal("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", obj.Objects[0].ObjectSize)
	}

	log.Println("9. test showstorage")

	res, err := sh.ShowStorage(shell.SetAddress(addr))
	if err != nil {
		log.Fatal("show storage err : ", err)
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

func transferTo(value *big.Int, addr string) {
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	log.Println("ethclient.Dial success")

	privateKey, err := crypto.HexToECDSA("928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("crypto.HexToECDSA success")

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	log.Println("cast public key to ECDSA success")

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("client.PendingNonceAt success")
	gasLimit := uint64(21000) // in units

	gasPrice := big.NewInt(30000000000) // in wei (30 gwei)
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("client.SuggestGasPrice success")

	toAddress := common.HexToAddress(addr[2:])
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("client.NetworkID error,use the default chainID")
		chainID = big.NewInt(666)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("types.SignTx success")

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("transfer ", value.String(), "to", addr)
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
