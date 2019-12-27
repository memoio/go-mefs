package provider

import (
	"log"
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

func handleChallengeBls12(km *metainfo.KeyMeta, metaValue, from string) error {
	ops := km.GetOptions()

	if len(ops) < 1 {
		return nil
	}

	userID := km.GetMid()
	log.Println("receive", userID, " 's challenge from", from)
	pubKey, err := getNewUserConfig(userID, from)
	if err != nil {
		log.Println("get new user`s config from:", from, "failed, error :", err)
		return err
	}

	hProto := &pb.Chalnum{}
	hByte, _ := b58.Decode(metaValue)
	err = proto.Unmarshal(hByte, hProto)
	if err != nil {
		log.Println("unmarshal h failed, err: ", err)
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
		buf.WriteString(userID)
		buf.WriteString(metainfo.BLOCK_DELIMITER)
		buf.WriteString(bid)
		blockID := cid.NewCidV2([]byte(buf.String()))
		buf.WriteString(metainfo.BLOCK_DELIMITER)
		buf.WriteString(strconv.Itoa(electedOffset))
		electedIndex := buf.String()
		tmpdata, tmptag, err := localNode.Data..GetSegAndTag(blockID, uint64(electedOffset))
		if err != nil {
			faultBlocks = append(faultBlocks, index)
		} else {
			isTrue := mcl.VerifyTag(tmpdata, tmptag, electedIndex, pubKey)
			if !isTrue {
				log.Println("verify tag failed")
				//验证失败，则在本地删除此块
				err := localNode.Data.DeleteBlock(blockID)
				if err != nil {
					log.Println("Delete block", blockID.String(), "error:", err)
				}
				faultBlocks = append(faultBlocks, index)
			} else {
				data = append(data, tmpdata)
				tag = append(tag, tmptag)
				chal.Indices = append(chal.Indices, electedIndex)
			}
		}
	}

	proof, err := mcl.GenProof(pubKey, chal, data, tag)
	if err != nil {
		log.Println("GenProof err: ", err)
		return err
	}

	// 在发送之前检查生成的proof
	boo, err := mcl.VerifyProof(pubKey, chal, proof)
	if err != nil {
		log.Println("verify proof failed, err is: ", err)
		return err
	}

	if !boo {
		log.Println("proof is false")
		return mcl.ErrProofVerifyInProvider
	}

	log.Println("proof is right")

	retKm, err := metainfo.NewKeyMeta(userID, metainfo.Proof, ops[0])
	if err != nil {
		return err
	}

	retValue := proof

	if len(faultBlocks) > 0 {
		retValue = retValue + metainfo.DELIMITER + b58.Encode([]byte(strings.Join(faultBlocks, metainfo.DELIMITER)))
	}

	// provider发回挑战结果,其中proof结构体序列化，作为字符串用Proof返回
	_, err = localNode.Data.SendMetaRequest(retKm, retValue, from)
	if err != nil {
		log.Println("send proof err: ", err)
	}
	return nil
}
