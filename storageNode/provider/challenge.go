package provider

import (
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
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

	hProto := &mpb.ChalInfo{}
	err = proto.Unmarshal(metaValue, hProto)
	if err != nil {
		utils.MLogger.Error("unmarshal h failed: ", err)
	}

	var chal mcl.Challenge
	chal.Seed = mcl.GenChallenge(hProto)

	// 聚合
	var data, tag [][]byte
	var faultBlocks []string
	var electedOffset int
	var buf, cbuf strings.Builder
	ctx := p.context
	for _, index := range hProto.Blocks {
		if len(index) == 0 {
			continue
		}
		buf.Reset()
		bid, off, err := utils.SplitIndex(index)
		if err != nil {
			continue
		}
		if off < 0 {
			faultBlocks = append(faultBlocks, index)
			continue
		} else if off > 0 {
			electedOffset = chal.Seed % off
		} else {
			electedOffset = 0
		}
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(bid)
		blockID := buf.String()
		buf.WriteString(metainfo.BlockDelimiter)
		buf.WriteString(strconv.Itoa(electedOffset))
		electedIndex := buf.String()

		cbuf.Reset()
		cbuf.WriteString(blockID)
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(int(mpb.KeyType_Block)))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString(strconv.Itoa(electedOffset))
		cbuf.WriteString(metainfo.DELIMITER)
		cbuf.WriteString("1")

		tmpdata, err := p.ds.GetBlock(ctx, cbuf.String(), nil, "local")
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag at %d failed: %s", blockID, electedOffset, err)
			faultBlocks = append(faultBlocks, index)
			continue
		}

		tmpseg, tmptag, isTrue := df.GetSegAndTag(tmpdata.RawData(), blockID, blskey)
		if !isTrue {
			utils.MLogger.Warnf("verify %s data and tag failed", blockID)
			faultBlocks = append(faultBlocks, index)
			continue
		}
		data = append(data, tmpseg[0])
		tag = append(tag, tmptag[0])
		chal.Indices = append(chal.Indices, electedIndex)
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
	if err != nil {
		utils.MLogger.Error("verify proof failed, err is: ", err)
		utils.MLogger.Error("gen proof for blocks: ", chal.Indices)
		return err
	}

	if !boo {
		utils.MLogger.Warn("proof is false")
		return mcl.ErrProofVerifyInProvider
	}

	utils.MLogger.Info("handle challenge: ", km.ToString(), " gen right proof")

	mustr := b58.Encode(proof.Mu)
	nustr := b58.Encode(proof.Nu)
	deltastr := b58.Encode(proof.Delta)

	retValue := mustr + metainfo.DELIMITER + nustr + metainfo.DELIMITER + deltastr

	if len(faultBlocks) > 0 {
		retValue = retValue + metainfo.DELIMITER + b58.Encode([]byte(strings.Join(faultBlocks, metainfo.DELIMITER)))
	}

	// provider发回挑战结果,其中proof结构体序列化，作为字符串用Proof返回
	_, err = p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(retValue), nil, from)
	if err != nil {
		utils.MLogger.Info("send proof err: ", err)
	}
	return nil
}
