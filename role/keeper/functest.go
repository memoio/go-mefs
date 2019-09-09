package keeper

import (
	"errors"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/contracts/upKeeping"
	"github.com/memoio/go-mefs/core"
	fr "github.com/memoio/go-mefs/repo/fsrepo"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
)

var ErrTestFail = errors.New("test Failed")

var serverKids = []string{"8MHS9fZzRaHNj4mP1kYDebwySmLzaw", "8MGRZbvn8caS431icB2P1uT74B3EHh", "8MJCzFbpXCvdfzmJy5L8jiw4w1qPdY", "8MKX58Ko5vBeJUkfgpkig53jZzwqoW", "8MHYzNkm6dF9SWU5u7Py8MJ31vJrzS", "8MK2saApPQMoNfVmnRDiApoAWFzo2K"}
var serverPids = []string{"8MHXst83NnSfYHnyqWMVjwjt2GiutV", "8MGrkL5cUpPsPbePvCfwCx6HemwDvy", "8MJ71X96BcnUNkhSFjc6CCsemL6nSQ", "8MGZ5nYsYw3Kmt8zC44W4V1NYaTGcE", "8MGhVo1ib6C6PmFhfQK4Hr3hHwQjC9", "8MJcdk2cyQvZknpxYf2AmGKDHRSRJP", "8MG9ZMYoZrZxjc7bVMeqJkaxAdb3Wx", "8MGqojupxiCesALno7sA73NhJkcSY5", "8MKAiRexSQG4SpGrpEQb4s9wjxJimX", "8MKU1DT94SB3aHTrMqWcJa2oLRtTzv", "8MJaFY7yAyYAvnjnM5hTbTfpjXhTHx", "8MGUGzCk1RUvq1aTPd9uuorrZ7FRhx", "8MHSARkgxWkjx5hKPm9vhX2v1VZ6GT"}

var testGid, testPid string
var testStartTime int64 //用全局变量，保证每次测试的对象不变

//测试时空值正确性的函数，每隔一段时间，上传数据，计算之前的时空值，返回实际计算的值，与理论值对比
func ResultSummaryTest() string {
	log.Println("===================ResultSummaryTest============================")
	defer log.Println("===================ResultSummaryTest============================")
	if testGid == "" {
		PInfo.Range(func(groupid, groupsinfo interface{}) bool {
			thisgroupid := groupid.(string)
			testGid = thisgroupid
			return false
		})
	}
	if testPid == "" {
		thisGroupsInfo, ok := getGroupsInfo(testGid)
		if !ok {
			log.Println("getGroupsInfo(testGid) error")
			return "0"
		}
		testPid = thisGroupsInfo.Providers[0]
	}
	if testStartTime == int64(0) {
		testStartTime = utils.GetUnixNow()
	} //本次测试为第一次测试，则选定测试用provider、user

	testEndTime := utils.GetUnixNow()
	actual, _ := resultSummary(testGid, testPid, testStartTime, testEndTime)
	log.Println("actual:", actual)
	return actual.String()
}

func SmartContractTest(kCount int, pCount int, amount *big.Int) error {
	log.Println(">>>>>>>>>>>>>>>>>>>>>SmartContractTest>>>>>>>>>>>>>>>>>>>>>")
	defer log.Println("===================SmartContractTestEnd============================")

	localAddr, _ := ad.GetAddressFromID(localNode.Identity.Pretty()) //将id转化成智能合约中的address格式
	mapKeeperAddr := make(map[common.Address]*big.Int)
	mapProviderAddr := make(map[common.Address]*big.Int)
	listKeeperAddr := []common.Address{localAddr}
	listProviderAddr := []common.Address{}
	mapKeeperAddr[localAddr] = queryBalanceTest(localAddr) //本节点作为一个keeper进行合约部署

	i := 0
	for _, serverKid := range serverKids { //得到keeper地址 并且查询初始余额
		tempAddr, _ := ad.GetAddressFromID(serverKid)
		mapKeeperAddr[tempAddr] = queryBalanceTest(tempAddr)
		listKeeperAddr = append(listKeeperAddr, tempAddr)
		if i++; i == kCount {
			break
		}
	}
	i = 0
	for _, serverPid := range serverPids { //得到provider地址 并查询初始余额
		tempAddr, _ := ad.GetAddressFromID(serverPid)
		mapProviderAddr[tempAddr] = queryBalanceTest(tempAddr)
		listProviderAddr = append(listProviderAddr, tempAddr)
		if i++; i == pCount {
			break
		}
	}

	hexPk, err := fr.GetHexPrivKeyFromKS(localNode.Identity, localNode.Password) //获得本节点私钥
	if err != nil {
		log.Println("获取私钥错误:", err)
		return err
	}

	ukaddr, uk, err := deployUKTest(hexPk, localAddr, listKeeperAddr, listProviderAddr) //部署一个合约，并且获得其实例
	if err != nil {
		log.Println("deployUKTest()err:", err)
		return err
	}
	log.Println("合约部署成功,合约地址:", ukaddr.Hex())
	log.Println("等待2分钟 查询合约金额和部署合约开销")
	time.Sleep(120 * time.Second)
	amountUk := queryBalanceTest(ukaddr)
	log.Println("合约金额：", amountUk)
	amountLocal := queryBalanceTest(localAddr)
	amountCost := big.NewInt(0)
	amountCost.Sub(amountLocal, mapKeeperAddr[localAddr])
	log.Println("本地节点金额变化：", amountCost)
	mapKeeperAddr[localAddr] = amountLocal

	err = contracts.SpaceTimePay(uk, localAddr, listProviderAddr[0], hexPk, amount)
	if err != nil {
		log.Println("时空支付错误:", err)
		return err
	}
	log.Println("时空支付成功,等待2分钟，进行金额变化计算")
	time.Sleep(120 * time.Second)
	log.Println("---------keeper金额变化----------")
	for kAddr, amount := range mapKeeperAddr {
		amountNow := queryBalanceTest(kAddr)
		amountCost := big.NewInt(0)
		amountCost.Sub(amountNow, amount)
		if kAddr == localAddr {
			log.Println(kAddr.String(), "(本节点):", amountCost)
		} else {
			log.Println(kAddr.String(), ":", amountCost)
		}
	}
	log.Println("---------provider金额变化----------")
	for pAddr, amount := range mapProviderAddr {
		amountNow := queryBalanceTest(pAddr)
		amountCost := big.NewInt(0)
		amountCost.Sub(amountNow, amount)
		log.Println(pAddr.String(), ":", amountCost)
		if listProviderAddr[0] == pAddr && amountCost.Sign() <= 0 {
			return errors.New("被支付的provider金额变化为0")
		}
	}
	return nil
}

