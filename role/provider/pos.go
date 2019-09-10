package provider

import (
	"context"
	"crypto/sha256"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/crypto/aes"
	df "github.com/memoio/go-mefs/data-format"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"github.com/memoio/go-mefs/utils/pos"
)

const (
	LowWater  = 0.8 // 数据生成为总量的80%
	HighWater = 0.9 // 使用量达到90%后，删除10%的块
)

// uid is defined in utils/pos

var curGid int = -1024
var curSid int = -1
var posID string
var posAddr string
var posCidPrefix string
var inGenerate int
var keeperIDs []string
var posSkByte []byte

// 因只考虑生成3+2个stripe，故测试Rs时，文件长度不超过3M；测试Mul时，文件长度不超过1M
const (
	mullen = 100 * 1024 * 1024
	offset = 25599
)

var opt = &df.DataEncoder{
	DataCount:   1,
	ParityCount: 4,
	TagFlag:     df.BLS12,
	SegmentSize: df.DefaultSegmentSize,
}

func PosSerivce() {
	// 获取合约地址一次，主要是获取keeper，用于发送block meta
	// handleUserDeployedContracts()
	posID = pos.GetPosId()
	posAddr = pos.GetPosAddr()
	posSkByte = pos.GetPosSkByte()

	retryCount := 0
	for {
		if retryCount > 10 {
			log.Println("Save upkeeping in posService error, exit from pos mode.")
			return
		}
		err := SaveUpkeeping(posID)
		if err == nil {
			break
		}
		retryCount++
	}

	if value, ok := ProContracts.upKeepingBook.Load(posID); ok {
		keeperIDs = value.(contracts.UpKeepingItem).KeeperIDs
	}

	//填充opt.KeySet
	getConfig := false

	for _, tmpKeeper := range keeperIDs {
		if err := getUserConifg(posID, tmpKeeper); err == nil {
			getConfig = true
			break
		}
	}

	if !getConfig {
		log.Println("Cannot get userconfig, start pos fails")
		return
	}

	//从磁盘读取存储的Cidprefix
	posKM, err := metainfo.NewKeyMeta(posID, metainfo.PosMeta)
	if err != nil {
		log.Println("NewKeyMeta posKM error :", err)
	} else {
		log.Println("posKm :", posKM.ToString())
		posValue, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(posKM.ToString(), "local")
		if err != nil {
			log.Println("Get posKM from local error :", err)
		} else {
			log.Println("posvalue :", string(posValue))
			posCidPrefix = string(posValue)
			cidInfo, err := metainfo.GetBlockMeta(string(posValue) + "_0")
			if err != nil {
				log.Println("get block meta in posRegular error :", err)
			} else {
				curGid, err = strconv.Atoi(cidInfo.GetGid()[utils.IDLength:])
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

	//开始pos
	posRegular(context.Background())
}

// getDiskUsage gets the disk usage
func getDiskUsage() (uint64, error) {
	dataStore := localNode.Repo.Datastore()
	DataSpace, err := ds.DiskUsage(dataStore)
	if err != nil {
		log.Println("get disk usage failed :", err)
		return 0, err
	}
	return DataSpace, nil
}

// getDiskTotal gets the disk total space which is set in config
func getDiskTotal() (float64, error) {
	cfg, err := localNode.Repo.Config()
	if err != nil {
		log.Println("getDiskTotal error :", err)
		return 0, err
	}
	maxSpaceStr := strings.Replace(cfg.Datastore.StorageMax, "GB", "", 1)
	maxSpaceInGB, err := strconv.ParseFloat(maxSpaceStr, 64)
	if err != nil {
		log.Println("PraseUint maxSpaceStr to maxspace error :", err)
		return 0, err
	}

	if maxSpaceInGB == 0 {
		return 0, errors.New("max space is zero")
	}

	maxSpaceInByte := maxSpaceInGB * 1024 * 1024 * 1024
	return maxSpaceInByte, nil
}

// getDiskUsage gets the disk total space which is set in config
func getFreeSpace() {
	return
}

// posRegular checks posBlocks and decide to add/delete
func posRegular(ctx context.Context) {
	log.Println("posRegular() start!")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if inGenerate == 0 {
				// 如果超过了90%，则删除10%容量的posBlocks；如果低于80%，则生成到80%
				go doGenerateOrDelete()
			}
		}
	}
}

func doGenerateOrDelete() {
	inGenerate = 1
	defer func() {
		inGenerate = 0
	}()
	usedSpace, err := getDiskUsage()
	if err != nil {
		return
	}
	totalSpace, err := getDiskTotal()
	if err != nil {
		return
	}
	ratio := float64(usedSpace) / totalSpace
	log.Println("usedSpace is: ", usedSpace, ", totalSpace is: ", totalSpace, ",(usedSpace)/totalSpace is: ", ratio)
	if ratio <= LowWater {
		generatePosBlocks(uint64(totalSpace / 10))
	} else if ratio >= HighWater {
		deletePosBlocks(uint64(usedSpace / 10))
	}
}

func UploadMulpolicy(data []byte) ([][]byte, int, error) {
	opt.Policy = df.MulPolicy
	// 构建加密秘钥
	buckid := localNode.Identity.Pretty() + strconv.Itoa(curGid)
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
func generatePosBlocks(increaseSpace uint64) {
	// fillRandom()
	// DataEncodeToMul()
	// send BlockMeta to keepers
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
		//更新stripeID、bucketID
		curSid = (curSid + 1) % 1024
		if curSid == 0 {
			curGid += 1024
		}
		posCidPrefix = posID + "_" + localNode.Identity.Pretty() + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		data, _, err := UploadMulpolicy(tmpData)
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
			err = localNode.Blocks.PutBlock(newblk)
			if err != nil {
				log.Println("add block failed, error :", err)
			}

			boff := blockID + "_" + strconv.Itoa(offset)

			blockList = append(blockList, boff)
		}

		// 向keeper发送元数据
		metaValue := strings.Join(blockList, metainfo.DELIMITER)
		km, err := metainfo.NewKeyMeta(localNode.Identity.Pretty(), metainfo.PosAdd)
		for _, keeper := range keeperIDs {
			sendMetaRequest(km, metaValue, keeper)
		}

		// 本地更新
		posKM, err := metainfo.NewKeyMeta(posID, metainfo.PosMeta)
		if err != nil {
			continue
		}
		posValue := posCidPrefix
		log.Println("posKM :", posKM.ToString(), ", posValue :", posValue)
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(posKM.ToString(), posValue, "local")
		if err != nil {
			log.Println("CmdPutTo posKM error :", err)
			continue
		}
	}
}

