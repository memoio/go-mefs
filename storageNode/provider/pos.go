package provider

import (
	"context"
	"math/big"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/memoio/go-mefs/crypto/pdp"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

const (
	lowWater  = 0.70 // 数据生成为总量的70%
	highWater = 0.85 // 使用量达到85%后，删除10%的块
)

// uid is defined in utils/pos

var curSid = -1
var postID string
var groupID string
var postAddr string
var inGenerate int
var bucketNum int
var bm *metainfo.BlockMeta

var opt = &df.DataCoder{
	Prefix: &mpb.BlockOptions{
		Bopts: &mpb.BucketOptions{
			Version:      1,
			Policy:       df.MulPolicy,
			DataCount:    1,
			ParityCount:  pos.Reps - 1,
			TagFlag:      pdp.PDPV0,
			SegmentSize:  pos.SegSize,
			Encryption:   0,
			SegmentCount: pos.SegCount,
		},
	},
}

// PostService starts post
func (p *Info) PostService(ctx context.Context, gc bool) error {
	// 获取合约地址一次，主要是获取keeper，用于发送block meta
	// handleUserDeployedContracts()
	utils.MLogger.Info("Start Post Service")

	//从磁盘读取存储的Cidprefix
	postKM, err := metainfo.NewKey(p.localID, mpb.KeyType_PosMeta)
	if err != nil {
		utils.MLogger.Debug("NewKeyMeta postKM error :", err)
		return err
	}

	postValue, err := p.ds.GetKey(ctx, postKM.ToString(), "local")
	if err != nil {
		utils.MLogger.Debug("Get postKM from local error :", err)
	} else {
		utils.MLogger.Info("get postKM value: ", string(postValue))
		cidInfo, err := metainfo.NewBlockFromString(string(postValue))
		if err != nil {
			utils.MLogger.Debug("get block meta in postRegular error :", err)
		} else {
			sid, err := strconv.Atoi(cidInfo.GetSid())
			if err != nil {
				utils.MLogger.Debug("strconv.Atoi Sid in postReguar error :", err)
			} else {
				curSid = sid
			}
		}
	}

	utils.MLogger.Info("before traverse post blocks reaches sid: ", curSid)

	p.StoragePostUsed = uint64(pos.DLen * pos.Reps * (curSid + 1))

	postID = pos.GetPostId()
	postAddr = pos.GetPostAddr()

	qItem, err := role.GetLatestQuery(postID)
	if err != nil {
		utils.MLogger.Error("get query of postID err: ", err)
		return err
	}

	groupID = qItem.QueryID

	localNum, err := address.GetNodeIDFromID(p.localID)
	if err != nil {
		utils.MLogger.Error("GetNodeIDFromID err: ", err)
		return err
	}

	bucketNum = int(localNum)

	gp := p.getGroupInfo(postID, groupID, true)
	if gp == nil {
		utils.MLogger.Info("get group of postID ", postID, " and groupID ", groupID, " is nil.")
		return role.ErrEmptyData
	}
	utils.MLogger.Info("status in groupInfo of postID, groupID is: ", gp.status)
	utils.MLogger.Info("storageUsed and storageTotal in groupInfo of postID, groupID is: ", gp.storageUsed, gp.storageTotal)
	utils.MLogger.Info("keepers in groupInfo of postID, groupID is: ", gp.keepers)
	utils.MLogger.Info("providers in groupInfo of postID, groupID is: ", gp.providers)

	km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, postID)
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
			utils.MLogger.Info("Send Post add to keepers:", keeper)
			p.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), km.ToString(), []byte(p.localID), nil, keeper)
		}

		time.Sleep(10 * time.Minute)
		gp.loadContracts(p.localID, true)
	}

	//填充opt.KeySet
	mkey, err := pdp.GenKeySetV1WithSeed(pos.GetPostSeed(), pdp.SCount)
	if err != nil {
		utils.MLogger.Info("Init bls config for post user fail: ", err)
		return err
	}

	p.userConfigs.Add(groupID, mkey)

	opt.BlsKey = mkey

	opt.PreCompute()

	p.traversePath(gc)

	utils.MLogger.Info("after traverse post blocks reaches sid: ", curSid)

	newbm, err := metainfo.NewBlockMeta(groupID, strconv.Itoa(bucketNum), strconv.Itoa(curSid), "0")
	if err != nil {
		return err
	}

	bm = newbm

	// send last one
	metaValue := bm.ToString()
	for _, keeper := range gp.keepers {
		p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(metaValue), nil, keeper)
	}

	//开始post
	p.postRegular(ctx)
	return nil
}

