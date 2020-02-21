package user

import (
	"bufio"
	"context"
	"crypto/cipher"
	"io"
	"math/big"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/role"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
)

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

type downloadTask struct {
	bucketID     int64
	encrypt      int32
	sKey         [32]byte
	group        *groupInfo            //groupInfo
	decoder      *dataformat.DataCoder //用于解码数据
	curStripe    int64                 //当前已进行到哪个stripe
	segOffset    int64                 //此次下载起始offset，表示在stripe中的起始segment
	dStart       int64                 // startPos
	dLength      int64                 //下载所需大小，用于后续指定范围下载
	sizeReceived int64                 //可以统计下载进度
	startTime    time.Time
	writer       io.Writer
	completeFunc []CompleteFunc //完成任务或出错的通知函数
}

// GetObject constructs lfs download process
func (l *LfsInfo) GetObject(ctx context.Context, bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts *DownloadOptions) error {
	utils.MLogger.Info("Download Object: ", objectName, " from bucket: ", bucketName)
	if !l.online || l.meta.buckets == nil {
		return ErrLfsServiceNotReady
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return ErrBucketNotExist
	}

	object, ok := bucket.objects[objectName]
	if !ok {
		return ErrObjectNotExist
	}

	object.RLock()
	defer object.RUnlock()

	if object.Deletion {
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

	segStripeSize := int64(bo.SegmentSize * bo.DataCount)
	stripeSize := int64(bo.SegmentCount) * segStripeSize

	// 下载的开始条带内偏移
	stripePos := start / stripeSize
	// 下载开始的segment
	segPos := (start % stripeSize) / segStripeSize
	// segment内的偏移
	dPos := start % segStripeSize

	bopt := &mpb.BlockOptions{
		Bopts:   bo,
		Start:   0,
		UserID:  l.userID,
		QueryID: l.fsID,
	}

	decoder := dataformat.NewDataCoderWithPrefix(l.keySet, bopt)

	dl := &downloadTask{
		bucketID:     bucket.BucketID,
		group:        l.gInfo,
		decoder:      decoder,
		startTime:    time.Now(),
		curStripe:    stripePos,
		segOffset:    segPos,
		dStart:       dPos,
		dLength:      length,
		writer:       writer,
		completeFunc: completeFuncs,
		encrypt:      bo.Encryption,
	}

	// default AES
	if bo.Encryption == 1 {
		dl.sKey = aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.OPart.Start)
	}

	return dl.Start(ctx)
}

