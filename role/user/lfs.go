package user

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"time"

	mcl "github.com/memoio/go-mefs/bls12"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/source/data"
	"github.com/memoio/go-mefs/utils"
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

// Start starts user's info
func (l *LfsInfo) Start(ctx context.Context) error {
	if l.gInfo == nil {
		return ErrLfsServiceNotReady
	}
	// 证明该user已经启动
	if l.online || (l.gInfo.state == groupStarted) {
		return nil
	}
	if l.gInfo.state >= starting && l.gInfo.state < groupStarted {
		return ErrLfsStarting
	}
	l.gInfo.state = starting
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

		if l.gInfo.userID == l.gInfo.shareToID {
			if l.keySet == nil || l.keySet.Sk == nil {
				seed := sha256.Sum256([]byte(l.privateKey + l.fsID))
				mkey, err := initBLS12Config(seed[:])
				if err != nil {
					utils.MLogger.Info("Init bls config fail: ", err)
					return err
				}
				l.keySet = mkey
				l.putUserConfig(l.context)
			}
		} else {
			if l.keySet == nil {
				return role.ErrEmptyBlsKey
			}
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
	_, err = checkMetaPath(l.fsID)
	if err != nil {
		return err
	}
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
			bucket.Lock()
			err = l.loadObjectsInfo(bucket) //再加载Object元数据
			bucket.Unlock()
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

		utils.MLogger.Infof("ser root: %s success", val)
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
