package user

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"io"
	"strconv"
	"sync"
	"time"

	ggio "github.com/gogo/protobuf/io"
	mcl "github.com/memoio/go-mefs/bls12"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	mt "gitlab.com/NebulousLabs/merkletree"
)

// LfsInfo has lfs info
type LfsInfo struct {
	userID     string
	fsID       string // use query addr as fsID
	privateKey string // of userID
	gInfo      *groupInfo
	ds         data.Service
	keySet     *mcl.KeySet
	meta       *lfsMeta //内存数据结构，存有当前的IpfsNode、SuperBlock和全部的Inode
	online     bool
	writable   bool // only one user can write
	context    context.Context
	cancelFunc context.CancelFunc
}

// Logs records lfs metainfo
type lfsMeta struct {
	sb             *superBlock
	bucketIDToName map[int64]string        //bucketID-> bucketName
	buckets        map[string]*superBucket //bucketName -> bucket
}

// superBlock has lfs bucket info
type superBlock struct {
	mpb.SuperBlockInfo
	sync.RWMutex
	dirty bool //看看superBlock是否需要更新（仅在新创建Bucket时需要）
}

// superBucket has lfs objects info
type superBucket struct {
	mpb.BucketInfo
	objects map[string]*objectInfo
	dirty   bool
	sync.RWMutex
	mtree *mt.Tree
}

// objectInfo stores an object meta info
type objectInfo struct {
	mpb.ObjectInfo
	sync.RWMutex
}

// Start starts user's info
func (l *LfsInfo) Start(ctx context.Context) error {
	// 证明该user已经启动
	if l.online || (l.gInfo != nil && l.gInfo.state > starting) {
		return nil
	}

	l.online = false
	l.writable = true

	has, err := l.gInfo.start(ctx)
	if err != nil {
		utils.MLogger.Error("Start group: ", l.fsID, " for: ", l.userID, " fail: ", err)
		return err
	}

	for _, kinof := range l.gInfo.keepers {
		if kinof.sessionID != l.gInfo.sessionID {
			utils.MLogger.Infof("%s starts in readonly mode, has session %s, want session: %s ", l.userID, l.gInfo.sessionID.String(), kinof.sessionID.String())
			l.writable = false
			break
		}
	}

	if has {
		// init or send bls config
		err = l.loadBLS12Config()
		if err != nil {
			utils.MLogger.Warn("Load bls config fail: ", err)
		}

		if l.keySet.Sk == nil {
			seed := sha256.Sum256([]byte(l.privateKey + l.fsID))
			mkey, err := initBLS12Config(seed[:])
			if err != nil {
				utils.MLogger.Info("Init bls config fail: ", err)
				return err
			}
			l.keySet = mkey
			l.putUserConfig(l.context)
		}
	}

	if !has || err != nil {
		seed := sha256.Sum256([]byte(l.privateKey + l.fsID))
		mkey, err := initBLS12Config(seed[:])
		if err != nil {
			utils.MLogger.Info("Init bls config fail: ", err)
			return err
		}

		l.keySet = mkey
		l.putUserConfig(l.context)
	}

	// in case persist is cancel
	err = l.startLfs(l.context)
	if err != nil {
		utils.MLogger.Error("Start lfs: ", l.fsID, " for: ", l.userID, " fail: ", err)
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
		utils.MLogger.Warn("Load superblock fail, so begin to init Lfs :", l.fsID)
		l.meta, err = initLfs() //初始化
		if err != nil {
			return err
		}
	} else {
		err = l.loadBucketInfo() //再加载Group元数据
		if err != nil {          //*错误处理
			utils.MLogger.Info("Load bucket info fail: ", err)
			return err
		}
		for _, bucket := range l.meta.buckets {
			err = l.loadObjectsInfo(bucket) //再加载Object元数据
			if err != nil {
				utils.MLogger.Error("Load objects in bucket", bucket.Name, " fail: ", err)
				continue
			}
			utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " is loaded as name: ", bucket.Name)
		}
	}
	utils.MLogger.Infof("Lfs Service %s is ready for: %s", l.fsID, l.userID)
	go l.persistMetaBlock(ctx)
	go l.persistRoot(ctx)
	go l.sendHeartBeat(ctx)
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
	return &lfsMeta{
		sb:             sb,
		bucketIDToName: make(map[int64]string),
		buckets:        make(map[string]*superBucket),
	}, nil
}

