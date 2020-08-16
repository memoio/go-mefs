package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/memoio/go-mefs/contracts"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
	mt "gitlab.com/NebulousLabs/merkletree"
)

func (k *Info) stPrePayRegular(ctx context.Context) {
	utils.MLogger.Info("SpaceTime Pre Pay start!")
	time.Sleep(4 * time.Minute)
	k.stPrePayAll(ctx)
	ticker := time.NewTicker(spaceTimePayTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.stPrePayAll(ctx)
		}
	}
}

func (k *Info) stPrePayAll(ctx context.Context) {
	uqs := k.getQUKeys()
	for _, uq := range uqs {
		if uq.qid == uq.uid {
			continue
		}

		thisGroup := k.getGroupInfo(uq.uid, uq.qid, false)
		if thisGroup == nil {
			continue
		}

		thisGroup.loadContracts(true)

		if thisGroup.upkeeping == nil {
			continue
		}

		if uq.uid == pos.GetPosId() {
			utils.MLogger.Info("SpaceTime Pay for pos user")
			thisGroup.upkeeping.Price = k.getPosPrice()
		}

		for _, proID := range thisGroup.providers {
			thisGroup.stPrePay(k.context, proID, k.sk, k.localID, k.ds)
		}
	}
}

func (k *Info) stPayRegular(ctx context.Context) {
	utils.MLogger.Info("SpaceTime Pay start!")
	time.Sleep(5 * time.Minute)
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			k.stPayAll(ctx)
		}
	}
}

func (k *Info) stPayAll(ctx context.Context) {
	uqs := k.getQUKeys()
	for _, uq := range uqs {
		if uq.qid == uq.uid {
			continue
		}

		thisGroup := k.getGroupInfo(uq.uid, uq.qid, false)
		if thisGroup == nil || thisGroup.upkeeping == nil {
			continue
		}

		for _, proID := range thisGroup.providers {
			err := thisGroup.stPay(k.context, proID, k.sk, k.localID, k.ds)
			if err != nil {
				continue
			}
			k.savePay(uq.uid, uq.qid, proID)
		}
	}

	balance := role.GetBalance(k.localID)
	ba, _ := new(big.Float).SetInt(balance).Float64()
	k.ms.balance.Set(ba)
}

