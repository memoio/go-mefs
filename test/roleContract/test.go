package main

import (
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/test"
)

const (
	kpAddr  = "0x208649111Fd9253B76950e9f827a5A6dd616340d"
	kpSk    = "8f9eb151ffaebf2fe963e6185f0d1f8c1e8397e5905b616958d765e7753329ea"
	adminSk = "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4"
	moneyTo = 1000000000000000
)

var ethEndPoint, qethEndPoint string

func main() {
	flag.String("testnet", "--eth=http://39.100.146.21:8101 --qeth=http://47.92.5.51:8101", "testnet commands")
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address for set;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address for query;")
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

	if err := testRole(); err != nil {
		log.Fatal(err)
	}
}

func testDeploy() error {
	contracts.EndPoint = ethEndPoint
	err := contracts.DeployKPMap(adminSk)
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

	contracts.EndPoint = ethEndPoint
	err = contracts.SetKeeper(keeperAddr, adminSk, true)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
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
	contracts.EndPoint = ethEndPoint
	proAddr, _, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	test.TransferTo(big.NewInt(moneyTo), proAddr, ethEndPoint, qethEndPoint)

	providerAddr := common.HexToAddress(proAddr[2:])

	err = contracts.SetProvider(providerAddr, adminSk, true)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	log.Println("test get result of set provider")
	contracts.EndPoint = qethEndPoint
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
	return nil
	log.Println("test set add kp")
	contracts.EndPoint = ethEndPoint
	err = contracts.AddKeeperProvidersToKPMap(keeperAddr, kpSk, keeperAddr, []common.Address{providerAddr})
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

	log.Println("test set provider second")
	contracts.EndPoint = ethEndPoint
	proAddr2, _, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	test.TransferTo(big.NewInt(moneyTo), proAddr2, ethEndPoint, qethEndPoint)

	providerAddr2 := common.HexToAddress(proAddr2[2:])

	err = contracts.SetProvider(providerAddr2, adminSk, true)
	if err != nil {
		log.Println("setProvider2 err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		res, err := contracts.IsProvider(providerAddr2)
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
	err = contracts.AddKeeperProvidersToKPMap(keeperAddr, kpSk, keeperAddr, []common.Address{providerAddr2})
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
		time.Sleep(30 * time.Second)
		pids, err := contracts.GetProviderInKPMap(providerAddr2, keeperAddr)
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
	err = contracts.DeleteProviderFromKPMap(keeperAddr, adminSk, keeperAddr, providerAddr)
	if err != nil {
		log.Fatal("delete provider from kpmap err:", err)
	}

	contracts.EndPoint = qethEndPoint
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

	contracts.EndPoint = ethEndPoint
	err = contracts.DeleteKeeperFromKPMap(keeperAddr, kpSk, keeperAddr)
	if err != nil {
		log.Fatal("delete keeper from kpmap err:", err)
	}

	contracts.EndPoint = qethEndPoint
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
	contracts.EndPoint = ethEndPoint
	err = contracts.SetKeeper(keeperAddr, adminSk, false)
	if err != nil {
		log.Println("setKeeper err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
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
	contracts.EndPoint = ethEndPoint
	err = contracts.SetProvider(providerAddr, adminSk, false)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
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

	log.Println("test set provider false second")
	contracts.EndPoint = ethEndPoint
	err = contracts.SetProvider(providerAddr2, adminSk, false)
	if err != nil {
		log.Println("setProvider err:", err)
		return err
	}

	contracts.EndPoint = qethEndPoint
	retryCount = 0
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		res, err := contracts.IsProvider(providerAddr2)
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