func (do *downloadTask) Start(ctx context.Context) error {
	curStripe := do.curStripe
	segStart := do.segOffset
	dStart := do.dStart // 0
	dc := int64(do.decoder.Prefix.Bopts.DataCount)
	segStripeSize := int64(do.decoder.Prefix.Bopts.SegmentSize) * dc
	stripeSize := int64(do.decoder.Prefix.Bopts.SegmentCount) * segStripeSize

	//下载的第一个stripe前已经有多少数据，等于此文件追加在后面
	segPos := segStart * segStripeSize
	readUnit := int64(transNum) * segStripeSize

	var bEnc cipher.BlockMode
	if do.encrypt == 1 {
		tmpEnc, err := aes.ContructAes(do.sKey[:])
		if err != nil {
			return err
		}
		bEnc = tmpEnc
	}

	var remain int64

	breakFlag := false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("download cancel")
			return nil
		default:
			if (do.sizeReceived+segPos+do.dStart)%stripeSize == 0 {
				curStripe++
				segStart = 0
			}

			remain = (curStripe-do.curStripe+1)*stripeSize - segPos - do.dStart - do.sizeReceived

			if remain > do.dLength-do.sizeReceived {
				remain = do.dLength - do.sizeReceived
			}

			if remain > stripeSize {
				remain = stripeSize
			}

			if remain > readUnit {
				remain = readUnit
			}

			data, n, err := do.rangeRead(ctx, curStripe, segStart, dStart, remain)
			if err != nil {
				if err.Error() == role.ErrWrongMoney.Error() {
					do.group.loadContracts("")
					continue
				} else {
					do.Complete(err)
					return err
				}
			}

			if do.encrypt == 1 {
				padding := aes.BlockSize - ((dStart+remain-1)%aes.BlockSize + 1)
				data = data[:dStart+remain+padding]
				decrypted := make([]byte, len(data))
				bEnc.CryptBlocks(decrypted, data)
				data = decrypted[:len(data)-int(padding)]
			}

			if remain+dStart > int64(len(data)) {
				return ErrCannotGetEnoughBlock
			}

			wl, err := do.writer.Write(data[dStart : dStart+remain])
			if err != nil {
				return err
			}

			if int64(wl) != remain {
				utils.MLogger.Warn("write length is not equal")
			}

			if n != remain {
				utils.MLogger.Warn("length is not match, got: ", n, ", want: ", remain)
			}

			do.sizeReceived += n

			if do.sizeReceived >= do.dLength {
				breakFlag = true
			} else {
				dStart = 0
				segStart += (1 + (n-1)/(segStripeSize))
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
func (do *downloadTask) rangeRead(ctx context.Context, stripeID, segStart, offset, remain int64) ([]byte, int64, error) {
	//首先设置一些本次stripe下载的基本参数
	dataCount := do.decoder.Prefix.Bopts.DataCount
	parityCount := do.decoder.Prefix.Bopts.ParityCount
	blockCount := dataCount + parityCount
	segSize := do.decoder.Prefix.Bopts.SegmentSize
	bm, err := metainfo.NewBlockMeta(do.group.groupID, strconv.Itoa(int(do.bucketID)), strconv.Itoa(int(stripeID)), "")
	if err != nil {
		utils.MLogger.Error("Download failed: ", err)
		return nil, 0, err
	}

	segRemains := int(1 + (remain-1)/int64(dataCount*segSize))

	tagSize, ok := dataformat.TagMap[int(do.decoder.Prefix.Bopts.TagFlag)]
	if !ok {
		tagSize = 48
	}

	do.decoder.Prefix.Start = int32(segStart)
	_, preLen, err := bf.PrefixEncode(do.decoder.Prefix)
	if err != nil {
		return nil, 0, err
	}

	eachLen := preLen + segRemains*(int(segSize)+int(2+(parityCount-1)/dataCount)*tagSize)

	needRepair := false //是否需要修复
	datas := make([][]byte, blockCount)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	parllel := int32(0)
	success := int32(0)
	fail := int32(0)
	wrongMoney := int32(0)
	for i := 0; i < int(blockCount); i++ {
		// fails too many, no need to download
		if atomic.LoadInt32(&fail) > parityCount {
			utils.MLogger.Error("Download obeject failed too much: ", ErrCannotGetEnoughBlock)
			continue
		}

		if i > int(dataCount)-1 {
			needRepair = true
		}

		atomic.AddInt32(&parllel, 1)
		bm.SetCid(strconv.Itoa(i))
		ncid := bm.ToString()
		go func(inum int, chunkid string) {
			defer atomic.AddInt32(&parllel, -1)
			defer atomic.AddInt32(&fail, 1)
			provider, _, err := do.group.getBlockProviders(chunkid)
			if err != nil || provider == do.group.groupID {
				utils.MLogger.Warnf("Get Block %s 's provider from keeper failed: %s", chunkid, err)
				return
			}

			pinfo, ok := do.group.providers[provider]
			if !ok {
				utils.MLogger.Warn(provider, " is not my provider")
				return
			}

			//user给channel合约签名，发给provider
			mes, money, err := do.getChannelSign(pinfo.chanItem, eachLen)
			if err != nil {
				if do.group.userID != do.group.groupID {
					utils.MLogger.Warnf("get channel fails: %s", err)
					return
				}
			}

			//获取数据块
			bgm, _ := metainfo.NewKey(chunkid, mpb.KeyType_Block, strconv.Itoa(int(segStart)), strconv.Itoa(segRemains))
			b, err := do.group.ds.GetBlock(ctx, bgm.ToString(), mes, provider)
			if err != nil {
				utils.MLogger.Warnf("Get Block %s from %s failed: %s", ncid, provider, err)
				if err.Error() == role.ErrWrongMoney.Error() {
					utils.MLogger.Infof("Try load channel value from %s", provider)
					atomic.AddInt32(&wrongMoney, 1)
				}

				if err.Error() == role.ErrNotEnoughMoney.Error() {
					atomic.AddInt32(&wrongMoney, 1)
					do.group.loadContracts(provider)
				}
				return
			}
			blkData := b.RawData()
			ok, err = dataformat.VerifyBlockLength(blkData, int(segStart), segRemains)
			if !ok || err != nil {
				utils.MLogger.Errorf("Verify Block %s from %s offset unmatched, Err: %s", chunkid, provider, err)
				return
			}

			_, _, ok = do.decoder.VerifyBlock(blkData, chunkid)
			if !ok {
				utils.MLogger.Warn("Fail to verify block: ", chunkid, " from:", provider)
				return
			}

			//下载数据成功，将内存的channel的value更改
			if pinfo.chanItem != nil {
				pinfo.chanItem.Value = money
				pinfo.chanItem.Sig = mes
				pinfo.chanItem.Dirty = true
				utils.MLogger.Info("Download success，change channel.value: ", pinfo.chanItem.ChannelID, " to: ", money.String())

				key, err := metainfo.NewKey(pinfo.chanItem.ChannelID, mpb.KeyType_Channel)
				if err == nil {
					do.group.ds.PutKey(ctx, key.ToString(), mes, nil, "local")
				}
			}

			datas[inum] = blkData
			atomic.AddInt32(&success, 1)
			atomic.AddInt32(&fail, -1)
		}(i, ncid)

		if i >= int(dataCount-1) {
			for {
				if atomic.LoadInt32(&parllel)+atomic.LoadInt32(&success) < dataCount {
					break
				}

				if atomic.LoadInt32(&success) == dataCount {
					break
				}
				time.Sleep(time.Second)
			}

			if atomic.LoadInt32(&success) == dataCount {
				break
			}
		}
	}

	for {
		if atomic.LoadInt32(&parllel) == 0 {
			break
		}

		time.Sleep(time.Second)
	}

	if success < dataCount {
		utils.MLogger.Errorf("Download object failed: %s", ErrCannotGetEnoughBlock)
		//  handle channel money problem
		if atomic.LoadInt32(&wrongMoney) > blockCount-dataCount {
			return nil, 0, role.ErrWrongMoney
		}
		return nil, 0, ErrCannotGetEnoughBlock
	}

	do.decoder.Repair = needRepair
	// decode returns bytes of 16B
	data, err := do.decoder.Decode(datas, 0, int(offset+remain))
	if err != nil {
		utils.MLogger.Errorf("Download failed due to decode err: ", err)
		return nil, 0, err
	}

	utils.MLogger.Debugf("Download get length: %d, need %d, from %d", len(data), offset+remain, offset)

	return data, int64(len(data)), nil
}

func (do *downloadTask) getChannelSign(cItem *role.ChannelItem, readLen int) ([]byte, *big.Int, error) {
	hexSK := do.group.privKey
	channelID := do.group.groupID

	if cItem != nil {
		money := big.NewInt(int64(readLen) * utils.READPRICEPERMB / (1024 * 1024))
		money.Add(money, cItem.Value) //100 + valueBase
		if money.Cmp(cItem.Money) > 0 {
			utils.MLogger.Warn("need to redeploy channel contract for ", cItem.ProID)
		}

		channelID = cItem.ChannelID

		mes, err := role.SignForChannel(channelID, hexSK, money)
		if err != nil {
			utils.MLogger.Errorf("Signature about channelID %s fails: %s", channelID, err)
			return nil, nil, err
		}

		return mes, money, nil
	}
	return nil, nil, role.ErrTestUser
}
