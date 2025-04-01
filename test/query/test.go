package main

import (
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

var (
	ethEndPoint  string
	qethEndPoint string
)

const (
	moneyTo  = 1000000000000000 //1e15
	waitTime = 10 * time.Second
)

func main() {
	utils.StartLogger()
	//--eth=http://119.147.213.219:8101 --qeth=http://119.147.213.219:8101      testnet网
	eth := flag.String("eth", "http://119.147.213.220:8191", "eth api address;")   //dev网
	qeth := flag.String("qeth", "http://119.147.213.220:8194", "eth api address;") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	contracts.EndPoint = ethEndPoint

	var (
		capacity int64 = 10
		duration int64 = 10
		price          = big.NewInt(100)
		ks             = 3
		ps             = 5
		reDeploy       = true
	)

	//ethEndPoint = *qeth
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

	log.Println("===============start test deployQuery================")
	defer log.Println("==============finish test deployQuery===============")

	//ethEndPoint = *eth 
	log.Println("start deploy query")
	cMarket := contracts.NewCM(localAddr, userSk)
	queryAddr, err := cMarket.DeployQuery(capacity, duration, price, ks, ps, reDeploy)
	if err != nil {
		log.Fatal("deploy Query fails ", err)
	}

	contracts.EndPoint = qethEndPoint

	time.Sleep(waitTime)
	queryGot, err := cMarket.GetQueryAddrs(localAddr)
	if err != nil {
		log.Fatal("get query addrs fails ", err)
	}
	if len(queryGot) < 1 {
		log.Fatal("get empty queryAddrs")
	}
	if queryGot[len(queryGot)-1].String() != queryAddr.String() {
		log.Fatal(queryAddr.String(), " set different from got:", queryGot[len(queryGot)-1].String())
	}

	log.Println("start get 'completed' params")
	localID, _ := address.GetIDFromAddress(localAddr.String())
	queryID, _ := address.GetIDFromAddress(queryAddr.String())
	qItem, err := role.GetLatestQuery(localID)
	if err != nil {
		log.Fatal("get query fails: ", err)
	}

	if qItem.KeeperNums != int32(ks) || qItem.ProviderNums != int32(ps) {
		log.Fatal("query info is different from set")
	}

	if qItem.Completed {
		log.Fatal("completed info is wrong")
	}

	log.Println("start set completed")
	contracts.EndPoint = ethEndPoint
	err = cMarket.SetQueryCompleted(queryAddr)
	if err != nil {
		log.Fatal("set query completed fails:", err)
	}

	time.Sleep(waitTime)
	contracts.EndPoint = qethEndPoint
	log.Println("start get 'completed' params")
	qItem, err = role.GetQueryInfo(localID, queryID)
	if err != nil {
		log.Fatal("get query fails:", err)
	}

	if !qItem.Completed {
		log.Fatal("set completed fails:")
	}
	log.Println("*****test pass*****")
}
