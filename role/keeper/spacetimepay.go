package keeper

import (
	"context"
	"encoding/base64"
	"log"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/utils"
	ad "github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

//chalpay: for one pay informations
type chalpay struct {
	beginTime int64    // last end
	endTime   int64    // this end
	spacetime *big.Int // space time value
	signature string   // signature of spacetime
	proof     string
}

func spaceTimePayRegular(ctx context.Context) {
	log.Println("SpaceTime Pay start!")
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

func spaceTimePay() {
	pus := getPUKeysFromukpInfo()
	for _, pu := range pus {
		//only master pay
		if !isMasterKeeper(pu.uid, pu.pid) {
			continue
		}

		log.Println(">>>>>>>>>>>>spacetimepay>>>>>>>>>>>>")
		defer log.Println("========spacetimepay========")

		log.Printf("userid:%s:\npid:%s\n", pu.uid, pu.pid)
		ukItem, err := getUpkeeping(pu.uid)
		if err != nil {
			log.Println("contracts.GetUKItem err: ", err)
			return
		}

		// TODO: exit when balance is too low
		ukBalance, err := contracts.QueryBalance(ukItem.UpKeepingAddr)
		if err != nil {
			log.Println("contracts.QueryBalance() err: ", err)
			return
		}
		log.Printf("ukaddr:%s has balance:%s\n", ukItem.UpKeepingAddr, ukBalance.String())

		// check again
		found := false
		for _, ProID := range ukItem.ProviderIDs {
			if pu.pid == ProID {
				found = true
				break
			}
		}

		// PosAdd
		if !found {
			if pu.uid == pos.GetPosId() {
				providerAddr, err := ad.GetAddressFromID(pu.pid)
				if err != nil {
					return
				}

				userAddr, err := ad.GetAddressFromID(pos.GetPosId())
				if err != nil {
					return
				}
				err = contracts.AddProvider(pos.PosSkStr, userAddr, []common.Address{providerAddr})
				if err != nil {
					log.Println("st AddProvider() error", err)
					return
				}

				saveUpkeeping(pu.uid, true)
			} else {
				continue
			}
		}

		price := ukItem.Price
		if pu.uid == pos.GetPosId() {
			price = pos.GetPosPrice()
		}

		startTime := checkLastPayTime(pu)
		spaceTime, lastTime := resultSummary(pu, startTime, utils.GetUnixNow())
		amount := convertSpacetime(spaceTime, price)
		if amount.Sign() > 0 {
			pAddr, _ := ad.GetAddressFromID(pu.pid) //providerAddress
			scGroupid, _ := ad.GetAddressFromID(pu.uid)
			ukAddr := common.HexToAddress(ukItem.UpKeepingAddr[2:])
			skByte, _ := localNode.PrivateKey.Bytes()
			ipfsSk := base64.StdEncoding.EncodeToString(skByte)
			hexSk, err := utils.IPFSskToEthsk(ipfsSk)
			if err != nil {
				log.Println("GetHexSk failed: ", err)
				return
			}
			log.Printf("amount:%d\nbeginTime:%s\nlastTime:%s\n", amount, utils.UnixToTime(startTime), utils.UnixToTime(lastTime))

			err = contracts.SpaceTimePay(ukAddr, scGroupid, pAddr, hexSk, amount) //进行支付
			if err != nil {
				log.Println("contracts.SpaceTimePay() failed: ", err)
				return
			}
		}

		km, metaValue, err := saveLastPay(pu, "signature", "proof", startTime, lastTime, spaceTime)
		if err != nil {
			log.Println("saveLastPay() failed: ", err)
			return
		}
		if amount.Sign() > 0 {
			km.SetKeyType(metainfo.Sync)
			metaSyncTo(km, metaValue) //send this value to other keepers
		}
		log.Println("spaceTimePay complete!")
	}
}

//price: MB/day
func convertSpacetime(spacetime *big.Int, price int64) *big.Int {
	amount := big.NewInt(0)
	if spacetime.Sign() <= 0 || price <= 0 {
		log.Println("error! spaceTime:", spacetime.String(), "price:", price)
		return amount
	}
	amount.Mul(spacetime, big.NewInt(price))
	amount.Quo(amount, big.NewInt(1024*1024*60*60*24)) //注意这里先用时空值×单位，计算出来更加准确
	if amount.Sign() <= 0 {
		log.Println("error! spaceTime:", spacetime, "amount:", amount, "price:", price)
		return amount
	}
	return amount
}

// challeng results to spacetime value
// lastTime is the lastest challenge time which is before Now
func resultSummary(thisPU puKey, timeStart int64, timeEnd int64) (*big.Int, int64) {
	timeList, lenghList := fetchChalresult(thisPU, timeStart, timeEnd) //取数据
	spacetime := big.NewInt(0)
	if len(timeList) <= 1 || len(lenghList) <= 1 {
		log.Println("no enough challenge data")
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
		log.Println("error spacetime<0!\ntimeList:", timeList, "\nlenghlist:", lenghList)
	}
	return spacetime, timepre
}

type timesortlist []int64                 //该结构用来对挑战结果按时间进行排序，以便计算时空值
func (p timesortlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p timesortlist) Len() int           { return len(p) }
func (p timesortlist) Less(i, j int) bool { return p[i] < p[j] }

//get from ledger chalMap
func fetchChalresult(thisPU puKey, timestart int64, timeend int64) ([]int64, []int64) {
	var timeList []int64  //存放挑战时间序列
	var lenghList []int64 //存放与挑战时间同序的数据长度序列
	var tsl timesortlist  //用来对挑战时间排序

	thischalinfo, ok := getChalinfo(thisPU)
	if !ok {
		log.Println("fetchChalresult(),getchalinfo error!")
		return timeList, lenghList
	}
	thischalinfo.chalMap.Range(func(key, value interface{}) bool {
		// remove paid challenges
		if key.(int64) < timestart {
			thischalinfo.chalMap.Delete(key)
		} else if key.(int64) < timeend {
			tsl = append(tsl, key.(int64))
		}

		return true
	})
	sort.Sort(tsl) //取出传入的时间区间内的时间数据，进行排序
	for _, key := range tsl {
		chalres, ok := thischalinfo.chalMap.Load(key)
		if !ok {
			log.Println("fetch challenge results err, time:", utils.UnixToTime(key))
		}
		timeList = append(timeList, key)
		lengthtemp := chalres.(*chalresult).length
		lenghList = append(lenghList, lengthtemp)
	} //用排好序的key，整理出时间与长度的列表
	return timeList, lenghList
}

func saveLastPay(thisPU puKey, signature, proof string, beginTime, endTime int64, spaceTime *big.Int) (*metainfo.KeyMeta, string, error) {
	//key: `uid/"local"/"lastpay"/pid`
	//value: `beginTime/endTime/spacetime/signature/proof`
	//for get
	kmLast, err := metainfo.NewKeyMeta(thisPU.uid, metainfo.Local, metainfo.SyncTypeLastPay, thisPU.pid)
	if err != nil {
		log.Println("doSpaceTimePay()NewKeyMeta()err: ", err)
		return nil, "", err
	}
	valueLast := strings.Join([]string{utils.UnixToString(beginTime), utils.UnixToString(endTime), spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
	localNode.Data.PutKey(context.Backgroud(), kmLast.ToString(), []byte(valueLast), "local")
	//key: `uid/"sync"/"chalpay"/pid/beginTime/endTime`
	//value: `spacetime/signature/proof`
	//for storing
	km, err := metainfo.NewKeyMeta(thisPU.uid, metainfo.Local, metainfo.SyncTypeChalPay, thisPU.pid, utils.UnixToString(beginTime), utils.UnixToString(endTime))
	if err != nil {
		log.Println("doSpaceTimePay()NewKeyMeta()err: ", err)
		return nil, "", err
	}
	metaValue := strings.Join([]string{spaceTime.String(), "signature", "proof"}, metainfo.DELIMITER)
	localNode.Data.PutKey(context.Backgroud(), km.ToString(), []byte(metaValue), "local")

	//将此次支付作为最近一次支付，保存在内存中
	thisChalPay := &chalpay{
		beginTime: beginTime,
		endTime:   endTime,
		proof:     "proof",
		signature: "signature",
		spacetime: spaceTime,
	}

	thisChalinfo, ok := getChalinfo(thisPU)
	if ok {
		thisChalinfo.lastPay = thisChalPay
	}

	return km, metaValue, nil
}

//获得最后一次支付的信息,最后一次的支付信息由master进行同步，会同时保存在内存和本地，先检查内存中的保存结果，若没有，则检查本地
func checkLastPayTime(thisPU puKey) int64 {
	failtime := int64(0)

	thisChalinfo, ok := getChalinfo(thisPU)
	if !ok {
		return failtime
	}

	if thisChalinfo.lastPay == nil {
		kmLast, err := metainfo.NewKeyMeta(thisPU.uid, metainfo.Local, metainfo.SyncTypeLastPay, thisPU.uid)
		if err != nil {
			log.Println(err)
			return failtime
		}
		// get from leveldb
		valueByte, err := localNode.Data.GetKey(kmLast.ToString(), "local")
		if err != nil {
			log.Println("no lastTime data, return Unix(0)")
			return failtime
		}
		valueString := string(valueByte)
		_, thisChalPay, err := parseLastPayKV(kmLast, valueString)
		if err != nil {
			log.Println("checkLastPayTime() parseLastPayKV() err: ", err)
			return failtime
		}
		return thisChalPay.endTime
	}

	return thisChalinfo.lastPay.endTime
}

//parseLastPayKV 传入lastPay的KV，解析成 PU和*chalpay结构体
//`uid/"local"/"lastpay"/pid` ,`beginTime/endTime/spacetime/signature/proof`
func parseLastPayKV(keyMeta *metainfo.KeyMeta, value string) (puKey, *chalpay, error) {
	splitedValue := strings.Split(value, metainfo.DELIMITER)
	if len(splitedValue) < 5 {
		return puKey{}, nil, errParaseMetaFailed
	}
	uidString := keyMeta.GetMid()
	options := keyMeta.GetOptions()
	if len(options) < 2 {
		return puKey{}, nil, metainfo.ErrIllegalKey
	}
	pidString := options[1]
	thisPU := puKey{
		pid: pidString,
		uid: uidString,
	}
	st, ok := big.NewInt(0).SetString(splitedValue[2], 0)
	if !ok {
		log.Println("SetString()err!value: ", splitedValue[2])
	}
	begintime := utils.StringToUnix(splitedValue[0])
	endtime := utils.StringToUnix(splitedValue[1])
	if begintime == 0 || endtime == 0 {
		log.Println("key:", keyMeta.ToString(), "\nvalue:", value)
		return puKey{}, nil, metainfo.ErrIllegalValue
	}
	thischalPay := &chalpay{
		beginTime: begintime,
		endTime:   endtime,
		spacetime: st,
		signature: splitedValue[3],
		proof:     splitedValue[4],
	}
	return thisPU, thischalPay, nil
}
