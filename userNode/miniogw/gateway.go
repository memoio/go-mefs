package miniogw

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/memoio/go-mefs/core"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/userNode/user"
	rbtree "github.com/memoio/go-mefs/utils/RbTree"
	"github.com/memoio/go-mefs/utils/address"

	"github.com/minio/cli"
	minio "github.com/minio/minio/cmd"
	"github.com/minio/minio/pkg/auth"
	"github.com/minio/minio/pkg/policy"
)

var (
	errLfsServiceNotReady   = errors.New("lfs service not ready")
	errNoObjectsToBeDeleted = errors.New("no objects to be deleted")
	errDeleteObjects        = errors.New("error(s) occurred while deleting objects")
)

// Start gateway
func Start(addr, pwd, endPoint string) error {
	minio.RegisterGatewayCommand(cli.Command{
		Name:            "lfs",
		Usage:           "Mefs Log File System Service (LFS)",
		Action:          mefsGatewayMain,
		HideHelpCommand: true,
	})

	err := os.Setenv("MINIO_ACCESS_KEY", addr)
	if err != nil {
		return err
	}
	err = os.Setenv("MINIO_SECRET_KEY", pwd)
	if err != nil {
		return err
	}

	rootpath, _ := fsrepo.BestKnownPath()
	gwConf := rootpath + "/gwConf"

	// ”memoriae“ is app name
	// "gateway" represents gatewat mode; respective, "server" represents server mode
	// "lfs" is subcommand, should equal to RegisterGatewayCommand{Name}
	go minio.Main([]string{"memoriae", "gateway", "lfs",
		"--address", endPoint, "--config-dir", gwConf})

	return nil
}

// Handler for 'minio gateway oss' command line.
func mefsGatewayMain(ctx *cli.Context) {
	minio.StartGateway(ctx, &Mefs{"lfs"})
}

// Mefs implements Lfs Gateway.
type Mefs struct {
	host string
}

// Name implements Gateway interface.
func (g *Mefs) Name() string {
	return "lfs"
}

// NewGatewayLayer implements Gateway interface and returns LFS ObjectLayer.
func (g *Mefs) NewGatewayLayer(creds auth.Credentials) (minio.ObjectLayer, error) {
	uid, err := address.GetIDFromAddress(creds.AccessKey)
	if err != nil {
		return nil, err
	}
	uploads := NewMultipartUploads()

	var lfs user.FileSyetem
	userIns, ok := core.LocalNode.Inst.(*user.Info)
	if !ok {
		log.Println("warn: please check user Instance before use gateway service")
	} else {
		lfs = userIns.GetUser(uid)
		if lfs == nil {
			log.Println("warn: please start lfs first to use gateway service")
		}
	}
	return &lfsGateway{
		userID:    uid,
		multipart: uploads,
		lfs:       lfs,
	}, nil
}

// Production - oss is production ready.
func (g *Mefs) Production() bool {
	return false
}

// lfsGateway implements gateway.
type lfsGateway struct {
	minio.GatewayUnsupported
	lfs       user.FileSyetem
	userID    string
	multipart *MultipartUploads
}

func (l *lfsGateway) checkLfs(ctx context.Context) error {
	userIns, ok := core.LocalNode.Inst.(*user.Info)
	if !ok {
		return errLfsServiceNotReady
	}
	lfs := userIns.GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return errLfsServiceNotReady
	}
	l.lfs = lfs
	return nil
}

// Shutdown saves any gateway metadata to disk
// if necessary and reload upon next restart.
func (l *lfsGateway) Shutdown(ctx context.Context) error {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}
	return l.lfs.Stop()
}

// StorageInfo is not relevant to LFS backend.
func (l *lfsGateway) StorageInfo(ctx context.Context) (si minio.StorageInfo) {
	si.Backend.Type = minio.BackendGateway
	si.Backend.GatewayOnline = true

	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return si
		}
	}

	use, _ := l.lfs.ShowStorage(ctx)
	si.Used = []uint64{use}
	return si
}

// MakeBucketWithLocation creates a new container on LFS backend.
func (l *lfsGateway) MakeBucketWithLocation(ctx context.Context, bucket, options string) error {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	bucketOptions := &mpb.BucketOptions{}
	err := json.Unmarshal([]byte(options), bucketOptions)
	if err != nil {
		bucketOptions = df.DefaultBucketOptions()
		lfsIns, ok := l.lfs.(*user.LfsInfo)
		if !ok {
			return convertToMinioError(errLfsServiceNotReady, "", "")
		}

		conpro, unconpro, err := lfsIns.GetGroup().GetProviders(ctx, -1)
		if err != nil {
			return err
		}

		proCount := int32(len(conpro) + len(unconpro))

		if proCount < (bucketOptions.GetDataCount() + bucketOptions.GetParityCount()) {
			bucketOptions.DataCount = proCount - 3
			bucketOptions.ParityCount = 2
		}
	}
	_, err = l.lfs.CreateBucket(ctx, bucket, bucketOptions)

	return convertToMinioError(err, bucket, "")
}

