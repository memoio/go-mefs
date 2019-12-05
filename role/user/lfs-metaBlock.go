package user

import (
	"bytes"
	"container/list"
	"context"
	"io"
	"log"
	"strconv"

	ggio "github.com/gogo/protobuf/io"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/bitset"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const metaTagFlag = dataformat.BLS12

//----------------------Flush superBlock---------------------------

//刷新超级块
func (lfs *LfsService) flushSuperBlock() error {
	err := lfs.flushSuperBlockLocal()
	if err != nil {
		return err
	}
	return lfs.flushSuperBlockToProvider()
}

//保存超级块信息到本地，传入参数为超级快结构体
func (lfs *LfsService) flushSuperBlockLocal() error {
	sb := lfs.meta.sb
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

	bm, err := metainfo.NewBlockMeta(lfs.userid, "0", "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, _, err := dataformat.DataEncodeToMul(data, ncidPrefix, 1, 0, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
	if err != nil {
		return err
	}
	ncid := bm.ToString()
	bcid := cid.NewCidV2([]byte(ncid))
	b, err := blocks.NewBlockWithCid(dataEncoded[0], bcid)
	if err != nil {
		return err
	}
	err = localNode.Blocks.DeleteBlock(bcid)
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	err = localNode.Blocks.PutBlock(b)
	if err != nil {
		return ErrCannotAddBlock
	}

	return nil
}

func (lfs *LfsService) flushSuperBlockToProvider() error {
	sb := lfs.meta.sb
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

	bm, err := metainfo.NewBlockMeta(lfs.userid, "0", "0", "")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, offset, err := dataformat.DataEncodeToMul(data, ncidPrefix, 1, sb.MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
	if err != nil {
		return err
	}
	providers, err := getGroupService(lfs.userid).getProviders(int(sb.MetaBackupCount))
	if err != nil {
		return err
	}
	for j := 0; j < len(providers); j++ { //
		bm.SetBid(strconv.Itoa(j))
		ncid := bm.ToString()

		km, err := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(int(offset)))
		if err != nil {
			continue
		}
		updateKey := km.ToString()
		bcid := cid.NewCidV2([]byte(updateKey))
		b, err := blocks.NewBlockWithCid(dataEncoded[j], bcid)
		if err != nil {
			return err
		}
		err = localNode.Blocks.PutBlockTo(b, providers[j])
		if err != nil {
			return err
		}
		err = getGroupService(lfs.userid).putDataMetaToKeepers(ncid, providers[j], int(offset))
		if err != nil {
			return err
		}
	}
	return nil
}

//-----------------------Flush BucketMeta----------------------------

func (lfs *LfsService) flushBucketInfo(bucket *superBucket) error {
	err := lfs.flushBucketInfoLocal(bucket)
	if err != nil {
		return err
	}
	return lfs.flushBucketInfoToProvider(bucket)
}

func (lfs *LfsService) flushBucketInfoLocal(bucket *superBucket) error {
	bucket.RLock()
	defer bucket.RUnlock()
	BucketBuffer := bytes.NewBuffer(nil)
	BucketDelimitedWriter := ggio.NewDelimitedWriter(BucketBuffer)
	err := BucketDelimitedWriter.WriteMsg(&bucket.BucketInfo)
	if err != nil {
		return err
	}

	bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucket.BucketID)), "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	dataEncoded, _, err := dataformat.DataEncodeToMul(BucketBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
	if err != nil {
		return err
	}
	err = BucketDelimitedWriter.Close()
	if err != nil {
		return err
	}
	ncid := bm.ToString()
	bcid := cid.NewCidV2([]byte(ncid))
	b, err := blocks.NewBlockWithCid(dataEncoded[0], bcid)
	if err != nil {
		return err
	}
	err = localNode.Blocks.DeleteBlock(bcid)
	if err != nil && err != bs.ErrNotFound {
		return err
	}
	err = localNode.Blocks.PutBlock(b)
	if err != nil {
		return err
	}
	return nil
}

