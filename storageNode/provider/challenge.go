package provider

import (
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
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
		return nil
	}

	fsID := km.GetMid()
	userID := ops[0]
	utils.MLogger.Info("receive: ", fsID, " 's challenge from: ", from)
	blskey, err := p.getNewUserConfig(userID, fsID)
	if err != nil {
		utils.MLogger.Warnf("get new user %s config from failed: %s ", fsID, err)
		return err
	}

	if blskey == nil || blskey.Pk == nil {
		utils.MLogger.Warn("get empty user`s config for: ", fsID)
		return nil
	}

	chalInfo := &mpb.ChalInfo{}
	err = proto.Unmarshal(metaValue, chalInfo)
	if err != nil {
		utils.MLogger.Error("unmarshal h failed: ", err)
		return err
	}

	var chal mcl.Challenge
	chal.Seed = mcl.GenChallenge(chalInfo)

	bset := bitset.New(uint(chalInfo.GetBucketNum()))
	err = bset.UnmarshalBinary(chalInfo.GetChunkMap())
	if err != nil {
		return err
	}

	totalNum := bset.Count()
	startPos := uint(chal.Seed) % bset.Len()
	if startPos < uint(chalInfo.GetBucketNum()+1)*3 {
		startPos = uint(chalInfo.GetBucketNum()+1) * 3
	}

	var data, tag [][]byte
	var buf, cbuf strings.Builder
	failchunk := false
	ctx := p.context
	for i, e := bset.NextSet(0); e && i < uint(chalInfo.BucketNum+1)*3; i, e = bset.NextSet(i + 1) {
		buf.Reset()
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(-int(i / 3)))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString("0")
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(int(i % 3)))
		blockID := buf.String()

		cbuf.Reset()
		cbuf.WriteString(blockID)
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(int(mpb.KeyType_Block)))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.FormatInt(chal.Seed+int64(i), 10))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString("1") // length

		tmpdata, err := p.ds.GetBlock(ctx, cbuf.String(), nil, "local")
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag at %d failed: %s", blockID, chal.Seed, err)
			bset.SetTo(i, false)
			failchunk = true
			continue
		}

		tmpseg, tmptag, segStart, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag failed", blockID)
			bset.SetTo(i, false)
			failchunk = true
			continue
		}

		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])

		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(segStart))
		chal.Indices = append(chal.Indices, buf.String())
	}

	count := uint(0)
	bucket := 1
	chunkNum := chalInfo.GetChunkNum()[0]
	stripeNum := 3 * (chalInfo.GetBucketNum() + 1)
	for i, e := bset.NextSet(startPos); e; i, e = bset.NextSet(i + 1) {
		count++
		for j := bucket; j <= int(chalInfo.GetBucketNum()); j++ {
			if stripeNum+chalInfo.GetStripeNum()[j-1]*int64(chalInfo.GetChunkNum()[j-1]) < int64(i) {
				break
			}
			bucket = j
			chunkNum = chalInfo.GetChunkNum()[j-1]
			stripeNum += chalInfo.GetStripeNum()[j-1] * int64(chalInfo.GetChunkNum()[j-1])
		}

		if int64(i) < stripeNum {
			break
		}

		buf.Reset()
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(bucket))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.FormatInt((int64(i)-stripeNum)/int64(chunkNum), 10))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.FormatInt((int64(i)-stripeNum)%int64(chunkNum), 10))
		blockID := buf.String()

		cbuf.Reset()
		cbuf.WriteString(blockID)
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(int(mpb.KeyType_Block)))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.FormatInt(chal.Seed+int64(i), 10))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString("1") // length

		tmpdata, err := p.ds.GetBlock(ctx, cbuf.String(), nil, "local")
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag at %d failed: %s", blockID, chal.Seed, err)
			bset.SetTo(i, false)
			failchunk = true
			continue
		}

		tmpseg, tmptag, segStart, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag failed", blockID)
			bset.SetTo(i, false)
			failchunk = true
			continue
		}

		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])

		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(segStart))
		chal.Indices = append(chal.Indices, buf.String())
		if count > totalNum/100 {
			break
		}
	}

	for i, e := bset.NextSet(uint(chalInfo.BucketNum+1) * 3); e && i < startPos; i, e = bset.NextSet(i + 1) {
		if count > totalNum/100 {
			break
		}
		count++
		for j := bucket; j <= int(chalInfo.GetBucketNum()); j++ {
			if stripeNum+chalInfo.GetStripeNum()[j-1]*int64(chalInfo.GetChunkNum()[j-1]) < int64(i) {
				break
			}
			bucket = j
			chunkNum = chalInfo.GetChunkNum()[j-1]
			stripeNum += chalInfo.GetStripeNum()[j-1] * int64(chalInfo.GetChunkNum()[j-1])
		}

		if int64(i) < stripeNum {
			break
		}

		buf.Reset()
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(bucket))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.FormatInt((int64(i)-stripeNum)/int64(chunkNum), 10))
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.FormatInt((int64(i)-stripeNum)%int64(chunkNum), 10))
		blockID := buf.String()

		cbuf.Reset()
		cbuf.WriteString(blockID)
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(int(mpb.KeyType_Block)))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.FormatInt(chal.Seed+int64(i), 10))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString("1") // length

		tmpdata, err := p.ds.GetBlock(ctx, cbuf.String(), nil, "local")
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag at %d failed: %s", blockID, chal.Seed, err)
			bset.SetTo(i, false)
			failchunk = true
			continue
		}

		tmpseg, tmptag, segStart, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag failed", blockID)
			bset.SetTo(i, false)
			failchunk = true
			continue
		}

		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])

		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(segStart))
		chal.Indices = append(chal.Indices, buf.String())
		if count > totalNum/100 {
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
	_, err = p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(retValue), nil, from)
	if err != nil {
		utils.MLogger.Info("send proof err: ", err)
	}
	return nil
}