//本节点部署一个测试用的uk合约
//传入参数为endPoint,私钥，合约中的user地址、keeper地址列表、provider地址列表，返回测试用用的合约实例，没有放入mapper，测试结束后，该实例丢失
func deployUKTest(hexKey string, userAddr common.Address, listKeeperAddr, listProviderAddr []common.Address) (common.Address, *upKeeping.UpKeeping, error) {
	key, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		log.Println("HexToECDSAErr:", err)
		return common.Address{}, nil, err
	}
	auth := bind.NewKeyedTransactor(key)
	auth.Value = big.NewInt(2457600)
	client := contracts.GetClient(contracts.EndPoint)
	ukaddr, _, uk, err := upKeeping.DeployUpKeeping(auth, client, userAddr, listKeeperAddr, listProviderAddr, big.NewInt(10), big.NewInt(1024), big.NewInt(1111))
	if err != nil {
		return common.Address{}, nil, err
	}
	return ukaddr, uk, nil
}

//输入一个节点的地址 查它的余额
func queryBalanceTest(addr common.Address) *big.Int {
	queryBalanceRes, _ := contracts.QueryBalance(addr.Hex())
	return queryBalanceRes
}

func DeployKeeperContractTest() {
	log.Println("===================DeployKeeperContractTestBegin========================")
	defer log.Println("===================DeployKeeperContractTestEnd=======================")
	config, err := localNode.Repo.Config()
	if err != nil {
		log.Println("获取config错误:", err)
		return
	}
	endPoint := config.Eth
	log.Println("endPoint:", endPoint)

	hexPk, err := fr.GetHexPrivKeyFromKS(localNode.Identity, localNode.Password) //获得本节点私钥
	if err != nil {
		log.Println("获取私钥错误:", err)
		return
	}
	log.Println("私钥hexPk:", hexPk)

	localAddr, _ := ad.GetAddressFromID(localNode.Identity.Pretty()) //将id转化成智能合约中的address格式
	log.Println("本节点地址localAddr：", localAddr.String())

	queryBalanceRes, err := contracts.QueryBalance(localAddr.Hex())
	if err != nil {
		log.Println("查询余额错误：", err)
		return
	}
	log.Println("查询余额信息：", queryBalanceRes)

	err = contracts.KeeperContract(hexPk)
	if err != nil {
		log.Println("keeper合约部署错误:", err)
		return
	}
	log.Println("keeper部署合约成功")
}

