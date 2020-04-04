package user

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// DeleteObject deletes a object in lfs
func (l *LfsInfo) DeleteObject(ctx context.Context, bucketName, objectName string) (*mpb.ObjectInfo, error) {
	//????1??
	ok := l.Sm.TryAcquire(1)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(1)
	if !l.Online() || l.meta.buckets == nil {
		return nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, ErrLfsReadOnly
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	err = checkObjectName(objectName)
	if err != nil {
		return nil, ErrObjectNameInvalid
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	// TODO:具体实现
	bucket.Lock()
	defer bucket.Unlock()
	if bucket.Objects == nil {
		return nil, ErrObjectNotExist
	}

	object, ok := bucket.Objects.Find(MetaName(objectName)).(*ObjectInfo)
	if !ok {
		return nil, ErrObjectNotExist
	}

	object.Lock()
	defer object.Unlock()

	deleteObject := mpb.DeleteObject{
		Name:     object.GetInfo().GetName(),
		ObjectID: object.GetInfo().GetObjectID(),
		Time:     time.Now().Unix(),
	}

	payload, _ := proto.Marshal(&deleteObject)
	op := &mpb.OpRecord{
		OpType:  mpb.LfsOp_OpDelete,
		OpID:    bucket.GetNextOpID(),
		Payload: payload,
	}

	// leaf is OpID + PayLoad
	tag, err := proto.Marshal(op)
	if err != nil {
		return nil, err
	}
	bucket.mtree.Push(tag)
	bucket.Root = bucket.mtree.Root()

	l.flushObjectMeta(bucket, false, op)
	bucket.NextOpID++

	object.Deletion = true
	bucket.dirty = true
	bucket.Objects.Delete(MetaName(object.GetInfo().GetName()))
	bucket.DeletedObject = append(bucket.DeletedObject, object)
	return &object.ObjectInfo, nil
}

// HeadObject get the info of an object
func (l *LfsInfo) HeadObject(ctx context.Context, bucketName, objectName string) (*mpb.ObjectInfo, error) {
	//需要1资源
	ok := l.Sm.TryAcquire(1)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(1)
	if l.meta.buckets == nil { //只读不需要Online
		return nil, ErrLfsServiceNotReady
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	err = checkObjectName(objectName)
	if err != nil {
		return nil, ErrObjectNameInvalid
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	if bucket.Objects == nil {
		return nil, ErrObjectNotExist
	}

	object, ok := bucket.Objects.Find(MetaName(objectName)).(*ObjectInfo)
	if !ok || bucket.Deletion {
		return nil, ErrObjectNotExist
	}

	return &object.ObjectInfo, nil
}

// ListObjects lists all objects of a bucket
func (l *LfsInfo) ListObjects(ctx context.Context, bucketName, prefix string, opts ListObjectsOptions) ([]*mpb.ObjectInfo, error) {
	//????2??
	ok := l.Sm.TryAcquire(2)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(2) //只读不需要Online
	if l.meta.buckets == nil {
		return nil, ErrLfsServiceNotReady
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	bucket.RLock()
	defer bucket.RUnlock()
	var objects []*mpb.ObjectInfo
	objectIter := bucket.Objects.Iterator()
	for objectIter != nil {
		object := objectIter.Value.(*ObjectInfo)
		if object.Deletion {
			continue
		}

		if strings.HasPrefix(object.GetInfo().GetName(), prefix) {
			objects = append(objects, &object.ObjectInfo)
		}
		objectIter = objectIter.Next()
	}
	return objects, nil
}

func (l *LfsInfo) GetsuperBucket(ctx context.Context, bucketName string) (*superBucket, error) {
	if l.meta.buckets == nil { //只读不需要Online
		return nil, ErrLfsServiceNotReady
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	return bucket, nil
}

// ShowStorage show lfs used space without appointed bucket
func (l *LfsInfo) ShowStorage(ctx context.Context) (uint64, error) {
	//????2??
	ok := l.Sm.TryAcquire(2)
	if !ok {
		return 0, ErrResourceUnavailable
	}
	defer l.Sm.Release(2)
	if l.meta.buckets == nil { //只读不需要Online
		return 0, ErrLfsServiceNotReady
	}

	var storageSpace uint64
	for _, bucket := range l.meta.buckets {
		bucketStorage, err := l.ShowBucketStorage(ctx, bucket.Name)
		if err != nil {
			continue
		}
		storageSpace += uint64(bucketStorage)
	}

	return storageSpace, nil
}

// ShowBucketStorage show lfs used spaceBucket
func (l *LfsInfo) ShowBucketStorage(ctx context.Context, bucketName string) (uint64, error) {
	if l.meta.buckets == nil { //只读不需要Online
		return 0, ErrLfsServiceNotReady
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return 0, ErrBucketNameInvalid
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return 0, ErrBucketNotExist
	}
	bucket.RLock()
	defer bucket.RUnlock()
	var storageSpace uint64
	objectIter := bucket.Objects.Iterator()
	for ; objectIter != nil; objectIter = objectIter.Next() {
		object := objectIter.Value.(*ObjectInfo)
		if object.Deletion {
			continue
		}
		storageSpace += uint64(object.GetLength())
	}
	return storageSpace, nil
}

func (l *LfsInfo) getLastChalTime(ctx context.Context, blockID string) (time.Time, error) {
	latestTime := time.Unix(0, 0)
	gp := l.gInfo
	conkeepers, _, err := gp.GetKeepers(l.context, -1)
	if err != nil {
		return latestTime, err
	}
	if len(conkeepers) == 0 {
		return latestTime, ErrNoKeepers
	}

	km, err := metainfo.NewKey(blockID, mpb.KeyType_ChalTime)
	if err != nil {
		return latestTime, err
	}

	var tempTime time.Time
	for _, keeper := range conkeepers {
		res, err := l.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), km.ToString(), nil, nil, keeper)
		if err != nil {
			continue
		}
		unixTime := utils.StringToUnix(string(res))
		tempTime = time.Unix(unixTime, 0)
		if tempTime.After(latestTime) {
			latestTime = tempTime
		}
	}
	return latestTime, err
}

// GetObjectAvailTime get available time of objects
func (l *LfsInfo) GetObjectAvailTime(ctx context.Context, object *mpb.ObjectInfo) (string, error) {
	//????2??
	ok := l.Sm.TryAcquire(2)
	if !ok {
		return "", ErrResourceUnavailable
	}
	defer l.Sm.Release(2)

	if len(object.Parts) == 0 {
		return time.Unix(object.GetCTime(), 0).Format(utils.BASETIME), nil
	}
	latestTime := time.Unix(0, 0)
	bucketName, ok := l.meta.bucketIDToName[object.GetInfo().GetBucketID()]
	if !ok {
		return latestTime.Format(utils.BASETIME), ErrBucketNotExist
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok {
		return latestTime.Format(utils.BASETIME), ErrBucketNotExist
	}

	blockCount := bucket.BOpts.DataCount + bucket.BOpts.ParityCount

	bo := bucket.BOpts

	stripeID := object.Parts[0].GetStart() / int64(bo.SegmentCount*bo.SegmentSize*bo.DataCount)

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(object.GetInfo().GetBucketID())), strconv.Itoa(int(stripeID)), "")
	if err != nil {
		return "", err
	}
	for i := 0; i < int(blockCount); i++ {
		bm.SetCid(strconv.Itoa(i))
		blockID := bm.ToString()
		blockAvailTime, err := l.getLastChalTime(ctx, blockID)
		if err != nil {
			utils.MLogger.Warn("Get block: %s's availTime failed: %s", blockID, err)
			continue
		}
		if blockAvailTime.After(latestTime) {
			latestTime = blockAvailTime
		}
	}
	return latestTime.Format(utils.BASETIME), nil
}
