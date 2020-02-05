package user

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"sort"
	"strconv"
	"strings"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	mt "gitlab.com/NebulousLabs/merkletree"
)

// CreateBucket create a bucket for a specified LFSservice
func (l *LfsInfo) CreateBucket(ctx context.Context, bucketName string, options *pb.BucketOptions) (*pb.BucketInfo, error) {
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

	objects := make(map[string]*list.Element)
	bucket := &superBucket{
		BucketInfo: pb.BucketInfo{
			Name:         bucketName,
			BucketID:     bucketID,
			BOpts:        options,
			CurStripe:    0,
			NextSeg:      0,
			Ctime:        time.Now().Unix(),
			Deletion:     false,
			NextObjectID: 0,
		},
		dirty:          true,
		objects:        objects,
		orderedObjects: list.New(),
		mtree:          mt.New(sha256.New()),
	}

	bucket.mtree.SetIndex(0)
	bucket.mtree.Push([]byte(bucketName))
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(bucket.Ctime))
	bucket.mtree.Push(buf)

	//将此Bucket信息添加到LFS中
	l.meta.sb.NextBucketID++
	l.meta.sb.bitsetInfo.Set(uint(bucketID))
	l.meta.sb.dirty = true

	l.meta.bucketByID[bucket.BucketID] = bucket
	l.meta.bucketNameToID[bucket.Name] = bucket.BucketID
	return &bucket.BucketInfo, nil
}

// DeleteBucket deletes a bucket from a specified LFSservice
func (l *LfsInfo) DeleteBucket(ctx context.Context, bucketName string) (*pb.BucketInfo, error) {
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
	bucket.Name = bucket.Name + "/" + strconv.Itoa(int(bucket.BucketID))
	l.meta.bucketNameToID[bucket.Name] = bucket.BucketID
	bucket.dirty = true
	bucket.Unlock()
	defer l.meta.sb.Unlock()
	return &bucket.BucketInfo, nil
}

// HeadBucket get a superBucket's metainfo
func (l *LfsInfo) HeadBucket(ctx context.Context, bucketName string) (*pb.BucketInfo, error) {
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
func (l *LfsInfo) ListBuckets(ctx context.Context, prefix string) ([]*pb.BucketInfo, error) {
	if !l.online {
		return nil, ErrLfsServiceNotReady
	}

	if l.meta.bucketByID == nil {
		return nil, ErrBucketNotExist
	}

	var lsuperBucket BucketsInfo
	for _, bs := range l.meta.bucketByID {
		if bs.Deletion {
			continue
		}
		if strings.HasPrefix(bs.Name, prefix) {
			lsuperBucket = append(lsuperBucket, &bs.BucketInfo)
		}
	}

	sort.Sort(lsuperBucket)
	return lsuperBucket, nil
}

type BucketsInfo []*pb.BucketInfo

func (b BucketsInfo) Len() int { // 重写 Len() 方法
	return len(b)
}
func (b BucketsInfo) Swap(i, j int) { // 重写 Swap() 方法
	b[i], b[j] = b[j], b[i]
}
func (b BucketsInfo) Less(i, j int) bool { // 重写 Less() 方法， 从大到小排序
	return b[j].Name < b[i].Name
}