//SetKeeperTest set an account as keeper
func SetKeeperTest(address string, node *core.MefsNode, isKeeper bool) error {
	log.Println("===================SetKeeperTestBegin========================")
	defer log.Println("===================SetKeeperTestEnd=======================")
	config, err := node.Repo.Config()
	if err != nil {
		log.Println("获得config错误：", err)
		return err
	}
	endPoint := config.Eth
	log.Println("endPoint:", endPoint)

	hexPk := "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4" //此私钥部署过keeper合约
	log.Println("账户地址address：", address)

	queryBalanceRes, err := contracts.QueryBalance(address)
	if err != nil {
		log.Println("查询余额错误：", err)
		return err
	}
	log.Println("查询余额信息：", queryBalanceRes)

	err = contracts.SetKeeper(common.HexToAddress(address), hexPk, isKeeper)
	if err != nil {
		log.Println("setKeeper错误:", err)
		return err
	}
	log.Println("setKeeper成功")
	// log.Println("===开始测试IsKeeper===")
	// time.Sleep(2 * time.Minute) //使上面的setKeeper交易执行
	// isKeeper, err := contracts.IsKeeper(endPoint, common.HexToAddress(address))
	// if err != nil {
	// 	log.Println("IsKeeper错误：", err)
	// 	return err
	// }
	// if !isKeeper {
	// 	log.Println("===IsKeeper测试失败===")
	// 	return err
	// }
	// log.Println("===IsKeeper测试成功===")
	return nil
}

func DeployProviderContractTest() {
	log.Println("===================DeployProviderContractTestBegin========================")
	defer log.Println("===================DeployProviderContractTestEnd=======================")
	config, _ := localNode.Repo.Config()
	endPoint := config.Eth
	log.Println("endPoint:", endPoint)

	hexPk, err := fr.GetHexPrivKeyFromKS(localNode.Identity, localNode.Password) //获得本节点私钥
	if err != nil {
		log.Println("获取私钥错误:", err)
		return
	}
	log.Println("私钥hexPk:", hexPk)

	localAddr, _ := ad.GetAddressFromID(localNode.Identity.Pretty()) //将id转化成智能合约中的address格式
	log.Println("本节点地址localAddr：", localAddr.String())

	queryBalanceRes, err := contracts.QueryBalance(localAddr.Hex())
	if err != nil {
		log.Println("查询余额错误：", err)
		return
	}
	log.Println("查询余额信息：", queryBalanceRes)

	err = contracts.ProviderContract(hexPk)
	if err != nil {
		log.Println("provider合约部署错误:", err)
		return
	}
	log.Println("provider部署合约成功")
}

//SetProviderTest set an account as provider
func SetProviderTest(address string, node *core.MefsNode, isProvider bool) error {
	log.Println("===================SetProviderTestBegin========================")
	defer log.Println("===================SetProviderTestEnd=======================")
	config, err := node.Repo.Config()
	if err != nil {
		log.Println("获得config错误：", err)
		return err
	}
	endPoint := config.Eth
	log.Println("endPoint:", endPoint)

	hexPk := "928969b4eb7fbca964a41024412702af827cbc950dbe9268eae9f5df668c85b4"

	log.Println("账户地址address：", address)

	queryBalanceRes, err := contracts.QueryBalance(address)
	if err != nil {
		log.Println("查询余额错误：", err)
		return err
	}
	log.Println("查询余额信息：", queryBalanceRes)

	err = contracts.SetProvider(common.HexToAddress(address), hexPk, isProvider)
	if err != nil {
		log.Println("setProvider错误:", err)
		return err
	}
	log.Println("setProvider成功")
	// log.Println("===开始测试IsProvider===")
	// isProvider, err := contracts.IsProvider(endPoint, common.HexToAddress(address))
	// if err != nil {
	// 	log.Println("IsProvider错误：", err)
	// 	return err
	// }
	// if !isProvider {
	// 	log.Println("===IsProvider测试失败===")
	// 	return err
	// }
	// log.Println("===IsProvider测试成功===")
	return nil
}

//testSaveChalPay 测试支付信息的保存和读取操作
func SaveChalPayTest() error {
	log.Println(">>>>>>>>>>>>>>>>>>>>>SaveChalPayTest>>>>>>>>>>>>>>>>>>>>>")
	defer log.Println("===================SaveChalPayTestEnd============================")
	groupid := "testGroupid"
	pid := "testPid"
	signature := "signature"
	proof := "proof"
	beginTime := int64(512)
	endTime := int64(1024)
	spaceTime := big.NewInt(2048)
	log.Println("没有信息，直接获取数据")
	thisTime := checkLastPayTime(groupid, pid)
	if thisTime != int64(0) {
		log.Println("获取空数据出错")
		return ErrTestFail
	}
	log.Println("ok!\n保存测试数据")
	_, _, err := saveLastPay(groupid, pid, signature, proof, beginTime, endTime, spaceTime)
	if err != nil {
		log.Println("saveLastPay()error:", err)
		return ErrTestFail
	}
	log.Println("ok!\n再次获取数据")
	thisTime = checkLastPayTime(groupid, pid)
	if thisTime != endTime {
		log.Println("获取空数据出错")
		return ErrTestFail
	}
	log.Print("测试成功")
	return nil
}

//这个函数用于写代码时临时测试各种功能.
func FreeTest() {
	err := localNode.Routing.(*dht.IpfsDHT).CmdPutTo("testkey", "testvalue", "local")
	if err != nil {
		log.Println("error!!:", err)
	}
	value, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom("testkey", "local")
	if err != nil {
		log.Println("error!!:", err)
	} else {
		log.Println(string(value))
	}
}
