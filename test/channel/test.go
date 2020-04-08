package main

import (
	"flag"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/test"
	"github.com/memoio/go-mefs/utils/address"
)

const (
	moneyTo = 1000000000000000
)

var ethEndPoint, qethEndPoint string

func main() {
	flag.String("testnet", "--eth=http://47.92.5.51:8101 --qeth=http://39.100.146.21:8101", "testnet commands")
	eth := flag.String("eth", "http://212.64.28.207:8101", "eth api address for set;")
	qeth := flag.String("qeth", "http://39.100.146.165:8101", "eth api address for query;")
	flag.Parse()
	ethEndPoint = *eth
	qethEndPoint = *qeth

	if err := testCloseChannel(); err != nil {
		log.Fatal(err)
	}

	if err := testChannelTimeout(); err != nil {
		log.Fatal(err)
	}
}

func testChannelTimeout() (err error) {
	log.Println("==========test channel timeout=========")
	contracts.EndPoint = ethEndPoint

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

	proAddr, _, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	err = test.TransferTo(big.NewInt(moneyTo), proAddr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Println(err)
		return err
	}

	providerAddr := common.HexToAddress(proAddr[2:])
	localAddr := common.HexToAddress(userAddr[2:])

	timeout := big.NewInt(5 * 60)
	moneyToChannel := big.NewInt(1000000)

	log.Println("test deploy channel contract")
	channelAddr, err := contracts.DeployChannelContract(userSk, localAddr, localAddr, providerAddr, timeout, moneyToChannel, true)
	if err != nil {
		log.Fatal("deployChannelErr:", err)
		return err
	}

	log.Println("test query channel balance: ", channelAddr.String())
	contracts.EndPoint = qethEndPoint
	retryCount := 0
	cbalance := test.QueryBalance(channelAddr.String(), qethEndPoint)
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		cbalance = test.QueryBalance(channelAddr.String(), qethEndPoint) //查看部署的channel合约的账户余额
		if cbalance.Cmp(big.NewInt(0)) == 0 {
			if retryCount > 20 {
				log.Fatal("channel contract has no balance")
			}
			continue
		}
		if cbalance.Cmp(moneyToChannel) != 0 {
			log.Fatal("channel contract has wrong balance")
		}
		break
	}

	log.Println("test query channel addr")
	addGot, err := contracts.DeployChannelContract(userSk, localAddr, localAddr, providerAddr, timeout, moneyToChannel, false)
	if err != nil {
		log.Println("Get Channel Err: ", err)
		return err
	}

	if addGot.String() != channelAddr.String() {
		log.Println("Get Wrong Channel")
	}

	log.Println("test channel timeout")
	//触发channelTimeout()
	contracts.EndPoint = ethEndPoint
	err = contracts.ChannelTimeout(channelAddr, userSk)
	if err != nil {
		log.Println("call channelTimeout fail")
		return err
	}

	retryCount = 0
	nbalance := test.QueryBalance(channelAddr.String(), qethEndPoint)
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		nbalance = test.QueryBalance(channelAddr.String(), qethEndPoint) //查看部署的channel合约的账户余额
		if nbalance.Cmp(cbalance) == 0 {
			if retryCount > 20 {
				log.Fatal("call channel timeout failed")
			}
			continue
		}

		if nbalance.Cmp(big.NewInt(0)) != 0 {
			log.Fatal("call channel timeout has wrong balance")
		}

		log.Println("test channel timeout success")

		break
	}

	return nil
}

