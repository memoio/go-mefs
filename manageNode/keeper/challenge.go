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
	"github.com/memoio/go-mefs/utils/bitset"
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

				utils.MLogger.Infof("Challenge for user %s fsID %s", pu.uid, pu.qid)

				chalTime := time.Now().Unix()
				mtime := thisGroup.upkeeping.StartTime
				if thisGroup.rootID != "" {
					gottime, _, err := role.GetLatestMerkleRoot(thisGroup.rootID)
					if err == nil {
						if chalTime < gottime {
							// maybe user can set large mtime
							utils.MLogger.Infof("latest merkle root time %d but chal time is %d", mtime, chalTime)
						} else {
							mtime = gottime
						}
					}
				}

				for _, proID := range thisGroup.providers {
					key, value, err := thisGroup.genChallengeBLS(k.localID, pu.uid, pu.qid, proID, mtime)
					if err != nil {
						utils.MLogger.Infof("Challenge for user %s fsID %s at provider %s fails: %s", pu.uid, pu.qid, proID, err)
						continue
					}
					utils.MLogger.Infof("Challenge for user %s fsID %s at provoder %s", pu.uid, pu.qid, proID)
					k.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, value, nil, proID)
				}

				// in case povider cannot get it
				go k.getUserBLS12Config(pu.uid, pu.qid)
			}
		}
	}
}

