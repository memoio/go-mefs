package user

import (
	"bytes"
	"container/list"
	"context"
	"errors"
	"io"
	"log"
	"strconv"
	"sync"
	"time"

	ggio "github.com/gogo/protobuf/io"
	mcl "github.com/memoio/go-mefs/bls12"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/source/data"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var persistMetaInterval time.Duration //持久化s检查间隔

const metaTagFlag = dataformat.BLS12

// LfsInfo has lfs info
type LfsInfo struct {
	userID     string
	fsID       string // use query addr as fsID
	privateKey []byte
	gInfo      *groupInfo
	ds         data.Service
	keySet     *mcl.KeySet
	meta       *lfsMeta //内存数据结构，存有当前的IpfsNode、SuperBlock和全部的Inode
	online     bool
	inProcess  int //atomic
	context    context.Context
	cancelFunc context.CancelFunc
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
	if l.online || (l.gInfo != nil && l.gInfo.state > starting) {
		return errors.New("The user is running")
	}

	l.online = false

	has, err := l.gInfo.start(l.context)
	if err != nil {
		log.Println("start group err: ", err)
		return err
	}

	if has {
		// init or send bls config
		err = l.loadBLS12Config()
		if err != nil {
			log.Println("load bls config err: ", err)
		}
	}
	if !has || err != nil {
		mkey, err := initBLS12Config()
		if err != nil {
			log.Println("init bls config err: ", err)
			return err
		}

		l.keySet = mkey
		l.putUserConfig()
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
		log.Println("load superblock fail, so begin to init Lfs :", l.fsID)
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
	log.Println("Lfs Service is ready for: ", l.userID)
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

// Stop user's info
func (l *LfsInfo) Stop() error {
	//用于通知资源释放
	l.cancelFunc()
	return nil
}

// Online user's info
func (l *LfsInfo) Online() bool {
	//用于通知资源释放
	return l.online
}

func (l *LfsInfo) GetGroup() *groupInfo {
	return l.gInfo
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
		log.Println("Flush Superblock to local finish. The uid is ", l.userID)
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
			log.Printf("Flush %s BucketInfo and objects Info to local finish. The uid is %s\n", bucket.Name, l.userID)
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
			log.Printf("Flush %s BucketInfo and objects Info to provider finish. The uid is %s\n", bucket.Name, l.userID)
		}
	}
	return nil
}

//----------------------Flush superBlock---------------------------

//刷新超级块
func (l *LfsInfo) flushSuperBlock() error {
	err := l.flushSuperBlockLocal()
	if err != nil {
		return err
	}
	return l.flushSuperBlockToProvider()
}

//保存超级块信息到本地，传入参数为超级快结构体
func (l *LfsInfo) flushSuperBlockLocal() error {
	sb := l.meta.sb
	sb.BucketsSet = sb.bitsetInfo.Bytes()
	SbBuffer := bytes.NewBuffer(nil)
	SbDelimitedWriter := ggio.NewDelimitedWriter(SbBuffer)
	err := SbDelimitedWriter.WriteMsg(&sb.SuperBlockInfo)
	if err != nil {
		log.Println("SbDelimitedWriter.WriteMsg(sb) failed ", err)
		return err
	}
	err = SbDelimitedWriter.Close()
	if err != nil {
		log.Println("SbDelimitedWriter.Close() failed ", err)
		return err
	}

	data := SbBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}

	bm, err := metainfo.NewBlockMeta(l.fsID, "0", "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, _, err := dataformat.DataEncodeToMul(data, ncidPrefix, 1, 0, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
	if err != nil {
		return err
	}
	ncid := bm.ToString()
	km, err := metainfo.NewKeyMeta(ncid, metainfo.Block)
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = l.ds.DeleteBlock(ctx, km.ToString(), "local")
	if err != nil && err != bs.ErrNotFound {
		return err
	}

	err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[0], "local")
	if err != nil {
		return ErrCannotAddBlock
	}

	return nil
}

