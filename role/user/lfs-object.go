package user

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	files "github.com/ipfs/go-ipfs-files"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (lfs *LfsService) DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	var bucket *pb.BucketInfo
	var ok bool
	var object *pb.ObjectInfo
	if lfs.CurrentLog.BucketByName == nil {
		return nil, ErrObjectNotExist
	}
	if bucket, ok = lfs.CurrentLog.BucketByName[bucketName]; !ok || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	// TODO:具体实现
	lfs.CurrentLog.State[bucket.BucketID].Mu.Lock()
	defer lfs.CurrentLog.State[bucket.BucketID].Mu.Unlock()
	if lfs.CurrentLog.Entries == nil {
		return nil, ErrObjectNotExist
	}
	objectMap, ok := lfs.CurrentLog.Entries[bucket.BucketID]
	if !ok || objectMap == nil {
		return nil, ErrObjectNotExist
	}
	object, ok = objectMap[objectName]
	if !ok || object == nil || object.Deletion {
		return nil, ErrObjectNotExist
	}
	lfs.CurrentLog.Entries[bucket.BucketID][objectName].Deletion = true
	lfs.CurrentLog.State[bucket.BucketID].Dirty = true
	return object, nil
}

func (lfs *LfsService) HeadObject(bucketName, objectName string) (*pb.ObjectInfo, string, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, "", err
	}
	if lfs.CurrentLog.BucketByName == nil {
		return nil, "", ErrObjectNotExist
	}
	if bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; ok && !bucket.Deletion {
		if objectMap, ok := lfs.CurrentLog.Entries[bucket.BucketID]; ok {
			if object := objectMap[objectName]; object != nil {
				AvailTime, _ := lfs.GetObjectAvailTime(object)
				return object, AvailTime, nil
			}
			return nil, "", ErrObjectNotExist
		}
		return nil, "", ErrBucketNotExist
	}
	return nil, "", ErrBucketNotExist
}

func (lfs *LfsService) ListObject(bucketName, pre string, avail bool) ([]*pb.ObjectInfo, []string, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, nil, err
	}
	if lfs.CurrentLog.BucketByName == nil {
		return nil, nil, ErrObjectNotExist
	}
	var objects []*pb.ObjectInfo
	var availTimes []string
	if Bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; ok && !Bucket.Deletion {
		for _, Object := range lfs.CurrentLog.Entries[Bucket.BucketID] {
			if len(objects) > MAXLISTVALUE { //返回不要过多，应指定好过滤条件
				break
			}
			if Object.Deletion {
				continue
			}
			if avail {
				if strings.HasPrefix(Object.ObjectName, pre) {
					objects = append(objects, Object)
					availTime, _ := lfs.GetObjectAvailTime(Object)
					availTimes = append(availTimes, availTime)
				}
			} else {
				if strings.HasPrefix(Object.ObjectName, pre) {
					objects = append(objects, Object)
				}
			}
		}
		return objects, availTimes, nil
	}
	return nil, nil, ErrBucketNotExist
}

func (lfs *LfsService) ShowStorageSpace(bucketName, pre string) (int, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return 0, err
	}
	if lfs.CurrentLog.BucketByName == nil {
		return 0, ErrBucketNotExist
	}
	if lfs.CurrentLog.Entries == nil {
		return 0, ErrObjectNotExist
	}

	var storageSpace int
	if Bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; ok && !Bucket.Deletion {
		for _, Object := range lfs.CurrentLog.Entries[Bucket.BucketID] {
			if Object.Deletion {
				continue
			}
			storageSpace += int(Object.GetObjectSize())
		}
		return storageSpace, nil
	}
	return 0, ErrBucketNotExist
}

type Upload struct { //一个上传任务实例
	LfsService  *LfsService
	Object      *pb.ObjectInfo
	BucketID    int32
	ObjectName  string
	Reader      io.Reader
	NamePrefix  string
	startTime   time.Time
	segmentSize int32
	TagFlag     int32
	keepers     []string
}

