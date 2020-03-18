package user

import (
	"context"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"io"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
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
	encrypt   int32
	sKey      [32]byte
	bucketID  int64
	begin     int64
	length    int64
	rawLen    int64
	sucLen    int64
	etag      string
	gInfo     *groupInfo
	reader    io.Reader
	startTime time.Time
	encoder   *dataformat.DataCoder
}

// PutObject constructs upload process
func (l *LfsInfo) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts ObjectOptions) (*mpb.ObjectInfo, error) {
	utils.MLogger.Infof("Upload object: %s to bucket: %s begin", objectName, bucketName)

	bucket, object, err := l.getBucketAndObjectInfo(bucketName, objectName, true)
	if err != nil {
		return nil, err
	}

	return l.addObjectData(ctx, bucket, object, reader)
}

// AppendObject constructs upload process
func (l *LfsInfo) AppendObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts ObjectOptions) (*mpb.ObjectInfo, error) {
	utils.MLogger.Infof("Upload append object: %s to bucket: %s begin", objectName, bucketName)
	bucket, object, err := l.getBucketAndObjectInfo(bucketName, objectName, false)
	if err != nil {
		return nil, err
	}

	return l.addObjectData(ctx, bucket, object, reader)
}

func (l *LfsInfo) getBucketAndObjectInfo(bucketName, objectName string, creation bool) (*superBucket, *objectInfo, error) {
	if !l.online || l.meta.buckets == nil {
		return nil, nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, nil, ErrLfsReadOnly
	}

	err := checkObjectName(objectName)
	if err != nil {
		return nil, nil, ErrObjectNameInvalid
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, nil, ErrBucketNotExist
	}

	objectElement, ok := bucket.objects[objectName]
	if ok || objectElement != nil {
		return bucket, objectElement, nil
	}

	bucket.Lock()
	defer bucket.Unlock()

	if creation {
		//add Object
		ct := time.Now().UnixNano()
		oInfo := &mpb.Object{
			Name:     objectName,
			BucketID: bucket.BucketID,
			CTime:    ct,
			ObjectID: bucket.NextObjectID,
			Dir:      false,
		}

		object := &objectInfo{
			ObjectInfo: mpb.ObjectInfo{
				Info:      oInfo,
				PartCount: 0,
				Parts:     make([]*mpb.ObjectPart, 0, 1),
				Deletion:  false,
				CTime:     ct,
				MTime:     ct,
			},
		}

		bucket.objects[objectName] = object
		bucket.NextObjectID++

		payload, err := proto.Marshal(oInfo)
		if err != nil {
			return nil, nil, err
		}
		op := &mpb.OpRecord{
			OpType:  mpb.LfsOp_OpAdd,
			OpID:    bucket.GetNextOpID(),
			Payload: payload,
		}

		l.flushObjectMeta(bucket, false, op)
		bucket.NextOpID++

		//gen_root
		tag, err := proto.Marshal(op)
		if err != nil {
			return nil, nil, err
		}
		bucket.mtree.Push(tag)
		bucket.Root = bucket.mtree.Root()
		bucket.dirty = true
		utils.MLogger.Infof("Upload create object: %s in bucket: %s", objectName, bucketName)
		return bucket, object, nil
	}
	return nil, nil, ErrObjectNotExist
}

