package user

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/memoio/go-mefs/role/user/pb"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var persistMetaInterval time.Duration //持久化检查间隔

func ConstructLfsService(userID string, privKey []byte) *LfsService {
	return &LfsService{
		UserID:     userID,
		PrivateKey: privKey,
	}
}

func (lfs *LfsService) StartLfsService(ctx context.Context) error {
	err := lfs.startLfs(ctx)
	if err != nil {
		fmt.Println("Lfs start err : ", err)
		return err
	}
	go lfs.PersistMetaBlock(ctx)
	return nil
}

//lfs节点启动，从本地或者本节点provider处获取currentLog信息进行填充，填充不了才进行currentlog的初始化操作
//填充顺序：超级块-Bucket数据-Bucket中Object数据
func (lfs *LfsService) startLfs(ctx context.Context) error {
	var err error
	lfs.CurrentLog, err = lfs.loadSuperBlock() //先加载超级块
	if err != nil || lfs.CurrentLog == nil {
		log.Println("Cannot get metaBlock", err)
		//启动失败，证明本地无metablock
		state, err := GetUserServiceState(lfs.UserID)
		if err != nil {
			return err
		}
		if state < GroupStarted { //这里 保证groupservice已经启动完成
			tick := time.Tick(10 * time.Second)
			tickCount := 0
		LoopNoInit:
			for {
				select {
				case <-tick:
					if tickCount >= 60 { //超过十分钟还没有Keeper，出故障了
						fmt.Println("Cannot start lfs service-", ErrNoKeepers)
						return ErrNoKeepers
					}

					state, err := GetUserServiceState(lfs.UserID)
					if err != nil {
						return err
					}
					if state >= GroupStarted {
						break LoopNoInit
					}
					tickCount++
				case <-ctx.Done():
					return nil
				}
			}
		}
		lfs.CurrentLog, err = lfs.loadSuperBlock() //找到keeper再加载一次超级块
		if err != nil || lfs.CurrentLog == nil {
			fmt.Println("load superblock fail, so begin to init Lfs :", lfs.UserID)
			lfs.CurrentLog, err = initLfs() //初始化
			if err != nil {
				log.Println(ErrCannotStartLfsService)
				return ErrCannotStartLfsService
			}
			fmt.Println(lfs.UserID + " : Lfs Service is ready")
			err = SetUserState(lfs.UserID, BothStarted)
			if err != nil {
				fmt.Println("SetUserState failed")
			}
			return nil
		}
	}
	err = lfs.loadBucketInfo() //再加载Group元数据
	if err != nil {            //*错误处理
		return err
	}
	for _, Bucket := range lfs.CurrentLog.BucketByID {
		err = lfs.loadObjectsInfo(Bucket) //再加载Object元数据
		if err != nil {
			log.Println(ErrCannotStartLfsService, err)
			return err
		}
		fmt.Println("Objects in bucket-", Bucket.BucketName, "is loaded")
	}
	fmt.Println(lfs.UserID + " : Lfs Service is ready")
	err = SetUserState(lfs.UserID, BothStarted)
	if err != nil {
		fmt.Println("SetUserState failed")
	}
	return nil
}

func initLfs() (*Logs, error) {
	log, err := InitLogs()
	if err != nil {
		return nil, err
	}
	return log, err
}

func InitLogs() (*Logs, error) {
	sb := newSuperBlock()
	bucketByID := make(map[int32]*Bucket)
	bucketNameToID := make(map[string]int32)
	return &Logs{
		Sb:             sb,
		BucketByID:     bucketByID,
		BucketNameToID: bucketNameToID,
	}, nil
}

func newSuperBlock() *SuperBlock {
	buckets := make(map[int32]string)
	bitset := bitset.New(256)
	return &SuperBlock{
		SuperBlockInfo: pb.SuperBlockInfo{
			BucketsSet:      nil,
			MetaBackupCount: defaultMetaBackupCount,
			NextBucketID:    1, //从1开始是因为SuperBlock的元数据块抢占了Bucket编号0的位置
			MagicNumber:     0xfb,
			Version:         1},
		Buckets: buckets,
		Bitset:  bitset,
		Dirty:   true,
	}
}

