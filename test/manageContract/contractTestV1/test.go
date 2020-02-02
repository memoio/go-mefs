package main

import (
	"flag"
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

	//ethEndPoint = *qeth //用正常的链（http://39.100.146.21:8101）给新建账户转账
	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails", err)
	}

	test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)

	log.Println("create account success")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式

	log.Println("===============start test================")
	defer log.Println("==============finish test===============")

	//ethEndPoint = *eth //用不正常的链（http://47.92.5.51:8101）部署query合约
	log.Println("=====start set mapper addr=====")
	contracts.EndPoint = ethEndPoint
	addrSet, _, err := contracts.GetMapperFromAdminV1(localAddr, localAddr, "test", userSk, true)
	if err != nil {
		log.Fatal("set addr fails", err)
	}

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	addrGot, mapperInstance, err := contracts.GetMapperFromAdminV1(localAddr, localAddr, "test", userSk, false)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	if addrSet.String() != addrGot.String() {
		log.Fatal(addrSet.String(), "set different from got:", addrGot.String())
	}

	log.Println("=====start add addr first=====")
	contracts.EndPoint = ethEndPoint

	err = contracts.AddToMapper(localAddr, localAddr, userSk, mapperInstance)
	if err != nil {
		log.Fatal("set addr fails", err)
	}

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	aGot, err := contracts.GetAddrsFromMapper(localAddr, mapperInstance)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	le := len(aGot)

	if aGot[le-1].String() != localAddr.String() {
		log.Fatal(localAddr.String(), " set different from got:", aGot[le-1].String())
	}

	log.Println("=====start add addr second=====")
	contracts.EndPoint = ethEndPoint

	err = contracts.AddToMapper(localAddr, addrSet, userSk, mapperInstance)
	if err != nil {
		log.Fatal("set addr fails", err)
	}

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	aGot, err = contracts.GetAddrsFromMapper(localAddr, mapperInstance)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	le = len(aGot)

	if aGot[le-1].String() != addrSet.String() {
		log.Fatal(addrSet.String(), " set different from got:", aGot[le-1].String())
	}

	log.Println("=====test pass=====")
}