func (lfs *LfsService) ConstructUpload(objectName, namePrefix, bucketName string, file files.Node) (*Upload, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketByName == nil {
		return nil, ErrObjectNotExist
	}
	if Bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; ok && !Bucket.Deletion {
		if object, ok := lfs.CurrentLog.Entries[Bucket.BucketID][namePrefix+objectName]; ok || object != nil {
			return nil, ErrObjectAlreadyExist
		}
		var err error
		gp := GetGroupService(lfs.UserID)
		_, conkeepers, err := gp.GetKeepers(gp.keeperSLA)
		if err != nil {
			return nil, err
		}
		var fileNext files.File
		switch f := file.(type) {
		case files.Directory:
			return nil, errors.New("unsupported now")
		case files.File:
			fileNext = f
		}
		ul := &Upload{
			LfsService:  lfs,
			TagFlag:     int32(Bucket.GetTagFlag()),
			segmentSize: int32(Bucket.GetSegmentSize()),
			startTime:   time.Now(),
			NamePrefix:  namePrefix,
			Reader:      fileNext,
			BucketID:    Bucket.BucketID,
			keepers:     conkeepers,
			ObjectName:  objectName,
		}
		return ul, nil
	}
	return nil, ErrBucketNotExist

}

//上传文件
func (ul *Upload) PutObject(ctx context.Context) (*pb.ObjectInfo, error) {
	if ul == nil {
		return nil, ErrLfsIsNotRunning
	}
	if err := checkObjectName(ul.ObjectName); err != nil {
		return nil, err
	}
	ObjectName := ul.NamePrefix + ul.ObjectName
	if len(ul.ObjectName) > maxObjectNameLen {
		return nil, ErrObjectNameToolong
	}

	ul.LfsService.CurrentLog.State[ul.BucketID].Mu.Lock()
	defer ul.LfsService.CurrentLog.State[ul.BucketID].Mu.Unlock()
	if object, ok := ul.LfsService.CurrentLog.Entries[ul.BucketID][ObjectName]; ok || object != nil {
		return nil, ErrObjectAlreadyExist
	}

	object := &pb.ObjectInfo{
		ObjectName:  ObjectName,
		BucketID:    ul.BucketID,
		Ctime:       time.Now().Format(utils.BASETIME),
		StripeStart: ul.LfsService.CurrentLog.BucketByID[ul.BucketID].CurStripe,
		OffsetStart: ul.LfsService.CurrentLog.BucketByID[ul.BucketID].NextOffset,
		Deletion:    false,
		Dir:         false,
	}
	ul.Object = object
	ul.LfsService.CurrentLog.Entries[ul.BucketID][ObjectName] = object
	bucketInfo := ul.LfsService.CurrentLog.BucketByID[ul.BucketID]
	if bucketInfo.Policy != dataformat.RsPolicy && bucketInfo.Policy != dataformat.MulPolicy {
		return nil, ErrPolicy
	}
	encodeOpt := dataformat.NewDataformat(bucketInfo.Policy, "", bucketInfo.DataCount, bucketInfo.ParityCount,
		int32(bucketInfo.TagFlag), bucketInfo.NextOffset, bucketInfo.SegmentSize, GetGroupService(ul.LfsService.UserID).GetKeyset())
	err := ul.putObject(ctx, encodeOpt)
	if err != nil {
		ul.LfsService.CurrentLog.State[ul.BucketID].Dirty = true //需要记录，可能上传一部分然后失败，空间已占用
		return nil, err
	}
	ul.LfsService.CurrentLog.State[ul.BucketID].Dirty = true
	return object, nil
}