func (l *LfsInfo) flushSuperBlockToProvider() error {
	sb := l.meta.sb
	sb.BucketsSet = sb.bitsetInfo.Bytes()
	SbBuffer := bytes.NewBuffer(nil)
	SbDelimitedWriter := ggio.NewDelimitedWriter(SbBuffer)
	err := SbDelimitedWriter.WriteMsg(&sb.SuperBlockInfo)
	if err != nil {
		log.Println("SbDelimitedWriter.WriteMsg(sb) failed ", err)
		return err
	}
	err = SbDelimitedWriter.Close()
	if err != nil {
		log.Println("SbDelimitedWriter.Close() failed ", err)
		return err
	}

	data := SbBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}

	bm, err := metainfo.NewBlockMeta(l.fsID, "0", "0", "")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, offset, err := dataformat.DataEncodeToMul(data, ncidPrefix, 1, sb.MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
	if err != nil {
		return err
	}
	providers, _, err := l.gInfo.GetProviders(int(sb.MetaBackupCount))
	if err != nil && len(providers) == 0 {
		return err
	}

	ctx := context.Background()

	for j := 0; j < len(providers); j++ { //
		bm.SetCid(strconv.Itoa(j))
		ncid := bm.ToString()

		km, err := metainfo.NewKeyMeta(ncid, metainfo.Block)
		if err != nil {
			continue
		}
		updateKey := km.ToString()

		err = l.ds.PutBlock(ctx, updateKey, dataEncoded[j], providers[j])
		if err != nil {
			return err
		}
		err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
		if err != nil {
			return err
		}
	}
	return nil
}

//-----------------------Flush BucketMeta----------------------------

func (l *LfsInfo) flushBucketInfo(bucket *superBucket) error {
	err := l.flushBucketInfoLocal(bucket)
	if err != nil {
		return err
	}
	return l.flushBucketInfoToProvider(bucket)
}

func (l *LfsInfo) flushBucketInfoLocal(bucket *superBucket) error {
	bucket.RLock()
	defer bucket.RUnlock()
	BucketBuffer := bytes.NewBuffer(nil)
	BucketDelimitedWriter := ggio.NewDelimitedWriter(BucketBuffer)
	err := BucketDelimitedWriter.WriteMsg(&bucket.BucketInfo)
	if err != nil {
		return err
	}

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, _, err := dataformat.DataEncodeToMul(BucketBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
	if err != nil {
		return err
	}
	err = BucketDelimitedWriter.Close()
	if err != nil {
		return err
	}
	ncid := bm.ToString()
	km, err := metainfo.NewKeyMeta(ncid, metainfo.Block)
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = l.ds.DeleteBlock(ctx, km.ToString(), "local")
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[0], "local")
	if err != nil {
		return err
	}
	return nil
}

func (l *LfsInfo) flushBucketInfoToProvider(bucket *superBucket) error {
	bucket.RLock()
	defer bucket.RUnlock()
	MetaBackupCount := l.meta.sb.MetaBackupCount
	providers, _, err := l.gInfo.GetProviders(int(MetaBackupCount))
	if err != nil && len(providers) == 0 {
		return err
	}
	BucketBuffer := bytes.NewBuffer(nil)
	BucketDelimitedWriter := ggio.NewDelimitedWriter(BucketBuffer)
	err = BucketDelimitedWriter.WriteMsg(&bucket.BucketInfo)
	if err != nil {
		return err
	}

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	BucketBytes := BucketBuffer.Bytes()
	dataEncoded, offset, err := dataformat.DataEncodeToMul(BucketBytes, ncidPrefix, 1, MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)

	ctx := context.Background()

	for j := 0; j < int(MetaBackupCount); j++ { //
		bm.SetCid(strconv.Itoa(j))
		ncid := bm.ToString()
		if err != nil {
			return err
		}
		km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)
		err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[j], providers[j])
		if err != nil {
			return err
		}

		err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
		if err != nil {
			return err
		}
	}
	err = BucketDelimitedWriter.Close()
	if err != nil {
		return err
	}
	return nil
}

//---------------------Flush objects' Meta for given superBucket--------

