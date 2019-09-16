package keeper

import (
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
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

//输入一个节点的地址 查它的余额
func queryBalanceTest(addr common.Address) *big.Int {
	queryBalanceRes, _ := contracts.QueryBalance(addr.Hex())
	return queryBalanceRes
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
