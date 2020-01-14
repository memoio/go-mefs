package user

import (
	"context"
	"io"

	pb "github.com/memoio/go-mefs/proto"
)

// FileSyetem defines user's function
type FileSyetem interface {
	Start() error
	Stop() error
	Fsync(bool) error
	Online() bool

	ListBuckets(prefix string) ([]*pb.BucketInfo, error)
	CreateBucket(bucketName string, options *pb.BucketOptions) (*pb.BucketInfo, error)
	HeadBucket(bucketName string) (*pb.BucketInfo, error)
	DeleteBucket(bucketName string) (*pb.BucketInfo, error)

	ListObjects(bucketName, prefix string, opts ObjectOptions) ([]*pb.ObjectInfo, error)

	PutObject(bucketName, objectName string, reader io.Reader) (*pb.ObjectInfo, error)
	GetObject(bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error
	HeadObject(bucketName, objectName string, opts ObjectOptions) (*pb.ObjectInfo, error)
	DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error)

	ShowStorage() (uint64, error)
	ShowBucketStorage(bucketName string) (uint64, error)
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