// 具体实现
func (ul *Upload) putObject(ctx context.Context, encodeOpt dataformat.DataformatOption) error {
	dataCount := encodeOpt.DataCount
	parityCount := encodeOpt.ParityCount
	blockCount := dataCount + parityCount
	var readByte int32
	// true为纠删
	switch encodeOpt.Policy {
	case dataformat.RsPolicy:
		readByte = SegementCount * ul.segmentSize * dataCount //每一次读取的数据，尽量读一个整的
	case dataformat.MulPolicy:
		readByte = SegementCount * ul.segmentSize //每一次读取的数据，尽量读一个整的，多副本只读一个
	default:
		return ErrPolicy
	}
	var breakFlag = false
	h := md5.New()
	for { //循环上传每一个块
		encodeOpt.BeginOffset = ul.LfsService.CurrentLog.BucketByID[ul.BucketID].NextOffset
		curOffset := encodeOpt.BeginOffset
		var data []byte
		switch encodeOpt.Policy {
		case dataformat.RsPolicy:
			if curOffset == 0 { //不为零则为追加模式
				data = make([]byte, readByte)
			} else {
				data = make([]byte, readByte-curOffset*ul.segmentSize*dataCount)
			}
		case dataformat.MulPolicy:
			if curOffset == 0 { //不为零则为追加模式
				data = make([]byte, readByte)
			} else {
				data = make([]byte, readByte-curOffset*ul.segmentSize)
			}
		default:
		}
		tempData := data
		count := 0
		for {
			n, err := ul.Reader.Read(tempData[count:]) //IPFS提供的读取固定一次读4k(方便IPLD)，下次可以找到去除限制
			count += n
			if err != nil && err != io.EOF {
				return err
			} else if err == io.EOF {
				breakFlag = true
				break
			} else if count == len(data) {
				break
			}
		}
		data = data[:count]
		// 对整个文件的数据进行MD5校验
		h.Write(data)
		// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
		tmpkey := ul.LfsService.PrivateKey
		// ul.LfsService.CurrentLog.node.PrivateKey.Bytes()
		tmpkey = append(tmpkey, byte(ul.BucketID))
		skey := sha256.Sum256(tmpkey)
		// ObjectSize记录明文数据长度，解密时也根据ObjectSize计算去除padding
		ul.Object.ObjectSize += int32(len(data))
		if len(data)%aes.BlockSize != 0 {
			data = aes.PKCS5Padding(data)
		}
		data, err := aes.AesEncrypt(data, skey[:])
		if err != nil {
			return err
		}
		bm, err := metainfo.NewBlockMeta(ul.LfsService.UserID, strconv.Itoa(int(ul.BucketID)), strconv.Itoa(int(ul.LfsService.CurrentLog.BucketByID[ul.BucketID].CurStripe)), "0")
		if err != nil {
			return err
		}
		encodeOpt.CidPrefix = bm.ToString(3)
		encodedData, offset, err := encodeOpt.Encode(data)
		if err != nil {
			return err
		}
		if curOffset == 0 {
			providers, _ := GetGroupService(ul.LfsService.UserID).GetProviders(int(blockCount))
			if providers == nil || len(providers) < int(dataCount+parityCount/2) {
				return ErrNoEnoughProvider
			}
			if len(providers) > int(blockCount) {
				providers = providers[:blockCount]
			}
			var i int
			var count int
			for ; i < len(providers); i++ {
				bm.SetBid(strconv.Itoa(i))
				ncid := bm.ToString()
				km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(int(offset)))
				updateKey := km.ToString()
				bcid := cid.NewCidV2([]byte(updateKey))
				b, err := blocks.NewBlockWithCid(encodedData[i], bcid)
				if err != nil {
					return err
				}
				err = localNode.Blocks.PutBlockTo(b, providers[i])
				if err != nil {
					fmt.Println("Put Block", ncid, "to", providers[i], "failed:", err)
				} else {
					key, err := metainfo.NewKeyMeta(ncid, metainfo.HasBlock)
					if err != nil {
						fmt.Println("Put Block", ncid, "to", providers[i], "failed:", err)
						continue
					}
					MaxCount := 20
					flag := 0
					for {
						if flag >= MaxCount {
							fmt.Println("Put Block", ncid, "to", providers[i], "failed:", err)
							break
						}
						flag++
						time.Sleep(500 * time.Millisecond)
						offByte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(key.ToString(), providers[i])
						if err != nil {
							continue
						}
						off, err := strconv.Atoi(string(offByte))
						if err != nil {
							fmt.Println(err)
						}
						if off == offset {
							count++
							break
						}
					}
				}
				err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, providers[i], offset)
				if err != nil {
					return err
				}
			}
			if count < int(dataCount+parityCount/2) {
				return ErrNoEnoughProvider
			}
			for ; i < int(blockCount); i++ {
				bm.SetBid(strconv.Itoa(i))
				ncid := bm.ToString()
				err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, ul.LfsService.UserID, offset)
				if err != nil {
					return err
				}
			}
		} else {
			var i int
			var count int
			for ; i < int(blockCount); i++ {
				bm.SetBid(strconv.Itoa(i))
				ncid := bm.ToString()
				provider, _, err := GetGroupService(ul.LfsService.UserID).GetBlockProviders(ncid)
				if err != nil || provider == ul.LfsService.UserID {
					fmt.Println("Append Block to", provider, "failed:", err)
					if _, err := peer.IDB58Decode(provider); provider != "" && err == nil {
						err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, provider, offset)
						if err != nil {
							return err
						}
					} else {
						err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, ul.LfsService.UserID, offset)
						if err != nil {
							return err
						}
					}
					continue
				}

				km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "append", strconv.Itoa(int(curOffset)), strconv.Itoa(int(offset)))
				appendKey := km.ToString()
				bcid := cid.NewCidV2([]byte(appendKey))
				b, err := blocks.NewBlockWithCid(encodedData[i], bcid)
				if err != nil {
					return err
				}
				err = localNode.Blocks.PutBlockTo(b, provider)
				if err != nil {
					fmt.Println("Append Block", ncid, "to", provider, "failed:", err)
					err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, provider, offset)
					if err != nil {
						return err
					}
					continue
				} else {
					key, err := metainfo.NewKeyMeta(ncid, metainfo.HasBlock)
					if err != nil {
						continue
					}

					MaxCount := 20
					flag := 0
					for {
						if flag >= MaxCount {
							fmt.Println("Append Block", ncid, "to", provider, "failed:", err)
							break
						}
						flag++
						time.Sleep(500 * time.Millisecond)
						offByte, err := localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(key.ToString(), provider)
						if err != nil {
							continue
						}
						off, err := strconv.Atoi(string(offByte))
						if err != nil {
							fmt.Println(err)
						}
						if off == offset {
							count++
							break
						}
					}
				}
				err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, provider, offset)
				if err != nil {
					return err
				}
			}
			if count < int(dataCount+parityCount/2) {
				return ErrNoEnoughProvider
			}
		}
		if offset >= int(SegementCount-1) { //如果写满了一个stripe
			ul.LfsService.CurrentLog.BucketByID[ul.BucketID].CurStripe++
			ul.LfsService.CurrentLog.BucketByID[ul.BucketID].NextOffset = 0
		} else {
			ul.LfsService.CurrentLog.BucketByID[ul.BucketID].NextOffset = int32(offset + 1)
		}
		if breakFlag {
			ul.Object.ETag = hex.EncodeToString(h.Sum(nil))
			break
		}
	}
	return nil
}