// make sure bucket and object is ont empty
func (l *LfsInfo) addObjectData(ctx context.Context, bucket *superBucket, object *objectInfo, reader io.Reader) (*mpb.ObjectInfo, error) {
	bucket.Lock()
	defer bucket.Unlock()

	object.Lock()
	defer object.Unlock()

	if object.Info.Dir {
		return &object.ObjectInfo, ErrObjectIsDir
	}

	bopt := &mpb.BlockOptions{
		Bopts:   bucket.BOpts,
		Start:   0,
		UserID:  l.userID,
		QueryID: l.fsID,
	}

	encoder := dataformat.NewDataCoderWithPrefix(l.keySet, bopt)

	//append Object
	opart := &mpb.ObjectPart{
		Name:     object.GetInfo().GetName(),
		ObjectID: object.GetInfo().GetObjectID(),
		PartID:   object.GetPartCount(),
		Start:    bucket.GetLength(),
		CTime:    time.Now().UnixNano(),
	}

	ul := &uploadTask{
		startTime: time.Now(), // for queue?
		reader:    reader,
		gInfo:     l.gInfo,
		bucketID:  bucket.BucketID,
		begin:     bucket.GetLength(),
		encoder:   encoder,
		encrypt:   bucket.BOpts.Encryption,
	}

	if bucket.BOpts.Encryption == 1 {
		ul.sKey = aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.GetInfo().GetObjectID())
	}

	err := ul.Start(ctx)
	if err != nil {
		utils.MLogger.Infof("Add data to object: %s in bucket: %s fails %s", object.GetInfo().GetName(), bucket.GetName(), err)
		return &object.ObjectInfo, err
	}

	// upload success length is less than expected,
	// treated as error
	padding := int64(0)
	if ul.encrypt == 1 {
		if ul.length%aes.BlockSize != 0 {
			padding = int64(aes.BlockSize) - ul.length%int64(aes.BlockSize)
		}
	}

	if ul.length+padding != ul.sucLen {
		utils.MLogger.Infof("upload %d, but success %d", ul.length, ul.sucLen)
		return &object.ObjectInfo, ErrUpload
	}

	// opart
	opart.ETag = ul.etag
	opart.Length = int64(ul.length)

	// object
	object.Parts = append(object.Parts, opart)
	object.PartCount++
	object.Length += int64(ul.length)
	object.ETag = calulateETag(object)
	object.MTime = opart.CTime

	// bucket
	bucket.Length += int64(ul.rawLen)
	payload, err := proto.Marshal(opart)
	if err != nil {
		return nil, err
	}
	op := &mpb.OpRecord{
		OpType:  mpb.LfsOp_OpAppend,
		OpID:    bucket.GetNextOpID(),
		Payload: payload,
	}

	l.flushObjectMeta(bucket, false, op)
	bucket.NextOpID++

	// leaf is OpID + PayLoad
	tag, err := proto.Marshal(op)
	if err != nil {
		return nil, err
	}
	bucket.mtree.Push(tag)
	bucket.Root = bucket.mtree.Root()

	bucket.dirty = true
	utils.MLogger.Infof("Add data to object: %s in bucket: %s end, length is: %d", object.GetInfo().GetName(), bucket.GetName(), opart.Length)
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
	end      int
	start    int
}

