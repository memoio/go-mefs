package provider

import (
	"context"
	"math/rand"
	"strconv"
	"strings"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

const (
	lowWater  = 0.70 // 数据生成为总量的70%
	highWater = 0.85 // 使用量达到85%后，删除10%的块
	mullen    = 100 * 1024 * 1024
	rep       = 10 // 10备份
)

// uid is defined in utils/pos

var curGid = -1024
var curSid = -1
var posID string
var groupID string
var posAddr string
var posCidPrefix string
var inGenerate int

var opt = &df.DataCoder{
	Prefix: &mpb.BlockOptions{
		Bopts: &mpb.BucketOptions{
			Version:      1,
			Policy:       df.MulPolicy,
			DataCount:    1,
			ParityCount:  rep - 1,
			TagFlag:      df.BLS12,
			SegmentSize:  df.DefaultSegmentSize,
			Encryption:   0,
			SegmentCount: 25600,
		},
	},
}

// PosService starts pos
func (p *Info) PosService(ctx context.Context, gc bool) error {
	// 获取合约地址一次，主要是获取keeper，用于发送block meta
	// handleUserDeployedContracts()
	utils.MLogger.Info("Start Pos Service")
	posID = pos.GetPosId()
	posAddr = pos.GetPosAddr()

	qItem, err := role.GetLatestQuery(posID)
	if err != nil {
		return err
	}

	groupID = qItem.QueryID

	gp := p.getGroupInfo(posID, groupID, true)
	if gp == nil {
		return role.ErrEmptyData
	}

	km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, posID)
	if err != nil {
		return err
	}

	has := false
	for {
		// for challenge
		for _, pid := range gp.providers {
			if p.localID == pid {
				has = true
			}
		}

		if has {
			break
		}

		for _, keeper := range gp.keepers {
			utils.MLogger.Info("Send Pos add to keepers:", keeper)
			p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), km.ToString(), []byte(p.localID), nil, keeper)
		}

		time.Sleep(10 * time.Minute)
		gp.loadContracts(p.localID, true)
	}

	//填充opt.KeySet
	mkey, err := mcl.GenKeySetWithSeed(pos.GetPosSeed(), mcl.TagAtomNum, mcl.PDPCount)
	if err != nil {
		utils.MLogger.Info("Init bls config for pos user fail: ", err)
		return err
	}
	opt.BlsKey = mkey

	opt.PreCompute()

	//从磁盘读取存储的Cidprefix
	posKM, err := metainfo.NewKey(p.localID, mpb.KeyType_PosMeta)
	if err != nil {
		utils.MLogger.Info("NewKeyMeta posKM error :", err)
		return err
	}

	posValue, err := p.ds.GetKey(ctx, posKM.ToString(), "local")
	if err != nil {
		utils.MLogger.Info("Get posKM from local error :", err)
	} else {
		utils.MLogger.Info("get posKM value: ", string(posValue))
		posCidPrefix = string(posValue)
		cidInfo, err := metainfo.GetBlockMeta(string(posValue) + "_0")
		if err != nil {
			utils.MLogger.Info("get block meta in posRegular error :", err)
		} else {
			curGid, err = strconv.Atoi(cidInfo.GetBid()[utils.IDLength:])
			if err != nil {
				utils.MLogger.Info("strconv.Atoi Gid in posReguar error :", err)
			}
			curSid, err = strconv.Atoi(cidInfo.GetSid())
			if err != nil {
				utils.MLogger.Info("strconv.Atoi Sid in posReguar error :", err)
			}
		}
	}

	utils.MLogger.Info("before traverse pos blocks reaches gid: ", curGid, ", sid: ", curSid)

	p.traversePath(gc)

	utils.MLogger.Info("after traverse pos blocks reaches gid: ", curGid, ", sid: ", curSid)

	//开始pos
	p.posRegular(ctx)
	return nil
}

