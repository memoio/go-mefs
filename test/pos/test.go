package main

import (
	"context"
	"crypto/ecdsa"
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/role"
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

	balance := queryBalance(ukAddr.String())
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		transferTo(big.NewInt(moneyTo), ukAddr.String())
	}

	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(ukAddr.String())
		if balance.Cmp(big.NewInt(moneyTo)) >= 0 {
			break
		}

		log.Println(ukAddr, "'s Balance now:", balance.String(), ", waiting for transfer success")
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

func transferTo(value *big.Int, addr string) {
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	log.Println("ethclient.Dial success")

	privateKey, err := crypto.HexToECDSA("928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("crypto.HexToECDSA success")

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	log.Println("cast public key to ECDSA success")

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("client.PendingNonceAt success")
	gasLimit := uint64(21000) // in units

	gasPrice := big.NewInt(30000000000) // in wei (30 gwei)
	gasPrice, err = client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("client.SuggestGasPrice success")

	toAddress := common.HexToAddress(addr[2:])
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Println("client.NetworkID error,use the default chainID")
		chainID = big.NewInt(666)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("types.SignTx success")

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("transfer ", value.String(), "to", addr)
	log.Printf("tx sent: %s\n", signedTx.Hash().Hex())
}

func queryBalance(addr string) *big.Int {
	var result string
	client, err := rpc.Dial(qethEndPoint)
	if err != nil {
		log.Fatal("rpc.dial err:", err)
	}
	err = client.Call(&result, "eth_getBalance", addr, "latest")
	if err != nil {
		log.Fatal("client.call err:", err)
	}
	return utils.HexToBigInt(result)
}
