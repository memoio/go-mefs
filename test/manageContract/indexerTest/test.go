//deploy indexer-contract for test
package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/indexer"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils"
)

var (
	ethEndPoint string
	resolverKey string
)

const (
	adminAddrStr    = "0x1a249DB4cc739BD53b05E2082D3724b7e033F74F"
	adminSk         = "aca26228a9ed5ca4da2dd08d225b1b1e049d80e1b126c0d7e644d04d0fb910a3"
	indexerAddress  = "0xA36D0F4e56b76B89532eBbca8108d90d8cA006c2"
	localAddress    = "0x9Fe60D25A7D676C1Dabc65ECc1557F43acF83cAd"
	localSk         = "124833f1b4b31a88cea6f82bbe42a92e1bc50c83aaff8142fee5ebee96e2dab5"
	resolverAddress = "0x3395d0586D773DB7500FCa1b713cB99Fdc47f433"
)

func main() {
	utils.StartLogger()
	eth := flag.String("eth", "http://119.147.213.219:8101", "eth api address;")
	flag.Parse()
	ethEndPoint = *eth
	contracts.EndPoint = ethEndPoint

	log.Println("=============start test indexer=============")
	defer log.Println("============finish test indexer===========")

	log.Println("start query admin's balance")
	adminBalance := test.QueryBalance(adminAddrStr, ethEndPoint)
	fmt.Println("the admin's balance is", adminBalance)
	if adminBalance.Cmp(big.NewInt(0)) == 0 {
		log.Fatal("admin's balance is 0")
	}

	//=======first test, we need deploy indexer=======
	// log.Println("start deploy indexer")
	// indexerAddress, indexerInstance, err := contracts.DeployIndexer(adminSk)
	// if err != nil {
	// 	log.Fatal("deploy indexer fails: ", err)
	// }
	// fmt.Println("indexer-contract address is ", indexerAddress.Hex())

	//build indexer instance by indexerAddress
	indexerInstance, err := indexer.NewIndexer(common.HexToAddress(indexerAddress), contracts.GetClient(ethEndPoint))
	if err != nil {
		log.Fatal("new indexerInstance fails: ", err)
	}

	log.Println("start get owner of indexer-contract")
	adminAddr := common.HexToAddress(adminAddrStr)
	indexerOwnerAddr, err := contracts.GetIndexerOwner(adminAddr, indexerInstance)
	if err != nil {
		log.Fatal("get owner of indexer fails: ", err)
	}
	fmt.Println("the owner of indexer is", indexerOwnerAddr.Hex())

	if indexerOwnerAddr.Hex() != adminAddrStr {
		log.Fatal("owner of indexer is different from admin")
	}

	resolverKey = "testV3"

	log.Println("start add resolverAddress to indexer-contract")
	err = test.TransferTo(big.NewInt(int64(1000000)), localAddress, ethEndPoint, ethEndPoint)
	if err != nil {
		log.Fatal("transfer money to localAddress from admin fails: ", err)
	}
	localAddrBalance := test.QueryBalance(localAddress, ethEndPoint)
	fmt.Println("balance of localAddress is ", localAddrBalance)
	err = contracts.AddToIndexer(common.HexToAddress(localAddress), common.HexToAddress(resolverAddress), resolverKey, localSk, indexerInstance)
	if err != nil {
		log.Fatal("add to indexer fails: ", err)
	}

	log.Println("start get information from indexer-contract")
	resolverAddr, resolverOwner, err := contracts.GetAddrFromIndexer(adminAddr, resolverKey, indexerInstance)
	if err != nil {
		log.Fatal("get information from indexer fails: ", err)
	}
	fmt.Println("resolverAddr: ", resolverAddr.Hex(), "resolverOwner: ", resolverOwner.Hex())

	if resolverAddr.Hex() != resolverAddress || resolverOwner.Hex() != localAddress {
		log.Fatal("get resolver information is different")
	}

	log.Println("*****test pass*****")
}