type Download struct { //一个下载任务实例
	LfsService         *LfsService
	BucketID           int32
	Object             *pb.ObjectInfo
	sizeReceived       int32 //可以统计下载进度
	startTime          time.Time
	pipeReader         io.Reader
	pipeWriter         io.Writer
	closePipeWithError func(error) bool
}

func (lfs *LfsService) ConstructDownload(bucketName, objectName string) (*Download, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; !ok || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	BucketID := lfs.CurrentLog.BucketByName[bucketName].BucketID
	var object *pb.ObjectInfo
	if object = lfs.CurrentLog.Entries[BucketID][objectName]; object == nil || object.Deletion {
		return nil, ErrObjectNotExist
	}
	piper, pipew := io.Pipe()
	//bufw := bufio.NewWriterSize(pipew, DefaultBufSize)

	checkErrAndClosePipe := func(err error) bool {
		if err != nil {
			err = pipew.CloseWithError(err)
			if err != nil {
				return false
			}
			return true
		}
		err = pipew.Close()
		if err != nil {
			return false
		}
		return false
	}

	dl := &Download{
		LfsService:         lfs,
		Object:             object,
		BucketID:           BucketID,
		startTime:          time.Now(),
		pipeReader:         piper,
		pipeWriter:         pipew,
		closePipeWithError: checkErrAndClosePipe,
	}
	return dl, nil
}

func (dl *Download) GetObject(ctx context.Context) (io.Reader, error) {
	switch dl.LfsService.CurrentLog.BucketByID[dl.BucketID].Policy {
	case dataformat.RsPolicy:
		go func() {
			err := dl.getObjectWithEC(ctx)
			if err != nil {
				fmt.Println("dl.getObjectWithEC(ctx) failed ", err)
			}
		}()
	case dataformat.MulPolicy:
		go func() {
			err := dl.getObjectWithMultireplic(ctx)
			if err != nil {
				fmt.Println("dl.getObjectWithMultireplic(ctx) failed ", err)
			}
		}()
	default:
		return nil, ErrPolicy
	}
	return dl.pipeReader, nil
}