func newSuperBlock() *superBlock {
	return &superBlock{
		SuperBlockInfo: mpb.SuperBlockInfo{
			MetaBackupCount: defaultMetaBackupCount,
			NextBucketID:    1, //从1开始是因为SuperBlock的元数据块抢占了Bucket编号0的位置
			Version:         1001},
		dirty: true,
	}
}

// Stop user's info
func (l *LfsInfo) Stop() error {
	//用于通知资源释放
	l.gInfo.stop(l.context)
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

func (l *LfsInfo) sendHeartBeat(ctx context.Context) error {
	utils.MLogger.Infof("Send Heartbeat %s is ready for: %s", l.fsID, l.userID)
	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if l.online && l.writable {
				l.gInfo.heartbeat(ctx)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

//每隔一段时间，会检查元数据快是否为脏，决定要不要持久化
func (l *LfsInfo) persistRoot(ctx context.Context) error {
	utils.MLogger.Infof("Persist Lfs root %s is ready for: %s", l.fsID, l.userID)
	tick := time.NewTicker(30 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if l.online && l.writable {
				l.genRoot()
			}
		case <-ctx.Done():
			if l.online && l.writable {
				l.genRoot()
			}
			return nil
		}
	}
}

func (l *LfsInfo) genRoot() {
	l.meta.sb.RLock()
	bucketNum := l.meta.sb.GetNextBucketID() - 1
	if bucketNum == 0 {
		return
	}
	ctime := time.Now().Unix()

	lr := &mpb.LfsRoot{
		BRoots: make([]*mpb.BucketRoot, bucketNum),
		CTime:  ctime,
	}

	for _, bucket := range l.meta.buckets {
		bucket.RLock()
		i := int(bucket.BucketID - 1)
		if i >= int(bucketNum) {
			utils.MLogger.Errorf("bucketID is %d, but total is %d", bucket.BucketID, bucketNum)
		}

		lr.BRoots[i] = &mpb.BucketRoot{
			BucketID:    bucket.BucketID,
			Root:        bucket.Root,
			ObjectCount: bucket.NextObjectID,
		}
		bucket.RUnlock()
	}
	l.meta.sb.RUnlock()

	mtree := mt.New(sha256.New())
	mtree.SetIndex(0)

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(ctime))
	mtree.Push([]byte(l.fsID))
	mtree.Push(buf)

	for i := 0; i < int(bucketNum); i++ {
		if lr.BRoots[i] == nil {
			continue
		}
		bbuf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, uint64(lr.BRoots[i].ObjectCount))
		bbuf = append(bbuf, lr.BRoots[i].Root...)
		mtree.Push(bbuf)
	}

	lr.Root = mtree.Root()

	// add root to contract

	l.meta.sb.Lock()
	l.meta.sb.LRoot = append(l.meta.sb.LRoot, lr)
	l.meta.sb.dirty = true
	l.meta.sb.Unlock()

	if l.gInfo.userID != l.gInfo.rootID {
		var val [32]byte
		copy(val[:], lr.Root[:32])
		role.SetMerkleRoot(l.privateKey, l.gInfo.rootID, ctime, val)
		utils.MLogger.Debugf("set root %d for %s", ctime, l.gInfo.groupID)
		keyTime, res, err := role.GetLatestMerkleRoot(l.gInfo.rootID)
		if err != nil {
			return
		}
		if keyTime != ctime {
			utils.MLogger.Debugf("get root expected: %d, but got %d", ctime, keyTime)
		}

		if bytes.Compare(res[:], val[:]) != 0 {
			utils.MLogger.Debugf("get root expected: %s, but got %d", val, res)
		}
	}
}

