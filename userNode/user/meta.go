package user

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"sync"

	ggio "github.com/gogo/protobuf/io"
	"github.com/gogo/protobuf/proto"
	config "github.com/memoio/go-mefs/config"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	rbtree "github.com/memoio/go-mefs/utils/RbTree"
	"github.com/memoio/go-mefs/utils/metainfo"
	mt "gitlab.com/NebulousLabs/merkletree"
)

const meta = "meta"
const maxCacheSize = 4 * 1024

// Logs records lfs metainfo
type lfsMeta struct {
	sb             *superBlock
	bucketIDToName map[int64]string        //bucketID-> bucketName
	buckets        map[string]*superBucket //bucketName -> bucket
	deletedBuckets []*superBucket
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
	Objects       *rbtree.Tree
	DeletedObject []*ObjectInfo
	obMetaCache   []byte
	obCacheSize   int //obMetaCache 已经用了多少
	dirty         bool
	sync.RWMutex
	mtree *mt.Tree
}

// objectInfo stores an object meta info
type ObjectInfo struct {
	mpb.ObjectInfo
	sync.RWMutex
}

type MetaName string

func (x MetaName) LessThan(y interface{}) bool {
	yStr := y.(MetaName)
	return x < yStr
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

	writeToMeta(data, l.fsID, "0")

	err = l.putDataToBlocks(data, int(defaultMetaBackupCount), "0", "0")
	if err != nil {
		return err
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

	data := bucketBuffer.Bytes()

	writeToMeta(data, l.fsID, strconv.Itoa(int(bucket.BucketID)))

	return l.putDataToBlocks(data, int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "0")
}

//---------------------Flush objects' Meta for given superBucket--------
func (l *LfsInfo) flushObjectsInfo(bucket *superBucket) error {
	if bucket == nil || bucket.Objects == nil {
		return nil
	}

	//先把cache都刷盘
	l.flushObjectMeta(bucket, true)

	data, err := readFromMeta(l.fsID, strconv.FormatInt(bucket.BucketID, 10)+".object")
	if err != nil {
		return err
	}

	bucket.ObjectsBlockSize = int64(len(data))

	return l.putDataToBlocks(data, int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "1")
}

//--------------------Load superBlock--------------------------
//lfs启动时加载超级块操作，返回结构体Meta,主要填充其中的superblock字段
//先从本地查找超级快信息，若没找到，就找自己的provider获取
func (l *LfsInfo) loadSuperBlock() (*lfsMeta, error) {
	utils.MLogger.Info("Load superblock: ", l.fsID, " for user:", l.userID)

	data, err := readFromMeta(l.fsID, "0")
	if err != nil || len(data) == 0 {
		datagot, err := l.getDataFromBlock(int(defaultMetaBackupCount), "0", "0")
		if err != nil {
			return nil, err
		}
		if len(datagot) > len(data) {
			data = datagot
			writeToMeta(data, l.fsID, "0")
		}
	}

	if len(data) > 0 {
		pbSuperBlock := mpb.SuperBlockInfo{}
		SbBuffer := bytes.NewBuffer(data)
		SbDelimitedReader := ggio.NewDelimitedReader(SbBuffer, len(data))
		err = SbDelimitedReader.ReadMsg(&pbSuperBlock)
		if err != nil {
			utils.MLogger.Info("Protobuf ReadMsg fail: ", err)
			return nil, err
		}

		lm := &lfsMeta{
			sb: &superBlock{
				SuperBlockInfo: pbSuperBlock,
				dirty:          false,
			},
			buckets:        make(map[string]*superBucket),
			bucketIDToName: make(map[int64]string),
		}

		if l.userID != l.gInfo.rootID {
			gotTime, gotRoot, err := role.GetLatestMerkleRoot(l.gInfo.rootID)
			if err == nil {
				has := false
				for _, lr := range lm.sb.LRoot {
					if lr.CTime == gotTime && bytes.Compare(lr.Root, gotRoot[:]) == 0 {
						has = true
					}
				}
				if has {
					utils.MLogger.Info("local fs has contract merkle root")
				} else {
					utils.MLogger.Info("local fs has not contract merkle root")
				}
			}
		}

		return lm, nil
	}
	utils.MLogger.Warn("Cannot load Lfs superblock.")
	return nil, ErrCannotLoadSuperBlock
}

//lfs启动进行元数据的加载，对Log中的字段进行初始化 填充除superblock、Entries字段之外的字段
func (l *LfsInfo) loadBucketInfo() error {
	var err error
	for bucketID := int64(1); bucketID < l.meta.sb.NextBucketID; bucketID++ {
		err = l.loadSingleBucketInfo(bucketID)
		if err != nil {
			utils.MLogger.Error("Load BucketInfo failed , bucketID: ", bucketID)
		}
	}
	return nil
}

func (l *LfsInfo) loadSingleBucketInfo(bucketID int64) error {
	data, err := readFromMeta(l.fsID, strconv.FormatInt(bucketID, 10))
	if err != nil || len(data) == 0 {
		datagot, err := l.getDataFromBlock(int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucketID)), "0")
		if err != nil {
			return err
		}

		if len(datagot) > len(data) {
			data = datagot
			writeToMeta(data, l.fsID, strconv.FormatInt(bucketID, 10))
		}
	}

	if len(data) > 0 {
		bucket := mpb.BucketInfo{}
		BucketBuffer := bytes.NewBuffer(data)
		BucketDelimitedReader := ggio.NewDelimitedReader(BucketBuffer, len(data))
		err = BucketDelimitedReader.ReadMsg(&bucket)
		if err != nil {
			utils.MLogger.Info("Protobuf ReadMsg fail: ", err)
			return err
		}

		tsb := newsuperBucket(bucket, false)

		tsb.mtree.SetIndex(0)
		tsb.mtree.Push([]byte(l.fsID + bucket.Name))

		bname := bucket.Name
		if bucket.Deletion {
			l.meta.deletedBuckets = append(l.meta.deletedBuckets, tsb)
		} else {
			l.meta.buckets[bname] = tsb
			l.meta.bucketIDToName[bucketID] = bname
		}
		return nil
	}
	return ErrCannotLoadMetaBlock
}

