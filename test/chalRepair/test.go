package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	df "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	shell "github.com/memoio/mefs-go-http-client"
)

//随机文件最大大小
const randomDataSize = 1024 * 1024 * 10
const bucketName = "Bucket01"
const dataCount = 3
const parityCount = 2

const moneyTo = 1000000000000000000

var ethEndPoint string

func main() {
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address")
	flag.Parse()
	ethEndPoint = *eth

	if err := ChallengeTest(); err != nil {
		log.Fatal(err)
	}
}

func ChallengeTest() error {
	sh := shell.NewShell("localhost:5001")
	testuser, err := sh.CreateUser()
	if err != nil {
		log.Fatal("Create user failed :", err)
		return err
	}
	addr := testuser.Address
	uid, err := address.GetIDFromAddress(addr)
	if err != nil {
		log.Fatal("address to id failed")
		return err
	}

	test := true
	for test {
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
		test = false
		break
	}

	err = sh.StartUser(addr)
	if err != nil {
		log.Fatal("Start user failed :", err)
		return err
	}
	for {
		_, err := sh.ShowStorage(shell.SetAddress(addr))
		if err != nil {
			time.Sleep(20 * time.Second)
			log.Println(addr, " not start, waiting..., err : ", err)
			continue
		}
		var opts []func(*shell.RequestBuilder) error
		//设置某些选项
		opts = append(opts, shell.SetAddress(addr))
		opts = append(opts, shell.SetDataCount(dataCount))
		opts = append(opts, shell.SetParityCount(parityCount))
		opts = append(opts, shell.SetPolicy(df.RsPolicy))
		bk, err := sh.CreateBucket(bucketName, opts...)
		if err != nil {
			time.Sleep(20 * time.Second)
			log.Println(addr, " not start, waiting, err : ", err)
			continue
		}
		log.Println(bk, "addr:", addr)
		log.Println(addr, "started, begin to upload")
		break
	}
	//然后开始上传文件
	r := rand.Int63n(randomDataSize)
	data := make([]byte, r)
	fillRandom(data)
	buf := bytes.NewBuffer(data)
	objectName := addr + "_" + strconv.Itoa(int(r))
	log.Println("Begin to upload", objectName, "Size is", ToStorageSize(r), "addr", addr)
	beginTime := time.Now().Unix()
	ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Println(addr, "Upload failed", err)
		return err
	}
	storagekb := float64(r) / 1024.0
	endTime := time.Now().Unix()
	speed := storagekb / float64(endTime-beginTime)
	log.Println("Upload", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", "addr", addr)
	log.Println(ob.String() + "address: " + addr)

	check := 0
	var LastChallengeTime string
	for {
		check++
		if check > 5 {
			log.Fatal("Challenge time not change")
			return errors.New("ChallengeTest failed, Last challenge time not change")
		}
		time.Sleep(2 * time.Minute)
		getOb, err := sh.ListObjects(bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println("List Objects failed :", err)
			return err
		}
		log.Println("Object Name :", getOb.Objects[0].Name, "\nObject LastChallenge Time :", getOb.Objects[0].LatestChalTime)
		if check >= 2 && strings.Compare(LastChallengeTime, getOb.Objects[0].LatestChalTime) != 0 {
			log.Println("Challenge success")
			break
		}
		LastChallengeTime = getOb.Objects[0].LatestChalTime
	}

	keepers, err := sh.ListKeepers(shell.SetAddress(addr))
	if err != nil {
		log.Println("list keepers error :", err)
		return err
	}
	keeper := keepers.Peers[0].PeerID
	log.Println("keeper :", keepers.Peers[0].PeerID)
	bm, err := metainfo.NewBlockMeta(uid, "1", "0", "0")
	if err != nil {
		log.Println(err)
		return err
	}
	cid := bm.ToString()
	kmBlock, err := metainfo.NewKeyMeta(cid, metainfo.BlockPos)
	if err != nil {
		log.Println(err)
		return err
	}
	blockMeta := kmBlock.ToString()
	log.Println("got blockMeta: ", blockMeta, " from: ", keeper)
	var provider string
	resPid, err := sh.GetFrom(blockMeta, keeper)
	if err == nil {
		provider = strings.Split(resPid.Extra, metainfo.DELIMITER)[0]
		log.Println("provider :", provider)
	} else {
		log.Println("get blockmeta error :", err)
		return err
	}

	ret, err := getBlock(sh, cid, provider) //获取块的MD5
	if err != nil || ret == "" {
		log.Println("get block from old provider error :", err)
		return err
	}
	log.Println("md5 of block`s rawdata :", ret)

	//在provider上删除指定块
	km, err := metainfo.NewKeyMeta(cid, metainfo.Block)
	if err != nil {
		log.Println("construct del block KV error :", err)
		return err
	}

	_, err = sh.DeleteFrom(km.ToString(), provider)
	if err != nil {
		fmt.Println("run dht delete error :", err)
		return err
	}

	time.Sleep(1 * time.Minute)
	nret, err := getBlock(sh, cid, provider) //获取块的MD5
	if nret != "" && err == nil {
		log.Println("get block from provider: ", provider, " expcted not")
		return err
	}

	log.Println("successfully delete block :", cid, " in provider", provider)

	// read whole file again
	outer, err := sh.GetObject(objectName, bucketName, shell.SetAddress(addr))
	if err != nil {
		log.Fatal("download file ", objectName, " err:", err)
		return err
	}

	obuf := new(bytes.Buffer)
	obuf.ReadFrom(outer)
	if obuf.Len() != int(r) {
		log.Fatal("download file ", objectName, "failed, got: ", obuf.Len(), "expected: ", r)
	}

	log.Println("successfully get object :", objectName, " in bucket:", bucketName)

	time.Sleep(50 * time.Minute)
	//获取新的provider，从新的provider上获得块的MD5
	var newProvider string
	res, err := sh.GetFrom(blockMeta, keeper)
	if err == nil {
		newProvider = strings.Split(res.Extra, metainfo.DELIMITER)[0]
		log.Println("newProvider :", newProvider)
	} else {
		log.Println("get newblockmeta error :", err)
		return err
	}

	newRet, err := getBlock(sh, cid, newProvider)
	if err != nil || newRet == "" {
		log.Println("get block from new provider error :", err)
		return err
	}

	log.Println("md5 of repaired block`s rawdata :", newRet)
	if ret == newRet {
		log.Println("Repair success")
	} else {
		log.Println("old and new block`s rawdata md5 not match")
		return errors.New("repair failed")
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
	log.Printf("tx sent: %s\n", signedTx.Hash().Hex())
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

func getBlock(sh *shell.Shell, cid, provider string) (string, error) {
	i := 0
	for i < 10 {
		ret, err := sh.GetBlockFrom(cid, provider)
		if err == nil && ret != "" {
			log.Println("Getblock success in ", i+1, " try")
			return ret, nil
		}
		i++
	}
	return "", errors.New("Tried Too Many Times")
}
