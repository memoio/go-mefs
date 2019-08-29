package keeper

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/memoio/go-mefs/contracts"
	fr "github.com/memoio/go-mefs/repo/fsrepo"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//PayInfo 最近一次支付信息的记录
//key为 PU结构体，value为*chalpay
var PayInfo sync.Map

//chalpay 一次支付结果在内存中的记录
//作为PayInfo结构体的
type chalpay struct {
	pid        string   //挑战对象
	uid        string   //挑战的数据所属对象
	begin_time int64    // 上次支付结算的结束时间，当前的起始时间
	end_time   int64    // 这次支付结算的结束时间
	spacetime  *big.Int // 此次结算的时空值，根据过去一段时间的结果计算
	signature  string   // 对spacetime值的签名
	proof      string   // 挑战结果，预留
}

func spaceTimePayRegular(ctx context.Context) {
	fmt.Println("spaceTimePayRegular() start!")
	ticker := time.NewTicker(SPACETIMEPAYTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			spaceTimePay()
		}
	}
}

//spaceTimePay 每隔一段时间触发的时空支付过程。对管理的所有user-provider进行一次支付
func spaceTimePay() {
	type payMaterial struct {
		groupid string
		pid     string
		price   int64
	}
	var materials []*payMaterial
	PInfo.Range(func(groupid, groupsInfo interface{}) bool { //循环user
		thisGroupid, ok := groupid.(string)
		if !ok {
			fmt.Println("spaceTimePay()", ErrPInfoTypeAssert)
			return true
		}
		thisGroupsInfo, ok := groupsInfo.(*GroupsInfo)
		if !ok {
			fmt.Println("spaceTimePay()", ErrPInfoTypeAssert)
			return true
		}
		for _, pidString := range thisGroupsInfo.Providers { //循环当前user的provider
			materials = append(materials, &payMaterial{
				groupid: thisGroupid,
				pid:     pidString,
				price:   GetUpkeeping(thisGroupsInfo).Price,
			})
		}
		return true
	})
	//避免链操作时间过长，先收集数据，再进行支付操作
	for _, material := range materials {
		doSpaceTimePay(material.groupid, material.pid, material.price)
	}

}

//dospacetimePay 时空支付函数，每过一段时间，对管理的provider进行一次支付
//注意：同一个组内的keeper，选出一个keeper对provider进行支付，不可重复支付，
//支付需要的信息：支付者 接受者 金额（时空值） 时间（段） proof
//先根据本地保存的挑战信息 汇总一段时间内的挑战结果，然后将挑战结果发给同组的其他keeper（同步），收到其他keeper的确认信息后，进行支付操作
func doSpaceTimePay(groupid string, pidString string, price int64) {
	if localKeeperIsMaster(groupid) { //只有master节点进行支付过程
		fmt.Println(">>>>>>>>>>>>spacetimepay>>>>>>>>>>>>")
		defer fmt.Println("=====spacetimepay=====")
		fmt.Printf("groupid:%s:\npid:%s\n", groupid, pidString)
		config, _ := localNode.Repo.Config()
		endPoint := config.Eth                                              //获取endPoint
		scGroupid, _ := ad.GetAddressFromID(groupid)                        //获得userAddress
		ukaddr, uk, err := contracts.GetUKFromResolver(endPoint, scGroupid) //查询合约
		if err != nil {
			fmt.Println("contracts.GetUKFromResolver() err:", err)
			return
		}
		ukBalance, err := contracts.QueryBalance(endPoint, ukaddr) //查询合约价格
		if err != nil {
			fmt.Println("contracts.QueryBalance() err:", err)
			return
		}
		fmt.Printf("ukaddr:%s\nukbalance:%s\n", ukaddr, ukBalance.String())

		startTime := checkLastPayTime(groupid, pidString)
		spaceTime, lastTime := resultSummary(groupid, pidString, startTime, utils.GetUnixNow()) //根据时间段获取时空值
		amount := convertSpacetime(spaceTime, price)                                            //将时空值转换成支付金额
		if amount.Sign() <= 0 {
			return
		}
		pAddr, _ := ad.GetAddressFromID(pidString)                                   //providerAddress
		hexPk, err := fr.GetHexPrivKeyFromKS(localNode.Identity, localNode.Password) //得到本节点的私钥
		if err != nil {
			fmt.Println("GetHexPrivKeyFromKS() failed:", err)
			return
		}
		fmt.Printf("amount:%d\nbegin_time:%s\nlast_time:%s\n", amount, utils.UnixToTime(startTime), utils.UnixToTime(lastTime))

		err = contracts.SpaceTimePay(uk, endPoint, scGroupid, pAddr, hexPk, amount) //进行支付
		if err != nil {
			fmt.Println("contracts.SpaceTimePay() failed:", err)
			return
		}

		km, metaValue, err := saveLastPay(groupid, pidString, "signature", "proof", startTime, lastTime, spaceTime)
		if err != nil {
			fmt.Println("saveLastPay() failed:", err)
			return
		}
		km.SetKeyType(metainfo.Sync)
		metaSyncTo(km, metaValue) //此次支付结果同步到其他的节点
		fmt.Println("spaceTimePay complete!")
	}
}

