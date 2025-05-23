package provider

import (
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/pdp"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

// key: qid/"Challenge"/uid/pid/kid/chaltime
func (p *Info) handleChallengeBls12(km *metainfo.Key, metaValue []byte, from string) error {
	utils.MLogger.Info("handle challenge: ", km.ToString(), " from: ", from)

	ops := km.GetOptions()
	if len(ops) < 4 {
		return role.ErrWrongKey
	}

	fsID := km.GetMainID()
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
	gp := p.getGroupInfo(userID, fsID, true)
	if gp == nil || !gp.status {
		return role.ErrServiceNotReady
	}

	blskey, err := p.getNewUserConfig(userID, fsID)
	if err != nil {
		utils.MLogger.Warnf("get new user %s config from failed: %s ", fsID, err)
		return err
	}

	if blskey == nil || blskey.PublicKey() == nil {
		utils.MLogger.Warn("get empty user`s config for: ", fsID)
		return nil
	}

	var proof pdp.ProofWithVersion
	proof.Ver = pdp.PDPV1
	var faultValue string
	switch cr.GetPolicy() {
	case "smart", "meta":
		pr, err := p.handleChallenge(cr, blskey)
		if err != nil {
			return err
		}
		proof.Proof = pr
		if len(cr.GetFailMap()) > 0 {
			faultValue = metainfo.DELIMITER + b58.Encode(cr.GetFailMap())
		}
	case "random100":
		pr, err := p.handleChallengeRandom(cr, blskey)
		if err != nil {
			return err
		}
		proof.Proof = pr
		if len(cr.GetFaultBlocks()) > 0 {
			faultValue = metainfo.DELIMITER + b58.Encode([]byte(strings.Join(cr.GetFaultBlocks(), metainfo.DELIMITER)))
		}
	default:
	}

	utils.MLogger.Info("handle challenge: ", km.ToString(), " gen right proof")

	pf, err := proof.Serialize()
	if err != nil {
		utils.MLogger.Warnf("proof serialize failed : %s, %s", fsID, err)
		return err
	}
	retValue := b58.Encode(pf) + faultValue

	// provider发回挑战结果,其中proof结构体序列化，作为字符串用Proof返回
	_, err = p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(retValue), nil, from)
	if err != nil {
		utils.MLogger.Info("send proof err: ", err)
	}
	return nil
}

func (p *Info) handleChallenge(cr *mpb.ChalInfo, blskey pdp.KeySet) (pdp.Proof, error) {
	var data, tag [][]byte
	failchunk := false

	var chal pdp.ChallengeV1
	chal.R = pdp.GenChallengeV1(cr)

	bset := bitset.New(0)
	err := bset.UnmarshalBinary(cr.GetChunkMap())
	if err != nil {
		return nil, err
	}

	chalNum := bset.Count()
	meta := false

	startPost := uint(chal.R) % bset.Len()

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

	var buf, cbuf strings.Builder
	for i, e := bset.NextSet(startPost); e; i, e = bset.NextSet(i + 1) {
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
		buf.WriteString(cr.GetQueryID())
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
		electedOffset = int((chal.R + int64(i)) % int64(segNum))

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

	for i, e := bset.NextSet(0); e && i < startPost; i, e = bset.NextSet(i + 1) {
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
		buf.WriteString(cr.GetQueryID())
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
		electedOffset = int((chal.R + int64(i)) % int64(segNum))

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
		utils.MLogger.Errorf("GenProof for %s fails due to no available data", cr.GetQueryID())
		return nil, role.ErrEmptyData
	}

	proof, err := blskey.PublicKey().GenProof(&chal, data, tag, 32)
	if err != nil {
		utils.MLogger.Error("GenProof err: ", err)
		return nil, err
	}

	// 在发送之前检查生成的proof
	boo, err := blskey.VerifyKey().VerifyProof(&chal, proof)
	if err != nil {
		utils.MLogger.Errorf("gen proof for blocks: %s failed: %s", chal.Indices, err)
		return nil, err
	}

	if !boo {
		return nil, role.ErrWrongState
	}

	if failchunk {
		failMap, err := bset.MarshalBinary()
		if err != nil {
			return nil, err
		}
		cr.FailMap = failMap
	}

	return proof, nil
}

func (p *Info) handleChallengeRandom(cr *mpb.ChalInfo, blskey pdp.KeySet) (pdp.Proof, error) {
	var chal pdp.ChallengeV1
	chal.R = pdp.GenChallengeV1(cr)

	// 聚合
	var data, tag [][]byte
	var faultBlocks []string
	var electedOffset int
	var buf, cbuf strings.Builder
	ctx := p.context
	for _, index := range cr.Blocks {
		if len(index) == 0 {
			continue
		}
		buf.Reset()
		bid, off, err := metainfo.GetBidAndOffset(index)
		if err != nil {
			continue
		}
		if off < 0 {
			faultBlocks = append(faultBlocks, index)
			continue
		} else if off > 0 {
			electedOffset = int(chal.R) % off
		} else {
			electedOffset = 0
		}
		buf.WriteString(cr.GetQueryID())
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(bid)
		blockID := buf.String()

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
			faultBlocks = append(faultBlocks, index)
			continue
		}

		tmpseg, tmptag, segStart, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag at %d failed", blockID, electedOffset)
			faultBlocks = append(faultBlocks, index)
			continue
		}

		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])

		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(segStart))
		chal.Indices = append(chal.Indices, buf.String())
	}

	if len(chal.Indices) == 0 {
		utils.MLogger.Errorf("GenProof random for %s fails due to no available data", cr.GetQueryID())
		return nil, role.ErrEmptyData
	}

	proof, err := blskey.PublicKey().GenProof(&chal, data, tag, 32)
	if err != nil {
		utils.MLogger.Error("GenProof err: ", err)
		return nil, err
	}

	boo, err := blskey.VerifyKey().VerifyProof(&chal, proof)
	if err != nil {
		utils.MLogger.Errorf("gen proof for blocks: %s failed: %s", chal.Indices, err)
		return nil, err
	}

	if !boo {
		return nil, role.ErrWrongState
	}

	cr.FaultBlocks = faultBlocks

	return proof, nil
}
