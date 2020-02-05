package user

import (
	"bufio"
	"context"
	"errors"
	"io"
	"math/big"
	"strconv"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
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
	fsID           string
	bucketID       int32
	encrypt        int32
	sKey           [32]byte
	decoder        *dataformat.DataCoder //用于解码数据
	group          *groupInfo            //groupInfo
	buffer         [][]byte
	blockCompleted int       //下载成功了几个块
	writer         io.Writer //用于调度下载，以及返回是否出错
}

func newDownloadJob(uid string, bid, cry int32, sk [32]byte, dec *dataformat.DataCoder, group *groupInfo, writer io.Writer) *downloadJob {
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
	encrypt      int32
	sKey         [32]byte
	group        *groupInfo            //groupInfo
	decoder      *dataformat.DataCoder //用于解码数据
	state        TaskState
	curStripe    int64 //当前已进行到哪个stripe
	segOffset    int64 //此次下载起始offset，表示在stripe中的起始segment
	dStart       int64 // startPos
	dLength      int64 //下载所需大小，用于后续指定范围下载
	sizeReceived int   //可以统计下载进度
	startTime    time.Time
	writer       io.Writer
	completeFunc []CompleteFunc //完成任务或出错的通知函数
}

// GetObject constructs lfs download process
func (l *LfsInfo) GetObject(ctx context.Context, bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error {
	utils.MLogger.Info("Download Object: ", objectName, " from bucket: ", bucketName)
	if !l.online || l.meta.bucketNameToID == nil {
		return ErrLfsServiceNotReady
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
		length = object.OPart.GetLength() - opts.Start
	}

	if opts.Start+length > object.OPart.GetLength() {
		return ErrObjectOptionsInvalid
	}

	start := opts.Start + object.OPart.GetStart()

	bo := bucket.BOpts

	segStripeSize := int64(bo.SegmentSize)
	stripeSize := int64(bo.SegmentCount*bo.DataCount) * segStripeSize

	// 下载的开始条带
	stripePos := start / stripeSize
	// 下载开始的segment
	segPos := (start % stripeSize) / segStripeSize
	// segment的偏移
	dPos := start % segStripeSize

	decoder := dataformat.NewDataCoderWithBopts(bo, l.keySet)

	dl := &downloadTask{
		fsID:         l.fsID,
		bucketID:     bucket.BucketID,
		group:        l.gInfo,
		decoder:      decoder,
		state:        Pending,
		startTime:    time.Now(),
		curStripe:    stripePos,
		segOffset:    segPos,
		dStart:       dPos,
		dLength:      length,
		writer:       writer,
		completeFunc: completeFuncs,
	}

	// default AES
	if bo.Encryption == 1 {
		// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
		dl.sKey = CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucketID, object.OPart.Start)
		dl.encrypt = 1
	}

	return dl.Start(ctx)
}