func (g *groupInfo) stPrePay(ctx context.Context, proID, localSk, localID string, ds data.Service) error {
	// only master triggers,
	// todo round-robin
	if !g.isMaster(proID) {
		return nil
	}

	if g.upkeeping == nil {
		return nil
	}

	price := g.upkeeping.Price

	// check again
	found := false
	startTime := g.upkeeping.StartTime
	for _, pInfo := range g.upkeeping.Providers {
		pid, err := address.GetIDFromAddress(pInfo.Addr.String())
		if err != nil {
			return err
		}

		if pInfo.Stop {
			return nil
		}

		if pid == proID {
			found = true
			startTime = pInfo.StEnd.Int64()
			if time.Now().Unix()-startTime < payInternval {
				utils.MLogger.Infof("SpaceTimePay is not on time for user %s fsID %s at %s", g.userID, g.groupID, proID)
				return nil
			}
			break
		}
	}

	if !found {
		return role.ErrNotMyProvider
	}

	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return role.ErrNotMyProvider
	}

	if startTime >= g.upkeeping.EndTime && g.userID != pos.GetPosId() {
		utils.MLogger.Infof("SpaceTimePay expired for user %s fsID %s at %s", g.userID, g.groupID, proID)
		return role.ErrUkExpire
	}

	utils.MLogger.Infof("SpaceTimePay start for user %s fsID %s at %s", g.userID, g.groupID, proID)

	if thisLinfo.currentPay == nil {
		endTime := time.Now().Unix()
		if endTime > g.upkeeping.EndTime {
			endTime = g.upkeeping.EndTime
		}

		amount, mroot := thisLinfo.stSummary(price, startTime, endTime)
		chalfrequency := thisLinfo.stShare(startTime, endTime)
		if amount.Sign() > 0 && len(mroot) >= 32 {
			needPay := new(big.Int).Add(g.upkeeping.NeedPay, amount)
			if needPay.Cmp(g.upkeeping.Money) > 0 {
				utils.MLogger.Infof("SpaceTimePay start pay for user %s fsID %s pro %s from %d fails due to no enough money in upkeeping", g.userID, g.groupID, proID, startTime)
				return role.ErrNotEnoughBalance
			}

			knum := len(g.keepers)
			cpay := &chalpay{
				STValue: mpb.STValue{
					Start:  startTime,
					Length: endTime - startTime,
					Value:  amount.Bytes(),
					Status: int32(knum * 2 / 3), // need at least 2/3
					Sign:   make([][]byte, knum),
					Share:  make([]int64, knum+1),
					Root:   mroot,
				},
				checkNum: 0,
			}

			thisLinfo.currentPay = cpay

			mkey, err := metainfo.NewKey(g.groupID, mpb.KeyType_StPayShare, g.userID, proID, localID, time.Unix(startTime, 0).Format(utils.BASETIME), time.Unix(endTime, 0).Format(utils.BASETIME))
			if err != nil {
				return err
			}
			key := mkey.ToString()
			for i, kid := range g.keepers {
				if kid == localID {
					cpay.Share[i] = int64(chalfrequency)
					continue
				}
				go ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, nil, nil, kid)
			}

			var shareSum int64
			for i := 0; i < 6; i++ {
				shareSum = 0
				time.Sleep(30 * time.Second)
				tmp := 0
				for j := 0; j < knum; j++ {
					shareSum += cpay.Share[j]
					if cpay.Share[j] != 0 {
						tmp++
					}
				}
				if tmp == knum {
					break
				}
			}
			cpay.Share[knum] = shareSum
			utils.MLogger.Infof("get Share:", g.keepers, cpay.Share)

			st := big.NewInt(cpay.Start)
			sl := big.NewInt(cpay.Length)
			sv := new(big.Int).SetBytes(cpay.Value)
			hash, err := role.GetHashForST(g.upkeeping.UpKeepingID, proID, st, sl, sv, cpay.Root[:32], cpay.Share)
			if err != nil {
				return err
			}

			sign, err := id.Sign(localSk, hash)
			if err != nil {
				return err
			}

			mkey, err = metainfo.NewKey(g.groupID, mpb.KeyType_StPaySign, g.userID, proID, localID, st.String(), sl.String())
			if err != nil {
				return err
			}

			// sync to other keepers？
			// get enough signs; then
			key = mkey.ToString()
			for i, kid := range g.keepers {
				if kid == localID {
					cpay.Lock()
					cpay.Sign[i] = sign
					cpay.Status--
					cpay.Unlock()
					continue
				}
				go ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, hash, sign, kid)
			}

			utils.MLogger.Infof("SpaceTimePay start for user %s fsID %s for provider %s at %d ", g.userID, g.groupID, proID, thisLinfo.currentPay.Start)
		} else {
			utils.MLogger.Infof("SpaceTimePay start for user %s fsID %s for provider %s at zero ", g.userID, g.groupID, proID)
		}
		return role.ErrEmptyData
	}

	return role.ErrEmptyData
}

func (g *groupInfo) stPay(ctx context.Context, proID, localSk, localID string, ds data.Service) error {
	// only master triggers,
	// todo round-robin
	if !g.isMaster(proID) {
		return nil
	}

	if g.upkeeping == nil {
		return nil
	}

	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return role.ErrNotMyProvider
	}

	if thisLinfo.currentPay != nil {
		cpay := thisLinfo.currentPay
		cpay.Lock()
		if cpay.Status <= 0 {
			pAddr, err := address.GetAddressFromID(proID)
			if err != nil {
				return err
			}
			ukAddr, err := address.GetAddressFromID(g.upkeeping.UpKeepingID)
			if err != nil {
				return err
			}

			st := big.NewInt(cpay.Start)
			sl := big.NewInt(cpay.Length)
			sv := new(big.Int).SetBytes(cpay.Value)

			utils.MLogger.Infof("SpaceTimePay start pay for user %s fsID %s pro %s from %d, length %d value %d", g.userID, g.groupID, proID, st, sl, sv)

			var root [32]byte
			copy(root[:], cpay.Root[:32])
			err = contracts.SpaceTimePay(ukAddr, pAddr, localSk, st, sl, sv, root, cpay.Share, cpay.Sign)
			if err != nil {
				utils.MLogger.Infof("SpaceTimePay start pay for user %s fsID %s pro %s from %s, length %s value %s failed %s", g.userID, g.groupID, proID, st.String(), sl.String(), sv.String(), err)
				cpay.Unlock()
				thisLinfo.currentPay = nil
				return err
			}

			if g.userID == pos.GetPosId() {
				utils.MLogger.Infof("SpaceTimePay start pay for pos user %s fsID %s pro %s from %s, length %s value %s success", g.userID, g.groupID, proID, st.String(), sl.String(), sv.String())
			} else {
				utils.MLogger.Infof("SpaceTimePay start pay for user %s fsID %s pro %s from %s, length %s value %s success", g.userID, g.groupID, proID, st.String(), sl.String(), sv.String())
			}

			cpay.Status = 0
			cpay.Unlock()
			thisLinfo.lastPay = cpay
			thisLinfo.currentPay = nil
			g.loadContracts(true)
			return nil
		}

		cpay.checkNum++
		if cpay.checkNum > 3 {
			thisLinfo.currentPay = nil
			cpay.Unlock()
			return role.ErrEmptyData
		}

		// resend to collet sign
		mkey, err := metainfo.NewKey(g.groupID, mpb.KeyType_StPaySign, g.userID, proID, localID, strconv.FormatInt(cpay.GetStart(), 10), strconv.FormatInt(cpay.GetLength(), 10))
		if err != nil {
			cpay.Unlock()
			return err
		}

		key := mkey.ToString()

		var sign []byte
		for i, kid := range g.keepers {
			if kid == localID {
				sign = cpay.GetSign()[i]
				break
			}
		}

		for _, kid := range g.keepers {
			if kid == localID {
				continue
			}
			go ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, cpay.hash, sign, kid)
		}

		cpay.Status = int32(len(g.keepers) * 2 / 3)

		cpay.Unlock()
	}

	return nil
}

