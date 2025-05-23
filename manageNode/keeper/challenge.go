package keeper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
	b58 "github.com/mr-tron/base58/base58"
)

func (k *Info) challengeRegular(ctx context.Context) {
	utils.MLogger.Info("Challenge service start!")
	ticker := time.NewTicker(chalTime)
	cdata := int64(0)
	count := 0
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
				if thisGroup == nil || thisGroup.upkeeping == nil || !thisGroup.status {
					continue
				}

				chalTime := time.Now().Unix()

				if thisGroup.upkeeping.EndTime < chalTime {
					utils.MLogger.Infof("Challenge for user %s fsID %s upkeeping has expired", pu.uid, pu.qid)
					continue
				}

				utils.MLogger.Infof("Challenge for user %s fsID %s", pu.uid, pu.qid)

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

				utils.MLogger.Infof("Challenge for user %s fsID %s at rootTime %d", pu.uid, pu.qid, mtime)
				count = 0
				for _, proID := range thisGroup.providers {
					if pu.uid == pos.GetPostId() {
						key, value, err := thisGroup.genChallengeRandom100(k.localID, pu.uid, pu.qid, proID, mtime)
						if err != nil {
							utils.MLogger.Infof("Challenge data for post user %s fsID %s at provider %s fails: %s", pu.uid, pu.qid, proID, err)
							continue
						}
						count++
						utils.MLogger.Infof("Challenge post data: %s", key)
						k.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, value, nil, proID)
						continue
					}
					if cdata%2 == 0 {
						key, value, err := thisGroup.genChallengeData(k.localID, pu.uid, pu.qid, proID, mtime)
						if err != nil {
							utils.MLogger.Infof("Challenge data for user %s fsID %s at provider %s fails: %s", pu.uid, pu.qid, proID, err)
							continue
						}
						count++
						utils.MLogger.Infof("Challenge data: %s", key)
						k.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, value, nil, proID)
					} else {
						key, value, err := thisGroup.genChallengeMeta(k.localID, pu.uid, pu.qid, proID, mtime)
						if err != nil {
							utils.MLogger.Infof("Challenge meta for user %s fsID %s at provider %s fails: %s", pu.uid, pu.qid, proID, err)
							continue
						}
						count++
						utils.MLogger.Infof("Challenge meta: %s", key)
						k.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, value, nil, proID)
					}

				}

				if count > 0 {
					// in case povider cannot get it
					go k.getUserBLS12Config(pu.uid, pu.qid)
				}
			}
			cdata++
		}
	}
}