//刷新具体某一个Bucket的object数据
func (l *LfsInfo) flushObjectsInfo(bucket *superBucket) error {
	if bucket == nil || bucket.objects == nil {
		return nil
	}
	err := l.flushObjectsInfoLocal(bucket)
	if err != nil {
		return err
	}
	return l.flushObjectsInfoToProvider(bucket)
}

func (l *LfsInfo) flushObjectsInfoLocal(bucket *superBucket) error {
	if bucket == nil || bucket.objects == nil {
		return nil
	}
	bucket.RLock()
	defer bucket.RUnlock()
	objectsBuffer := bytes.NewBuffer(nil)
	objectDelimitedWriter := ggio.NewDelimitedWriter(objectsBuffer)

	bucketID := bucket.BucketID
	objectsStripeID := 1 //ObjectInfo的stripe从1开始
	objectsBlockLength := 0
	ctx := context.Background()
	for objectElement := bucket.orderedObjects.Front(); objectElement != nil; objectElement = objectElement.Next() {
		object, ok := objectElement.Value.(*objectInfo)
		if !ok {
			continue
		}
		if objectsBuffer.Len() >= utils.BlockSize { //如果object的总长度大于规定的size，则分块
			objectsBlockLength += objectsBuffer.Len()

			bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
			if err != nil {
				return err
			}
			ncidPrefix := bm.ToString(3)
			dataEncoded, _, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
			if err != nil {
				return err
			}
			ncid := bm.ToString()

			km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)

			err = l.ds.DeleteBlock(ctx, km.ToString(), "local")
			if err != nil && err != bs.ErrNotFound {
				return err
			}
			err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[0], "local")
			if err != nil {
				return ErrCannotAddBlock
			}
			err = objectDelimitedWriter.Close()
			if err != nil {
				return err
			}
			objectsBuffer = bytes.NewBuffer(nil) //重新开始处理下一个块
			objectDelimitedWriter = ggio.NewDelimitedWriter(objectsBuffer)
			objectsStripeID++
		}
		err := objectDelimitedWriter.WriteMsg(&object.ObjectInfo)
		if err != nil {
			return err
		}
	}

	if objectsBuffer.Len() != 0 { //处理最后的剩余部分
		objectsBlockLength += objectsBuffer.Len()
		bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
		if err != nil {
			return err
		}
		ncidPrefix := bm.ToString(3)
		dataEncoded, _, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
		if err != nil {
			return err
		}
		ncid := bm.ToString()
		km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)

		err = l.ds.DeleteBlock(ctx, km.ToString(), "local")
		if err != nil && err != bs.ErrNotFound {
			return err
		}
		err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[0], "local")
		if err != nil {
			return ErrCannotAddBlock
		}
		err = objectDelimitedWriter.Close() //结束
		if err != nil {
			return err
		}
	}
	l.meta.bucketByID[bucketID].ObjectsBlockSize = int64(objectsBlockLength)
	return nil
}

