package keeper

import (
	"context"
	"math/big"
	"sort"
	"time"

	"github.com/memoio/go-mefs/contracts"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/pos"
)

func (k *Info) stPayRegular(ctx context.Context) {
	utils.MLogger.Info("SpaceTime Pay start!")
	ticker := time.NewTicker(spaceTimePayTime)
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
	// only master triggers,
	// todo round-robin
	if !g.isMaster(proID) {
		return nil
	}

	if g.upkeeping != nil {
		return nil
	}

	// TODO: exit when balance is too low

	price := g.upkeeping.Price

	// check again
	found := false
	stEnd := big.NewInt(0)
	for _, pInfo := range g.upkeeping.Providers {
		pid, err := address.GetIDFromAddress(pInfo.Addr.String())
		if err != nil {
			return err
		}
		if pid == proID {
			found = true
			stEnd = pInfo.StEnd
			break
		}
	}

	// PosAdd
	if !found {
		if g.userID != pos.GetPosId() {
			return nil
		}

		price = pos.GetPosPrice()
	}

	thisIlinfo, ok := g.ledgerMap.Load(proID)
	if !ok {
		return role.ErrNotMyProvider
	}

	thisLinfo := thisIlinfo.(*lInfo)

	var startTime int64
	if thisLinfo.currentPay == nil {
		startTime = g.upkeeping.StartTime
		if thisLinfo.lastPay == nil {
			thisLinfo.lastPay = &chalpay{
				STValue: mpb.STValue{
					Start:  startTime,
					Length: 0,
					Status: 0,
				},
			}
		}

		// handle last pay
		if thisLinfo.lastPay.Status != 0 {
			return nil
		}

		startTime = thisLinfo.lastPay.Start + thisLinfo.lastPay.Length

		spaceTime, lastTime := thisLinfo.resultSummary(startTime, time.Now().Unix())
		amount := convertSpacetime(spaceTime, price)
		if amount.Sign() > 0 {
			thisLinfo.currentPay = &chalpay{
				STValue: mpb.STValue{
					Start:  startTime,
					Length: lastTime - startTime,
					Value:  amount.Bytes(),
					Status: int32(len(g.keepers)),
				},
			}
		}
		// sync to other keepers？
		// get enough signs; then
	}

	//TODO
	stStart := stEnd
	stLength := big.NewInt(time.Now().Unix() - stEnd.Int64())
	merkleRoot := [32]byte{0}
	share := []int{}
	sign := [][]byte{}

	spaceTime, _ := thisLinfo.resultSummary(startTime, time.Now().Unix())
	amount := convertSpacetime(spaceTime, price)
	if amount.Sign() > 0 {
		pAddr, _ := address.GetAddressFromID(proID) //providerAddress
		ukAddr, _ := address.GetAddressFromID(g.upkeeping.UpKeepingID)

		err := contracts.SpaceTimePay(ukAddr, pAddr, localSk, stStart, stLength, amount, merkleRoot, share, sign) //进行支付
		if err != nil {
			utils.MLogger.Info("contracts.SpaceTimePay() failed: ", err)
			return err
		}
	}

	thisLinfo.currentPay.Status = 0
	thisLinfo.lastPay = thisLinfo.currentPay
	thisLinfo.currentPay = nil

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

type timeValue struct {
	time  int64
	space int64
}

// challeng results to spacetime value
// lastTime is the lastest challenge time which is before Now
func (l *lInfo) resultSummary(start, end int64) (*big.Int, int64) {
	var tsl []timeValue //用来对挑战时间排序
	spacetime := big.NewInt(0)

	var deletes []int64

	l.chalMap.Range(func(k, value interface{}) bool {
		// remove paid challenges
		key := k.(int64)
		if key < start {
			deletes = append(deletes, key)
		} else if key < end {
			chalres := value.(*mpb.ChalInfo)
			tv := timeValue{
				time:  key,
				space: chalres.TotalLength,
			}
			tsl = append(tsl, tv)
		}

		return true
	})

	for _, d := range deletes {
		l.chalMap.Delete(d)
	}

	if len(tsl) <= 1 {
		utils.MLogger.Info("no enough challenge data")
		return spacetime, 0
	}

	sort.Slice(tsl, func(i, j int) bool {
		return tsl[i].time < tsl[j].time
	})

	timepre := tsl[0].time
	lengthpre := tsl[0].space
	for _, tv := range tsl[1:] {
		spacetime.Add(spacetime, big.NewInt((tv.time-timepre)*int64(lengthpre+tv.space)/2))
		timepre = tv.time
		lengthpre = tv.space
	}
	if spacetime.Sign() <= 0 {
		utils.MLogger.Info("error spacetime<=0")
	}
	return spacetime, timepre
}