func (dl *Download) getObjectWithEC(ctx context.Context) error {
	bucket := dl.LfsService.CurrentLog.BucketByID[dl.BucketID]
	dataCount := bucket.DataCount
	parityCount := bucket.ParityCount
	blockCount := dataCount + parityCount
	stripeID := dl.Object.StripeStart
	offsetStart := dl.Object.OffsetStart
	// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
	tmpkey := dl.LfsService.PrivateKey
	tmpkey = append(tmpkey, byte(dl.BucketID))
	skey := sha256.Sum256(tmpkey)
	cfg, err := localNode.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed.")
		return err
	}

	for {
		bm, err := metainfo.NewBlockMeta(dl.LfsService.UserID, strconv.Itoa(int(dl.BucketID)), strconv.Itoa(int(stripeID)), "")
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		datas := make([][]byte, blockCount)
		var tempReceiveBlockCount int32
		var needRepair = true //是否需要修复
		var data []byte
		for i := 0; i < int(blockCount); i++ {
			bm.SetBid(strconv.Itoa(i))
			ncid := bm.ToString()
			provider, _, err := GetGroupService(dl.LfsService.UserID).GetBlockProviders(ncid)
			if err != nil || provider == dl.LfsService.UserID {
				log.Printf("Get Block %s's provider from keeper failed, Err: %v\n", ncid, err)
				continue
			}

			//user给channel合约签名，发给provider
			mes, money, err := dl.getMessage(ncid, provider)
			if err != nil {
				continue
			}

			b, err := localNode.Blocks.GetBlockFrom(ctx, provider, ncid, DefaultGetBlockDelay, mes)
			if err != nil {
				log.Printf("Get Block %s from %s failed, Err: %v\n", ncid, provider, err)
				continue
			}
			err = localNode.Blocks.DeleteBlock(cid.NewCidV2([]byte(ncid)))
			if err != nil {
				log.Println("Delete block", ncid, "failed:", err)
			}
			blkData := b.RawData()
			//需要检查数据块的长度也没问题
			dif := dl.Object.ObjectSize - dl.sizeReceived
			ok, err := dataformat.VerifyBlockLength(blkData, int(offsetStart), int(bucket.TagFlag), int(bucket.SegmentSize), int(dataCount), int(parityCount), int(dif), bucket.Policy)
			if !ok {
				log.Printf("Block %s from %s offset unmatched, Err: %v\n", ncid, provider, err)
				continue
			}

			if ok := dataformat.VerifyBlock(blkData, ncid, GetGroupService(dl.LfsService.UserID).GetKeyset().Pk); !ok || err != nil {
				log.Println("Verify Block failed.", ncid, "from:", provider)
				continue
			}

			if !cfg.Test {
				//下载数据成功，将内存的channel的value更改
				cs := GetContractService(dl.LfsService.UserID)
				if cs == nil {
					dl.closePipeWithError(err)
					return ErrUserNotExist
				}
				channel, err := cs.GetChannelItem(provider)
				if err != nil {
					dl.closePipeWithError(err)
					return err
				}
				fmt.Println("下载成功，更改内存中channel.value", channel.ChannelAddr, money.String())
				channel.Value = money
				cs.channelBook[provider] = channel
			}
			datas[i] = blkData
			tempReceiveBlockCount++
			if tempReceiveBlockCount >= dataCount {
				if i == int(dataCount)-1 {
					needRepair = false
				}
				break
			}
		}
		if tempReceiveBlockCount < dataCount {
			dl.closePipeWithError(ErrCannotGetEnoughBlock)
			return ErrCannotGetEnoughBlock
		}
		if needRepair {
			recoveredData, err := dataformat.RecoverData(datas, int(dataCount), int(parityCount), -1)
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
			data, err = dataformat.GetFileDataFromSripe(recoveredData, int(dataCount), int(offsetStart), -1)
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
		} else {
			data, err = dataformat.GetFileDataFromSripe(datas, int(dataCount), int(offsetStart), -1)
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
		}

		if dl.sizeReceived+int32(len(data)) >= dl.Object.ObjectSize {
			// 先解密，再去padding
			padding := aes.BlockSize - ((dl.Object.ObjectSize-1)%aes.BlockSize + 1)
			data = data[:dl.Object.ObjectSize-dl.sizeReceived+padding] //此处时因为获取的为整块，而文件所需只占里面一部分
			data, err = aes.AesDecrypt(data, skey[:])
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
			data = data[:len(data)-int(padding)]
			written, err := dl.pipeWriter.Write(data)
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
			dl.sizeReceived += int32(written)
			break
		}

		data, err = aes.AesDecrypt(data, skey[:])
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		written, err := dl.pipeWriter.Write(data)
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		dl.sizeReceived += int32(written)
		stripeID++
		offsetStart = 0
	}
	dl.closePipeWithError(nil)
	return nil
}

