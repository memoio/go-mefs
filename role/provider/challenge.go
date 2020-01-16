package provider

import (
	"context"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	pb "github.com/memoio/go-mefs/role/pb"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

// key: qid/"Challenge"/uid/pid/kid/chaltime
func (p *Info) handleChallengeBls12(km *metainfo.KeyMeta, metaValue []byte, from string) error {
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

	hProto := &pb.Chalnum{}
	err = proto.Unmarshal(metaValue, hProto)
	if err != nil {
		utils.MLogger.Error("unmarshal h failed: ", err)
	}

	var chal mcl.Challenge
	chal.C = int(hProto.PubC)

	// 聚合
	var data, tag [][]byte
	var faultBlocks []string
	var electedOffset int
	var buf strings.Builder
	for _, index := range hProto.Indices {
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
			electedOffset = chal.C % off
		} else {
			electedOffset = 0
		}
		buf.WriteString(fsID)
		buf.WriteString(metainfo.BLOCK_DELIMITER)
		buf.WriteString(bid)
		blockID := cid.NewCidV2([]byte(buf.String()))
		buf.WriteString(metainfo.BLOCK_DELIMITER)
		buf.WriteString(strconv.Itoa(electedOffset))
		electedIndex := buf.String()
		tmpdata, tmptag, err := p.ds.BlockStore().GetSegAndTag(blockID, uint64(electedOffset))
		if err != nil {
			utils.MLogger.Warnf("get %s data and tag  at %d failed: %s", blockID, electedOffset, err)
			faultBlocks = append(faultBlocks, index)
		} else {
			isTrue := blskey.VerifyTag(tmpdata, tmptag, electedIndex)
			if !isTrue {
				utils.MLogger.Warnf("verify %s data and tag failed", blockID)
				//验证失败，则在本地删除此块
				err := p.ds.DeleteBlock(context.Background(), blockID.String(), "local")
				if err != nil {
					utils.MLogger.Info("Delete block", blockID.String(), "error:", err)
				}
				faultBlocks = append(faultBlocks, index)
			} else {
				data = append(data, tmpdata)
				tag = append(tag, tmptag)
				chal.Indices = append(chal.Indices, electedIndex)
			}
		}
	}

	if len(chal.Indices) == 0 {
		utils.MLogger.Error("GenProof fails due to no available data")
		return nil
	}

	proof, err := blskey.GenProof(chal, data, tag, 32)
	if err != nil {
		utils.MLogger.Error("GenProof err: ", err)
		return err
	}

	// 在发送之前检查生成的proof
	boo, err := blskey.VerifyProof(chal, proof)
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
	_, err = p.ds.SendMetaRequest(context.Background(), int32(metainfo.Put), km.ToString(), []byte(retValue), nil, from)
	if err != nil {
		utils.MLogger.Info("send proof err: ", err)
	}
	return nil
}
