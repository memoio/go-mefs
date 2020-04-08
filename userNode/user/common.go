package user

import (
	"errors"

	"github.com/minio/minio-go/v6/pkg/s3utils"
)

const (
	defaultMetaBackupCount int32 = 3
	flushLocalBackup             = 1

	defaultTransNum = 32 * 8
	// DefaultBufSize used for read
	DefaultBufSize = 1024 * 1024 * 4

	MaxListKeys = 1000
)

var transNum = defaultTransNum

var (
	ErrPolicy               = errors.New("policy is error")
	ErrLfsServiceNotReady   = errors.New("lfs service is not ready")
	ErrLfsReadOnly          = errors.New("lfs service is read only")
	ErrLfsStarting          = errors.New("Another lfs instance is starting")
	ErrCannotGetEnoughBlock = errors.New("cannot get enough block")
	ErrCannotLoadMetaBlock  = errors.New("cannot load metaBlock")
	ErrCannotLoadSuperBlock = errors.New("cannot load superblock")
	ErrUpload               = errors.New("upload fails")
	ErrResourceUnavailable  = errors.New("resource unavailable")

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
	ErrObjectIsDir          = errors.New("object is directory")
	ErrNoEnoughBlockUpload  = errors.New("block uploaded is not enough")
)

//检查文件名合法性
func checkBucketName(bucketName string) error {
	return s3utils.CheckValidBucketName(bucketName)
}

func checkObjectName(objectName string) error {
	err := s3utils.CheckValidObjectName(objectName)
	if err != nil {
		return err
	}

	return nil
}
