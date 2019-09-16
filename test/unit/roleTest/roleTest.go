package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"flag"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	config "github.com/memoio/go-mefs/config"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	kpAddr  = "0x208649111Fd9253B76950e9f827a5A6dd616340d"
	kpSk    = "8f9eb151ffaebf2fe963e6185f0d1f8c1e8397e5905b616958d765e7753329ea"
	adminSk = "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4"
)

var ethEndPoint string

func main() {
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address")
	flag.Parse()
	ethEndPoint = *eth

	balance := queryBalance(kpAddr)
	if balance.Cmp(big.NewInt(1000000000000)) <= 0 {
		transferTo(big.NewInt(1000000000000), kpAddr)
	}

	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(kpAddr)
		if balance.Cmp(big.NewInt(1000000000000)) > 0 {
			break
		}

		log.Println(kpAddr, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}

	//if err := testDeploy(); err != nil {
	//	log.Fatal(err)
	//}

	if err := testRole(); err != nil {
		log.Fatal(err)
	}
}

func testDeploy() error {
	err := contracts.DeployKeeperProviderMap(adminSk)
	if err != nil {
		log.Println("DeployKeeperProviderMap err:", err)
		return err
	}

	time.Sleep(2 * time.Minute)

	return nil
}

func testRole() (err error) {
	log.Println("==========role=========")

	log.Println("test set keeper")

	keeperAddr := common.HexToAddress(kpAddr[2:])

	err = contracts.SetKeeper(keeperAddr, adminSk, true)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}

	retryCount := 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		res, err := contracts.IsKeeper(keeperAddr)
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set keeper fails")
			}
			continue
		}
		log.Println("set keeper success")
		break
	}

	log.Println("test set provider")
	proAddr, proSk, err := createAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	providerAddr := common.HexToAddress(proAddr[2:])

	err = contracts.SetProvider(providerAddr, adminSk, true)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		res, err := contracts.IsProvider(providerAddr)
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set provider fails")
			}
			continue
		}
		log.Println("set provider success")
		break
	}

	log.Println("test set add kp")

	err = contracts.AddKeeperProvidersToKPMap(keeperAddr, kpSk, keeperAddr, []common.Address{providerAddr})
	if err != nil {
		log.Println("Add Keeper Providers To KPMap err:", err)
		return err
	}

	log.Println("test get keeper from kpmap")
	retryCount = 0
	flag := false
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		kps, err := contracts.GetAllKeeperInKPMap(keeperAddr)
		if err != nil {
			if retryCount > 20 {
				log.Fatal("Get All Keeper In KPMap fails")
			}
			continue
		}

		flag := false
		for _, kc := range kps {
			if kc.String() == keeperAddr.String() {
				flag = true
				break
			}
		}

		if flag {
			log.Println("get keeper from kpmap success")
		} else {
			if retryCount > 20 {
				log.Fatal("Get Keeper fails")
			}
			continue
		}
		break
	}

	log.Println("test get provider from kpmap")
	retryCount = 0
	flag = false
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		pids, err := contracts.GetProviderInKPMap(providerAddr, keeperAddr)
		if err != nil {
			if retryCount > 20 {
				log.Fatal("Get Provider In KPMap fails")
			}
			continue
		}

		flag = false
		for _, pidr := range pids {
			if pidr.String() == providerAddr.String() {
				flag = true
				break
			}
		}

		if flag {
			log.Println("get provider from kpmap success")
		} else {
			if retryCount > 20 {
				log.Fatal("Get Provider fails")
			}
			continue
		}
		break
	}
	log.Println("add kp to kpmap success")

	log.Println("test delete provider from kpmap")

	err = contracts.DeleteProviderFromKPMap(providerAddr, proSk, keeperAddr, providerAddr)
	if err != nil {
		log.Fatal("delete provider from kpmap err:", err)
	}

	retryCount = 0
	flag = false
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		pids, err := contracts.GetProviderInKPMap(providerAddr, keeperAddr)
		if err != nil {
			if retryCount > 20 {
				log.Fatal("Get Provider In KPMap fails")
			}
			continue
		}

		flag = false
		for _, pidr := range pids {
			if pidr.String() == providerAddr.String() {
				flag = true
				break
			}
		}

		if flag {
			if retryCount > 20 {
				log.Fatal("Delete Provider In KPMap Fails")
			}
			continue
		} else {
			log.Println("delete provider from kpmap success")
		}
		break
	}

	log.Println("test delete keeper from kpmap")

	err = contracts.DeleteKeeperFromKPMap(providerAddr, proSk, keeperAddr)
	if err != nil {
		log.Fatal("delete keeper from kpmap err:", err)
	}

	retryCount = 0
	flag = false
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		kps, err := contracts.GetAllKeeperInKPMap(keeperAddr)
		if err != nil {
			if retryCount > 20 {
				log.Fatal("Get All Keeper In KPMap fails")
			}
			continue
		}

		flag = false
		for _, kc := range kps {
			if kc.String() == keeperAddr.String() {
				flag = true
				break
			}
		}

		if flag {
			if retryCount > 20 {
				log.Fatal("Delete Keeper In KPMap Fails")
			}
			continue
		} else {
			log.Println("delete keeper from kpmap success")
		}
		break
	}

	log.Println("test set keeper false")

	err = contracts.SetKeeper(keeperAddr, adminSk, false)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		res, err := contracts.IsKeeper(keeperAddr)
		if err != nil || res == true {
			if retryCount > 20 {
				log.Fatal("set keeper false fails")
			}
			continue
		}
		log.Println("set keeper fasle success")
		break
	}

	log.Println("test set provider false")
	err = contracts.SetProvider(providerAddr, adminSk, false)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		res, err := contracts.IsProvider(providerAddr)
		if err != nil || res == true {
			if retryCount > 20 {
				log.Fatal("set provider false fails")
			}
			continue
		}
		log.Println("set provider false success")
		break
	}

	return nil
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

	balance := queryBalance(addressHex)
	if balance.Cmp(big.NewInt(1000000000000)) < 0 {
		transferTo(big.NewInt(1000000000000), addressHex)
	}

	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(addressHex)
		if balance.Cmp(big.NewInt(1000000000000)) >= 0 {
			break
		}

		log.Println(addressHex, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}

	return addressHex, string(enc), nil
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
	client, err := ethclient.Dial(ethEndPoint)
	if err != nil {
		log.Println("rpc.Dial err", err)
		log.Fatal(err)
	}
	Address := common.HexToAddress(addr[2:])
	balance, err := client.PendingBalanceAt(context.Background(), Address)
	if err != nil {
		log.Fatal(err)
	}
	return balance
}
