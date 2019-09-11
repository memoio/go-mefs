package main

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	userAddr = "0x208649111Fd9253B76950e9f827a5A6dd616340d"
	userSk   = "8f9eb151ffaebf2fe963e6185f0d1f8c1e8397e5905b616958d765e7753329ea"
)

var serverKids = []string{"8MHS9fZzRaHNj4mP1kYDebwySmLzaw", "8MGRZbvn8caS431icB2P1uT74B3EHh", "8MJCzFbpXCvdfzmJy5L8jiw4w1qPdY", "8MKX58Ko5vBeJUkfgpkig53jZzwqoW", "8MHYzNkm6dF9SWU5u7Py8MJ31vJrzS", "8MK2saApPQMoNfVmnRDiApoAWFzo2K"}
var serverPids = []string{"8MHXst83NnSfYHnyqWMVjwjt2GiutV", "8MGrkL5cUpPsPbePvCfwCx6HemwDvy", "8MJ71X96BcnUNkhSFjc6CCsemL6nSQ", "8MGZ5nYsYw3Kmt8zC44W4V1NYaTGcE", "8MGhVo1ib6C6PmFhfQK4Hr3hHwQjC9", "8MJcdk2cyQvZknpxYf2AmGKDHRSRJP", "8MG9ZMYoZrZxjc7bVMeqJkaxAdb3Wx", "8MGqojupxiCesALno7sA73NhJkcSY5", "8MKAiRexSQG4SpGrpEQb4s9wjxJimX", "8MKU1DT94SB3aHTrMqWcJa2oLRtTzv", "8MJaFY7yAyYAvnjnM5hTbTfpjXhTHx", "8MGUGzCk1RUvq1aTPd9uuorrZ7FRhx", "8MHSARkgxWkjx5hKPm9vhX2v1VZ6GT"}

const ethEndPoint = "http://212.64.28.207:8101"

func main() {
	kCount := 3
	pCount := 5
	amount := big.NewInt(1230)

	balance := queryBalance(userAddr)
	if balance.Cmp(big.NewInt(10000000000)) <= 0 {
		transferTo(big.NewInt(10000000000), userAddr)
	}

	for {
		time.Sleep(30 * time.Second)
		balance := queryBalance(userAddr)
		if balance.Cmp(big.NewInt(10000000000)) > 0 {
			break
		}

		log.Println(userAddr, "'s Balance now:", balance.String(), ", waiting for transfer success")
	}

	if err := SmartContractTest(kCount, pCount, amount); err != nil {
		log.Fatal(err)
	}
}