type timeValue struct {
	time  int64
	space int64
}

// challeng results to spacetime value
// lastTime is the lastest challenge time which is before Now
func (l *lInfo) stSummary(price *big.Int, start, end int64) (*big.Int, []byte) {
	spacetime := big.NewInt(0)
	var tsl []timeValue //用来对挑战时间排序

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
				space: chalres.SuccessLength,
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
		return spacetime, nil
	}

	sort.Slice(tsl, func(i, j int) bool {
		return tsl[i].time < tsl[j].time
	})

	var newTsl []timeValue

	if tsl[0].time > start && tsl[0].time < start+3600 {
		tv := timeValue{
			time:  start,
			space: tsl[0].space,
		}
		newTsl = append(newTsl, tv)
	} else {
		tv := timeValue{
			time:  start,
			space: 0,
		}
		newTsl = append(newTsl, tv)
	}

	// at least once per hour
	ftime := start + 3600
	i := 0
	for {
		if i >= len(tsl) || ftime < tsl[i].time {
			tv := timeValue{
				time:  ftime,
				space: 0,
			}
			ftime += 3600
			newTsl = append(newTsl, tv)
		} else {
			newTsl = append(newTsl, tsl[i])
			ftime = tsl[i].time + 3600
			i++
		}

		if ftime > end && i == len(tsl) {
			break
		}
	}

	tLen := len(newTsl)

	if newTsl[tLen-1].time < end && newTsl[tLen-1].time > end-3600 {
		tv := timeValue{
			time:  end,
			space: newTsl[tLen-1].space,
		}
		newTsl = append(newTsl, tv)
	} else {
		tv := timeValue{
			time:  end,
			space: 0,
		}
		newTsl = append(newTsl, tv)
	}

	mtree := mt.New(sha256.New())
	mtree.SetIndex(0)

	var nbuf bytes.Buffer
	enc := gob.NewEncoder(&nbuf)

	timepre := newTsl[0].time
	lengthpre := newTsl[0].space
	for _, tv := range newTsl[1:] {
		spacetime.Add(spacetime, big.NewInt((tv.time-timepre)*int64(lengthpre+tv.space)/2))
		timepre = tv.time
		lengthpre = tv.space
		enc.Encode(tv)
		mtree.Push(nbuf.Bytes())
	}
	if spacetime.Sign() <= 0 {
		utils.MLogger.Info("error spacetime<=0")
	}

	spacetime.Mul(spacetime, price)
	spacetime.Quo(spacetime, big.NewInt(1024*1024*60*60))

	stWei := new(big.Float).SetInt(spacetime)
	stWei.Quo(stWei, role.GetMemoPrice())
	stWei.Int(spacetime)

	if spacetime.Sign() <= 0 {
		utils.MLogger.Info("error!amount:", spacetime, "price:", price)
	}

	utils.MLogger.Debug("spacetime  calc is:", spacetime)
	utils.MLogger.Debug(tsl)
	return spacetime, mtree.Root()
}

//get challenge frequency
func (l *lInfo) stShare(start, end int64) int {
	chalFrequency := 0

	l.chalMap.Range(func(k, value interface{}) bool {
		key := k.(int64)
		if key >= start && key < end {
			chalFrequency++
		}

		return true
	})
	return chalFrequency
}
