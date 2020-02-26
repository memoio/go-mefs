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
	"time"

	"github.com/memoio/go-mefs/core"
	df "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role/user"
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

	lfs := core.LocalNode.Inst.(*user.Info).GetUser(uid)
	if lfs == nil {
		log.Println("warn: please start lfs first to use gateway service")
	}

	return &lfsGateway{
		userID:    uid,
		multipart: uploads,
	}, nil
}

// Production - oss is production ready.
func (g *Mefs) Production() bool {
	return false
}

// lfsGateway implements gateway.
type lfsGateway struct {
	minio.GatewayUnsupported
	userID    string
	multipart *MultipartUploads
}

// Shutdown saves any gateway metadata to disk
// if necessary and reload upon next restart.
func (l *lfsGateway) Shutdown(ctx context.Context) error {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return errLfsServiceNotReady
	}
	return lfs.Stop()
}

// StorageInfo is not relevant to LFS backend.
func (l *lfsGateway) StorageInfo(ctx context.Context) (si minio.StorageInfo) {
	si.Backend.Type = minio.BackendGateway
	si.Backend.GatewayOnline = true

	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return si
	}

	use, _ := lfs.ShowStorage(ctx)
	si.Used = []uint64{use}
	return si
}

// MakeBucketWithLocation creates a new container on LFS backend.
func (l *lfsGateway) MakeBucketWithLocation(ctx context.Context, bucket, options string) error {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return errLfsServiceNotReady
	}

	bucketOptions := &mpb.BucketOptions{}
	err := json.Unmarshal([]byte(options), bucketOptions)
	if err != nil {
		bucketOptions = df.DefaultBucketOptions()
	}
	_, err = lfs.CreateBucket(ctx, bucket, bucketOptions)

	return convertToMinioError(err, bucket, "")
}

// GetBucketInfo gets bucket metadata.
func (l *lfsGateway) GetBucketInfo(ctx context.Context, bucket string) (bi minio.BucketInfo, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return bi, errLfsServiceNotReady
	}

	bucketInfo, err := lfs.HeadBucket(ctx, bucket)
	if err != nil {
		return bi, convertToMinioError(err, bucket, "")
	}
	bi.Name = bucket
	bi.Created = time.Unix(bucketInfo.Ctime, 0).UTC()
	return bi, nil
}

// ListBuckets lists all LFS buckets.
func (l *lfsGateway) ListBuckets(ctx context.Context) (buckets []minio.BucketInfo, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return nil, user.ErrLfsServiceNotReady
	}

	bucketsInfo, err := lfs.ListBuckets(ctx, "")
	if err != nil {
		return nil, convertToMinioError(err, "", "")
	}

	buckets = make([]minio.BucketInfo, len(bucketsInfo))
	for i, v := range bucketsInfo {
		buckets[i].Name = v.Name
		buckets[i].Created = time.Unix(v.Ctime, 0).UTC()
	}
	return buckets, nil
}

// DeleteBucket deletes a bucket on LFS.
func (l *lfsGateway) DeleteBucket(ctx context.Context, bucket string) error {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return errLfsServiceNotReady
	}

	_, err := lfs.DeleteBucket(ctx, bucket)
	return convertToMinioError(err, bucket, "")
}

// ListObjects lists all blobs in LFS bucket filtered by prefix.
func (l *lfsGateway) ListObjects(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (loi minio.ListObjectsInfo, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return loi, errLfsServiceNotReady
	}

	var ops user.ObjectOptions

	objs, err := lfs.ListObjects(ctx, bucket, prefix, ops)
	if err != nil {
		return loi, convertToMinioError(err, bucket, "")
	}

	loi.Objects = make([]minio.ObjectInfo, len(objs))
	for i, v := range objs {
		loi.Objects[i].Bucket = bucket
		loi.Objects[i].ETag = v.OPart.ETag
		loi.Objects[i].Name = v.OPart.Name
		loi.Objects[i].Size = v.OPart.Length
		loi.Objects[i].ModTime = time.Unix(v.Ctime, 0).UTC()
	}
	return loi, nil
}

