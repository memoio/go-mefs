package user

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	"github.com/memoio/go-mefs/role"
	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/address"
	"github.com/memoio/go-mefs/utils/metainfo"
)

const defaultJobsCount = 1

//DownloadOptions 下载时的一些参数
type DownloadOptions struct {
	// Start and end length
	Start, Length int64
}

// DefaultDownloadOptions returns
func DefaultDownloadOptions() *DownloadOptions {
	return &DownloadOptions{
		Start:  0,
		Length: -1,
	}
}

type writeFunc func([]byte, int32, error)

type notif struct {
	err error
	id  int32
}

//针对Bucket内stripe的下载，可复用
type downloadJob struct {
	fsID           string
	bucketID       int32
	encrypt        bool
	sKey           [32]byte
	decoder        *dataformat.DataCoder //用于解码数据
	group          *groupInfo            //groupInfo
	buffer         [][]byte
	blockCompleted int       //下载成功了几个块
	writer         io.Writer //用于调度下载，以及返回是否出错
}

func newDownloadJob(uid string, bid int32, cry bool, sk [32]byte, dec *dataformat.DataCoder, group *groupInfo, writer io.Writer) *downloadJob {
	return &downloadJob{
		fsID:     uid,
		bucketID: bid,
		encrypt:  cry,
		sKey:     sk,
		decoder:  dec,
		group:    group,
		writer:   writer,
	}
}

//下载一整个对象的下载任务
type downloadTask struct {
	fsID         string
	bucketID     int32
	encrypt      bool
	sKey         [32]byte
	group        *groupInfo            //groupInfo
	decoder      *dataformat.DataCoder //用于解码数据
	state        TaskState
	curStripe    int64 //当前已进行到哪个stripe
	segOffset    int64 //此次下载起始offset，表示在stripe中的起始segment
	dStart       int64 //数据其实起始
	dLength      int64 //下载所需大小，用于后续指定范围下载
	sizeReceived int   //可以统计下载进度
	startTime    int64
	writer       io.Writer
	completeFunc []CompleteFunc //完成任务或出错的通知函数
}

// GetObject constructs lfs download process
func (l *LfsInfo) GetObject(bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error {
	log.Println("GetObject: ", bucketName, objectName)
	if !l.online {
		return ErrLfsServiceNotReady
	}

	if l.meta.bucketNameToID == nil {
		return ErrBucketNotExist
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return ErrBucketNotExist
	}
	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return ErrBucketNotExist
	}

	objectElement, ok := bucket.objects[objectName]
	if !ok {
		return ErrObjectNotExist
	}
	object, ok := objectElement.Value.(*objectInfo)
	if !ok || object == nil || object.Deletion {
		return ErrObjectNotExist
	}

	length := opts.Length
	if opts.Length < 0 {
		length = object.GetSize() - opts.Start
	}

	if opts.Start+length > object.Size {
		return ErrObjectOptionsInvalid
	}

	stripeSize := int64(utils.BlockSize * bucket.DataCount)
	segStripeSize := int64(bucket.SegmentSize) * int64(bucket.DataCount)
	//计算出下载的起始参数
	segStart := object.OffsetStart * segStripeSize //在此object之前同一个stripe由其他object占据的空间
	// 下载的开始条带
	stripePos := (opts.Start+segStart)/stripeSize + object.StripeStart
	// 下载开始的segment
	segPos := ((opts.Start + segStart) % stripeSize) / segStripeSize
	// segment的偏移
	offsetPos := opts.Start % segStripeSize

	decoder := dataformat.NewDataCoder(bucket.Policy, bucket.DataCount, bucket.ParityCount, int32(bucket.TagFlag), bucket.SegmentSize, l.keySet)

	dl := &downloadTask{
		fsID:         l.fsID,
		bucketID:     bucket.BucketID,
		group:        l.gInfo,
		decoder:      decoder,
		state:        Pending,
		startTime:    time.Now().Unix(),
		curStripe:    stripePos,
		segOffset:    segPos,
		dStart:       offsetPos,
		dLength:      length,
		writer:       writer,
		completeFunc: completeFuncs,
	}

	if bucket.Encryption {
		// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
		tmpkey := l.privateKey
		tmpkey = append(tmpkey, byte(bucket.BucketID))
		dl.sKey = sha256.Sum256(tmpkey)
		dl.encrypt = true
	}

	return dl.Start(context.Background())
}

