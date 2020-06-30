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
	mpb "github.com/memoio/go-mefs/pb"
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
	dirty          bool
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
	obCacheSize   int //obMetaCintache 已经用了多少
	applyOpID     int64
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

	sbBuffer := bytes.NewBuffer(nil)
	sbDelimitedWriter := ggio.NewDelimitedWriter(sbBuffer)
	defer sbDelimitedWriter.Close()

	err := sbDelimitedWriter.WriteMsg(&l.meta.sb.SuperBlockInfo)
	if err != nil {
		return err
	}

	data := sbBuffer.Bytes()
	if len(data) == 0 {
		return nil
	}

	writeToMeta(data, l.fsID, "0")

	l.putDataToBlocks(data, int(defaultMetaBackupCount), "0", "0")

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

	l.putDataToBlocks(data, int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "0")

	return nil
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

	if len(data) == 0 {
		return nil
	}

	l.putDataToBlocks(data, int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "1")

	return nil
}

//--------------------Load superBlock--------------------------
//lfs启动时加载超级块操作，返回结构体Meta,主要填充其中的superblock字段
//先从本地查找超级快信息，若没找到，就找自己的provider获取
func (l *LfsInfo) loadSuperBlock(update bool) error {
	utils.MLogger.Info("Load superblock: ", l.fsID, " for user:", l.userID)

	localSb := mpb.SuperBlockInfo{}
	lm := &lfsMeta{
		sb: &superBlock{
			dirty: false,
		},
		buckets:        make(map[string]*superBucket),
		bucketIDToName: make(map[int64]string),
	}

	localHas := false
	if !update {
		l.meta = lm
		// read from local at start
		ldata, _ := readFromMeta(l.fsID, "0")
		if len(ldata) > 0 {
			localHas = true
			sdReader := ggio.NewDelimitedReader(bytes.NewBuffer(ldata), len(ldata))
			err := sdReader.ReadMsg(&localSb)
			if err != nil {
				utils.MLogger.Info("Protobuf ReadMsg local fail: ", err)
			}
			l.meta.sb.SuperBlockInfo = localSb
		}
	}

	remoteSb := mpb.SuperBlockInfo{}
	rdatas, err := l.getDataFromBlock(int(defaultMetaBackupCount), "0", "0")
	if err == nil {
		for _, rdata := range rdatas {
			if len(rdata) == 0 {
				continue
			}
			sdReader := ggio.NewDelimitedReader(bytes.NewBuffer(rdata), len(rdata))
			err := sdReader.ReadMsg(&remoteSb)
			if err != nil {
				utils.MLogger.Info("Protobuf ReadMsg remote fail: ", err)
				continue
			}
			if l.meta.sb.GetNextBucketID() < remoteSb.GetNextBucketID() || (l.meta.sb.GetNextBucketID() == remoteSb.GetNextBucketID() && len(l.meta.sb.GetLRoot()) < len(remoteSb.GetLRoot())) {
				// remote has newer
				utils.MLogger.Info("remote superblock has newer buckets: ", remoteSb.GetNextBucketID()-1)
				l.meta.sb.SuperBlockInfo = remoteSb
				l.meta.sb.dirty = true
			}
		}

	}

	// verify at start
	if !update && l.meta.sb.GetNextBucketID() > 0 {
		utils.MLogger.Infof("%s has %d buckets", l.fsID, l.meta.sb.GetNextBucketID()-1)
		if l.userID != l.gInfo.rootID {
			gotTime, gotRoot, err := role.GetLatestMerkleRoot(l.gInfo.rootID)
			if err == nil {
				has := false
				for _, lr := range lm.sb.GetLRoot() {
					if lr.CTime == gotTime && bytes.Compare(lr.Root, gotRoot[:]) == 0 {
						has = true
					}
				}
				if has {
					utils.MLogger.Infof("local fs contains contract merkle root: %s at %d", hex.EncodeToString(gotRoot[:]), gotTime)
				} else {
					utils.MLogger.Info("local fs has not contract merkle root")
				}
			}
		}
	}

	if !localHas && len(rdatas) == 0 {
		utils.MLogger.Warn("Cannot load Lfs superblock.")
		return ErrCannotLoadSuperBlock
	}
	return nil
}