func SmartContractTest(kCount int, pCount int, amount *big.Int) error {
	log.Println(">>>>>>>>>>>>>>>>>>>>>SmartContractTest>>>>>>>>>>>>>>>>>>>>>")
	defer log.Println("===================SmartContractTestEnd============================")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式
	mapKeeperAddr := make(map[common.Address]*big.Int)
	mapProviderAddr := make(map[common.Address]*big.Int)
	listKeeperAddr := []common.Address{localAddr}
	listProviderAddr := []common.Address{}
	mapKeeperAddr[localAddr] = queryBalance(localAddr.String())

	i := 0
	for _, serverKid := range serverKids { //得到keeper地址 并且查询初始余额
		tempAddr, _ := address.GetAddressFromID(serverKid)
		mapKeeperAddr[tempAddr] = queryBalance(tempAddr.String())
		listKeeperAddr = append(listKeeperAddr, tempAddr)
		if i++; i == kCount-1 {
			break
		}
	}
	i = 0
	for _, serverPid := range serverPids { //得到provider地址 并查询初始余额
		tempAddr, _ := address.GetAddressFromID(serverPid)
		mapProviderAddr[tempAddr] = queryBalance(tempAddr.String())
		listProviderAddr = append(listProviderAddr, tempAddr)
		if i++; i == pCount {
			break
		}
	}

	key, err := crypto.HexToECDSA(userSk)
	if err != nil {
		log.Fatal("HexToECDSAErr:", err)
		return err
	}
	auth := bind.NewKeyedTransactor(key)
	auth.Value = big.NewInt(234500)
	auth.GasPrice = big.NewInt(100)

	client := contracts.GetClient(ethEndPoint)
	ukaddr, trans, uk, err := upKeeping.DeployUpKeeping(auth, client, localAddr, listKeeperAddr, listProviderAddr, big.NewInt(10), big.NewInt(1024), big.NewInt(111))
	if err != nil {
		log.Println("deploy Upkeping err:", err)
		return err
	}
	log.Println("depoly upkepping success, contract addr: ", ukaddr.Hex(), ", trans hash: ", trans.Hash().String())

	log.Println("begin to query upkeeping's balance")
	retryCount := 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := queryBalance(ukaddr.String())
		if amountUk.Cmp(big.NewInt(100)) > 0 {
			log.Println("contract balance", amountUk)
			if amountUk.Cmp(big.NewInt(234500)) != 0 {
				log.Fatal("Contract balance is not equal to preset: 234500")
			}

			amountLocal := queryBalance(userAddr)
			amountCost := big.NewInt(0)
			amountCost.Sub(amountLocal, mapKeeperAddr[localAddr])
			log.Println("user balance change due to deploy：", amountCost)
			mapKeeperAddr[localAddr] = amountLocal
			break
		}
		if retryCount > 20 {
			log.Fatal("Upkeeping has no balance")
		}
	}

	log.Println("begin to query upkeeping's information")

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)

		item, err := contracts.GetUpkeepingInfo(localAddr, uk)
		if err != nil {
			if retryCount > 20 {
				log.Fatal("Upkeeping has no information, err: ", err)
				break
			}
			continue
		}

		if item.Duration != int64(10) {
			log.Fatal("Contract duration", item.Duration, " is not equal to preset: 10")
		}

		if item.Capacity != int64(1024) {
			log.Fatal("Contract duration ", item.Capacity, " is not equal to preset: 1024")
		}

		if item.Price != int64(111) {
			log.Fatal("Contract price ", item.Price, " is not equal to preset: 111")
		}

		knum := 0

		for _, kp := range listKeeperAddr {
			kid, _ := address.GetIDFromAddress(kp.String())
			for _, keeper := range item.KeeperIDs {
				if kid == keeper {
					knum++
				}
			}
		}

		if knum != kCount {
			log.Fatal("Contract keeper count is not equal to preset: ", kCount+1)
		}

		pnum := 0

		for _, kp := range listProviderAddr {
			kid, _ := address.GetIDFromAddress(kp.String())
			for _, keeper := range item.ProviderIDs {
				if kid == keeper {
					pnum++
				}
			}
		}

		if pnum != pCount {
			log.Fatal("Contract provider count is not equal to preset: ", pCount)
		}

		log.Println("upkeeping's information is right")
		break
	}

	log.Println("begin to initiate spacetime pay")
	err = contracts.SpaceTimePay(uk, localAddr, listProviderAddr[0], userSk, amount)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	log.Println("spacetime pay complete")

	log.Println("begin to query results of spacetime pay")

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := queryBalance(ukaddr.String())
		if amountUk.Cmp(big.NewInt(234500)) < 0 {
			log.Println("keeper's balance change")
			for kAddr, amount := range mapKeeperAddr {
				amountNow := queryBalance(kAddr.String())
				amountCost := big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(kAddr.String(), ":", amountCost)
				if kAddr != localAddr {
					if amountCost.Cmp(big.NewInt(41)) < 0 {
						log.Fatal("keeper gets wrong pay")
					}
				}

			}

			log.Println("provider's balance change")
			for pAddr, amount := range mapProviderAddr {
				amountNow := queryBalance(pAddr.String())
				amountCost := big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(pAddr.String(), ":", amountCost)
				if listProviderAddr[0] == pAddr && amountCost.Cmp(big.NewInt(123*9)) < 0 {
					log.Fatal("provider gets wrong pay")
				}
			}
			break
		}

		if retryCount > 20 {
			log.Fatal("st pay fails")
		}
	}

	return nil
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