//每隔一段时间，会检查元数据快是否为脏，决定要不要持久化
func (l *LfsInfo) persistMetaBlock(ctx context.Context) error {
	utils.MLogger.Infof("Persist Lfs %s is ready for: %s", l.fsID, l.userID)
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			if l.online && l.writable { //LFS没启动不刷新
				err := l.Fsync(false)
				if err != nil {
					utils.MLogger.Warn("Cannot Persist MetaBlock: ", err)
				}
			}
		case <-ctx.Done():
			if l.online && l.writable { //LFS没启动不刷新
				err := l.Fsync(true)
				if err != nil {
					utils.MLogger.Warn("Cannot Persist MetaBlock: ", err)
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

	if !l.writable {
		return nil
	}

	err := l.flushSuperBlock(isForce)
	if err != nil {
		return err
	}

	for _, bucket := range l.meta.buckets {
		err := l.flushBucketAndObjects(bucket, isForce)
		if err != nil {
			return err
		}
	}

	if isForce {
		l.gInfo.saveChannelValue()
	}

	return nil
}

//----------------------Flush superBlock---------------------------
func (l *LfsInfo) flushSuperBlock(isForce bool) error {
	l.meta.sb.RLock()
	defer l.meta.sb.RUnlock()

	if !isForce && !l.meta.sb.dirty {
		return nil
	}

	sb := l.meta.sb
	sbBuffer := bytes.NewBuffer(nil)
	sbDelimitedWriter := ggio.NewDelimitedWriter(sbBuffer)
	defer sbDelimitedWriter.Close()

	err := sbDelimitedWriter.WriteMsg(&sb.SuperBlockInfo)
	if err != nil {
		return err
	}

	data := sbBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}

	bm, err := metainfo.NewBlockMeta(l.fsID, "0", "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)

	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, int(defaultMetaBackupCount)-1, l.userID, l.fsID)
	dataEncoded, offset, err := enc.Encode(data, ncidPrefix, 0)
	if err != nil {
		return err
	}
	ncid := bm.ToString()
	km, err := metainfo.NewKey(ncid, mpb.KeyType_Block)
	if err != nil {
		return err
	}

	ctx := context.Background()

	l.ds.PutBlock(ctx, km.ToString(), dataEncoded[0], "local")

	providers, _, err := l.gInfo.GetProviders(int(sb.MetaBackupCount))
	if err != nil && len(providers) == 0 {
		return err
	}
	for j := 0; j < len(providers); j++ { //
		bm.SetCid(strconv.Itoa(j))
		ncid := bm.ToString()

		km, err := metainfo.NewKey(ncid, mpb.KeyType_Block)
		if err != nil {
			continue
		}
		updateKey := km.ToString()

		err = l.ds.PutBlock(ctx, updateKey, dataEncoded[j], providers[j])
		if err != nil {
			continue
		}

		err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
		if err != nil {
			continue
		}
	}

	utils.MLogger.Infof("user %s lfs %s superblock persist. ", l.userID, l.fsID)
	l.meta.sb.dirty = false
	return nil
}

func (l *LfsInfo) flushBucketAndObjects(bucket *superBucket, flag bool) error {
	bucket.RLock()
	defer bucket.RUnlock()

	if bucket.dirty || flag {
		err := l.flushObjectsInfo(bucket)
		if err != nil {
			return err
		}

		err = l.flushBucketInfo(bucket)
		if err != nil {
			return err
		}
		utils.MLogger.Infof("Flush user %s %s BucketInfo and its objects finish.", l.fsID, bucket.Name)
	}
	bucket.dirty = false
	return nil
}

//-----------------------Flush BucketMeta----------------------------
func (l *LfsInfo) flushBucketInfo(bucket *superBucket) error {
	bucketBuffer := bytes.NewBuffer(nil)
	bucketDelimitedWriter := ggio.NewDelimitedWriter(bucketBuffer)
	defer bucketDelimitedWriter.Close()
	err := bucketDelimitedWriter.WriteMsg(&bucket.BucketInfo)
	if err != nil {
		return err
	}

	if bucketBuffer.Len() == 0 {
		return nil
	}

	metaBackupCount := int(l.meta.sb.MetaBackupCount)
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), "0", "0")
	if err != nil {
		return err
	}

	ncidPrefix := bm.ToString(3)
	dataEncoded, offset, err := enc.Encode(bucketBuffer.Bytes(), ncidPrefix, 0)
	if err != nil {
		return err
	}

	ctx := context.Background()
	l.ds.PutBlock(ctx, bm.ToString(), dataEncoded[0], "local")

	providers, _, err := l.gInfo.GetProviders(metaBackupCount)
	if err != nil && len(providers) == 0 {
		return err
	}

	for j := 0; j < metaBackupCount && j < len(providers); j++ { //
		bm.SetCid(strconv.Itoa(j))
		ncid := bm.ToString()
		km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block)
		err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[j], providers[j])
		if err != nil {
			continue
		}

		err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
		if err != nil {
			continue
		}
	}
	return nil
}

