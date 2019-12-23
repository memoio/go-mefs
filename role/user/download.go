package user

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/role/user/pb"
	"github.com/memoio/go-mefs/utils"
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
	bucket         *superBucket
	decoder        *dataformat.DataDecoder //用于解码数据
	lfs            *LfsInfo                //lfsservice
	group          *groupInfo              //groupInfo
	buffer         [][]byte
	blockCompleted int       //下载成功了几个块
	writer         io.Writer //用于调度下载，以及返回是否出错
}

func newDownloadJob(bucket *superBucket, decoder *dataformat.DataDecoder, lfs *LfsInfo, group *groupInfo, writer io.Writer) (*downloadJob, error) {
	return &downloadJob{
		bucket:  bucket,
		decoder: decoder,
		lfs:     lfs,
		group:   group,
		writer:  writer,
	}, nil
}

//下载一整个对象的下载任务
type downloadTask struct {
	superBucket  *superBucket
	lService     *LfsInfo   //lfsService
	group        *groupInfo //groupInfo
	object       *objectInfo
	decoder      *dataformat.DataDecoder //用于解码数据
	State        TaskState
	curStripe    int64 //当前已进行到哪个stripe
	offsetStart  int64 //此次下载起始offset，表示在stripe中的起始segment
	indexStart   int64 //此次下载起始index，表示在segment中的起始字节，用于后续指定范围下载
	length       int64 //下载所需大小，用于后续指定范围下载
	sizeReceived int   //可以统计下载进度
	startTime    int64
	writer       io.Writer
	completeFunc []CompleteFunc //完成任务或出错的通知函数
}

// ConstructDownload constructs lfs download process
func (l *LfsInfo) ConstructDownload(bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, options *DownloadOptions) (Job, error) {
	ok := IsOnline(l.userID)
	if !ok {
		return nil, err
	}
	if l.meta.bucketNameToID == nil {
		return nil, ErrBucketNotExist
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return nil, ErrBucketNotExist
	}
	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, ErrBucketNotExist
	}

	objectElement, ok := bucket.objects[objectName]
	if !ok {
		return nil, ErrObjectNotExist
	}
	object, ok := objectElement.Value.(*objectInfo)
	if !ok || object == nil || object.Deletion {
		return nil, ErrObjectNotExist
	}
	if options.Start+options.Length > object.Size {
		return nil, ErrObjectOptionsInvalid
	}
	var curStripe int64
	var offsetStart int64
	var indexStart int64
	var length = options.Length

	if options.Length < 0 {
		length = object.GetSize() - options.Start
	}

	dataCount := bucket.DataCount
	stripeSize := int64(utils.BlockSize * dataCount)
	stripeSegmentSize := int64(bucket.SegmentSize) * int64(dataCount)
	//计算出下载的起始参数
	extraDataSize := object.OffsetStart * int64(bucket.SegmentSize) * int64(dataCount) //在此object之前同一个stripe由其他object占据的空间
	curStripe = (options.Start+extraDataSize)/stripeSize + object.StripeStart
	offsetStart = ((options.Start + extraDataSize) % stripeSize) / stripeSegmentSize
	indexStart = options.Start % stripeSegmentSize

	group := getGroup(l.userid)

	decoder, _ := dataformat.NewDataDecoder(bucket.Policy, bucket.DataCount, bucket.ParityCount)
	return &downloadTask{
		superBucket:  bucket,
		lService:     lfs,
		group:        group,
		object:       object,
		decoder:      decoder,
		State:        Pending,
		startTime:    time.Now().Unix(),
		curStripe:    curStripe,
		offsetStart:  offsetStart,
		indexStart:   indexStart,
		length:       length,
		writer:       writer,
		completeFunc: completeFuncs,
	}, nil
}

