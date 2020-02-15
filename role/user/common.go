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
	ErrPolicy                = errors.New("policy is error")
	ErrBalance               = errors.New("balance is insufficient")
	ErrKeySetIsNil           = errors.New("bls keyset is nil")
	ErrUserNotExist          = errors.New("user does not exist")
	ErrLfsServiceNotReady    = errors.New("lfs service is not ready")
	ErrCannotStartLfsService = errors.New("cannot start lfs service")
	ErrReadOnly              = errors.New("lfs is read only")

	errGetContractItem = errors.New("cannot get contract Item")
	errTimeOut         = errors.New("Time out")

	ErrNoProviders           = errors.New("there is no providers")
	ErrNoKeepers             = errors.New("there is no keepers")
	ErrNoEnoughProvider      = errors.New("no enough providers")
	ErrNoEnoughKeeper        = errors.New("no enough keepers")
	ErrCannotConnectNetwork  = errors.New("cannot connect")
	ErrCannotDeleteMetaBlock = errors.New("cannot delete metablock in provider")

	ErrBucketNotExist     = errors.New("bucket not exist")
	ErrBucketAlreadyExist = errors.New("bucket already exists")
	ErrBucketNotEmpty     = errors.New("bucket is not empty")
	ErrBucketNameInvalid  = errors.New("bucket name is invalid")

	ErrObjectNotExist       = errors.New("object not exist")
	ErrObjectAlreadyExist   = errors.New("object already exist")
	ErrObjectNameToolong    = errors.New("object name is too long")
	ErrObjectNameInvalid    = errors.New("object name is invalid")
	ErrObjectOptionsInvalid = errors.New("object option is invalid")

	ErrCannotGetEnoughBlock = errors.New("cannot get enough Block")
	ErrCannotLoadMetaBlock  = errors.New("cannot load MetaBlock")
	ErrCannotAddBlock       = errors.New("cannot put this block")
	ErrCannotLoadSuperBlock = errors.New("cannot load superblock")
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
