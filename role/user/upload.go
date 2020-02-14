package user

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
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
	encrypt   int32
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
func (l *LfsInfo) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader) (*mpb.ObjectInfo, error) {
	if !l.online || l.meta.bucketNameToID == nil {
		return nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, ErrReadOnly
	}

	err := checkObjectName(objectName)
	if err != nil {
		return nil, ErrObjectNameInvalid
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}

	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	bucket.Lock()
	defer bucket.Unlock()

	if objectElement, ok := bucket.objects[objectName]; ok || objectElement != nil {
		return nil, ErrObjectAlreadyExist
	}

	bo := bucket.BOpts
	if bo.Policy != dataformat.RsPolicy && bo.Policy != dataformat.MulPolicy {
		return nil, ErrPolicy
	}

	utils.MLogger.Infof("Upload object: %s to bucket: %s begin", objectName, bucketName)

	segStripeSize := int64(bo.SegmentSize)
	stripeSize := int64(bo.SegmentCount*bo.DataCount) * segStripeSize

	start := bucket.CurStripe*stripeSize + bucket.NextSeg*segStripeSize

	opart := &mpb.ObjectPart{
		Name:  objectName,
		Start: start,
	}

	object := &objectInfo{
		ObjectInfo: mpb.ObjectInfo{
			ObjectID: bucket.NextObjectID,
			OPart:    opart,
			BucketID: bucketID,
			Ctime:    time.Now().Unix(),
			Deletion: false,
			Dir:      false,
		},
	}

	bucket.objects[objectName] = object

	encoder := dataformat.NewDataCoderWithBopts(bucket.BOpts, l.keySet)

	ul := &uploadTask{
		fsID:      l.fsID,
		startTime: time.Now(), // for queue?
		reader:    reader,
		gInfo:     l.gInfo,
		bucketID:  bucketID,
		curStripe: bucket.CurStripe,
		curOffset: bucket.NextSeg,
		encoder:   encoder,
		encrypt:   bo.Encryption,
	}

	if bo.Encryption == 1 {
		ul.sKey = aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucketID, start)
	}

	err = ul.Start(ctx)
	if err != nil && ul.length == 0 {
		delete(bucket.objects, objectName)
		return &object.ObjectInfo, err
	}

	object.OPart.ETag = ul.etag
	object.OPart.Length = ul.length
	bucket.NextObjectID++
	bucket.CurStripe = ul.curStripe
	bucket.NextSeg = ul.curOffset

	// leaf is objectname + md5
	bucket.mtree.Push([]byte(objectName + object.OPart.ETag))
	bucket.Root = bucket.mtree.Root()

	bucket.dirty = true
	utils.MLogger.Infof("Upload object: %s to bucket: %s end, length is: %d", objectName, bucketName, object.OPart.Length)
	return &object.ObjectInfo, err
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
	dc := enc.Prefix.Bopts.DataCount
	pc := enc.Prefix.Bopts.ParityCount
	bc := int(dc + pc)
	least := int(dc + pc/2)
	segSize := enc.Prefix.Bopts.SegmentSize
	readUnit := int(segSize) * int(dc) //每一次读取的数据，尽量读一个整的
	readByte := int(enc.Prefix.Bopts.SegmentCount) * readUnit

	rdata := make([]byte, readByte)
	extra := make([]byte, 0, readByte)

	h := md5.New()
	parllel := int32(0)
	var wg sync.WaitGroup

	breakFlag := false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("upload cancel")
			return nil
		default:
			pros, _, _ := u.gInfo.GetProviders(bc)
			if len(pros) >= least {
				breakFlag = true
			} else {
				utils.MLogger.Warn("cannot get enough providers")
				time.Sleep(60 * time.Second)
			}
		}
	}

	breakFlag = false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("upload cancel")
			return nil
		default:
			// clear itself
			data := make([]byte, 0, readByte)
			extraLen := len(extra)
			readLen := readByte - int(u.curOffset)*readUnit - extraLen
			if extraLen > 0 {
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

			// plus extra
			n += extraLen

			// 对整个文件的数据进行MD5校验
			h.Write(data)

			endOffset := int(u.curOffset) + (n-1)/int(u.encoder.Prefix.Bopts.SegmentSize*u.encoder.Prefix.Bopts.DataCount)

			utils.MLogger.Debugf("Upload object: stripe: %d, seg offset: %d, length: %d", u.curStripe, u.curOffset, n)

			// handle it before
			if endOffset >= int(u.encoder.Prefix.Bopts.GetSegmentCount()) {
				utils.MLogger.Error("Wrong offset, need to handle: ", endOffset)
				return errors.New("Read length unexpected err")
			}

			u.length += int64(n)

			for {
				if atomic.LoadInt32(&parllel) < 32 {
					break
				}
				time.Sleep(time.Second)
			}
			atomic.AddInt32(&parllel, 1)
			wg.Add(1)

			go func(data []byte, stripeID, start int) {
				defer wg.Done()
				defer atomic.AddInt32(&parllel, -1)
				// encrypt
				if u.encrypt == 1 {
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
					return
				}

				count := int32(0)
				blockMetas := make([]blockMeta, bc)
				var pwg sync.WaitGroup

				count = 0
				if start == 0 {
					pros, _, _ := u.gInfo.GetProviders(bc)
					if len(pros) < least {
						return
					}

					for i := 0; i < bc; i++ {
						bm.SetCid(strconv.Itoa(i))
						ncid := bm.ToString()
						km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block)
						blockMetas[i].cid = ncid
						blockMetas[i].offset = offset
						blockMetas[i].provider = u.fsID
						if i >= len(pros) {
							continue
						}

						blockMetas[i].provider = pros[i]
						pwg.Add(1)
						go func(edata []byte, proID string) {
							defer pwg.Done()
							for k := 0; k < 3; k++ {
								err := u.gInfo.ds.PutBlock(ctx, km.ToString(), edata, proID)
								if err != nil {
									utils.MLogger.Warn("Put Block: ", km.ToString(), " to: ", proID, " failed: ", err)
									time.Sleep(30 * time.Second)
									continue
								} else {
									atomic.AddInt32(&count, 1)
									break
								}
							}

						}(encodedData[i], pros[i])
					}
					pwg.Wait()
				} else {
					for i := 0; i < bc; i++ {
						bm.SetCid(strconv.Itoa(i))
						ncid := bm.ToString()
						km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block, strconv.Itoa(int(start)), strconv.Itoa(offset-start+1))
						blockMetas[i].cid = ncid
						blockMetas[i].offset = offset
						blockMetas[i].provider = u.fsID

						provider, _, err := u.gInfo.getBlockProviders(ncid)
						if err != nil || provider == u.fsID {
							continue
						}
						blockMetas[i].provider = provider

						pwg.Add(1)
						go func(edata []byte, proID string) {
							defer pwg.Done()
							for k := 0; k < 3; k++ {
								err = u.gInfo.ds.AppendBlock(ctx, km.ToString(), edata, proID)
								if err != nil {
									utils.MLogger.Warn("Append Block: ", km.ToString(), " to: ", proID, " failed: ", err)
									time.Sleep(30 * time.Second)
									continue
								} else {
									atomic.AddInt32(&count, 1)
									break
								}
							}
						}(encodedData[i], provider)
					}
					pwg.Wait()
				}

				//没有达到最低安全标准，返回错误
				if count >= dc {
					for _, v := range blockMetas {
						err = u.gInfo.putDataMetaToKeepers(v.cid, v.provider, v.offset)
						if err != nil {
							utils.MLogger.Errorf("put metablock: %s to keepers failed", v.cid)
						}
					}
				}
			}(data, int(u.curStripe), int(u.curOffset))

			if endOffset == int(u.encoder.Prefix.Bopts.GetSegmentCount())-1 { //如果写满了一个stripe
				u.curStripe++
				u.curOffset = 0
			} else {
				u.curOffset = int64(endOffset + 1)
			}
		}
	}

	u.etag = hex.EncodeToString(h.Sum(nil))

	wg.Wait()

	return nil
}