//Start 上传文件
func (u *uploadTask) Start(ctx context.Context) error {
	enc := u.encoder
	dc := enc.Prefix.Bopts.DataCount
	pc := enc.Prefix.Bopts.ParityCount
	bc := int(dc + pc)
	least := int(dc + pc/2)
	segSize := enc.Prefix.Bopts.SegmentSize

	segStripeSize := int(segSize * dc)
	stripeSize := int(enc.Prefix.Bopts.SegmentCount) * segStripeSize

	curStripe := int(u.begin) / stripeSize
	curOffset := (int(u.begin) % stripeSize) / segStripeSize

	var pros []string
	breakFlag := false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("upload cancel")
			return nil
		default:
			pros, _, _ = u.gInfo.GetProviders(ctx, bc)
			if len(pros) >= least {
				breakFlag = true
			} else {
				utils.MLogger.Warn("cannot get enough providers")
				time.Sleep(60 * time.Second)
			}
		}
	}

	var bEnc cipher.BlockMode
	if u.encrypt == 1 {
		tmpEnc, err := aes.ContructAesEnc(u.sKey[:])
		if err != nil {
			return err
		}
		bEnc = tmpEnc
	}

	tn := os.Getenv("MEFS_TRANS")
	if tn != "" {
		tNum, err := strconv.Atoi(tn)
		if err != nil {
			transNum = DefaultTransNum
		} else {
			transNum = tNum
		}
	} else {
		transNum = DefaultTransNum
	}

	h := md5.New()
	parllel := int32(0)
	var wg sync.WaitGroup
	rdata := make([]byte, stripeSize)
	extra := make([]byte, 0, stripeSize)

	breakFlag = false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("upload cancel")
			return nil
		default:
			// clear itself
			data := make([]byte, 0, stripeSize)
			extraLen := len(extra)
			readLen := stripeSize - int(curOffset)*segStripeSize - extraLen
			if extraLen > 0 {
				data = append(data, extra...)
				extra = extra[:0]
			}

			utils.MLogger.Debugf("Upload object: stripe: %d, seg offset: %d, expected length: %d", curStripe, curOffset, readLen)

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

			endOffset := curOffset + (n-1)/segStripeSize + 1

			utils.MLogger.Debugf("Upload object: stripe: %d, seg offset: %d, length: %d", curStripe, curOffset, n)

			// handle it before
			if endOffset > int(u.encoder.Prefix.Bopts.GetSegmentCount()) {
				utils.MLogger.Error("Wrong offset, need to handle: ", endOffset)
				return role.ErrRead
			}

			u.rawLen += int64((endOffset - curOffset) * segStripeSize)
			u.length += int64(n)

			// encrypt
			if u.encrypt == 1 {
				if len(data)%aes.BlockSize != 0 {
					data = aes.PKCS5Padding(data)
				}
				crypted := make([]byte, len(data))
				bEnc.CryptBlocks(crypted, data)
				copy(data, crypted)
			}

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

				bm, _ := metainfo.NewBlockMeta(u.gInfo.groupID, strconv.Itoa(int(u.bucketID)), strconv.Itoa(stripeID), "0")

				blockMetas := make([]blockMeta, bc)
				for i := 0; i < bc; i++ {
					bm.SetCid(strconv.Itoa(i))
					ncid := bm.ToString()
					blockMetas[i].cid = ncid
					if i < len(pros) {
						blockMetas[i].provider = pros[i]
					} else {
						blockMetas[i].provider = u.gInfo.groupID
					}

					if start != 0 {
						provider, _, err := u.gInfo.getBlockProviders(ctx, ncid)
						if err != nil {
							continue
						}
						blockMetas[i].provider = provider
					}
				}
				count := int32(0)
				for be := 0; be*segStripeSize < len(data); be += transNum {
					var transData []byte
					if len(data) > (be+transNum)*segStripeSize {
						transData = data[be*segStripeSize : (be+transNum)*segStripeSize]
					} else {
						// last one
						transData = data[be*segStripeSize:]
					}

					encodedData, offset, err := enc.Encode(transData, bm.ToString(3), start)
					if err != nil {
						return
					}

					count = 0
					var pwg sync.WaitGroup
					for i := 0; i < bc; i++ {
						blockMetas[i].end = int(offset)
						blockMetas[i].start = int(start)
						provider := blockMetas[i].provider
						if provider == u.gInfo.groupID {
							continue
						}
						pwg.Add(1)
						if start == 0 {
							go func(num int, edata []byte, proID string) {
								defer pwg.Done()
								km, _ := metainfo.NewKey(blockMetas[num].cid, mpb.KeyType_Block)
								for k := 0; k < 10; k++ {
									err := u.gInfo.ds.PutBlock(ctx, km.ToString(), edata, proID)
									if err != nil {
										utils.MLogger.Warn("Put Block: ", km.ToString(), " to: ", proID, "  failed: ", err)
										if u.gInfo.ds.Connect(ctx, proID) {
											tdelay := rand.Int63n(int64(k+1) * 60000000000)
											time.Sleep(time.Duration(60000000000*int64(k) + tdelay))
											continue
										}
										break
									} else {
										atomic.AddInt32(&count, 1)
										break
									}
								}
							}(i, encodedData[i], provider)
						} else {
							go func(num int, edata []byte, proID string) {
								defer pwg.Done()
								km, _ := metainfo.NewKey(blockMetas[num].cid, mpb.KeyType_Block, strconv.Itoa(blockMetas[num].start), strconv.Itoa(blockMetas[num].end-blockMetas[num].start))
								for k := 0; k < 10; k++ {
									err := u.gInfo.ds.AppendBlock(ctx, km.ToString(), edata, proID)
									if err != nil {
										utils.MLogger.Warn("Append Block: ", km.ToString(), " to: ", proID, " failed: ", err)
										if u.gInfo.ds.Connect(ctx, proID) {
											tdelay := rand.Int63n(int64(k+1) * 60000000000)
											time.Sleep(time.Duration(60000000000*int64(k) + tdelay))
											continue
										}
										break
									} else {
										atomic.AddInt32(&count, 1)
										break
									}
								}
							}(i, encodedData[i], provider)
						}
					}

					pwg.Wait()
					start = offset
					if count >= dc {
						atomic.AddInt64(&u.sucLen, int64(len(transData)))
					}
				}
				if count >= dc {
					for _, v := range blockMetas {
						u.gInfo.putDataMetaToKeepers(ctx, v.cid, v.provider, v.end)
					}
				}
			}(data, curStripe, curOffset)

			if endOffset == int(u.encoder.Prefix.Bopts.GetSegmentCount()) { //如果写满了一个stripe
				curStripe++
				curOffset = 0
			} else {
				curOffset = endOffset
			}
		}
	}

	u.etag = hex.EncodeToString(h.Sum(nil))

	wg.Wait()

	return nil
}

func calulateETag(ob *objectInfo) string {
	if len(ob.GetParts()) == 1 {
		return ob.GetParts()[0].ETag
	}

	var hashes []byte
	for i := 0; i < len(ob.GetParts()); i++ {
		md5, err := hex.DecodeString(ob.GetParts()[i].ETag)
		if err != nil {
			continue
		}
		hashes = append(hashes, md5...)
	}

	sum := md5.Sum(hashes)
	return hex.EncodeToString(sum[:])
}
