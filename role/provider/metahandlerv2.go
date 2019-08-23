package provider

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	mcl "github.com/memoio/go-mefs/bls12"
	rs "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

const (
	RepairFailed  = "Repair Failed"
	RepairSuccess = "Repair Successes"
)

// ProviderHandlerV2 provider角色回调接口的实现，
type ProviderHandlerV2 struct {
	Role string
}

// HandleMetaMessage provider角色层metainfo的回调函数,传入对方节点发来的kv，和对方节点的peerid
//没有返回值时，返回complete，或者返回规定信息
func (provider *ProviderHandlerV2) HandleMetaMessage(metaKey, metaValue, from string) (string, error) {
	km, err := metainfo.GetKeyMeta(metaKey)
	if err != nil {
		return "", err
	}
	keytype := km.GetKeyType()
	switch keytype {
	case metainfo.Test:
		go handleTest(km)
	case metainfo.UserInitReq:
		fmt.Println("keytype：UserInitReq 不处理")
	case metainfo.UserDeployedContracts:
		go handleUserDeployedContracts(km, metaKey, from)
	case metainfo.Challenge:
		go handleChallengeBls12(km, metaValue, from)
	case metainfo.Repair:
		go handleRepair(km, metaValue, from)
	case metainfo.StorageSync:
		go hanldeStorageSync(from)
	case metainfo.DeleteBlock:
		go handleDeleteBlock(km, from)
	case metainfo.GetBlock:
		res, err := handleGetBlock(km, from)
		if err == nil {
			return res, nil
		}
	case metainfo.PutBlock:
		handlePutBlock(km, metaValue, from)
	default: //没有匹配的信息，报错
		return "", metainfo.ErrWrongType
	}
	return metainfo.MetaHandlerComplete, nil
}

// 获取这个节点的角色信息，返回错误说明provider还没有启动好
func (provider *ProviderHandlerV2) GetRole() (string, error) {
	return provider.Role, nil
}

func handleTest(km *metainfo.KeyMeta) {
	fmt.Println("测试用回调函数")
	fmt.Println("km.mid:", km.GetMid())
	fmt.Println("km.options", km.GetOptions())
}