// GetBucketInfo gets bucket metadata.
func (l *lfsGateway) GetBucketInfo(ctx context.Context, bucket string) (bi minio.BucketInfo, err error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return bi, convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	bucketInfo, err := l.lfs.HeadBucket(ctx, bucket)
	if err != nil {
		return bi, convertToMinioError(err, bucket, "")
	}
	bi.Name = bucket
	bi.Created = time.Unix(bucketInfo.GetCTime(), 0).UTC()
	return bi, nil
}

// ListBuckets lists all LFS buckets.
func (l *lfsGateway) ListBuckets(ctx context.Context) (buckets []minio.BucketInfo, err error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return nil, user.ErrLfsServiceNotReady
		}
	}

	bucketsInfo, err := l.lfs.ListBuckets(ctx, "")
	if err != nil {
		return nil, convertToMinioError(err, "", "")
	}

	buckets = make([]minio.BucketInfo, len(bucketsInfo))
	for i, v := range bucketsInfo {
		buckets[i].Name = v.Name
		buckets[i].Created = time.Unix(v.GetCTime(), 0).UTC()
	}
	return buckets, nil
}

// DeleteBucket deletes a bucket on LFS.
func (l *lfsGateway) DeleteBucket(ctx context.Context, bucket string) error {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	_, err := l.lfs.DeleteBucket(ctx, bucket)
	return convertToMinioError(err, bucket, "")
}

// ListObjects lists all blobs in LFS bucket filtered by prefix.
func (l *lfsGateway) ListObjects(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (loi minio.ListObjectsInfo, err error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return loi, convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}
	lfsInfo := l.lfs.(*user.LfsInfo)
	sbucket, err := lfsInfo.GetsuperBucket(ctx, bucket)
	if err != nil {
		return loi, convertToMinioError(err, bucket, "")
	}

	//list_object需要2资源
	err = lfsInfo.Sm.Acquire(ctx, 2)
	if err != nil {
		return loi, convertToMinioError(err, "", "")
	}
	defer lfsInfo.Sm.Release(2)

	sbucket.RLock()
	//给Bucket解锁
	defer sbucket.RUnlock()

	objsTree := sbucket.Objects

	if maxKeys > user.MaxListKeys || maxKeys == 0 {
		maxKeys = user.MaxListKeys
	}
	if maxKeys > objsTree.Size() {
		maxKeys = objsTree.Size()
	}
	loi.Objects = make([]minio.ObjectInfo, 0, maxKeys)
	recursiveKey := ""
	entryPrefixMatch := prefix
	if len(delimiter) > 0 {
		lastIndex := strings.LastIndex(prefix, delimiter)
		if lastIndex > 0 && lastIndex < len(prefix) {
			entryPrefixMatch = prefix[:lastIndex+1]
		}
	}
	//没有delimiter的时候全返回（recursive），有的时候只返回一层（no-recursive）
	recursive := len(delimiter) <= 0
	prefixLen := len(entryPrefixMatch)
	objectIter := rbtree.NewNode()
	if marker == "" {
		objectIter = objsTree.Iterator()
	} else {
		objectIter = objsTree.FindIt(user.MetaName(marker))
		if objectIter != nil {
			objectIter = objectIter.Next()
		}
	}
	limit := maxKeys

	hasPrefixKey := false

	//用于过滤prefix
	first := true
	for ; limit > 0 && objectIter != nil; objectIter = objectIter.Next() {
		object := objectIter.Value.(*user.ObjectInfo)
		if object.Deletion {
			continue
		}
		//如果没读完，则需要接着进行
		if limit == 1 {
			loi.NextMarker = object.GetInfo().GetName()
		}
		name := object.GetInfo().GetName()
		//首先进行前缀判定
		if !strings.HasPrefix(name, prefix) {
			continue
		}

		//非递归，只返回一级目录
		if !recursive {
			//已经新建有用Prefix命名的了
			if name == entryPrefixMatch {
				hasPrefixKey = true
			}
			//看看这个能不能抽象出文件夹
			index := strings.Index(name[prefixLen:], delimiter)
			//无"/"，简单对象
			if index < 0 {
				loi.Objects = append(loi.Objects, minio.ObjectInfo{
					Bucket:      bucket,
					Name:        name,
					ModTime:     time.Unix(object.GetMTime(), 0),
					Size:        object.GetLength(),
					IsDir:       object.GetInfo().GetDir(),
					ETag:        object.GetETag(),
					ContentType: object.GetInfo().GetContentType(),
				})
			} else {
				//有"/"，获取文件夹抽象
				if first {
					first = false
					recursiveKey = name[:prefixLen+index+1]
					loi.Prefixes = append(loi.Prefixes, recursiveKey)
				} else {
					//这个前缀已经记录，不管
					if strings.HasPrefix(name, recursiveKey) {
						continue
					}

					//一个新前缀
					recursiveKey = name[:prefixLen+index+1]
					loi.Prefixes = append(loi.Prefixes, recursiveKey)
				}
			}
		} else { //递归获取，全部返回
			loi.Objects = append(loi.Objects, minio.ObjectInfo{
				Bucket:      bucket,
				Name:        object.GetInfo().GetName(),
				ModTime:     time.Unix(object.GetMTime(), 0),
				Size:        object.GetLength(),
				IsDir:       object.GetInfo().GetDir(),
				ETag:        object.GetETag(),
				ContentType: object.GetInfo().GetContentType(),
			})
		}
		limit--
	}

	//!recursive时返回的结果依然包括Prefix，抽象成文件夹
	if len(prefix) > 0 && !recursive && !hasPrefixKey {
		loi.Objects = append(loi.Objects, minio.ObjectInfo{
			Bucket:  bucket,
			Name:    entryPrefixMatch,
			Size:    0,
			IsDir:   true,
			ModTime: time.Now().UTC(),
		})
	}
	//没读完
	if objectIter != nil {
		loi.IsTruncated = true
	}
	return loi, nil
}

