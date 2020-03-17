package keeper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

func (k *Info) challengeRegular(ctx context.Context) {
	utils.MLogger.Info("Challenge service start!")
	ticker := time.NewTicker(chalTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.MLogger.Info("Regular challenge start")
			pus := k.getQUKeys()
			for _, pu := range pus {
				thisGroup := k.getGroupInfo(pu.uid, pu.qid, false)
				if thisGroup == nil {
					continue
				}

				for _, proID := range thisGroup.providers {
					key, value, err := thisGroup.genChallengeBLS(k.localID, pu.uid, pu.qid, proID)
					if err != nil {
						continue
					}
					utils.MLogger.Debug("Challenge: ", key)
					go k.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, value, nil, proID)
				}

				// in case povider cannot get it
				go k.getUserBLS12Config(pu.uid, pu.qid)
			}
		}
	}
}

func (g *groupInfo) genChallengeBLS(localID, userID, qid, proID string) (string, []byte, error) {
	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return "", nil, role.ErrEmptyData
	}

	// last chanllenge has not complete
	if thisLinfo.inChallenge {
		thisLinfo.cleanLastChallenge()
	}

	thisLinfo.inChallenge = true

	// at most challenge 100 blocks
	ret := make([]string, 100)
	chalnum := 0
	thisLinfo.blockMap.Range(func(key, value interface{}) bool {
		if chalnum >= 100 {
			return false
		}
		ret[chalnum] = key.(string)
		chalnum++
		return true
	})

	psum := 0
	for i := 0; i < len(ret); i++ {
		cInfo, ok := thisLinfo.blockMap.Load(ret[i])
		if !ok {
			continue
		}

		bids := strings.Split(ret[i], metainfo.BlockDelimiter)
		bi := g.getBucketInfo(bids[0], false)
		if bi == nil {
			continue
		}

		bSize := int(bi.bops.GetSegmentSize())
		ret[i] = ret[i] + metainfo.BlockDelimiter + strconv.Itoa(cInfo.(*blockInfo).offset)
		psum += (cInfo.(*blockInfo).offset * bSize)
	}

	// no data
	if len(ret) == 0 || psum == 0 {
		thisLinfo.inChallenge = false
		return "", nil, role.ErrEmptyData
	}

	challengetime := time.Now().Unix()

	thischalresult := &mpb.ChalInfo{
		KeeperID:    localID,
		ProviderID:  proID,
		QueryID:     qid,
		UserID:      userID,
		ChalTime:    challengetime,
		ChalLength:  int64(psum),
		Blocks:      ret,
		TotalLength: thisLinfo.maxlength,
	}

	hByte, err := proto.Marshal(thischalresult)
	if err != nil {
		thisLinfo.inChallenge = false
		return "", nil, err
	}

	thisLinfo.chalMap.Store(challengetime, thischalresult)
	thisLinfo.chalCid = ret
	thisLinfo.lastChalTime = challengetime

	// key: qid/"Challenge"/uid/pid/kid/chaltime
	km, err := metainfo.NewKey(qid, mpb.KeyType_Challenge, userID, proID, localID, utils.UnixToString(challengetime))
	if err != nil {
		return "", nil, err
	}
	return km.ToString(), hByte, nil
}

func (l *lInfo) cleanLastChallenge() {
	if !l.inChallenge {
		return
	}

	failChallTime := l.lastChalTime
	thischalresult, ok := l.chalMap.Load(failChallTime)
	if !ok {
		return
	}

	chalResult := thischalresult.(*mpb.ChalInfo)
	chalResult.Res = false

	l.inChallenge = false
}