func handleUserDeployedContracts(km *metainfo.KeyMeta, metaValue, from string) error {
	fmt.Println("NewUserDeployedContracts", km.ToString(), metaValue, "From:", from)
	err := SaveUpkeeping(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Upkeeping err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Upkeeping success")
	}
	err = SaveChannel(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Channel err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Channel success")
	}
	err = SaveQuery(km.GetMid())
	if err != nil {
		fmt.Println("Save ", km.GetMid(), "'s Query err", err)
	} else {
		fmt.Println("Save ", km.GetMid(), "'s Query success")
	}
	err = SaveOffer()
	if err != nil {
		fmt.Println("Save ", localNode.Identity.Pretty(), "'s Offer err", err)
	} else {
		fmt.Println("Save ", localNode.Identity.Pretty(), "'s Offer success")
	}
	return nil
}

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

	//fmt.Printf("Receive challenge-Random:%d, FoundBlock-%s, FaultBlock-%s\n", chal.C, FoundBlock, FaultBlock)
	//splitedMetaKey[2]是h,splitedMetaKey[3]是挑战发起时间
	proof, err := mcl.GenProof(pubKey, chal, data, tag)
	if err != nil {
		fmt.Println("GenProof err-", err)
		return err
	}
	// 在发送之前检查生成的proof
	boo, err := mcl.Verify(pubKey, chal, proof)
	if err != nil {
		fmt.Println("verify proof failed")
		return err
	}
	if !boo {
		fmt.Println("boo :", boo)
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

func handleRepair(km *metainfo.KeyMeta, rpids, keeper string) error {
	var nbid int
	var cids []string
	var ret string
	var sig []byte
	blockID := km.GetMid()
	userID := blockID[:IDLength]
	tmpPubKey, ok := usersConfigs.Load(userID)
	if !ok {
		tmpUserCongfig, err := getNewUserConfig(userID, keeper)
		if err != nil {
			fmt.Println("get new user`s config failed,error :", err)
			return err
		}
		usersConfigs.Store(userID, tmpUserCongfig.PubKey)
		tmpPubKey = tmpUserCongfig.PubKey
	}
	cpids := strings.Split(rpids, metainfo.DELIMITER)
	stripe := make([][]byte, len(cpids)-1)
	for _, cpid := range cpids {
		if len(cpid) > 0 {
			splitcpid := strings.Split(cpid, metainfo.REPAIR_DELIMETER)
			cids = append(cids, splitcpid[0])
			if strings.Compare(splitcpid[0], blockID) != 0 {
				pid := splitcpid[1]
				blk, err := localNode.Blocks.GetBlockFrom(localNode.Context(), pid, splitcpid[0], time.Minute, sig)
				if blk != nil && err == nil {
					right := rs.VerifyBlock(blk.RawData(), splitcpid[0], tmpPubKey.(*mcl.PublicKey))
					if right {
						blkMeta, err := metainfo.GetBlockMeta(splitcpid[0])
						if err != nil {
							fmt.Println("get block meta error :", err)
							return err
						}
						i, err := strconv.Atoi(blkMeta.GetBid())
						if err != nil {
							fmt.Println("strconv.Atoi error :", err)
							return err
						}
						if i >= len(stripe) {
							for j := len(stripe); j <= i; j++ {
								stripe = append(stripe, nil)
							}
						}
						stripe[i] = make([]byte, len(blk.RawData()))
						stripe[i] = blk.RawData()
					} else {
						fmt.Println("block rawdata verify failed, error :", err)
						return err
					}
				} else {
					fmt.Println("GetBlock error :", err)
				}
			} else {
				cidMeta, err := metainfo.GetBlockMeta(blockID)
				if err != nil {
					fmt.Println("get block meta error :", err)
					return err
				}
				nbid, err = strconv.Atoi(cidMeta.GetBid())
				if err != nil {
					fmt.Println("strconv.Atoi error :", err)
					return err
				}
				//ret = cid|pid|offset
				ret = splitcpid[0] + metainfo.DELIMITER + splitcpid[1] + metainfo.DELIMITER + splitcpid[2]
				if nbid >= len(stripe) {
					for j := len(stripe); j <= nbid; j++ {
						stripe = append(stripe, nil)
					}
				}
			}
		}
	}

	retKm, err := metainfo.NewKeyMeta(blockID, metainfo.RepairRes)
	if err != nil {
		return err
	}
	newstripe, err := rs.Repair(stripe)
	if err != nil {
		fmt.Println("修复失败 ：", blockID, "\nrepair error :", err)
		retMetaValue := RepairFailed + metainfo.DELIMITER + ret
		fmt.Println("repair response metavalue :", retMetaValue)
		_, err = sendMetaRequest(retKm, retMetaValue, keeper)
		if err != nil {
			return err
		}
		//删除修复时get到的block；
		for _, tmpCid := range cids {
			if len(tmpCid) > 0 {
				temcid := cid.NewCidV2([]byte(tmpCid))
				err = localNode.Blocks.DeleteBlock(temcid)
				if err != nil && err != bs.ErrNotFound {
					fmt.Println("delete error :", err)
					return err
				}
			}
		}
		return nil
	}
	ncid := cid.NewCidV2([]byte(blockID))
	newblk, err := blocks.NewBlockWithCid(newstripe[nbid], ncid)
	if err != nil {
		fmt.Println("New block failed, error :", err)
		return err
	}
	//删除修复时get到的block；
	for _, tmpCid := range cids {
		if len(tmpCid) > 0 {
			temcid := cid.NewCidV2([]byte(tmpCid))
			err = localNode.Blocks.DeleteBlock(temcid)
			if err != nil && err != bs.ErrNotFound {
				fmt.Println("delete error :", err)
				return err
			}
		}
	}
	//把修复好的block放到本地；
	err = localNode.Blocks.PutBlock(newblk)
	if err != nil {
		fmt.Println("add block failed, error :", err)
		return err
	}
	retMetaValue := RepairSuccess + metainfo.DELIMITER + ret
	fmt.Println("repair response metavalue :", retMetaValue)
	fmt.Println("修复成功 ：", blockID)
	_, err = sendMetaRequest(retKm, retMetaValue, keeper)
	if err != nil {
		fmt.Println("repair response err :", err)
		return err
	}
	return nil
}

func hanldeStorageSync(kid string) error {
	cfg, err := localNode.Repo.Config()
	if err != nil {
		fmt.Println("get config failed :", err)
		return err
	}
	maxSpace := cfg.Datastore.StorageMax
	dataStore := localNode.Repo.Datastore()
	actulDataSpace, err := ds.DiskUsage(dataStore)
	if err != nil {
		fmt.Println("get disk usage failed :", err)
		return err
	}
	rawDataSpace := actulDataSpace
	km, err := metainfo.NewKeyMeta(kid, metainfo.StorageSync)
	if err != nil {
		fmt.Println("construct StorageSync KV error :", err)
		return err
	}
	value := maxSpace + metainfo.DELIMITER + strconv.FormatUint(actulDataSpace, 10) + metainfo.DELIMITER + strconv.FormatUint(rawDataSpace, 10)
	_, err = sendMetaRequest(km, value, kid)
	if err != nil {
		fmt.Println("send error :", err)
		return err
	}
	return nil
}

func handleDeleteBlock(km *metainfo.KeyMeta, from string) error {
	blockID := km.GetMid()
	bcid := cid.NewCidV2([]byte(blockID))
	err := localNode.Blocks.DeleteBlock(bcid)
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	return nil
}