//每隔一段时间，会检查元数据快是否为脏，决定要不要持久化
func (lfs *LfsService) PersistMetaBlock(ctx context.Context) error {
	persistMetaInterval = 10 * time.Second
	tick := time.NewTicker(persistMetaInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			state, err := GetUserServiceState(lfs.UserID)
			if err != nil {
				return err
			}
			if state == BothStarted { //LFS没启动不刷新
				err := lfs.Fsync(false)
				if err != nil {
					log.Println("Cannot Persist MetaBlock : ", err)
				}
			}
		case <-ctx.Done():
			state, err := GetUserServiceState(lfs.UserID)
			if err != nil {
				return err
			}
			if state == BothStarted { //LFS没启动不刷新
				err := lfs.Fsync(false)
				if err != nil {
					log.Println("Cannot Persist MetaBlock : ", err)
				}
			}
			return nil
		}
	}
}

//现在只刷新metaBlock，以后可以删除数据块的时候先只标记，然后再在Fsync统一刷新
func (lfs *LfsService) Fsync(isForce bool) error {
	state, err := GetUserServiceState(lfs.UserID)
	if err != nil {
		return err
	}
	if state < GroupStarted {
		return ErrGroupServiceNotReady
	}
	if state < BothStarted {
		return ErrLfsIsNotRunning
	}
	// 持久化保存channel的price
	cs := GetContractService(lfs.UserID)
	if cs == nil {
		return ErrUserNotExist
	}
	gp := GetGroupService(lfs.UserID)
	providers, err := gp.GetProviders(-1)
	if err != nil {
		return err
	}
	for _, provider := range providers {
		channel, err := cs.GetChannelItem(provider)
		if err != nil {
			fmt.Println("GetChannelItem err:", provider, err)
			continue
		}
		// 保存本地形式：K-provider，V-channel此时的value
		km, err := metainfo.NewKeyMeta(channel.ChannelAddr, metainfo.Local, metainfo.SyncTypeChannelValue)
		if err != nil {
			fmt.Println("NewKeyMeta err:", provider, err)
			continue
		}
		err = localNode.Routing.(*dht.IpfsDHT).CmdPutTo(km.ToString(), channel.Value.String(), "local")
		if err != nil {
			fmt.Println("CmdPutTo error", provider, err)
			continue
		}
	}
	if lfs.CurrentLog.Sb.Dirty || isForce { //将超级块信息保存在本地
		err := lfs.flushSuperBlockLocal(lfs.CurrentLog.Sb)
		if err != nil {
			return err
		}
		fmt.Println("Flush Superblock to local finish. The uid is ", lfs.UserID)
	}

	for _, bucket := range lfs.CurrentLog.BucketByID { //bucket信息和object信息保存在本地
		if bucket.Dirty || isForce {
			err := lfs.flushObjectsInfoLocal(bucket)
			if err != nil {
				return err
			}
			err = lfs.flushBucketInfoLocal(bucket)
			if err != nil {
				return err
			}
			fmt.Printf("Flush %s BucketInfo and Objects Info to local finish. The uid is %s\n", bucket.BucketName, lfs.UserID)
		}
	}

	if lfs.CurrentLog.Sb.Dirty || isForce {
		err := lfs.flushSuperBlockToProvider(lfs.CurrentLog.Sb)
		if err != nil {
			return err
		}
		lfs.CurrentLog.Sb.Dirty = false
		fmt.Println("Flush Superblock to provider finish. The uid is ", lfs.UserID)
	}

	for _, bucket := range lfs.CurrentLog.BucketByID {
		if bucket.Dirty || isForce {
			err := lfs.flushObjectsInfoToProvider(bucket)
			if err != nil {
				return err
			}
			err = lfs.flushBucketInfoToProvider(bucket)
			if err != nil {
				return err
			}
			bucket.Dirty = false
			fmt.Printf("Flush %s BucketInfo and Objects Info to provider finish. The uid is %s\n", bucket.BucketName, lfs.UserID)
		}
	}

	return nil
}

//检查文件名合法性，文件名中不能含有"/"
func checkObjectName(objectName string) error {
	if len(objectName) > maxObjectNameLen {
		return ErrObjectNameToolong
	}
	if objectName == "" || len(objectName) == 0 {
		return ErrObjectNameInvalid
	}
	for i := 0; i < len(objectName); i++ {
		if objectName[i] == '/' || objectName[i] == '\\' || objectName[i] == '\n' {
			return ErrObjectNameInvalid
		}
	}
	return nil
}

func isStart(uid string) error {
	state, err := GetUserServiceState(uid)
	if err != nil {
		return err
	}
	switch state {
	case Starting, Collecting, CollectCompleted:
		return ErrGroupServiceNotReady
	case GroupStarted:
		return ErrLfsIsNotRunning
	case BothStarted:
		return nil
	default:
		return ErrWrongState
	}
}