func (do *downloadTask) Start(ctx context.Context) error {
	curStripe := do.curStripe + 1
	segStart := do.segOffset
	dStart := do.dStart
	dc := do.decoder.DataCount
	segSize := do.decoder.SegmentSize
	stripeSize := int64(utils.BlockSize * dc)

	//下载的第一个stripe前已经有多少数据，等于此文件追加在后面
	segPos := segStart * int64(segSize) * int64(dc)

	job := newDownloadJob(do.fsID, do.bucketID, do.encrypt, do.sKey, do.decoder, do.group, do.writer)
	//构造任务并运行
	var remain int64
	//只有下载任务的第一个stripe才可能从非0的offset开始
	if do.dLength <= stripeSize-segPos-do.dStart {
		//第一种情况，处在某一个stripe的中间
		remain = do.dLength
	} else {
		//第二种情况，此stripe填到末尾
		remain = stripeSize - segPos - do.dStart
	}

	for {
		//重复读取stripe中所需内容
		n, err := job.rangeRead(ctx, curStripe-1, segStart, dStart, remain)
		if err != nil {
			do.Complete(err)
			return err
		}

		if n != remain {
			log.Println("length is not match")
		}

		do.sizeReceived += int(n)

		if int64(do.sizeReceived) >= do.dLength {
			break
		}

		curStripe++
		dStart = 0
		segStart = 0

		if do.dLength <= (curStripe-do.curStripe)*stripeSize-segPos-do.dStart {
			//最后剩下一部分
			remain = (do.dLength + segPos + do.dStart) % (stripeSize)
		} else {
			//填满一整个stripe
			remain = stripeSize

		}
	}
	if w, ok := do.writer.(*bufio.Writer); ok {
		w.Flush()
	}
	do.Complete(nil)
	return nil
}

func (do *downloadTask) Stop(ctx context.Context) error {
	return nil
}

func (do *downloadTask) Cancel(ctx context.Context) error {
	return nil
}

func (do *downloadTask) Complete(err error) {
	for _, f := range do.completeFunc {
		f(err)
	}
}

func (do *downloadTask) Info() (interface{}, error) {
	return nil, nil
}

type (
	// Destination 日后用于区分是直接写到某路径，还是通过api的
	Destination struct {
		typ    int
		path   string
		writer io.Writer
	}

	// CompleteFunc is a function type that is called when the download completed.
	CompleteFunc func(error) error
)

