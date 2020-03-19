package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/memoio/go-mefs/role"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	moneyTo   = 10000000000000000
	moneyToUK = 234500
)

var serverKaddrs = []string{"0x25a239c463415fF09767EDd051323385C9CE670c", "0xc67F94895F9626506857919D997e8dA7ffd95bF7", "0x9ADb6BC98FD4eE2bFF716034B9653dC5F0558B5f", "0xf904237239a79f535bdc77622CCfB31E3B3f83C9", "0x6Bd50cA3Ba83151f8Cb133B3C90737E173243adf", "0xd61E260aAA4AF3D64B899029E8c4025c96Ab31ec"}
var keeperSk = []string{"0xa7026c19010aa9fc55393d6efdcd5df3a5b08ccf2f0432af97093e7ed5a4282c", "0xba38f489b2ad7cf6220e9fd0e3166dd45639bac684cd9c1ef47c94ec416374d5"}
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
	amount := big.NewInt(1200)

	contracts.EndPoint = ethEndPoint

	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails:", err)
	}
	fmt.Println("userAddr:", userAddr)
	fmt.Println("userSk:", userSk)

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
	listKeeperSk := []string{userSk, keeperSk[0], keeperSk[1]}
	listProviderAddr := []common.Address{}
	mapKeeperAddr[localAddr] = test.QueryBalance(localAddr.String(), qethEndPoint)

	i := 0
	for _, serverKaddr := range serverKaddrs { //得到keeper地址 并且查询初始余额
		tempAddr := common.HexToAddress(serverKaddr)
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

	log.Println("1.begin to deploy upkeeping first")
	uAddr, err := contracts.DeployUpkeeping(userSk, localAddr, listKeeperAddr[0], listKeeperAddr, listProviderAddr, 600, 1024, 111, 90, big.NewInt(moneyToUK), false)
	if err != nil {
		log.Println("deploy Upkeping err:", err)
		return err
	}
	log.Println("upKeeping contract address:", uAddr.String())

	log.Println("2.begin to reget upkeeping's addr")
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

	log.Println("3.begin to query upkeeping's balance")
	retryCount := 0
	for {
		retryCount++
		time.Sleep(time.Minute)
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		if amountUk.Cmp(big.NewInt(100)) > 0 {
			log.Println("contract balance", amountUk)
			if amountUk.Cmp(big.NewInt(moneyToUK)) != 0 {
				log.Fatal("Contract balance is not equal to preset: ", moneyToUK)
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

	log.Println("4.begin to query upkeeping's information")
	contracts.EndPoint = ethEndPoint
	queryAddrGet, _, providers, timeGet, sizeG, priceG, createDate, endDate, cycle, needPay, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	createdate := big.NewInt(0)
	if (queryAddrGet != listKeeperAddr[0]) || (timeGet.Cmp(big.NewInt(600)) != 0) || (sizeG.Cmp(big.NewInt(1024)) != 0) || (priceG.Cmp(big.NewInt(111)) != 0) || (endDate.Cmp(createdate.Add(createDate, timeGet)) != 0) || (cycle.Cmp(big.NewInt(90)) != 0) || (needPay.Cmp(big.NewInt(0)) != 0) {
		log.Fatal("ukGetOrder get wrong parameters:", queryAddrGet.String(), " ", timeGet, sizeG, priceG, createDate, endDate, cycle, needPay)
	}

	log.Println("5.begin to first initiate spacetime pay , stLength is 30")
	if providers[0].Addr.String() != listProviderAddr[0].String() {
		log.Fatal("providers' order is wrong", providers)
	}
	stStart := providers[0].StEnd
	stLength := big.NewInt(30)
	merkleRoot := [32]byte{0}
	share := []int{4, 3, 3, 10} //keeper在本次支付中挑战的次数，share[kCount]代表挑战总次数
	signs, err := getSigs(listKeeperAddr, listKeeperSk, listProviderAddr[0], ukaddr, stStart, stLength, amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
	}
	err = contracts.SpaceTimePay(ukaddr, listProviderAddr[0], userSk, stStart, stLength, amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	log.Println("spacetime pay trigger")

	log.Println("6.begin to query results of first stPay")
	retryCount = 0
	for {
		if retryCount > 20 {
			log.Fatal("first stPay fails")
		}
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		log.Println("contract balance", amountUk)
		if amountUk.Cmp(big.NewInt(moneyToUK)) == 0 { //合约金额应该不变
			log.Println("contract balance not change")
			_, keepers, providers, _, _, _, _, _, _, needPay, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
			if err != nil {
				continue
			}
			if (len(providers[0].Money) == 1) && (providers[0].Money[0].Int64() == 120*9) && (len(keepers[1].Money) == 1) && (keepers[1].Money[0].Int64() == 36) && (needPay.Cmp(amount) == 0) && (providers[0].StEnd.Cmp(createdate.Add(createDate, stLength)) == 0) { //参数结果符合要求
				log.Println("parameters are right")
				//检查provider的余额变化
				amount := mapProviderAddr[listProviderAddr[0]]
				amountNow := test.QueryBalance(listProviderAddr[0].String(), ethEndPoint)
				amountCost := big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(listProviderAddr[0].String(), ":", amountCost)
				if amountCost.Cmp(big.NewInt(0)) == 0 {
					log.Println("provider balance not change")
				} else {
					continue
				}
				//检查keeper[1]的余额变化
				amount = mapKeeperAddr[listKeeperAddr[1]]
				amountNow = test.QueryBalance(listKeeperAddr[1].String(), ethEndPoint)
				amountCost = big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(listKeeperAddr[1].String(), ":", amountCost)
				if amountCost.Cmp(big.NewInt(0)) == 0 {
					log.Println("keeper[1] balance not change")
				} else {
					continue
				}
				break //all is right
			}
		}
	}

	log.Println("8.begin to second initiate spacetime pay, stLength is 100")
	_, _, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	stStart = providers[0].StEnd
	stLength = big.NewInt(100)
	merkleRoot = [32]byte{0}
	share = []int{4, 3, 3, 10} //keeper在本次支付中挑战的次数，share[kCount]代表挑战总次数
	signs, err = getSigs(listKeeperAddr, listKeeperSk, listProviderAddr[0], ukaddr, stStart, stLength, amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
	}
	err = contracts.SpaceTimePay(ukaddr, listProviderAddr[0], userSk, stStart, stLength, amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	log.Println("spacetime pay trigger")

	log.Println("9.begin to query results of second stPay")
	retryCount = 0
	for ; retryCount < 20; retryCount++ {
		retryCount++
		time.Sleep(30 * time.Second)
		_, keepers, providers, _, _, _, createDate, _, _, _, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
		if err != nil {
			continue
		}
		if (len(providers[0].Money) == 1) && (providers[0].Money[0].Int64() == 120*9*2) && (len(keepers[1].Money) == 1) && (keepers[1].Money[0].Int64() == 36*2) && (providers[0].StEnd.Cmp(createdate.Add(createDate, big.NewInt(130))) == 0) { //参数结果符合要求
			log.Println("parameters are right")
			break
		}
	}
	if retryCount == 20 {
		log.Fatal("second stPay fails")
	}

	log.Println("10.begin to test addProvider")
	providerAddr, err := address.GetAddressFromID(serverPids[pCount])
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}
	err = contracts.AddProvider(userSk, localAddr, localAddr, ukaddr, []common.Address{providerAddr})
	if err != nil {
		log.Fatal("ukAddProvider AddProvider() error", err)
		return err
	}

	//等待now > endDate,触发第三次时空支付
	for {
		nowTime := time.Now().Unix()
		if nowTime >= endDate.Int64() {
			break
		}
	}
	log.Println("11.begin to third initiate spacetime pay , stLength is 100")
	_, _, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	stStart = providers[0].StEnd
	stLength = big.NewInt(100)
	merkleRoot = [32]byte{0}
	share = []int{4, 3, 3, 10} //keeper在本次支付中挑战的次数，share[kCount]代表挑战总次数
	signs, err = getSigs(listKeeperAddr, listKeeperSk, listProviderAddr[0], ukaddr, stStart, stLength, amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
	}
	err = contracts.SpaceTimePay(ukaddr, listProviderAddr[0], userSk, stStart, stLength, amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	log.Println("spacetime pay trigger")

	log.Println("12.begin to query results of third stPay")
	retryCount = 0
	for {
		if retryCount > 20 {
			log.Fatal("third stPay fails")
		}
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		log.Println("contract balance", amountUk)
		if amountUk.Int64() == (moneyToUK - amount.Int64()*3) { //合约金额理应减少amount*3
			log.Println("contract balance reduce ", amount.Int64()*3)
			_, keepers, providers, _, _, _, _, _, _, needPay, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
			if err != nil {
				continue
			}
			if (len(providers[0].Money) == 2) && (providers[0].Money[1].Int64() == 120*9) && (len(keepers[1].Money) == 2) && (keepers[1].Money[1].Int64() == 36) && (needPay.Int64() == 0) && (providers[0].StEnd.Cmp(createdate.Add(createDate, big.NewInt(230))) == 0) { //参数结果符合要求
				log.Println("parameters are right")
				//检查provider的余额变化
				amount := mapProviderAddr[listProviderAddr[0]]
				amountNow := test.QueryBalance(listProviderAddr[0].String(), ethEndPoint)
				amountCost := big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(listProviderAddr[0].String(), ":", amountCost)
				if amountCost.Cmp(big.NewInt(120*9*3)) == 0 {
					log.Println("provider's balance increased 3240")
				} else {
					continue
				}
				//检查keeper[1]的余额变化
				amount = mapKeeperAddr[listKeeperAddr[1]]
				amountNow = test.QueryBalance(listKeeperAddr[1].String(), ethEndPoint)
				amountCost = big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(listKeeperAddr[1].String(), ":", amountCost)
				if amountCost.Cmp(big.NewInt(36*3)) == 0 {
					log.Println("keeper[1] balance increased 108")
				} else {
					continue
				}
				break //all is right
			}
		}
	}

	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		log.Println("contract balance", amountUk)
		if amountUk.Cmp(big.NewInt(moneyToUK)) < 0 {
			log.Println("keeper's balance change")
			for kAddr, amount := range mapKeeperAddr {
				amountNow := test.QueryBalance(kAddr.String(), qethEndPoint)
				amountCost := big.NewInt(0)
				amountCost.Sub(amountNow, amount)
				log.Println(kAddr.String(), ":", amountCost)
				if kAddr != localAddr {
					if amountCost.Cmp(big.NewInt(108)) < 0 {
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
				if listProviderAddr[0] == pAddr && amountCost.Cmp(big.NewInt(120*9*3)) < 0 {
					log.Fatal("provider gets wrong pay")
				}
			}
			break
		}

		if retryCount > 20 {
			log.Fatal("st pay fails")
		}
	}

	log.Println("9.begin to test extendTime")
	contracts.EndPoint = ethEndPoint
	addTime := int64(60)
	err = contracts.ExtendTime(userSk, localAddr, localAddr, localAddr.String(), addTime)
	if err != nil {
		log.Fatal("extend uk storage time error", err)
	}
	_, _, _, timeNewGet, _, _, _, _, _, _, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	if timeNewGet.Cmp(timeGet.Add(timeGet, big.NewInt(addTime))) != 0 {
		log.Fatal("storage time extended is not right", err)
	}

	log.Println("upkeeping's tests pass")

	return nil
}

func getSigs(keeperAddress []common.Address, keeperSk []string, providerAddr, upKeepingAddr common.Address, stStart, stLength, stValue *big.Int, merkleRoot [32]byte, share []int) ([][]byte, error) {
	sigs := [][]byte{}
	for i := 0; i < len(keeperAddress); i++ {
		sig, err := role.SignForStPay(upKeepingAddr, providerAddr, keeperSk[i], stStart, stLength, stValue, merkleRoot, share)
		if err != nil {
			log.Println("signForstPay error:", err)
			return sigs, err
		}
		sigs = append(sigs, sig)
	}
	return sigs, nil
}
