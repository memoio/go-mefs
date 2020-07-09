package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
)

var ethEndPoint, qethEndPoint string

func main() {
	flag.String("testnet", "--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101", "testnet commands")
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address for set;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address for query;")
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	contracts.EndPoint = ethEndPoint

	//检查交易信息
	txHex := "0x758493a866fc738290a57850690e11ef8d379416bffa7be916d814e9c40913ff"
	receipt := contracts.GetTransactionReceipt(common.HexToHash(txHex))
	txReceipt, err := receipt.MarshalJSON()
	if err != nil {
		log.Println("marshal json fails, err: ", err)
		return
	}
	log.Println("TxReceipt:", string(txReceipt))

	//检查区块信息
	blockNumber := receipt.BlockNumber
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		log.Fatal("rpc.Dial err", err)
	}
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("blockNumberInTx:", receipt.BlockNumber.Uint64())
	fmt.Println("blockNumberInBlock:", block.Number().Uint64()) // 5671744
	fmt.Println("blockTime:", block.Time())                     // 1527211625
	fmt.Println("blockTime: ", time.Unix(int64(block.Time()), 0).Format(utils.BASETIME))
	fmt.Println("blockHash:", block.Hash().Hex()) // 0x9e8751ebb5069389b855bba72d949
	count, err := client.TransactionCount(context.Background(), block.Hash())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("txCountInBlock: ", count)
}