//handleProof handles the challenge result from provider
//key: qid/"Challenge"/uid/pid/kid/chaltime,value: proof[/FaultBlocks]
func (k *Info) handleProof(km *metainfo.Key, value []byte) {
	utils.MLogger.Info("handleProof: ", km.ToString())
	ops := km.GetOptions()
	if len(ops) != 4 {
		return
	}

	qid := km.GetMid()
	userID := ops[0]
	proID := ops[1]
	kid := ops[2]
	chaltime := ops[3]
	if kid != k.localID {
		return
	}

	proInfo, ok := k.providers.Load(proID)
	if !ok {
		return
	}
	proInfo.(*pInfo).credit--

	thisGroup := k.getGroupInfo(userID, qid, false)
	if thisGroup == nil {
		return
	}

	thisLinfo := thisGroup.getLInfo(proID, false)
	if thisLinfo == nil {
		return
	}

	defer func() {
		thisLinfo.inChallenge = false
	}()

	challengetime := utils.StringToUnix(chaltime)
	if thisLinfo.lastChalTime != challengetime {
		return
	}

	thischalresult, ok := thisLinfo.chalMap.Load(challengetime)
	if !ok {
		return
	}

	chalResult := thischalresult.(*mpb.ChalInfo)

	spliteProof := strings.Split(string(value), metainfo.DELIMITER)
	if len(spliteProof) < 3 {
		return
	}

	var splitedindex []string
	if len(spliteProof) == 4 {
		indices, _ := b58.Decode(spliteProof[3])
		splitedindex = strings.Split(string(indices), metainfo.DELIMITER)
	}

	var chal mcl.Challenge
	var slength int64 //success length
	var electedOffset int
	var buf strings.Builder

	chal.Seed = mcl.GenChallenge(chalResult)

	// key: bucketid_stripeid_blockid_offset
	set := make(map[string]struct{}, len(splitedindex))
	// key: bucketid_stripeid_blockid
	cset := make(map[string]struct{}, len(splitedindex))
	if len(splitedindex) != 0 {
		utils.MLogger.Debug(proID, " Fault or NotFound blocks :", qid, metainfo.BlockDelimiter, splitedindex)
		for _, s := range splitedindex {
			if len(s) == 0 {
				continue
			}
			set[s] = struct{}{}
			chcid, _, err := utils.SplitIndex(s)
			if err != nil {
				continue
			}
			cset[chcid] = struct{}{}
		}
	}

	for _, index := range thisLinfo.chalCid {
		_, ok := set[index]
		if ok {
			continue
		}
		buf.Reset()
		bids := strings.Split(index, metainfo.BlockDelimiter)
		if len(bids) != 4 {
			continue
		}

		bi := thisGroup.getBucketInfo(bids[0], false)
		if bi == nil {
			continue
		}

		off, err := strconv.Atoi(bids[3])
		if err != nil {
			continue
		}

		if off > 0 {
			electedOffset = chal.Seed % off
		} else if off == 0 {
			electedOffset = 0
		} else {
			continue
		}

		buf.WriteString(qid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(bids[0])
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(bids[1])
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(bids[2])
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))

		chal.Indices = append(chal.Indices, buf.String())

		slength += int64(off * int(bi.bops.GetSegmentSize()))
	}

	// recheck the status again
	if len(chal.Indices) == 0 {
		return
	}

	muByte, err := b58.Decode(spliteProof[0])
	if err != nil {
		return
	}
	nuByte, err := b58.Decode(spliteProof[1])
	if err != nil {
		return
	}
	deltaByte, err := b58.Decode(spliteProof[2])
	if err != nil {
		return
	}
	pf := &mcl.Proof{
		Mu:    muByte,
		Nu:    nuByte,
		Delta: deltaByte,
	}

	res, err := thisGroup.blsKey.VerifyProof(chal, pf)
	if err != nil {
		utils.MLogger.Error("proof of ", qid, " from provider: ", proID, "verify fails: ", err)
		utils.MLogger.Warn("verify blocks: ", chal.Indices)
		return
	}

	if res {
		utils.MLogger.Info("proof of ", qid, " from provider: ", proID, " verify success.")

		// update thischalinfo.cidMap;
		// except fault blocks, others are considered as "good"
		thisLinfo.blockMap.Range(func(k, v interface{}) bool {
			_, ok := cset[k.(string)]
			if ok {
				utils.MLogger.Debugf("do not change %s availtime for %s", k.(string), qid)
				return true
			}
			cInfo := v.(*blockInfo)
			cInfo.repair = 0
			cInfo.availtime = challengetime
			return true
		})

		chalResult.Res = true
		chalResult.SuccessLength = int64((float64(slength) / float64(chalResult.ChalLength)) * float64(chalResult.TotalLength))
		proInfo.(*pInfo).credit += 2
		if proInfo.(*pInfo).credit > 100 {
			proInfo.(*pInfo).credit = 100
		}
	} else {
		chalResult.Res = false
		chalResult.SuccessLength = 0
		utils.MLogger.Info("handle proof of ", qid, "from provider: ", proID, " verify fail.")
	}

	//update thischalinfo.chalMap
	chalResult.BlsProof = strings.Join(spliteProof[:3], metainfo.DELIMITER)

	hByte, err := proto.Marshal(chalResult)
	if err != nil {
		return
	}

	k.putKey(k.context, km.ToString(), hByte, nil, "local", thisGroup.clusterID, thisGroup.bft)

	return
}
