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
	pool "github.com/libp2p/go-buffer-pool"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var defaultTaskWorkerCount int64 = 8

//UploadOptions 下载时的一些参数
type UploadOptions struct {
	// Start and end length
	Length int64
}

// uploadTask has info for upload
type uploadTask struct { //一个上传任务实例
	encrypt         int32
	sKey            [32]byte
	bucketID        int64
	begin           int64
	length          int64
	rawLen          int64
	sucLen          int64
	etag            string
	taskWorkerCount int64
	gInfo           *groupInfo
	reader          io.Reader
	startTime       time.Time
	encoder         *dataformat.DataCoder
}

// PutObject constructs upload process
func (l *LfsInfo) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts PutObjectOptions) (*mpb.ObjectInfo, error) {
	//上传需要10资源
	ok := l.Sm.TryAcquire(10)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(10)
	utils.MLogger.Infof("Upload object: %s to bucket: %s begin", objectName, bucketName)

	bucket, err := l.getBucketInfo(bucketName)
	if err != nil {
		return nil, err
	}

	objectElement := bucket.Objects.Find(MetaName(objectName))
	if objectElement != nil {
		return nil, ErrObjectAlreadyExist
	}

	bucket.Lock()
	defer bucket.Unlock()

	//add Object
	ct := time.Now().Unix()
	oInfo := &mpb.Object{
		Name:     objectName,
		BucketID: bucket.BucketID,
		CTime:    ct,
		ObjectID: bucket.NextObjectID,
		Dir:      false,
	}

	object := &ObjectInfo{
		ObjectInfo: mpb.ObjectInfo{
			Info:      oInfo,
			PartCount: 0,
			Parts:     make([]*mpb.ObjectPart, 0, 1),
			Deletion:  false,
			CTime:     ct,
			MTime:     ct,
		},
	}

	object.Lock()
	defer object.Unlock()

	obj, opart, err := l.addObjectData(ctx, bucket, object, reader)
	if err != nil {
		return &obj.ObjectInfo, err
	}

	err = l.insertObject(bucket, obj)
	if err != nil {
		return &obj.ObjectInfo, err
	}

	err = l.appendPart(bucket, obj, opart)
	if err != nil {
		return &obj.ObjectInfo, err
	}

	return &obj.ObjectInfo, nil
}

// AppendObject constructs upload process
func (l *LfsInfo) AppendObject(ctx context.Context, bucketName, objectName string, reader io.Reader, opts UploadOptions) (*mpb.ObjectInfo, error) {
	//上传需要10资源
	ok := l.Sm.TryAcquire(10)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(10)
	utils.MLogger.Infof("Upload append object: %s to bucket: %s begin", objectName, bucketName)
	bucket, object, err := l.getBucketAndObjectInfo(bucketName, objectName)
	if err != nil {
		return nil, err
	}

	obj, opart, err := l.addObjectData(ctx, bucket, object, reader)
	if err != nil {
		return &obj.ObjectInfo, err
	}

	err = l.appendPart(bucket, obj, opart)
	if err != nil {
		return &obj.ObjectInfo, err
	}

	return &obj.ObjectInfo, nil
}

func (l *LfsInfo) insertObject(bucket *superBucket, object *ObjectInfo) error {
	bucket.Objects.Insert(MetaName(object.GetInfo().GetName()), object)
	bucket.NextObjectID++

	payload, err := proto.Marshal(object.ObjectInfo.Info)
	if err != nil {
		return err
	}
	op := &mpb.OpRecord{
		OpType:  mpb.LfsOp_OpAdd,
		OpID:    bucket.GetNextOpID(),
		Payload: payload,
	}

	l.flushObjectMeta(bucket, false, op)
	bucket.applyOpID = op.GetOpID()
	bucket.NextOpID++

	//gen_root
	tag, err := proto.Marshal(op)
	if err != nil {
		return err
	}
	bucket.mtree.Push(tag)
	bucket.Root = bucket.mtree.Root()
	bucket.dirty = true
	bucket.MTime = time.Now().Unix()

	l.meta.dirty = true
	utils.MLogger.Infof("Upload create object: %s in bucket: %s", object.GetInfo().GetName(), bucket.GetName())
	return nil
}