//lfs启动进行元数据的加载，对Log中的字段进行初始化 填充除superblock、Entries字段之外的字段
func (l *LfsInfo) loadBucketInfo(update bool) error {
	for bucketID := int64(1); bucketID < l.meta.sb.GetNextBucketID(); bucketID++ {
		err := l.loadSingleBucketInfo(bucketID, update)
		if err != nil {
			utils.MLogger.Errorf("Load BucketInfo %d failed: %s", bucketID, err)
		}
	}
	return nil
}

func (l *LfsInfo) loadSingleBucketInfo(bucketID int64, update bool) error {
	idName := strconv.FormatInt(bucketID, 10)
	localbucket := mpb.BucketInfo{
		Name:     idName,
		BucketID: bucketID,
		Deletion: true,
	}

	tsb := &superBucket{}
	if !update {
		ldata, _ := readFromMeta(l.fsID, idName)
		if len(ldata) > 0 {
			bdReader := ggio.NewDelimitedReader(bytes.NewBuffer(ldata), len(ldata))
			err := bdReader.ReadMsg(&localbucket)
			if err != nil {
				utils.MLogger.Info("Protobuf ReadMsg local fail: ", err)
			}
		}
		tsb = newsuperBucket(localbucket, false)
		tsb.mtree.SetIndex(0)
		tsb.mtree.Push([]byte(l.fsID + idName))
		tsb.Root = tsb.mtree.Root()
	} else {
		bname, ok := l.meta.bucketIDToName[bucketID]
		if ok {
			sb, ok := l.meta.buckets[bname]
			if ok {
				tsb = sb
			}
		}
	}

	remotebucket := mpb.BucketInfo{
		NextObjectID: -1,
	}

	rdatas, err := l.getDataFromBlock(int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucketID)), "0")
	if err == nil && len(rdatas) > 0 {
		for _, rdata := range rdatas {
			if len(rdata) == 0 {
				continue
			}

			bdReader := ggio.NewDelimitedReader(bytes.NewBuffer(rdata), len(rdata))
			err := bdReader.ReadMsg(&remotebucket)
			if err != nil {
				utils.MLogger.Info("Protobuf ReadMsg remote fail: ", err)
				continue
			}

			if tsb.GetNextOpID() < remotebucket.GetNextOpID() || (tsb.GetName() != remotebucket.GetName() && remotebucket.GetName() != idName) {
				tsb = newsuperBucket(remotebucket, false)
				tsb.mtree.SetIndex(0)
				tsb.mtree.Push([]byte(l.fsID + strconv.FormatInt(bucketID, 10)))
				tsb.Root = tsb.mtree.Root()
				tsb.dirty = true
				utils.MLogger.Infof("remote has newer ops %d for bucket %d", remotebucket.GetNextOpID()-1, tsb.GetBucketID())
			}
		}

	}

	if tsb.GetDeletion() {
		l.meta.deletedBuckets = append(l.meta.deletedBuckets, tsb)
	} else {
		l.meta.buckets[tsb.GetName()] = tsb
		l.meta.bucketIDToName[bucketID] = tsb.GetName()
	}

	if tsb.GetName() == idName && tsb.GetDeletion() {
		utils.MLogger.Info("Construct delete buckets: ", bucketID)
	}

	return nil
}

