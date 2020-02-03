package user

import (
	"encoding/binary"
	"errors"
	"golang.org/x/crypto/blake2b"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/minio/minio-go/v6/pkg/s3utils"
)

//-------Group Type------

const (
	KeeperSLA             = 2 //暂定
	ProviderSLA           = 6
	DefaultCapacity int64 = 1000 //单位：MB
	DefaultDuration int64 = 100  //单位：天

	//LFS
	maxObjectNameLen = 4096 //设定文件名和路径可占用的最长字节数

	DefaultGetBlockDelay = 30 * time.Second

	defaultMetaBackupCount int32 = 3
	flushLocalBackup             = 1
)

const DefaultBufSize = 1024 * 1024 * 4

var (
	ErrPolicy                = errors.New("policy is error")
	ErrBalance               = errors.New("balance is insufficient")
	ErrKeySetIsNil           = errors.New("bls keyset is nil")
	ErrUserNotExist          = errors.New("user does not exist")
	ErrLfsServiceNotReady    = errors.New("lfs service is not ready")
	ErrCannotStartLfsService = errors.New("cannot start lfs service")

	errGetContractItem = errors.New("cannot get contract Item")
	ErrTimeOut         = errors.New("Time out")

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
	ErrDirNotExist          = errors.New("directory not exist")
	ErrObjectAlreadyExist   = errors.New("object already exist")
	ErrObjectNameToolong    = errors.New("object name is too long")
	ErrObjectNameInvalid    = errors.New("object name is invalid")
	ErrObjectOptionsInvalid = errors.New("object option is invalid")

	ErrCannotGetEnoughBlock = errors.New("cannot get enough Block")
	ErrCannotLoadMetaBlock  = errors.New("cannot load MetaBlock")
	ErrCannotAddBlock       = errors.New("cannot put this block")
	ErrCannotLoadSuperBlock = errors.New("cannot load superblock")
)

func DefaultBucketOptions() *pb.BucketOptions {
	return &pb.BucketOptions{
		Version:      1,
		Policy:       dataformat.RsPolicy,
		DataCount:    3,
		ParityCount:  2,
		SegmentSize:  dataformat.DefaultSegmentSize,
		TagFlag:      dataformat.BLS12,
		SegmentCount: dataformat.DefaultSegmentCount,
		Encryption:   0,
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

// CreateAesKey creates
func CreateAesKey(privateKey, queryID []byte, bucketID int32, objectStart int64) [32]byte {
	tmpkey := make([]byte, len(privateKey)+len(queryID)+12)
	copy(tmpkey, privateKey)
	copy(tmpkey[len(privateKey):], queryID)
	binary.LittleEndian.PutUint32(tmpkey[len(privateKey)+len(queryID):], uint32(bucketID))
	binary.LittleEndian.PutUint64(tmpkey[len(privateKey)+len(queryID)+4:], uint64(objectStart))
	return blake2b.Sum256(tmpkey)
}