func (l *LfsInfo) flushObjectsInfoToProvider(bucket *superBucket) error {
	if bucket == nil || bucket.objects == nil {
		return nil
	}
	bucket.RLock()
	defer bucket.RUnlock()
	MetaBackupCount := l.meta.sb.MetaBackupCount
	providers, _, err := l.gInfo.GetProviders(int(MetaBackupCount))
	if err != nil && len(providers) == 0 {
		return err
	}
	objectsBuffer := bytes.NewBuffer(nil)
	objectDelimitedWriter := ggio.NewDelimitedWriter(objectsBuffer)

	bucketID := bucket.BucketID
	objectsStripeID := 1
	objectsBlockLength := 0

	ctx := context.Background()

	for objectElement := bucket.orderedObjects.Front(); objectElement != nil; objectElement = objectElement.Next() {
		object, ok := objectElement.Value.(*objectInfo)
		if !ok {
			continue
		}
		if objectsBuffer.Len() >= utils.BlockSize { //如果object的总长度大于规定的size，则分块
			objectsBlockLength += objectsBuffer.Len()
			bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
			if err != nil {
				return err
			}
			ncidPrefix := bm.ToString(3)
			dataEncoded, offset, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, 1, MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
			if err != nil {
				return err
			}
			for j := 0; j < len(providers); j++ {
				bm.SetCid(strconv.Itoa(j))
				ncid := bm.ToString()
				km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)

				err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[j], providers[j])
				if err != nil {
					return err
				}

				err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
				if err != nil {
					return err
				}
			}
			objectsStripeID++
			err = objectDelimitedWriter.Close()
			if err != nil {
				return err
			}
			objectsBuffer = bytes.NewBuffer(nil) //重新开始处理下一个块
			objectDelimitedWriter = ggio.NewDelimitedWriter(objectsBuffer)
		}
		err := objectDelimitedWriter.WriteMsg(&object.ObjectInfo)
		if err != nil {
			return err
		}
	}

	if objectsBuffer.Len() != 0 { //处理最后的剩余部分
		objectsBlockLength += objectsBuffer.Len()
		bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
		if err != nil {
			return err
		}
		ncidPrefix := bm.ToString(3)
		dataEncoded, offset, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, 1, MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, l.keySet)
		if err != nil {
			return err
		}
		for j := 0; j < len(providers); j++ {
			bm.SetCid(strconv.Itoa(j))
			ncid := bm.ToString()
			km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)
			err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[j], providers[j])
			if err != nil {
				return err
			}

			err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
			if err != nil {
				return err
			}
		}
		err = objectDelimitedWriter.Close()
		if err != nil {
			return err
		}
	}
	l.meta.bucketByID[bucketID].ObjectsBlockSize = int64(objectsBlockLength)
	return nil
}

//--------------------Load superBlock--------------------------
//lfs启动时加载超级块操作，返回结构体Meta,主要填充其中的superblock字段
//先从本地查找超级快信息，若没找到，就找自己的provider获取
func (l *LfsInfo) loadSuperBlock() (*lfsMeta, error) {
	log.Println("Begin to load superblock : ", l.fsID, "for user:", l.userID)
	var b blocks.Block
	var err error
	sig, err := BuildSignMessage()
	if err != nil {
		return nil, err
	}

	bm, err := metainfo.NewBlockMeta(l.fsID, "0", "0", "0")
	if err != nil {
		return nil, err
	}
	ncidlocal := bm.ToString()
	km, _ := metainfo.NewKeyMeta(ncidlocal, metainfo.Block)

	if l.keySet == nil {
		return nil, ErrKeySetIsNil
	}

	ctx := context.Background()

	b, err = l.ds.GetBlock(ctx, km.ToString(), nil, "local")
	if err == nil && b != nil && dataformat.VerifyBlock(b.RawData(), ncidlocal, l.keySet.Pk) { //如果本地有这个块的话，无需麻烦Provider
	} else { //若本地无超级块，向自己的provider进行查询
		err = l.ds.DeleteBlock(ctx, km.ToString(), "local")
		if err != nil && err != bs.ErrNotFound {
			return nil, err
		}
		log.Println("Try to get it from remote servers:", ncidlocal)
		for j := 0; j < int(defaultMetaBackupCount); j++ {
			bm.SetCid(strconv.Itoa(j))
			ncid := bm.ToString()
			provider, _, err := l.gInfo.getBlockProviders(ncid) //获取数据块的保存位置
			if (provider == "" || err != nil) && j < int(defaultMetaBackupCount)-1 {
				continue
			} else if err != nil {
				log.Println("Cannot load Lfs superblock.", err)
				return nil, ErrCannotLoadMetaBlock
			}

			km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)

			b, err = l.ds.GetBlock(ctx, km.ToString(), sig, provider) //向指定provider查询超级块
			if err != nil {                                           //*错误处理
				log.Printf("Get metablock %s from %s failed: %s.\n", ncid, provider, err)
				continue
			}
			if b != nil { //获取到有效数据块，跳出
				if ok := dataformat.VerifyBlock(b.RawData(), ncid, l.keySet.Pk); !ok {
					log.Println("Verify Block failed.", ncid, "from:", provider)
				} else {
					log.Println("load superblock in block", ncid, "from Provider", provider)
					break
				}
			}
		}
	}

	if b != nil {
		data, err := dataformat.GetDataFromRawData(b.RawData()) //Tag暂时没用
		if err != nil {                                         //*错误处理
			log.Println("GetDataFromRawData err!", err)
			return nil, err
		}
		pbSuperBlock := pb.SuperBlockInfo{}
		SbBuffer := bytes.NewBuffer(data)
		SbDelimitedReader := ggio.NewDelimitedReader(SbBuffer, 5*utils.BlockSize)
		err = SbDelimitedReader.ReadMsg(&pbSuperBlock)
		if err == io.EOF {
		} else if err != nil {
			log.Println("Cannot load Lfs superblock.", err)
			return nil, err
		}
		bucketByID := make(map[int32]*superBucket)
		bucketNameToID := make(map[string]int32)

		log.Println("Lfs superBlock is loaded.")
		return &lfsMeta{
			sb: &superBlock{
				SuperBlockInfo: pbSuperBlock,
				dirty:          false,
				bitsetInfo:     bitset.From(pbSuperBlock.BucketsSet),
			},
			bucketByID:     bucketByID,
			bucketNameToID: bucketNameToID,
		}, nil
	}
	log.Println("Cannot load Lfs superblock. Get metablock failed")
	return nil, ErrCannotLoadSuperBlock
}