// postRegular checks postBlocks and decide to add/delete
func (p *Info) postRegular(ctx context.Context) {
	utils.MLogger.Info("Post start!")

	p.doGenerateOrDelete()
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if inGenerate == 0 {
				// 如果超过了90%，则删除10%容量的postBlocks；如果低于80%，则生成到80%
				go p.doGenerateOrDelete()
			}
		}
	}
}

func (p *Info) traversePath(gc bool) {
	if gc {
		utils.MLogger.Info("clean post blocks first")
	}

	notfound := 0
	var res strings.Builder
	sid := 0
	for sid = 0; ; sid++ {
		notfound = 0
		for i := 0; i < pos.Reps; i++ {
			res.Reset()
			res.WriteString(groupID)
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(bucketNum))
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(sid))
			res.WriteString(metainfo.BlockDelimiter)
			res.WriteString(strconv.Itoa(i))
			ncid := cid.NewCidV2([]byte(res.String()))
			exist, err := p.ds.BlockStore().Has(ncid)
			if err != nil {
				utils.MLogger.Infof("post has %s failed: %s", res.String(), err)
				notfound++
				continue
			}

			if exist {
				if gc {
					p.ds.DeleteBlock(p.context, res.String(), "local")
				}
			} else {
				notfound++
				utils.MLogger.Infof("post not has %s", res.String())
			}
		}

		if notfound >= pos.Reps {
			break
		}
	}

	if gc {
		curSid = -1
	} else {
		curSid = sid - 1
	}

	p.StoragePostUsed = uint64(pos.DLen * pos.Reps * (curSid + 1))

	bm.SetSid(strconv.Itoa(curSid))
}

func (p *Info) doGenerateOrDelete() {
	inGenerate = 1
	defer func() {
		inGenerate = 0
	}()

	lsinfo, err := role.GetDiskSpaceInfo()
	if err != nil || lsinfo.Total == 0 {
		return
	}

	freeRatio := float64(lsinfo.Free) / float64(lsinfo.Total)

	usedSpace, err := p.getDiskUsage()
	if err != nil {
		return
	}

	totalSpace := p.getDiskTotal()

	ratio := float64(usedSpace) / float64(totalSpace)
	utils.MLogger.Infof("Space used by mefs is: %d, pledge space is: %d, used ratio is: %.4f", usedSpace, totalSpace, ratio)
	utils.MLogger.Infof("Local free space is: %d, local total space is: %d, free ratio is: %.4f", lsinfo.Free, lsinfo.Total, freeRatio)

	if ratio < lowWater && freeRatio > (1-lowWater) {
		generateSpace := uint64(float64(totalSpace) * (lowWater - ratio))
		if generateSpace > uint64(pos.Reps)*uint64(pos.DLen)*30 {
			generateSpace = uint64(pos.Reps) * uint64(pos.DLen) * 30
		}

		if lsinfo.Free < generateSpace {
			utils.MLogger.Infof("Local only has space: %d", lsinfo.Free)
			return
		}

		p.generatePostBlocks(generateSpace)
	} else {
		if ratio > highWater || freeRatio < (1-highWater) {
			p.deletePostBlocks(uint64(usedSpace / 10))
		}
	}

	km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, postID)
	gp := p.getGroupInfo(postID, groupID, false)
	if gp == nil {
		return
	}
	for _, keeper := range gp.keepers {
		p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(bm.ToString()+metainfo.BlockDelimiter+strconv.Itoa(pos.SegCount)), nil, keeper)
	}
}

