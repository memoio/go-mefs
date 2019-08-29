package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
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
	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/utils/address"
	shell "github.com/memoio/mefs-go-http-client"
)

//随机文件最大大小
const randomDataSize = 1024 * 1024 * 10
const dataCount = 3
const parityCount = 2

func main() {
	if err := UploadTest(); err != nil {
		log.Fatal(err)
	}
}

func UploadTest() error {
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

	fmt.Println("GetIDFromAddress success,uid is",uid)
	transferTo(big.NewInt(1000000000000000000), addr)
	for {
		balance := queryBalance(addr)
		if balance.Cmp(big.NewInt(10000000000)) > 0 {
			break
		}
		fmt.Println(addr, "'s Balance now:", balance.String(), ", waiting for transfer success")
		time.Sleep(10 * time.Second)
	}
	err = sh.StartUser(addr)
	if err != nil {
		fmt.Println("Start user failed :", err)
		return err
	}

	for {
		err := sh.ShowStorage(shell.SetAddress(addr))
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
		bucketName := "Bucket0"
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
	//upload file
	bucketName := "Bucket0"
	bucketNum := 0
	errNum := 0
	fileNum := 0
	for {
		r := rand.Int63n(randomDataSize)
		data := make([]byte, r)
		fillRandom(data)
		buf := bytes.NewBuffer(data)
		objectName := addr + "_" + strconv.Itoa(int(r))
		fmt.Println("  Begin to upload file",fileNum,"，Filename is", objectName, "Size is", ToStorageSize(r), "addr", addr)
		beginTime := time.Now().Unix()
		ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println(addr, "Upload failed in file ",fileNum,",", err)
			if errNum == 0 {
				bucketNum++
				var opts []func(*shell.RequestBuilder) error
				opts = append(opts, shell.SetAddress(addr))
				opts = append(opts, shell.SetDataCount(dataCount))
				opts = append(opts, shell.SetParityCount(parityCount))
				opts = append(opts, shell.SetPolicy(df.RsPolicy))
				bucketName = "Bucket"+string(bucketNum)
				_, errBucket := sh.CreateBucket(bucketName, opts...)
				if errBucket != nil {
					time.Sleep(20 * time.Second)
					fmt.Println(addr, " not start, waiting, err : ", err)
					continue
				}
				errNum=1
			} else {
				return err
			}
		}
		storagekb := float64(r) / 1024.0
		endTime := time.Now().Unix()
		speed := storagekb / float64(endTime-beginTime)
		fmt.Println("  Upload file",fileNum,"success，Filename is", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", "addr", addr)
		fmt.Println(ob.String() + "address: " + addr)
		fileNum++
		errNum = 0
	}
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

const ethEndPoint = "http://212.64.28.207:8101"

func transferTo(value *big.Int, addr string) {
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		fmt.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA("928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4")
	if err != nil {
		log.Fatal(err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	gasLimit := uint64(21000) // in units

	gasPrice := big.NewInt(30000000000) // in wei (30 gwei)
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	toAddress := common.HexToAddress(addr[2:])
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("transfer ", value.String(), "to", addr)
	fmt.Printf("tx sent: %s\n", signedTx.Hash().Hex())
}

func queryBalance(addr string) *big.Int {
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		fmt.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	Address := common.HexToAddress(addr[2:])
	balance, err := client.PendingBalanceAt(context.Background(), Address)
	if err != nil {
		log.Fatal(err)
	}
	return balance
}

