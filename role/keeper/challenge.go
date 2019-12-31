package keeper

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/pb"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

func (k *Info) challengeRegular(ctx context.Context) {
	log.Println("Challenge service start!")
	ticker := time.NewTicker(CHALTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			log.Println("Challenge start at: ", utils.GetTimeNow())
			pus := k.getUQKeys()
			for _, pu := range pus {
				thisIgroup, ok := k.ukpGroup.Load(pu.qid)
				if !ok {
					continue
				}
				thisGroup := thisIgroup.(*groupInfo)

				for _, proID := range thisGroup.providers {
					thisIl, ok := thisGroup.ledgerMap.Load(proID)
					if !ok {
						continue
					}
					thisLinfo := thisIl.(*lInfo)
					key, value, err := thisLinfo.genChallengeBLS(k.netID, pu.qid, proID)
					if err != nil {
						continue
					}
					go k.ds.SendMetaRequest(ctx, int32(metainfo.Get), key, value, nil, proID)
				}

				// in case povider cannot get it
				go k.getUserBLS12Config(pu.uid, pu.qid)
			}
		}
	}
}

func (l *lInfo) genChallengeBLS(localID, qid, proID string) (string, []byte, error) {
	// last chanllenge has not complete
	if l.inChallenge {
		l.cleanLastChallenge()
	}

	l.inChallenge = true

	// at most challenge 100 blocks
	ret := make([]string, 0, 100)
	psum := 0
	chalnum := 0
	l.blockMap.Range(func(key, value interface{}) bool {
		cInfo := value.(*blockInfo)
		ret = append(ret, key.(string)+metainfo.BLOCK_DELIMITER+strconv.Itoa(cInfo.offset))
		psum += cInfo.offset + 1
		chalnum++
		if chalnum >= 100 {
			return false
		}
		return true
	})

	// no data
	if len(ret) == 0 || psum == 0 {
		l.inChallenge = false
		return "", nil, nil
	}

	challengetime := utils.GetUnixNow()
	// timestamp as random source
	// need more parameters to securely generate random source
	chal := mcl.GenChallenge(challengetime, ret)

	thischalresult := &chalresult{
		kid:        localID,
		pid:        proID,
		qid:        qid,
		chalTime:   challengetime,
		sum:        int64(psum) * df.DefaultSegmentSize,
		totalSpace: l.maxlength,
		h:          chal.C,
	}

	hProto := &pb.Chalnum{
		PubC:    int64(chal.C),
		Indices: chal.Indices,
	}
	hByte, err := proto.Marshal(hProto)
	if err != nil {
		l.inChallenge = false
		return "", nil, err
	}

	l.chalMap.Store(challengetime, thischalresult)
	l.chalCid = ret
	l.lastChalTime = challengetime

	// key: qid/"Challenge"/proID/chalTime
	km, err := metainfo.NewKeyMeta(qid, metainfo.Challenge, proID, utils.UnixToString(challengetime))
	if err != nil {
		log.Println("construct challenge KV error :", err)
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
		log.Println("thischalinfo.chalMap.Load error!challengetime: ", failChallTime)
		return
	}

	chalResult := thischalresult.(*chalresult)
	chalResult.res = false
	chalResult.length = 0

	l.inChallenge = false
}

//handleProof handles the challenge result from provider
//key: qid/"Challenge"/pid/chaltime,value: proof[/FaultBlocks]
func (k *Info) handleProof(km *metainfo.KeyMeta, proof []byte, pid string, blskey *mcl.PublicKey) bool {
	ops := km.GetOptions()
	if len(ops) < 1 {
		return false
	}

	chaltime := ops[0]

	thischalinfo, ok := l.getChalinfo(pu)
	if !ok {
		log.Println("getChalinfo error!pu: ", pu)
		return false
	}

	defer func() {
		thischalinfo.inChallenge = false
	}()

	spliteProof := strings.Split(string(proof), metainfo.DELIMITER)
	if len(spliteProof) < 3 {
		return false
	}

	var splitedindex []string
	if len(spliteProof) == 4 {
		indices, _ := b58.Decode(spliteProof[3])
		splitedindex = strings.Split(string(indices), metainfo.DELIMITER)
	}

	challengetime := utils.StringToUnix(chaltime)
	if thischalinfo.lastChalTime != challengetime {
		return false
	}

	thischalresult, ok := thischalinfo.chalMap.Load(challengetime)
	if !ok {
		log.Println("thischalinfo.chalMap.Load error!challengetime:", challengetime)
		return false
	}

	chalResult := thischalresult.(*chalresult)

	var chal mcl.Challenge
	var slength int64 //success length
	var electedOffset int
	var buf strings.Builder

	chal.C = chalResult.h

	// key: bucketid_stripeid_blockid_offset
	set := make(map[string]struct{}, len(splitedindex))
	// key: bucketid_stripeid_blockid
	cset := make(map[string]struct{}, len(splitedindex))
	if len(splitedindex) != 0 {
		log.Println("Fault or NotFound blocks :", pu.qid, metainfo.BLOCK_DELIMITER, splitedindex)
		for _, s := range splitedindex {
			if len(s) == 0 {
				continue
			}
			set[s] = struct{}{}
			chcid, _, err := utils.SplitIndex(s)
			if err != nil {
				log.Println("SplitIndex err:", err)
				continue
			}
			cset[chcid] = struct{}{}
		}
	}

	for _, index := range thischalinfo.chalCid {
		_, ok := set[index]
		if ok {
			continue
		}
		buf.Reset()
		chcid, off, err := utils.SplitIndex(index)
		if err != nil {
			log.Println("SplitIndex err:", err)
			continue
		}

		if off > 0 {
			electedOffset = chal.C % off
		} else if off == 0 {
			electedOffset = 0
		} else {
			continue
		}

		buf.WriteString(pu.qid)
		buf.WriteString(metainfo.BLOCK_DELIMITER)
		buf.WriteString(chcid)
		buf.WriteString(metainfo.BLOCK_DELIMITER)
		buf.WriteString(strconv.Itoa(electedOffset))

		chal.Indices = append(chal.Indices, buf.String())
		slength += int64(off + 1)
	}

	slength *= df.DefaultSegmentSize

	// recheck the status again
	if len(chal.Indices) == 0 {
		return false
	}

	blsProof := strings.Join(spliteProof[:3], metainfo.DELIMITER)

	res, err := mcl.VerifyProof(blskey, chal, blsProof)
	if err != nil {
		log.Println("handle proof of ", pu.qid, "from provider: ", pid, "verify err:", err)
		return false
	}
	if res {
		log.Println("handle proof of ", pu.qid, "from provider: ", pid, " verify success.")

		// update thischalinfo.cidMap;
		// except fault blocks, others are considered as "good"
		thischalinfo.cidMap.Range(func(k, v interface{}) bool {
			_, ok := cset[k.(string)]
			if ok {
				return true
			}
			cInfo := v.(*cidInfo)
			cInfo.repair = 0
			cInfo.availtime = challengetime
			return true
		})

		//update thischalinfo.chalMap
		chalResult.proof = blsProof
		chalResult.res = true
		chalResult.length = int64((float64(slength) / float64(chalResult.sum)) * float64(chalResult.totalSpace))
		return true
	} else {
		log.Println("handle proof of ", pu.qid, "from provider: ", pid, " verify fail.")
	}

	return false
}