func (lfs *LfsService) flushBucketInfoToProvider(bucket *superBucket) error {
	bucket.RLock()
	defer bucket.RUnlock()
	MetaBackupCount := lfs.meta.sb.MetaBackupCount
	providers, _ := getGroupService(lfs.userid).getProviders(int(MetaBackupCount))
	if providers == nil {
		return ErrNoProviders
	}
	BucketBuffer := bytes.NewBuffer(nil)
	BucketDelimitedWriter := ggio.NewDelimitedWriter(BucketBuffer)
	err := BucketDelimitedWriter.WriteMsg(&bucket.BucketInfo)
	if err != nil {
		return err
	}

	bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucket.BucketID)), "0", "0")
	if err != nil {
		return err
	}
	ncidPrefix := bm.ToString(3)
	BucketBytes := BucketBuffer.Bytes()
	dataEncoded, offset, err := dataformat.DataEncodeToMul(BucketBytes, ncidPrefix, 1, MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
	for j := 0; j < int(MetaBackupCount); j++ { //
		bm.SetBid(strconv.Itoa(j))
		ncid := bm.ToString()
		if err != nil {
			return err
		}
		km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(int(offset)))
		updateKey := km.ToString()
		bcid := cid.NewCidV2([]byte(updateKey))
		b, err := blocks.NewBlockWithCid(dataEncoded[j], bcid)
		if err != nil {
			return err
		}
		err = localNode.Blocks.PutBlockTo(b, providers[j])
		if err != nil {
			return err
		}
		err = getGroupService(lfs.userid).putDataMetaToKeepers(ncid, providers[j], int(offset))
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
func (lfs *LfsService) flushObjectsInfo(bucket *superBucket) error {
	if bucket == nil || bucket.objects == nil {
		return nil
	}
	err := lfs.flushObjectsInfoLocal(bucket)
	if err != nil {
		return err
	}
	return lfs.flushObjectsInfoToProvider(bucket)
}

func (lfs *LfsService) flushObjectsInfoLocal(bucket *superBucket) error {
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
	for objectElement := bucket.orderedObjects.Front(); objectElement != nil; objectElement = objectElement.Next() {
		object, ok := objectElement.Value.(*objectInfo)
		if !ok {
			continue
		}
		if objectsBuffer.Len() >= utils.BlockSize { //如果object的总长度大于规定的size，则分块
			objectsBlockLength += objectsBuffer.Len()
			// dataEncoded, _, err := dataformat.DataEncode(objectsBuffer.Bytes(), dataformat.DefaultSegmentSize, metaTagFlag)
			bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
			if err != nil {
				return err
			}
			ncidPrefix := bm.ToString(3)
			dataEncoded, _, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
			// dataEncoded, _, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, dataformat.BLS, AllUsers.getGroupService(lfs.userid).keySet)
			if err != nil {
				return err
			}
			ncid := bm.ToString()
			bcid := cid.NewCidV2([]byte(ncid))
			b, err := blocks.NewBlockWithCid(dataEncoded[0], bcid)
			if err != nil {
				return err
			}
			err = localNode.Blocks.DeleteBlock(bcid)
			if err != nil && err != bs.ErrNotFound {
				return err
			}
			err = localNode.Blocks.PutBlock(b)
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
		bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
		if err != nil {
			return err
		}
		ncidPrefix := bm.ToString(3)
		dataEncoded, _, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, flushLocalBackup, 0, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
		if err != nil {
			return err
		}
		ncid := bm.ToString()
		bcid := cid.NewCidV2([]byte(ncid))
		b, err := blocks.NewBlockWithCid(dataEncoded[0], bcid)
		if err != nil {
			return err
		}
		err = localNode.Blocks.DeleteBlock(bcid)
		if err != nil && err != bs.ErrNotFound {
			return err
		}
		err = localNode.Blocks.PutBlock(b)
		if err != nil {
			return ErrCannotAddBlock
		}
		err = objectDelimitedWriter.Close() //结束
		if err != nil {
			return err
		}
	}
	lfs.meta.bucketByID[bucketID].ObjectsBlockSize = int64(objectsBlockLength)
	return nil
}

func (lfs *LfsService) flushObjectsInfoToProvider(bucket *superBucket) error {
	if bucket == nil || bucket.objects == nil {
		return nil
	}
	bucket.RLock()
	defer bucket.RUnlock()
	MetaBackupCount := lfs.meta.sb.MetaBackupCount
	providers, _ := getGroupService(lfs.userid).getProviders(int(MetaBackupCount))
	if providers == nil {
		return ErrNoProviders
	}
	objectsBuffer := bytes.NewBuffer(nil)
	objectDelimitedWriter := ggio.NewDelimitedWriter(objectsBuffer)

	bucketID := bucket.BucketID
	objectsStripeID := 1
	objectsBlockLength := 0
	for objectElement := bucket.orderedObjects.Front(); objectElement != nil; objectElement = objectElement.Next() {
		object, ok := objectElement.Value.(*objectInfo)
		if !ok {
			continue
		}
		if objectsBuffer.Len() >= utils.BlockSize { //如果object的总长度大于规定的size，则分块
			objectsBlockLength += objectsBuffer.Len()
			bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
			if err != nil {
				return err
			}
			ncidPrefix := bm.ToString(3)
			dataEncoded, offset, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, 1, MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
			if err != nil {
				return err
			}
			for j := 0; j < len(providers); j++ {
				bm.SetBid(strconv.Itoa(j))
				ncid := bm.ToString()
				km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(int(offset)))
				updateKey := km.ToString()
				bcid := cid.NewCidV2([]byte(updateKey))
				b, err := blocks.NewBlockWithCid(dataEncoded[j], bcid)
				if err != nil {
					return err
				}
				err = localNode.Blocks.PutBlockTo(b, providers[j])
				if err != nil {
					return err
				}

				err = getGroupService(lfs.userid).putDataMetaToKeepers(ncid, providers[j], int(offset))
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
		bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucketID)), strconv.Itoa(objectsStripeID), "0")
		if err != nil {
			return err
		}
		ncidPrefix := bm.ToString(3)
		dataEncoded, offset, err := dataformat.DataEncodeToMul(objectsBuffer.Bytes(), ncidPrefix, 1, MetaBackupCount-1, dataformat.DefaultSegmentSize, metaTagFlag, getGroupService(lfs.userid).getKeyset())
		if err != nil {
			return err
		}
		for j := 0; j < len(providers); j++ {
			bm.SetBid(strconv.Itoa(j))
			ncid := bm.ToString()
			km, _ := metainfo.NewKeyMeta(ncid, metainfo.PutBlock, "update", "0", strconv.Itoa(int(offset)))
			updateKey := km.ToString()
			bcid := cid.NewCidV2([]byte(updateKey))
			b, err := blocks.NewBlockWithCid(dataEncoded[j], bcid)
			if err != nil {
				return err
			}
			err = localNode.Blocks.PutBlockTo(b, providers[j])
			if err != nil {
				return err
			}

			err = getGroupService(lfs.userid).putDataMetaToKeepers(ncid, providers[j], int(offset))
			if err != nil {
				return err
			}
		}
		err = objectDelimitedWriter.Close()
		if err != nil {
			return err
		}
	}
	lfs.meta.bucketByID[bucketID].ObjectsBlockSize = int64(objectsBlockLength)
	return nil
}

