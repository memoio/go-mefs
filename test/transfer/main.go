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
	//--eth=http://119.147.213.219:8101 --qeth=http://119.147.213.219:8101      testnet网
	eth := flag.String("eth", "http://119.147.213.219:8101", "eth api address;")             //dev网
	addr := flag.String("addr", "0x1a249DB4cc739BD53b05E2082D3724b7e033F74F", "transfer to") //dev网，用于keeper、provider连接
	money := flag.Int64("money", 6, "transfer money")

	flag.Parse()
	ethEndPoint = *eth
	toAddr := *addr
	toMoney := *money

	contracts.EndPoint = ethEndPoint
	qethEndPoint = ethEndPoint

	num := test.QueryBalance("0x1a249DB4cc739BD53b05E2082D3724b7e033F74F", ethEndPoint)

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
