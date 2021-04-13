package main

import (
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils"
)

var (
	ethEndPoint  string
	qethEndPoint string
)

const (
	moneyTo  = 1000000000000000
	waitTime = 3 * time.Second
)

func main() {
	utils.StartLogger()
	//--eth=http://119.147.213.220:8193 --qeth=http://119.147.213.220:8196      testnet网
	eth := flag.String("eth", "http://119.147.213.220:8193", "eth api address;")   //dev网
	qeth := flag.String("qeth", "http://119.147.213.220:8196", "eth api address;") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth
	contracts.EndPoint = ethEndPoint

	num := test.QueryBalance("0x1a249DB4cc739BD53b05E2082D3724b7e033F74F", ethEndPoint)
	log.Println("用于转账的账号余额:", num)

	//ethEndPoint = *qeth //用正常的链（http://119.147.213.220:8192）给新建账户转账
	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails", err)
	}

	test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)

	log.Println("create account success")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式

	log.Println("===============start test================")
	defer log.Println("==============finish test===============")

	//ethEndPoint = *eth
	log.Println("=====start set mapper addr=====")
	contracts.EndPoint = ethEndPoint
	cManage := contracts.NewCManage(localAddr, userSk)
	addrSet, _, err := cManage.GetMapperFromAdmin(localAddr, "test", true)
	if err != nil {
		log.Fatal("set addr fails", err)
	}
	time.Sleep(waitTime)

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	addrGot, mapperInstance, err := cManage.GetMapperFromAdmin(localAddr, "test", false)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	if addrSet.String() != addrGot.String() {
		log.Fatal(addrSet.String(), "set different from got:", addrGot.String())
	}

	log.Println("=====start add addr first=====")
	contracts.EndPoint = ethEndPoint

	err = cManage.AddToMapper(localAddr, mapperInstance)
	if err != nil {
		log.Fatal("set addr fails", err)
	}
	time.Sleep(waitTime)

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	aGot, err := cManage.GetAddressFromMapper(mapperInstance)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	le := len(aGot)

	if aGot[le-1].String() != localAddr.String() {
		log.Fatal(localAddr.String(), " set different from got:", aGot[le-1].String())
	}

	log.Println("=====start add addr second=====")
	contracts.EndPoint = ethEndPoint

	err = cManage.AddToMapper(addrSet, mapperInstance)
	if err != nil {
		log.Fatal("set addr fails", err)
	}
	time.Sleep(waitTime)

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	aGot, err = cManage.GetAddressFromMapper(mapperInstance)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	le = len(aGot)

	if aGot[le-1].String() != addrSet.String() {
		log.Fatal(addrSet.String(), " set different from got:", aGot[le-1].String())
	}

	log.Println("=====test pass=====")
}