func (dl *Download) getObjectWithMultireplic(ctx context.Context) error {
	bucket := dl.LfsService.CurrentLog.BucketByID[dl.BucketID]
	dataCount := bucket.DataCount
	parityCount := bucket.ParityCount
	blockCount := dataCount + parityCount
	stripeID := dl.Object.StripeStart
	offsetStart := dl.Object.OffsetStart
	// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
	tmpkey := dl.LfsService.PrivateKey
	tmpkey = append(tmpkey, byte(dl.BucketID))
	skey := sha256.Sum256(tmpkey)
	node := localNode
	cfg, err := node.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed.")
		return err
	}
	for {
		bm, err := metainfo.NewBlockMeta(dl.LfsService.UserID, strconv.Itoa(int(dl.BucketID)), strconv.Itoa(int(stripeID)), "")
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		var blkData []byte
		var flag int
		for i := 0; i < int(blockCount); i++ {
			bm.SetBid(strconv.Itoa(i))
			ncid := bm.ToString()
			provider, _, err := GetGroupService(dl.LfsService.UserID).GetBlockProviders(ncid)
			if err != nil || provider == dl.LfsService.UserID {
				log.Printf("Get Block %s's provider from keeper failed.\n", ncid)
			}

			//user给channel合约签名，发给provider
			mes, money, err := dl.getMessage(ncid, provider)
			if err != nil {
				continue
			}

			b, err := localNode.Blocks.GetBlockFrom(ctx, provider, ncid, DefaultGetBlockDelay, mes)
			if b != nil && err == nil {
				blkData = b.RawData()
				err := localNode.Blocks.DeleteBlock(cid.NewCidV2([]byte(ncid)))
				if err != nil {
					log.Println("Delete block", ncid, "failed:", err)
				}
				//需要检查数据块的长度也没问题
				dif := dl.Object.ObjectSize - dl.sizeReceived
				ok, err := dataformat.VerifyBlockLength(blkData, int(offsetStart), int(bucket.TagFlag), int(bucket.SegmentSize), int(dataCount), int(parityCount), int(dif), bucket.Policy)
				if !ok {
					log.Printf("Block %s from %s offset unmatched, Err: %v\n", ncid, provider, err)
					continue
				}
				if ok := dataformat.VerifyBlock(blkData, ncid, GetGroupService(dl.LfsService.UserID).GetKeyset().Pk); !ok || err != nil {
					fmt.Println("Verify Block failed.", ncid, "from:", provider)
					continue
				}

				if !cfg.Test {
					// 下载数据成功，更新内存channel合约的value值
					cs := GetContractService(dl.LfsService.UserID)
					if cs == nil {
						dl.closePipeWithError(err)
						return ErrUserNotExist
					}
					channel, err := cs.GetChannelItem(provider)
					if err != nil {
						dl.closePipeWithError(err)
						return err
					}
					fmt.Println("下载成功，更改内存中channel.value", channel.ChannelAddr, money.String())
					channel.Value = money
					cs.channelBook[provider] = channel
				}
				flag++
				break
			}
		}
		if flag < 1 {
			dl.closePipeWithError(ErrCannotGetEnoughBlock)
			return ErrCannotGetEnoughBlock
		}
		data, err := dataformat.GetSegsFromData(blkData, int(offsetStart), -1)
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		if dl.sizeReceived+int32(len(data)) >= dl.Object.ObjectSize {
			// 先解密，再去padding
			padding := aes.BlockSize - ((dl.Object.ObjectSize-1)%aes.BlockSize + 1)
			data = data[:dl.Object.ObjectSize-dl.sizeReceived+padding]
			data, err = aes.AesDecrypt(data, skey[:])
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
			data = data[:len(data)-int(padding)]
			written, err := dl.pipeWriter.Write(data)
			if err != nil {
				dl.closePipeWithError(err)
				return err
			}
			dl.sizeReceived += int32(written)
			break
		}

		data, err = aes.AesDecrypt(data, skey[:])
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		written, err := dl.pipeWriter.Write(data)
		if err != nil {
			dl.closePipeWithError(err)
			return err
		}
		dl.sizeReceived += int32(written)
		offsetStart = 0
		stripeID++
	}
	dl.closePipeWithError(nil)
	return nil
}

