package user

import (
	"context"
	"io"

	pb "github.com/memoio/go-mefs/proto"
)

// FileSyetem defines user's function
type FileSyetem interface {
	Start(ctx context.Context) error
	Stop() error
	Fsync(bool) error
	Online() bool

	ListBuckets(ctx context.Context, prefix string) ([]*pb.BucketInfo, error)
	CreateBucket(ctx context.Context, bucketName string, options *pb.BucketOptions) (*pb.BucketInfo, error)
	HeadBucket(ctx context.Context, bucketName string) (*pb.BucketInfo, error)
	DeleteBucket(ctx context.Context, bucketName string) (*pb.BucketInfo, error)

	ListObjects(ctx context.Context, bucketName, prefix string, opts ObjectOptions) ([]*pb.ObjectInfo, error)

	PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader) (*pb.ObjectInfo, error)
	GetObject(ctx context.Context, bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error
	HeadObject(ctx context.Context, bucketName, objectName string, opts ObjectOptions) (*pb.ObjectInfo, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) (*pb.ObjectInfo, error)

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