//---------------------Flush objects' Meta for given superBucket--------
func (l *LfsInfo) flushObjectsInfo(bucket *superBucket) error {
	if bucket == nil || bucket.objects == nil {
		return nil
	}

	objectsBuffer := bytes.NewBuffer(nil)
	objectDelimitedWriter := ggio.NewDelimitedWriter(objectsBuffer)
	defer objectDelimitedWriter.Close()

	metaBackupCount := l.meta.sb.MetaBackupCount
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, int(metaBackupCount-1), l.userID, l.fsID)

	providers, _, err := l.gInfo.GetProviders(int(metaBackupCount))
	if err != nil && len(providers) == 0 {
		return err
	}

	bucketID := bucket.BucketID
	objectsStripeID := 1
	objectsBlockLength := 0
	ctx := context.Background()

	for _, object := range bucket.objects {
		err := objectDelimitedWriter.WriteMsg(&object.ObjectInfo)
		if err != nil {
			continue
		}
	}

	if objectsBuffer.Len() != 0 { //处理最后的剩余部分
		objectsBlockLength += objectsBuffer.Len()
		bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
		if err != nil {
			return err
		}

		ncidPrefix := bm.ToString(3)
		dataEncoded, offset, err := enc.Encode(objectsBuffer.Bytes(), ncidPrefix, 0)
		if err != nil {
			return err
		}

		l.ds.PutBlock(ctx, bm.ToString(), dataEncoded[0], "local")

		for j := 0; j < len(providers); j++ {
			bm.SetCid(strconv.Itoa(j))
			ncid := bm.ToString()
			km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block)
			err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[j], providers[j])
			if err != nil {
				continue
			}

			err = l.gInfo.putDataMetaToKeepers(ncid, providers[j], int(offset))
			if err != nil {
				continue
			}
		}

		l.meta.buckets[bucket.Name].ObjectsBlockSize = int64(objectsBlockLength)
	}

	return nil
}