//--------------------Load superBlock--------------------------
//lfs启动时加载超级块操作，返回结构体Meta,主要填充其中的superblock字段
//先从本地查找超级快信息，若没找到，就找自己的provider获取
func (lfs *LfsService) loadSuperBlock() (*lfsMeta, error) {
	log.Println("Begin to load superblock : ", lfs.userid)
	var b blocks.Block
	var err error
	sig, err := BuildSignMessage()
	if err != nil {
		return nil, err
	}

	bm, err := metainfo.NewBlockMeta(lfs.userid, "0", "0", "0")
	if err != nil {
		return nil, err
	}
	ncidlocal := bm.ToString()
	bcidlocal := cid.NewCidV2([]byte(ncidlocal))
	if getGroupService(lfs.userid).getKeyset() == nil {
		return nil, ErrKeySetIsNil
	}
	if b, err = localNode.Blocks.GetBlock(context.Background(), bcidlocal); b != nil && err == nil && dataformat.VerifyBlock(b.RawData(), ncidlocal, getGroupService(lfs.userid).getKeyset().Pk) { //如果本地有这个块的话，无需麻烦Provider
	} else { //若本地无超级块，向自己的provider进行查询
		err = localNode.Blocks.DeleteBlock(bcidlocal)
		if err != nil && err != bs.ErrNotFound {
			return nil, err
		}
		log.Printf("Cannot Get superBlock in block %s from local datastore. Try to get it from remote servers.\n", ncidlocal)
		for j := 0; j < int(defaultMetaBackupCount); j++ {
			bm.SetBid(strconv.Itoa(j))
			ncid := bm.ToString()
			provider, _, err := getGroupService(lfs.userid).getBlockProviders(ncid) //获取数据块的保存位置
			if (provider == "" || err != nil) && j < int(defaultMetaBackupCount)-1 {
				continue
			} else if err != nil {
				log.Println("Cannot load Lfs superblock.", err)
				return nil, ErrCannotLoadMetaBlock
			}
			b, err = localNode.Blocks.GetBlockFrom(localNode.Context(), provider, ncid, DefaultGetBlockDelay, sig) //向指定provider查询超级块
			if err != nil {                                                                                        //*错误处理
				log.Printf("Get metablock %s from %s failed: %s.\n", ncid, provider, err)
				continue
			}
			if b != nil { //获取到有效数据块，跳出
				if ok := dataformat.VerifyBlock(b.RawData(), ncid, getGroupService(lfs.userid).getKeyset().Pk); !ok {
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
func (lfs *LfsService) loadBucketInfo() error {
	sig, err := BuildSignMessage()
	if err != nil {
		return err
	}

	state, err := getUserState(lfs.userid)
	if err != nil {
		return err
	}
	if state < groupStarted {
		return ErrGroupServiceNotReady
	}
	if getGroupService(lfs.userid).getKeyset() == nil {
		return ErrKeySetIsNil
	}

	for bucketID, ok := lfs.meta.sb.bitsetInfo.NextSet(0); ok; bucketID, ok = lfs.meta.sb.bitsetInfo.NextSet(bucketID + 1) {
		if !ok {
			break
		}
		var b blocks.Block
		var err error
		bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucketID)), "0", "0")
		if err != nil {
			return err
		}
		ncidlocal := bm.ToString()
		bcidlocal := cid.NewCidV2([]byte(ncidlocal))
		if b, err = localNode.Blocks.GetBlock(context.Background(), bcidlocal); b != nil && err == nil && dataformat.VerifyBlock(b.RawData(), ncidlocal, getGroupService(lfs.userid).getKeyset().Pk) { //如果本地有这个块的话，无需麻烦Provider
		} else {
			err = localNode.Blocks.DeleteBlock(bcidlocal)
			if err != nil && err != bs.ErrNotFound {
				return err
			}
			log.Printf("Cannot Get BucketInfo in block %s from local datastore. Maybe block is lost or broken.\n", ncidlocal)
			ncidprefix := bm.ToString(3)
			for j := 0; j < int(lfs.meta.sb.MetaBackupCount); j++ {
				ncid := ncidprefix + "_" + strconv.Itoa(j)
				provider, _, err := getGroupService(lfs.userid).getBlockProviders(ncid) //获取保存位置
				if err != nil && j == int(lfs.meta.sb.MetaBackupCount)-1 {
					log.Printf("load superBucket: %d's block: %s from provider: %s falied.\n", bucketID, ncid, provider)
					continue
				}
				b, err = localNode.Blocks.GetBlockFrom(localNode.Context(), provider, ncid, DefaultGetBlockDelay, sig) //获取数据块
				if b != nil && err == nil {
					if ok := dataformat.VerifyBlock(b.RawData(), ncid, getGroupService(lfs.userid).getKeyset().Pk); !ok {
						log.Println("Verify Block failed.", ncid, "from:", provider)
					} else {
						break
					}
				} else if err != nil && j == int(lfs.meta.sb.MetaBackupCount)-1 {
					log.Println("load superBucket error:", bucketID, err)
				}
			}

		}

		if b != nil {
			data, err := dataformat.GetDataFromRawData(b.RawData()) //Tag暂时没用
			if err != nil {                                         //*错误处理
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
			lfs.meta.bucketByID[int32(bucketID)] = &superBucket{
				BucketInfo:     bucket,
				objects:        objects,
				orderedObjects: list.New(),
				dirty:          false,
			}
			lfs.meta.bucketNameToID[bucket.Name] = bucket.BucketID
			log.Println("superBucket-ID:", bucket.BucketID, "Name-", bucket.Name, "is loaded")
		}
	}
	return nil
}

//------------------------------Load Objectinfo---------------------------------------
//填充Entries字段，传入参数为bucket,记录传入bucket的数据信息
func (lfs *LfsService) loadObjectsInfo(bucket *superBucket) error {
	sig, err := BuildSignMessage()
	if err != nil {
		return err
	}
	state, err := getUserState(lfs.userid)
	if err != nil {
		return err
	}
	if state < groupStarted {
		return ErrGroupServiceNotReady
	}
	if getGroupService(lfs.userid).getKeyset() == nil {
		return ErrKeySetIsNil
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
		bm, err := metainfo.NewBlockMeta(lfs.userid, strconv.Itoa(int(-bucket.BucketID)), strconv.Itoa(stripeID), "0")
		if err != nil {
			return err
		}
		ncidlocal := bm.ToString()
		bcidlocal := cid.NewCidV2([]byte(ncidlocal))
		if b, err = localNode.Blocks.GetBlock(context.Background(), bcidlocal); b != nil && err == nil && dataformat.VerifyBlock(b.RawData(), ncidlocal, getGroupService(lfs.userid).getKeyset().Pk) { //如果本地有这个块的话，无需麻烦Provider
		} else {
			err = localNode.Blocks.DeleteBlock(bcidlocal)
			if err != nil && err != bs.ErrNotFound {
				return err
			}
			log.Printf("Cannot Get ObjectInfo in block %s from local datastore. Maybe block is lost or broken.\n", ncidlocal)
			for j := 0; j < int(lfs.meta.sb.MetaBackupCount); j++ {
				bm.SetBid(strconv.Itoa(j))
				ncid := bm.ToString()
				provider, _, err := getGroupService(lfs.userid).getBlockProviders(ncid)
				if err != nil && j == int(lfs.meta.sb.MetaBackupCount)-1 {
					return ErrCannotLoadMetaBlock
				}
				b, err = localNode.Blocks.GetBlockFrom(localNode.Context(), provider, ncid, DefaultGetBlockDelay, sig)
				if b != nil && err == nil {
					if ok := dataformat.VerifyBlock(b.RawData(), ncid, getGroupService(lfs.userid).getKeyset().Pk); !ok {
						log.Println("Verify Block failed.", ncid, "from:", provider)
					} else {
						break
					}
				} else if err != nil && j == int(lfs.meta.sb.MetaBackupCount)-1 {
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
