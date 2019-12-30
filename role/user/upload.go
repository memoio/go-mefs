package user

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"strconv"
	"time"

	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
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
	fsID      string
	encrypt   bool
	sKey      [32]byte
	bucketID  int32
	curStripe int64
	curOffset int64
	length    int64
	etag      string
	gInfo     *groupInfo
	reader    io.Reader
	startTime time.Time
	encoder   *dataformat.DataCoder
}

// PutObject constructs upload process
func (l *LfsInfo) PutObject(bucketName, objectName string, reader io.Reader) (*pb.ObjectInfo, error) {
	log.Println("PutObject: ", bucketName, objectName)
	if !l.online {
		return nil, errors.New("user is not running")
	}

	if l.meta.bucketNameToID == nil {
		return nil, ErrBucketNotExist
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
	if objectElement, ok := bucket.objects[objectName]; ok || objectElement != nil {
		return nil, ErrObjectAlreadyExist
	}

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

	encoder := dataformat.NewDataCoder(bucket.Policy, bucket.DataCount, bucket.ParityCount,
		int32(bucket.TagFlag), bucket.SegmentSize, l.keySet)

	ul := &uploadTask{
		fsID:      l.fsID,
		startTime: time.Now(), // for queue?
		reader:    reader,
		gInfo:     l.gInfo,
		bucketID:  bucketID,
		curStripe: bucket.CurStripe,
		curOffset: bucket.NextOffset,
		encoder:   encoder,
	}

	if bucket.Encryption {
		// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
		tmpkey := l.privateKey
		tmpkey = append(tmpkey, byte(bucket.BucketID))
		ul.encrypt = true
		ul.sKey = sha256.Sum256(tmpkey)
	}

	err := ul.Start(context.Background())
	if err != nil {
		if ul.length > 0 {
			object.ETag = ul.etag
			object.Size = ul.length
			bucket.CurStripe = ul.curStripe
			bucket.NextOffset = ul.curOffset
			bucket.dirty = true //需要记录，可能上传一部分然后失败，空间已占用
		} else { //没有占用任何空间，清除信息
			objectElement := bucket.objects[objectName]
			bucket.orderedObjects.Remove(objectElement)
			delete(bucket.objects, objectName)
		}
		return &object.ObjectInfo, err
	}

	object.ETag = ul.etag
	object.Size = ul.length
	bucket.CurStripe = ul.curStripe
	bucket.NextOffset = ul.curOffset
	bucket.dirty = true

	return &object.ObjectInfo, nil
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
	if u == nil {
		return nil, ErrObjectNotExist
	}
	return u, nil
}

type blockMeta struct {
	cid      string
	provider string
	offset   int
}

//Start 上传文件
func (u *uploadTask) Start(ctx context.Context) error {
	enc := u.encoder
	dc := enc.DataCount
	pc := enc.ParityCount
	bc := int(dc + pc)
	least := int(dc + pc/2)
	segSize := enc.SegmentSize
	readUnit := int(segSize) * int(dc) //每一次读取的数据，尽量读一个整的
	readByte := utils.SegementCount * readUnit

	breakFlag := false

	blockMetas := make([]blockMeta, bc)
	data := make([]byte, readByte)

	h := md5.New()
Loop:
	for { //循环上传每一个块
		select {
		case <-ctx.Done():
			log.Println("上传取消")
			return nil
		default:
			readLen := readByte - int(u.curOffset)*readUnit
			//尽量一次性读一整个stripe所需数据
			n, err := io.ReadAtLeast(u.reader, data, readLen)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				breakFlag = true
			} else if err != nil {
				return err
			}

			tmp := data[:n]

			// 对整个文件的数据进行MD5校验
			h.Write(tmp)

			// encrypt
			if u.encrypt {
				if len(tmp)%aes.BlockSize != 0 {
					tmp = aes.PKCS5Padding(tmp)
				}
				tmp, err = aes.AesEncrypt(tmp, u.sKey[:])
				if err != nil {
					return err
				}
			}

			//在这里先将数据编码
			//这里只输入BucketID和StripeID，blockID后面动态改变
			bm, err := metainfo.NewBlockMeta(u.fsID, strconv.Itoa(int(u.bucketID)), strconv.Itoa(int(u.curStripe)), "0")
			if err != nil {
				return err
			}

			encodedData, offset, err := enc.Encode(tmp, bm.ToString(3), int32(u.curOffset))
			if err != nil {
				log.Println("encodedData", err)
				return err
			}

			//如果是新的一个stripe，则需要重新找provider
			var pros []string
			if u.curOffset == 0 {
				pros, _, err = u.gInfo.GetProviders(bc)
				if pros == nil || len(pros) < least {
					log.Println("putobject err：", ErrNoEnoughProvider)
					return ErrNoEnoughProvider
				}
			}

			//总共上传成功几个块
			var count int
			var provider string
			for i := 0; i < bc; i++ {
				bm.SetBid(strconv.Itoa(i))
				ncid := bm.ToString()
				var km *metainfo.KeyMeta

				//如果是追加，则要找到上一个provider
				if u.curOffset == 0 {
					if i < len(pros) {
						provider = pros[i]
						km, _ = metainfo.NewKeyMeta(ncid, metainfo.Block)
					} else {
						blockMetas[i].cid = ncid
						blockMetas[i].offset = offset
						blockMetas[i].provider = u.fsID
						continue
					}
					err = u.gInfo.ds.PutBlock(ctx, km.ToString(), encodedData[i], provider)
					if err != nil {
						log.Println("Put Block", ncid, u.curOffset, offset, "to", provider, "failed:", err)
						continue
					}
					count++
				} else {
					provider, _, err = u.gInfo.getBlockProviders(ncid)
					if err != nil || provider == u.fsID {
						log.Println("Append Block to", provider, "failed:", err)
						_, err := peer.IDB58Decode(provider)

						if err == nil && provider != "" {
							blockMetas[i].cid = ncid
							blockMetas[i].offset = offset
							blockMetas[i].provider = provider
						} else {
							blockMetas[i].cid = ncid
							blockMetas[i].offset = offset
							blockMetas[i].provider = u.fsID
						}
						continue
					}
					km, _ = metainfo.NewKeyMeta(ncid, metainfo.Block, strconv.Itoa(int(u.curOffset)), strconv.Itoa(offset))
					err = u.gInfo.ds.AppendBlock(ctx, km.ToString(), encodedData[i], provider)
					if err != nil {
						log.Println("Put Block", ncid, u.curOffset, offset, "to", provider, "failed:", err)
						continue
					}
					count++
				}

				blockMetas[i].cid = ncid
				blockMetas[i].offset = offset
				blockMetas[i].provider = provider
			}
			//没有达到最低安全标准，返回错误
			if count < least {
				return ErrNoEnoughProvider
			}

			for _, v := range blockMetas {
				err = u.gInfo.putDataMetaToKeepers(v.cid, v.provider, v.offset)
				if err != nil {
					log.Println("putobject", err)
					return err
				}
			}
			u.length += int64(n)
			if offset >= int(utils.SegementCount-1) { //如果写满了一个stripe
				u.curStripe++
				u.curOffset = 0
			} else {
				u.curOffset = int64(offset + 1)
			}
			if breakFlag {
				u.etag = hex.EncodeToString(h.Sum(nil))
				break Loop
			}
		}
	}
	return nil
}
