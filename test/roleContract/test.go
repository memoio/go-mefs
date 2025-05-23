package main

import (
	"errors"
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils"
)

const (
	kpAddr   = "0x208649111Fd9253B76950e9f827a5A6dd616340d"
	kpSk     = "8f9eb151ffaebf2fe963e6185f0d1f8c1e8397e5905b616958d765e7753329ea"
	adminSk  = "aca26228a9ed5ca4da2dd08d225b1b1e049d80e1b126c0d7e644d04d0fb910a3"
	moneyTo  = 1000000000000000000
	waitTime = 3 * time.Second
)

var ethEndPoint, qethEndPoint string

func main() {
	utils.StartLogger()
	flag.String("testnet", "--eth=http://119.147.213.219:8101 --qeth=http://119.147.213.219:8101", "testnet commands")
	eth := flag.String("eth", "http://119.147.213.220:8191", "eth api address for set;")
	qeth := flag.String("qeth", "http://119.147.213.220:8194", "eth api address for query;")
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	balance := test.QueryBalance(kpAddr, qethEndPoint)
	if balance.Cmp(big.NewInt(moneyTo)) <= 0 {
		test.TransferTo(big.NewInt(moneyTo), kpAddr, ethEndPoint, qethEndPoint)
	}

	//if err := testDeploy(); err != nil {
	//	log.Fatal(err)
	//}

	if err := testKeeper(); err != nil {
		log.Fatal(err)
	}

	if err := testProvider(); err != nil {
		log.Fatal(err)
	}
}

func testDeploy() error {
	contracts.EndPoint = ethEndPoint
	var tmpAddr common.Address
	cRole := contracts.NewCR(tmpAddr, adminSk)
	err := cRole.DeployKPMap()
	if err != nil {
		log.Println("DeployKeeperProviderMap err:", err)
		return err
	}

	time.Sleep(2 * time.Minute)

	return nil
}

func testKeeper() (err error) {
	log.Println("==========test keeper=========")

	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Println(err)
		return err
	}

	err = test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("1. test set keeper")

	keeperAddr := common.HexToAddress(userAddr[2:])

	contracts.EndPoint = ethEndPoint

	cRole := contracts.NewCR(common.HexToAddress(userAddr), adminSk)
	err = cRole.SetKeeper(keeperAddr, true)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}
	time.Sleep(waitTime)

	contracts.EndPoint = qethEndPoint

	retryCount := 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsKeeper(common.HexToAddress(userAddr))
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set keeper fails")
			}
			continue
		}
		log.Println("set keeper success")
		break
	}

	price, err := cRole.GetKeeperPrice()
	if err != nil {
		return err
	}
	log.Println("set keeper need price is: ", price)
	log.Println("set keeper price is: ", utils.KeeperDeposit)
	contracts.EndPoint = ethEndPoint

	log.Println("2. test pledge keeper")
	amount := new(big.Int).SetInt64(utils.KeeperDeposit)
	cRole = contracts.NewCR(common.HexToAddress(userAddr), userSk)
	err = cRole.PledgeKeeper(amount)
	if err != nil {
		return err
	}
	time.Sleep(waitTime)

	_, keeperContract, err := contracts.GetKeeperContractFromIndexer(keeperAddr)
	if err != nil {
		log.Println("keeperContracterr:", err)
		return err
	}
	active, banned, money, time, err := keeperContract.Info(&bind.CallOpts{
		From: keeperAddr,
	}, keeperAddr)
	if err != nil {
		return err
	}

	log.Println("get keepe info: ", active, banned, money, time)

	if money.Cmp(amount) != 0 {
		return errors.New("wrong parameters")
	}

	log.Println("3. test set to non-keeper")
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(userAddr), adminSk)
	err = cRole.SetKeeper(keeperAddr, false)
	if err != nil {
		log.Println("setKeeper false err:", err)
		return err
	}

	return nil
}

func testProvider() (err error) {
	log.Println("==========test provider=========")

	userAddr, userSk, err := test.CreateAddr()
	if err != nil {
		log.Println(err)
		return err
	}

	err = test.TransferTo(big.NewInt(moneyTo), userAddr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("1. test set provider")

	proAddr := common.HexToAddress(userAddr[2:])

	contracts.EndPoint = ethEndPoint
	cRole := contracts.NewCR(common.HexToAddress(userAddr), adminSk)
	err = cRole.SetProvider(proAddr, true)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}
	time.Sleep(waitTime)

	contracts.EndPoint = qethEndPoint
	retryCount := 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsProvider(common.HexToAddress(userAddr))
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set provider fails")
			}
			continue
		}
		break
	}

	log.Println("set provider success")

	price, err := cRole.GetProviderPrice()
	if err != nil {
		return err
	}

	log.Println("pledge provide need price is: ", price)
	contracts.EndPoint = ethEndPoint
	log.Println("2. test pledge provider")
	size := new(big.Int).SetInt64(utils.DepositCapacity / 1024)

	cRole = contracts.NewCR(common.HexToAddress(userAddr), userSk)
	err = cRole.PledgeProvider(size)
	if err != nil {
		return err
	}
	time.Sleep(waitTime)

	_, providerContract, err := contracts.GetProviderContractFromIndexer(common.HexToAddress(userAddr))
	if err != nil {
		log.Println("providerContracterr:", err)
		return err
	}
	active, banned, money, time, err := providerContract.Info(&bind.CallOpts{
		From: common.HexToAddress(userAddr),
	}, proAddr)
	if err != nil {
		return err
	}

	log.Println("get provider info: ", active, banned, money, time)

	if money.Cmp(big.NewInt(307199999999998976)) != 0 {
		return errors.New("wrong parameters")
	}

	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(userAddr), adminSk)
	err = cRole.SetProvider(proAddr, false)
	if err != nil {
		log.Println("setProvider false err:", err)
		return err
	}

	return nil
}