//--------------------Load superBlock--------------------------
//lfs启动时加载超级块操作，返回结构体Meta,主要填充其中的superblock字段
//先从本地查找超级快信息，若没找到，就找自己的provider获取
func (l *LfsInfo) loadSuperBlock() (*lfsMeta, error) {
	utils.MLogger.Info("Load superblock: ", l.fsID, " for user:", l.userID)
	if l.keySet == nil {
		return nil, role.ErrEmptyBlsKey
	}
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, int(defaultMetaBackupCount-1), l.userID, l.fsID)

	var data []byte

	bm, err := metainfo.NewBlockMeta(l.fsID, "0", "0", "0")
	if err != nil {
		return nil, err
	}
	ncidlocal := bm.ToString()
	km, _ := metainfo.NewKey(ncidlocal, mpb.KeyType_Block)
	ctx := context.Background()
	b, err := l.ds.GetBlock(ctx, km.ToString(), nil, "local")
	if err == nil && b != nil {
		_, _, ok := enc.VerifyBlock(b.RawData(), ncidlocal)
		if ok {
			data = append(data, b.RawData()...)
		}
	}

	sig, err := role.BuildSignMessage()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		utils.MLogger.Info("Try to get: ", ncidlocal, " from remote servers")
		for j := 0; j < int(defaultMetaBackupCount); j++ {
			bm.SetCid(strconv.Itoa(j))
			ncid := bm.ToString()
			provider, _, err := l.gInfo.getBlockProviders(ncid) //获取数据块的保存位置
			if err != nil || provider == "" {
				continue
			}

			km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block)

			b, err := l.ds.GetBlock(ctx, km.ToString(), sig, provider)
			if err == nil && b != nil { //获取到有效数据块，跳出
				_, _, ok := enc.VerifyBlock(b.RawData(), ncid)
				if ok {
					data = append(data, b.RawData()...)
					utils.MLogger.Warn("Load superblock in block: ", ncid, " from provider: ", provider)
					break
				}
			}
		}
	}

	if len(data) > 0 {
		res := make([][]byte, 1)
		res[0] = data
		data, err := enc.Decode(res, 0, -1)
		if err != nil {
			utils.MLogger.Info("Decode data fail: ", err)
			return nil, err
		}
		pbSuperBlock := mpb.SuperBlockInfo{}
		SbBuffer := bytes.NewBuffer(data)
		SbDelimitedReader := ggio.NewDelimitedReader(SbBuffer, 5*dataformat.BlockSize)
		err = SbDelimitedReader.ReadMsg(&pbSuperBlock)
		if err == io.EOF {
		} else if err != nil {
			return nil, err
		}

		return &lfsMeta{
			sb: &superBlock{
				SuperBlockInfo: pbSuperBlock,
				dirty:          false,
			},
			buckets:        make(map[string]*superBucket),
			bucketIDToName: make(map[int64]string),
		}, nil
	}
	utils.MLogger.Warn("Cannot load Lfs superblock.")
	return nil, ErrCannotLoadSuperBlock
}

//lfs启动进行元数据的加载，对Log中的字段进行初始化 填充除superblock、Entries字段之外的字段
func (l *LfsInfo) loadBucketInfo() error {
	sig, err := role.BuildSignMessage()
	if err != nil {
		return err
	}

	metaBackupCount := int(l.meta.sb.MetaBackupCount)

	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)
	ctx := context.Background()
	for bucketID := int64(1); bucketID < l.meta.sb.NextBucketID; bucketID++ {
		var data []byte
		bm, _ := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), "0", "0")
		ncidlocal := bm.ToString()
		b, err := l.ds.GetBlock(ctx, ncidlocal, nil, "local")
		if err == nil && b != nil {
			_, _, ok := enc.VerifyBlock(b.RawData(), ncidlocal)
			if ok {
				data = append(data, b.RawData()...)
			}
		}

		if len(data) == 0 {
			for j := 0; j < int(l.meta.sb.MetaBackupCount); j++ {
				bm.SetCid(strconv.Itoa(j))
				ncid := bm.ToString()
				provider, _, err := l.gInfo.getBlockProviders(ncid)
				if err != nil || provider == "" {
					continue
				}
				b, err = l.ds.GetBlock(ctx, ncid, sig, provider)
				if err == nil && b != nil {
					_, _, ok := enc.VerifyBlock(b.RawData(), ncid)
					if ok {
						data = append(data, b.RawData()...)
						break
					}
				}
			}
		}

		if len(data) > 0 {
			res := make([][]byte, 1)
			res[0] = data
			data, err := enc.Decode(res, 0, -1) //Tag暂时没用
			if err != nil {
				utils.MLogger.Info("Decode data fail: ", err)
				continue
			}
			bucket := mpb.BucketInfo{}
			BucketBuffer := bytes.NewBuffer(data)
			BucketDelimitedReader := ggio.NewDelimitedReader(BucketBuffer, 5*dataformat.BlockSize)
			err = BucketDelimitedReader.ReadMsg(&bucket)
			if err != nil && err != io.EOF {
				continue
			}
			objects := make(map[string]*objectInfo)
			tsb := &superBucket{
				BucketInfo: bucket,
				objects:    objects,
				dirty:      false,
				mtree:      mt.New(sha256.New()),
			}

			tsb.mtree.SetIndex(0)
			tsb.mtree.Push([]byte(l.fsID + bucket.Name))

			bname := bucket.Name
			if bucket.Deletion {
				bname = bucket.Name + "." + strconv.Itoa(int(bucket.BucketID))
			}
			l.meta.buckets[bname] = tsb
			l.meta.bucketIDToName[bucketID] = bname
		}
	}
	return nil
}