// generatePostBlocks generate block accoding to the free space
func (p *Info) generatePostBlocks(increaseSpace uint64) {
	utils.MLogger.Infof("generate post blocks for space: %d", increaseSpace)

	postKM, err := metainfo.NewKey(p.localID, mpb.KeyType_PosMeta)
	if err != nil {
		return
	}

	totalIncreased := uint64(0)
	for {
		if totalIncreased >= increaseSpace {
			break
		}
		tmpData := make([]byte, pos.DLen)
		totalIncreased += uint64(10 * len(tmpData))
		rand.Seed(time.Now().UnixNano())
		fillRandom(tmpData)
		curSid++

		bm.SetSid(strconv.Itoa(curSid))
		data, offset, err := opt.Encode(tmpData, bm.ToString(3), 0)
		if err != nil {
			utils.MLogger.Info("UploadMulpolicy in generate Post Blocks error: ", err)
			continue
		}

		blockList := []string{}

		//做成块，放到本地
		for i, dataBlock := range data {
			bm.SetCid(strconv.Itoa(i))
			blockID := bm.ToString()

			err := p.ds.PutBlock(p.context, blockID, dataBlock, "local")
			if err != nil {
				utils.MLogger.Info("add block failed, error :", err)
				continue
			}

			p.StoragePostUsed += uint64(pos.DLen)

			res := strings.SplitAfterN(blockID, metainfo.BlockDelimiter, 2)
			if len(res) != 2 {
				continue
			}

			boff := res[1] + metainfo.BlockDelimiter + strconv.Itoa(offset)

			blockList = append(blockList, boff)
		}

		gp := p.getGroupInfo(postID, groupID, false)
		if gp == nil {
			return
		}

		// 向keeper发送元数据
		metaValue := strings.Join(blockList, metainfo.DELIMITER)
		km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, postID)

		for _, keeper := range gp.keepers {
			p.ds.SendMetaRequest(p.context, int32(mpb.OpType_Put), km.ToString(), []byte(metaValue), nil, keeper)
		}

		// 本地更新
		err = p.ds.PutKey(p.context, postKM.ToString(), []byte(bm.ToString()), nil, "local")
		if err != nil {
			utils.MLogger.Info("CmdPutTo postKM error :", err)
			continue
		}
		utils.MLogger.Info("postKM :", postKM.ToString(), ", postValue :", bm.ToString())
	}
}

func (p *Info) deletePostBlocks(decreseSpace uint64) {
	utils.MLogger.Info("data is about to exceed the space limit, delete post blocks")

	postKM, err := metainfo.NewKey(p.localID, mpb.KeyType_PosMeta)
	if err != nil {
		return
	}

	// delete last blocks
	var totalDecresed uint64
	for {
		if curSid == -1 {
			return
		}

		if totalDecresed >= decreseSpace {
			break
		}
		//删除块
		deleteBlocks := []string{}
		for i := pos.Reps - 1; i >= 0; i-- {
			bm.SetCid(strconv.Itoa(i))
			blockID := bm.ToString()
			err := p.ds.DeleteBlock(p.context, blockID, "local")
			if err != nil {
				utils.MLogger.Info("delete block: ", blockID, " error :", err)
				continue
			}
			utils.MLogger.Info("delete block : ", blockID, " success")
			p.StoragePostUsed -= uint64(pos.DLen)
			totalDecresed += uint64(pos.DLen)
			deleteBlocks = append(deleteBlocks, blockID)
		}

		//更新Gid,Sid
		curSid--

		bm.SetSid(strconv.Itoa(curSid))

		err = p.ds.PutKey(p.context, postKM.ToString(), []byte(bm.ToString()), nil, "local")
		if err != nil {
			utils.MLogger.Info("CmdPutTo postKM error :", err)
			continue
		}
		utils.MLogger.Info("after delete, sid is: ", curSid)

		// send BlockMeta deletion to keepers
		//发送元数据到keeper
		km, err := metainfo.NewKey(groupID, mpb.KeyType_Pos, postID)
		if err != nil {
			utils.MLogger.Info("construct put blockMeta KV error :", err)
			return
		}

		gp := p.getGroupInfo(postID, groupID, false)
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

func getPostPreIncome(ukAddrs []common.Address, localAddr common.Address) *big.Int {
	postPreIncome := big.NewInt(0)
	localID, err := address.GetIDFromAddress(localAddr.Hex())
	if err != nil {
		utils.MLogger.Debug("getIDFromAddress err: ", err, "address: ", localAddr.Hex())
		return postPreIncome
	}

	for _, ukAddr := range ukAddrs {
		ukID, err := address.GetIDFromAddress(ukAddr.Hex())
		if err != nil {
			utils.MLogger.Debug("getIDFromAddress err: ", err, "address: ", ukAddr.Hex())
			continue
		}
		ukItem, err := role.GetUpkeepingInfo(localID, ukID)
		if err != nil {
			utils.MLogger.Debug("GetUpkeepingInfo err: ", err, "localID: ", localID, "ukID: ", ukID)
			continue
		}
		for _, pInfo := range ukItem.Providers {
			if pInfo.Addr.Hex() == localAddr.Hex() {
				postPreIncome.Add(postPreIncome, calculatePreIncome(pInfo.Money, int(pInfo.PayIndex.Int64())))
				break
			}
		}
	}
	return postPreIncome
}

func calculatePreIncome(money []*big.Int, payIndex int) *big.Int {
	count := big.NewInt(0)
	for ; payIndex < len(money); payIndex++ {
		count.Add(count, money[payIndex])
	}
	return count
}