// ListObjectsV2 lists all blobs in LFS bucket filtered by prefix
func (l *lfsGateway) ListObjectsV2(ctx context.Context, bucket, prefix, continuationToken, delimiter string, maxKeys int,
	fetchOwner bool, startAfter string) (loiv2 minio.ListObjectsV2Info, err error) {
	marker := continuationToken
	if marker == "" {
		marker = startAfter
	}

	loi, err := l.ListObjects(ctx, bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		return loiv2, err
	}

	loiv2 = minio.ListObjectsV2Info{
		IsTruncated:           loi.IsTruncated,
		ContinuationToken:     continuationToken,
		NextContinuationToken: loi.NextMarker,
		Objects:               loi.Objects,
		Prefixes:              loi.Prefixes,
	}
	return loiv2, err
}

// GetObjectNInfo - returns object info and locked object ReadCloser
func (l *lfsGateway) GetObjectNInfo(ctx context.Context, bucket, object string, rs *minio.HTTPRangeSpec, h http.Header, lockType minio.LockType, opts minio.ObjectOptions) (gr *minio.GetObjectReader, err error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return gr, convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	objInfo, err := l.GetObjectInfo(ctx, bucket, object, opts)
	if err != nil {
		return gr, convertToMinioError(err, bucket, object)
	}

	piper, pipew := io.Pipe()
	bufw := bufio.NewWriterSize(pipew, user.DefaultBufSize)
	checkErrAndClosePipe := func(err error) error {
		if err != nil {
			err = pipew.CloseWithError(err)
			return err
		}
		err = pipew.Close()
		return err
	}
	var complete []user.CompleteFunc
	complete = append(complete, checkErrAndClosePipe)
	start, length, err := rs.GetOffsetLength(objInfo.Size)
	if err != nil {
		return gr, err
	}
	go l.lfs.GetObject(ctx, bucket, object, bufw, complete, user.DownloadObjectOptions{Start: start, Length: length})

	// Setup cleanup function to cause the above go-routine to
	// exit in case of partial read
	pipeCloser := func() { piper.Close() }
	return minio.NewGetObjectReaderFromReader(piper, objInfo, opts.CheckCopyPrecondFn, pipeCloser)
}

// GetObject reads an object on LFS. Supports additional
// parameters like offset and length which are synonymous with
// HTTP Range requests.
//
// startOffset indicates the starting read location of the object.
// length indicates the total length of the object.
func (l *lfsGateway) GetObject(ctx context.Context, bucket, key string, startOffset, length int64, writer io.Writer, etag string, opts minio.ObjectOptions) error {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	var errRes error
	bufw := bufio.NewWriterSize(writer, user.DefaultBufSize)
	checkErrAndClosePipe := func(err error) error {
		errRes = err
		return nil
	}
	var complete []user.CompleteFunc
	complete = append(complete, checkErrAndClosePipe)
	err := l.lfs.GetObject(ctx, bucket, key, bufw, complete, user.DownloadObjectOptions{Start: startOffset, Length: length})

	if err != nil {
		return convertToMinioError(err, bucket, "")
	}

	return errRes
}

