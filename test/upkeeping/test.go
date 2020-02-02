package main

import (
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	moneyTo = 1000000000000000
)

var serverKids = []string{"8MHS9fZzRaHNj4mP1kYDebwySmLzaw", "8MGRZbvn8caS431icB2P1uT74B3EHh", "8MJCzFbpXCvdfzmJy5L8jiw4w1qPdY", "8MKX58Ko5vBeJUkfgpkig53jZzwqoW", "8MHYzNkm6dF9SWU5u7Py8MJ31vJrzS", "8MK2saApPQMoNfVmnRDiApoAWFzo2K"}
var serverPids = []string{"8MHXst83NnSfYHnyqWMVjwjt2GiutV", "8MGrkL5cUpPsPbePvCfwCx6HemwDvy", "8MJ71X96BcnUNkhSFjc6CCsemL6nSQ", "8MGZ5nYsYw3Kmt8zC44W4V1NYaTGcE", "8MGhVo1ib6C6PmFhfQK4Hr3hHwQjC9", "8MJcdk2cyQvZknpxYf2AmGKDHRSRJP", "8MG9ZMYoZrZxjc7bVMeqJkaxAdb3Wx", "8MGqojupxiCesALno7sA73NhJkcSY5", "8MKAiRexSQG4SpGrpEQb4s9wjxJimX", "8MKU1DT94SB3aHTrMqWcJa2oLRtTzv", "8MJaFY7yAyYAvnjnM5hTbTfpjXhTHx", "8MGUGzCk1RUvq1aTPd9uuorrZ7FRhx", "8MHSARkgxWkjx5hKPm9vhX2v1VZ6GT"}

var ethEndPoint, qethEndPoint string

func main() {
	flag.String("testnet", "--eth=http://39.100.146.21:8101 --qeth=http://47.92.5.51:8101", "testnet commands")
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address;")
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	kCount := 3
	pCount := 5
	amount := big.NewInt(1230)

	contracts.EndPoint = ethEndPoint

	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails:", err)
	}

	test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)

	if err := ukTest(kCount, pCount, amount, userAddr, userSk); err != nil {
		log.Fatal(err)
	}
}

func ukTest(kCount int, pCount int, amount *big.Int, userAddr, userSk string) error {
	log.Println(">>>>>>>>>>>>>>>>>>>>>SmartContractTest>>>>>>>>>>>>>>>>>>>>>")
	defer log.Println("===================SmartContractTestEnd============================")

	localAddr := common.HexToAddress(userAddr[2:]) //将id转化成智能合约中的address格式
	mapKeeperAddr := make(map[common.Address]*big.Int)
	mapProviderAddr := make(map[common.Address]*big.Int)
	listKeeperAddr := []common.Address{localAddr}
	listProviderAddr := []common.Address{}
	mapKeeperAddr[localAddr] = test.QueryBalance(localAddr.String(), qethEndPoint)

	i := 0
	for _, serverKid := range serverKids { //得到keeper地址 并且查询初始余额
		tempAddr, _ := address.GetAddressFromID(serverKid)
		mapKeeperAddr[tempAddr] = test.QueryBalance(tempAddr.String(), qethEndPoint)
		listKeeperAddr = append(listKeeperAddr, tempAddr)
		if i++; i == kCount-1 {
			break
		}
	}
	i = 0
	for _, serverPid := range serverPids { //得到provider地址 并查询初始余额
		tempAddr, _ := address.GetAddressFromID(serverPid)
		mapProviderAddr[tempAddr] = test.QueryBalance(tempAddr.String(), qethEndPoint)
		listProviderAddr = append(listProviderAddr, tempAddr)
		if i++; i == pCount {
			break
		}
	}

	log.Println("begin to deploy upkeeping first")
	uAddr, err := contracts.DeployUpkeeping(userSk, localAddr, listKeeperAddr[0], listKeeperAddr, listProviderAddr, 10, 1024, 111, big.NewInt(234500), false)
	if err != nil {
		log.Println("deploy Upkeping err:", err)
		return err
	}

	log.Println("begin to reget upkeeping's addr")
	contracts.EndPoint = qethEndPoint
	ukaddr, _, err := contracts.GetUpkeeping(localAddr, localAddr, listKeeperAddr[0].String())
	if err != nil {
		log.Fatal("cannnot get upkeeping contract: ", err)
		return err
	}

	if uAddr.String() != ukaddr.String() {
		log.Fatal("set is different from get")
		return err
	}

	log.Println("begin to query upkeeping's balance")
	retryCount := 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		if amountUk.Cmp(big.NewInt(100)) > 0 {
			log.Println("contract balance", amountUk)
			if amountUk.Cmp(big.NewInt(234500)) != 0 {
				log.Fatal("Contract balance is not equal to preset: 234500")
			}

			amountLocal := test.QueryBalance(userAddr, qethEndPoint)
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

	log.Println("begin to initiate spacetime pay")
	contracts.EndPoint = ethEndPoint
	err = contracts.SpaceTimePay(ukaddr, listProviderAddr[0], userSk, amount)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	log.Println("spacetime pay trigger")

	log.Println("begin to query results of spacetime pay")
	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		if amountUk.Cmp(big.NewInt(234500)) < 0 {
			log.Println("keeper's balance change")
			for kAddr, amount := range mapKeeperAddr {
				amountNow := test.QueryBalance(kAddr.String(), qethEndPoint)
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
				amountNow := test.QueryBalance(pAddr.String(), qethEndPoint)
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

	log.Println("begin to test addProvider")
	contracts.EndPoint = ethEndPoint
	providerAddr, err := address.GetAddressFromID(serverPids[pCount])
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	err = contracts.AddProvider(userSk, localAddr, localAddr, []common.Address{providerAddr}, localAddr.String())
	if err != nil {
		log.Fatal("ukAddProvider AddProvider() error", err)
		return err
	}

	log.Println("upkeeping's tests pass")

	return nil
}