//------------------------------Load Objectinfo---------------------------------------
//填充Entries字段，传入参数为bucket,记录传入bucket的数据信息
func (l *LfsInfo) loadObjectsInfo(bucket *superBucket) error {
	sig, err := role.BuildSignMessage()
	if err != nil {
		return err
	}
	objectsBlockSize := bucket.ObjectsBlockSize
	if objectsBlockSize == 0 {
		return nil
	}

	fullData := make([]byte, 0, objectsBlockSize)

	metaBackupCount := int(l.meta.sb.MetaBackupCount)
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)

	stripeID := 1 //ObjectsBlock的Stripe从1开始计算
	ctx := context.Background()

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), strconv.Itoa(stripeID), "0")
	if err != nil {
		return err
	}
	ncidlocal := bm.ToString()

	var data []byte
	b, err := l.ds.GetBlock(ctx, ncidlocal, nil, "local")
	if b != nil && err == nil {
		_, _, ok := enc.VerifyBlock(b.RawData(), ncidlocal)
		if ok {
			data = append(data, b.RawData()...)
		}
	}

	if len(data) == 0 {
		for j := 0; j < int(l.meta.sb.MetaBackupCount); j++ {
			bm.SetCid(strconv.Itoa(j))
			ncid := bm.ToString()
			provider, _, err := l.gInfo.getBlockProviders(ncid)
			if err != nil || provider == "" {
				continue
			}
			km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block)
			b, err := l.ds.GetBlock(ctx, km.ToString(), sig, provider)
			if b != nil && err == nil {
				_, _, ok := enc.VerifyBlock(b.RawData(), ncid)
				if ok {
					data = append(data, b.RawData()...)
					break
				}
			}
		}
	}

	if len(data) > 0 {
		res := make([][]byte, 1)
		res[0] = data
		data, err := enc.Decode(res, 0, -1)
		if err != nil {
			return err
		}

		if len(data) < int(objectsBlockSize) {
			utils.MLogger.Warn("data length is not equal")
		}

		fullData = append(fullData, data...)

		objectSlice := make([]*mpb.ObjectInfo, bucket.NextObjectID)

		objectsBuffer := bytes.NewBuffer(fullData)
		objectsDelimitedReader := ggio.NewDelimitedReader(objectsBuffer, 2*dataformat.BlockSize)
		for {
			object := mpb.ObjectInfo{}
			err := objectsDelimitedReader.ReadMsg(&object)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}

			if object.GetOPart() == nil {
				continue
			}

			if object.GetOPart().GetLength() == 0 {
				continue
			}

			oName := object.GetOPart().Name
			if object.Deletion {
				oName = object.GetOPart().Name + "." + strconv.Itoa(int(object.ObjectID))
			}

			objectSlice[int(object.ObjectID)] = &object

			bucket.objects[oName] = &objectInfo{
				ObjectInfo: object,
			}
		}

		for i := 0; i < int(bucket.NextObjectID); i++ {
			if objectSlice[i] != nil {
				continue
			}
			bucket.mtree.Push([]byte(objectSlice[i].GetOPart().GetName() + objectSlice[i].GetOPart().GetETag()))
		}
	}
	return nil
}
