package user

import (
	"bytes"
	"context"
	"encoding/binary"
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
	"github.com/memoio/go-mefs/utils/metainfo"
	mt "gitlab.com/NebulousLabs/merkletree"
)

const meta = "meta"
const MaxCacheSize = 4 * 1024 * 1024

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
	objects     map[string]*objectInfo
	obMetaCache []byte
	obCacheSize int    //obMetaCache 已经用了多少
	lenBuf      []byte //用于写objectinfo暂存序列化后长度
	dirty       bool
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
	err = l.FlushSbMeta(data)
	if err != nil {
		return err
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

	err = l.ds.PutBlock(ctx, km.ToString(), dataEncoded[0], "local")
	if err != nil {
		utils.MLogger.Errorf("user %s lfs %s superblock persist to local failed. ", l.userID, l.fsID)
		return err
	}
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

	data := bucketBuffer.Bytes()

	err = l.FlushBucketMeta(bucket.GetBucketID(), data)
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, offset, err := enc.Encode(data, ncidPrefix, 0)
	if err != nil {
		return err
	}

	ctx := context.Background()
	err = l.ds.PutBlock(ctx, bm.ToString(), dataEncoded[0], "local")
	if err != nil {
		utils.MLogger.Errorf("user %s lfs %s bucket %s info persist to local failed. ", l.userID, l.fsID, bucket.Name)
		return err
	}
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

	//先把cache都刷盘
	l.FlushObjectMeta(bucket, true)
	metapath := getMetaPath(l.fsID)
	obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
	obMetaFile, err := os.Open(obpath)
	//文件不存在，证明没有刷东西进来
	//TODO:假设是用户自己在刷盘后将还未同步的数据删掉，可能会造成元数据丢失
	if os.IsNotExist(err) {
		obMetaFile.Close()
		return nil
	}
	stat, err := obMetaFile.Stat()
	if stat.Size() == 0 {
		obMetaFile.Close()
		return nil
	}
	//现在每次上传全部
	//以后应该改成索引更新，比如若累计数据大于一个Segment
	//那只用将后面的数据追加进去就行了，这样每次最多更新一个Segment大小的元数据
	data, err := ioutil.ReadAll(obMetaFile)
	if err != nil {
		obMetaFile.Close()
		return err
	}
	obMetaFile.Close()

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

	if len(data) != 0 { //处理最后的剩余部分
		objectsBlockLength += len(data)
		bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
		if err != nil {
			return err
		}

		ncidPrefix := bm.ToString(3)
		dataEncoded, offset, err := enc.Encode(data, ncidPrefix, 0)
		if err != nil {
			return err
		}

		err = l.ds.PutBlock(ctx, bm.ToString(), dataEncoded[0], "local")
		if err != nil {
			utils.MLogger.Errorf("user %s lfs %s objectinfo of %s persist to local failed. ", l.userID, l.fsID, bucket.Name)
			return err
		}

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

		bucket.ObjectsBlockSize = int64(objectsBlockLength)
	}

	return nil
}