func (do *downloadTask) Start(ctx context.Context) error {
	curStripe := do.curStripe + 1
	segStart := do.segOffset
	dStart := do.dStart // 0
	dc := do.decoder.Prefix.Bopts.DataCount
	segSize := do.decoder.Prefix.Bopts.SegmentSize
	stripeSize := int64(do.decoder.Prefix.Bopts.SegmentCount * segSize * dc)

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

	breakFlag := false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("download cancel")
			return nil
		default:
			n, err := job.rangeRead(ctx, curStripe-1, segStart, dStart, remain)
			if err != nil {
				do.Complete(err)
				return err
			}

			if n != remain {
				utils.MLogger.Warn("length is not match, got: ", n, ", want: ", remain)
			}

			do.sizeReceived += int(n)

			if int64(do.sizeReceived) >= do.dLength {
				breakFlag = true
			} else {
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
	dataCount := ds.decoder.Prefix.Bopts.DataCount
	parityCount := ds.decoder.Prefix.Bopts.ParityCount
	blockCount := dataCount + parityCount
	bm, err := metainfo.NewBlockMeta(ds.fsID, strconv.Itoa(int(ds.bucketID)), strconv.Itoa(int(stripeID)), "")
	if err != nil {
		utils.MLogger.Error("Download failed: ", err)
		return 0, err
	}

	var count int32
	needRepair := true //是否需要修复
	datas := make([][]byte, blockCount)

	for i := 0; i < int(blockCount); i++ {
		// fails too many, no need to download
		if blockCount-int32(i)+count < dataCount {
			utils.MLogger.Error("Download Obeject failed: ", ErrCannotGetEnoughBlock)
			return 0, ErrCannotGetEnoughBlock
		}

		bm.SetCid(strconv.Itoa(i))
		ncid := bm.ToString()
		provider, _, err := ds.group.getBlockProviders(ncid)
		if err != nil || provider == ds.fsID {
			utils.MLogger.Warnf("Get Block %s 's provider from keeper failed, Err: %s", ncid, err)
			continue
		}

		//user给channel合约签名，发给provider
		mes, money, err := ds.getChannelSign(ncid, provider)
		if err != nil {
			continue
		}

		//获取数据块
		b, err := ds.group.ds.GetBlock(ctx, ncid, mes, provider)
		if err != nil {
			utils.MLogger.Warnf("Get Block %s from %s failed, Err: %s", ncid, provider, err)
			continue
		}
		blkData := b.RawData()
		//需要检查数据块的长度也没问题f
		ok, err := dataformat.VerifyBlockLength(blkData, int(segStart), int(remain))
		if !ok || err != nil {
			utils.MLogger.Errorf("Verify Block %s from %s offset unmatched, Err: %s", ncid, provider, err)
			continue
		}

		ok = ds.decoder.VerifyBlock(blkData, ncid)
		if !ok {
			utils.MLogger.Warn("Fail to verify block: ", ncid, " from:", provider)
			continue
		}

		//下载数据成功，将内存的channel的value更改
		pinfo, ok := ds.group.providers[provider]
		if !ok {
			continue
		}
		if pinfo.chanItem != nil {
			pinfo.chanItem.Value = money
			pinfo.chanItem.Sig = mes
			pinfo.chanItem.Dirty = true
			utils.MLogger.Info("Download success，change channel.value: ", pinfo.chanItem.ChannelID, " to: ", money.String())
		}

		if ds.decoder.Prefix.Bopts.Policy == dataformat.RsPolicy {
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
		utils.MLogger.Error("Download failed: ", ErrCannotGetEnoughBlock)
		return 0, ErrCannotGetEnoughBlock
	}

	ds.decoder.Repair = needRepair
	// decode returns bytes of 16B
	data, err := ds.decoder.Decode(datas, int(segStart), int(offset+remain))
	if err != nil {
		utils.MLogger.Errorf("Download failed due to decode err: ", err)
		return 0, err
	}

	utils.MLogger.Debugf("Download get length: %d, need %d, from %d", len(data), offset+remain, offset)

	if ds.encrypt == 1 {
		padding := aes.BlockSize - ((offset+remain-1)%aes.BlockSize + 1)
		data = data[:offset+remain+padding]
		data, err = aes.AesDecrypt(data, ds.sKey[:])
		if err != nil {
			utils.MLogger.Info("Download failed due to decrypt err: ", err)
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

	wl, err := ds.writer.Write(data[offset : offset+remain])
	if err != nil {
		return 0, err
	}

	if int64(wl) != remain {
		utils.MLogger.Warn("write length is not equal")
	}

	return remain, nil
}

func (ds *downloadJob) getChannelSign(ncid string, provider string) ([]byte, *big.Int, error) {
	// for test
	money := big.NewInt(0)
	hexSK := ds.group.privKey
	channelID := ds.fsID

	pinfo, ok := ds.group.providers[provider]
	if !ok {
		utils.MLogger.Warn(provider, " is not my provider")
		return nil, nil, errors.New("No such provider")
	}

	if pinfo.chanItem != nil {
		channelID = pinfo.chanItem.ChannelID
		addValue := int64((dataformat.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
		money = money.Add(pinfo.chanItem.Value, big.NewInt(addValue)) //100 + valueBase
	}

	sig, err := role.SignForChannel(channelID, hexSK, money)
	if err != nil {
		utils.MLogger.Errorf("Signature about Block %s from %s failed.", ncid, provider)
		return nil, nil, err
	}

	pubKey, err := utils.GetPkFromEthSk(hexSK)
	if err != nil {
		utils.MLogger.Error("Get public key fail: ", err)
		return nil, nil, err
	}

	message := &pb.ChannelSign{
		Sig:       sig,
		PubKey:    pubKey,
		Value:     money.Bytes(),
		ChannelID: channelID,
	}

	mes, err := proto.Marshal(message)
	if err != nil {
		return nil, nil, err
	}

	return mes, money, nil
}