func testRole() (err error) {
	log.Println("==========role=========")

	log.Println("test set keeper")

	keeperAddr := common.HexToAddress(kpAddr[2:])

	contracts.EndPoint = ethEndPoint
	cRole := contracts.NewCR(common.HexToAddress(kpAddr), adminSk)
	err = cRole.SetKeeper(keeperAddr, true)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount := 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsKeeper(common.HexToAddress(kpAddr))
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set keeper fails")
			}
			continue
		}
		log.Println("set keeper success")
		break
	}

	err = cRole.SetKeeper(keeperAddr, false)
	if err != nil {
		log.Println("setKeeper false err:", err)
		return err
	}

	log.Println("test set provider")
	contracts.EndPoint = ethEndPoint
	proAddr, _, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	test.TransferTo(big.NewInt(moneyTo), proAddr, ethEndPoint, qethEndPoint)

	providerAddr := common.HexToAddress(proAddr[2:])

	err = cRole.SetProvider(providerAddr, true)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	log.Println("test get result of set provider")
	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsProvider(common.HexToAddress(proAddr))
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set provider fails")
			}
			continue
		}
		log.Println("set provider success")
		break
	}

	err = cRole.SetProvider(providerAddr, true)
	if err != nil {
		log.Println("setProvider false err:", err)
		return err
	}

	log.Println("test set add kp")
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(kpAddr), kpSk)
	err = cRole.AddKeeperProvidersToKPMap(keeperAddr, []common.Address{providerAddr})
	if err != nil {
		log.Println("Add Keeper Providers To KPMap err:", err)
		return err
	}

	log.Println("test get keeper from kpmap")
	contracts.EndPoint = qethEndPoint
	retryCount = 0
	flag := false
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		kps, err := cRole.GetAllKeeperInKPMap()
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
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		pids, err := cRole.GetProviderInKPMap(keeperAddr)
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

	log.Println("test set provider second")
	contracts.EndPoint = ethEndPoint
	proAddr2, _, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	test.TransferTo(big.NewInt(moneyTo), proAddr2, ethEndPoint, qethEndPoint)

	providerAddr2 := common.HexToAddress(proAddr2[2:])

	cRole = contracts.NewCR(common.HexToAddress(kpAddr), adminSk)
	err = cRole.SetProvider(providerAddr2, true)
	if err != nil {
		log.Println("setProvider2 err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsProvider(common.HexToAddress(proAddr2))
		if err != nil || res == false {
			if retryCount > 20 {
				log.Fatal("set provider2 fails")
			}
			continue
		}
		log.Println("set provider2 success")
		break
	}

	log.Println("test set add kp second")
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(kpAddr), kpSk)
	err = cRole.AddKeeperProvidersToKPMap(keeperAddr, []common.Address{providerAddr2})
	if err != nil {
		log.Println("Add Keeper Providers To KPMap err:", err)
		return err
	}

	log.Println("test get provider from kpmap second")
	contracts.EndPoint = qethEndPoint
	retryCount = 0
	flag = false
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		cRole = contracts.NewCR(common.HexToAddress(proAddr2), kpSk)
		pids, err := cRole.GetProviderInKPMap(keeperAddr)
		if err != nil {
			if retryCount > 20 {
				log.Fatal("Get Provider2 In KPMap fails")
			}
			continue
		}

		flag = false
		for _, pidr := range pids {
			if pidr.String() == providerAddr2.String() {
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
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(kpAddr), adminSk)
	err = cRole.DeleteProviderFromKPMap(keeperAddr, providerAddr)
	if err != nil {
		log.Fatal("delete provider from kpmap err:", err)
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	flag = false
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		cRole = contracts.NewCR(common.HexToAddress(proAddr), kpSk)
		pids, err := cRole.GetProviderInKPMap(keeperAddr)
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

	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(kpAddr), kpSk)
	err = cRole.DeleteKeeperFromKPMap(keeperAddr)
	if err != nil {
		log.Fatal("delete keeper from kpmap err:", err)
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	flag = false
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		kps, err := cRole.GetAllKeeperInKPMap()
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
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(kpAddr), adminSk)
	err = cRole.SetKeeper(keeperAddr, false)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsKeeper(common.HexToAddress(kpAddr))
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
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(proAddr), adminSk)
	err = cRole.SetProvider(providerAddr, false)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsProvider(common.HexToAddress(proAddr))
		if err != nil || res == true {
			if retryCount > 20 {
				log.Fatal("set provider false fails")
			}
			continue
		}
		log.Println("set provider false success")
		break
	}

	log.Println("test set provider false second")
	contracts.EndPoint = ethEndPoint
	cRole = contracts.NewCR(common.HexToAddress(proAddr2), adminSk)
	err = cRole.SetProvider(providerAddr2, false)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(test.RetryGetInfoSleepTime * time.Duration(retryCount))
		res, err := cRole.IsProvider(common.HexToAddress(proAddr2))
		if err != nil || res == true {
			if retryCount > 20 {
				log.Fatal("set provider2 false fails")
			}
			continue
		}
		log.Println("set provider2 false success")
		break
	}

	return nil
}