//----------------------------Load BucketInfo-----------------------------------
//lfs启动进行元数据的加载，对Log中的字段进行初始化 填充除superblock、Entries字段之外的字段
func (l *LfsInfo) loadBucketInfo() error {
	sig, err := BuildSignMessage()
	if err != nil {
		return err
	}

	for bucketID, ok := l.meta.sb.bitsetInfo.NextSet(0); ok; bucketID, ok = l.meta.sb.bitsetInfo.NextSet(bucketID + 1) {
		if !ok {
			break
		}
		var b blocks.Block
		var err error
		bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), "0", "0")
		if err != nil {
			return err
		}
		ncidlocal := bm.ToString()
		ctx := context.Background()
		if b, err = l.ds.GetBlock(ctx, ncidlocal, nil, "local"); b != nil && err == nil && dataformat.VerifyBlock(b.RawData(), ncidlocal, l.keySet.Pk) { //如果本地有这个块的话，无需麻烦Provider
		} else {
			err = l.ds.DeleteBlock(ctx, ncidlocal, "local")
			if err != nil && err != bs.ErrNotFound {
				return err
			}
			log.Printf("Cannot Get BucketInfo in block %s from local datastore. Maybe block is lost or broken.\n", ncidlocal)
			ncidprefix := bm.ToString(3)
			for j := 0; j < int(l.meta.sb.MetaBackupCount); j++ {
				ncid := ncidprefix + "_" + strconv.Itoa(j)
				provider, _, err := l.gInfo.getBlockProviders(ncid) //获取保存位置
				if err != nil && j == int(l.meta.sb.MetaBackupCount)-1 {
					log.Printf("load superBucket: %d's block: %s from provider: %s falied.\n", bucketID, ncid, provider)
					continue
				}
				b, err = l.ds.GetBlock(ctx, ncid, sig, provider)
				if b != nil && err == nil {
					if ok := dataformat.VerifyBlock(b.RawData(), ncid, l.keySet.Pk); !ok {
						log.Println("Verify Block failed.", ncid, "from:", provider)
					} else {
						break
					}
				} else if err != nil && j == int(l.meta.sb.MetaBackupCount)-1 {
					log.Println("load superBucket error:", bucketID, err)
				}
			}
		}

		if b != nil {
			data, err := dataformat.GetDataFromRawData(b.RawData()) //Tag暂时没用
			if err != nil {
				log.Println("GetDataFromRawData err!", err)
				return err
			}
			bucket := pb.BucketInfo{}
			BucketBuffer := bytes.NewBuffer(data)
			BucketDelimitedReader := ggio.NewDelimitedReader(BucketBuffer, 5*utils.BlockSize)
			err = BucketDelimitedReader.ReadMsg(&bucket)
			if err != nil && err != io.EOF {
				return err
			}
			objects := make(map[string]*list.Element)
			l.meta.bucketByID[int32(bucketID)] = &superBucket{
				BucketInfo:     bucket,
				objects:        objects,
				orderedObjects: list.New(),
				dirty:          false,
			}
			l.meta.bucketNameToID[bucket.Name] = bucket.BucketID
			log.Println("superBucket-ID:", bucket.BucketID, "Name-", bucket.Name, "is loaded")
		}
	}
	return nil
}

