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

var persistMetaInterval time.Duration //持久化s检查间隔

// LfsInfo has lfs info
type LfsInfo struct {
	userID     string
	privateKey []byte
	gInfo      *groupInfo
	keySet     *mcl.KeySet
	meta       *lfsMeta //内存数据结构，存有当前的IpfsNode、SuperBlock和全部的Inode
	online     bool
	inProcess  int //atomic
	context    context.Context
	cancelFunc context.CancelFunc
}

// Service defines user's function
type Service interface {
	Start() error
	Stop()
	Flush() error
	ListBucket(prefix string) ([]*pb.BucketInfo, error)
	CreateBucket(bucketName string, options *pb.BucketOptions) (*pb.BucketInfo, error)
	HeadBucket(bucketName string) (*pb.BucketInfo, error)
	DeleteBucket(bucketName string) (*pb.BucketInfo, error)
	DeleteObject(bucketName, objectName string) (*pb.ObjectInfo, error)
	HeadObject(bucketName, objectName string, avail bool) (*pb.ObjectInfo, error)
	PutObject(bucketName, objectName string, reader io.Reader) error
	GetObject(bucketName, objectName string)
	IsOnline() bool
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

// Start starts user's info
func (l *LfsInfo) Start() error {
	// 证明该user已经启动
	if l.online || (l.gInfo != nil && l.gInfo.state >= starting) {
		return errors.New("The user is running")
	}

	l.state = false

	has, err = u.gInfo.start(l.context)
	if err != nil {
		log.Println("start group err: ", err)
		return err
	}

	if has {
		// init or send bls config
		mkey, err := loadBLS12Config(userID, l.gInfo.tempKeepers, l.privateKey)

		if err != nil || mkey == nil {
			log.Println("no bls config err: ", err)
			return err
		}
		l.keySet = mkey
	} else {
		mkey, err := userBLS12ConfigInit()
		if err != nil {
			log.Println("init bls config err: ", err)
			return err
		}

		putUserConfig(userID, l.gInfo.tempKeepers, l.privateKey, mkey)

		l.keySet = mkey
	}

	err = l.startLfs(l.context)
	if err != nil {
		log.Println("StartLfsService()err")
		return err
	}
	l.online = true
	return nil
}

// lfs启动，从本地或者本节点provider处获取LfsMeta信息进行填充，填充不了才进行LfsMeta的初始化操作
//填充顺序：超级块-Bucket数据-Bucket中Object数据
func (l *LfsInfo) startLfs(ctx context.Context) error {
	var err error
	l.meta, err = l.loadSuperBlock() //先加载超级块
	if err != nil || l.meta == nil {
		//启动失败，证明本地无metablock
		log.Println("load superblock fail, so begin to init Lfs :", l.userID)
		l.meta, err = initLfs() //初始化
		if err != nil {
			log.Println(ErrCannotStartLfsService)
			return ErrCannotStartLfsService
		}
	} else {
		err = l.loadBucketInfo() //再加载Group元数据
		if err != nil {          //*错误处理
			return err
		}
		for _, bucket := range l.meta.bucketByID {
			err = l.loadObjectsInfo(bucket) //再加载Object元数据
			if err != nil {
				log.Println(ErrCannotStartLfsService, err)
				return err
			}
			log.Println("objects in bucket-", bucket.Name, "is loaded")
		}
	}
	log.Println(l.userID, " : Lfs Service is ready")
	l.online = true
	go l.persistMetaBlock(ctx)
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

// Start starts user's info
func (l *LfsInfo) Stop() error {
	//用于通知资源释放
	l.cancelFunc()
}

//每隔一段时间，会检查元数据快是否为脏，决定要不要持久化
func (l *LfsInfo) persistMetaBlock(ctx context.Context) error {
	persistMetaInterval = 10 * time.Second
	tick := time.NewTicker(persistMetaInterval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if l.online { //LFS没启动不刷新
				err := l.Fsync(false)
				if err != nil {
					log.Println("Cannot Persist MetaBlock : ", err)
				}
			}
		case <-ctx.Done():
			if l.online { //LFS没启动不刷新
				err := l.Fsync(false)
				if err != nil {
					log.Println("Cannot Persist MetaBlock : ", err)
				}
			}
			return nil
		}
	}
}

//Fsync 现在只刷新metaBlock，以后可以删除数据块的时候先只标记，然后再在Fsync统一刷新
func (l *LfsInfo) Fsync(isForce bool) error {
	if !l.online {
		return ErrLfsServiceNotReady
	}

	l.meta.sb.RLock()
	if l.meta.sb.dirty || isForce { //将超级块信息保存在本地
		err := l.flushSuperBlockLocal()
		if err != nil {
			l.meta.sb.RUnlock()
			return err
		}
		log.Println("Flush Superblock to local finish. The uid is ", l.userid)
	}
	l.meta.sb.RUnlock()

	for _, bucket := range l.meta.bucketByID { //bucket信息和object信息保存在本地
		if bucket.dirty || isForce {
			err := l.flushObjectsInfoLocal(bucket)
			if err != nil {
				return err
			}
			err = l.flushBucketInfoLocal(bucket)
			if err != nil {
				return err
			}
			log.Printf("Flush %s BucketInfo and objects Info to local finish. The uid is %s\n", bucket.Name, l.userid)
		}
	}

	l.meta.sb.RLock()
	if l.meta.sb.dirty || isForce {
		err := l.flushSuperBlockToProvider()
		if err != nil {
			l.meta.sb.RUnlock()
			return err
		}
		l.meta.sb.dirty = false
		log.Println("Flush Superblock to provider finish. The uid is ", l.userID)
	}
	l.meta.sb.RUnlock()

	for _, bucket := range l.meta.bucketByID {
		if bucket.dirty || isForce {
			err := l.flushObjectsInfoToProvider(bucket)
			if err != nil {
				return err
			}
			err = l.flushBucketInfoToProvider(bucket)
			if err != nil {
				return err
			}
			bucket.dirty = false
			log.Printf("Flush %s BucketInfo and objects Info to provider finish. The uid is %s\n", bucket.Name, l.useriID)
		}
	}
	saveChannelValue(l.userID)
	return nil
}
