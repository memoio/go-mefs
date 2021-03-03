package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
)

var (
	ethEndPoint  string
	qethEndPoint string
)

const (
	moneyTo = 1000000000000000
)

func main() {
	//--eth=http://119.147.213.219:8101 --qeth=http://119.147.213.219:8101      testnet网
	eth := flag.String("eth", "http://119.147.213.219:8101", "eth api address;")   //dev网
	qeth := flag.String("qeth", "http://119.147.213.219:8101", "eth api address;") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth
	contracts.EndPoint = ethEndPoint

	num := test.QueryBalance("0x1a249DB4cc739BD53b05E2082D3724b7e033F74F", ethEndPoint)
	fmt.Println("managed account has: ", num)

	var (
		capacity int64 = 1000
		duration int64 = 10000
		price          = big.NewInt(100000)
	)

	//ethEndPoint = *qeth //用正常的链（http://119.147.213.219:8101）给新建账户转账
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

	//ethEndPoint = *eth //用不正常的链（http://119.147.213.219:8101）部署query合约
	log.Println("start deploy offer")
	offerAddr, err := contracts.DeployOffer(localAddr, userSk, capacity, duration, price, false)
	if err != nil {
		log.Fatal("deploy offer fails", err)
	}

	log.Println("start get offerInfo from remote")
	contracts.EndPoint = qethEndPoint
	offerGot, err := contracts.GetOfferAddrs(localAddr, localAddr)
	if err != nil {
		log.Fatal("get offer from remote fails: ", err)
	}
	if len(offerGot) < 1 {
		log.Fatal("get empty offerAddrs")
	}
	if offerGot[len(offerGot)-1].String() != offerAddr.String() {
		log.Fatal(offerAddr.String(), "set different from got:", offerGot[len(offerGot)-1].String())
	}

	localID, _ := address.GetIDFromAddress(localAddr.String())
	oItem, err := role.GetLatestOffer(localID, localID)
	if err != nil {
		log.Fatal("get offer item fails:", err)
	}
	if oItem.Capacity != capacity || oItem.Duration != duration || oItem.Price.Cmp(price) != 0 {
		log.Fatal("offer info is different from set")
	}

	log.Println("start deploy offer again")
	offerAddr, err = contracts.DeployOffer(localAddr, userSk, capacity, duration, price, true)
	if err != nil {
		log.Fatal("redo deploy offer fails", err)
	}

	log.Println("start get offerInfo from remote")
	contracts.EndPoint = qethEndPoint
	offerGot, err = contracts.GetOfferAddrs(localAddr, localAddr)
	if err != nil {
		log.Fatal("get offer from remote fails")
	}
	if len(offerGot) < 1 {
		log.Fatal("get empty offerAddrs")
	}

	if offerGot[len(offerGot)-1].String() != offerAddr.String() {
		log.Fatal(offerAddr.String(), " set different from got:", offerGot[len(offerGot)-1].String())
	}

	oItem, err = role.GetLatestOffer(localID, localID)
	if err != nil {
		log.Fatal("get offer item fails:", err)
	}
	if oItem.Capacity != capacity || oItem.Duration != duration || oItem.Price.Cmp(price) != 0 {
		log.Fatal("offer info is different from set")
	}

	log.Println("*****test pass*****")
}
