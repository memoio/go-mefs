package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	df "github.com/memoio/go-mefs/data-format"
	shell "github.com/memoio/mefs-go-http-client"
)

//每个用户上传对象数目
const ObjectCount = 1

//随机文件最大大小
const randomDataSize = 1024 * 1024 * 3
const bucketName = "Bucket01"
const dataCount = 3
const parityCount = 2

func main() {
	if err := ChallengeTest(); err != nil {
		log.Fatal(err)
	}
}

func ChallengeTest() error {
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
		//设置某些选项
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
	//然后开始上传文件
	for j := 0; j < ObjectCount; j++ {
		r := rand.Int63n(randomDataSize)
		data := make([]byte, r)
		fillRandom(data)
		buf := bytes.NewBuffer(data)
		objectName := addr + "_" + strconv.Itoa(int(r))
		fmt.Println("  Begin to upload", objectName, "Size is", ToStorageSize(r), "addr", addr)
		beginTime := time.Now().Unix()
		ob, err := sh.PutObject(buf, objectName, bucketName, shell.SetAddress(addr))
		if err != nil {
			log.Println(addr, "Upload failed", err)
			return err
		}
		storagekb := float64(r) / 1024.0
		endTime := time.Now().Unix()
		speed := storagekb / float64(endTime-beginTime)
		fmt.Println("  Upload", objectName, "Size is", ToStorageSize(r), "speed is", speed, "KB/s", "addr", addr)
		fmt.Println(ob.String() + "address: " + addr)

	}
	check := 0
	var LastChallengeTime string
	for {
		check++
		if check > 5 {
			return errors.New("ChallengeTest failed, Last challenge time not change")
		}
		time.Sleep(5 * time.Minute)
		getOb, err := sh.ListObjects(bucketName, shell.SetAddress(addr))
		if err != nil {
			fmt.Println("List Objects failed :", err)
			return err
		}
		log.Println("Object Name :", getOb.Objects[0].ObjectName, "\nObject LastChallenge Time :", getOb.Objects[0].LatestChalTime)
		if check >= 2 && strings.Compare(LastChallengeTime, getOb.Objects[0].LatestChalTime) != 0 {
			log.Println("Challenge success")
			break
		}
		LastChallengeTime = getOb.Objects[0].LatestChalTime
	}

	keepers, err := sh.ListKeepers(shell.SetAddress(addr))
	if err != nil {
		fmt.Println("list keepers error :", err)
		return err
	}
	keeper := keepers.Peers[0].PeerID
	fmt.Println("keeper :", keepers.Peers[0].PeerID)
	bm, err := metainfo.NewBlockMeta(uid, "1", "0", "0")
	if err != nil {
		fmt.Println(err)
		return err
	}
	cid := bm.ToString()
	kmBlock, err := metainfo.NewKeyMeta(cid, metainfo.Local, metainfo.SyncTypeBlock)
	if err != nil {
		fmt.Println(err)
		return err
	}
	blockMeta := kmBlock.ToString()
	fmt.Println("blockMeta : ", blockMeta)
	var provider string
	resPid, err := sh.GetFrom(blockMeta, keeper)
	if err == nil {
		provider = strings.Split(resPid.Extra, "|")[0]
		fmt.Println("provider :", provider)
	} else {
		fmt.Println("get blockmeta error :", err)
		return err
	}
	ret, err := getBlock(sh, cid, provider) //获取块的MD5
	if err != nil || ret == "" {
		fmt.Println("get block from old provider error :", err)
		return err
	}
	fmt.Println("md5 of block`s rawdata :", ret)

	//在provider上删除指定块
	km, err := metainfo.NewKeyMeta(cid, metainfo.DeleteBlock)
	if err != nil {
		fmt.Println("construct del block KV error :", err)
		return err
	}
	_, err = sh.ChallengeTest(km.ToString(), provider)
	if err != nil {
		fmt.Println("run dht challengeTest error :", err)
		return err
	}
	fmt.Println("delete block :", cid, " in provider")
	time.Sleep(42 * time.Minute)

	//获取新的provider，从新的provider上获得块的MD5
	var newProvider string
	res, err := sh.GetFrom(blockMeta, keeper)
	if err == nil {
		newProvider = strings.Split(res.Extra, "|")[0]
		fmt.Println("newProvider :", newProvider)
	} else {
		fmt.Println("get newblockmeta error :", err)
		return err
	}

	newRet, err := getBlock(sh, cid, newProvider)
	if err != nil || newRet == "" {
		fmt.Println("get block from new provider error :", err)
		return err
	}
	fmt.Println("md5 of repaired block`s rawdata :", newRet)
	if ret == newRet {
		fmt.Println("Repair success")
	} else {
		fmt.Println("old and new block`s rawdata md5 not match")
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

func getBlock(sh *shell.Shell, cid, provider string) (string, error) {
	i := 0
	var err error
	for i < 10 {
		ret, err := sh.GetBlockFrom(cid, provider)
		if err == nil {
			fmt.Println("Getblock success in ", i+1, " try")
			return ret, nil
		}
		fmt.Println("get block failed, now try again")
		i++
	}
	return "", err
}
