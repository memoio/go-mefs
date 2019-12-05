package user

import (
	"container/list"
	"strings"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
)

// CreateBucket create a bucket for a specified LFSservice
func (lfs *LfsService) CreateBucket(bucketName string, options *pb.BucketOptions) (*pb.BucketInfo, error) {
	// TODO judge datacount + parity count <= providers

	err := isStart(lfs.userid)
	if err != nil {
		return nil, err
	}
	if lfs.meta.bucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	err = CheckValidBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	if _, ok := lfs.meta.bucketNameToID[bucketName]; ok {
		return nil, ErrBucketAlreadyExist
	}
	lfs.meta.sb.Lock()
	defer lfs.meta.sb.Unlock()
	// 多副本策略
	switch options.Policy {
	case dataformat.MulPolicy:
		Sum := options.DataCount + options.ParityCount
		options.DataCount = 1
		options.ParityCount = Sum - 1
	case dataformat.RsPolicy:
	default:
		return nil, dataformat.ErrWrongPolicy
	}

	bucketID := lfs.meta.sb.NextBucketID

	objects := make(map[string]*list.Element)
	bucket := &superBucket{
		BucketInfo: pb.BucketInfo{
			Name:        bucketName,
			BucketID:    bucketID,
			Policy:      options.Policy,
			DataCount:   options.DataCount,
			ParityCount: options.ParityCount,
			CurStripe:   0,
			NextOffset:  0,
			Ctime:       time.Now().Format(utils.BASETIME),
			SegmentSize: options.SegmentSize,
			TagFlag:     options.TagFlag,
			Deletion:    false,
			Encryption:  options.Encryption,
		},
		dirty:          true,
		objects:        objects,
		orderedObjects: list.New(),
	}
	//将此Bucket信息添加到LFS中
	lfs.meta.sb.NextBucketID++
	lfs.meta.sb.bitsetInfo.Set(uint(bucketID))
	lfs.meta.sb.dirty = true

	lfs.meta.bucketByID[bucket.BucketID] = bucket
	lfs.meta.bucketNameToID[bucket.Name] = bucket.BucketID
	return &bucket.BucketInfo, nil
}

// DeleteBucket deletes a bucket from a specified LFSservice
func (lfs *LfsService) DeleteBucket(bucketName string) (*pb.BucketInfo, error) {
	err := isStart(lfs.userid)
	if err != nil {
		return nil, err
	}
	if lfs.meta.bucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := lfs.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	if bucket.CurStripe > 0 || bucket.NextOffset > 0 {
		return nil, ErrBucketNotEmpty
	}
	bucket.Deletion = true
	bucket.dirty = true
	return &bucket.BucketInfo, nil
}

// HeadBucket get a superBucket's metainfo
func (lfs *LfsService) HeadBucket(bucketName string) (*pb.BucketInfo, error) {
	err := isStart(lfs.userid)
	if err != nil {
		return nil, err
	}
	if lfs.meta.bucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := lfs.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	return &bucket.BucketInfo, nil
}

// ListBucket lists all superBucket in a lfsservice
func (lfs *LfsService) ListBucket(pre string) ([]*pb.BucketInfo, error) {
	err := isStart(lfs.userid)
	if err != nil {
		return nil, err
	}
	if lfs.meta.bucketByID == nil {
		return nil, ErrBucketNotExist
	}
	var lsuperBucket []*pb.BucketInfo
	for _, bs := range lfs.meta.bucketByID {
		if bs.Deletion {
			continue
		}
		if strings.HasPrefix(bs.Name, pre) {
			lsuperBucket = append(lsuperBucket, &bs.BucketInfo)
		}
	}
	return lsuperBucket, nil
}