//--------------------Load superBlock--------------------------
//lfs启动时加载超级块操作，返回结构体Meta,主要填充其中的superblock字段
//先从本地查找超级快信息，若没找到，就找自己的provider获取
func (l *LfsInfo) loadSuperBlock() (*lfsMeta, error) {
	utils.MLogger.Info("Load superblock: ", l.fsID, " for user:", l.userID)
	metapath := getMetaPath(l.fsID)
	sbpath := path.Join(metapath, "0")
	sbMetaFile, err := os.Open(sbpath)
	if !os.IsNotExist(err) {
		pbSuperBlock := mpb.SuperBlockInfo{}
		SbDelimitedReader := ggio.NewDelimitedReader(sbMetaFile, dataformat.BlockSize)
		err = SbDelimitedReader.ReadMsg(&pbSuperBlock)
		if err == nil {
			SbDelimitedReader.Close()
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
	sbMetaFile.Close()

	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, int(defaultMetaBackupCount-1), l.userID, l.fsID)

	data, err := l.getSuperBlockData(enc)
	if err != nil {
		return nil, err
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

func (l *LfsInfo) getSuperBlockData(enc *dataformat.DataCoder) ([]byte, error) {
	if l.keySet == nil {
		return nil, role.ErrEmptyBlsKey
	}

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
	if len(data) == 0 {
		sig, err := role.BuildSignMessage()
		if err != nil {
			return nil, err
		}
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
		data, err = enc.Decode(res, 0, -1)
		if err != nil {
			utils.MLogger.Info("Decode data fail: ", err)
			return nil, err
		}
		metapath := getMetaPath(l.fsID)
		sbpath := path.Join(metapath, "0")
		sbMetaFile, err := os.Create(sbpath)
		defer sbMetaFile.Close()
		if err != nil {
			return data, nil
		}
		sbMetaFile.Write(data)
	}
	return data, nil
}

//lfs启动进行元数据的加载，对Log中的字段进行初始化 填充除superblock、Entries字段之外的字段
func (l *LfsInfo) loadBucketInfo() error {
	var err error
	metaBackupCount := int(l.meta.sb.MetaBackupCount)
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)
	for bucketID := int64(1); bucketID < l.meta.sb.NextBucketID; bucketID++ {
		err = l.loadSingleBucketInfo(bucketID, enc)
		if err != nil {
			utils.MLogger.Error("Load BucketInfo failed , bucketID: ", bucketID)
		}
	}
	return nil
}

func (l *LfsInfo) loadSingleBucketInfo(bucketID int64, enc *dataformat.DataCoder) error {
	metapath := getMetaPath(l.fsID)
	bupath := path.Join(metapath, strconv.FormatInt(bucketID, 10))
	buMetaFile, err := os.Open(bupath)
	if !os.IsNotExist(err) {
		bucket := mpb.BucketInfo{}
		BucketDelimitedReader := ggio.NewDelimitedReader(buMetaFile, dataformat.BlockSize)
		err = BucketDelimitedReader.ReadMsg(&bucket)
		if err == nil {
			tsb := newSuperBucket(bucket, false)

			tsb.mtree.SetIndex(0)
			tsb.mtree.Push([]byte(l.fsID + bucket.Name))

			bname := bucket.Name
			if bucket.Deletion {
				bname = bucket.Name + "." + strconv.Itoa(int(bucket.BucketID))
			}
			l.meta.buckets[bname] = tsb
			l.meta.bucketIDToName[bucketID] = bname
			buMetaFile.Close()
			return nil
		}
	}
	buMetaFile.Close()

	data, err := l.getBucketInfoData(bucketID, enc)
	if err != nil {
		return err
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

		tsb := newSuperBucket(bucket, false)

		tsb.mtree.SetIndex(0)
		tsb.mtree.Push([]byte(l.fsID + bucket.Name))

		bname := bucket.Name
		if bucket.Deletion {
			bname = bucket.Name + "." + strconv.Itoa(int(bucket.BucketID))
		}
		l.meta.buckets[bname] = tsb
		l.meta.bucketIDToName[bucketID] = bname
		return nil
	}
	return ErrCannotLoadMetaBlock
}

func (l *LfsInfo) getBucketInfoData(bucketID int64, enc *dataformat.DataCoder) ([]byte, error) {
	sig, err := role.BuildSignMessage()
	if err != nil {
		return nil, err
	}
	var data []byte
	ctx := context.Background()
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
		data, err = enc.Decode(res, 0, -1) //Tag暂时没用
		if err != nil {
			utils.MLogger.Info("Decode data fail: ", err)
			return nil, err
		}
		metapath := getMetaPath(l.fsID)
		bupath := path.Join(metapath, strconv.FormatInt(bucketID, 10))
		buMetaFile, err := os.Create(bupath)
		defer buMetaFile.Close()
		if err != nil {
			return nil, err
		}
		buMetaFile.Write(data)
		return data, nil
	}
	return nil, ErrCannotLoadMetaBlock
}

//------------------------------Load Objectinfo---------------------------------------
//填充Entries字段，传入参数为bucket,记录传入bucket的数据信息
func (l *LfsInfo) loadObjectsInfo(bucket *superBucket) error {
	//先从metapath找，是否需要一个验证版本的方法？
	err := l.loadObjectsInfoFromMeta(bucket)
	if err == nil {
		return nil
	}

	objectsBlockSize := bucket.ObjectsBlockSize
	if objectsBlockSize == 0 {
		return nil
	}

	data, err := l.getobjectInfoData(bucket)
	if err != nil {
		return err
	}
	if int64(len(data)) >= objectsBlockSize {
		data = data[:objectsBlockSize]
		utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " has objects: ", bucket.NextObjectID)
		objectsBuffer := bytes.NewBuffer(data)
		objectsDelimitedReader := ggio.NewDelimitedReader(objectsBuffer, len(data))
		op := mpb.OpRecord{}
		for {
			err := objectsDelimitedReader.ReadMsg(&op)
			if err != nil {
				break
			}
			err = applyOp(bucket, &op)
			if err != nil {
				continue
			}
			tag := append([]byte(strconv.FormatInt(op.GetOpID(), 10)), op.GetPayload()...)
			bucket.mtree.Push(tag)
		}
		return nil
	}
	return ErrCannotLoadMetaBlock
}