// posRegular checks posBlocks and decide to add/delete
func (p *Info) posRegular(ctx context.Context) {
	utils.MLogger.Info("Pos start!")

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
		utils.MLogger.Info("clean pos blocks first")
	}
	exist := false
	gid := 0
	var res strings.Builder
	for {
		sid := 0
		for sid = 0; sid < 1024; sid++ {
			for i := 0; i < rep; i++ {
				res.Reset()
				res.WriteString(groupID)
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(p.localID)
				res.WriteString(strconv.Itoa(gid))
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(strconv.Itoa(sid))
				res.WriteString(metainfo.BlockDelimiter)
				res.WriteString(strconv.Itoa(i))
				ncid := cid.NewCidV2([]byte(res.String()))
				exist, err := p.ds.BlockStore().Has(ncid)
				if err != nil {
					continue
				}

				if exist {
					if gc {
						p.ds.DeleteBlock(p.context, res.String(), "local")
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
		break
	}

	posCidPrefix = posID + "_" + p.localID + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
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
	utils.MLogger.Info("usedSpace is: ", usedSpace, ", totalSpace is: ", totalSpace, ",ratio is: ", ratio)

	if ratio < lowWater {
		p.generatePosBlocks(uint64(float64(totalSpace) * (lowWater - ratio)))
	} else if ratio > highWater {
		p.deletePosBlocks(uint64(usedSpace / 10))
	}
}

// generatePosBlocks generate block accoding to the free space
func (p *Info) generatePosBlocks(increaseSpace uint64) {
	utils.MLogger.Info("generate pos blocks for space: %d", increaseSpace)

	posKM, err := metainfo.NewKey(p.localID, mpb.KeyType_PosMeta)
	if err != nil {
		return
	}

	totalIncreased := uint64(0)
	for {
		if totalIncreased >= increaseSpace {
			break
		}
		tmpData := make([]byte, mullen)
		totalIncreased += uint64(10 * len(tmpData))
		rand.Seed(time.Now().UnixNano())
		fillRandom(tmpData)
		// 配置部分
		// 更新stripeID、bucketID
		curSid = (curSid + 1) % 1024
		if curSid == 0 {
			curGid += 1024
		}

		posCidPrefix = posID + "_" + p.localID + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		data, offset, err := opt.Encode(tmpData, posCidPrefix, 0)
		if err != nil {
			utils.MLogger.Info("UploadMulpolicy in generate Pos Blocks error: ", err)
			continue
		}

		blockList := []string{}

		//做成块，放到本地
		for i, dataBlock := range data {
			blockID := posCidPrefix + "_" + strconv.Itoa(i)

			err := p.ds.PutBlock(p.context, blockID, dataBlock, "local")
			if err != nil {
				utils.MLogger.Info("add block failed, error :", err)
			}
			boff := blockID + metainfo.BlockDelimiter + strconv.Itoa(offset)

			blockList = append(blockList, boff)
		}

		gp := p.getGroupInfo(posID, groupID, false)
		if gp == nil {
			return
		}

		// 向keeper发送元数据
		metaValue := strings.Join(blockList, metainfo.DELIMITER)
		km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, posID)

		for _, keeper := range gp.keepers {
			p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(metaValue), nil, keeper)
		}

		// 本地更新
		err = p.ds.PutKey(p.context, posKM.ToString(), []byte(posCidPrefix), nil, "local")
		if err != nil {
			utils.MLogger.Info("CmdPutTo posKM error :", err)
			continue
		}
		utils.MLogger.Info("posKM :", posKM.ToString(), ", posValue :", posCidPrefix)
	}
}

func (p *Info) deletePosBlocks(decreseSpace uint64) {
	utils.MLogger.Info("data is about to exceed the space limit, delete pos blocks")

	posKM, err := metainfo.NewKey(p.localID, mpb.KeyType_PosMeta)
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
		for i := 0; i < rep; i++ {
			blockID := posCidPrefix + metainfo.BlockDelimiter + strconv.Itoa(i)
			err := p.ds.DeleteBlock(p.context, blockID, "local")
			if err != nil {
				utils.MLogger.Info("delete block: ", blockID, " error :", err)
				j++
			} else {
				utils.MLogger.Info("delete block : ", blockID, " success")
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

		err = p.ds.PutKey(p.context, posKM.ToString(), []byte(posCidPrefix), nil, "local")
		if err != nil {
			utils.MLogger.Info("CmdPutTo posKM error :", err)
			continue
		}
		utils.MLogger.Info("after delete ,Gid is: ", curGid, ", sid is: ", curSid, ", cid prefix is: ", posCidPrefix)

		// send BlockMeta deletion to keepers
		//发送元数据到keeper
		km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, posID)
		if err != nil {
			utils.MLogger.Info("construct put blockMeta KV error :", err)
			return
		}

		gp := p.getGroupInfo(posID, groupID, false)
		if gp == nil {
			return
		}
		metavalue := strings.Join(deleteBlocks, metainfo.DELIMITER)
		for _, keeper := range gp.keepers {
			p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Delete), km.ToString(), []byte(metavalue), nil, keeper)
		}
	}
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