//从一个stripe内指定范围读取数据写入到writer内
func (ds *downloadJob) rangeRead(ctx context.Context, stripeID, segStart, offset, remain int64) (int64, error) {
	//首先设置一些本次stripe下载的基本参数
	dataCount := ds.decoder.DataCount
	parityCount := ds.decoder.ParityCount
	blockCount := dataCount + parityCount
	bm, err := metainfo.NewBlockMeta(ds.fsID, strconv.Itoa(int(ds.bucketID)), strconv.Itoa(int(stripeID)), "")
	if err != nil {
		log.Println("Download failed-", err)
		return 0, err
	}

	var count int32
	var data []byte
	needRepair := true //是否需要修复
	datas := make([][]byte, blockCount)

	for i := 0; i < int(blockCount); i++ {
		// fails too many, no need to download
		if blockCount-int32(i)+count < dataCount {
			log.Printf("Get Obeject failed, Err: %v\n", ErrCannotGetEnoughBlock)
			return 0, ErrCannotGetEnoughBlock
		}

		bm.SetCid(strconv.Itoa(i))
		ncid := bm.ToString()
		provider, _, err := ds.group.getBlockProviders(ncid)
		if err != nil || provider == ds.fsID {
			log.Printf("Get Block %s's provider from keeper failed, Err: %v\n", ncid, err)
			continue
		}

		//user给channel合约签名，发给provider
		mes, money, err := ds.getMessage(ncid, provider)
		if err != nil {
			continue
		}

		//获取数据块
		b, err := ds.group.ds.GetBlock(ctx, ncid, mes, provider)
		if err != nil {
			log.Printf("Get Block %s from %s failed, Err: %v\n", ncid, provider, err)
			continue
		}
		blkData := b.RawData()
		//需要检查数据块的长度也没问题
		ok, err := dataformat.VerifyBlockLength(blkData, int(segStart), int(ds.decoder.TagFlag), int(ds.decoder.SegmentSize), int(dataCount), int(parityCount), int(remain), ds.decoder.Policy)
		if !ok {
			log.Printf("Block %s from %s offset unmatched, Err: %v\n", ncid, provider, err)
			continue
		}

		if ok := dataformat.VerifyBlock(blkData, ncid, ds.decoder.KeySet.Pk); !ok || err != nil {
			log.Println("Verify Block failed.", ncid, "from:", provider)
			continue
		}

		//下载数据成功，将内存的channel的value更改
		pinfo, ok := ds.group.providers[provider]
		if !ok {
			continue
		}
		if pinfo.chanItem != nil {
			pinfo.chanItem.Value = money
			log.Println("download success，change channel.value", pinfo.chanItem.ChannelID, money.String())
		}

		if ds.decoder.Policy == dataformat.RsPolicy {
			datas[i] = blkData
		} else {
			datas[0] = blkData
		}
		count++

		if count >= dataCount {
			if i == int(dataCount)-1 {
				needRepair = false
			}
			break
		}
	}
	if count < dataCount {
		log.Println("Download failed-", ErrCannotGetEnoughBlock)
		return 0, ErrCannotGetEnoughBlock
	}
	data, err = ds.decoder.Decode(datas, int(segStart), needRepair)
	if err != nil {
		log.Println("Download failed-", err)
		return 0, err
	}

	if remain+offset > int64(len(data)) {
		return 0, ErrCannotGetEnoughBlock
	}

	if int64(len(data)) > remain+offset {
		if ds.encrypt {
			padding := aes.BlockSize - ((remain-1)%aes.BlockSize + 1)
			data = data[:remain+padding] //此处时因为获取的为整块，而文件所需只占里面一部分
			// 先解密，再去padding
			data, err = aes.AesDecrypt(data, ds.sKey[:])
			if err != nil {
				log.Println("Download failed-", err)
				return 0, err
			}
			data = data[:len(data)-int(padding)]
			if remain+offset > int64(len(data)) {
				return 0, ErrCannotGetEnoughBlock
			}
		}

		if remain+offset > int64(len(data)) {
			return 0, ErrCannotGetEnoughBlock
		}

		_, err = ds.writer.Write(data[offset : offset+remain])
		if err != nil {
			return 0, err
		}
		return remain, nil
	}

	if ds.encrypt {
		data, err = aes.AesDecrypt(data, ds.sKey[:])
		if err != nil {
			log.Println("Download failed-", err)
			return 0, err
		}
	}

	//使用匿名函数调度
	wl, err := ds.writer.Write(data[offset : offset+remain])
	if err != nil {
		return 0, err
	}

	if int64(wl) != remain {
		log.Println("write length is not equal")
	}

	return remain, nil
}

func (ds *downloadJob) getMessage(ncid string, provider string) ([]byte, *big.Int, error) {
	money := big.NewInt(0)
	//user给channel合约签名，发给provider
	userID := ds.group.userID
	hexSK := ds.group.privKey

	providerAddress, err := address.GetAddressFromID(provider)
	if err != nil {
		log.Println("GetProAddr err: ", err)
		return nil, nil, err
	}

	localAddress, err := address.GetAddressFromID(userID)
	if err != nil {
		log.Println("GetLocalAddr err: ", err)
		return nil, nil, err
	}

	var channelID string

	pinfo, ok := ds.group.providers[provider]
	if !ok {
		return nil, nil, errors.New("No provider")
	}
	if pinfo.chanItem != nil {
		addValue := int64((utils.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
		money = money.Add(pinfo.chanItem.Value, big.NewInt(addValue)) //100 + valueBase
	}
	moneyByte := money.Bytes()

	//签名
	sig, err := role.SignForChannel(channelID, hexSK, money)
	if err != nil {
		log.Printf("signature about Block %s from %s failed.\n", ncid, provider)
		return nil, nil, err
	}
	//将签名信息、user公钥、user地址、provider地址、签名金额一并发给provider
	pubKey, err := utils.GetCompressedPkFromHexSk(hexSK)
	if err != nil {
		log.Println("get public key error.")
		return nil, nil, err
	}

	message := &pb.SignForChannel{
		Sig:             sig,
		UserPK:          pubKey,
		UserAddress:     localAddress.String(),
		ProviderAddress: providerAddress.String(),
		Money:           moneyByte,
	}
	mes, err := proto.Marshal(message)
	if err != nil {
		log.Println("protoMarshal about Block", ncid, "from", provider, "failed. err:", err)
		return nil, nil, err
	}
	return mes, money, nil
}
