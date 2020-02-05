package main

import (
	"flag"
	"fmt"
	"math/big"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
)

var (
	ethEndPoint  string
	qethEndPoint string
)

const (
	moneyTo = 1000000000000000000
)

func main() {
	//--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101      testnet网
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address;")               //dev网
	addr := flag.String("addr", "0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6", "transfer to") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	toAddr := *addr
	contracts.EndPoint = ethEndPoint
	qethEndPoint = ethEndPoint

	num := test.QueryBalance("0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6", ethEndPoint)
	fmt.Println("用于转账的账号余额:", num)

	var balance *big.Int
	balance = test.QueryBalance(toAddr, qethEndPoint)
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		test.TransferTo(big.NewInt(moneyTo), toAddr, ethEndPoint, qethEndPoint)
	}

	return
}
