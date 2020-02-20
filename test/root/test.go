package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
)

var (
	ethEndPoint  string
	qethEndPoint string
)

const (
	moneyTo = 1000000000000000
)

func main() {
	//--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101      testnet网
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address;")    //dev网
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address;") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth
	contracts.EndPoint = ethEndPoint

	var (
		reDeploy = true
	)

	//ethEndPoint = *qeth //用正常的链（http://39.100.146.21:8101）给新建账户转账
	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails", err)
	}

	err = test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("create account success")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式

	log.Println("===============start test deployRoot================")
	defer log.Println("==============finish test deployRoot===============")

	//ethEndPoint = *eth //用不正常的链（http://47.92.5.51:8101）部署query合约
	log.Println("start deploy root")
	rootAddr, err := contracts.DeployRoot(userSk, localAddr, localAddr, reDeploy)
	if err != nil {
		log.Fatal("deploy root fails", err)
	}

	contracts.EndPoint = qethEndPoint
	log.Println("start get root contract")

	gotAddr, _, err := contracts.GetRoot(localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("get root contract fails: ", err)
	}

	if gotAddr.String() != rootAddr.String() {
		log.Fatal("get wrong root contract")
	}

	keyTime := time.Now().Unix()
	res, err := contracts.GetMerkleRoot(localAddr, gotAddr, keyTime)
	if err == nil {
		log.Fatal("get empty merkle root fail, should return err")
	}

	val := sha256.Sum256([]byte{'1'})
	err = contracts.SetMerkleRoot(userSk, gotAddr, keyTime, val)
	if err != nil {
		log.Fatal("set merkle root fails:", err)
	}

	res, err = contracts.GetMerkleRoot(localAddr, gotAddr, keyTime)
	if err != nil {
		log.Fatal("get empty merkle root:", err)
	}

	if bytes.Compare(val[:], res[:]) != 0 {
		log.Println("set merkle root:", val)
		log.Println("get merkle root:", res)
		log.Fatal("get wrong merkle root")
	}

	log.Println("*****test pass*****")
}