func deletePosBlocks(decreseSpace uint64) {
	// delete last blocks
	var totalDecresed uint64

	for {
		if totalDecresed >= decreseSpace {
			break
		}
		//删除块
		deleteBlocks := []string{}
		for i := 0; i < 5; i++ {
			blockID := posCidPrefix + "_" + strconv.Itoa(i)
			ncid := cid.NewCidV2([]byte(blockID))
			err := localNode.Blockstore.DeleteBlock(ncid)
			if err != nil {
				log.Println("delete block in func deletePosBlocks() error :", err, " blockID :", blockID)
				return
			}
			log.Println("delete block : ", blockID, " success")
			totalDecresed += uint64(5 * mullen)

			deleteBlocks = append(deleteBlocks, blockID)
		}
		// send BlockMeta deletion to keepers
		//发送元数据到keeper
		km, err := metainfo.NewKeyMeta(localNode.Identity.Pretty(), metainfo.PosDelete)
		if err != nil {
			log.Println("construct put blockMeta KV error :", err)
			return
		}
		metavalue := strings.Join(deleteBlocks, metainfo.DELIMITER)
		for _, keeper := range keeperIDs {
			sendMetaRequest(km, metavalue, keeper)
		}
		//更新Gid,Sid
		curSid = (curSid + 1023) % 1024
		if curSid == 1023 {
			curGid -= 1024
		}
		posCidPrefix = posID + "_" + localNode.Identity.Pretty() + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		log.Println("after delete ,Gid :", curGid, "\n sid :", curSid, "\ncid prefix :", posCidPrefix)
	}
}

func getUserConifg(userID, keeperID string) error {
	// 需要用私钥decode出bls的私钥，用user中的方法
	//获取公钥
	opt.KeySet = new(mcl.KeySet)
	tmpUserBls12Config, err := getNewUserConfig(userID, keeperID)
	if err != nil {
		log.Println("getNewUserConfig in get userconfig error :", err)
		return err
	}

	usersConfigs.Store(userID, tmpUserBls12Config.PubKey)

	opt.KeySet.Pk = tmpUserBls12Config.PubKey

	//获取私钥
	opt.KeySet.Sk, err = getUserPrivateKey(userID, keeperID)
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