// ListObjectsV2 lists all blobs in LFS bucket filtered by prefix
func (l *lfsGateway) ListObjectsV2(ctx context.Context, bucket, prefix, continuationToken, delimiter string, maxKeys int,
	fetchOwner bool, startAfter string) (loi minio.ListObjectsV2Info, err error) {

	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return loi, errLfsServiceNotReady
	}

	var ops user.ObjectOptions

	objs, err := lfs.ListObjects(ctx, bucket, prefix, ops)
	if err != nil {
		return loi, convertToMinioError(err, bucket, "")
	}
	loi.Objects = make([]minio.ObjectInfo, len(objs))
	for i, v := range objs {
		loi.Objects[i].Bucket = bucket
		loi.Objects[i].ETag = v.OPart.ETag
		loi.Objects[i].Name = v.OPart.Name
		loi.Objects[i].Size = v.OPart.Length
		loi.Objects[i].ModTime = time.Unix(v.Ctime, 0).UTC()
	}
	return loi, nil
}

// GetObjectNInfo - returns object info and locked object ReadCloser
func (l *lfsGateway) GetObjectNInfo(ctx context.Context, bucket, object string, rs *minio.HTTPRangeSpec, h http.Header, lockType minio.LockType, opts minio.ObjectOptions) (gr *minio.GetObjectReader, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return gr, errLfsServiceNotReady
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
	go lfs.GetObject(ctx, bucket, object, bufw, complete, user.DefaultDownloadOptions())

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
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return errLfsServiceNotReady
	}

	var errRes error
	bufw := bufio.NewWriterSize(writer, user.DefaultBufSize)
	checkErrAndClosePipe := func(err error) error {
		errRes = err
		return nil
	}
	var complete []user.CompleteFunc
	complete = append(complete, checkErrAndClosePipe)
	err := lfs.GetObject(ctx, bucket, key, bufw, complete, &user.DownloadOptions{
		Start:  startOffset,
		Length: length,
	})

	if err != nil {
		return convertToMinioError(err, bucket, "")
	}

	return errRes
}

// GetObjectInfo reads object info and replies back ObjectInfo.
func (l *lfsGateway) GetObjectInfo(ctx context.Context, bucket, object string, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return minio.ObjectInfo{}, errLfsServiceNotReady
	}

	var ops user.ObjectOptions
	obj, err := lfs.HeadObject(ctx, bucket, object, ops)
	if err != nil {
		return minio.ObjectInfo{}, convertToMinioError(err, bucket, object)
	}
	objInfo = minio.ObjectInfo{
		Bucket:      bucket,
		Name:        object,
		IsDir:       obj.Dir,
		ETag:        obj.OPart.ETag,
		ContentType: obj.ContentType,
		Size:        obj.OPart.Length,
	}

	return objInfo, nil
}

// PutObject creates a new object with the incoming data.
func (l *lfsGateway) PutObject(ctx context.Context, bucket, object string, r *minio.PutObjReader, opts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return minio.ObjectInfo{}, errLfsServiceNotReady
	}

	reader := bufio.NewReaderSize(r.Reader, user.DefaultBufSize)
	obj, err := lfs.PutObject(ctx, bucket, object, reader)
	if err != nil {
		return objInfo, convertToMinioError(err, bucket, object)
	}

	objInfo = minio.ObjectInfo{
		Bucket:      bucket,
		Name:        object,
		IsDir:       obj.Dir,
		ETag:        obj.OPart.ETag,
		ContentType: obj.ContentType,
		Size:        obj.OPart.Length,
	}

	return objInfo, err
}

// CopyObject copies an object from source bucket to a destination bucket.
func (l *lfsGateway) CopyObject(ctx context.Context, srcBucket, srcObject, dstBucket, dstObject string, srcInfo minio.ObjectInfo, srcOpts, dstOpts minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	return objInfo, nil
}

// DeleteObject deletes a blob in bucket.
func (l *lfsGateway) DeleteObject(ctx context.Context, bucket, object string) error {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return errLfsServiceNotReady
	}

	_, err := lfs.DeleteObject(ctx, bucket, object)

	return convertToMinioError(err, bucket, object)
}

func (l *lfsGateway) DeleteObjects(ctx context.Context, bucket string, objects []string) ([]error, error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return nil, errLfsServiceNotReady
	}

	errFlag := 0
	errs := make([]error, len(objects))
	for i, object := range objects {
		_, err := lfs.DeleteObject(ctx, bucket, object)
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
	return nil, errLfsServiceNotReady
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
	default:
		return err
	}
}
