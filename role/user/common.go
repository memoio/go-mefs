package user

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	"github.com/memoio/go-mefs/contracts"
	pb "github.com/memoio/go-mefs/role/user/pb"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/metainfo"
)

//-------Group Type------

const (
	KeeperSLA             = 4 //暂定
	ProviderSLA           = 6
	DefaultCapacity int64 = 100     //单位：MB
	DefaultDuration int64 = 10      //单位：天
	DefaultPrice    int64 = 1000000 //单位：wei

	DefaultPassword       = "123456"
	SegementCount   int32 = 256

	//LFS
	maxObjectNameLen = 4096 //设定文件名和路径可占用的最长字节数

	DefaultGetBlockDelay = 30 * time.Second

	BlockSize                    = 1024 * 1024 //暂定一个块中纯data的大小，1k
	defaultMetaBackupCount int32 = 3
	MAXLISTVALUE                 = 100
	flushLocalBackup             = 1
)

//KeeperInfo 此结构体记录Keeper的信息，存储Tendermint地址，让user也能访问链上数据
type KeeperInfo struct {
	IsBFT     bool //标识Keeper组采取的同步方法
	KeeperID  string
	Connected bool
}
type PeersInfo struct {
	Keepers   []*KeeperInfo
	Providers []string
}

//------Contracts Type--------
type ContractService struct {
	UserID        string
	channelBook   map[string]contracts.ChannelItem // 保存该user所部署的channel合约，K-provider地址，V-合约结构体
	offerBook     map[string]contracts.OfferItem   // 保存keeper选择后的provider的offer合约，K-provider地址，V-合约结构体
	upKeepingItem contracts.UpKeepingItem
	queryItem     contracts.QueryItem
}

type GroupService struct {
	Userid         string
	password       string
	initResMutex   sync.Mutex //目前同一时间只回复一个Keeper避免冲突
	localPeersInfo PeersInfo
	tempKeepers    []string //先收集Keeper和Provider暂存，然后到时间挑选（目前是随机，以后可让User自己选）
	tempProviders  []string
	PrivateKey     []byte
	KeySet         *mcl.KeySet
	storeDays      int64 //表示部署合约时的存储数据时间，单位是“天”
	storeSize      int64 //表示部署合约时的存储数据大小，单位是“MB”
	storePrice     int64 //表示部署合约时的存储价格大小，单位是“wei”
	keeperSLA      int   //表示部署合约时的keeper参数，目前是keeper数量
	providerSLA    int   //表示部署合约时的provider参数，目前是provider数量
}

//------LFS Type--------
type LfsService struct {
	CurrentLog *Logs //内存数据结构，存有当前的IpfsNode、SuperBlock和全部的Inode
	UserID     string
	PrivateKey []byte
}

type Logs struct {
	Sb           *pb.SuperBlock
	SbMux        sync.Mutex
	SbModified   bool                                //看看superBlock是否需要更新（仅在新创建Bucket时需要）
	BucketByName map[string]*pb.BucketInfo           //通过BucketName找到Bucket信息
	BucketByID   map[int32]*pb.BucketInfo            //通过BucketID知道到Bucket信息
	Entries      map[int32]map[string]*pb.ObjectInfo //通过BucketID检索Bucket下文件
	State        map[int32]*BucketState              //通过BucketID确定Bucket的状态
}

type BucketState struct {
	Dirty bool
	Mu    sync.Mutex
}

var (
	ErrPolicy                    = errors.New("the policy is error")
	ErrBalance                   = errors.New("your account's balance is insufficient, we will not deploy contract")
	ErrGetSecreteKey             = errors.New("get user's secrete key error")
	ErrKeySetIsNil               = errors.New("user's Keyset is nil")
	ErrUserNotExist              = errors.New("user not exist")
	ErrUserBookIsNil             = errors.New("the User book is nil")
	ErrCannotFindUserInUserBook  = errors.New("cannot find this user in userbook")
	ErrGetContractItem           = errors.New("Can't get contract Item")
	ErrContractServiceAlreadySet = errors.New("this contract Service already set")
	ErrGroupServiceAlreadySet    = errors.New("this group Service already set")
	ErrLfsServiceAlreadySet      = errors.New("this lfs Service already set")
	ErrTimeOut                   = errors.New("Time out")

	ErrNoProviders           = errors.New("there is no providers")
	ErrNoKeepers             = errors.New("there is no keepers")
	ErrCannotConnectKeeper   = errors.New("cannot connect Keeper")
	ErrCannotConnectProvider = errors.New("cannot connect this provider")
	ErrNoEnoughProvider      = errors.New("no Enough Providers")
	ErrNoEnoughKeeper        = errors.New("no Enough Keepers")
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

	ErrObjectNameToolong    = errors.New("the object's name is too long")
	ErrObjectNameInvalid    = errors.New("object name invalid")
	ErrCannotGetEnoughBlock = errors.New("cannot get enough Block")

	ErrCannotLoadMetaBlock  = errors.New("cannot Load MetaBlock")
	ErrCannotAddBlock       = errors.New("cannot Add this block")
	ErrCannotLoadSuperBlock = errors.New("cannot load superblock")
	ErrWrongState           = errors.New("wrong userservice state")
	ErrWrongInitState       = errors.New("wrong init state")
)

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
