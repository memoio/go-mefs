package user

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// ObjectOptions is
type ObjectOptions struct {
	UserDefined map[string]string
}

// DeleteObject deletes a object in lfs
func (l *LfsInfo) DeleteObject(ctx context.Context, bucketName, objectName string) (*mpb.ObjectInfo, error) {
	if !l.online || l.meta.buckets == nil {
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
	if bucket.objects == nil {
		return nil, ErrObjectNotExist
	}

	object, ok := bucket.objects[objectName]
	if !ok {
		return nil, ErrObjectNotExist
	}

	object.Lock()
	defer object.Unlock()

	object.Deletion = true
	bucket.dirty = true
	oName := object.Parts[0].GetName() + "." + strconv.Itoa(int(object.GetInfo().ObjectID))
	delete(bucket.objects, object.Parts[0].GetName())
	bucket.objects[oName] = object
	return &object.ObjectInfo, nil
}

// HeadObject get the info of an object
func (l *LfsInfo) HeadObject(ctx context.Context, bucketName, objectName string, opts ObjectOptions) (*mpb.ObjectInfo, error) {
	if !l.online || l.meta.buckets == nil {
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

	if bucket.objects == nil {
		return nil, ErrObjectNotExist
	}

	object, ok := bucket.objects[objectName]
	if !ok || bucket.Deletion {
		return nil, ErrObjectNotExist
	}

	return &object.ObjectInfo, nil
}

// ListObjects lists all objects of a bucket
func (l *LfsInfo) ListObjects(ctx context.Context, bucketName, prefix string, opts ObjectOptions) ([]*mpb.ObjectInfo, error) {
	if !l.online || l.meta.buckets == nil {
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

	var objects []*mpb.ObjectInfo
	for _, object := range bucket.objects {
		if object.Deletion {
			continue
		}

		if strings.HasPrefix(object.GetInfo().GetName(), prefix) {
			objects = append(objects, &object.ObjectInfo)
		}
	}
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].GetInfo().GetName() < objects[j].GetInfo().GetName()
	})
	return objects, nil
}

// ShowStorage show lfs used space without appointed bucket
func (l *LfsInfo) ShowStorage(ctx context.Context) (uint64, error) {
	if !l.online || l.meta.buckets == nil {
		return 0, ErrLfsServiceNotReady
	}

	var storageSpace uint64
	for _, bucket := range l.meta.buckets {
		bucket.RLock()
		for _, object := range bucket.objects {
			if object.Deletion {
				continue
			}

			storageSpace += uint64(object.GetLength())
		}
		bucket.RUnlock()
	}

	return storageSpace, nil
}

// ShowBucketStorage show lfs used spaceBucket
func (l *LfsInfo) ShowBucketStorage(ctx context.Context, bucketName string) (uint64, error) {
	if !l.online || l.meta.buckets == nil {
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

	var storageSpace uint64
	for _, object := range bucket.objects {
		if object.Deletion {
			continue
		}
		storageSpace += uint64(object.GetLength())
	}
	return storageSpace, nil
}

func (l *LfsInfo) getLastChalTime(blockID string) (time.Time, error) {
	latestTime := time.Unix(0, 0)
	gp := l.gInfo
	conkeepers, _, err := gp.GetKeepers(-1)
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
	ctx := context.Background()
	for _, keeper := range conkeepers {
		res, err := l.ds.SendMetaRequest(ctx, int32(mpb.OpType_Get), km.ToString(), nil, nil, keeper)
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

// GetObjectAvailTime get available time of objects
func (l *LfsInfo) GetObjectAvailTime(object *mpb.ObjectInfo) (string, error) {
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
		blockAvailTime, err := l.getLastChalTime(blockID)
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