func (g *groupInfo) genChallengeBLS(localID, userID, qid, proID string, rootTime int64) (string, []byte, error) {
	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return "", nil, role.ErrNotMyProvider
	}

	// last chanllenge has not complete
	if thisLinfo.inChallenge {
		thisLinfo.cleanLastChallenge()
	}

	thisLinfo.inChallenge = true

	bucketNum := g.bucketNum
	stripeNum := (bucketNum + 1) * 3
	stripes := make([]int64, bucketNum)
	chunks := make([]int32, bucketNum)

	var res strings.Builder
	cset := make(map[string]int)
	bset := bitset.New(uint(stripeNum))
	psum := 0
	bitLen := uint(0)
	// challenge superbucket
	for i := 0; i <= int(bucketNum); i++ {
		// superbucket 3 chunks and 4k segment
		binfo := g.getBucketInfo(strconv.Itoa(-i), true)
		if binfo == nil {
			utils.MLogger.Infof("missing bucket %d info", -i)
			continue
		}
		for j := 0; j < binfo.chunkNum; j++ {
			res.Reset()
			res.WriteString(strconv.Itoa(-i))
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString("0")
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(j))
			cInfo, ok := thisLinfo.blockMap.Load(res.String())
			if ok {
				bset.Set(uint(i)*3 + uint(j))
				bitLen = uint(i)*3 + uint(j)
				psum += (cInfo.(*blockInfo).offset * int(binfo.bops.GetSegmentSize()))
				cset[res.String()] = cInfo.(*blockInfo).offset
				bitLen = uint(i)*3 + uint(j)
				break
			}
		}
	}

	// challenge buckets
	for i := 1; i <= int(bucketNum); i++ {
		binfo := g.getBucketInfo(strconv.Itoa(i), false)
		if binfo == nil {
			utils.MLogger.Infof("missing bucket %d info", i)
			continue
		}
		// not challenge last one
		count := binfo.curStripes
		chunks[i-1] = int32(binfo.chunkNum)
		bset.Set(uint(stripeNum) + uint(count*binfo.chunkNum))
		for j := 0; j < count; j++ {
			for k := 0; k < binfo.chunkNum; k++ {
				res.Reset()
				res.WriteString(strconv.Itoa(i))
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(strconv.Itoa(j))
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(strconv.Itoa(k))
				cInfo, ok := thisLinfo.blockMap.Load(res.String())
				if ok {
					bset.Set(uint(stripeNum) + uint(j*binfo.chunkNum) + uint(k))
					bitLen = uint(stripeNum) + uint(j*binfo.chunkNum) + uint(k)
					psum += (cInfo.(*blockInfo).offset * int(binfo.bops.GetSegmentSize()))
					cset[res.String()] = cInfo.(*blockInfo).offset
					break
				}
			}
		}

		bset.SetTo(uint(stripeNum)+uint(count*binfo.chunkNum), false)
		stripes[i-1] = int64(count)
		stripeNum += int64(count * binfo.chunkNum)
	}

	// no data
	if psum == 0 {
		thisLinfo.inChallenge = false
		return "", nil, role.ErrEmptyData
	}

	challengetime := time.Now().Unix()

	bset.Shrink(bitLen)

	chunkMap, err := bset.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	if thisLinfo.maxlength < int64(psum) {
		thisLinfo.maxlength = int64(psum)
	}

	thischalresult := &mpb.ChalInfo{
		Policy:      "1%", // use different policy; such as "1%"
		KeeperID:    localID,
		ProviderID:  proID,
		QueryID:     qid,
		UserID:      userID,
		ChalTime:    challengetime,
		RootTime:    rootTime,
		TotalLength: thisLinfo.maxlength,
		BucketNum:   bucketNum,
		StripeNum:   stripes,
		ChunkNum:    chunks,
		ChunkMap:    chunkMap,
	}

	hByte, err := proto.Marshal(thischalresult)
	if err != nil {
		thisLinfo.inChallenge = false
		return "", nil, err
	}

	thisLinfo.chalMap.Store(challengetime, thischalresult)
	thisLinfo.lastChalTime = challengetime
	thisLinfo.chalCid = cset

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
		utils.MLogger.Warnf("handleProof: %s fails: wrong keeperID", km.ToString())
		return
	}

	proInfo, ok := k.providers.Load(proID)
	if !ok {
		utils.MLogger.Warnf("handleProof: %s fails: no proInfoD", km.ToString())
		return
	}
	proInfo.(*pInfo).credit--

	thisGroup := k.getGroupInfo(userID, qid, false)
	if thisGroup == nil {
		utils.MLogger.Warnf("handleProof: %s fails: no groupinfo", km.ToString())
		return
	}

	thisLinfo := thisGroup.getLInfo(proID, false)
	if thisLinfo == nil {
		utils.MLogger.Warnf("handleProof: %s fails: no legerinfo", km.ToString())
		return
	}

	defer func() {
		thisLinfo.inChallenge = false
	}()

	challengetime := utils.StringToUnix(chaltime)
	if thisLinfo.lastChalTime != challengetime {
		utils.MLogger.Warnf("handleProof: %s fails: no challengetime", km.ToString())
		return
	}

	thischalresult, ok := thisLinfo.chalMap.Load(challengetime)
	if !ok {
		utils.MLogger.Warnf("handleProof: %s fails: no challenge result", km.ToString())
		return
	}

	chalResult := thischalresult.(*mpb.ChalInfo)

	spliteProof := strings.Split(string(value), metainfo.DELIMITER)
	if len(spliteProof) < 3 {
		utils.MLogger.Warnf("handleProof: %s fails: proof is too short", km.ToString())
		return
	}

	fset := bitset.New(0)
	bset := bitset.New(0)
	flength := uint(0)
	if len(spliteProof) == 4 {
		fmap, err := b58.Decode(spliteProof[3])
		if err == nil {
			fset.UnmarshalBinary(fmap)
			flength = fset.Len()
			chalResult.FailMap = fmap
		}
	}

	err := bset.UnmarshalBinary(chalResult.ChunkMap)
	if err != nil {
		utils.MLogger.Warnf("handleProof: %s fails: %s", km.ToString(), err)
		return
	}

	var chal mcl.Challenge
	chal.Seed = mcl.GenChallenge(chalResult)

	totalNum := bset.Count()
	startPos := uint(chal.Seed) % bset.Len()
	if startPos < uint(chalResult.GetBucketNum()+1)*3 {
		startPos = uint(chalResult.GetBucketNum()+1) * 3
	}

	var slength, chalLength int64 //success length
	var electedOffset int
	var buf strings.Builder

	bucketID := 0
	stripeID := 1
	chunkID := 0
	failset := make(map[string]struct{})
	for i, e := bset.NextSet(0); e && i < uint(chalResult.BucketNum+1)*3; i, e = bset.NextSet(i + 1) {
		buf.Reset()
		bucketID = -int(i / 3)
		stripeID = 0
		chunkID = int(i % 3)
		buf.WriteString(strconv.Itoa(bucketID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum, ok := thisLinfo.chalCid[blockID]
		if !ok {
			continue
		}
		chalLength += int64(segNum * 4096)

		if flength != 0 && !fset.Test(i) {
			failset[blockID] = struct{}{}
			continue
		}

		slength += int64(segNum * 4096)
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		buf.Reset()
		buf.WriteString(qid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(blockID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))
		chal.Indices = append(chal.Indices, buf.String())
	}

	count := uint(0)
	bucketID = 1
	chunkNum := chalResult.GetChunkNum()[0]
	stripeNum := 3 * (chalResult.GetBucketNum() + 1)
	for i, e := bset.NextSet(startPos); e; i, e = bset.NextSet(i + 1) {
		count++
		for j := bucketID; j < int(chalResult.GetBucketNum()); j++ {
			if stripeNum+chalResult.GetStripeNum()[j-1]*int64(chalResult.GetChunkNum()[j-1]) < int64(i) {
				break
			}
			bucketID = j
			chunkNum = chalResult.GetChunkNum()[j-1]
			stripeNum += chalResult.GetStripeNum()[j-1] * int64(chalResult.GetChunkNum()[j-1])
		}

		stripeID = int((int64(i) - stripeNum) / int64(chunkNum))
		chunkID = int((int64(i) - stripeNum) % int64(chunkNum))

		buf.Reset()
		buf.WriteString(strconv.Itoa(bucketID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum, ok := thisLinfo.chalCid[blockID]
		if !ok {
			continue
		}

		bi := thisGroup.getBucketInfo(strconv.Itoa(bucketID), false)
		if bi == nil {
			continue
		}

		chalLength += int64(segNum * int(bi.bops.GetSegmentSize()))

		if flength != 0 && !fset.Test(i) {
			failset[blockID] = struct{}{}
			continue
		}

		slength += int64(segNum * int(bi.bops.GetSegmentSize()))
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		buf.Reset()
		buf.WriteString(qid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(blockID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))
		chal.Indices = append(chal.Indices, buf.String())

		if count > totalNum/100 {
			break
		}
	}

	for i, e := bset.NextSet(uint(chalResult.BucketNum+1) * 3); e && i < startPos; i, e = bset.NextSet(i + 1) {
		if count > totalNum/100 {
			break
		}
		count++
		for j := bucketID; j < int(chalResult.GetBucketNum()); j++ {
			if stripeNum+chalResult.GetStripeNum()[j-1]*int64(chalResult.GetChunkNum()[j-1]) < int64(i) {
				break
			}
			bucketID = j
			chunkNum = chalResult.GetChunkNum()[j-1]
			stripeNum += chalResult.GetStripeNum()[j-1] * int64(chalResult.GetChunkNum()[j-1])
		}

		stripeID = int((int64(i) - stripeNum) / int64(chunkNum))
		chunkID = int((int64(i) - stripeNum) % int64(chunkNum))

		buf.Reset()
		buf.WriteString(strconv.Itoa(bucketID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum, ok := thisLinfo.chalCid[blockID]
		if !ok {
			continue
		}

		bi := thisGroup.getBucketInfo(strconv.Itoa(bucketID), false)
		if bi == nil {
			continue
		}

		chalLength += int64(segNum * int(bi.bops.GetSegmentSize()))

		if flength != 0 && !fset.Test(i) {
			failset[blockID] = struct{}{}
			continue
		}

		slength += int64(segNum * int(bi.bops.GetSegmentSize()))
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		buf.Reset()
		buf.WriteString(qid)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(blockID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))
		chal.Indices = append(chal.Indices, buf.String())

		if count > totalNum/100 {
			break
		}
	}

	// recheck the status again
	if len(chal.Indices) == 0 {
		utils.MLogger.Warnf("handleProof: %s fails: chal empty", km.ToString())
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

	blsKey, err := k.getUserBLS12Config(userID, qid)
	if err != nil {
		return
	}

	res, err := blsKey.VerifyProof(chal, pf, true)
	if err != nil {
		utils.MLogger.Error("proof of ", qid, " from provider: ", proID, " verify fails: ", err)
		utils.MLogger.Warn("verify blocks: ", chal.Indices)
		return
	}

	if res {
		utils.MLogger.Info("proof of ", qid, " from provider: ", proID, " verify success.")

		// update thischalinfo.cidMap;
		// except fault blocks, others are considered as "good"

		thisLinfo.blockMap.Range(func(k, v interface{}) bool {
			_, ok := failset[k.(string)]
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
		chalResult.ChalLength = chalLength
		chalResult.SuccessLength = int64((float64(slength) / float64(chalLength)) * float64(chalResult.TotalLength))
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
