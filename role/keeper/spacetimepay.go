package keeper

import (
	"context"
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

func (k *Info) stPayRegular(ctx context.Context) {
	log.Println("SpaceTime Pay start!")
	ticker := time.NewTicker(SPACETIMEPAYTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uqs := k.getUQKeys()
			for _, uq := range uqs {
				thisIGroup, ok := k.ukpGroup.Load(uq.pid)
				if !ok {
					continue
				}

				thisGroup := thisIGroup.(*groupInfo)

				for _, proID := range thisGroup.providers {
					err := thisGroup.spaceTimePay(proID)
					if err != nil {
						continue
					}
					k.savePay(qid, proID)
				}

			}
		}
	}
}

func (g *groupInfo) spaceTimePay(proID string) error {

	log.Println(">>>>>>>>>>>>spacetimepay>>>>>>>>>>>>")
	defer log.Println("========spacetimepay========")
	if !g.isMaster(proID) {
		return errors.New("fail to pay")
	}

	if g.upkeeping != nil {
		return errors.New("fail to pay")
	}

	// TODO: exit when balance is too low
	ukBalance, err := contracts.QueryBalance(g.upkeeping.UpKeepingAddr)
	if err != nil {
		log.Println("contracts.QueryBalance() err: ", err)
		return err
	}
	log.Printf("ukaddr:%s has balance:%s\n", ukItem.UpKeepingAddr, ukBalance.String())

	price := g.upkeeping.Price

	// check again
	found := false
	for _, pid := range g.upkeeping.ProviderIDs {
		if pid == ProID {
			found = true
			break
		}
	}

	// PosAdd
	if !found {
		if g.groupID == pos.GetPosId() {
			providerAddr, err := ad.GetAddressFromID(proID)
			if err != nil {
				return err
			}

			userAddr, err := ad.GetAddressFromID(pos.GetPosId())
			if err != nil {
				return err
			}
			err = contracts.AddProvider(pos.PosSkStr, userAddr, []common.Address{providerAddr})
			if err != nil {
				log.Println("st AddProvider() error", err)
				return err
			}

			g.saveUpkeeping()
			price = pos.GetPosPrice()
		} else {
			return
		}
	}

	thisIlinfo, ok := g.ledgerMap.Load(proID)
	if !ok {
		return errors.New("No such provider")
	}

	thisLinfo := thisIlinfo.(*lInfo)

	startTime := utils.StringToTime(g.upkeeping.StartTime)
	if thisLinfo.lastPay != nil {
		startTime = thisLinfo.checkLastPayTime()
	}

	spaceTime, lastTime := thisLinfo.resultSummary(startTime, utils.GetUnixNow())
	amount := convertSpacetime(spaceTime, price)
	if amount.Sign() > 0 {
		pAddr, _ := ad.GetAddressFromID(pu.pid) //providerAddress
		scGroupid, _ := ad.GetAddressFromID(pu.qid)
		ukAddr := common.HexToAddress(ukItem.UpKeepingAddr[2:])
		log.Printf("amount:%d\nbeginTime:%s\nlastTime:%s\n", amount, utils.UnixToTime(startTime), utils.UnixToTime(lastTime))

		err = contracts.SpaceTimePay(ukAddr, scGroupid, pAddr, k.sk, amount) //进行支付
		if err != nil {
			log.Println("contracts.SpaceTimePay() failed: ", err)
			return err
		}
	}

	thisLinfo.lastPay = &chalpay{
		beginTime: startTime,
		endTime:   lastTime,
		proof:     "proof",
		signature: "signature",
		spacetime: spaceTime,
	}
	// sync to other keepers？
	return nil
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
func (l *lInfo) resultSummary(proID string, start, end int64) (*big.Int, int64) {
	var timeList []int64  //存放挑战时间序列
	var lenghList []int64 //存放与挑战时间同序的数据长度序列
	var tsl timesortlist  //用来对挑战时间排序
	spacetime := big.NewInt(0)

	var deletes []int64

	l.chalMap.Range(func(k, value interface{}) bool {
		// remove paid challenges
		key := k.(int64)
		if key < timeStart {
			deletes = append(deletes, key)
		} else if key < timeEnd {
			tsl = append(tsl, key)
		}

		return true
	})

	for _, d := range deletes {
		l.chalMap.Delete(d)
	}

	sort.Sort(tsl) //取出传入的时间区间内的时间数据，进行排序
	for _, key := range tsl {
		chalres, ok := l.chalMap.Load(key)
		if !ok {
			log.Println("fetch challenge results err, time:", utils.UnixToTime(key))
		}
		timeList = append(timeList, key)
		lengthtemp := chalres.(*chalresult).length
		lenghList = append(lenghList, lengthtemp)
	}

	if len(timeList) <= 1 || len(lenghList) <= 1 {
		log.Println("no enough challenge data")
		return spacetime, 0
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

//parseLastPayKV 传入lastPay的KV，解析成 PU和*chalpay结构体
//`qid/"lastpay"/pid` ,`beginTime/endTime/spacetime/signature/proof`
func (l *lInfo) parseLastPayKV(value []byte) error {
	splitedValue := strings.Split(string(value), metainfo.DELIMITER)
	if len(splitedValue) < 5 {
		return errParaseMetaFailed
	}

	splitedKey := strings.Split(key, metainfo.DELIMITER)
	if len(splitedKey) < 3 {
		return metainfo.ErrIllegalKey
	}

	st, ok := big.NewInt(0).SetString(splitedValue[2], 10)
	if !ok {
		log.Println("SetString()err!value: ", splitedValue[2])
	}
	begintime := utils.StringToUnix(splitedValue[0])
	endtime := utils.StringToUnix(splitedValue[1])
	if begintime == 0 || endtime == 0 {
		log.Println("key:", keyMeta.ToString(), "\nvalue:", value)
		return metainfo.ErrIllegalValue
	}
	l.lastPay = &chalpay{
		beginTime: begintime,
		endTime:   endtime,
		spacetime: st,
		signature: splitedValue[3],
		proof:     splitedValue[4],
	}

	return nil
}
