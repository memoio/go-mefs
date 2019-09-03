package provider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/crypto/aes"
	df "github.com/memoio/go-mefs/data-format"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const (
	LowWater  = 0.8 // 数据生成为总量的80%
	HighWater = 0.9 // 使用量达到90%后，删除10%的块
)

// uid is defined in utils/pos

var skbyte = []byte{179, 233, 48, 97, 94, 148, 140, 7, 78, 102, 169, 48, 136, 124, 152, 101, 76, 69, 210, 14, 38, 15, 176, 227, 73, 41, 135, 17, 170, 138, 242, 69}

//var buckid = 1
var posUid string = "8MGxCuiT75bje883b7uFb6eMrJt5cP"
var curGid int = -10
var curSid int = -1
var inGenerate int

// 因只考虑生成3+2个stripe，故测试Rs时，文件长度不超过3M；测试Mul时，文件长度不超过1M
var Mullen = 1 * 1024 * 1024
var opt = &df.DataEncoder{
	//CidPrefix:   "8MGxCuiT75bje883b7uFb6eMrJt5cP_1_0",
	DataCount:   1,
	ParityCount: 4,
	TagFlag:     df.BLS12,
	SegmentSize: df.DefaultSegmentSize,
	KeySet:      getKeySet(),
}

func getKeySet() *mcl.KeySet {
	mcl.Init(mcl.BLS12_381)
	keyset, _ := mcl.GenKeySet()
	return keyset
}

func posSerivce() {
	// 获取合约地址一次，主要是获取keeper，用于发送block meta
	// handleUserDeployedContracts()
}

// getDiskUsage gets the disk usage
func getDiskUsage() (uint64, error) {
	dataStore := localNode.Repo.Datastore()
	DataSpace, err := ds.DiskUsage(dataStore)
	if err != nil {
		fmt.Println("get disk usage failed :", err)
		return 0, err
	}
	return DataSpace, nil
}

// getDiskUsage gets the disk total space which is set in config
func getDiskTotal() (float64, error) {
	cfg, err := localNode.Repo.Config()
	if err != nil {
		fmt.Println("getDiskTotal error :", err)
		return 0, err
	}
	maxSpaceStr := strings.Replace(cfg.Datastore.StorageMax, "GB", "", 1)
	maxSpaceInGB, err := strconv.ParseFloat(maxSpaceStr, 64)
	//Uint(maxSpaceStr, 10, 64)
	if err != nil {
		fmt.Println("PraseUint maxSpaceStr to maxspace error :", err)
		return 0, err
	}
	maxSpaceInByte := maxSpaceInGB * 1024 * 1024 * 1024
	return maxSpaceInByte, nil
}

// getDiskUsage gets the disk total space which is set in config
func getFreeSpace() {
	return
}