func testCloseChannel() (err error) {
	log.Println("test close channel")
	contracts.EndPoint = ethEndPoint

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

	proAddr, proSk, err := test.CreateAddr()
	if err != nil {
		log.Fatal("create provider fails")
	}

	err = test.TransferTo(big.NewInt(moneyTo), proAddr, ethEndPoint, qethEndPoint)
	if err != nil {
		log.Println(err)
		return err
	}

	providerAddr := common.HexToAddress(proAddr[2:])
	localAddr := common.HexToAddress(userAddr[2:])

	log.Println("test deploy channel contract")
	timeout := big.NewInt(5 * 60)
	moneyToChannel := big.NewInt(1000000)
	channelAddr, err := contracts.DeployChannelContract(userSk, localAddr, localAddr, providerAddr, timeout, moneyToChannel, true)
	if err != nil {
		log.Fatal("deployChannelErr:", err)
		return err
	}

	log.Println("test query channel balance: ", channelAddr.String())
	retryCount := 0
	cbalance := test.QueryBalance(channelAddr.String(), qethEndPoint)
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		cbalance = test.QueryBalance(channelAddr.String(), qethEndPoint) //查看部署的channel合约的账户余额
		if cbalance.Cmp(big.NewInt(0)) == 0 {
			if retryCount > 20 {
				log.Fatal("channel contract has no balance")
			}
			continue
		}
		if cbalance.Cmp(moneyToChannel) != 0 {
			log.Fatal("channel contract has wrong balance")
		}

		log.Println(channelAddr.String(), "has channel balance: ", cbalance.String())
		break
	}

	log.Println("test query channel contract")
	contracts.EndPoint = qethEndPoint
	addGot, _, err := contracts.GetLatestChannel(localAddr, localAddr, providerAddr, localAddr)
	if err != nil {
		log.Fatal("GetChannelAddr fails:", err)
		return err
	}

	if addGot.String() != channelAddr.String() {
		log.Fatal("Get Wrong ChannelAddr")
	}

	log.Println("test close channel contract")
	balance := test.QueryBalance(userAddr, qethEndPoint) //查看账户余额
	//签名
	value := big.NewInt(11111)
	contracts.EndPoint = ethEndPoint
	chanID, _ := address.GetIDFromAddress(channelAddr.String())
	mes, err := role.SignForChannel(chanID, userSk, value)
	if err != nil {
		log.Fatal("SignForChannelErr:", err)
		return err
	}

	cSign := new(mpb.ChannelSign)
	err = proto.Unmarshal(mes, cSign)
	if err != nil {
		log.Fatal("Unmarshal SignForChannelErr:", err)
	}

	//provider触发CloseChannel()
	err = contracts.CloseChannel(channelAddr, proSk, cSign.GetSig(), value)
	if err != nil {
		log.Fatal("CloseChannelErr:", err)
		return err
	}

	retryCount = 0
	nbalance := test.QueryBalance(channelAddr.String(), qethEndPoint)
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		nbalance = test.QueryBalance(channelAddr.String(), qethEndPoint) //查看部署的channel合约的账户余额
		if nbalance.Cmp(big.NewInt(0)) > 0 {
			if retryCount > 20 {
				log.Fatal("call close channel error, balance has: ", nbalance.String())
				return
			}
			continue
		}
		log.Println("call close channel success: channel has no balance")
		break
	}

	retryCount = 0
	ubalance := test.QueryBalance(userAddr, qethEndPoint)
	for {
		retryCount++
		time.Sleep(30 * time.Second)
		ubalance = test.QueryBalance(userAddr, qethEndPoint)
		if ubalance.Cmp(balance) == 0 {
			if retryCount > 20 {
				log.Fatal("channel contract has balance")
			}
			continue
		}

		ubalance.Add(ubalance, value)
		ubalance.Sub(ubalance, moneyToChannel)

		if ubalance.Cmp(balance) != 0 {
			log.Fatal("call close channel failed")
		}

		log.Println("call close channel success: user has refund his remain value")
		break
	}

	log.Println("test get download income from chain")
	chAddrs := []common.Address{channelAddr}
	total, daily, err := contracts.GetDownloadIncome(chAddrs, providerAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("totalIncome: ", total.String(), "\ndailyIncome: ", daily.String())
	if total.Cmp(value) != 0 {
		log.Fatal("test get income failed")
	}
	return nil
}