//-------------------------Load Objectinfo----------------------------
//填充Entries字段，传入参数为bucket,记录传入bucket的数据信息
func (l *LfsInfo) loadObjectsInfo(bucket *superBucket) error {
	//先从metapath找，是否需要一个验证版本的方法？
	objectsBlockSize := bucket.ObjectsBlockSize

	data, err := readFromMeta(l.fsID, strconv.FormatInt(bucket.BucketID, 10)+".object")
	if err != nil || int64(len(data)) < objectsBlockSize {
		datagot, err := l.getDataFromBlock(int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "1")
		if err != nil {
			return err
		}

		if len(datagot) > len(data) {
			data = datagot
			writeToMeta(data, l.fsID, strconv.FormatInt(bucket.BucketID, 10)+".object")
		}
	}

	broot := new(mpb.BucketRoot)
	if len(l.meta.sb.GetLRoot()) > 0 {
		lroot := l.meta.sb.GetLRoot()[len(l.meta.sb.GetLRoot())-1]
		if len(lroot.GetBRoots()) < int(bucket.BucketID) {
			utils.MLogger.Error("Objects in bucket: ", bucket.BucketID, " has inconsistent objects")
		} else {
			broot = lroot.GetBRoots()[bucket.BucketID-1]
		}
	}

	if int64(len(data)) >= objectsBlockSize {
		data = data[:objectsBlockSize]
		utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " has objects: ", bucket.NextObjectID)
		objectsBuffer := bytes.NewBuffer(data)
		objectsDelimitedReader := ggio.NewDelimitedReader(objectsBuffer, len(data))
		op := mpb.OpRecord{}
		var opNum int64
		for {
			err := objectsDelimitedReader.ReadMsg(&op)
			if err != nil {
				break
			}
			err = applyOp(bucket, &op)
			if err != nil {
				continue
			}
			opNum++
			tag := append([]byte(strconv.FormatInt(op.GetOpID(), 10)), op.GetPayload()...)
			bucket.mtree.Push(tag)
			if opNum == broot.GetOpCount() {
				if bytes.Compare(broot.GetRoot(), bucket.mtree.Root()) != 0 {
					utils.MLogger.Errorf("bucket %s expect root %s, but got %s", bucket.Name, hex.EncodeToString(broot.GetRoot()), hex.EncodeToString(bucket.mtree.Root()))
				}
			}
		}

		bucket.Root = bucket.mtree.Root()
		// verify root
		// verify ops
		if opNum != bucket.GetNextOpID() {
			utils.MLogger.Infof("Load ops is not correct, expect: %d, but got %d", bucket.GetNextOpID(), opNum)
		}

		return nil
	}
	return ErrCannotLoadMetaBlock
}

