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

const ( //indexerHex indexerAddress, it is well known
	indexerHex = "0x9e4af0964ef92095ca3d2ae0c05b472837d8bd37"
	moneyTo    = 1000000000000000
)

func main() {
	//--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101      testnet网
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address;")    //dev网，用于user连接
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address;") //dev网，用于keeper、provider连接
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
	resAddr, resInsatnce, err := contracts.DeployMapper(localAddr, userSk)
	if err != nil {
		log.Fatal("deploy mapper fails:", err)
	}

	log.Println("start add to mapper")
	err = contracts.AddToMapper(localAddr, resAddr, userSk, resInsatnce)
	if err != nil {
		log.Fatal("add mapper fails: ", err)
	}

	log.Println("start get address from mapper remote")
	contracts.EndPoint = qethEndPoint
	mapperAddr, err := contracts.GetAddrsFromMapper(localAddr, resInsatnce)
	if err != nil {
		log.Fatal("get mapper fails: ", err)
	}

	if resAddr != mapperAddr[len(mapperAddr)-1] {
		log.Fatal("address is different from remote")
	}
	log.Println("*****test pass*****")
}
