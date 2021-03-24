package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	moneyTo      = 10000000000000000
	moneyToUK    = 234500
	defaultCycle = 300
	sDuration    = 2000
	sSize        = 1024
	sPrice       = 111
	sLength      = 290
	perMoney     = 120
)

var (
	amount = big.NewInt(10 * perMoney)
	kCount = 3
	pCount = 5
)

//7个 + 7
var serverKaddrs = []string{"0x8c4d5d6d57574Ef7Bb21BD369969185BF781cBCD",
	"0x25a239c463415fF09767EDd051323385C9CE670c", "0xc67F94895F9626506857919D997e8dA7ffd95bF7", "0x615ef0593E5Ac607d4e9097675d7A7eF492A884a", "0xc2c74adab42496d52b9dfdd77bd1a331b360014d", "0x877c9831072D176BE63024e8482B9C75678d770F", "0x2272274F5322df7BF6573038A7c9e6404A133aDc",
	"0x972e5Cc7FdcAb6B07553Afe2182DD9C5fDa04565", "0x55198c7EEffcFf491880A2Ab48D2f94cf631C9B7",
	"0x9076AE4bcee2EB7D99C2165AF0d0dD3Ea693Dd58", "0x3047C99CC029a9E36AeBf8Ac6aE01b835141f23b", "0xBf33F2106b21dbD8a0D9be0Ad393019fC09DE320", "0x63698bD23246DB27a5Fe17CB7B7239A8E762da60", "0x5EAf9be454c08E8f2466CCD6573823d44037AcB8"}

var keeperSk = []string{"25e5246b92c190dbf993ae4eeb1d3a27133d1ad3ed8109e4593bde81fa7451b0",
	"a7026c19010aa9fc55393d6efdcd5df3a5b08ccf2f0432af97093e7ed5a4282c", "ba38f489b2ad7cf6220e9fd0e3166dd45639bac684cd9c1ef47c94ec416374d5", "3dccc64bb83c0cf37a15ae2636a3437d8f095321a4c3b7cf475074179058b90f", "bda6fde734f83e84631969b559a0ccfa7cb1b09665dd8145d6e66ccea19cf6e2", "bcca3d97fef5ce3599f5ae592a8aa010fc56339ef15b7d355cc8d4e59ce86e96", "81dfac254a4ae3b9d489a74acd1492d0dc42c948d0d2e2c5f355e936f3ff0803",
	"60e003d7b5bb6a358bd1e41bba7aa75e2dce11fe92254630713511ece0ebbd5b", "9cff834cdb325d9b476579e31d2edf400d23ed83465fef47cf2465ade25d7cdc",
	"f680f12645de0e50ca4374c9607d2a52d1f8a2c5e9337186b0738129d61a8f26", "4ef3aa1b2476f3f0aaca4480493475e6411ba988083ee1533b5bf96102e244e5", "6948f591408d1618db14ad231a033ee46504f3c41f5c37cc81fcc54589fb4be0", "5ee75e4ddec0faae19ccddc0f00cb6b4eb07d4f2a2e97c41a1e8aebc80e5b555", "86013431bdee6263d05f434db474baac1849de62a4bcb3f007ee1d05264dbd0a"}

//13个
var serverPids = []string{"8MHXst83NnSfYHnyqWMVjwjt2GiutV", "8MGrkL5cUpPsPbePvCfwCx6HemwDvy", "8MJ71X96BcnUNkhSFjc6CCsemL6nSQ", "8MGZ5nYsYw3Kmt8zC44W4V1NYaTGcE", "8MGhVo1ib6C6PmFhfQK4Hr3hHwQjC9", "8MJcdk2cyQvZknpxYf2AmGKDHRSRJP", "8MG9ZMYoZrZxjc7bVMeqJkaxAdb3Wx", "8MGqojupxiCesALno7sA73NhJkcSY5", "8MKAiRexSQG4SpGrpEQb4s9wjxJimX", "8MKU1DT94SB3aHTrMqWcJa2oLRtTzv", "8MJaFY7yAyYAvnjnM5hTbTfpjXhTHx", "8MGUGzCk1RUvq1aTPd9uuorrZ7FRhx", "8MHSARkgxWkjx5hKPm9vhX2v1VZ6GT"}

var ethEndPoint, qethEndPoint string
var success bool

