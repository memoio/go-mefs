package user

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

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

	bucket.Lock()
	defer bucket.Unlock()

	utils.MLogger.Info("Upload object: ", objectName, " to bucket: ", bucketName, " begin")

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

	encoder := dataformat.NewDataCoder(int(bucket.Policy), int(bucket.DataCount), int(bucket.ParityCount), dataformat.CurrentVersion, int(bucket.TagFlag), int(bucket.SegmentSize), dataformat.DefaultSegmentCount, l.keySet)

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

	utils.MLogger.Info("Upload object: ", objectName, " to bucket: ", bucketName, " end")
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
	dc := enc.Prefix.DataCount
	pc := enc.Prefix.ParityCount
	bc := int(dc + pc)
	least := int(dc + pc/2)
	segSize := enc.Prefix.SegmentSize
	readUnit := int(segSize) * int(dc) //每一次读取的数据，尽量读一个整的
	readByte := utils.SegementCount * readUnit

	rdata := make([]byte, readByte)
	extra := make([]byte, 0, readByte)

	h := md5.New()
	parllel := int32(0)
	var wg sync.WaitGroup

	breakFlag := false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("upload cancle")
			return nil
		default:
			// clear itself
			data := make([]byte, 0, readByte)
			readLen := readByte - int(u.curOffset)*readUnit - len(extra)
			if len(extra) > 0 {
				data = append(data, extra...)
				extra = extra[:0]
			}

			n, err := io.ReadAtLeast(u.reader, rdata, readLen)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				breakFlag = true
			} else if err != nil {
				return err
			}

			// need handle n > readLen
			if n > readLen {
				data = append(data, rdata[:readLen]...)
				extra = append(extra, rdata[readLen:n]...)
				n = readLen
			} else {
				data = append(data, rdata[:n]...)
			}

			// 对整个文件的数据进行MD5校验
			h.Write(data)

			endOffset := int(u.curOffset) + 1 + (len(data)-1)/int(u.encoder.Prefix.SegmentSize*u.encoder.Prefix.DataCount)
			u.length += int64(n)
			if endOffset > int(u.encoder.Prefix.GetSegmentCount()) {
				utils.MLogger.Error("wrong offset, need to handle: ", endOffset)
				return errors.New("read length unexpected err")
			}

			if endOffset == int(u.encoder.Prefix.GetSegmentCount()) { //如果写满了一个stripe
				u.curStripe++
				u.curOffset = 0
			} else {
				u.curOffset = int64(endOffset)
			}

			for {
				if atomic.LoadInt32(&parllel) < 32 {
					break
				}
				time.Sleep(5 * time.Second)
			}
			atomic.AddInt32(&parllel, 1)
			wg.Add(1)

			go func(data []byte, stripeID, start int) {
				defer wg.Done()
				defer atomic.AddInt32(&parllel, -1)
				// encrypt
				if u.encrypt {
					if len(data)%aes.BlockSize != 0 {
						data = aes.PKCS5Padding(data)
					}
					data, err = aes.AesEncrypt(data, u.sKey[:])
					if err != nil {
						return
					}
				}

				bm, _ := metainfo.NewBlockMeta(u.fsID, strconv.Itoa(int(u.bucketID)), strconv.Itoa(stripeID), "0")

				encodedData, offset, err := enc.Encode(data, bm.ToString(3), start)
				if err != nil {
					utils.MLogger.Info("encodedData", err)
					return
				}

				var count int
				blockMetas := make([]blockMeta, bc)
				for k := 0; k < 3; k++ {
					count = 0
					if start == 0 {
						pros, _, _ := u.gInfo.GetProviders(bc)
						if len(pros) < least {
							utils.MLogger.Info("putobject err：", ErrNoEnoughProvider)
							return
						}

						for i := 0; i < bc; i++ {
							bm.SetCid(strconv.Itoa(i))
							ncid := bm.ToString()
							km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)
							blockMetas[i].cid = ncid
							blockMetas[i].offset = offset
							blockMetas[i].provider = u.fsID
							if i > len(pros) {
								continue
							}
							err := u.gInfo.ds.PutBlock(ctx, km.ToString(), encodedData[i], pros[i])
							if err != nil {
								utils.MLogger.Info("Put Block", ncid, u.curOffset, offset, "to", pros[i], "failed:", err)
								continue
							}
							count++
							blockMetas[i].provider = pros[i]
						}
					} else {
						for i := 0; i < bc; i++ {
							bm.SetCid(strconv.Itoa(i))
							ncid := bm.ToString()
							km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block, strconv.Itoa(int(start)), strconv.Itoa(offset))
							blockMetas[i].cid = ncid
							blockMetas[i].offset = offset
							blockMetas[i].provider = u.fsID

							provider, _, err := u.gInfo.getBlockProviders(ncid)
							if err != nil || provider == u.fsID {
								continue
							}

							err = u.gInfo.ds.AppendBlock(ctx, km.ToString(), encodedData[i], provider)
							if err != nil {
								utils.MLogger.Info("Put Block", ncid, u.curOffset, offset, "to", provider, "failed:", err)
								continue
							}
							count++
							blockMetas[i].provider = provider
						}
					}

					//没有达到最低安全标准，返回错误
					if count >= least {
						time.Sleep(60 * time.Second)
						break
					}
				}

				for _, v := range blockMetas {
					err = u.gInfo.putDataMetaToKeepers(v.cid, v.provider, v.offset)
					if err != nil {
						utils.MLogger.Info("putobject", err)
						return
					}
				}
			}(data, int(u.curStripe), int(u.curOffset))
		}
	}

	u.etag = hex.EncodeToString(h.Sum(nil))

	wg.Wait()

	return nil
}
