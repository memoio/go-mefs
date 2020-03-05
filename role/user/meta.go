package user

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"sync"

	ggio "github.com/gogo/protobuf/io"
	"github.com/golang/protobuf/proto"
	config "github.com/memoio/go-mefs/config"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/repo/fsrepo"
	"github.com/memoio/go-mefs/role"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	mt "gitlab.com/NebulousLabs/merkletree"
)

const meta = "meta"

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
	objects   map[string]*objectInfo
	metaCache []byte
	dirty     bool
	sync.RWMutex
	mtree *mt.Tree
}

// objectInfo stores an object meta info
type objectInfo struct {
	mpb.ObjectInfo
	sync.RWMutex
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

	var opRecord mpb.OpRecord
	for _, object := range bucket.objects {
		opRecord.OpType = mpb.LfsOp_OpAdd
		opRecord.OpID = bucket.GetNextOpID()
		bucket.NextOpID++
		opRecord.Payload, err = proto.Marshal(&object.ObjectInfo)
		objectDelimitedWriter.WriteMsg(&opRecord)
		for _, part := range object.Parts {
			opRecord.OpType = mpb.LfsOp_OpAppend
			opRecord.OpID = bucket.GetNextOpID()
			bucket.NextOpID++
			opRecord.Payload, err = proto.Marshal(part)
			objectDelimitedWriter.WriteMsg(&opRecord)
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
	metapath := getMetaPath()
	sbpath := path.Join(metapath, "0")
	sbMetaFile, err := os.Open(sbpath)
	if !os.IsNotExist(err) {
		pbSuperBlock := mpb.SuperBlockInfo{}
		SbDelimitedReader := ggio.NewDelimitedReader(sbMetaFile, dataformat.BlockSize)
		err = SbDelimitedReader.ReadMsg(&pbSuperBlock)
		if err == nil || err == io.EOF {
			return &lfsMeta{
				sb: &superBlock{
					SuperBlockInfo: pbSuperBlock,
					dirty:          false,
				},
				buckets:        make(map[string]*superBucket),
				bucketIDToName: make(map[int64]string),
			}, nil
		}
	}
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
		SbDelimitedReader := ggio.NewDelimitedReader(SbBuffer, len(data))
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
		metapath := getMetaPath()
		bupath := path.Join(metapath, strconv.FormatInt(bucketID, 10))
		buMetaFile, err := os.Open(bupath)
		if !os.IsNotExist(err) {
			bucket := mpb.BucketInfo{}
			BucketDelimitedReader := ggio.NewDelimitedReader(buMetaFile, dataformat.BlockSize)
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
			continue
		}

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
			BucketDelimitedReader := ggio.NewDelimitedReader(BucketBuffer, len(data))
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
	//先从metapath找，是否需要一个验证版本的方法？
	err := l.loadObjectsInfoForMeta(bucket)
	if err != nil {
		return nil
	}

	if len(data) > 0 {
		metapath := getMetaPath()
		obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
		obMetaFile, err := os.Create(obpath)
		obMetaFile.Write(data)
		res := make([][]byte, 1)
		res[0] = data
		data, err := enc.Decode(res, 0, -1)
		if err != nil {
			return err
		}

		if len(data) < int(objectsBlockSize) {
			utils.MLogger.Warn("data length is not equal")
		}
		data = data[:objectsBlockSize]
		opRecordSlice := make([]*mpb.OpRecord, bucket.NextOpID)
		utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " has objects: ", bucket.NextObjectID)
		objectsBuffer := bytes.NewBuffer(data)
		objectsDelimitedReader := ggio.NewDelimitedReader(objectsBuffer, len(data))
		for {
			op := mpb.OpRecord{}
			err := objectsDelimitedReader.ReadMsg(&op)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			applyOp(bucket, &op)
			opRecordSlice[op.GetOpID()] = &op
		}

		for i := 0; i < int(bucket.NextOpID); i++ {
			op := opRecordSlice[i]
			if op == nil {
				continue
			}
			tag := append([]byte(strconv.FormatInt(op.GetOpID(), 10)), op.GetPayload()...)
			bucket.mtree.Push(tag)
		}
	}
	return nil
}

func (l *LfsInfo) getobjectInfoData(bucket *superBucket) ([]byte,error) {
	sig, err := role.BuildSignMessage()
	if err != nil {
		return nil,err
	}
	objectsBlockSize := bucket.ObjectsBlockSize
	if objectsBlockSize == 0 {
		return nil,nil
	}

	metaBackupCount := int(l.meta.sb.MetaBackupCount)
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)

	stripeID := 1 //ObjectsBlock的Stripe从1开始计算
	ctx := context.Background()

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), strconv.Itoa(stripeID), "0")
	if err != nil {
		return nil,err
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
		metapath := getMetaPath()
		obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
		obMetaFile, err := os.Create(obpath)
		if err != 0 {
			return nil,err
		}
		obMetaFile.Write(data)
	}
	return data,nil
}