//-------------------------Load Objectinfo----------------------------
//填充Entries字段，传入参数为bucket,记录传入bucket的数据信息
func (l *LfsInfo) loadObjectsInfo(bucket *superBucket, update bool) error {
	obSize := bucket.ObjectsBlockSize

	broot := new(mpb.BucketRoot)
	if len(l.meta.sb.GetLRoot()) > 0 {
		lroot := l.meta.sb.GetLRoot()[len(l.meta.sb.GetLRoot())-1]
		if len(lroot.GetBRoots()) < int(bucket.BucketID) || bucket.BucketID == 0 {
			utils.MLogger.Error("bucket: ", bucket.BucketID, " may has inconsistent objects or is newly created")
		} else {
			broot = lroot.GetBRoots()[bucket.BucketID-1]
		}
	}

	utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " has objects: ", bucket.NextObjectID)

	var localOps, remoteOps, remoteMax int64
	var data []byte
	op := mpb.OpRecord{}
	if !update {
		ldata, _ := readFromMeta(l.fsID, strconv.FormatInt(bucket.BucketID, 10)+".object")
		if len(ldata) > 0 {
			data = ldata
			odReader := ggio.NewDelimitedReader(bytes.NewBuffer(ldata), len(ldata))
			for {
				err := odReader.ReadMsg(&op)
				if err != nil {
					break
				}

				// filter invalid reccord
				if op.GetOpType() == mpb.LfsOp_OpErr {
					continue
				}
				localOps++
			}
		}
	} else {
		localOps = bucket.applyOpID + 1
	}

	remoteMax = localOps

	rdatas, err := l.getDataFromBlock(int(l.meta.sb.MetaBackupCount), strconv.Itoa(int(-bucket.BucketID)), "1")
	if err == nil && len(rdatas) > 0 {
		for _, rdata := range rdatas {
			if len(rdata) == 0 {
				continue
			}

			remoteOps = 0

			odReader := ggio.NewDelimitedReader(bytes.NewBuffer(rdata), len(rdata))
			for {
				err := odReader.ReadMsg(&op)
				if err != nil {
					break
				}

				if op.GetOpType() == mpb.LfsOp_OpErr {
					continue
				}

				remoteOps++
			}

			if localOps < remoteOps && remoteMax < remoteOps {
				remoteMax = remoteOps
				data = rdata
				utils.MLogger.Infof("remote has newer objects at ops %d in bucket: %d", remoteOps-1, bucket.GetBucketID())
				bucket.dirty = true
			}
		}
	}

	if len(data) == 0 && bucket.GetNextOpID() > 0 {
		return ErrCannotLoadMetaBlock
	}

	if len(data) > 0 && (!update || localOps < remoteMax) {
		if int64(len(data)) < obSize {
			utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " miss some objects")
		}

		odReader := ggio.NewDelimitedReader(bytes.NewBuffer(data), len(data))
		var opNum int64
		for {
			err := odReader.ReadMsg(&op)
			if err != nil {
				break
			}

			if opNum > bucket.applyOpID {
				err = applyOp(bucket, &op)
				if err != nil {
					continue
				}

				l.flushObjectMeta(bucket, false, &op)

				if opNum != op.GetOpID() {
					utils.MLogger.Errorf("ops store %d and calc %d in bucketID %d are mismatch", op.GetOpID(), opNum, bucket.GetBucketID())
				}
				tag, err := proto.Marshal(&op)
				if err != nil {
					continue
				}
				bucket.mtree.Push(tag)
			}

			opNum++

			if bytes.Compare(broot.GetRoot(), bucket.mtree.Root()) == 0 {
				utils.MLogger.Infof("bucket %s at ops %d store op %d expect root %s", bucket.Name, opNum, broot.GetOpCount(), hex.EncodeToString(broot.GetRoot()))
			}

			if broot.GetBucketID() != 0 && opNum == broot.GetOpCount() {
				if bytes.Compare(broot.GetRoot(), bucket.mtree.Root()) != 0 {
					utils.MLogger.Errorf("bucket %s at ops %d expect root %s, but got %s", bucket.Name, opNum, hex.EncodeToString(broot.GetRoot()), hex.EncodeToString(bucket.mtree.Root()))
				}
			}
		}

		bucket.Root = bucket.mtree.Root()
		// verify root
		// verify ops
		if opNum != bucket.GetNextOpID() {
			utils.MLogger.Errorf("Load ops is not correct, expect: %d, but got %d", bucket.GetNextOpID(), opNum)
			if opNum > bucket.GetNextOpID() {
				bucket.NextOpID = opNum
				bucket.dirty = true
			}
		}
	}

	return nil
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
		bucket.applyOpID = op.GetOpID()
		utils.MLogger.Info("Add Object: ", info.GetName(), " in bucket: ", bucket.Name)
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
		ob.ETag = calculateETagForNewPart(ob.ETag, part.ETag)
		ob.Length += part.Length
		ob.MTime = part.GetCTime()
		ob.Unlock()
		bucket.applyOpID = op.GetOpID()
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
		bucket.applyOpID = op.GetOpID()
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
	_, err = os.Stat(obpath)
	if err != nil {
		os.Create(obpath)
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
			obMetaFile, err := os.OpenFile(obpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
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
	if !l.writable {
		return ErrLfsReadOnly
	}

	enc, err := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)
	if err != nil {
		return err
	}

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

	providers, _, err := l.gInfo.GetProviders(ctx, -1)
	if err != nil && len(providers) == 0 {
		return err
	}

	pros := utils.DisorderArray(providers)

	i := 0
	for j := 0; i < metaBackupCount && j < len(pros); j++ {
		bm.SetCid(strconv.Itoa(i))
		ncid := bm.ToString()
		km, _ := metainfo.NewKey(ncid, mpb.KeyType_Block)
		err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[i], pros[j])
		if err != nil {
			continue
		}

		err = l.gInfo.putDataMetaToKeepers(ctx, ncid, pros[j], int(offset))
		if err != nil {
			continue
		}
		i++
	}

	if i < metaBackupCount-1 {
		utils.MLogger.Infof("No enough online providers for metablock:%s", bm.ToString())
	}

	return nil
}