//applyOp 应用某个操作
func applyOp(bucket *superBucket, op *mpb.OpRecord) error {
	var err error
	payload := op.GetPayload()
	switch op.OpType {
	case mpb.LfsOp_OpAdd:
		info := mpb.Object{}
		err = proto.Unmarshal(payload, &info)
		if err != nil {
			utils.MLogger.Error("OpAdd payload parse failed, bucket: ", bucket.GetName())
			return err
		}
		if ob := bucket.Objects.Find(MetaName(info.GetName())); ob != nil {
			return ErrObjectAlreadyExist
		}
		bucket.Objects.Insert(MetaName(info.GetName()), &ObjectInfo{
			ObjectInfo: mpb.ObjectInfo{
				Info:      &info,
				Deletion:  false,
				Length:    0,
				CTime:     info.GetCTime(),
				MTime:     info.GetCTime(),
				PartCount: 0,
				Parts:     make([]*mpb.ObjectPart, 0, 1),
			},
		})
		utils.MLogger.Info("Add Object-", info.GetName(), " in bucket: ", bucket.Name)
	case mpb.LfsOp_OpAppend:
		part := mpb.ObjectPart{}
		err = proto.Unmarshal(payload, &part)
		if err != nil {
			utils.MLogger.Error("OpAppend payload parse failed, bucket: ", bucket.GetName())
			return err
		}
		ob, ok := bucket.Objects.Find(MetaName(part.GetName())).(*ObjectInfo)
		if !ok || ob == nil {
			utils.MLogger.Error("Add Part-", part.GetPartID(), "to an inexistent object-", part.GetName())
			return ErrObjectNotExist
		}

		ob.Lock()
		if part.PartID < ob.PartCount {
			ob.Unlock()
			return ErrObjectAlreadyExist
		}
		ob.Parts = append(ob.Parts, &part)
		ob.PartCount++
		ob.ETag = calulateETag(ob)
		ob.Length += part.Length
		ob.MTime = part.GetCTime()
		ob.Unlock()
	case mpb.LfsOp_OpDelete:
		mes := mpb.DeleteObject{}
		err = proto.Unmarshal(payload, &mes)
		if err != nil {
			utils.MLogger.Error("OpDelete payload parse failed, bucket: ", bucket.GetName())
			return err
		}
		ob, ok := bucket.Objects.Find(MetaName(mes.GetName())).(*ObjectInfo)
		if !ok || ob == nil {
			utils.MLogger.Error("Delete an inexistent object-", mes.GetName())
			return err
		}
		ob.Lock()
		ob.Deletion = true
		bucket.Objects.Delete(MetaName(mes.GetName()))
		bucket.DeletedObject = append(bucket.DeletedObject, ob)
		ob.Unlock()
	case mpb.LfsOp_OpCancel:
		return errors.New("Undefined")
		//这里暂时不实现
	default:
		return errors.New("Undefined")
	}

	return nil
}

func (l *LfsInfo) flushObjectMeta(bucket *superBucket, force bool, ops ...*mpb.OpRecord) error {
	//先检查本地有没有
	metapath, err := checkMetaPath(l.fsID)
	if err != nil {
		return err
	}
	obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
	stat, err := os.Stat(obpath)
	if os.IsNotExist(err) {
		// 本地这个文件被删了，要先读取回来
		data, err := l.getDataFromBlock(int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "1")
		if err != nil {
			return err
		}
		writeToMeta(data, l.fsID, strconv.FormatInt(bucket.BucketID, 10)+".object")
	} else if stat.IsDir() {
		err = os.Rename(metapath, metapath+".bak")
		if err != nil {
			err = os.Remove(metapath)
			if err != nil {
				return err
			}
		}
		// 本地这个文件被删了，要先读取回来
		data, err := l.getDataFromBlock(int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "1")
		if err != nil {
			return err
		}
		writeToMeta(data, l.fsID, strconv.FormatInt(bucket.BucketID, 10)+".object")
	}

	if len(ops) == 0 && bucket.obCacheSize == 0 {
		return nil
	}

	lenBuf := make([]byte, binary.MaxVarintLen64)

	for i := 0; i < len(ops); i++ {
		data, err := proto.Marshal(ops[i])
		if err != nil {
			continue
		}
		length := uint64(len(data))
		n := binary.PutUvarint(lenBuf, length)

		if bucket.obCacheSize+len(data)+n > maxCacheSize {
			obMetaFile, err := os.OpenFile(obpath, os.O_APPEND|os.O_WRONLY, 0666)
			if err != nil {
				obMetaFile.Close()
				return err
			}
			obMetaFile.Write(bucket.obMetaCache[:bucket.obCacheSize])
			obMetaFile.Write(lenBuf[:n])
			obMetaFile.Write(data)
			obMetaFile.Sync()
			obMetaFile.Close()
			bucket.obCacheSize = 0
			continue
		}

		copy(bucket.obMetaCache[bucket.obCacheSize:], lenBuf[:n])
		bucket.obCacheSize += n
		copy(bucket.obMetaCache[bucket.obCacheSize:], data)
		bucket.obCacheSize += len(data)
	}

	if force && bucket.obCacheSize > 0 {
		//采用追加模式
		obMetaFile, err := os.OpenFile(obpath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			obMetaFile.Close()
			return err
		}
		obMetaFile.Write(bucket.obMetaCache[:bucket.obCacheSize])
		obMetaFile.Sync()
		obMetaFile.Close()
		bucket.obCacheSize = 0
	}
	return nil
}