// GetObjectInfo reads object info and replies back ObjectInfo.
func (l *lfsGateway) GetObjectInfo(ctx context.Context, bucket, object string, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return minio.ObjectInfo{}, convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	obj, err := l.lfs.HeadObject(ctx, bucket, object)
	if err != nil {
		return minio.ObjectInfo{}, convertToMinioError(err, bucket, object)
	}
	// need handle ETag
	objInfo = minio.ObjectInfo{
		Bucket:      bucket,
		Name:        object,
		IsDir:       obj.GetInfo().GetDir(),
		ETag:        obj.GetETag(),
		ContentType: obj.GetInfo().GetContentType(),
		Size:        obj.GetLength(),
	}

	return objInfo, nil
}

// PutObject creates a new object with the incoming data.
func (l *lfsGateway) PutObject(ctx context.Context, bucket, object string, r *minio.PutObjReader, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return minio.ObjectInfo{}, convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	ops := user.DefaultUploadOption()
	ops.UserDefined = opts.UserDefined
	reader := bufio.NewReaderSize(r.Reader, user.DefaultBufSize)
	obj, err := l.lfs.PutObject(ctx, bucket, object, reader, ops)
	if err != nil {
		return objInfo, convertToMinioError(err, bucket, object)
	}

	objInfo = minio.ObjectInfo{
		Bucket:      bucket,
		Name:        object,
		IsDir:       obj.GetInfo().GetDir(),
		ETag:        obj.GetETag(),
		ContentType: obj.GetInfo().GetContentType(),
		Size:        obj.GetLength(),
	}

	return objInfo, err
}

// CopyObject copies an object from source bucket to a destination bucket.
func (l *lfsGateway) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string, srcInfo minio.ObjectInfo, srcOpts, dstOpts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	return objInfo, nil
}

// DeleteObject deletes a blob in bucket.
func (l *lfsGateway) DeleteObject(ctx context.Context, bucket, object string) error {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	_, err := l.lfs.DeleteObject(ctx, bucket, object)

	return convertToMinioError(err, bucket, object)
}

func (l *lfsGateway) DeleteObjects(ctx context.Context, bucket string, objects []string) ([]error, error) {
	if l.lfs == nil || !l.lfs.Online() {
		//再检查一次
		err := l.checkLfs(ctx)
		if err != nil {
			return nil, convertToMinioError(errLfsServiceNotReady, "", "")
		}
	}

	errFlag := 0
	errs := make([]error, len(objects))
	for i, object := range objects {
		_, err := l.lfs.DeleteObject(ctx, bucket, object)
		if err != nil {
			errFlag = 1
			errs[i] = convertToMinioError(err, bucket, object)
		}
	}

	if errFlag != 0 {
		return errs, errDeleteObjects
	}

	return errs, nil
}

// SetBucketPolicy sets policy on bucket.
// LFS supports three types of bucket policies:
// oss.ACLPublicReadWrite: readwrite in minio terminology
// oss.ACLPublicRead: readonly in minio terminology
// oss.ACLPrivate: none in minio terminology
func (l *lfsGateway) SetBucketPolicy(ctx context.Context, bucket string, bucketPolicy *policy.Policy) error {
	return nil
}

// GetBucketPolicy will get policy on bucket.
func (l *lfsGateway) GetBucketPolicy(ctx context.Context, bucket string) (*policy.Policy, error) {
	return nil, convertToMinioError(errLfsServiceNotReady, "", "")
}

// DeleteBucketPolicy deletes all policies on bucket.
func (l *lfsGateway) DeleteBucketPolicy(ctx context.Context, bucket string) error {
	return nil
}

// IsCompressionSupported returns whether compression is applicable for this layer.
func (l *lfsGateway) IsCompressionSupported() bool {
	return false
}

func convertToMinioError(err error, bucket, object string) error {
	switch err {
	case errLfsServiceNotReady:
		return minio.BackendDown{}
	case user.ErrLfsReadOnly:
		return minio.PrefixAccessDenied{Bucket: bucket, Object: object}
	case user.ErrBucketNameInvalid:
		return minio.BucketNameInvalid{Bucket: bucket}
	case user.ErrBucketNotExist:
		return minio.BucketNotFound{Bucket: bucket}
	case user.ErrBucketAlreadyExist:
		return minio.BucketExists{Bucket: bucket}
	case user.ErrObjectNameInvalid:
		return minio.ObjectNameInvalid{Bucket: bucket, Object: object}
	case user.ErrObjectAlreadyExist:
		return minio.ObjectAlreadyExists{Bucket: bucket, Object: object}
	case user.ErrObjectNotExist:
		return minio.ObjectNotFound{Bucket: bucket, Object: object}
	case user.ErrObjectIsDir:
		return minio.ObjectExistsAsDirectory{Bucket: bucket, Object: object}
	case nil:
		return nil
	default:
		return minio.PrefixAccessDenied{Bucket: bucket, Object: object}
	}
}