func (do *downloadTask) Start(ctx context.Context) error {
	stripeStart := do.curStripe
	offsetStart := do.offsetStart
	indexStart := do.indexStart
	dataCount := do.superBucket.DataCount
	stripeSize := int64(utils.BlockSize * dataCount)

	//下载的第一个stripe前已经有多少数据，等于此文件追加在后面
	extraData := do.offsetStart * int64(do.superBucket.SegmentSize) * int64(dataCount)

	job, _ := newDownloadJob(do.superBucket, do.decoder, do.lService, do.group, do.writer)
	//构造任务并运行
	for {
		var remain int64
		//如果当前offset从0开始
		if offsetStart == 0 {
			if do.length >= (do.curStripe-stripeStart+1)*stripeSize-extraData-do.indexStart {
				//填满一整个stripe
				remain = stripeSize
			} else {
				//最后剩下一部分
				remain = (do.length + extraData + do.indexStart) % (stripeSize)
			}
		} else {
			//只有下载任务的第一个stripe才可能从非0的offset开始
			if do.length <= stripeSize-extraData-do.indexStart {
				//第一种情况，处在某一个stripe的中间
				remain = do.length
			} else {
				//第二种情况，此stripe填到末尾
				remain = stripeSize - extraData - do.indexStart
			}
		}
		//重复读取stripe中所需内容
		n, err := job.rangeRead(ctx, do.curStripe, offsetStart, indexStart, remain)
		if err != nil {
			do.Complete(err)
			return err
		}
		do.sizeReceived += int(n)
		if int64(do.sizeReceived) >= do.length {
			break
		}
		do.curStripe++
		indexStart = 0
		offsetStart = 0
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
func (ds *downloadJob) rangeRead(ctx context.Context, stripeID, offsetStart, indexStart, remain int64) (int64, error) {
	//首先设置一些本次stripe下载的基本参数
	dataCount := ds.bucket.DataCount
	parityCount := ds.bucket.ParityCount
	blockCount := dataCount + parityCount
	bm, err := metainfo.NewBlockMeta(ds.l.userid, strconv.Itoa(int(ds.bucket.BucketID)), strconv.Itoa(int(stripeID)), "")
	if err != nil {
		log.Println("Download failed-", err)
		return 0, err
	}
	// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
	tmpkey := ds.l.privateKey
	tmpkey = append(tmpkey, byte(ds.bucket.BucketID))
	skey := sha256.Sum256(tmpkey)

	cfg, err := localNode.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed, err: ", err)
		return 0, err
	}

	datas := make([][]byte, blockCount)
	var tempReceiveBlockCount int32
	var needRepair = true //是否需要修复
	var data []byte
Loop:
	for i := 0; i < int(blockCount); i++ {
		select {
		case <-ctx.Done(): //取消了本次下载
			fmt.Println("Cancel download")
			return 0, errors.New("Task canceled")
		default:
			if blockCount-int32(i)+tempReceiveBlockCount < dataCount {
				log.Printf("Get Obeject failed, Err: %v\n", ErrCannotGetEnoughBlock)
				return 0, ErrCannotGetEnoughBlock
			}
			bm.SetBid(strconv.Itoa(i))
			ncid := bm.ToString()
			provider, _, err := ds.group.getBlockProviders(ncid)
			if err != nil || provider == ds.l.userid {
				log.Printf("Get Block %s's provider from keeper failed, Err: %v\n", ncid, err)
				continue Loop
			}

			//user给channel合约签名，发给provider
			mes, money, err := ds.getMessage(ncid, provider)
			if err != nil {
				continue Loop
			}

			//获取数据块
			b, err := localNode.Blocks.GetBlockFrom(ctx, provider, ncid, DefaultGetBlockDelay, mes)
			if err != nil {
				log.Printf("Get Block %s from %s failed, Err: %v\n", ncid, provider, err)
				continue Loop
			}
			blkData := b.RawData()
			//需要检查数据块的长度也没问题
			ok, err := dataformat.VerifyBlockLength(blkData, int(offsetStart), int(ds.bucket.TagFlag), int(ds.bucket.SegmentSize), int(dataCount), int(parityCount), int(remain), ds.bucket.Policy)
			if !ok {
				log.Printf("Block %s from %s offset unmatched, Err: %v\n", ncid, provider, err)
				continue Loop
			}

			if ok := dataformat.VerifyBlock(blkData, ncid, ds.group.getKeyset().Pk); !ok || err != nil {
				log.Println("Verify Block failed.", ncid, "from:", provider)
				continue Loop
			}

			if !cfg.Test {
				//下载数据成功，将内存的channel的value更改
				cItem, err := getChannelItem(ds.l.userid, provider)
				if err == nil && cItem != nil {
					log.Println("下载成功，更改内存中channel.value", cItem.ChannelAddr, money.String())
					cItem.Value = money
				}
			}

			if ds.bucket.Policy == dataformat.RsPolicy {
				datas[i] = blkData
			} else {
				datas[0] = blkData
			}
			tempReceiveBlockCount++

			if tempReceiveBlockCount >= dataCount {
				if i == int(dataCount)-1 {
					needRepair = false
				}
				break Loop
			}
		}
	}
	if tempReceiveBlockCount < dataCount {
		log.Println("Download failed-", ErrCannotGetEnoughBlock)
		return 0, ErrCannotGetEnoughBlock
	}
	data, err = ds.decoder.Decode(datas, int(offsetStart), needRepair)
	if err != nil {
		log.Println("Download failed-", err)
		return 0, err
	}
	if int64(len(data)) > remain+indexStart {
		if ds.bucket.Encryption {
			padding := aes.BlockSize - ((remain-1)%aes.BlockSize + 1)
			data = data[:remain+padding] //此处时因为获取的为整块，而文件所需只占里面一部分
			// 先解密，再去padding
			data, err = aes.AesDecrypt(data, skey[:])
			if err != nil {
				log.Println("Download failed-", err)
				return 0, err
			}
			data = data[:len(data)-int(padding)]
			if remain+indexStart > int64(len(data)) {
				return 0, ErrCannotGetEnoughBlock
			}
		}

		if remain+indexStart > int64(len(data)) {
			return 0, ErrCannotGetEnoughBlock
		}

		_, err = ds.writer.Write(data[indexStart : indexStart+remain])
		if err != nil {
			return 0, err
		}
		return remain, nil
	}

	if remain+indexStart > int64(len(data)) {
		return 0, ErrCannotGetEnoughBlock
	}
	if ds.bucket.Encryption {
		data, err = aes.AesDecrypt(data, skey[:])
		if err != nil {
			log.Println("Download failed-", err)
			return 0, err
		}
	}
	//使用匿名函数调度
	_, err = ds.writer.Write(data[indexStart : indexStart+remain])
	if err != nil {
		return 0, err
	}
	return remain, nil
}

func (ds *downloadJob) getMessage(ncid string, provider string) ([]byte, *big.Int, error) {
	money := big.NewInt(0)
	//user给channel合约签名，发给provider
	userID := ds.l.userid
	privateKey := ds.l.privateKey
	localAddress, providerAddress, hexSK, err := buildSignParams(userID, provider, privateKey)
	if err != nil {
		log.Printf("buildSignParams about Block %s from %s failed.\n", ncid, provider)
		return nil, nil, err
	}

	var channelAddr common.Address

	//判断是不是测试user，如果是，就将channelAddress设为0
	cfg, err := localNode.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed.")
		return nil, nil, err
	}
	if cfg.Test {
		channelAddress := contracts.InvalidAddr
		channelAddr = common.HexToAddress(channelAddress)
		//设置此次下载需要签名的金额，money此时不用变仍为0
	} else {
		cItem, err := getChannelItem(userID, provider)
		if err == nil && cItem != nil {
			channelAddr = common.HexToAddress(cItem.ChannelAddr)
			// 此次下载需要签名的金额，在valueBase的基础上再加上此次下载需要支付的money，就是此次签名的value
			addValue := int64((utils.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
			money = money.Add(cItem.Value, big.NewInt(addValue)) //100 + valueBase
		}
	}
	moneyByte := money.Bytes()

	//签名
	sig, err := contracts.SignForChannel(channelAddr, money, hexSK)
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
