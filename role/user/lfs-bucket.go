package user

import (
	"strings"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
)

// CreateBucket create a bucket for a specified LFSservice
func (lfs *LfsService) CreateBucket(bucketName string, policy int, dataCount, parityCount int) (*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	if _, ok := lfs.CurrentLog.BucketNameToID[bucketName]; ok {
		return nil, ErrBucketAlreadyExist
	}
	lfs.CurrentLog.Sb.SbMux.Lock()
	defer lfs.CurrentLog.Sb.SbMux.Unlock()
	// 多副本策略
	switch policy {
	case dataformat.MulPolicy:
		Sum := dataCount + parityCount
		dataCount = 1
		parityCount = Sum - 1
	case dataformat.RsPolicy:
	default:
		return nil, dataformat.ErrWrongPolicy
	}

	bucketID := lfs.CurrentLog.Sb.NextBucketID

	objects := make(map[string]*Object)
	bucket := &Bucket{
		BucketInfo: pb.BucketInfo{
			BucketName:  bucketName,
			BucketID:    bucketID,
			Policy:      int32(policy),
			DataCount:   int32(dataCount),
			ParityCount: int32(parityCount),
			CurStripe:   0,
			NextOffset:  0,
			Ctime:       time.Now().Format(utils.BASETIME),
			SegmentSize: dataformat.DefaultSegmentSize,
			TagFlag:     dataformat.BLS12,
			Deletion:    false,
			Encryption:  true,
		},
		Dirty:   true,
		Objects: objects,
	}
	//将此Bucket信息添加到LFS中
	lfs.CurrentLog.Sb.NextBucketID++
	lfs.CurrentLog.Sb.Bitset.Set(uint(bucketID))
	lfs.CurrentLog.Sb.Dirty = true

	lfs.CurrentLog.BucketByID[bucket.BucketID] = bucket
	lfs.CurrentLog.BucketNameToID[bucket.BucketName] = bucket.BucketID
	return &bucket.BucketInfo, nil
}

// DeleteBucket deletes a bucket from a specified LFSservice
func (lfs *LfsService) DeleteBucket(bucketName string) (*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	if bucket.CurStripe > 0 || bucket.NextOffset > 0 {
		return nil, ErrBucketNotEmpty
	}
	bucket.Deletion = true
	bucket.Dirty = true
	return &bucket.BucketInfo, nil
}

// HeadBucket get a Bucket's metainfo
func (lfs *LfsService) HeadBucket(bucketName string) (*pb.BucketInfo, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	return &bucket.BucketInfo, nil
}

// ListBucket lists all buckets in a lfsservice
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
			buckets = append(buckets, &Bucket.BucketInfo)
		}
	}
	return buckets, nil
}
