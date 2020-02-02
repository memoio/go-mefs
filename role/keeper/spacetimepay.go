package keeper

import (
	"context"
	"errors"
	"math/big"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

func (k *Info) stPayRegular(ctx context.Context) {
	utils.MLogger.Info("SpaceTime Pay start!")
	ticker := time.NewTicker(SPACETIMEPAYTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			uqs := k.getQUKeys()
			for _, uq := range uqs {
				if uq.qid == uq.uid {
					continue
				}
				qid := uq.qid

				thisGroup := k.getGroupInfo(uq.uid, qid, false)
				if thisGroup == nil {
					continue
				}

				for _, proID := range thisGroup.providers {
					err := thisGroup.spaceTimePay(proID, k.sk)
					if err != nil {
						continue
					}
					k.savePay(qid, proID)
				}

			}
		}
	}
}

func (g *groupInfo) spaceTimePay(proID, localSk string) error {

	utils.MLogger.Info(">>>>>>>>>>>>spacetimepay>>>>>>>>>>>>")
	defer utils.MLogger.Info("========spacetimepay========")
	if !g.isMaster(proID) {
		return errors.New("fail to pay")
	}

	if g.upkeeping != nil {
		return errors.New("fail to pay")
	}

	// TODO: exit when balance is too low

	price := g.upkeeping.Price

	// check again
	found := false
	for _, pid := range g.upkeeping.ProviderIDs {
		if pid == proID {
			found = true
			break
		}
	}

	// PosAdd
	if !found {
		if g.userID == pos.GetPosId() {
			providerAddr, err := address.GetAddressFromID(proID)
			if err != nil {
				return err
			}

			userAddr, err := address.GetAddressFromID(pos.GetPosId())
			if err != nil {
				return err
			}

			queryAddr, err := address.GetAddressFromID(pos.GetPosGID())
			if err != nil {
				return err
			}

			err = contracts.AddProvider(pos.PosSkStr, userAddr, userAddr, []common.Address{providerAddr}, queryAddr.String())
			if err != nil {
				utils.MLogger.Info("st AddProvider() error: ", err)
				return err
			}

			price = pos.GetPosPrice()
		} else {
			return errors.New("fail to pay")
		}
	}

	thisIlinfo, ok := g.ledgerMap.Load(proID)
	if !ok {
		return errors.New("No such provider")
	}

	thisLinfo := thisIlinfo.(*lInfo)

	startTime := g.upkeeping.StartTime
	if thisLinfo.lastPay != nil {
		startTime = thisLinfo.lastPay.endTime
	}

	spaceTime, lastTime := thisLinfo.resultSummary(startTime, utils.GetUnixNow())
	amount := convertSpacetime(spaceTime, price)
	if amount.Sign() > 0 {
		pAddr, _ := address.GetAddressFromID(proID) //providerAddress
		ukAddr, _ := address.GetAddressFromID(g.upkeeping.UpKeepingID)
		utils.MLogger.Infof("amount:%d,beginTime:%s, lastTime:%s", amount, utils.UnixToTime(startTime), utils.UnixToTime(lastTime))

		err := contracts.SpaceTimePay(ukAddr, pAddr, localSk, amount) //进行支付
		if err != nil {
			utils.MLogger.Info("contracts.SpaceTimePay() failed: ", err)
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
		utils.MLogger.Error("error! spaceTime:", spacetime.String(), ", price:", price)
		return amount
	}
	amount.Mul(spacetime, big.NewInt(price))
	amount.Quo(amount, big.NewInt(1024*1024*60*60*24)) //注意这里先用时空值×单位，计算出来更加准确
	if amount.Sign() <= 0 {
		utils.MLogger.Info("error! spaceTime:", spacetime, "amount:", amount, "price:", price)
		return amount
	}
	return amount
}

// challeng results to spacetime value
// lastTime is the lastest challenge time which is before Now
func (l *lInfo) resultSummary(start, end int64) (*big.Int, int64) {
	var timeList []int64  //存放挑战时间序列
	var lenghList []int64 //存放与挑战时间同序的数据长度序列
	var tsl timesortlist  //用来对挑战时间排序
	spacetime := big.NewInt(0)

	var deletes []int64

	l.chalMap.Range(func(k, value interface{}) bool {
		// remove paid challenges
		key := k.(int64)
		if key < start {
			deletes = append(deletes, key)
		} else if key < end {
			tsl = append(tsl, key)
		}

		return true
	})

	for _, d := range deletes {
		l.chalMap.Delete(d)
	}

	sort.Sort(tsl) //取出传入的时间区间内的时间数据，进行排序
	for _, key := range tsl {
		chalresI, ok := l.chalMap.Load(key)
		if !ok {
			utils.MLogger.Info("fetch challenge results err, time:", utils.UnixToTime(key))
		}

		chalres := chalresI.(*pb.ChalInfo)

		timeList = append(timeList, key)
		lenghList = append(lenghList, chalres.TotalLength)
	}

	if len(timeList) <= 1 || len(lenghList) <= 1 {
		utils.MLogger.Info("no enough challenge data")
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
		utils.MLogger.Info("error spacetime<0!\ntimeList:", timeList, "\nlenghlist:", lenghList)
	}
	return spacetime, timepre
}

type timesortlist []int64                 //该结构用来对挑战结果按时间进行排序，以便计算时空值
func (p timesortlist) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p timesortlist) Len() int           { return len(p) }
func (p timesortlist) Less(i, j int) bool { return p[i] < p[j] }

// parseLastPayKV from value to payInfo
//`qid/"lastpay"/pid` ,`beginTime/endTime/spacetime/signature/proof`
func (l *lInfo) parseLastPayKV(value []byte) error {
	splitedValue := strings.Split(string(value), metainfo.DELIMITER)
	if len(splitedValue) < 5 {
		return errParaseMetaFailed
	}

	st, ok := big.NewInt(0).SetString(splitedValue[2], 10)
	if !ok {
		utils.MLogger.Info("SetString()err!value: ", splitedValue[2])
	}
	begintime := utils.StringToUnix(splitedValue[0])
	endtime := utils.StringToUnix(splitedValue[1])
	if begintime == 0 || endtime == 0 {
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