func (l *LfsInfo) getDataFromBlock(metaBackupCount int, buc, stripe string) ([][]byte, error) {
	if l.keySet == nil {
		return nil, role.ErrEmptyBlsKey
	}

	enc, err := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)
	if err != nil {
		return nil, err
	}

	var data [][]byte
	ctx := l.context

	bm, err := metainfo.NewBlockMeta(l.fsID, buc, stripe, "0")
	if err != nil {
		return nil, err
	}

	sig, err := role.BuildSignMessage()
	if err != nil {
		return nil, err
	}

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
			_, _, _, ok := enc.VerifyBlock(b.RawData(), ncid)
			if ok {
				if len(b.RawData()) > 0 {
					res := make([][]byte, 1)
					res[0] = b.RawData()
					rdata, err := enc.Decode(res, 0, -1)
					if err != nil {
						continue
					}
					data = append(data, rdata)
				}
			}
		}
	}

	bm.SetCid(strconv.Itoa(0))
	ncidlocal := bm.ToString()
	km, _ := metainfo.NewKey(ncidlocal, mpb.KeyType_Block)
	b, err := l.ds.GetBlock(ctx, km.ToString(), nil, "local")
	if err == nil && b != nil {
		_, _, _, ok := enc.VerifyBlock(b.RawData(), ncidlocal)
		if ok {
			if len(b.RawData()) > 0 {
				res := make([][]byte, 1)
				res[0] = b.RawData()
				rdata, err := enc.Decode(res, 0, -1)
				if err == nil && len(rdata) > 0 {
					data = append(data, rdata)
				}
			}
		}
	}

	if len(data) > 0 {
		return data, nil
	}

	return nil, role.ErrEmptyData
}

func writeToMeta(data []byte, fsID, buc string) error {
	metapath, err := checkMetaPath(fsID)
	if err != nil {
		return err
	}

	sbpath := path.Join(metapath, buc)
	sbMetaFile, err := os.Create(sbpath)
	if err != nil {
		return err
	}
	defer sbMetaFile.Close()
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
	if err != nil {
		return nil, err
	}
	defer sbMetaFile.Close()
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
