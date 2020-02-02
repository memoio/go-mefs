package main

import (
	"flag"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/pos"
)

const (
	userAddr = "0x208649111Fd9253B76950e9f827a5A6dd616340d"
	userSk   = "8f9eb151ffaebf2fe963e6185f0d1f8c1e8397e5905b616958d765e7753329ea"
	moneyTo  = 1000000000000000
)

var ethEndPoint, qethEndPoint string

func main() {
	flag.String("testnet", "--eth=http://39.100.146.21:8101 --qeth=http://47.92.5.51:8101", "testnet commands")
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address for set;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address for query;")
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	contracts.EndPoint = ethEndPoint
	userAddr := pos.GetPosAddr()

	localAddr := common.HexToAddress(userAddr[2:])

	ukAddr, _, err := contracts.GetUpkeeping(localAddr, localAddr, "latest")
	if err != nil {
		log.Fatal(userAddr, "has not deployed upkeeping")
		return
	}

	balance := test.QueryBalance(ukAddr.String(), qethEndPoint)
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		test.TransferTo(big.NewInt(moneyTo), ukAddr.String(), ethEndPoint, qethEndPoint)
	}

	localID, _ := address.GetIDFromAddress(localAddr.String())
	ukID, _ := address.GetIDFromAddress(ukAddr.String())
	uItem, err := role.GetUpkeepingInfo(localID, ukID)
	if err != nil {
		log.Fatal("Upkeeping has no information, err: ", err)
		return
	}

	var tempPro []string
	for _, pid := range uItem.ProviderIDs {
		if utils.CheckDup(tempPro, pid) {
			tempPro = append(tempPro, pid)
		}
	}

	log.Println(userAddr, "'s upkeeping addr: ", ukAddr, " has balance: ", balance)
	log.Println(userAddr, "has keeper: ", uItem.KeeperIDs)
	log.Println(userAddr, "has provider: ", uItem.ProviderIDs)
	log.Println(userAddr, "has dedeup provider: ", tempPro)
}
