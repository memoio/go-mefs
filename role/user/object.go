package user

import (
	"context"
	"strconv"
	"strings"
	"time"

	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// ObjectOptions is
type ObjectOptions struct {
	UserDefined map[string]string
}

// DeleteObject deletes a object in lfs
func (l *LfsInfo) DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error) {
	if !l.online {
		return nil, ErrLfsServiceNotReady
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
	// TODO:具体实现
	bucket.Lock()
	defer bucket.Unlock()
	if bucket.objects == nil {
		return nil, ErrObjectNotExist
	}
	objectElement, ok := bucket.objects[objectName]
	if !ok || objectElement == nil {
		return nil, ErrObjectNotExist
	}
	object, ok := objectElement.Value.(*objectInfo)
	if !ok {
		return nil, ErrObjectNotExist
	}

	delete(bucket.objects, objectName)
	object.Deletion = true
	// move deletions to special name
	object.Name = objectName + "/" + time.Now().Format(utils.BASETIME)
	bucket.objects[object.Name] = objectElement
	bucket.dirty = true
	return &object.ObjectInfo, nil
}

// HeadObject get the info of an object
func (l *LfsInfo) HeadObject(bucketName, objectName string, opts ObjectOptions) (*pb.ObjectInfo, error) {
	if !l.online {
		return nil, ErrLfsServiceNotReady
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
	// TODO:具体实现
	if bucket.objects == nil {
		return nil, ErrObjectNotExist
	}
	objectElement, ok := bucket.objects[objectName]
	if !ok || objectElement == nil {
		return nil, ErrObjectNotExist
	}
	object, ok := objectElement.Value.(*objectInfo)
	if !ok {
		return nil, ErrObjectNotExist
	}
	//var AvailTime string
	//if avail {
	//	AvailTime, _ = l.GetObjectAvailTime(object)
	//}
	return &object.ObjectInfo, nil
}

// ListObjects lists all objects of a bucket
func (l *LfsInfo) ListObjects(bucketName, prefix string, opts ObjectOptions) ([]*pb.ObjectInfo, error) {
	if !l.online {
		return nil, ErrLfsServiceNotReady
	}

	if l.meta.bucketNameToID == nil {
		return nil, ErrObjectNotExist
	}
	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	var objects []*pb.ObjectInfo
	for objectElement := bucket.orderedObjects.Front(); objectElement != nil; objectElement = objectElement.Next() {
		if objectElement == nil {
			continue
		}
		object, ok := objectElement.Value.(*objectInfo)
		if !ok || object.Deletion {
			continue
		}
		//if avail {
		//	if strings.HasPrefix(object.Name, pre) {
		//		objects = append(objects, &object.ObjectInfo)
		//		availTime, _ := l.GetObjectAvailTime(object)
		//		availTimes = append(availTimes, availTime)
		//	}
		if strings.HasPrefix(object.Name, prefix) {
			objects = append(objects, &object.ObjectInfo)
		}
	}
	return objects, nil
}

// ShowStorage show lfs used space without appointed bucket
func (l *LfsInfo) ShowStorage() (uint64, error) {
	if !l.online {
		return 0, ErrLfsServiceNotReady
	}

	if l.meta.bucketNameToID == nil {
		return 0, ErrBucketNotExist
	}

	var storageSpace uint64
	for _, bucket := range l.meta.bucketByID {

		for _, objectElement := range bucket.objects {
			if objectElement == nil {
				continue
			}

			object, ok := objectElement.Value.(*objectInfo)
			if !ok || object.Deletion {
				continue
			}

			storageSpace += uint64(object.GetSize())
		}
	}

	return storageSpace, nil
}

// ShowBucketStorage show lfs used spaceBucket
func (l *LfsInfo) ShowBucketStorage(bucketName string) (uint64, error) {
	if !l.online {
		return 0, ErrLfsServiceNotReady
	}

	if l.meta.bucketNameToID == nil {
		return 0, ErrBucketNotExist
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return 0, ErrBucketNotExist
	}
	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return 0, ErrBucketNotExist
	}
	var storageSpace uint64
	for _, objectElement := range bucket.objects {
		if objectElement == nil {
			continue
		}
		object, ok := objectElement.Value.(*objectInfo)
		if !ok || object.Deletion {
			continue
		}
		storageSpace += uint64(object.GetSize())
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

	km, err := metainfo.NewKeyMeta(blockID, metainfo.ChalTime)
	if err != nil {
		return latestTime, err
	}

	var tempTime time.Time
	ctx := context.Background()
	for _, keeper := range conkeepers {
		res, err := l.ds.SendMetaRequest(ctx, int32(metainfo.Get), km.ToString(), nil, nil, keeper)
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
func (l *LfsInfo) GetObjectAvailTime(object *pb.ObjectInfo) (string, error) {
	latestTime := time.Unix(0, 0)
	bucket := l.meta.bucketByID[object.BucketID]
	blockCount := bucket.DataCount + bucket.ParityCount
	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(object.BucketID)), strconv.Itoa(int(object.StripeStart)), "")
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
