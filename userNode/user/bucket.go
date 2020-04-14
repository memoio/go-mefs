package user

import (
	"context"
	"crypto/sha256"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	rbtree "github.com/memoio/go-mefs/utils/RbTree"
	"github.com/memoio/go-mefs/utils/metainfo"
	mt "gitlab.com/NebulousLabs/merkletree"
)

func newsuperBucket(binfo mpb.BucketInfo, dirty bool) *superBucket {
	return &superBucket{
		BucketInfo:  binfo,
		dirty:       dirty,
		Objects:     rbtree.NewTree(),
		obMetaCache: make([]byte, maxCacheSize),
		obCacheSize: 0,
		mtree:       mt.New(sha256.New()),
	}
}

// CreateBucket create a bucket for a specified LFSservice
func (l *LfsInfo) CreateBucket(ctx context.Context, bucketName string, options *mpb.BucketOptions) (*mpb.BucketInfo, error) {
	//操作需要1资源
	ok := l.Sm.TryAcquire(1)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(1)
	// TODO judge datacount + parity count <= providers
	if !l.Online() || l.meta.buckets == nil {
		return nil, ErrLfsServiceNotReady
	}

	if !l.writable {
		return nil, ErrLfsReadOnly
	}

	// 多副本策略
	switch options.Policy {
	case dataformat.MulPolicy:
		Sum := options.DataCount + options.ParityCount
		options.DataCount = 1
		options.ParityCount = Sum - 1
	case dataformat.RsPolicy:
	default:
		return nil, ErrPolicy
	}

	// datacount + parityCount should <= providerSLA
	if options.DataCount+options.ParityCount > int32(l.gInfo.providerSLA) {
		utils.MLogger.Errorf("data count and parity count is too large, should not large than provider number %d", l.gInfo.providerSLA)
		return nil, ErrPolicy
	}

	// tagCount决定了最大的segmentSize
	if int(options.GetSegmentSize()) > 32*l.keySet.Pk.TagCount {
		utils.MLogger.Errorf("segmentSize is set large than: %d", 32*l.keySet.Pk.TagCount)
		return nil, ErrPolicy
	}

	err := checkBucketName(bucketName)
	if err != nil {
		utils.MLogger.Errorf("bucketName %s is not valid %s", bucketName, err)
		return nil, ErrBucketNameInvalid
	}

	l.meta.sb.Lock()

	if _, ok := l.meta.buckets[bucketName]; ok {
		l.meta.sb.Unlock()
		return nil, ErrBucketAlreadyExist
	}

	utils.MLogger.Infof("create bucket %s in lfs %s", bucketName, l.fsID)

	bucketID := l.meta.sb.NextBucketID
	binfo := mpb.BucketInfo{
		Name:         bucketName,
		BucketID:     bucketID,
		BOpts:        options,
		Length:       0,
		CTime:        time.Now().Unix(),
		Deletion:     false,
		NextObjectID: 0,
		NextOpID:     0,
	}

	bucket := newsuperBucket(binfo, true)

	bucket.mtree.SetIndex(0)
	bucket.mtree.Push([]byte(l.fsID + bucketName))

	//将此Bucket信息添加到LFS中
	l.meta.sb.NextBucketID++
	l.meta.sb.dirty = true

	l.meta.buckets[bucket.Name] = bucket
	l.meta.bucketIDToName[bucketID] = bucketName
	l.meta.sb.Unlock()
	l.meta.dirty = true

	bk, _ := metainfo.NewKey(l.fsID, mpb.KeyType_Bucket, l.userID, strconv.FormatInt(bucketID, 10))

	val, err := proto.Marshal(&binfo)
	if err == nil {
		l.gInfo.putDataToKeepers(ctx, bk.ToString(), val)
	}

	return &bucket.BucketInfo, nil
}

// DeleteBucket deletes a bucket from a specified LFSservice
func (l *LfsInfo) DeleteBucket(ctx context.Context, bucketName string) (*mpb.BucketInfo, error) {
	//操作需要1资源
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

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	l.meta.sb.Lock()
	defer l.meta.sb.Unlock()
	bucket.Lock()
	bucket.Deletion = true
	delete(l.meta.buckets, bucket.Name)
	delete(l.meta.bucketIDToName, bucket.GetBucketID())
	l.meta.deletedBuckets = append(l.meta.deletedBuckets, bucket)
	bucket.dirty = true
	bucket.Unlock()
	l.meta.dirty = true

	bk, _ := metainfo.NewKey(l.fsID, mpb.KeyType_Bucket, l.userID, strconv.FormatInt(bucket.GetBucketID(), 10))

	val, err := proto.Marshal(&bucket.BucketInfo)
	if err == nil {
		l.gInfo.putDataToKeepers(ctx, bk.ToString(), val)
	}

	return &bucket.BucketInfo, nil
}

// HeadBucket get a superBucket's metainfo
func (l *LfsInfo) HeadBucket(ctx context.Context, bucketName string) (*mpb.BucketInfo, error) {
	//操作需要1资源
	ok := l.Sm.TryAcquire(1)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(1)
	if !l.Online() || l.meta.buckets == nil {
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
	return &bucket.BucketInfo, nil
}

// ListBuckets lists all Buckets information
func (l *LfsInfo) ListBuckets(ctx context.Context, prefix string) ([]*mpb.BucketInfo, error) {
	//操作需要2资源
	ok := l.Sm.TryAcquire(2)
	if !ok {
		return nil, ErrResourceUnavailable
	}
	defer l.Sm.Release(2)
	if !l.Online() || l.meta.buckets == nil {
		return nil, ErrLfsServiceNotReady
	}

	var lsuperBucket []*mpb.BucketInfo
	for _, bs := range l.meta.buckets {
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