// posRegular checks posBlocks and decide to add/delete
func PosRegular(ctx context.Context) {
	fmt.Println("posRegular() start!")
	posKM, err := metainfo.NewKeyMeta(posUid, metainfo.Local)
	if err != nil {
		fmt.Println("NewKeyMeta posKM error :", err)
	} else {
		fmt.Println("posKm :", posKM.ToString())
		posValue, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(posKM.ToString(), "local")
		if err != nil {
			fmt.Println("CmdGetFrom(posKM, 'local') error :", err)
		} else {
			fmt.Println("posvalue :", string(posValue))
			opt.CidPrefix = string(posValue)
			cidInfo, err := metainfo.GetBlockMeta(string(posValue) + "_0")
			if err != nil {
				fmt.Println("get block meta in posRegular error :", err)
			} else {
				curGid, err = strconv.Atoi(cidInfo.GetGid()[utils.IDLength:])
				if err != nil {
					fmt.Println("strconv.Atoi Gid in posReguar error :", err)
				}
				curSid, err = strconv.Atoi(cidInfo.GetSid())
				if err != nil {
					fmt.Println("strconv.Atoi Sid in posReguar error :", err)
				}
			}
		}
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if inGenerate == 0 {
				go doGenerateOrDelete()
			}
			// 如果超过了90%，则删除10%容量的posBlocks；如果低于80%，则生成到80%
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
	fmt.Println("usedSpace :", usedSpace)
	totalSpace, err := getDiskTotal()
	if err != nil {
		return
	}
	fmt.Println("totalSpace :", totalSpace, "\n(usedSpace)/totalSpace", float64(usedSpace)/totalSpace)
	if float64(usedSpace)/totalSpace <= LowWater {
		generatePosBlocks(uint64(totalSpace / 10))
	} else if float64(usedSpace)/totalSpace >= HighWater {
		deletePosBlocks(uint64(usedSpace / 10))
	}
}

func UploadMulpolicy(data []byte) ([][]byte, int, error) {
	opt.Policy = df.MulPolicy
	// 构建加密秘钥
	buckid := localNode.Identity.Pretty() + strconv.Itoa(curGid)
	tmpkey := []byte(string(skbyte) + buckid)
	skey := sha256.Sum256(tmpkey)
	// 加密、Encode
	data, err := aes.AesEncrypt(data, skey[:])
	if err != nil {
		return nil, 0, err
	}
	encodeData, offset, err := opt.Encode(data, 0)
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
		tmpData := make([]byte, Mullen)
		totalIncreased += uint64(5 * len(tmpData))
		rand.Seed(time.Now().UnixNano())
		fillRandom(tmpData)
		// 配置部分
		//更新stripeID、bucketID
		curSid = (curSid + 1) % 10
		if curSid == 0 {
			curGid += 10
		}
		opt.CidPrefix = posUid + "_" + localNode.Identity.Pretty() + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		data, _, err := UploadMulpolicy(tmpData)
		if err != nil {
			fmt.Println("UploadMulpolicy in generate Pos Blocks error :", err)
			return
		}

		//做成块，放到本地
		for i, dataBlock := range data {
			blockID := opt.CidPrefix + "_" + strconv.Itoa(i)
			ncid := cid.NewCidV2([]byte(blockID))
			newblk, err := blocks.NewBlockWithCid(dataBlock, ncid)
			if err != nil {
				fmt.Println("New block failed, error :", err)
				return
			}
			fmt.Println("New block success :", newblk.Cid())
			err = localNode.Blocks.PutBlock(newblk)
			if err != nil {
				fmt.Println("add block failed, error :", err)
				return
			}

			//向keeper发送元数据
			/*kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.BlockMetaInfo, metainfo.SyncTypeBlock)
			if err != nil {
				fmt.Println("construct put blockMeta KV error :", err)
				return
			}
			metaValue := localNode.Identity.Pretty() + metainfo.DELIMITER + strconv.Itoa(offset)*/
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
		for i := 0; i < 5; i++ {
			blockID := opt.CidPrefix + "_" + strconv.Itoa(i)
			ncid := cid.NewCidV2([]byte(blockID))
			err := localNode.Blockstore.DeleteBlock(ncid)
			if err != nil {
				fmt.Println("delete block in func deletePosBlocks() error :", err, " blockID :", blockID)
				return
			}
			fmt.Println("delete block :", blockID, "success")
			totalDecresed += uint64(5 * Mullen)

			// send BlockMeta deletion to keepers
			//发送元数据到keeper
			/*kmBlock, err := metainfo.NewKeyMeta(blockID, metainfo.DeleteBlock)
			if err != nil {
				fmt.Println("construct put blockMeta KV error :", err)
				return
			}*/
		}
		//跟新Gid,Sid
		fmt.Println("current gid :", curGid, "\n cursid :", curSid)
		curSid = (curSid + 9) % 10
		if curSid == 9 {
			curGid -= 10
		}
		opt.CidPrefix = posUid + "_" + localNode.Identity.Pretty() + strconv.Itoa(curGid) + "_" + strconv.Itoa(curSid)
		fmt.Println("after delete ,Gid :", curGid, "\n sid :", curSid, "\ncid prefix :", opt.CidPrefix)
	}
}

func getUserConifg() {
	// 需要用私钥decode出bls的私钥，用user中的方法
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