func (l *LfsInfo) putDataToBlocks(data []byte, metaBackupCount int, buc, stripe string) error {
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)

	bm, err := metainfo.NewBlockMeta(l.fsID, buc, stripe, "0")
	if err != nil {
		return err
	}

	ncidPrefix := bm.ToString(3)
	dataEncoded, offset, err := enc.Encode(data, ncidPrefix, 0)
	if err != nil {
		return err
	}

	ctx := l.context
	err = l.ds.PutBlock(ctx, bm.ToString(), dataEncoded[0], "local")
	if err != nil {
		utils.MLogger.Errorf("user %s lfs %s bucket %s info persist to local failed. ", l.userID, l.fsID, buc)
		return err
	}
	providers, _, err := l.gInfo.GetProviders(ctx, metaBackupCount)
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

		err = l.gInfo.putDataMetaToKeepers(ctx, ncid, providers[j], int(offset))
		if err != nil {
			continue
		}
	}
	return nil
}

func (l *LfsInfo) getDataFromBlock(metaBackupCount int, buc, stripe string) ([]byte, error) {
	if l.keySet == nil {
		return nil, role.ErrEmptyBlsKey
	}

	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)

	var data []byte

	bm, err := metainfo.NewBlockMeta(l.fsID, buc, stripe, "0")
	if err != nil {
		return nil, err
	}
	ncidlocal := bm.ToString()
	km, _ := metainfo.NewKey(ncidlocal, mpb.KeyType_Block)
	ctx := l.context
	b, err := l.ds.GetBlock(ctx, km.ToString(), nil, "local")
	if err == nil && b != nil {
		_, _, ok := enc.VerifyBlock(b.RawData(), ncidlocal)
		if ok {
			data = append(data, b.RawData()...)
		}
	}
	if len(data) == 0 {
		sig, err := role.BuildSignMessage()
		if err != nil {
			return nil, err
		}
		utils.MLogger.Info("Try to get: ", ncidlocal, " from remote servers")

		for j := 0; j < metaBackupCount; j++ {
			bm.SetCid(strconv.Itoa(j))
			ncid := bm.ToString()
			provider, _, err := l.gInfo.getBlockProviders(ctx, ncid) //获取数据块的保存位置
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
		data, err = enc.Decode(res, 0, -1)
		if err != nil {
			utils.MLogger.Info("Decode data fail: ", err)
			return nil, err
		}
	}
	return data, nil
}

func writeToMeta(data []byte, fsID, buc string) error {
	metapath, err := checkMetaPath(fsID)
	if err != nil {
		return err
	}

	sbpath := path.Join(metapath, buc)
	sbMetaFile, err := os.Create(sbpath)
	defer sbMetaFile.Close()
	if err != nil {
		return err
	}
	sbMetaFile.Write(data)
	sbMetaFile.Sync()
	return nil
}

func readFromMeta(fsID, buc string) ([]byte, error) {
	metapath, err := checkMetaPath(fsID)
	if err != nil {
		return nil, err
	}

	sbpath := path.Join(metapath, buc)
	sbMetaFile, err := os.Open(sbpath)
	defer sbMetaFile.Close()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(sbMetaFile)
}

func getMetaPath(fsID string) string {
	rootpath, _ := fsrepo.BestKnownPath()
	metapath, _ := config.Path(rootpath, meta)
	fsmetapath := path.Join(metapath, fsID)
	return fsmetapath
}

//检查path是否存在，不存在则新建一个
func checkMetaPath(fsID string) (string, error) {
	metapath := getMetaPath(fsID)
	stat, err := os.Stat(metapath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(metapath, 0777)
		return metapath, err
	}
	if !stat.IsDir() {
		err := os.Rename(metapath, metapath+".bak")
		if err != nil {
			os.Remove(metapath)
		}
		err = os.MkdirAll(metapath, 0777)
		return metapath, err
	}
	return metapath, nil
}
