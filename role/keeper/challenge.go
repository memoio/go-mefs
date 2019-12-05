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

func challengeRegular(ctx context.Context) {
	log.Println("Challenge service start!")
	ticker := time.NewTicker(CHALTIME)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			challengeProviderBLS12()
		}
	}
}

func challengeProviderBLS12() {
	log.Println("Challenge start at: ", utils.GetTimeNow())

	pus := getPUKeysFromukpInfo()
	for _, pu := range pus {
		thisInfo, ok := ledgerInfo.Load(pu)
		if !ok {
			continue
		}

		thischalinfo := thisInfo.(*chalinfo)

		// last chanllenge has not complete
		if thischalinfo.inChallenge {
			cleanLastChallenge(pu)
		}

		thischalinfo.inChallenge = true

		// at most challenge 100 blocks
		ret := make([]string, 0, 100)
		psum := 0
		chalnum := 0
		thischalinfo.cidMap.Range(func(key, value interface{}) bool {
			cInfo := value.(*cidInfo)
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
			thischalinfo.inChallenge = false
			continue
		}

		challengetime := utils.GetUnixNow()
		// timestamp as random source
		// need more parameters to securely generate random source
		chal := mcl.GenChallenge(challengetime, ret)

		thischalresult := &chalresult{
			kid:           localNode.Identity.Pretty(),
			pid:           pu.pid,
			uid:           pu.uid,
			challengeTime: challengetime,
			sum:           int64(psum) * df.DefaultSegmentSize,
			totalSpace:    thischalinfo.maxlength,
			h:             chal.C,
		}

		hProto := &pb.Chalnum{
			PubC:    int64(chal.C),
			Indices: chal.Indices,
		}
		hByte, err := proto.Marshal(hProto)
		if err != nil {
			log.Println("marshal h failed, err: ", err)
			thischalinfo.inChallenge = false
			continue
		}

		thischalinfo.chalMap.Store(challengetime, thischalresult)
		thischalinfo.chalCid = ret
		thischalinfo.lastChalTime = challengetime

		go doChallengeBLS12(pu, hByte, challengetime)
	}
}

func doChallengeBLS12(pu puKey, hByte []byte, chaltime int64) {
	fail := false
	// clean before return
	defer func() {
		if fail {
			cleanLastChallenge(pu)
		}
	}()

	// get user config once; in case provider cannot get it
	getUserBLS12Config(pu.uid)

	km, err := metainfo.NewKeyMeta(pu.uid, metainfo.Challenge, utils.UnixToString(chaltime))
	if err != nil {
		log.Println("construct challenge KV error :", err)
		fail = true
		return
	}
	metaValue := b58.Encode(hByte)
	_, err = sendMetaRequest(km, metaValue, pu.pid)
	if err != nil {
		log.Println("DoChallengeBLS12 error :", err)
		fail = true
		return
	}
	return
}

func cleanLastChallenge(pu puKey) {
	thischalinfo, ok := getChalinfo(pu)
	if !ok {
		log.Println("getChalinfo error!pu: ", pu)
		return
	}

	if !thischalinfo.inChallenge {
		return
	}

	failChallTime := thischalinfo.lastChalTime
	thischalresult, ok := thischalinfo.chalMap.Load(failChallTime)
	if !ok {
		log.Println("thischalinfo.chalMap.Load error!challengetime: ", failChallTime)
		return
	}

	chalResult := thischalresult.(*chalresult)
	chalResult.res = false
	chalResult.length = 0

	thischalinfo.inChallenge = false
}

//handleProofResultBls12 handles the challenge result from provider
//key: uid/"proof"/chaltime,value: proof[/FaultBlocks]
func handleProofResultBls12(km *metainfo.KeyMeta, proof, pid string) {
	ops := km.GetOptions()
	if len(ops) < 1 {
		return
	}

	chaltime := ops[0]
	uid := km.GetMid()

	log.Println("handle proof of ", uid, "from provider: ", pid)

	pu := puKey{
		pid: pid,
		uid: uid,
	}

	defer func() {
		cleanLastChallenge(pu)
	}()

	thischalinfo, ok := getChalinfo(pu)
	if !ok {
		log.Println("getChalinfo error!pu: ", pu)
		return
	}

	spliteProof := strings.Split(proof, metainfo.DELIMITER)
	if len(spliteProof) < 3 {
		return
	}

	var splitedindex []string
	if len(spliteProof) == 4 {
		indices, _ := b58.Decode(spliteProof[3])
		splitedindex = strings.Split(string(indices), metainfo.DELIMITER)
	}

	challengetime := utils.StringToUnix(chaltime)
	if thischalinfo.lastChalTime != challengetime {
		return
	}

	thischalresult, ok := thischalinfo.chalMap.Load(challengetime)
	if !ok {
		log.Println("thischalinfo.chalMap.Load error!challengetime:", challengetime)
		return
	}

	chalResult := thischalresult.(*chalresult)

	var chal mcl.Challenge
	var slength int64 //success length
	var electedOffset int
	var buf strings.Builder

	chal.C = chalResult.h

	set := make(map[string]struct{}, len(splitedindex))
	if len(splitedindex) != 0 {
		log.Println("Fault or NotFound blocks :", splitedindex)
		reduceCredit(pid)
		for _, s := range splitedindex {
			if len(s) == 0 {
				continue
			}
			set[s] = struct{}{}
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

		buf.WriteString(uid)
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
		return
	}

	pubKey, err := getUserBLS12Config(uid)
	if err != nil {
		log.Println("getUserBLS12Config error! uid:", uid)
		return
	}

	blsProof := strings.Join(spliteProof[:3], metainfo.DELIMITER)

	res, err := mcl.VerifyProof(pubKey, chal, blsProof)
	if err != nil {
		log.Println("mcl.VerifyProof err: ", err)
		return
	}
	if res {
		log.Println("verify success: ", uid)

		//update thischalinfo.cidMap;
		// except fault blocks, others are considered as "good"
		thischalinfo.cidMap.Range(func(k, v interface{}) bool {
			_, ok := set[k.(string)]
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

		// TODO: store in disk
		addCredit(pid)
	} else {
		log.Println("verify failed cid: ", chal.Indices)
		reduceCredit(pid)
	}

	thischalinfo.inChallenge = false

	return
}
