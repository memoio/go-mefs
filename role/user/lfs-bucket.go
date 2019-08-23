package user

import (
	"strings"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
)

//Create a bucket for a specified LFSservice
func (lfs *LfsService) CreateBucket(bucketName string, policy int, dataCount, parityCount int) (*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; ok || bucket != nil {
		return nil, ErrBucketAlreadyExist
	}
	lfs.CurrentLog.SbMux.Lock()
	defer lfs.CurrentLog.SbMux.Unlock()
	// 多副本策略
	if policy == dataformat.MulPolicy {
		Sum := dataCount + parityCount
		dataCount = 1
		parityCount = Sum - 1
	}
	bucket := &pb.BucketInfo{
		BucketName:  bucketName,
		BucketID:    lfs.CurrentLog.Sb.NextBucketID,
		Policy:      int32(policy),
		DataCount:   int32(dataCount),
		ParityCount: int32(parityCount),
		CurStripe:   0,
		NextOffset:  0,
		Ctime:       time.Now().Format(utils.BASETIME),
		SegmentSize: dataformat.DefaultSegmentSize,
		TagFlag:     dataformat.BLS12,
		Deletion:    false,
	}
	//将此Bucket信息添加到LFS中
	lfs.CurrentLog.Sb.Buckets[bucket.BucketID] = bucket.BucketName
	lfs.CurrentLog.Sb.NextBucketID++
	lfs.CurrentLog.SbModified = true

	lfs.CurrentLog.BucketByID[bucket.BucketID] = bucket
	lfs.CurrentLog.BucketByName[bucket.BucketName] = bucket
	lfs.CurrentLog.Entries[bucket.BucketID] = make(map[string]*pb.ObjectInfo)
	lfs.CurrentLog.State[bucket.BucketID] = &BucketState{Dirty: true}
	return bucket, nil
}

//Delete a bucket from a specified LFSservice
func (lfs *LfsService) DeleteBucket(bucketName string) (*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketByName == nil {
		return nil, ErrBucketNotExist
	}

	bucket, ok := lfs.CurrentLog.BucketByName[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	if bucket.CurStripe > 0 || bucket.NextOffset > 0 {
		return nil, ErrBucketNotEmpty
	}
	bucket.Deletion = true
	lfs.CurrentLog.State[bucket.BucketID].Dirty = true
	return bucket, nil
}

//Get a Bucket's metainfo
func (lfs *LfsService) HeadBucket(bucketName string) (*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketByName == nil {
		return nil, ErrBucketNotExist
	}
	if bucket, ok := lfs.CurrentLog.BucketByName[bucketName]; ok && bucket != nil && !bucket.Deletion {
		return bucket, nil
	}
	return nil, ErrBucketNotExist
}

//list all buckets in a lfsservice
func (lfs *LfsService) ListBucket(pre string) ([]*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketByID == nil {
		return nil, ErrBucketNotExist
	}
	var buckets []*pb.BucketInfo
	for _, Bucket := range lfs.CurrentLog.BucketByID {
		if len(buckets) > MAXLISTVALUE {
			break
		}
		if Bucket.Deletion {
			continue
		}
		if strings.HasPrefix(Bucket.BucketName, pre) {
			buckets = append(buckets, Bucket)
		}
	}
	return buckets, nil
}