func main() {
	utils.StartLogger()
	flag.String("testnet", "--eth=http://119.147.213.219:8101 --qeth=http://119.147.213.219:8101", "testnet commands")
	eth := flag.String("eth", "http://119.147.213.219:8101", "eth api address;")
	qeth := flag.String("qeth", "http://119.147.213.219:8101", "eth api address;")
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	contracts.EndPoint = ethEndPoint

	if err := ukTest(); err != nil {
		log.Fatal(err)
	}
}

func ukTest() error {
	log.Println(">>>>>>>>>>>>>>>>>>>>>SmartContractTest>>>>>>>>>>>>>>>>>>>>>")
	defer log.Println("===================SmartContractTestEnd============================")

	kAddrList := []common.Address{}
	kBalanceMap := make(map[common.Address]*big.Int)
	var kSkList []string

	i := 0
	for _, serverKaddr := range serverKaddrs { //得到keeper地址 并且查询初始余额
		tempAddr := common.HexToAddress(serverKaddr)
		kBalanceMap[tempAddr] = test.QueryBalance(tempAddr.String(), qethEndPoint)
		kAddrList = append(kAddrList, tempAddr)
		kSkList = append(kSkList, keeperSk[i])
		if i++; i == kCount {
			break
		}
	}

	if kBalanceMap[kAddrList[0]].Cmp(big.NewInt(moneyTo)) < 0 {
		test.TransferTo(big.NewInt(moneyTo), kAddrList[0].String(), ethEndPoint, qethEndPoint)
	}

	log.Println("kSkList:", kSkList[0])

	pAddrList := []common.Address{}
	pBalanceMap := make(map[common.Address]*big.Int)
	i = 0
	for _, serverPid := range serverPids { //得到provider地址 并查询初始余额
		tempAddr, _ := address.GetAddressFromID(serverPid)
		pBalanceMap[tempAddr] = test.QueryBalance(tempAddr.String(), qethEndPoint)
		pAddrList = append(pAddrList, tempAddr)
		if i++; i == pCount {
			break
		}
	}

	oldBlock, err := contracts.GetLatestBlock()
	if err != nil {
		log.Fatal("getLatestBlock fails:", err)
		return err
	}

	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create user fails:", err)
		return err
	}
	fmt.Println("userAddr:", userAddr, ", userSk:", userSk)
	localAddr := common.HexToAddress(userAddr[2:])

	test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)

	log.Println("1.begin to deploy upkeeping first")
	contracts.EndPoint = ethEndPoint
	uAddr, err := contracts.DeployUpkeeping(userSk, localAddr, localAddr, kAddrList, pAddrList, sDuration, sSize, big.NewInt(sPrice), defaultCycle, big.NewInt(moneyToUK), false)
	if err != nil {
		log.Println("deploy Upkeping err:", err)
		return err
	}
	log.Println("upKeeping contract address:", uAddr.String())

	log.Println("2.begin to reget upkeeping's addr")
	contracts.EndPoint = qethEndPoint
	ukaddr, _, err := contracts.GetUpkeeping(localAddr, localAddr, localAddr.String())
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
		amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
		if amountUk.Sign() > 0 {
			log.Println("contract balance", amountUk)
			if amountUk.Cmp(big.NewInt(moneyToUK)) != 0 {
				log.Fatal("Contract balance is not equal to preset: ", moneyToUK)
			}

			break
		}

		if retryCount > 20 {
			log.Fatal("Upkeeping has no balance")
		}

		time.Sleep(10 * time.Second)
	}

	log.Println("4.begin to query upkeeping's information")
	queryAddrGet, _, providers, timeGet, sizeG, priceG, createDate, endDate, cycle, needPay, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
		return err
	}

	createdate := big.NewInt(0)
	if (queryAddrGet != localAddr) || (timeGet.Cmp(big.NewInt(sDuration)) != 0) || (sizeG.Cmp(big.NewInt(sSize)) != 0) || (priceG.Cmp(big.NewInt(sPrice)) != 0) || (endDate.Cmp(createdate.Add(createDate, timeGet)) != 0) || (cycle.Cmp(big.NewInt(defaultCycle)) != 0) || (needPay.Cmp(big.NewInt(0)) != 0) {
		log.Fatal("uk get wrong parameters:", queryAddrGet.String(), timeGet, sizeG, priceG, createDate, endDate, cycle, needPay)
	}

	if providers[0].Addr.String() != pAddrList[0].String() {
		log.Fatal("providers' order is wrong", providers)
	}

	log.Println("5.begin to first initiate spacetime pay to provider 0, stLength is:", sLength)

	stStart := providers[0].StEnd
	stLength := big.NewInt(sLength)
	merkleRoot := [32]byte{0}
	share := []int64{4, 3, 3, 10}
	signs, err := getSigs(kAddrList, kSkList, pAddrList[0], ukaddr, stStart, stLength, amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
		return err
	}

	err = contracts.SpaceTimePay(ukaddr, pAddrList[0], kSkList[0], stStart, stLength, amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}

	firstNow := time.Now().Unix()
	threeCycle := createDate.Int64() + 3*cycle.Int64()
	log.Println("spacetime pay trigger", "nowTime:", firstNow, "createDate+3*cycle:", threeCycle)

	log.Println("6.begin to query results of first stPay")
	amountUk := test.QueryBalance(ukaddr.String(), qethEndPoint)
	log.Println("contract balance", amountUk)
	_, keepers, providers, _, _, _, _, _, _, needPay, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("getOrder err: ", err)
	}
	log.Println(keepers)
	log.Println(providers)
	log.Println(needPay)
	if amountUk.Cmp(big.NewInt(moneyToUK)) == 0 { //合约金额不变,时间未超过startTime+3*cycle,没有真实支付
		log.Println("contract balance not change")
		if (len(providers[0].Money) == 1) && (providers[0].Money[0].Int64() == perMoney*9) && (len(keepers[1].Money) == 1) && (keepers[1].Money[0].Int64() == perMoney*3/10) && (needPay.Cmp(amount) == 0) && (providers[0].StEnd.Cmp(createdate.Add(createDate, stLength)) == 0) {
			log.Println("parameters are right")
			success = true
		}
	} else if firstNow > threeCycle && amountUk.Cmp(big.NewInt(moneyToUK)) < 0 { //触发了真实支付
		log.Println("contract balance reduce ", moneyToUK-amountUk.Int64())
		if (len(providers[0].Money) == 1) && (providers[0].Money[0].Int64() == perMoney*9) && (len(keepers[1].Money) == 1) && (keepers[1].Money[0].Int64() == perMoney*3/10) && (needPay.Int64() == 0) && (providers[0].StEnd.Cmp(createdate.Add(createDate, stLength)) == 0) {
			log.Println("parameters are right")
			success = true
		}
	}

	if !success {
		log.Fatal("parameters not right")
	}
	success = false
	//检查provider的余额变化
	money := new(big.Int)
	amountNow := test.QueryBalance(pAddrList[0].String(), ethEndPoint)
	log.Println("provider[0] balance increased:", money.Sub(amountNow, pBalanceMap[pAddrList[0]]))
	//检查keeper[1]的余额变化
	amountNow = test.QueryBalance(kAddrList[1].String(), ethEndPoint)
	log.Println("keeper[1] balance increased:", money.Sub(amountNow, kBalanceMap[kAddrList[1]]))

	log.Println("7. begin to test setProviderStop to stop provider 1")
	contracts.EndPoint = ethEndPoint
	setStopSigns, err := getSetStopSigns(kAddrList, kSkList, pAddrList[1], ukaddr)
	if err != nil {
		log.Fatal("get setStopSigns error:", err)
	}
	invalidAddr := common.HexToAddress(contracts.InvalidAddr)
	err = contracts.SetProviderStop(kSkList[0], kAddrList[0], localAddr, pAddrList[1], invalidAddr, localAddr.String(), setStopSigns)
	if err != nil {
		log.Fatal("set provider stop fails: ", err)
	}
	_, _, providerInfo, _, _, _, _, _, _, _, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	if len(providerInfo) != pCount || !providerInfo[1].Stop {
		log.Fatal("provider is not stopped, error")
	}

	log.Println("8. begin to test stPay for stopped provider 1, stLength is:", sLength)
	stStart = providerInfo[1].StEnd
	if stStart.Cmp(createDate) != 0 {
		log.Fatal("provider[1] stEnd is wrong, ", providerInfo)
	}
	merkleRoot = [32]byte{0}
	share = []int64{4, 3, 3, 10}
	signs, err = getSigs(kAddrList, kSkList, pAddrList[1], ukaddr, stStart, big.NewInt(sLength), amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
	}
	err = contracts.SpaceTimePay(ukaddr, pAddrList[1], kSkList[0], stStart, big.NewInt(sLength), amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}

	stoppedNow := time.Now().Unix()
	log.Println("spacetime pay trigger", "nowTime:", stoppedNow, "createDate+4*cycle:", threeCycle+cycle.Int64())

	log.Println("9.begin to query results of stPay for stopped provider 1")
	amountUk = test.QueryBalance(ukaddr.String(), qethEndPoint)
	log.Println("contract balance", amountUk)
	_, keepers, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("getOrder err:", err)
	}
	log.Println(keepers)
	log.Println(providers)
	log.Println(needPay)
	if (len(providers[1].Money) == 1) && (providers[1].Money[0].Int64() == perMoney*9 && (len(keepers[1].Money) == 1) && (keepers[1].Money[0].Int64() == perMoney*3*2/10) && (providers[1].StEnd.Cmp(createdate.Add(createDate, stLength)) == 0)) {
		log.Println("not pay, parameters are right")
		//检查provider[1]的余额变化
		amountNow := test.QueryBalance(pAddrList[1].String(), ethEndPoint)
		if pBalanceMap[pAddrList[1]].Cmp(amountNow) == 0 {
			log.Println("provider[1] balance not change")
			success = true
		}
	} else if (firstNow > threeCycle) && (amountUk.Int64() == moneyToUK-perMoney*11) { //触发两次支付
		if (len(providers[1].Money) == 1) && (providers[1].Money[0].Int64() == perMoney*9 && (len(keepers[1].Money) == 2) && (keepers[1].Money[1].Int64() == perMoney*3/10) && (providers[1].StEnd.Cmp(createdate.Add(createDate, stLength)) == 0)) {
			log.Println("pay twice, parameters are right")
			amountNow := test.QueryBalance(pAddrList[1].String(), ethEndPoint)
			if pBalanceMap[pAddrList[1]].Cmp(amountNow) == 0 {
				log.Println("provider[1] balance not change")
				success = true
			}
		}
	} else if (firstNow < threeCycle+cycle.Int64()) && (amountUk.Int64() == moneyToUK-perMoney*10) { //只触发第一次对provider0的支付
		if (len(providers[1].Money) == 1) && (providers[1].Money[0].Int64() == perMoney*9 && (len(keepers[1].Money) == 2) && (keepers[1].Money[1].Int64() == perMoney*3/10) && (providers[1].StEnd.Cmp(createdate.Add(createDate, stLength)) == 0)) {
			log.Println("pay once, parameters are right")
			amountNow := test.QueryBalance(pAddrList[1].String(), ethEndPoint)
			if pBalanceMap[pAddrList[1]].Cmp(amountNow) == 0 {
				log.Println("provider[1] balance not change")
				success = true
			}
		}
	}
	if !success {
		log.Fatal("parameters not roight")
	}
	success = false

	log.Println("10.begin to test addProvider by user")
	providerAddr1, err := address.GetAddressFromID(serverPids[pCount])
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	var esigs [][]byte
	err = contracts.AddProvider(userSk, localAddr, ukaddr, []common.Address{providerAddr1}, esigs)
	if err != nil {
		log.Fatal("ukAddProvider user AddProvider() error", err)
		return err
	}

	log.Println("10.begin to test addProvider by keeper")

	providerAddr, err := address.GetAddressFromID(serverPids[pCount+1])
	if err != nil {
		log.Println("ukAddProvider GetAddressFromID() error", err)
		return err
	}

	addProviderAddrs := []common.Address{providerAddr}
	sigs, err := getAddProviderSigs(kAddrList, kSkList, ukaddr, addProviderAddrs)
	if err != nil {
		log.Println("ukAddProvider getsigs error", err)
		return err
	}

	err = contracts.AddProvider(kSkList[0], kAddrList[0], ukaddr, addProviderAddrs, sigs)
	if err != nil {
		log.Fatal("ukAddProvider AddProvider() error", err)
		return err
	}

	_, _, providers, _, _, _, _, _, _, _, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("getUk err:", err)
	}

	has := false
	for _, pinfo := range providers {
		if pinfo.Addr.String() == providerAddr.String() {
			has = true
			break
		}
	}

	if !has {
		log.Fatal("add provider failed")
	}

	log.Println("wait third cycle")
	//等待now > endDate + 60,触发第三次时空支付
	for {
		nowTime := time.Now().Unix()
		if nowTime >= createDate.Int64()+3*defaultCycle+60 {
			break
		}
		time.Sleep(time.Duration(createDate.Int64()+3*defaultCycle+60-nowTime) * time.Second)
	}

	log.Println("11.begin to second initiate spacetime pay to provider 0 after 3 cycles, stLength is: ", sLength)
	_, _, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	stStart = providers[0].StEnd
	stLength = big.NewInt(sLength)
	merkleRoot = [32]byte{0}
	share = []int64{4, 3, 3, 10} //keeper在本次支付中挑战的次数，share[kCount]代表挑战总次数
	signs, err = getSigs(kAddrList, kSkList, pAddrList[0], ukaddr, stStart, stLength, amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
	}
	err = contracts.SpaceTimePay(ukaddr, pAddrList[0], kSkList[0], stStart, stLength, amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	secondNow := time.Now().Unix()
	log.Println("spacetime pay trigger", "nowTime:", secondNow)

	log.Println("12.begin to query results of second stPay")
	amountUk = test.QueryBalance(ukaddr.String(), qethEndPoint)
	log.Println("contract balance", amountUk)
	_, keepers, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("get order fails")
	}
	log.Println(keepers)
	log.Println(providers)
	log.Println(needPay)
	if amountUk.Int64() < moneyToUK {
		if (len(providers[0].Money) == 2) && (providers[0].Money[1].Int64() == perMoney*9) && (providers[0].StEnd.Cmp(createdate.Add(createDate, big.NewInt(sLength*2))) == 0) { //参数结果符合要求
			log.Println("parameters are right")
			success = true
		} else if (len(providers[0].Money) == 1) && (providers[0].Money[0].Int64() == perMoney*18) && (providers[0].StEnd.Cmp(createdate.Add(createDate, big.NewInt(sLength*2))) == 0) {
			log.Println("parameters are right")
			success = true
		}
	}
	if !success {
		log.Fatal("parameters not right")
	}
	success = false

	log.Println("wait enddate")
	//等待now > endDate + 60,触发第三次时空支付
	for {
		nowTime := time.Now().Unix()
		if nowTime >= endDate.Int64()+60 {
			break
		}
		time.Sleep(time.Duration(endDate.Int64()+60-nowTime) * time.Second)
	}
	fmt.Println("nowTime:", time.Now().Unix(), "endDate:", endDate.Int64())

	log.Println("13.begin to third initiate spacetime pay to provider 0, stLength is: ", sLength)
	_, _, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	stStart = providers[0].StEnd
	stLength = big.NewInt(sLength)
	merkleRoot = [32]byte{0}
	share = []int64{4, 3, 3, 10} //keeper在本次支付中挑战的次数，share[kCount]代表挑战总次数
	signs, err = getSigs(kAddrList, kSkList, pAddrList[0], ukaddr, stStart, stLength, amount, merkleRoot, share)
	if err != nil {
		log.Fatal("getSigs error:", err)
	}
	err = contracts.SpaceTimePay(ukaddr, pAddrList[0], kSkList[0], stStart, stLength, amount, merkleRoot, share, signs)
	if err != nil {
		log.Fatal("spacetime pay err:", err)
		return err
	}
	log.Println("spacetime pay trigger")

	log.Println("14.begin to query results of third stPay")
	amountUk = test.QueryBalance(ukaddr.String(), qethEndPoint)
	log.Println("contract balance", amountUk)
	_, keepers, providers, _, _, _, _, _, _, needPay, _, err = contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("get order fails")
	}
	log.Println(keepers)
	log.Println(providers)
	log.Println(needPay)
	if amountUk.Int64() == (moneyToUK - amount.Int64()*3 - perMoney) { //合约金额理应减少amount*3(三次时空支付pro[0])和amount/10(一次时空支付pro[1],只有keeper拿到了)
		log.Println("contract balance reduce ", amount.Int64()*3+perMoney)

		amountp := pBalanceMap[pAddrList[0]]
		amountpNow := test.QueryBalance(pAddrList[0].String(), ethEndPoint)
		amountpCost := big.NewInt(0)
		amountpCost.Sub(amountpNow, amountp)
		log.Println(pAddrList[0].String(), ":", amountpCost)

		//检查keeper[1]的余额变化
		amountk := kBalanceMap[kAddrList[1]]
		amountkNow := test.QueryBalance(kAddrList[1].String(), ethEndPoint)
		amountkCost := big.NewInt(0)
		amountkCost.Sub(amountkNow, amountk)
		log.Println(kAddrList[1].String(), ":", amountkCost)

		if amountpCost.Cmp(big.NewInt(perMoney*9*3)) == 0 && amountkCost.Cmp(big.NewInt(perMoney*3*4/10)) == 0 {
			log.Println("provider0's balance increased 3240")
			log.Println("keeper[1] balance increased 144")
			success = true
		}
	}
	if !success {
		log.Fatal("parameters not right")
	}

	log.Println("15. begin to query stopped provider's balance")
	//检查provider[1]的余额变化
	amountBefore := pBalanceMap[pAddrList[1]]
	amountNow = test.QueryBalance(pAddrList[1].String(), ethEndPoint)
	amountCost := big.NewInt(0)
	amountCost.Sub(amountNow, amountBefore)
	log.Println(pAddrList[1].String(), ":", amountCost)
	if amountCost.Cmp(big.NewInt(0)) == 0 {
		log.Println("stopped provider's balance not change")
	} else {
		log.Fatal("error: stopped provider's balance changed")
	}

	log.Println("16.begin to test extendTime")
	contracts.EndPoint = ethEndPoint
	addTime := time.Now().Unix() - endDate.Int64() + int64(300)
	err = contracts.ExtendTime(userSk, localAddr, localAddr, localAddr.String(), addTime)
	if err != nil {
		log.Fatal("extend uk storage time error", err)
	}
	_, _, _, timeNewGet, _, _, _, endDateNew, _, _, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	if timeNewGet.Cmp(timeGet.Add(timeGet, big.NewInt(addTime))) != 0 || endDateNew.Int64() != (endDate.Int64()+addTime) {
		log.Fatal("storage time extended is not right", err)
	}

	//set kAddrList[1] stop
	log.Println("17. begin to test setKeeperStop")
	contracts.EndPoint = ethEndPoint
	setStopSigns, err = getSetStopSigns(kAddrList, kSkList, kAddrList[1], ukaddr)
	if err != nil {
		log.Fatal("get setStopSigns error:", err)
	}
	err = contracts.SetKeeperStop(kSkList[0], kAddrList[0], localAddr, kAddrList[1], localAddr.String(), setStopSigns)
	if err != nil {
		log.Fatal("set keeper stop fails: ", err)
	}
	_, keeperInfo, _, _, _, _, _, _, _, _, _, err := contracts.GetOrder(userSk, localAddr, localAddr, localAddr.String())
	if err != nil {
		log.Fatal("ukGetOrder error:", err)
	}
	if len(keeperInfo) != kCount || !keeperInfo[1].Stop {
		log.Fatal("keeper is not stopped, error")
	}

	newblk, err := contracts.GetLatestBlock()
	if err != nil {
		panic(err)
	}

	//query provider[0]'s income
	log.Println("18. begin to test getStorageIncome")
	upAddrs := []common.Address{uAddr}

	total, _, err := contracts.GetStorageIncome(upAddrs, pAddrList[0], oldBlock.Number().Int64(), newblk.Number().Int64())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("totalIncome: ", total.String())
	if total.Cmp(big.NewInt(perMoney*9*3)) != 0 {
		log.Fatal("query providers[0]'s storageIncome failed, it is not equal to 3240")
	}

	//query keeper[0]'s income
	log.Println("19. begin to test getManageIncome")
	total, _, err = contracts.GetStorageIncome(upAddrs, kAddrList[0], oldBlock.Number().Int64(), newblk.Number().Int64())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("totalIncome: ", total.String())
	if total.Cmp(big.NewInt(perMoney*4/10*4)) != 0 {
		log.Fatal("query keepers[0]'s manageIncome failed, it is not equal to 192")
	}

	log.Println("upkeeping's tests pass")

	return nil
}

func getSigs(keeperAddress []common.Address, keeperSk []string, providerAddr, upKeepingAddr common.Address, stStart, stLength, stValue *big.Int, merkleRoot [32]byte, share []int64) ([][]byte, error) {
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

func getAddProviderSigs(keeperAddress []common.Address, keeperSk []string, upKeepingAddr common.Address, providerAddr []common.Address) ([][]byte, error) {
	sigs := [][]byte{}
	for i := 0; i < len(keeperAddress); i++ {
		sig, err := role.SignForAddProvider(upKeepingAddr, providerAddr, keeperSk[i])
		if err != nil {
			log.Println("signForAddProvider error:", err)
			return sigs, err
		}
		sigs = append(sigs, sig)
	}
	return sigs, nil
}

func getSetStopSigns(keeperAddress []common.Address, keeperSk []string, providerAddr, upKeepingAddr common.Address) ([][]byte, error) {
	sigs := [][]byte{}
	for i := 0; i < len(keeperAddress); i++ {
		sig, err := role.SignForSetStop(upKeepingAddr, providerAddr, keeperSk[i])
		if err != nil {
			log.Println("signForSetStop error:", err)
			return sigs, err
		}
		sigs = append(sigs, sig)
	}
	return sigs, nil
}
