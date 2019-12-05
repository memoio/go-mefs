package user

import (
	"container/list"
	"context"
	"log"
	"sync"
	"time"

	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils/bitset"
)

var persistMetaInterval time.Duration //持久化检查间隔

// LfsService has lfs info
type LfsService struct {
	userid     string
	meta       *lfsMeta //内存数据结构，存有当前的IpfsNode、SuperBlock和全部的Inode
	inProcess  int      //表示此lfs上是否有操作，如上传下载，避免过程中user被Kill
	privateKey []byte
}

// Logs records lfs metainfo
type lfsMeta struct {
	sb             *superBlock
	bucketNameToID map[string]int32       //通过BucketName找到Bucket信息
	bucketByID     map[int32]*superBucket //通过BucketID知道到Bucket信息
}

// superBlock has lfs bucket info
type superBlock struct {
	pb.SuperBlockInfo
	bitsetInfo *bitset.BitSet
	sync.RWMutex
	dirty bool //看看superBlock是否需要更新（仅在新创建Bucket时需要）
	state int
}

// superBucket has lfs objects info
type superBucket struct {
	pb.BucketInfo
	objects        map[string]*list.Element //通过BucketID检索Bucket下文件
	orderedObjects *list.List               //用过map和list结合，构造一个有序Map
	dirty          bool
	sync.RWMutex
	state int
}

// objectInfo stores an object meta info
type objectInfo struct {
	pb.ObjectInfo
	sync.RWMutex
	state int
}

func constructLfsService(userID string, privKey []byte) *LfsService {
	return &LfsService{
		userid:     userID,
		privateKey: privKey,
	}
}

// StartLfsService is
func (lfs *LfsService) StartLfsService(ctx context.Context) error {
	err := lfs.startLfs(ctx)
	if err != nil {
		log.Println("Lfs start err : ", err)
		return err
	}
	go lfs.persistMetaBlock(ctx)
	return nil
}

//lfs节点启动，从本地或者本节点provider处获取LfsMeta信息进行填充，填充不了才进行LfsMeta的初始化操作
//填充顺序：超级块-Bucket数据-Bucket中Object数据
func (lfs *LfsService) startLfs(ctx context.Context) error {
	var err error
	lfs.meta, err = lfs.loadSuperBlock() //先加载超级块
	if err != nil || lfs.meta == nil {
		log.Println("Cannot get metaBlock", err)
		//启动失败，证明本地无metablock
		state, err := getUserState(lfs.userid)
		if err != nil {
			return err
		}
		if state < groupStarted { //这里 保证groupservice已经启动完成
			tick := time.Tick(10 * time.Second)
			tickCount := 0
		LoopNoInit:
			for {
				select {
				case <-tick:
					if tickCount >= 60 { //超过十分钟还没有Keeper，出故障了
						log.Println("Cannot start lfs service-", ErrNoKeepers)
						return ErrNoKeepers
					}

					state, err := getUserState(lfs.userid)
					if err != nil {
						return err
					}
					if state >= groupStarted {
						break LoopNoInit
					}
					tickCount++
				case <-ctx.Done():
					return nil
				}
			}
		}
		lfs.meta, err = lfs.loadSuperBlock() //找到keeper再加载一次超级块
		if err != nil || lfs.meta == nil {
			log.Println("load superblock fail, so begin to init Lfs :", lfs.userid)
			lfs.meta, err = initLfs() //初始化
			if err != nil {
				log.Println(ErrCannotStartLfsService)
				return ErrCannotStartLfsService
			}
			log.Println(lfs.userid + " : Lfs Service is ready")
			err = setUserState(lfs.userid, bothStarted)
			if err != nil {
				log.Println("setUserState failed")
			}
			return nil
		}
	}
	err = lfs.loadBucketInfo() //再加载Group元数据
	if err != nil {            //*错误处理
		return err
	}
	for _, bucket := range lfs.meta.bucketByID {
		err = lfs.loadObjectsInfo(bucket) //再加载Object元数据
		if err != nil {
			log.Println(ErrCannotStartLfsService, err)
			return err
		}
		log.Println("objects in bucket-", bucket.Name, "is loaded")
	}
	log.Println(lfs.userid + " : Lfs Service is ready")
	err = setUserState(lfs.userid, bothStarted)
	if err != nil {
		log.Println("setUserState failed")
	}
	return nil
}

