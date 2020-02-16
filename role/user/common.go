package user

import (
	"errors"

	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/minio/minio-go/v6/pkg/s3utils"
)

//-------Group Type------

const (
	// KeeperSLA is keeper needed
	KeeperSLA = 3
	// ProviderSLA is provider needed
	ProviderSLA = 6
	// DefaultCapacity is default store capacity
	DefaultCapacity int64 = 1000 //单位：MB
	// DefaultDuration is default store days
	DefaultDuration int64 = 100 //单位：天

	defaultMetaBackupCount int32 = 3
	flushLocalBackup             = 1

	// DefaultBufSize used for read
	DefaultBufSize = 1024 * 1024 * 4
)

var (
	ErrPolicy               = errors.New("policy is error")
	ErrLfsServiceNotReady   = errors.New("lfs service is not ready")
	ErrLfsReadOnly          = errors.New("lfs service is read only")
	ErrCannotGetEnoughBlock = errors.New("cannot get enough block")
	ErrCannotLoadMetaBlock  = errors.New("cannot load metaBlock")
	ErrCannotLoadSuperBlock = errors.New("cannot load superblock")

	ErrNoProviders      = errors.New("there is no providers")
	ErrNoKeepers        = errors.New("there is no keepers")
	ErrNoEnoughProvider = errors.New("no enough providers")
	ErrNoEnoughKeeper   = errors.New("no enough keepers")

	ErrBucketNotExist     = errors.New("bucket not exist")
	ErrBucketAlreadyExist = errors.New("bucket already exists")
	ErrBucketNotEmpty     = errors.New("bucket is not empty")
	ErrBucketNameInvalid  = errors.New("bucket name is invalid")

	ErrObjectNotExist       = errors.New("object not exist")
	ErrObjectAlreadyExist   = errors.New("object already exist")
	ErrObjectNameToolong    = errors.New("object name is too long")
	ErrObjectNameInvalid    = errors.New("object name is invalid")
	ErrObjectOptionsInvalid = errors.New("object option is invalid")
)

// DefaultBucketOptions is default bucket option
func DefaultBucketOptions() *mpb.BucketOptions {
	return &mpb.BucketOptions{
		Version:      1,
		Policy:       dataformat.RsPolicy,
		DataCount:    3,
		ParityCount:  2,
		SegmentSize:  dataformat.DefaultSegmentSize,
		TagFlag:      dataformat.BLS12,
		SegmentCount: dataformat.DefaultSegmentCount,
		Encryption:   1,
	}
}

//检查文件名合法性
func checkBucketName(bucketName string) error {
	return s3utils.CheckValidBucketName(bucketName)
}

func checkObjectName(objectName string) error {
	err := s3utils.CheckValidObjectName(objectName)
	if err != nil {
		return err
	}

	for i := 0; i < len(objectName); i++ {
		if objectName[i] == '\\' || objectName[i] == '\n' {
			return ErrObjectNameInvalid
		}
	}
	return nil
}
