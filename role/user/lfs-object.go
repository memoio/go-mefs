package user

import (
	"log"
	"strconv"
	"strings"
	"time"

	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

// DeleteObject deletes a object in lfs
func (l *lfsInfo) DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error) {
	err := IsOnline(l.userid)
	if err != nil {
		return nil, err
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
func (l *lfsInfo) HeadObject(bucketName, objectName string, avail bool) (*pb.ObjectInfo, string, error) {
	err := IsOnline(l.userid)
	if err != nil {
		return nil, "", err
	}
	if l.meta.bucketNameToID == nil {
		return nil, "", ErrBucketNotExist
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, "", ErrBucketNotExist
	}
	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, "", ErrBucketNotExist
	}
	// TODO:具体实现
	if bucket.objects == nil {
		return nil, "", ErrObjectNotExist
	}
	objectElement, ok := bucket.objects[objectName]
	if !ok || objectElement == nil {
		return nil, "", ErrObjectNotExist
	}
	object, ok := objectElement.Value.(*objectInfo)
	if !ok {
		return nil, "", ErrObjectNotExist
	}
	var AvailTime string
	if avail {
		AvailTime, _ = l.GetObjectAvailTime(object)
	}
	return &object.ObjectInfo, AvailTime, nil
}

// ListObject lists all objects of a bucket
func (l *lfsInfo) ListObject(bucketName, pre string, avail bool) ([]*pb.ObjectInfo, []string, error) {
	err := IsOnline(l.userid)
	if err != nil {
		return nil, nil, err
	}
	if l.meta.bucketNameToID == nil {
		return nil, nil, ErrObjectNotExist
	}
	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, nil, ErrBucketNotExist
	}
	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, nil, ErrBucketNotExist
	}
	var objects []*pb.ObjectInfo
	var availTimes []string
	for objectElement := bucket.orderedObjects.Front(); objectElement != nil; objectElement = objectElement.Next() {
		if objectElement == nil {
			continue
		}
		object, ok := objectElement.Value.(*objectInfo)
		if !ok || object.Deletion {
			continue
		}
		if avail {
			if strings.HasPrefix(object.Name, pre) {
				objects = append(objects, &object.ObjectInfo)
				availTime, _ := l.GetObjectAvailTime(object)
				availTimes = append(availTimes, availTime)
			}
		} else {
			if strings.HasPrefix(object.Name, pre) {
				objects = append(objects, &object.ObjectInfo)
			}
		}
	}
	return objects, availTimes, nil
}

// ShowStorageSpaceAll show lfs used space without appointed bucket
func (l *lfsInfo) ShowStorageSpaceAll() ([]uint64, error) {
	err := IsOnline(l.userid)
	if err != nil {
		return nil, err
	}

	if l.meta.bucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	storageSpaceAll := make([]uint64, 0, len(l.meta.bucketByID))
	for _, bucket := range l.meta.bucketByID {
		var storageSpace int64

		for _, objectElement := range bucket.objects {
			if objectElement == nil {
				continue
			}

			object, ok := objectElement.Value.(*objectInfo)
			if !ok || object.Deletion {
				continue
			}

			storageSpace += object.GetSize()
		}

		storageSpaceAll = append(storageSpaceAll, uint64(storageSpace))
	}

	return storageSpaceAll, nil
}

// ShowStorageSpace show lfs used space
func (l *lfsInfo) ShowStorageSpace(bucketName, pre string) (int, error) {
	err := IsOnline(l.userid)
	if err != nil {
		return 0, err
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
	var storageSpace int
	for _, objectElement := range bucket.objects {
		if objectElement == nil {
			continue
		}
		object, ok := objectElement.Value.(*objectInfo)
		if !ok || object.Deletion {
			continue
		}
		storageSpace += int(object.GetSize())
	}
	return storageSpace, nil
}

func (l *lfsInfo) getLastChalTime(blockID string) (time.Time, error) {
	latestTime := time.Unix(0, 0)
	gp := getGroup(l.userid)
	_, conkeepers, err := gp.getKeepers(-1)
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

// GetObjectAvailTime get available time of objects
func (l *lfsInfo) GetObjectAvailTime(object *objectInfo) (string, error) {
	latestTime := time.Unix(0, 0)
	bucket := l.meta.bucketByID[object.BucketID]
	blockCount := bucket.DataCount + bucket.ParityCount
	bm, err := metainfo.NewBlockMeta(l.userid, strconv.Itoa(int(object.BucketID)), strconv.Itoa(int(object.StripeStart)), "")
	if err != nil {
		return "", err
	}
	for i := 0; i < int(blockCount); i++ {
		bm.SetBid(strconv.Itoa(i))
		blockID := bm.ToString()
		blockAvailTime, err := l.getLastChalTime(blockID)
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
