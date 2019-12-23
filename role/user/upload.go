package user

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//UploadOptions 下载时的一些参数
type UploadOptions struct {
	// Start and end length
	Length int64
}

// uploadTask has info for upload
type uploadTask struct { //一个上传任务实例
	gInfo       *groupInfo
	sKey        []byte
	bucketID    int32
	curStripe   int64
	offset      int64
	length      int64
	segmentSize int32
	tagFlag     int32
	reader      io.Reader
	startTime   time.Time
	encoder     *dataformat.DataEncoder
}

// ConstructUpload constructs upload process
func (l *LfsInfo) ConstructUpload(objectName, prefix, bucketName string, reader io.Reader) (Job, error) {
	if !l.online {
		return nil, errors.New("user is not running")
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}

	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	if err := checkObjectName(objectName); err != nil {
		return nil, err
	}
	if objectElement, ok := bucket.objects[prefix+objectName]; ok || objectElement != nil {
		return nil, ErrObjectAlreadyExist
	}
	gp := l.gInfo

	objectName = prefix + objectName
	if bucket.Policy != dataformat.RsPolicy && bucket.Policy != dataformat.MulPolicy {
		return nil, ErrPolicy
	}

	object := &objectInfo{
		ObjectInfo: pb.ObjectInfo{
			Name:        objectName,
			BucketID:    bucketID,
			Ctime:       time.Now().Format(utils.BASETIME),
			StripeStart: bucket.CurStripe,
			OffsetStart: bucket.NextOffset,
			Deletion:    false,
			Dir:         false,
		},
	}
	objectElement := bucket.orderedObjects.PushBack(object)
	bucket.objects[objectName] = objectElement
	encoder := dataformat.NewDataEncoder(bucket.Policy, bucket.DataCount, bucket.ParityCount,
		int32(bucket.TagFlag), bucket.SegmentSize, getGroup(l.userid).getKeyset())

	skey := make([]byte, 0)
	if l.superBucket.Encryption {
		// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
		tmpkey := l.privateKey
		tmpkey = append(tmpkey, byte(bucket.BucketID))
		skey := sha256.Sum256(tmpkey)
	}

	ul := &uploadTask{
		lService:    lfs,
		superBucket: bucket,
		objectInfo:  object,
		sKey:        skey,
		TagFlag:     int32(bucket.GetTagFlag()),
		segmentSize: int32(bucket.GetSegmentSize()),
		startTime:   time.Now(),
		Prefix:      prefix,
		Reader:      reader,
		BucketID:    bucket.BucketID,
		keepers:     conkeepers,
		ObjectName:  objectName,
		encoder:     encoder,
	}
	return ul, nil
}

// Stop is
func (u *uploadTask) Stop(context.Context) error {
	return nil
}

// Cancel is
func (u *uploadTask) Cancel(context.Context) error {
	return nil
}

// Done is
func (u *uploadTask) Done() {
	return
}

// Info gets
func (u *uploadTask) Info() (interface{}, error) {
	if u == nil || u.objectInfo == nil {
		return nil, ErrObjectNotExist
	}
	return u, nil
}

//Start 上传文件
func (u *uploadTask) Start(ctx context.Context) error {
	err := u.putObject(ctx)
	if err != nil {
		if u.objectInfo.Size > 0 {
			u.superBucket.dirty = true //需要记录，可能上传一部分然后失败，空间已占用
		} else { //没有占用任何空间，清除信息
			objectElement := u.superBucket.objects[u.ObjectName]
			u.superBucket.orderedObjects.Remove(objectElement)
			delete(u.superBucket.objects, u.ObjectName)
		}
		return err
	}
	u.superBucket.dirty = true
	return nil
}

type blockMeta struct {
	cid      string
	provider string
	offset   int
}

