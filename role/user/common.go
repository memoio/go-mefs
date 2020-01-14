package user

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
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

// We support '.' with bucket names but we fallback to using path
// style requests instead for such superBucket.
var (
	validBucketName       = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9\.\-\_\:]{1,61}[A-Za-z0-9]$`)
	validBucketNameStrict = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	ipAddress             = regexp.MustCompile(`^(\d+\.){3}\d+$`)
)

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
		Policy:      dataformat.RsPolicy,
		DataCount:   3,
		ParityCount: 2,
		SegmentSize: dataformat.DefaultSegmentSize,
		TagFlag:     dataformat.BLS12,
		Encryption:  true,
	}
}

//检查文件名合法性，文件名中不能含有"/"
func checkObjectName(objectName string) error {
	if strings.TrimSpace(objectName) == "" {
		return errors.New("objectInfo name cannot be empty")
	}
	if len(objectName) > maxObjectNameLen {
		return ErrObjectNameToolong
	}
	if !utf8.ValidString(objectName) {
		return errors.New("objectInfo name with non UTF-8 strings are not supported")
	}
	for i := 0; i < len(objectName); i++ {
		if objectName[i] == '/' || objectName[i] == '\\' || objectName[i] == '\n' {
			return ErrObjectNameInvalid
		}
	}
	return nil
}

// CheckValidBucketName - checks if we have a valid input bucket name.
func CheckValidBucketName(bucketName string) (err error) {
	return checkBucketNameCommon(bucketName, false)
}

// CheckValidBucketNameStrict - checks if we have a valid input bucket name.
// This is a stricter version.
// - http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingBucket.html
func CheckValidBucketNameStrict(bucketName string) (err error) {
	return checkBucketNameCommon(bucketName, true)
}

// Common checker for both stricter and basic validation.
func checkBucketNameCommon(bucketName string, strict bool) (err error) {
	if strings.TrimSpace(bucketName) == "" {
		return errors.New("superBucket name cannot be empty")
	}
	if len(bucketName) < 3 {
		return errors.New("superBucket name cannot be smaller than 3 characters")
	}
	if len(bucketName) > 63 {
		return errors.New("superBucket name cannot be greater than 63 characters")
	}
	if ipAddress.MatchString(bucketName) {
		return errors.New("superBucket name cannot be an ip address")
	}
	if strings.Contains(bucketName, "..") || strings.Contains(bucketName, ".-") || strings.Contains(bucketName, "-.") {
		return errors.New("superBucket name contains invalid characters")
	}
	if strict {
		if !validBucketNameStrict.MatchString(bucketName) {
			err = errors.New("superBucket name contains invalid characters")
		}
		return err
	}
	if !validBucketName.MatchString(bucketName) {
		err = errors.New("superBucket name contains invalid characters")
	}
	return err
}

// BuildSignMessage builds sign message for test or repair
func BuildSignMessage() ([]byte, error) {
	money := big.NewInt(123)
	moneyByte := money.Bytes()
	message := &pb.ChannelSign{
		Value:     moneyByte,
		ChannelID: "test",
	}
	mes, err := proto.Marshal(message)
	if err != nil {
		utils.MLogger.Error("protoMarshal failed: ", err)
		return nil, err
	}
	return mes, nil
}
