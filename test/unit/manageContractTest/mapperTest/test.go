package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
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

	num := queryBalance("0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6")
	fmt.Println("用于转账的账号余额:", num)

	userAddr, userSk, err := createAddr()
	if err != nil {
		log.Fatal("create user fails", err)
	}
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

func createAddr() (string, string, error) {
	identity, err := config.CreateID(os.Stdout, 2048)
	if err != nil {
		return "", "", err
	}
	address, err := address.GetAddressFromID(identity.PeerID)
	if err != nil {
		return "", "", err
	}
	addressHex := address.Hex()
	sk, err := utils.IPFSskToEthsk(identity.PrivKey)
	if err != nil {
		return "", "", err
	}

	var balance *big.Int
	balance = queryBalance(addressHex)
	if err != nil {
		return "", "", err
	}
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		transferTo(big.NewInt(moneyTo), addressHex)
	}

	for i := 1; i <= 35; i++ {
		time.Sleep(30 * time.Second)
		balance = queryBalance(addressHex)
		if balance.Cmp(big.NewInt(moneyTo)) >= 0 {
			break
		}
		log.Println(addressHex, "'s Balance now:", balance.String(), ", waiting for transfer success")
		if (i % 6) == 0 {
			log.Println("第", i/6+1, "次触发转账")
			transferTo(big.NewInt(moneyTo), addressHex)
		}
	}

	if balance.Cmp(big.NewInt(moneyTo)) < 0 {
		return addressHex, sk, errors.New("转账失败")
	}

	return addressHex, sk, nil
}

func queryBalance(addr string) *big.Int {
	var result string
	client, err := rpc.Dial(ethEndPoint)
	if err != nil {
		log.Fatal("rpc.dial err:", err)
	}
	err = client.Call(&result, "eth_getBalance", addr, "latest")
	if err != nil {
		log.Fatal("client.call err:", err)
	}
	return utils.HexToBigInt(result)
}
