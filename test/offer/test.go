package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"

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

	num := test.QueryBalance("0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6", ethEndPoint)
	fmt.Println("managed account has: ", num)

	var (
		capacity int64 = 1000
		duration int64 = 10000
		price    int64 = 100000
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

	log.Println("===============start test deployOffer================")
	defer log.Println("==============finish test deployOffer successfully===============")

	//ethEndPoint = *eth //用不正常的链（http://47.92.5.51:8101）部署query合约
	log.Println("start deploy offer")
	offerAddr, err := contracts.DeployOffer(localAddr, userSk, capacity, duration, price, false)
	if err != nil {
		log.Fatal("deploy offer fails", err)
	}

	log.Println("start get offerInfo from remote")
	contracts.EndPoint = qethEndPoint
	offerGot, _, err := contracts.GetLatestOffer(localAddr, localAddr)
	if err != nil {
		log.Fatal("get offer from remote fails: ", err)
	}

	if offerGot.String() != offerAddr.String() {
		log.Fatal(offerAddr.String(), "set different from got:", offerGot.String())
	}

	log.Println("start deploy offer again")
	offerAddr, err = contracts.DeployOffer(localAddr, userSk, capacity, duration, price, true)
	if err != nil {
		log.Fatal("redo deploy offer fails", err)
	}

	log.Println("start get offerInfo from remote")
	contracts.EndPoint = qethEndPoint
	offerGot, _, err = contracts.GetLatestOffer(localAddr, localAddr)
	if err != nil {
		log.Fatal("get query from remote fails")
	}

	if offerGot.String() != offerAddr.String() {
		log.Fatal(offerAddr.String(), "set different from got:", offerGot.String())
	}

	log.Println("*****test pass*****")
}
