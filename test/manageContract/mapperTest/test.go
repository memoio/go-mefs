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

const ( //indexerHex indexerAddress, it is well known
	indexerHex = "0xA36D0F4e56b76B89532eBbca8108d90d8cA006c2"
	moneyTo    = 1000000000000000
	waitTime   = 3 * time.Second
)

func main() {
	utils.StartLogger()
	//--eth=http://119.147.213.219:8101 --qeth=http://119.147.213.219:8101      testnet网
	eth := flag.String("eth", "http://119.147.213.219:8101", "eth api address;")   //dev网，用于user连接
	qeth := flag.String("qeth", "http://119.147.213.219:8101", "eth api address;") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth
	contracts.EndPoint = ethEndPoint

	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails", err)
	}

	test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)

	log.Println("create account success")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式

	log.Println("=============start test mapper=============")
	defer log.Println("============finish test mapper===========")

	log.Println("start deploy mapper")
	cManage := contracts.NewCManage(localAddr, userSk)
	resAddr, resInsatnce, err := cManage.DeployMapper()
	if err != nil {
		log.Fatal("deploy mapper fails:", err)
	}
	time.Sleep(waitTime)

	log.Println("start add to mapper")
	err = cManage.AddToMapper(resAddr, resInsatnce)
	if err != nil {
		log.Fatal("add mapper fails: ", err)
	}
	time.Sleep(waitTime)

	log.Println("start get address from mapper remote")
	contracts.EndPoint = qethEndPoint
	mapperAddr, err := cManage.GetAddressFromMapper(resInsatnce)
	if err != nil {
		log.Fatal("get mapper fails: ", err)
	}

	if resAddr != mapperAddr[len(mapperAddr)-1] {
		log.Fatal("address is different from remote")
	}
	log.Println("*****test pass*****")
}
