package user

import (
	"context"
	"errors"
	"log"
	"math/big"
	"regexp"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang/protobuf/proto"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
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
	ErrPolicy                    = errors.New("the policy is error")
	ErrBalance                   = errors.New("your account's balance is insufficient, we will not deploy contract")
	ErrKeySetIsNil               = errors.New("user's Keyset is nil")
	ErrUserNotExist              = errors.New("user not exist")
	ErrUserBookIsNil             = errors.New("the User book is nil")
	ErrCannotFindUserInUserBook  = errors.New("cannot find this user in userbook")
	errGetContractItem           = errors.New("Can't get contract Item")
	ErrContractServiceAlreadySet = errors.New("this contract Service already set")
	ErrGroupServiceAlreadySet    = errors.New("this group Service already set")
	ErrLfsServiceAlreadySet      = errors.New("this lfs Service already set")
	ErrTimeOut                   = errors.New("Time out")

	ErrNoProviders           = errors.New("there is no providers")
	ErrNoKeepers             = errors.New("there is no keepers")
	ErrCannotConnectKeeper   = errors.New("cannot connect Keeper")
	ErrCannotConnectProvider = errors.New("cannot connect this provider")
	ErrNoEnoughProvider      = errors.New("no Enough providers")
	ErrNoEnoughKeeper        = errors.New("no Enough keepers")
	ErrCannotConnectNetwork  = errors.New("cannot connect NetWork")
	ErrCannotDeleteMetaBlock = errors.New("cannot delete metablock in provider,maybe it is not connected")
	ErrGroupServiceNotReady  = errors.New("group service is not ready")

	ErrCannotStartLfsService = errors.New("cannot start lfs service")
	ErrLfsIsNotRunning       = errors.New("lfs is not running")

	ErrObjectNotExist     = errors.New("object is not exist")
	ErrDirNotExist        = errors.New("directory is not exist")
	ErrObjectAlreadyExist = errors.New("file already exist")

	ErrBucketNotExist     = errors.New("bucket is not exist")
	ErrBucketAlreadyExist = errors.New("bucket Already Exist")
	ErrBucketNotEmpty     = errors.New("bucket is Not empty")
	ErrBucketNameInvalid  = errors.New("bucket name invalid")

	ErrObjectNameToolong    = errors.New("the object's name is too long")
	ErrObjectNameInvalid    = errors.New("object name invalid")
	ErrObjectOptionsInvalid = errors.New("object options invalid")

	ErrCannotGetEnoughBlock = errors.New("cannot get enough Block")
	ErrCannotLoadMetaBlock  = errors.New("cannot Load MetaBlock")
	ErrCannotAddBlock       = errors.New("cannot Add this block")
	ErrCannotLoadSuperBlock = errors.New("cannot load superblock")
	ErrWrongState           = errors.New("wrong userservice state")
	ErrWrongInitState       = errors.New("wrong init state")
)

func putKeyTo(key, value, node string) error {
	return localNode.Routing.(*dht.IpfsDHT).CmdPutTo(key, value, node)
}

func getKeyFrom(key, node string) ([]byte, error) {
	return localNode.Routing.(*dht.IpfsDHT).CmdGetFrom(key, node)
}

func sendMetaMessage(km *metainfo.KeyMeta, metaValue, to string) error {
	caller := ""
	for _, i := range []int{0, 1, 2, 3, 4} {
		pc, _, _, _ := runtime.Caller(i)
		caller += string(i) + ":" + runtime.FuncForPC(pc).Name() + "\n"
	}
	return localNode.Routing.(*dht.IpfsDHT).SendMetaMessage(km.ToString(), metaValue, to, caller)
}

func sendMetaRequest(km *metainfo.KeyMeta, metaValue, to string) (string, error) {
	caller := ""
	for _, i := range []int{0, 1, 2, 3, 4} {
		pc, _, _, _ := runtime.Caller(i)
		caller += string(i) + ":" + runtime.FuncForPC(pc).Name() + "\n"
	}
	return localNode.Routing.(*dht.IpfsDHT).SendMetaRequest(km.ToString(), metaValue, to, caller)
}

// broadcastMetaMessage 广播发送信息，现在只针对初始化流程写
func broadcastMetaMessage(km *metainfo.KeyMeta, metavalue string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = context.WithValue(ctx, "User_Init_Req", true)
	/*pc, _, _, _ := runtime.Caller(2)
	caller := runtime.FuncForPC(pc).Name()
	ctx = context.WithValue(ctx, "caller", caller)*/
	_, err := localNode.Routing.(*dht.IpfsDHT).GetValue(ctx, km.ToString())
	return err
}

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

func testConnect() error {
	waitTime := 0 //进行网络连接
	for {
		if waitTime > 60 { //连不上网？
			log.Println(ErrCannotConnectNetwork, "please restart and retry.")
			return ErrCannotConnectNetwork
		}
		if connPeers := localNode.PeerHost.Network().Peers(); len(connPeers) != 0 { //刚启动还没连接节点，等等
			break //连上网了，退出
		} else {
			log.Println(ErrCannotConnectNetwork, "waiting...")
			time.Sleep(10 * time.Second) //没联网，等联网
		}
		waitTime++
	}
	return nil
}

// BuildSignMessage builds sign message for test or repair
func BuildSignMessage() ([]byte, error) {
	money := big.NewInt(123)
	moneyByte := money.Bytes()
	message := &pb.SignForChannel{
		Money: moneyByte,
	}
	mes, err := proto.Marshal(message)
	if err != nil {
		log.Println("protoMarshal failed err: ", err)
		return nil, err
	}
	return mes, nil
}
