package main

import (
	"flag"
	"log"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
)

var (
	ethEndPoint  string
	qethEndPoint string
)

const (
	eth2Wei = 1000000000000000000
)

func main() {
	//--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101      testnet网
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address;")               //dev网
	addr := flag.String("addr", "0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6", "transfer to") //dev网，用于keeper、provider连接
	money := flag.Int64("money", 6, "transfer money")

	flag.Parse()
	ethEndPoint = *eth
	toAddr := *addr
	toMoney := *money

	contracts.EndPoint = ethEndPoint
	qethEndPoint = ethEndPoint

	num := test.QueryBalance("0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6", ethEndPoint)

	moneyTo := new(big.Int).Mul(big.NewInt(eth2Wei), big.NewInt(toMoney))

	log.Println("admin has:", num, " transfer: ", moneyTo, " to:", toAddr)

	oldbalance := test.QueryBalance(toAddr, qethEndPoint)
	test.TransferTo(moneyTo, toAddr, ethEndPoint, qethEndPoint)
	balance := test.QueryBalance(toAddr, qethEndPoint)
	oldbalance.Add(oldbalance, moneyTo)
	if balance.Cmp(oldbalance) < 0 {
		log.Println("transfer fails")
		return
	}

	log.Println("transfer success")

	return
}