// 具体实现
func (u *uploadTask) putObject(ctx context.Context) error {
	u.objectInfo.Lock()
	defer u.objectInfo.Unlock()

	encoder := u.encoder
	dataCount := encoder.DataCount
	parityCount := encoder.ParityCount
	blockCount := dataCount + parityCount

	var readByte int32
	//var err error
	readByte = utils.SegementCount * u.segmentSize * dataCount //每一次读取的数据，尽量读一个整的

	var breakFlag = false

	var curOffset int64
	blockMetas := make([]blockMeta, blockCount)
	//尽量复用内存
	var data []byte
	h := md5.New()
Loop:
	for { //循环上传每一个块
		select {
		case <-ctx.Done():
			fmt.Println("上传取消")
			return nil
		default:
			curOffset = u.superBucket.NextOffset
			if curOffset == 0 { //不为零则为追加模式
				if len(data) != int(readByte) {
					data = make([]byte, readByte)
				}
			} else {
				data = make([]byte, readByte-int32(curOffset)*int32(u.segmentSize)*int32(dataCount))
			}

			//尽量一次性读一整个stripe所需数据
			n, err := io.ReadAtLeast(u.Reader, data, len(data))
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				breakFlag = true
			} else if err != nil {
				return err
			}
			data = data[:n]
			//在此记录下本次上传多少，等发送完毕再计入
			tempSize := n
			// 对整个文件的数据进行MD5校验
			h.Write(data)

			//这里只输入BucketID和StripeID，blockID后面动态改变
			bm, err := metainfo.NewBlockMeta(u.lService.userid, strconv.Itoa(int(u.BucketID)), strconv.Itoa(int(u.superBucket.CurStripe)), "0")
			if err != nil {
				return err
			}

			if len(u.skey) > 0 {
				if len(data)%aes.BlockSize != 0 {
					data = aes.PKCS5Padding(data)
				}
				data, err = aes.AesEncrypt(data, u.skey[:])
				if err != nil {
					return err
				}
			}

			//在这里先将数据编码
			encodedData, offset, err := encoder.Encode(data, bm.ToString(3), int32(curOffset))
			if err != nil {
				log.Println("encodedData", err)
				return err
			}

			//如果是新的一个stripe，则需要重新找provider
			var providers []string
			if curOffset == 0 {
				providers, _ = getGroup(u.lService.userid).getProviders(int(blockCount))
				if providers == nil || len(providers) < int(dataCount+parityCount/2) {
					log.Println("putobject", ErrNoEnoughProvider)
					return ErrNoEnoughProvider
				}
				if len(providers) > int(blockCount) {
					providers = providers[:blockCount]
				}
			}

			//总共上传成功几个块
			var count int
			var provider string
			for i := 0; i < int(blockCount); i++ {
				bm.SetBid(strconv.Itoa(i))
				ncid := bm.ToString()
				var km *metainfo.KeyMeta

				//如果是追加，则要找到上一个provider
				if curOffset == 0 {
					if i < len(providers) {
						provider = providers[i]
						km, _ = metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(offset))
					} else {
						blockMetas[i].cid = ncid
						blockMetas[i].offset = offset
						blockMetas[i].provider = u.lService.userid
						continue
					}
				} else {
					provider, _, err = getGroup(u.lService.userid).getBlockProviders(ncid)
					if err != nil || provider == u.lService.userid {
						log.Println("Append Block to", provider, "failed:", err)
						if _, err := peer.IDB58Decode(provider); provider != "" && err == nil {
							blockMetas[i].cid = ncid
							blockMetas[i].offset = offset
							blockMetas[i].provider = provider
						} else {
							blockMetas[i].cid = ncid
							blockMetas[i].offset = offset
							blockMetas[i].provider = u.lService.userid
						}
						continue
					}
					km, _ = metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "append", strconv.Itoa(int(curOffset)), strconv.Itoa(offset))
				}

				//开始上传这个块
				Key := km.ToString()
				bcid := cid.NewCidV2([]byte(Key))
				b, err := blocks.NewBlockWithCid(encodedData[i], bcid)
				if err != nil {
					return err
				}
				err = localNode.Blocks.PutBlockTo(b, provider)
				if err != nil {
					log.Println("Put Block", ncid, curOffset, offset, "to", provider, "failed:", err)
					continue
				}
				count++

				blockMetas[i].cid = ncid
				blockMetas[i].offset = offset
				blockMetas[i].provider = provider
			}
			//没有达到最低安全标准，返回错误
			if count < int(dataCount+parityCount/2) {
				return ErrNoEnoughProvider
			}

			for _, v := range blockMetas {
				err = l.gInfo.putDataMetaToKeepers(v.cid, v.provider, v.offset)
				if err != nil {
					log.Println("putobject", err)
					return err
				}
			}
			u.objectInfo.Size += int64(tempSize)
			if offset >= int(utils.SegementCount-1) { //如果写满了一个stripe
				u.superBucket.CurStripe++
				u.superBucket.NextOffset = 0
			} else {
				u.superBucket.NextOffset = int64(offset + 1)
			}
			if breakFlag {
				u.objectInfo.ETag = hex.EncodeToString(h.Sum(nil))
				break Loop
			}
		}
	}
	return nil
}
