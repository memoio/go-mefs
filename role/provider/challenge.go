package provider

import (
	"fmt"
	"log"
	"strconv"

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
	userID := km.GetMid()
	pubKeyInterface, ok := usersConfigs.Load(userID)
	if !ok {
		tmpUserCongfig, err := getNewUserConfig(userID, from)
		if err != nil {
			fmt.Println("get new user`s config failed,error :", err)
			return err
		}
		usersConfigs.Store(userID, tmpUserCongfig.PubKey)
		pubKeyInterface = tmpUserCongfig.PubKey
	}
	pubKey := pubKeyInterface.(*mcl.PublicKey)

	usersConfigs.LoadOrStore(userID, pubKey)

	hProto := &pb.Chalnum{}
	hByte, _ := b58.Decode(metaValue)
	err := proto.Unmarshal(hByte, hProto)
	if err != nil {
		fmt.Println("unmarshal h failed")
	}

	var chal mcl.Challenge
	chal.C = int(hProto.PubC)

	// 聚合
	var data, tag [][]byte
	var FaultBlock, FoundBlock, uid string
	var electedOffset int
	for _, index := range hProto.Indices {
		if len(index) > 0 {
			bid, off, err := utils.SplitIndex(index)
			uid = bid[:IDLength]
			if err != nil {
				return err
			}
			if off < 0 {
				return mcl.ErrOffsetIsNegative
			} else if off > 0 {
				electedOffset = chal.C % off
			} else {
				electedOffset = 0
			}
			blockID := cid.NewCidV2([]byte(bid))
			electedIndex := bid + metainfo.BLOCK_DELIMITER + strconv.Itoa(electedOffset)
			offset := uint64(electedOffset)
			tmpdata, tmptag, err := localNode.Blockstore.GetSegAndTag(blockID, offset)
			if err != nil {
				FaultBlock = FaultBlock + metainfo.DELIMITER + electedIndex
			} else {
				isTrue := mcl.VerifyTag(tmpdata, tmptag, electedIndex, pubKey)
				if !isTrue {
					fmt.Println("verify tag failed")
					//验证失败，则在本地删除此块
					err := localNode.Blocks.DeleteBlock(blockID)
					if err != nil {
						log.Println("Delete block", blockID.String(), "error:", err)
					}
					FaultBlock = FaultBlock + metainfo.DELIMITER + electedIndex
				} else {
					data = append(data, tmpdata)
					tag = append(tag, tmptag)
					chal.Indices = append(chal.Indices, electedIndex)
					FoundBlock = FoundBlock + metainfo.DELIMITER + electedIndex
				}
			}
		}
	}

	proof, err := mcl.GenProof(pubKey, chal, data, tag)
	if err != nil {
		fmt.Println("GenProof err-", err)
		return err
	}
	// 在发送之前检查生成的proof
	boo, err := mcl.Verify(pubKey, chal, proof)
	if err != nil {
		fmt.Println("verify proof failed, err is: ", err)
		return err
	}
	if !boo {
		fmt.Println("proof is false")
		return mcl.ErrProofVerifyInProvider
	}
	retKm, err := metainfo.NewKeyMeta(uid, metainfo.Proof, b58.Encode([]byte(FaultBlock)), ops[0])
	if err != nil {
		return err
	}
	retMetaValue := proof
	// provider发回挑战结果,其中proof结构体序列化，作为字符串用Proof返回
	_, err = sendMetaRequest(retKm, retMetaValue, from)
	if err != nil {
		fmt.Println("send proof err :", err)
	}
	return nil
}