//convertSpacetime 将时空值转换成支付金额的函数
//price是部署upkeeping合约时设置的单价，单位：每MB每天
func convertSpacetime(spacetime *big.Int, price int64) *big.Int {
	amount := big.NewInt(0)
	if spacetime.Sign() <= 0 || price <= 0 {
		fmt.Println("error! spaceTime:", spacetime.String(), "price:", price)
		return amount
	}
	amount.Mul(spacetime, big.NewInt(price))
	amount.Quo(amount, big.NewInt(1024*1024*60*60*24)) //注意这里先用时空值×单位，计算出来更加准确
	if amount.Sign() <= 0 {
		fmt.Println("error! spaceTime:", spacetime, "amount:", amount, "price:", price)
		return amount
	}
	return amount
}

//进行一次挑战结果的汇总
//传入user和provider的id，返回时空值spacetime
func resultSummary(uid string, pid string, timeStart int64, timeEnd int64) (*big.Int, int64) {
	timeList, lenghList := fetchChalresult(uid, pid, timeStart, timeEnd) //取数据
	spacetime := big.NewInt(0)
	if len(timeList) <= 1 || len(lenghList) <= 1 {
		fmt.Println("no enough  challenge data")
		return big.NewInt(0), 0
	}
	timepre := timeList[0]
	lengthpre := lenghList[0]
	//初始化变量
	for index, timeafter := range timeList[1:] { //循环数组进行计算
		length := lenghList[index+1]
		spacetime.Add(spacetime, big.NewInt((timeafter-timepre)*int64(lengthpre+length)/2))
		timepre = timeafter
		lengthpre = length
	}
	if spacetime.Sign() < 0 {
		fmt.Println("error spacetime<0!\ntimeList:", timeList, "\nlenghlist:", lenghList)
	}
	return spacetime, timepre
}

type timesortlist []int64                 //该结构用来对挑战结果按时间进行排序，以便计算时空值
func (p timesortlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p timesortlist) Len() int           { return len(p) }
func (p timesortlist) Less(i, j int) bool { return p[i] < p[j] }

//从内存中取相应的挑战结果数据，并且进行排序
//传入user和provider的id号，返回两个数组，存放挑战结果中的时间和长度
//按照时间顺序排序
func fetchChalresult(uidString string, pidString string, timestart int64, timeend int64) ([]int64, []uint32) {
	var timeList []int64   //存放挑战时间序列
	var lenghList []uint32 //存放与挑战时间同序的数据长度序列
	var tsl timesortlist   //用来对挑战时间排序
	//取provider-user的挑战信息
	thisPU := PU{
		pid: pidString,
		uid: uidString,
	}
	thischalinfo, ok := getChalinfo(thisPU)
	if !ok {
		fmt.Println("fetchChalresult(),getchalinfo error!")
		return timeList, lenghList
	}
	thischalinfo.Time.Range(func(key, value interface{}) bool {
		if (key.(int64) >= timestart) && (key.(int64) < timeend) {
			tsl = append(tsl, key.(int64))
		}
		return true
	})
	sort.Sort(tsl) //取出传入的时间区间内的时间数据，进行排序
	for _, key := range tsl {
		chalres, ok := thischalinfo.Time.Load(key)
		if !ok {
			fmt.Println("获取挑战数据失败 time:", utils.UnixToTime(key))
		}
		timeList = append(timeList, key)
		lengthtemp := chalres.(*chalresult).length
		lenghList = append(lenghList, lengthtemp)
	} //用排好序的key，整理出时间与长度的列表
	return timeList, lenghList
}