func (lfs *LfsService) getLastChalTime(blockID string) (time.Time, error) {
	latestTime := time.Unix(0, 0)
	gp := GetGroupService(lfs.UserID)
	_, conkeepers, err := gp.GetKeepers(-1)
	if err != nil {
		return latestTime, err
	}
	if len(conkeepers) == 0 {
		return latestTime, ErrNoKeepers
	}

	km, err := metainfo.NewKeyMeta(blockID, metainfo.Query, metainfo.QueryTypeLastChal)
	if err != nil {
		return latestTime, err
	}
	var res string
	var tempTime time.Time
	for _, keeper := range conkeepers {
		res, err = sendMetaRequest(km, "", keeper)
		if err != nil {
			continue
		}
		unixTime := utils.StringToUnix(string(res))
		tempTime = utils.UnixToTime(unixTime)
		if tempTime.After(latestTime) {
			latestTime = tempTime
		}
	}
	return latestTime, err
}

func (lfs *LfsService) GetObjectAvailTime(object *pb.ObjectInfo) (string, error) {
	latestTime := time.Unix(0, 0)
	bucket := lfs.CurrentLog.BucketByID[object.BucketID]
	blockCount := bucket.DataCount + bucket.ParityCount
	bm, err := metainfo.NewBlockMeta(lfs.UserID, strconv.Itoa(int(object.BucketID)), strconv.Itoa(int(object.StripeStart)), "")
	if err != nil {
		return "", err
	}
	for i := 0; i < int(blockCount); i++ {
		bm.SetBid(strconv.Itoa(i))
		blockID := bm.ToString()
		blockAvailTime, err := lfs.getLastChalTime(blockID)
		if err != nil {
			log.Printf("Get block-%s's availTime failed!err: %v\n", blockID, err)
			continue
		}
		if blockAvailTime.After(latestTime) {
			latestTime = blockAvailTime
		}
	}
	return latestTime.Format(utils.BASETIME), nil
}

func (dl *Download) getMessage(ncid string, provider string) ([]byte, *big.Int, error) {
	money := big.NewInt(0)
	//user给channel合约签名，发给provider
	userID := dl.LfsService.UserID
	privateKey := dl.LfsService.PrivateKey
	localAddress, providerAddress, hexSK, err := getParamsForSign(userID, provider, privateKey)
	if err != nil {
		log.Printf("getParamsForSign about Block %s from %s failed.\n", ncid, provider)
		return nil, nil, err
	}

	var channelAddr common.Address

	//判断是不是测试user，如果是，就将channelAddress设为0
	node := localNode
	cfg, err := node.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed.")
		return nil, nil, err
	}
	if cfg.Test {
		channelAddress := contracts.InvalidAddr
		channelAddr = common.HexToAddress(channelAddress)
		//设置此次下载需要签名的金额，money此时不用变仍为0
	} else {
		cs := GetContractService(userID)
		if cs == nil {
			return nil, nil, ErrUserNotExist
		}
		Item, err := cs.GetChannelItem(provider)
		if err != nil {
			return nil, nil, err
		}
		channelAddr = common.HexToAddress(Item.ChannelAddr)
		// 此次下载需要签名的金额，在valueBase的基础上再加上此次下载需要支付的money，就是此次签名的value
		addValue := int64((BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
		money = money.Add(Item.Value, big.NewInt(addValue)) //100 + valueBase
	}
	moneyByte := money.Bytes()

	//签名
	sig, err := contracts.SignForChannel(channelAddr, money, hexSK)
	if err != nil {
		log.Printf("signature about Block %s from %s failed.\n", ncid, provider)
		return nil, nil, err
	}
	//将签名信息、user公钥、user地址、provider地址、签名金额一并发给provider
	pubKey, err := utils.GetCompressedPkFromHexSk(hexSK)
	if err != nil {
		log.Println("get public key error.")
		return nil, nil, err
	}

	message := &pb.SignForChannel{
		Sig:             sig,
		UserPK:          pubKey,
		UserAddress:     localAddress.String(),
		ProviderAddress: providerAddress.String(),
		Money:           moneyByte,
	}
	mes, err := proto.Marshal(message)
	if err != nil {
		log.Println("protoMarshal about Block", ncid, "from", provider, "failed. err:", err)
		return nil, nil, err
	}
	return mes, money, nil
}
