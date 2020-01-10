package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
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

const (
	moneyTo = 1000000000000000
)

func main() {
	//--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101      testnet网
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address;")    //dev网
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address;") //dev网，用于keeper、provider连接
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth
	contracts.EndPoint = ethEndPoint

	num := queryBalance("0x0eb5b66c31b3c5a12aae81a9d629540b6433cac6")
	fmt.Println("用于转账的账号余额:", num)

	//ethEndPoint = *qeth //用正常的链（http://39.100.146.21:8101）给新建账户转账
	userAddr, userSk, err := createAddr()
	if err != nil {
		log.Fatal("create user fails", err)
	}
	log.Println("create account success")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式

	log.Println("===============start test================")
	defer log.Println("==============finish test===============")

	//ethEndPoint = *eth //用不正常的链（http://47.92.5.51:8101）部署query合约
	log.Println("=====start set mapper addr=====")
	contracts.EndPoint = ethEndPoint
	addrSet, _, err := contracts.GetMapperFromAdminV1(localAddr, localAddr, "test", userSk, true)
	if err != nil {
		log.Fatal("set addr fails", err)
	}

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	addrGot, mapperInstance, err := contracts.GetMapperFromAdminV1(localAddr, localAddr, "test", userSk, false)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	if addrSet.String() != addrGot.String() {
		log.Fatal(addrSet.String(), "set different from got:", addrGot.String())
	}

	log.Println("=====start add addr first=====")
	contracts.EndPoint = ethEndPoint

	err = contracts.AddToMapper(localAddr, localAddr, userSk, mapperInstance)
	if err != nil {
		log.Fatal("set addr fails", err)
	}

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	aGot, err := contracts.GetAddrsFromMapper(localAddr, mapperInstance)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	le := len(aGot)

	if aGot[le-1].String() != localAddr.String() {
		log.Fatal(localAddr.String(), " set different from got:", aGot[le-1].String())
	}

	log.Println("=====start add addr second=====")
	contracts.EndPoint = ethEndPoint

	err = contracts.AddToMapper(localAddr, addrSet, userSk, mapperInstance)
	if err != nil {
		log.Fatal("set addr fails", err)
	}

	log.Println("=====start get addr from remote=====")
	contracts.EndPoint = qethEndPoint
	aGot, err = contracts.GetAddrsFromMapper(localAddr, mapperInstance)
	if err != nil {
		log.Fatal("got addr from remote fails: ", err)
	}

	le = len(aGot)

	if aGot[le-1].String() != addrSet.String() {
		log.Fatal(addrSet.String(), " set different from got:", aGot[le-1].String())
	}

	log.Println("=====test pass=====")
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
	skByteEth, err := utils.IPFSskToEthskByte(identity.PrivKey)
	if err != nil {
		return "", "", err
	}
	enc := make([]byte, len(skByteEth)*2)
	//对私钥进行十六进制编码，得到以太坊格式的私钥，此处不加上"0x"前缀
	hex.Encode(enc, skByteEth)

	var balance *big.Int
	balance = queryBalance(addressHex)
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		transferTo(big.NewInt(moneyTo), addressHex)
	}

	for i := 1; i <= 35; i++ {
		time.Sleep(time.Minute)
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
		return addressHex, string(enc), errors.New("转账失败")
	}

	return addressHex, string(enc), nil
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
