package provider

import (
	"context"
	"crypto/sha256"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/memoio/go-mefs/crypto/aes"
	df "github.com/memoio/go-mefs/data-format"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	pb "github.com/memoio/go-mefs/source/go-block-format/pb"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

const (
	lowWater  = 0.8 // 数据生成为总量的80%
	highWater = 0.9 // 使用量达到90%后，删除10%的块
	mullen    = 100 * 1024 * 1024
	offset    = 25599
)

// uid is defined in utils/pos

var curGid = -1024
var curSid = -1
var posID string
var groupID string
var posAddr string
var posCidPrefix string
var inGenerate int
var keeperIDs []string
var posSkByte []byte

var pre = &pb.Prefix{
	Version:     1,
	Policy:      df.MulPolicy,
	DataCount:   1,
	ParityCount: 4,
	TagFlag:     df.BLS12,
	SegmentSize: df.DefaultSegmentSize,
}

var opt = &df.DataCoder{
	Prefix: pre,
}

// PosService starts pos
func (p *Info) PosService(ctx context.Context, gc bool) {
	// 获取合约地址一次，主要是获取keeper，用于发送block meta
	// handleUserDeployedContracts()
	posID = pos.GetPosId()
	posAddr = pos.GetPosAddr()
	posSkByte = pos.GetPosSkByte()
	groupID := pos.GetPosGID()

	gp := p.getGroupInfo(groupID, posID, true)
	if gp == nil {
		return
	}
	err := gp.getContracts(p.localID)
	if err == nil {
		return
	}

	//填充opt.KeySet
	err = p.getUserConifg(groupID, posID)
	if err != nil {
		return
	}

	opt.PreCompute()

	//从磁盘读取存储的Cidprefix
	posKM, err := metainfo.NewKeyMeta(groupID, metainfo.PosMeta)
	if err != nil {
		log.Println("NewKeyMeta posKM error :", err)
	} else {
		log.Println("posKm :", posKM.ToString())
		posValue, err := p.ds.GetKey(ctx, posKM.ToString(), "local")
		if err != nil {
			log.Println("Get posKM from local error :", err)
		} else {
			log.Println("posvalue :", string(posValue))
			posCidPrefix = string(posValue)
			cidInfo, err := metainfo.GetBlockMeta(string(posValue) + "_0")
			if err != nil {
				log.Println("get block meta in posRegular error :", err)
			} else {
				curGid, err = strconv.Atoi(cidInfo.GetBid()[utils.IDLength:])
				if err != nil {
					log.Println("strconv.Atoi Gid in posReguar error :", err)
				}
				curSid, err = strconv.Atoi(cidInfo.GetSid())
				if err != nil {
					log.Println("strconv.Atoi Sid in posReguar error :", err)
				}
			}
		}
	}

	p.traversePath(gc)
	log.Println("pos blocks reaches gid: ", curGid, ", sid: ", curSid)

	//开始pos
	p.posRegular(ctx)
}

// posRegular checks posBlocks and decide to add/delete
func (p *Info) posRegular(ctx context.Context) {
	log.Println("Pos start!")

	p.doGenerateOrDelete()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if inGenerate == 0 {
				// 如果超过了90%，则删除10%容量的posBlocks；如果低于80%，则生成到80%
				go p.doGenerateOrDelete()
			}
		}
	}
}

func (p *Info) traversePath(gc bool) {
	if gc {
		log.Println("clean pos blocks first")
	}
	exist := false
	gid := 0
	for {
		sid := 0
		for sid = 0; sid < 1024; sid++ {
			for i := 0; i < 5; i++ {
				posCid := posID + "_" + p.localID + strconv.Itoa(gid) + "_" + strconv.Itoa(sid) + "_" + strconv.Itoa(i)
				ncid := cid.NewCidV2([]byte(posCid))
				exist, err := p.ds.BlockStore().Has(ncid)
				if err != nil {
					continue
				}

				if exist {
					if gc {
						p.ds.BlockStore().DeleteBlock(ncid)
					}
				} else {
					break
				}
			}
			if !exist {
				break
			}
		}
		if exist && sid == 1024 {
			gid += 1024
			continue
		}
		if gc {
			curSid = -1024
			curGid = -1
		} else {
			curSid = (sid + 1023) % 1024
			if curSid == 1023 {
				curGid = gid - 1024
				if curGid < 0 {
					curSid = -1
				}
			}
		}

		posCidPrefix = posID + "_" + p.localID + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)

		break
	}
}

func (p *Info) doGenerateOrDelete() {
	inGenerate = 1
	defer func() {
		inGenerate = 0
	}()
	usedSpace, err := p.getDiskUsage()
	if err != nil {
		return
	}

	totalSpace := p.getDiskTotal()

	ratio := float64(usedSpace) / float64(totalSpace)
	log.Println("usedSpace is: ", usedSpace, ", totalSpace is: ", totalSpace, ",ratio is: ", ratio)

	if ratio <= lowWater {
		p.generatePosBlocks(uint64(float64(totalSpace) * (lowWater - ratio)))
	} else if ratio >= highWater {
		p.deletePosBlocks(uint64(usedSpace / 10))
	}
}

func (p *Info) uploadMulpolicy(data []byte) ([][]byte, int, error) {
	// 构建加密秘钥
	buckid := p.localID + strconv.Itoa(curGid)
	tmpkey := []byte(string(posSkByte) + buckid)
	skey := sha256.Sum256(tmpkey)
	// 加密、Encode
	data, err := aes.AesEncrypt(data, skey[:])
	if err != nil {
		return nil, 0, err
	}
	encodeData, offset, err := opt.Encode(data, posCidPrefix, 0)
	if err != nil {
		return nil, 0, err
	}
	return encodeData, offset, nil
}