//saveLastPay 支付完成或者同步操作时，记录信息,返回支付信息的keymeta结构体和metavalue
func saveLastPay(groupidString, pidString, signature, proof string, beginTime, endTime int64, spaceTime *big.Int) (*metainfo.KeyMeta, string, error) {
	//最近一次支付信息，保存在本地 `uid/"local"/"lastpay"/pid`,`begin_time/end_time/spacetime/signature/proof`
	kmLast, err := metainfo.NewKeyMeta(groupidString, metainfo.Local, metainfo.SyncTypeLastPay, pidString)
	if err != nil {
		fmt.Println("doSpaceTimePay()NewKeyMeta()err", err)
		return nil, "", err
	}
	valueLast := strings.Join([]string{utils.UnixToString(beginTime), utils.UnixToString(endTime), spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(kmLast.ToString(), valueLast, "local")
	if err != nil {
		fmt.Println("CmdPutTo()error:", err)
		return nil, "", err
	}
	//支付信息，保存在内存和本地`uid/"sync"/"chalpay"/pid/begin_time/end_time` `spacetime/signature/proof`
	km, err := metainfo.NewKeyMeta(groupidString, metainfo.Local, metainfo.SyncTypeChalPay, pidString, utils.UnixToString(beginTime), utils.UnixToString(endTime))
	if err != nil {
		fmt.Println("doSpaceTimePay()NewKeyMeta()err", err)
		return nil, "", err
	}
	metaValue := strings.Join([]string{spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
	err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(km.ToString(), metaValue, "local") //持久化保存支付信息
	if err != nil {
		fmt.Println("CmdPutTo()error:", err)
		return nil, "", err
	}
	//将此次支付作为最近一次支付，保存在内存中
	thisPU := PU{
		pid: pidString,
		uid: groupidString,
	}
	thisChalPay := &chalpay{
		begin_time: beginTime,
		end_time:   endTime,
		pid:        pidString,
		proof:      "proof",
		signature:  "signature",
		spacetime:  spaceTime,
		uid:        groupidString,
	}
	PayInfo.Store(thisPU, thisChalPay)
	return km, metaValue, nil
}

//获得最后一次支付的信息,最后一次的支付信息由master进行同步，会同时保存在内存和本地，先检查内存中的保存结果，若没有，则检查本地
func checkLastPayTime(groupidString string, pidString string) int64 {
	failtime := int64(0) //出错时，返回0时间戳
	thisPU := PU{
		pid: pidString,
		uid: groupidString,
	}
	thisPayOfProvider, ok := PayInfo.Load(thisPU) //在内存中查找
	if !ok {                                      //没有找到 在本地查找
		kmLast, err := metainfo.NewKeyMeta(groupidString, metainfo.Local, metainfo.SyncTypeLastPay, pidString)
		if err != nil {
			fmt.Println(err)
			return failtime
		}
		valueByte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(kmLast.ToString(), "local") //在levelDB中查找最近一次支付信息
		if err != nil {                                                                           //硬盘中也没找到，说明没存。返回failtime
			fmt.Println("no lastTime data,return Unix(0)")
			return failtime
		}
		valueString := string(valueByte)
		_, thisChalPay, err := parseLastPayKV(kmLast, valueString)
		if err != nil { //解析出错，一般不会发生
			fmt.Println("checkLastPayTime() parseLastPayKV() error!", err)
			return failtime
		}
		return thisChalPay.end_time
	}

	thisChalPay := thisPayOfProvider.(*chalpay)
	return thisChalPay.end_time
}

//parseLastPayKV 传入lastPay的KV，解析成 PU和*chalpay结构体
//`uid/"local"/"lastpay"/pid` ,`begin_time/end_time/spacetime/signature/proof`
func parseLastPayKV(keyMeta *metainfo.KeyMeta, value string) (PU, *chalpay, error) {
	splitedValue := strings.Split(value, metainfo.DELIMITER)
	if len(splitedValue) < 5 {
		return PU{}, nil, ErrParaseMetaFailed
	}
	uidString := keyMeta.GetMid()
	options := keyMeta.GetOptions()
	if len(options) < 2 {
		return PU{}, nil, metainfo.ErrIllegalKey
	}
	pidString := options[1]
	thisPU := PU{
		pid: pidString,
		uid: uidString,
	}
	st, ok := big.NewInt(0).SetString(splitedValue[2], 0)
	if !ok {
		fmt.Println("SetString()err!value:", splitedValue[2])
	}
	beginTime := utils.StringToUnix(splitedValue[0])
	endTime := utils.StringToUnix(splitedValue[1])
	if beginTime == 0 || endTime == 0 {
		fmt.Println("key:", keyMeta.ToString(), "\nvalue:", value)
		return PU{}, nil, metainfo.ErrIllegalValue
	}
	thischalPay := &chalpay{
		pid:        pidString,
		uid:        uidString,
		begin_time: beginTime,
		end_time:   endTime,
		spacetime:  st,
		signature:  splitedValue[3],
		proof:      splitedValue[4],
	}
	return thisPU, thischalPay, nil
}