func (l *LfsInfo) loadObjectsInfoForMeta(bucket *superBucket) error {
	metapath := getMetaPath()
	obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
	obMetaFile, err := os.Open(obpath)
	if !os.IsNotExist(err) {
		opRecordSlice := make([]*mpb.OpRecord, bucket.NextOpID)
		utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " has objects: ", bucket.NextObjectID)
		objectsDelimitedReader := ggio.NewDelimitedReader(obMetaFile, dataformat.BlockSize)
		for {
			op := mpb.OpRecord{}
			err := objectsDelimitedReader.ReadMsg(&op)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			applyOp(bucket, &op)
			opRecordSlice[op.GetOpID()] = &op
		}

		for i := 0; i < int(bucket.NextOpID); i++ {
			op := opRecordSlice[i]
			if op == nil {
				continue
			}
			tag := append([]byte(strconv.FormatInt(op.GetOpID(), 10)), op.GetPayload()...)
			bucket.mtree.Push(tag)
		}
		return nil
	}
	return errors.New("Meta not exist")
}

//应用某个操作
func applyOp(bucket *superBucket, op *mpb.OpRecord) error {
	var err error
	payload := op.GetPayload()
	switch op.OpType {
	case mpb.LfsOp_OpAdd:
		info := mpb.Object{}
		err = proto.Unmarshal(payload, &info)
		if err != nil {
			return err
		}
		bucket.objects[info.GetName()] = &objectInfo{
			ObjectInfo: mpb.ObjectInfo{
				Info:         &info,
				Deletion:     false,
				Length:       0,
				LastModified: info.Ctime,
				PartCount:    0,
				Parts:        make([]*mpb.ObjectPart, 0, 1),
			},
		}
		utils.MLogger.Info("Add Object", info.GetName(), " in bucket: ")
	case mpb.LfsOp_OpAppend:
		part := mpb.ObjectPart{}
		err = proto.Unmarshal(payload, &part)
		if err != nil {
			return err
		}
		ob := bucket.objects[part.GetName()]
		if ob == nil {
			//按理说不会出现往已删除的Object追加数据，后面再调整
			deleteOName := part.GetName() + "." + strconv.FormatInt(part.GetObjectID(), 10)
			if deletedOb := bucket.objects[deleteOName]; deletedOb != nil {
				deletedOb.Lock()
				deletedOb.Parts = append(deletedOb.Parts, &part)
				deletedOb.Unlock()
				return nil
			} else {
				utils.MLogger.Error("Add Part", part.GetPartID(), "to an inexistent object-", part.GetName())
				return ErrObjectNotExist
			}
		}
		//需要上锁么？外层已经上锁了
		ob.Lock()
		ob.Parts = append(ob.Parts, &part)
		ob.Unlock()
	case mpb.LfsOp_OpDelete:
		mes := mpb.DeleteObject{}
		err = proto.Unmarshal(payload, &mes)
		if err != nil {
			return err
		}
		ob := bucket.objects[mes.GetName()]
		if ob == nil {
			utils.MLogger.Error("Delete an inexistent object-", mes.GetName())
			return err
		}
		ob.Deletion = true
		delete(bucket.objects, mes.GetName())
		deleteOName := mes.GetName() + "." + strconv.FormatInt(mes.GetObjectID(), 10)
		bucket.objects[deleteOName] = ob
	case mpb.LfsOp_OpCancel:
		return errors.New("Undefined")
		//这里暂时不实现
	default:
		return errors.New("Undefined")
	}

	return nil
}

func getMetaPath() string {
	rootpath, _ := fsrepo.BestKnownPath()
	metapath, _ := config.Path(rootpath, meta)
	return metapath
}

//检查path是否存在，不存在则新建一个
func checkMetaPath() error {
	metapath := getMetaPath()
	stat, err := os.Stat(metapath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(metapath, 0777)
		return err
	}
	if !stat.IsDir() {
		rootpath, _ := fsrepo.BestKnownPath()
		err := os.Rename(metapath, path.Join(rootpath, "meta.rename"))
		if err != nil {
			os.Remove(metapath)
		}
		err = os.MkdirAll(metapath, 0777)
		return err
	}
	return nil
}
