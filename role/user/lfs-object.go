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
	"strconv"
	"strings"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

func (lfs *LfsService) DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	// TODO:具体实现
	bucket.Lock.Lock()
	defer bucket.Lock.Unlock()
	if bucket.Objects == nil {
		return nil, ErrObjectNotExist
	}
	object, ok := bucket.Objects[objectName]
	if !ok || object == nil {
		return nil, ErrObjectNotExist
	}
	object.Deletion = true
	bucket.Dirty = true
	return &object.ObjectInfo, nil
}

func (lfs *LfsService) HeadObject(bucketName, objectName string) (*pb.ObjectInfo, string, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, "", err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, "", ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, "", ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, "", ErrBucketNotExist
	}
	// TODO:具体实现
	bucket.Lock.Lock()
	defer bucket.Lock.Unlock()
	if bucket.Objects == nil {
		return nil, "", ErrObjectNotExist
	}
	object, ok := bucket.Objects[objectName]
	if !ok || object == nil {
		return nil, "", ErrObjectNotExist
	}
	AvailTime, _ := lfs.GetObjectAvailTime(object)
	return &object.ObjectInfo, AvailTime, nil
}

func (lfs *LfsService) ListObject(bucketName, pre string, avail bool) ([]*pb.ObjectInfo, []string, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, nil, ErrObjectNotExist
	}
	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, nil, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, nil, ErrBucketNotExist
	}
	var objects []*pb.ObjectInfo
	var availTimes []string
	for _, Object := range bucket.Objects {
		if len(objects) > MAXLISTVALUE { //返回不要过多，应指定好过滤条件
			break
		}
		if Object.Deletion {
			continue
		}
		if avail {
			if strings.HasPrefix(Object.ObjectName, pre) {
				objects = append(objects, &Object.ObjectInfo)
				availTime, _ := lfs.GetObjectAvailTime(Object)
				availTimes = append(availTimes, availTime)
			}
		} else {
			if strings.HasPrefix(Object.ObjectName, pre) {
				objects = append(objects, &Object.ObjectInfo)
			}
		}
	}
	return objects, availTimes, nil
}

func (lfs *LfsService) ShowStorageSpace(bucketName, pre string) (int, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return 0, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return 0, ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return 0, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return 0, ErrBucketNotExist
	}
	var storageSpace int
	for _, Object := range bucket.Objects {
		if Object.Deletion {
			continue
		}
		storageSpace += int(Object.GetObjectSize())
	}
	return storageSpace, nil
}

type Upload struct { //一个上传任务实例
	LfsService  *LfsService
	Object      *Object
	BucketID    int32
	ObjectName  string
	Reader      io.Reader
	NamePrefix  string
	startTime   time.Time
	segmentSize int32
	TagFlag     int32
	keepers     []string
}

func (lfs *LfsService) ConstructUpload(objectName, namePrefix, bucketName string, file files.Node) (Job, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	if object, ok := bucket.Objects[namePrefix+objectName]; ok || object != nil {
		return nil, ErrObjectAlreadyExist
	}
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
		TagFlag:     int32(bucket.GetTagFlag()),
		segmentSize: int32(bucket.GetSegmentSize()),
		startTime:   time.Now(),
		NamePrefix:  namePrefix,
		Reader:      fileNext,
		BucketID:    bucket.BucketID,
		keepers:     conkeepers,
		ObjectName:  objectName,
	}
	return ul, nil
}

func (ul *Upload) Stop(context.Context) error {
	return nil
}
func (ul *Upload) Cancel(context.Context) error {
	return nil
}
func (ul *Upload) Done() {
	return
}
func (ul *Upload) Info() (interface{}, error) {
	if ul == nil || ul.Object == nil {
		return nil, ErrObjectNotExist
	}
	return ul, nil
}

//上传文件
func (ul *Upload) Start(ctx context.Context) error {
	if ul == nil {
		return ErrLfsIsNotRunning
	}
	if err := checkObjectName(ul.ObjectName); err != nil {
		return err
	}
	ObjectName := ul.NamePrefix + ul.ObjectName
	if len(ul.ObjectName) > maxObjectNameLen {
		return ErrObjectNameToolong
	}
	bucket, ok := ul.LfsService.CurrentLog.BucketByID[ul.BucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return ErrBucketNotExist
	}
	if bucket.Policy != dataformat.RsPolicy && bucket.Policy != dataformat.MulPolicy {
		return ErrPolicy
	}
	bucket.Lock.Lock()
	defer bucket.Lock.Unlock()
	if object, ok := bucket.Objects[ObjectName]; ok || object != nil {
		return ErrObjectAlreadyExist
	}

	object := &Object{
		ObjectInfo: pb.ObjectInfo{
			ObjectName:  ObjectName,
			BucketID:    ul.BucketID,
			Ctime:       time.Now().Format(utils.BASETIME),
			StripeStart: ul.LfsService.CurrentLog.BucketByID[ul.BucketID].CurStripe,
			OffsetStart: ul.LfsService.CurrentLog.BucketByID[ul.BucketID].NextOffset,
			Deletion:    false,
			Dir:         false,
		},
	}
	ul.Object = object
	bucket.Objects[ObjectName] = object
	encoder := dataformat.NewDataEncoder(bucket.Policy, bucket.DataCount, bucket.ParityCount,
		int32(bucket.TagFlag), bucket.SegmentSize, GetGroupService(ul.LfsService.UserID).GetKeyset())
	err := ul.putObject(ctx, encoder)
	if err != nil {
		bucket.Dirty = true //需要记录，可能上传一部分然后失败，空间已占用
		return err
	}
	bucket.Dirty = true
	return nil
}

// 具体实现
func (ul *Upload) putObject(ctx context.Context, encoder *dataformat.DataEncoder) error {
	ul.Object.Lock.Lock()
	defer ul.Object.Lock.Unlock()
	dataCount := encoder.DataCount
	parityCount := encoder.ParityCount
	blockCount := dataCount + parityCount
	var readByte int32
	// true为纠删
	switch encoder.Policy {
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
		curOffset := ul.LfsService.CurrentLog.BucketByID[ul.BucketID].NextOffset
		var data []byte
		switch encoder.Policy {
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

		encodedData, offset, err := encoder.Encode(data, bm.ToString(3), curOffset)
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
				km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(offset))
				updateKey := km.ToString()
				bcid := cid.NewCidV2([]byte(updateKey))
				b, err := blocks.NewBlockWithCid(encodedData[i], bcid)
				if err != nil {
					return err
				}
				err = localNode.Blocks.PutBlockTo(b, providers[i])
				if err != nil {
					fmt.Println("Put Block", ncid, "to", providers[i], "failed:", err)
					err = GetGroupService(ul.LfsService.UserID).PutDataMetaToKeepers(ncid, providers[i], offset)
					if err != nil {
						return err
					}
					continue
				}
				count++

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

				km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "append", strconv.Itoa(int(curOffset)), strconv.Itoa(offset))
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
				}
				count++
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

func (lfs *LfsService) GetObjectAvailTime(object *Object) (string, error) {
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