func (l *LfsInfo) getobjectInfoData(bucket *superBucket) ([]byte, error) {
	sig, err := role.BuildSignMessage()
	if err != nil {
		return nil, err
	}
	objectsBlockSize := bucket.ObjectsBlockSize
	if objectsBlockSize == 0 {
		metapath := getMetaPath(l.fsID)
		obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
		obMetaFile, err := os.Create(obpath)
		defer obMetaFile.Close()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}

	stripeID := 1 //ObjectsBlock的Stripe从1开始计算
	ctx := context.Background()

	metaBackupCount := int(l.meta.sb.MetaBackupCount)
	enc := dataformat.NewDataCoderWithDefault(l.keySet, dataformat.MulPolicy, 1, metaBackupCount-1, l.userID, l.fsID)

	bm, err := metainfo.NewBlockMeta(l.fsID, strconv.Itoa(int(-bucket.BucketID)), strconv.Itoa(stripeID), "0")
	if err != nil {
		return nil, err
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

	//将读取的数据先保存到MetaData
	if len(data) > 0 {
		res := make([][]byte, 1)
		res[0] = data
		data, err = enc.Decode(res, 0, -1)
		if err != nil {
			return nil, err
		}
		if int64(len(data)) >= objectsBlockSize {
			metapath := getMetaPath(l.fsID)
			obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
			obMetaFile, err := os.Create(obpath)
			defer obMetaFile.Close()
			if err != nil {
				return data, err
			}
			realData := data[:objectsBlockSize]
			obMetaFile.Write(realData)
			return realData, nil
		}
	}
	return nil, ErrCannotLoadMetaBlock
}

func (l *LfsInfo) loadObjectsInfoFromMeta(bucket *superBucket) error {
	metapath := getMetaPath(l.fsID)
	obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
	obMetaFile, err := os.Open(obpath)
	defer obMetaFile.Close()
	if !os.IsNotExist(err) {
		utils.MLogger.Info("Objects in bucket: ", bucket.BucketID, " has objects: ", bucket.NextObjectID)
		objectsDelimitedReader := ggio.NewDelimitedReader(obMetaFile, dataformat.BlockSize)
		op := mpb.OpRecord{}
		for {
			err := objectsDelimitedReader.ReadMsg(&op)
			if err != nil {
				break
			}
			err = applyOp(bucket, &op)
			if err != nil {
				continue
			}
			tag := append([]byte(strconv.FormatInt(op.GetOpID(), 10)), op.GetPayload()...)
			bucket.mtree.Push(tag)
		}

		return nil
	}
	return errors.New("Meta not exist")
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
		if ob, ok := bucket.objects[info.GetName()]; ob != nil || ok {
			return ErrObjectAlreadyExist
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
		utils.MLogger.Info("Add Object-", info.GetName(), " in bucket: ", bucket.Name)
	case mpb.LfsOp_OpAppend:
		part := mpb.ObjectPart{}
		err = proto.Unmarshal(payload, &part)
		if err != nil {
			utils.MLogger.Error("OpAppend payload parse failed, bucket: ", bucket.GetName())
			return err
		}
		ob := bucket.objects[part.GetName()]
		if ob == nil {
			utils.MLogger.Error("Add Part-", part.GetPartID(), "to an inexistent object-", part.GetName())
			return ErrObjectNotExist
		}

		ob.Lock()
		if part.PartID < ob.PartCount {
			return ErrObjectAlreadyExist
		}
		ob.Parts = append(ob.Parts, &part)
		ob.PartCount++
		ob.Length += part.Length
		ob.LastModified = part.GetTime()
		ob.Unlock()
	case mpb.LfsOp_OpDelete:
		mes := mpb.DeleteObject{}
		err = proto.Unmarshal(payload, &mes)
		if err != nil {
			utils.MLogger.Error("OpDelete payload parse failed, bucket: ", bucket.GetName())
			return err
		}
		ob := bucket.objects[mes.GetName()]
		if ob == nil {
			utils.MLogger.Error("Delete an inexistent object-", mes.GetName())
			return err
		}
		ob.Lock()
		ob.Deletion = true
		delete(bucket.objects, mes.GetName())
		deleteOName := mes.GetName() + "." + strconv.FormatInt(mes.GetObjectID(), 10)
		bucket.objects[deleteOName] = ob
		ob.Unlock()
	case mpb.LfsOp_OpCancel:
		return errors.New("Undefined")
		//这里暂时不实现
	default:
		return errors.New("Undefined")
	}

	return nil
}

func (l *LfsInfo) FlushSbMeta(data []byte) error {
	metapath, err := checkMetaPath(l.fsID)
	if err != nil {
		return err
	}
	sbpath := path.Join(metapath, "0")
	sbMetaFile, err := os.Create(sbpath)
	defer sbMetaFile.Close()
	if err != nil {
		return err
	}
	sbMetaFile.Write(data)
	return nil
}

func (l *LfsInfo) FlushBucketMeta(bucketID int64, data []byte) error {
	metapath, err := checkMetaPath(l.fsID)
	if err != nil {
		return err
	}
	bupath := path.Join(metapath, strconv.FormatInt(bucketID, 10))
	buMetaFile, err := os.Create(bupath)
	defer buMetaFile.Close()
	if err != nil {
		return err
	}
	buMetaFile.Write(data)
	return nil
}

func (l *LfsInfo) FlushObjectMeta(bucket *superBucket, force bool, ops ...*mpb.OpRecord) error {
	//先检查本地有没有
	metapath, err := checkMetaPath(l.fsID)
	if err != nil {
		return err
	}
	obpath := path.Join(metapath, strconv.FormatInt(bucket.BucketID, 10)+".object")
	stat, err := os.Stat(obpath)
	if os.IsNotExist(err) {
		// 本地这个文件被删了，要先读取回来
		_, err = l.getobjectInfoData(bucket)
		if err != nil {
			return err
		}
	} else if stat.IsDir() {
		err = os.Rename(metapath, metapath+".bak")
		if err != nil {
			err = os.Remove(metapath)
			if err != nil {
				return err
			}
		}
		// 本地这个文件被删了，要先读取回来
		l.getobjectInfoData(bucket)
		if err != nil {
			return err
		}
	}

	if len(ops) == 0 && bucket.obCacheSize == 0 {
		return nil
	}
	var data []byte
	var length uint64
	for _, op := range ops {
		data, _ = proto.Marshal(op)
		length = uint64(len(data))
		n := binary.PutUvarint(bucket.lenBuf, length)

		if bucket.obCacheSize+len(data)+n < MaxCacheSize {
			copy(bucket.obMetaCache[bucket.obCacheSize:], bucket.lenBuf[:n])
			bucket.obCacheSize += n
			copy(bucket.obMetaCache[bucket.obCacheSize:], data)
			bucket.obCacheSize += len(data)
			continue
		}

		//采用追加模式
		obMetaFile, err := os.OpenFile(obpath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			obMetaFile.Close()
			return err
		}
		obMetaFile.Write(bucket.obMetaCache[:bucket.obCacheSize])
		obMetaFile.Write(bucket.lenBuf[:n])
		obMetaFile.Write(data)
		obMetaFile.Close()
		bucket.obCacheSize = 0
	}
	if force && bucket.obCacheSize > 0 {
		//采用追加模式
		obMetaFile, err := os.OpenFile(obpath, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			obMetaFile.Close()
			return err
		}
		obMetaFile.Write(bucket.obMetaCache[:bucket.obCacheSize])
		obMetaFile.Close()
		bucket.obCacheSize = 0
	}
	return nil
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