func (l *LfsInfo) appendPart(bucket *superBucket, object *ObjectInfo, opart *mpb.ObjectPart) error {
	object.Parts = append(object.Parts, opart)
	object.PartCount++
	object.Length += int64(opart.GetLength())
	object.ETag = calculateETagForNewPart(object.ETag, opart.ETag)
	object.MTime = opart.CTime

	// bucket
	bucket.MTime = opart.CTime
	payload, err := proto.Marshal(opart)
	if err != nil {
		return err
	}
	op := &mpb.OpRecord{
		OpType:  mpb.LfsOp_OpAppend,
		OpID:    bucket.GetNextOpID(),
		Payload: payload,
	}

	l.flushObjectMeta(bucket, false, op)
	bucket.applyOpID = op.GetOpID()
	bucket.NextOpID++

	// leaf is OpID + PayLoad
	tag, err := proto.Marshal(op)
	if err != nil {
		return err
	}
	bucket.mtree.Push(tag)
	bucket.Root = bucket.mtree.Root()
	l.meta.dirty = true

	bucket.dirty = true
	utils.MLogger.Infof("Add data to object: %s in bucket: %s end, length is: %d", object.GetInfo().GetName(), bucket.GetName(), opart.Length)
	return nil

}

func (l *LfsInfo) getBucketInfo(bucketName string) (*superBucket, error) {
	if !l.Online() || l.meta.buckets == nil {
		return nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, ErrLfsReadOnly
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	return bucket, nil
}

func (l *LfsInfo) getBucketAndObjectInfo(bucketName, objectName string) (*superBucket, *ObjectInfo, error) {
	if !l.Online() || l.meta.buckets == nil {
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

	objectElement := bucket.Objects.Find(MetaName(objectName))
	if objectElement != nil {
		return bucket, objectElement.(*ObjectInfo), nil
	}

	return nil, nil, ErrObjectNotExist
}

// make sure bucket and object is ont empty
func (l *LfsInfo) addObjectData(ctx context.Context, bucket *superBucket, object *ObjectInfo, reader io.Reader) (*ObjectInfo, *mpb.ObjectPart, error) {
	if object.Info.Dir {
		return object, nil, ErrObjectIsDir
	}

	bopt := &mpb.BlockOptions{
		Bopts:   bucket.BOpts,
		Start:   0,
		UserID:  l.userID,
		QueryID: l.fsID,
	}

	encoder, err := dataformat.NewDataCoderWithPrefix(l.keySet, bopt)
	if err != nil {
		return nil, nil, err
	}

	//append Object
	opart := &mpb.ObjectPart{
		Name:     object.GetInfo().GetName(),
		ObjectID: object.GetInfo().GetObjectID(),
		PartID:   object.GetPartCount(),
		Start:    bucket.GetLength(),
		CTime:    time.Now().Unix(),
	}

	ul := &uploadTask{
		startTime:       time.Now(), // for queue?
		reader:          reader,
		gInfo:           l.gInfo,
		bucketID:        bucket.BucketID,
		begin:           bucket.GetLength(),
		encoder:         encoder,
		encrypt:         bucket.BOpts.Encryption,
		taskWorkerCount: defaultTaskWorkerCount,
	}

	if bucket.BOpts.Encryption == 1 {
		ul.sKey = aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.GetInfo().GetObjectID())
	}

	err = ul.Start(ctx)
	if err != nil {
		utils.MLogger.Infof("Add data to object: %s in bucket: %s fails %s", object.GetInfo().GetName(), bucket.GetName(), err)
		return object, opart, err
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
		return object, opart, ErrUpload
	}

	// opart
	opart.ETag = ul.etag
	opart.Length = int64(ul.length)

	// bucket
	bucket.Length += int64(ul.rawLen)

	return object, opart, nil
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
	//计算一些参数
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

	//获取存储矿工
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

	//加密设置
	var bEnc cipher.BlockMode
	if u.encrypt == 1 {
		tmpEnc, err := aes.ContructAesEnc(u.sKey[:])
		if err != nil {
			return err
		}
		bEnc = tmpEnc
	}

	//分片传输设置
	tn := os.Getenv("MEFS_TRANS")
	utils.MLogger.Infof("MEFS_TRANS in upload is set to %s", tn)
	if tn != "" {
		tNum, err := strconv.Atoi(tn)
		if err != nil {
			transNum = defaultTransNum
		} else {
			transNum = tNum
		}
	} else {
		transNum = defaultTransNum
	}

	h := md5.New()
	rdata := make([]byte, stripeSize)
	// var extra []byte
	errc := make(chan error)
	var errrt error
	breakFlag = false
	//减少内存分配
	poolbuf := new(pool.BufferPool)

	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("upload cancel")
			return nil
		case errrt = <-errc:
			utils.MLogger.Warn("upload encounter an err:", errrt)
			return errrt
		default:
			// clear itself
			data := poolbuf.Get(stripeSize)
			data = data[:0]
			readLen := stripeSize - int(curOffset)*segStripeSize

			utils.MLogger.Debugf("Upload object: stripe: %d, seg offset: %d, expected length: %d", curStripe, curOffset, readLen)

			n, err := io.ReadAtLeast(u.reader, rdata[:readLen], readLen)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				breakFlag = true
			} else if err != nil {
				return err
			} else if n != readLen {
				return ErrUpload
			}

			data = append(data, rdata[:n]...)
			if len(data) == 0 {
				break
			}

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

			// transfer to different providers
			newpro := utils.DisorderArray(pros)

			bm, _ := metainfo.NewBlockMeta(u.gInfo.groupID, strconv.Itoa(int(u.bucketID)), strconv.Itoa(curStripe), "0")

			blockMetas := make([]blockMeta, bc)
			for i := 0; i < bc; i++ {
				bm.SetCid(strconv.Itoa(i))
				ncid := bm.ToString()
				blockMetas[i].cid = ncid
				if i < len(newpro) {
					blockMetas[i].provider = newpro[i]
				} else {
					blockMetas[i].provider = u.gInfo.groupID
				}

				if curOffset != 0 {
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

				encodedData, offset, err := enc.Encode(transData, bm.ToString(3), curOffset)
				if err != nil {
					return err
				}

				count = 0
				var pwg sync.WaitGroup
				for i := 0; i < bc; i++ {
					blockMetas[i].end = int(offset)
					blockMetas[i].start = int(curOffset)
					provider := blockMetas[i].provider
					if provider == u.gInfo.groupID {
						continue
					}
					pwg.Add(1)
					if curOffset == 0 {
						go func(num int, edata []byte, proID string) {
							defer pwg.Done()
							km, _ := metainfo.NewKey(blockMetas[num].cid, mpb.KeyType_Block)
							for k := 0; k < 10; k++ {
								err := u.gInfo.ds.PutBlock(ctx, km.ToString(), edata, proID)
								if err != nil {
									utils.MLogger.Warn("Put Block: ", km.ToString(), " to: ", proID, "  failed: ", err)
									if _, success := u.gInfo.ds.Connect(ctx, proID); success {
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
									if _, success := u.gInfo.ds.Connect(ctx, proID); success {
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
				curOffset = offset
				if count >= dc {
					atomic.AddInt64(&u.sucLen, int64(len(transData)))
				}
			}
			if count >= dc {
				for _, v := range blockMetas {
					err := u.gInfo.putDataMetaToKeepers(ctx, v.cid, v.provider, v.end)
					if err != nil {
						return err
					}
				}
				poolbuf.Put(data)
			} else {
				return ErrNoEnoughBlockUpload
			}

			if endOffset == int(u.encoder.Prefix.Bopts.GetSegmentCount()) { //如果写满了一个stripe
				curStripe++
				curOffset = 0
			} else {
				curOffset = endOffset
			}
		}
	}

	u.etag = hex.EncodeToString(h.Sum(nil))

	return errrt
}

func calculateETag(ob *ObjectInfo) string {
	if len(ob.GetParts()) == 0 {
		return ""
	}
	if len(ob.GetParts()) == 1 {
		return ob.GetParts()[0].ETag
	}

	result, err := hex.DecodeString(ob.GetParts()[0].ETag)
	if err != nil {
		return ""
	}
	for i := 0; i < len(ob.GetParts()); i++ {
		temp, err := hex.DecodeString(ob.GetParts()[i].ETag)
		if err != nil {
			continue
		}
		err = xor(result, temp)
		if err != nil {
			return ""
		}
	}

	return hex.EncodeToString(result)
}

func calculateETagForNewPart(old, new string) string {
	var oldBytes, newBytes []byte
	var err error
	if len(old) == 0 {
		return new
	} else {
		oldBytes, err = hex.DecodeString(old)
		if err != nil {
			return ""
		}
	}
	if len(new) == 0 {
		return old
	} else {
		newBytes, err = hex.DecodeString(new)
		if err != nil {
			return ""
		}
	}
	err = xor(oldBytes, newBytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(oldBytes)
}

func xor(a []byte, b []byte) error {
	if len(a) != len(b) {
		return ErrWrongParameters
	}
	for i := 0; i < len(a); i++ {
		a[i] ^= b[i]
	}
	return nil
}