func initLfs() (*lfsMeta, error) {
	log, err := initLogs()
	if err != nil {
		return nil, err
	}
	return log, err
}

func initLogs() (*lfsMeta, error) {
	sb := newSuperBlock()
	bucketByID := make(map[int32]*superBucket)
	bucketNameToID := make(map[string]int32)
	return &lfsMeta{
		sb:             sb,
		bucketByID:     bucketByID,
		bucketNameToID: bucketNameToID,
	}, nil
}

func newSuperBlock() *superBlock {
	bitset := bitset.New(256)
	return &superBlock{
		SuperBlockInfo: pb.SuperBlockInfo{
			BucketsSet:      nil,
			MetaBackupCount: defaultMetaBackupCount,
			NextBucketID:    1, //从1开始是因为SuperBlock的元数据块抢占了Bucket编号0的位置
			MagicNumber:     0xfb,
			Version:         1},
		bitsetInfo: bitset,
		dirty:      true,
	}
}

//每隔一段时间，会检查元数据快是否为脏，决定要不要持久化
func (lfs *LfsService) persistMetaBlock(ctx context.Context) error {
	persistMetaInterval = 10 * time.Second
	tick := time.NewTicker(persistMetaInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			state, err := getUserState(lfs.userid)
			if err != nil {
				return err
			}
			if state == bothStarted { //LFS没启动不刷新
				err := lfs.Fsync(false)
				if err != nil {
					log.Println("Cannot Persist MetaBlock : ", err)
				}
			}
		case <-ctx.Done():
			state, err := getUserState(lfs.userid)
			if err != nil {
				return err
			}
			if state == bothStarted { //LFS没启动不刷新
				err := lfs.Fsync(false)
				if err != nil {
					log.Println("Cannot Persist MetaBlock : ", err)
				}
			}
			return nil
		}
	}
}

//Fsync 现在只刷新metaBlock，以后可以删除数据块的时候先只标记，然后再在Fsync统一刷新
func (lfs *LfsService) Fsync(isForce bool) error {
	state, err := getUserState(lfs.userid)
	if err != nil {
		return err
	}
	if state < groupStarted {
		return ErrGroupServiceNotReady
	}
	if state < bothStarted {
		return ErrLfsIsNotRunning
	}

	lfs.meta.sb.RLock()
	if lfs.meta.sb.dirty || isForce { //将超级块信息保存在本地
		err := lfs.flushSuperBlockLocal()
		if err != nil {
			return err
		}
		log.Println("Flush Superblock to local finish. The uid is ", lfs.userid)
	}
	for _, bucket := range lfs.meta.bucketByID { //bucket信息和object信息保存在本地
		if bucket.dirty || isForce {
			err := lfs.flushObjectsInfoLocal(bucket)
			if err != nil {
				return err
			}
			err = lfs.flushBucketInfoLocal(bucket)
			if err != nil {
				return err
			}
			log.Printf("Flush %s BucketInfo and objects Info to local finish. The uid is %s\n", bucket.Name, lfs.userid)
		}
	}

	if lfs.meta.sb.dirty || isForce {
		err := lfs.flushSuperBlockToProvider()
		if err != nil {
			return err
		}
		lfs.meta.sb.dirty = false
		log.Println("Flush Superblock to provider finish. The uid is ", lfs.userid)
	}
	lfs.meta.sb.RUnlock()
	for _, bucket := range lfs.meta.bucketByID {
		if bucket.dirty || isForce {
			err := lfs.flushObjectsInfoToProvider(bucket)
			if err != nil {
				return err
			}
			err = lfs.flushBucketInfoToProvider(bucket)
			if err != nil {
				return err
			}
			bucket.dirty = false
			log.Printf("Flush %s BucketInfo and objects Info to provider finish. The uid is %s\n", bucket.Name, lfs.userid)
		}
	}
	saveChannelValue(lfs.userid)
	return nil
}

func isStart(uid string) error {
	state, err := getUserState(uid)
	if err != nil {
		return err
	}
	switch state {
	case starting, collecting, collectCompleted:
		return ErrGroupServiceNotReady
	case groupStarted:
		return ErrLfsIsNotRunning
	case bothStarted:
		return nil
	default:
		return ErrWrongState
	}
}