//------------------------------Load Objectinfo---------------------------------------
//填充Entries字段，传入参数为bucket,记录传入bucket的数据信息
func (l *LfsInfo) loadObjectsInfo(bucket *superBucket) error {
	sig, err := BuildSignMessage()
	if err != nil {
		return err
	}
	ObjectsBlockSize := bucket.ObjectsBlockSize
	fullData := make([]byte, 0, ObjectsBlockSize)
	if ObjectsBlockSize == 0 { //证明此Bucket一个文件都没有
		return nil
	}
	var readCount int
	stripeID := 1 //ObjectsBlock的Stripe从1开始计算
	for {
		var b blocks.Block
		var err error
		bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), strconv.Itoa(stripeID), "0")
		if err != nil {
			return err
		}
		ncidlocal := bm.ToString()
		ctx := context.Background()
		if b, err = l.ds.GetBlock(ctx, ncidlocal, nil, "local"); b != nil && err == nil && dataformat.VerifyBlock(b.RawData(), ncidlocal, l.keySet.Pk) { //如果本地有这个块的话，无需麻烦Provider
		} else {
			err = l.ds.DeleteBlock(ctx, ncidlocal, "local")
			if err != nil && err != bs.ErrNotFound {
				return err
			}
			log.Printf("Cannot Get ObjectInfo in block %s from local datastore. Maybe block is lost or broken.\n", ncidlocal)
			for j := 0; j < int(l.meta.sb.MetaBackupCount); j++ {
				bm.SetCid(strconv.Itoa(j))
				ncid := bm.ToString()
				provider, _, err := l.gInfo.getBlockProviders(ncid)
				if err != nil && j == int(l.meta.sb.MetaBackupCount)-1 {
					return ErrCannotLoadMetaBlock
				}
				km, _ := metainfo.NewKeyMeta(ncid, metainfo.Block)
				b, err = l.ds.GetBlock(ctx, km.ToString(), sig, provider)
				if b != nil && err == nil {
					if ok := dataformat.VerifyBlock(b.RawData(), ncid, l.keySet.Pk); !ok {
						log.Println("Verify Block failed.", ncid, "from:", provider)
					} else {
						break
					}
				} else if err != nil && j == int(l.meta.sb.MetaBackupCount)-1 {
					return ErrCannotLoadMetaBlock
				}
			}
		}
		if b != nil {
			data, err := dataformat.GetDataFromRawData(b.RawData())
			if err != nil {
				return err
			}
			if readCount+len(data) >= int(ObjectsBlockSize) { //读入数据等于object信息大小时，跳出循环
				end := int(ObjectsBlockSize) - readCount
				fullData = append(fullData, data[0:end]...)
				break
			}
			fullData = append(fullData, data...)
			readCount += len(data)
		} else {
			return ErrCannotLoadMetaBlock
		}
		stripeID++
	}
	objectsBuffer := bytes.NewBuffer(fullData)
	objectsDelimitedReader := ggio.NewDelimitedReader(objectsBuffer, 2*utils.BlockSize)
	for {
		object := pb.ObjectInfo{}
		err := objectsDelimitedReader.ReadMsg(&object)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if object.Size == 0 {
			continue
		}

		objectElement := bucket.orderedObjects.PushBack(&objectInfo{
			ObjectInfo: object,
		})
		bucket.objects[object.Name] = objectElement
	}
	return nil
}