func (g *groupInfo) genChallengeData(localID, userID, qid, proID string, rootTime int64) (string, []byte, error) {
	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return "", nil, role.ErrNotMyProvider
	}

	// last chanllenge has not complete
	if thisLinfo.inChallenge {
		thisLinfo.cleanLastChallenge()
	}

	thisLinfo.inChallenge = true

	bucketNum := int(g.bucketNum + 1)
	bc := make([]*mpb.BucketContent, bucketNum)
	for i := 0; i < bucketNum; i++ {
		bi := &mpb.BucketContent{
			ChunkNum:  0,
			StripeNum: 0,
			SegCount:  0,
			SegSize:   0,
		}

		bc[i] = bi
	}

	challengetime := time.Now().Unix()

	var res strings.Builder
	cset := make(map[string]int)
	bset := bitset.New(0)
	psum := 0
	cNum := int64(0)

	// challenge buckets
	for i := 1; i < bucketNum; i++ {
		binfo := g.getBucketInfo(strconv.Itoa(i), false)
		if binfo == nil {
			utils.MLogger.Infof("missing bucket %d info", i)
			continue
		}

		count := binfo.curStripes
		if count < 0 {
			continue
		}

		bc[i].ChunkNum = int32(binfo.chunkNum)
		bc[i].SegSize = binfo.bops.GetSegmentSize()
		bc[i].SegCount = binfo.bops.GetSegmentCount()

		for k := 0; k < binfo.chunkNum; k++ {
			res.Reset()
			res.WriteString(strconv.Itoa(i))
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(count))
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(k))
			cInfo, ok := thisLinfo.blockMap.Load(res.String())
			if ok {
				if cInfo.(*blockInfo).offset != int(binfo.bops.GetSegmentCount()) {
					if count == 0 {
						bc[i].SegCount = int32(cInfo.(*blockInfo).offset)
						count++
					}
				} else {
					// seg is full
					count++
				}
				break
			}
		}

		bc[i].StripeNum = int64(count)

		bset.SetTo(uint(count*binfo.chunkNum), false)
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
					bset.Set(uint(cNum) + uint(j*binfo.chunkNum) + uint(k))
					psum += (cInfo.(*blockInfo).offset * int(binfo.bops.GetSegmentSize()))
					cset[res.String()] = cInfo.(*blockInfo).offset
					break
				}
			}
		}

		cNum += int64(count * binfo.chunkNum)
	}

	// no data
	if psum == 0 {
		thisLinfo.inChallenge = false
		return "", nil, role.ErrEmptyData
	}

	chunkMap, err := bset.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	if thisLinfo.maxlength < int64(psum) {
		thisLinfo.maxlength = int64(psum)
	}

	thischalresult := &mpb.ChalInfo{
		Policy:      "smart", // use different policy; such as "1%"
		KeeperID:    localID,
		ProviderID:  proID,
		QueryID:     qid,
		UserID:      userID,
		ChalTime:    challengetime,
		RootTime:    rootTime,
		TotalLength: thisLinfo.maxlength,
		Buckets:     bc,
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

func (g *groupInfo) genChallengeMeta(localID, userID, qid, proID string, rootTime int64) (string, []byte, error) {
	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return "", nil, role.ErrNotMyProvider
	}

	// last chanllenge has not complete
	if thisLinfo.inChallenge {
		thisLinfo.cleanLastChallenge()
	}

	thisLinfo.inChallenge = true

	bucketNum := int(g.bucketNum + 1)
	bc := make([]*mpb.BucketContent, bucketNum)
	for i := 0; i < bucketNum; i++ {
		bi := &mpb.BucketContent{
			ChunkNum:  0,
			StripeNum: 0,
			SegCount:  0,
			SegSize:   0,
		}

		bc[i] = bi
	}

	var res strings.Builder
	cset := make(map[string]int)
	bset := bitset.New(0)
	psum := 0
	stripeNum := int64(0)

	// challenge buckets
	for i := 0; i < bucketNum; i++ {
		binfo := g.getBucketInfo(strconv.Itoa(-i), false)
		if binfo == nil {
			utils.MLogger.Infof("missing bucket %d info", -i)
			continue
		}

		count := binfo.curStripes + 1
		if count <= 0 {
			continue
		}

		bi := &mpb.BucketContent{
			ChunkNum:  int32(binfo.chunkNum),
			StripeNum: int64(count),
			SegCount:  1, // only challenge first seg?
			SegSize:   binfo.bops.GetSegmentSize(),
		}

		bc[i] = bi

		bset.Set(uint(stripeNum) + uint(count*binfo.chunkNum))
		for j := 0; j < count; j++ {
			for k := 0; k < binfo.chunkNum; k++ {
				res.Reset()
				res.WriteString(strconv.Itoa(-i))
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(strconv.Itoa(j))
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(strconv.Itoa(k))
				cInfo, ok := thisLinfo.blockMap.Load(res.String())
				if ok {
					bset.Set(uint(stripeNum) + uint(j*binfo.chunkNum) + uint(k))
					psum += (cInfo.(*blockInfo).offset * int(binfo.bops.GetSegmentSize()))
					cset[res.String()] = cInfo.(*blockInfo).offset
					break
				}
			}
		}

		bset.SetTo(uint(stripeNum)+uint(count*binfo.chunkNum), false)
		stripeNum += int64(count * binfo.chunkNum)
	}

	// no data
	if psum == 0 {
		thisLinfo.inChallenge = false
		return "", nil, role.ErrEmptyData
	}

	challengetime := time.Now().Unix()

	chunkMap, err := bset.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	if thisLinfo.maxlength < int64(psum) {
		thisLinfo.maxlength = int64(psum)
	}

	thischalresult := &mpb.ChalInfo{
		Policy:      "meta", // use different policy; such as "1%"
		KeeperID:    localID,
		ProviderID:  proID,
		QueryID:     qid,
		UserID:      userID,
		ChalTime:    challengetime,
		RootTime:    rootTime,
		TotalLength: thisLinfo.maxlength,
		Buckets:     bc,
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

func (g *groupInfo) genChallengeRandom100(localID, userID, qid, proID string, rootTime int64) (string, []byte, error) {
	thisLinfo := g.getLInfo(proID, false)
	if thisLinfo == nil {
		return "", nil, role.ErrNotMyProvider
	}

	// last chanllenge has not complete
	if thisLinfo.inChallenge {
		thisLinfo.cleanLastChallenge()
	}

	thisLinfo.inChallenge = true

	challengetime := time.Now().Unix()

	// at most challenge 100 blocks
	cset := make(map[string]int)
	ret := make([]string, 0, 100)
	chalnum := 0
	psum := 0
	thisLinfo.blockMap.Range(func(key, value interface{}) bool {
		cInfo := value.(*blockInfo)
		cset[key.(string)] = cInfo.offset
		ret = append(ret, key.(string)+metainfo.BlockDelimiter+strconv.Itoa(cInfo.offset))
		psum += cInfo.offset
		chalnum++
		if chalnum >= 100 {
			return false
		}
		return true
	})

	// no data
	if psum == 0 {
		thisLinfo.inChallenge = false
		return "", nil, role.ErrEmptyData
	}

	psum *= (pos.DLen)

	if thisLinfo.maxlength < int64(psum) {
		thisLinfo.maxlength = int64(psum)
	}

	thischalresult := &mpb.ChalInfo{
		Policy:      "random100", // use different policy; such as "1%"
		KeeperID:    localID,
		ProviderID:  proID,
		QueryID:     qid,
		UserID:      userID,
		ChalTime:    challengetime,
		RootTime:    rootTime,
		TotalLength: thisLinfo.maxlength,
		ChalLength:  int64(psum),
		Blocks:      ret,
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
	chalResult.FailMap = chalResult.ChunkMap
	_, sucCids, faultCids, err := role.VerifyChallenge(chalResult, nil, false)
	if err != nil {
		faultCids = append(faultCids, sucCids...)
		sucCids = nil
	}

	if len(sucCids) > 0 {
		utils.MLogger.Debugf("proof of %s has sucCids: %s", chalResult.GetQueryID(), sucCids)
		for _, key := range sucCids {
			_, ok := l.faultCid.Load(key)
			if ok {
				l.faultCid.Delete(key)
			}
		}
	}

	if len(faultCids) > 0 {
		utils.MLogger.Debugf("proof of %s has faultCids: %s", chalResult.GetQueryID(), faultCids)
		for _, key := range faultCids {
			l.faultCid.Store(key, struct{}{})
		}
	}

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

	qid := km.GetMainID()
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
	if proInfo.(*pInfo).credit < -100 {
		proInfo.(*pInfo).credit = -100
	}

	thisGroup := k.getGroupInfo(userID, qid, false)
	if thisGroup == nil {
		utils.MLogger.Warnf("handleProof: %s fails: no groupinfo", km.ToString())
		return
	}

	if !thisGroup.status {
		utils.MLogger.Warnf("handleProof: %s fails: groupinfo is not ready", km.ToString())
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
	if len(spliteProof) < 1 {
		utils.MLogger.Warnf("handleProof: %s fails: proof is too short", km.ToString())
		return
	}

	var err error
	chalResult.BlsProof, err = b58.Decode(spliteProof[0])
	if err != nil {
		utils.MLogger.Warnf("handleProof: %s fails: proof b58 decode failed", km.ToString())
		return
	}
	switch chalResult.GetPolicy() {
	case "smart", "meta":
		if len(spliteProof) == 2 {
			fmap, err := b58.Decode(spliteProof[1])
			if err == nil {
				chalResult.FailMap = fmap
			}
		}
	case "random100":
		if len(spliteProof) == 2 {
			indices, err := b58.Decode(spliteProof[1])
			if err == nil {
				chalResult.FaultBlocks = strings.Split(string(indices), metainfo.DELIMITER)
			}
		}
	default:
	}

	blsKey, err := k.getUserBLS12Config(userID, qid)
	if err != nil {
		return
	}

	res, sucCids, faultCids, err := role.VerifyChallenge(chalResult, blsKey, false)
	if err != nil {
		utils.MLogger.Error("proof of ", qid, " from provider: ", proID, " verify fails: ", err)
		faultCids = append(faultCids, sucCids...)
		sucCids = nil
	}

	if len(sucCids) > 0 {
		utils.MLogger.Debugf("proof of %s has sucCids: %s", km.ToString(), sucCids)
		for _, key := range sucCids {
			_, ok := thisLinfo.faultCid.Load(key)
			if ok {
				thisLinfo.faultCid.Delete(key)
			}
		}
	}

	if len(faultCids) > 0 {
		utils.MLogger.Debugf("proof of %s has faultCids: %s", km.ToString(), faultCids)
		for _, key := range faultCids {
			thisLinfo.faultCid.Store(key, struct{}{})
		}
	}

	if res {
		utils.MLogger.Info("proof of ", qid, " from provider: ", proID, " verify success.")

		// update thischalinfo.cidMap;
		// except fault blocks, others are considered as "good"
		thisLinfo.blockMap.Range(func(k, v interface{}) bool {
			_, ok = thisLinfo.faultCid.Load(k.(string))
			if ok {
				utils.MLogger.Debugf("do not change faulted %s availtime for %s", k.(string), qid)
				return true
			}

			v.(*blockInfo).repair = 0
			v.(*blockInfo).availtime = challengetime
			return true
		})

		proInfo.(*pInfo).credit += 2
		if proInfo.(*pInfo).credit > 100 {
			proInfo.(*pInfo).credit = 100
		}
	} else {
		utils.MLogger.Info("handle proof of ", qid, "from provider: ", proID, " verify fail.")
		utils.MLogger.Info("User's verifykey is ", blsKey.Serialize(), "proof is ", chalResult.BlsProof)
	}

	//update thischalinfo.chalMap
	hByte, err := proto.Marshal(chalResult)
	if err != nil {
		return
	}

	k.putKey(k.context, km.ToString(), hByte, nil, "local", thisGroup.clusterID, thisGroup.bft)

}