// generatePosBlocks generate block accoding to the free space
func (p *Info) generatePosBlocks(increaseSpace uint64) {
	// fillRandom()
	// DataEncodeToMul()
	// send BlockMeta to keepers
	log.Println("generate pos blcoks")

	posKM, err := metainfo.NewKeyMeta(posID, metainfo.PosMeta)
	if err != nil {
		return
	}

	var totalIncreased uint64
	for {
		if totalIncreased >= increaseSpace {
			break
		}
		tmpData := make([]byte, mullen)
		totalIncreased += uint64(5 * len(tmpData))
		rand.Seed(time.Now().UnixNano())
		fillRandom(tmpData)
		// 配置部分
		// 更新stripeID、bucketID
		curSid = (curSid + 1) % 1024
		if curSid == 0 {
			curGid += 1024
		}

		posCidPrefix = posID + "_" + p.localID + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		data, _, err := p.uploadMulpolicy(tmpData)
		if err != nil {
			log.Println("UploadMulpolicy in generate Pos Blocks error :", err)
			continue
		}

		blockList := []string{}

		//做成块，放到本地
		for i, dataBlock := range data {
			blockID := posCidPrefix + "_" + strconv.Itoa(i)
			ncid := cid.NewCidV2([]byte(blockID))
			newblk, err := blocks.NewBlockWithCid(dataBlock, ncid)
			if err != nil {
				log.Println("New block failed, error :", err)
				continue
			}
			log.Println("New block success :", newblk.Cid())
			err = p.ds.BlockStore().Put(newblk)
			if err != nil {
				log.Println("add block failed, error :", err)
			}

			boff := blockID + "_" + strconv.Itoa(offset)

			blockList = append(blockList, boff)
		}

		// 向keeper发送元数据
		metaValue := strings.Join(blockList, metainfo.DELIMITER)
		km, err := metainfo.NewKeyMeta(p.localID, metainfo.Pos)
		for _, keeper := range keeperIDs {
			p.ds.SendMetaRequest(context.Background(), int32(metainfo.Put), km.ToString(), []byte(metaValue), nil, keeper)
		}

		// 本地更新
		posValue := posCidPrefix
		log.Println("posKM :", posKM.ToString(), ", posValue :", posValue)
		err = p.ds.PutKey(context.Background(), posKM.ToString(), []byte(posValue), "local")
		if err != nil {
			log.Println("CmdPutTo posKM error :", err)
			continue
		}
	}
}

func (p *Info) deletePosBlocks(decreseSpace uint64) {
	log.Println("data is about to exceed the space limit, delete pos blcoks")

	posKM, err := metainfo.NewKeyMeta(posID, metainfo.PosMeta)
	if err != nil {
		return
	}

	// delete last blocks
	var totalDecresed uint64
	for {
		if curGid == -1024 && curSid == -1 {
			return
		}

		if totalDecresed >= decreseSpace {
			break
		}
		//删除块
		deleteBlocks := []string{}
		j := 0
		for i := 0; i < 5; i++ {
			blockID := posCidPrefix + "_" + strconv.Itoa(i)
			ncid := cid.NewCidV2([]byte(blockID))
			err := p.ds.BlockStore().DeleteBlock(ncid)
			if err != nil {
				log.Println("delete block: ", blockID, " error :", err)
				j++
			} else {
				log.Println("delete block : ", blockID, " success")
				totalDecresed += uint64(mullen)
				deleteBlocks = append(deleteBlocks, blockID)
			}
		}

		//更新Gid,Sid
		curSid = (curSid + 1023) % 1024
		if curSid == 1023 {
			curGid -= 1024
		}

		if curGid == -1024 {
			curSid = -1
		}

		posCidPrefix = posID + "_" + p.localID + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		log.Println("after delete ,Gid :", curGid, ", sid :", curSid, ", cid prefix :", posCidPrefix)

		posValue := posCidPrefix
		err = p.ds.PutKey(context.Background(), posKM.ToString(), []byte(posValue), "local")
		if err != nil {
			log.Println("CmdPutTo posKM error :", err)
			continue
		}

		// send BlockMeta deletion to keepers
		//发送元数据到keeper
		if j < 5 {
			km, err := metainfo.NewKeyMeta(p.localID, metainfo.Pos)
			if err != nil {
				log.Println("construct put blockMeta KV error :", err)
				return
			}
			metavalue := strings.Join(deleteBlocks, metainfo.DELIMITER)
			for _, keeper := range keeperIDs {
				p.ds.SendMetaRequest(context.Background(), int32(metainfo.Delete), km.ToString(), []byte(metavalue), nil, keeper)
			}
		}
	}
}

func (p *Info) getUserConifg(userID, groupID string) error {
	// 需要用私钥decode出bls的私钥，用user中的方法
	//获取公钥

	pubKey, err := p.getNewUserConfig(userID, groupID)
	if err != nil {
		log.Println("getNewUserConfig in get userconfig error :", err)
		return err
	}

	opt.BlsKey = pubKey

	//获取私钥
	opt.BlsKey.Sk, err = p.getUserPrivateKey(userID, groupID)
	if err != nil {
		log.Println("getUserPrivateKey in get userconfig error ", err)
		return err
	}
	return nil
}

func fillRandom(p []byte) {
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
}
