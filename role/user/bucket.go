package user

import (
	"context"
	"crypto/sha256"
	"sort"
	"strings"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	mt "gitlab.com/NebulousLabs/merkletree"
)

// CreateBucket create a bucket for a specified LFSservice
func (l *LfsInfo) CreateBucket(ctx context.Context, bucketName string, options *mpb.BucketOptions) (*mpb.BucketInfo, error) {
	// TODO judge datacount + parity count <= providers
	if !l.online || l.meta.bucketNameToID == nil {
		return nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, ErrReadOnly
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	l.meta.sb.Lock()
	defer l.meta.sb.Unlock()

	if _, ok := l.meta.bucketNameToID[bucketName]; ok {
		return nil, ErrBucketAlreadyExist
	}

	utils.MLogger.Infof("create bucket %s in lfs %s", bucketName, l.fsID)

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

	bucketID := l.meta.sb.NextBucketID

	objects := make(map[string]*objectInfo)
	bucket := &superBucket{
		BucketInfo: mpb.BucketInfo{
			Name:         bucketName,
			BucketID:     bucketID,
			BOpts:        options,
			CurStripe:    0,
			NextSeg:      0,
			Ctime:        time.Now().Unix(),
			Deletion:     false,
			NextObjectID: 0,
		},
		dirty:   true,
		objects: objects,
		mtree:   mt.New(sha256.New()),
	}

	bucket.mtree.SetIndex(0)
	bucket.mtree.Push([]byte(l.fsID + bucketName))

	//将此Bucket信息添加到LFS中
	l.meta.sb.NextBucketID++
	l.meta.sb.bitsetInfo.Set(uint(bucketID))
	l.meta.sb.dirty = true

	l.meta.bucketByID[bucket.BucketID] = bucket
	l.meta.bucketNameToID[bucket.Name] = bucket.BucketID
	return &bucket.BucketInfo, nil
}

// DeleteBucket deletes a bucket from a specified LFSservice
func (l *LfsInfo) DeleteBucket(ctx context.Context, bucketName string) (*mpb.BucketInfo, error) {
	if !l.online || l.meta.bucketNameToID == nil {
		return nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, ErrReadOnly
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}

	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	l.meta.sb.Lock()
	bucket.Lock()
	bucket.Deletion = true
	delete(l.meta.bucketNameToID, bucket.Name)
	bucket.dirty = true
	bucket.Unlock()
	defer l.meta.sb.Unlock()
	return &bucket.BucketInfo, nil
}

// HeadBucket get a superBucket's metainfo
func (l *LfsInfo) HeadBucket(ctx context.Context, bucketName string) (*mpb.BucketInfo, error) {
	if !l.online || l.meta.bucketNameToID == nil {
		return nil, ErrLfsServiceNotReady
	}

	err := checkBucketName(bucketName)
	if err != nil {
		return nil, ErrBucketNameInvalid
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}

	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}
	return &bucket.BucketInfo, nil
}

// ListBuckets lists all Buckets information
func (l *LfsInfo) ListBuckets(ctx context.Context, prefix string) ([]*mpb.BucketInfo, error) {
	if !l.online {
		return nil, ErrLfsServiceNotReady
	}

	if l.meta.bucketByID == nil {
		return nil, ErrBucketNotExist
	}

	var lsuperBucket []*mpb.BucketInfo
	for _, bs := range l.meta.bucketByID {
		if bs.Deletion {
			continue
		}
		if strings.HasPrefix(bs.Name, prefix) {
			lsuperBucket = append(lsuperBucket, &bs.BucketInfo)
		}
	}

	sort.Slice(lsuperBucket, func(i, j int) bool {
		return lsuperBucket[i].Name < lsuperBucket[j].Name
	})
	return lsuperBucket, nil
}
