package provider

import (
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

// key: qid/"Challenge"/uid/pid/kid/chaltime
func (p *Info) handleChallengeBls12(km *metainfo.Key, metaValue []byte, from string) error {
	utils.MLogger.Info("handle challenge: ", km.ToString(), " from: ", from)

	var data, tag [][]byte
	var buf, cbuf strings.Builder
	failchunk := false

	ops := km.GetOptions()
	if len(ops) < 4 {
		return role.ErrWrongKey
	}

	fsID := km.GetMid()
	userID := ops[0]

	if p.localID != ops[1] {
		return role.ErrWrongKey
	}

	cr := &mpb.ChalInfo{}
	err := proto.Unmarshal(metaValue, cr)
	if err != nil {
		utils.MLogger.Error("unmarshal h failed: ", err)
		return err
	}

	if cr.GetUserID() != userID {
		return role.ErrInvalidInput
	}

	if cr.GetQueryID() != fsID {
		return role.ErrInvalidInput
	}

	// incase get block has no group info
	go func() {
		_, ok := p.fsGroup.Load(fsID)
		if !ok {
			p.getGroupInfo(userID, fsID, true)
		}
	}()

	blskey, err := p.getNewUserConfig(userID, fsID)
	if err != nil {
		utils.MLogger.Warnf("get new user %s config from failed: %s ", fsID, err)
		return err
	}

	if blskey == nil || blskey.Pk == nil {
		utils.MLogger.Warn("get empty user`s config for: ", fsID)
		return nil
	}

	var chal mcl.Challenge
	chal.Seed = mcl.GenChallenge(cr)

	bset := bitset.New(0)
	err = bset.UnmarshalBinary(cr.GetChunkMap())
	if err != nil {
		return err
	}

	startPos := uint(chal.Seed) % bset.Len()
	chalNum := bset.Count()
	meta := false

	switch cr.GetPolicy() {
	case "100":
		chalNum = 100
	case "1%":
		chalNum = chalNum / 100
	case "smart":
		if chalNum/100 < 100 {
			chalNum = 100
		} else {
			chalNum = chalNum / 100
		}
	case "meta":
		meta = true
		utils.MLogger.Info("handle meta challenge: ", km.ToString(), " from: ", from)
	default:
	}

	ctx := p.context

	bucketNum := len(cr.GetBuckets())
	bucketID := 0
	stripeID := 0
	chunkID := 0
	stripeNum := int64(0)
	chunkNum := 0
	count := uint(0)
	electedOffset := 0

	for i, e := bset.NextSet(startPos); e; i, e = bset.NextSet(i + 1) {
		count++
		for j := bucketID; j < bucketNum; j++ {
			if int64(i) >= stripeNum && int64(i) <
				stripeNum+cr.Buckets[j].GetStripeNum()*int64(cr.Buckets[j].GetChunkNum()) {
				bucketID = j
				chunkNum = int(cr.Buckets[j].GetChunkNum())
				break
			}
			stripeNum += cr.Buckets[j].GetStripeNum() * int64(cr.Buckets[j].GetChunkNum())
		}

		if int64(i) < stripeNum || chunkNum == 0 {
			break
		}

		stripeID = int((int64(i) - stripeNum) / int64(chunkNum))
		chunkID = int((int64(i) - stripeNum) % int64(chunkNum))
		buf.Reset()
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BlockDelimiter)
		if meta {
			buf.WriteString(strconv.Itoa(-bucketID))
		} else {
			buf.WriteString(strconv.Itoa(bucketID))
		}
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum := int(cr.Buckets[bucketID].GetSegCount())
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		cbuf.Reset()
		cbuf.WriteString(blockID)
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(int(mpb.KeyType_Block)))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(electedOffset))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString("1") // length

		tmpdata, err := p.ds.GetBlock(ctx, cbuf.String(), nil, "local")
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag at %d failed: %s", blockID, electedOffset, err)
			failchunk = true
			continue
		}

		tmpseg, tmptag, segStart, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag failed", blockID)
			failchunk = true
			continue
		}

		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])

		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(segStart))
		chal.Indices = append(chal.Indices, buf.String())
		bset.SetTo(i, false)
		if count > chalNum {
			break
		}
	}

	bucketID = 0
	stripeNum = 0

	for i, e := bset.NextSet(0); e && i < startPos; i, e = bset.NextSet(i + 1) {
		if count > chalNum {
			break
		}
		count++
		for j := bucketID; j < bucketNum; j++ {
			if int64(i) >= stripeNum && int64(i) <
				stripeNum+cr.Buckets[j].GetStripeNum()*int64(cr.Buckets[j].GetChunkNum()) {
				bucketID = j
				chunkNum = int(cr.Buckets[j].GetChunkNum())
				break
			}

			stripeNum += cr.Buckets[j].GetStripeNum() * int64(cr.Buckets[j].GetChunkNum())
		}

		if int64(i) < stripeNum || chunkNum == 0 {
			break
		}

		stripeID = int((int64(i) - stripeNum) / int64(chunkNum))
		chunkID = int((int64(i) - stripeNum) % int64(chunkNum))

		buf.Reset()
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BlockDelimiter)
		if meta {
			buf.WriteString(strconv.Itoa(-bucketID))
		} else {
			buf.WriteString(strconv.Itoa(bucketID))
		}
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(stripeID))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(chunkID))
		blockID := buf.String()

		segNum := int(cr.Buckets[bucketID].GetSegCount())
		electedOffset = int((chal.Seed + int64(i)) % int64(segNum))

		cbuf.Reset()
		cbuf.WriteString(blockID)
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(int(mpb.KeyType_Block)))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(electedOffset))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString("1") // length

		tmpdata, err := p.ds.GetBlock(ctx, cbuf.String(), nil, "local")
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag at %d failed: %s", blockID, electedOffset, err)
			failchunk = true
			continue
		}

		tmpseg, tmptag, segStart, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag failed", blockID)
			failchunk = true
			continue
		}

		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])

		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(segStart))
		chal.Indices = append(chal.Indices, buf.String())
		bset.SetTo(i, false)
		if count > chalNum {
			break
		}
	}

	if len(chal.Indices) == 0 {
		utils.MLogger.Errorf("GenProof for %s fails due to no available data", fsID)
		return nil
	}

	proof, err := blskey.GenProof(chal, data, tag, 32)
	if err != nil {
		utils.MLogger.Error("GenProof err: ", err)
		return err
	}

	// 在发送之前检查生成的proof
	boo, err := blskey.VerifyProof(chal, proof, true)
	if err != nil || !boo {
		utils.MLogger.Errorf("gen proof for blocks: %s failed: %s", chal.Indices, err)
		return err
	}

	utils.MLogger.Info("handle challenge: ", km.ToString(), " gen right proof")

	mustr := b58.Encode(proof.Mu)
	nustr := b58.Encode(proof.Nu)
	deltastr := b58.Encode(proof.Delta)

	retValue := mustr + metainfo.DELIMITER + nustr + metainfo.DELIMITER + deltastr

	if failchunk {
		failMap, err := bset.MarshalBinary()
		if err != nil {
			return err
		}
		retValue = retValue + metainfo.DELIMITER + b58.Encode(failMap)
	}

	// provider发回挑战结果,其中proof结构体序列化，作为字符串用Proof返回
	_, err = p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Put), km.ToString(), []byte(retValue), nil, from)
	if err != nil {
		utils.MLogger.Info("send proof err: ", err)
	}
	return nil
}
