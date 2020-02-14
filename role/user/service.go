package user

import (
	"context"
	"io"

	mpb "github.com/memoio/go-mefs/proto"
)

// FileSyetem defines user's function
type FileSyetem interface {
	Start(ctx context.Context) error
	Stop() error
	Fsync(bool) error
	Online() bool

	ListBuckets(ctx context.Context, prefix string) ([]*mpb.BucketInfo, error)
	CreateBucket(ctx context.Context, bucketName string, options *mpb.BucketOptions) (*mpb.BucketInfo, error)
	HeadBucket(ctx context.Context, bucketName string) (*mpb.BucketInfo, error)
	DeleteBucket(ctx context.Context, bucketName string) (*mpb.BucketInfo, error)

	ListObjects(ctx context.Context, bucketName, prefix string, opts ObjectOptions) ([]*mpb.ObjectInfo, error)

	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader) (*mpb.ObjectInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error
	HeadObject(ctx context.Context, bucketName, objectName string, opts ObjectOptions) (*mpb.ObjectInfo, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) (*mpb.ObjectInfo, error)

	ShowStorage(ctx context.Context) (uint64, error)
	ShowBucketStorage(ctx context.Context, bucketName string) (uint64, error)
}

// service is user's service
type service interface {
	Stop() error
	Fsync() error
	GetFS(userID string) FileSyetem
	NewFS(userID string) error
}

// group is used to init network parameters for user
type group interface {
	start(ctx context.Context) error
	connect(ctx context.Context) error
	// broadcast init information
	initGroup(ctx context.Context) error
	// notify keepers and providers
	notify()
	// confirm all keepers
	confirm(ctx context.Context)
	deployContract(ctx context.Context)
}
