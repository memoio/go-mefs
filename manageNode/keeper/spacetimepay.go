package keeper

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"math/big"
	"sort"
	"time"

	"github.com/memoio/go-mefs/contracts"
	id "github.com/memoio/go-mefs/crypto/identity"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
	mt "gitlab.com/NebulousLabs/merkletree"
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

				expireCount := 0
				for _, proID := range thisGroup.providers {
					err := thisGroup.spaceTimePay(k.context, proID, k.sk, k.localID, k.ds)
					if err != nil {
						if err.Error() == role.ErrUkExpire.Error() {
							expireCount++
						}
						continue
					}
					k.savePay(uq.uid, qid, proID)
				}

				if expireCount == len(thisGroup.providers) {
					// all pay ends;
					k.ukpGroup.Delete(qid)
				}
			}
		}
	}
}

func (g *groupInfo) spaceTimePay(ctx context.Context, proID, localSk, localID string, ds data.Service) error {
	// only master triggers,
	// todo round-robin
	if !g.isMaster(proID) {
		return nil
	}

	if g.upkeeping == nil {
		return nil
	}

	// TODO: exit when balance is too low

	price := g.upkeeping.Price
	if g.userID == pos.GetPosId() {
		// todo, price depends on total pledge data
		price = pos.GetPosPrice()
	}

	// check again
	found := false
	startTime := g.upkeeping.StartTime
	for _, pInfo := range g.upkeeping.Providers {
		pid, err := address.GetIDFromAddress(pInfo.Addr.String())
		if err != nil {
			return err
		}
		if pid == proID {
			found = true
			startTime = pInfo.StEnd.Int64()
			break
		}
	}

	// PosAdd
	if !found {
		return nil
	}

	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return role.ErrNotMyProvider
	}

	if thisLinfo.lastPay != nil {
		startTime = thisLinfo.lastPay.GetStart() + thisLinfo.lastPay.GetLength()
	}

	if startTime > g.upkeeping.EndTime && g.userID != pos.GetPosId() {
		utils.MLogger.Infof("SpaceTimePay expired for user %s fsID %s at %s", g.userID, g.groupID, proID)
		return role.ErrUkExpire
	}

	utils.MLogger.Infof("SpaceTimePay start for user %s fsID %s at %s", g.userID, g.groupID, proID)

	if thisLinfo.currentPay == nil {
		amount, lastTime, mroot := thisLinfo.resultSummary(price, startTime, time.Now().Unix())
		if amount.Sign() > 0 {
			knum := len(g.keepers)
			cpay := &chalpay{
				STValue: mpb.STValue{
					Start:  startTime,
					Length: lastTime - startTime,
					Value:  amount.Bytes(),
					Status: int32(knum * 2 / 3), // need at least 2/3
					Sign:   make([][]byte, knum),
					Share:  make([]int64, knum+1),
					Root:   mroot,
				},
			}

			for i := 0; i < knum; i++ {
				cpay.Share[i] = 1
			}
			cpay.Share[knum] = int64(knum)

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

			mkey, err := metainfo.NewKey(g.groupID, mpb.KeyType_Sign, g.userID, proID, localID, st.String(), sl.String())

			key := mkey.ToString()
			for i, kid := range g.keepers {
				if kid == localID {
					cpay.Sign[i] = sign
				}
			}

			thisLinfo.currentPay = cpay

			// sync to other keepers？
			// get enough signs; then
			for _, kid := range g.keepers {
				if kid == localID {
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

	if thisLinfo.currentPay != nil && thisLinfo.currentPay.Status <= 0 {
		pAddr, _ := address.GetAddressFromID(proID) //providerAddress
		ukAddr, _ := address.GetAddressFromID(g.upkeeping.UpKeepingID)

		st := big.NewInt(thisLinfo.currentPay.Start)
		sl := big.NewInt(thisLinfo.currentPay.Length)
		sv := new(big.Int).SetBytes(thisLinfo.currentPay.Value)

		utils.MLogger.Infof("SpaceTimePay start pay for %s from %s, length %s value %s", proID, st.String(), sl.String(), sv.String())

		var root [32]byte
		copy(root[:], thisLinfo.currentPay.Root[:32])
		err := contracts.SpaceTimePay(ukAddr, pAddr, localSk, st, sl, sv, root, thisLinfo.currentPay.Share, thisLinfo.currentPay.Sign)
		if err != nil {
			utils.MLogger.Infof("SpaceTimePay start pay for user %s fsID %s pro %s from %s, length %s value %s failed %s", g.userID, g.groupID, proID, st.String(), sl.String(), sv.String(), err)
			return err
		}

		utils.MLogger.Infof("SpaceTimePay start pay for user %s fsID %s pro %s from %s, length %s value %s success", g.userID, g.groupID, proID, st.String(), sl.String(), sv.String())

		thisLinfo.currentPay.Status = 0
		thisLinfo.lastPay = thisLinfo.currentPay
		thisLinfo.currentPay = nil
		g.loadContracts(true)
		return nil
	}

	return role.ErrEmptyData
}

type timeValue struct {
	time  int64
	space int64
}

// challeng results to spacetime value
// lastTime is the lastest challenge time which is before Now
func (l *lInfo) resultSummary(price, start, end int64) (*big.Int, int64, []byte) {
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
		return spacetime, 0, nil
	}

	sort.Slice(tsl, func(i, j int) bool {
		return tsl[i].time < tsl[j].time
	})

	mtree := mt.New(sha256.New())
	mtree.SetIndex(0)

	timepre := tsl[0].time
	lengthpre := tsl[0].space
	var nbuf bytes.Buffer        // 替代网络连接
	enc := gob.NewEncoder(&nbuf) // 将写入网络。
	for _, tv := range tsl[1:] {
		spacetime.Add(spacetime, big.NewInt((tv.time-timepre)*int64(lengthpre+tv.space)/2))
		timepre = tv.time
		lengthpre = tv.space
		enc.Encode(tv)
		mtree.Push(nbuf.Bytes())
	}
	if spacetime.Sign() <= 0 {
		utils.MLogger.Info("error spacetime<=0")
	}

	spacetime.Mul(spacetime, big.NewInt(price))
	spacetime.Quo(spacetime, big.NewInt(1024*1024*60*60))
	if spacetime.Sign() <= 0 {
		utils.MLogger.Info("error!amount:", spacetime, "price:", price)
	}
	return spacetime, timepre, mtree.Root()
}
