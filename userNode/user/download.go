package user

import (
	"bufio"
	"context"
	"crypto/cipher"
	"io"
	"math/big"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/memoio/go-mefs/contracts"
	"github.com/memoio/go-mefs/crypto/aes"
	"github.com/memoio/go-mefs/crypto/pdp"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	"github.com/memoio/go-mefs/role"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	"golang.org/x/sync/semaphore"
)

type downloadTask struct {
	bucketID     int64
	start        int64
	length       int64
	sizeReceived int64
	encrypt      int32
	sKey         [32]byte
	group        *groupInfo            //groupInfo
	decoder      *dataformat.DataCoder //用于解码数据
	startTime    time.Time
	writer       io.Writer
	completeFunc []CompleteFunc //完成任务或出错的通知函数
	cidMaps      sync.Map
}

// GetObject constructs lfs download process
func (l *LfsInfo) GetObject(ctx context.Context, bucketName, objectName string, writer io.Writer, completeFuncs []CompleteFunc, opts DownloadObjectOptions) error {
	//下载需要10资源
	ok := l.Sm.TryAcquire(10)
	if !ok {
		for _, f := range completeFuncs {
			f(ErrResourceUnavailable)
		}
		return ErrResourceUnavailable
	}
	defer l.Sm.Release(10)
	utils.MLogger.Info("Download Object: ", objectName, " from bucket: ", bucketName)
	if !l.Online() || l.meta.buckets == nil {
		for _, f := range completeFuncs {
			f(ErrLfsServiceNotReady)
		}
		return ErrLfsServiceNotReady
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		for _, f := range completeFuncs {
			f(ErrBucketNotExist)
		}
		return ErrBucketNotExist
	}

	objectRes := bucket.Objects.Find(MetaName(objectName))
	if objectRes == nil {
		for _, f := range completeFuncs {
			f(ErrObjectNotExist)
		}
		return ErrObjectNotExist
	}
	object := objectRes.(*ObjectInfo)
	object.RLock()
	defer object.RUnlock()

	if object.Deletion {
		for _, f := range completeFuncs {
			f(ErrObjectNotExist)
		}
		return ErrObjectNotExist
	}

	opStart := opts.Start

	length := opts.Length

	if length <= 0 {
		length = object.GetLength() - opStart
	}

	if opStart+length > object.GetLength() || length <= 0 {
		for _, f := range completeFuncs {
			f(ErrObjectOptionsInvalid)
		}
		return ErrObjectOptionsInvalid
	}

	bo := bucket.BOpts

	bopt := &mpb.BlockOptions{
		Bopts:   bo,
		Start:   0,
		UserID:  l.userID,
		QueryID: l.fsID,
	}

	decoder, err := dataformat.NewDataCoderWithPrefix(l.keySet, bopt)
	if err != nil {
		return err
	}

	dl := &downloadTask{
		bucketID:     bucket.BucketID,
		group:        l.gInfo,
		decoder:      decoder,
		startTime:    time.Now(),
		length:       length,
		writer:       writer,
		completeFunc: completeFuncs,
		encrypt:      bo.Encryption,
	}
	// default AES
	if bo.Encryption == 1 {
		dl.sKey = aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.GetInfo().GetObjectID())
	}

	i := 0
	readLen := int64(0)
	for readLen < length {
		if len(object.GetParts()) <= i {
			for _, f := range completeFuncs {
				f(ErrObjectOptionsInvalid)
			}
			return ErrObjectOptionsInvalid
		}
		dl.start = opStart + object.Parts[i].GetStart()
		dl.length = object.Parts[i].GetLength() - opStart
		if length-readLen < dl.length {
			dl.length = length - readLen
		}
		err := dl.Start(ctx)
		if err != nil {
			for _, f := range completeFuncs {
				f(err)
			}
			return err
		}
		opStart = 0
		readLen += dl.length
		i++
	}

	return nil
}

func (do *downloadTask) Start(ctx context.Context) error {
	dc := int64(do.decoder.Prefix.Bopts.DataCount)
	segStripeSize := int64(do.decoder.Prefix.Bopts.SegmentSize) * dc
	stripeSize := int64(do.decoder.Prefix.Bopts.SegmentCount) * segStripeSize
	readUnit := int64(transNum) * segStripeSize

	var bEnc cipher.BlockMode
	if do.encrypt == 1 {
		tmpEnc, err := aes.ContructAesDec(do.sKey[:])
		if err != nil {
			return err
		}
		bEnc = tmpEnc
	}

	// transNum can be adjusted by change MEFS_TRANS;
	// larger MEFS_TRANS has large rate when network is ok
	tn := os.Getenv("MEFS_TRANS")
	utils.MLogger.Infof("MEFS_TRANS in download is set to %s", tn)
	if tn != "" {
		tNum, err := strconv.Atoi(tn)
		if err != nil {
			transNum = defaultTransNum
		} else {
			transNum = tNum
		}
	} else {
		transNum = defaultTransNum
	}
	utils.MLogger.Debugf("download rate is: %d", transNum)

	var length int64
	breakFlag := false
	for !breakFlag {
		select {
		case <-ctx.Done():
			utils.MLogger.Warn("download cancel")
			do.Complete(nil)
			return nil
		default:
			start := do.start + do.sizeReceived
			// read at most one stripe
			length = stripeSize - start%stripeSize
			// read to end
			if length > do.length-do.sizeReceived {
				length = do.length - do.sizeReceived
			}
			// read slower due to network
			if length > readUnit {
				length = readUnit
			}

			data, n, err := do.rangeRead(ctx, start, length)
			if err != nil {
				if err.Error() == role.ErrWrongMoney.Error() {
					do.group.loadContracts(ctx, "")
					continue
				} else {
					do.Complete(err)
					return err
				}
			}

			if n < length {
				utils.MLogger.Warn("length is not match, got: ", n, ", want: ", length)
			}
			offset := start % segStripeSize

			if do.encrypt == 1 {
				padding := aes.BlockSize - ((offset+length-1)%aes.BlockSize + 1)
				data = data[:offset+length+padding]
				decrypted := make([]byte, len(data))
				bEnc.CryptBlocks(decrypted, data)
				data = decrypted[:len(data)-int(padding)]
			}

			wl, err := do.writer.Write(data[offset : offset+length])
			if err != nil {
				do.Complete(err)
				return err
			}

			if int64(wl) != length {
				utils.MLogger.Warn("write length is not equal")
			}

			do.sizeReceived += length

			if do.sizeReceived >= do.length {
				breakFlag = true
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
func (do *downloadTask) rangeRead(ctx context.Context, start, length int64) ([]byte, int64, error) {
	// bucket options
	dataCount := do.decoder.Prefix.Bopts.DataCount
	parityCount := do.decoder.Prefix.Bopts.ParityCount
	blockCount := dataCount + parityCount
	segSize := do.decoder.Prefix.Bopts.SegmentSize
	segStripeSize := int64(do.decoder.Prefix.Bopts.SegmentSize * dataCount)
	stripeSize := int64(do.decoder.Prefix.Bopts.SegmentCount) * segStripeSize
	tagSize, ok := pdp.TagMap[int(do.decoder.Prefix.Bopts.TagFlag)]
	if !ok {
		tagSize = 48
	}

	// convert start and length to stripe, segment parameters
	curStripe := start / stripeSize
	segStart := int(start % stripeSize / segStripeSize)
	offset := start % segStripeSize
	segNeed := int(1 + (offset+length-1)/segStripeSize)

	do.decoder.Prefix.Start = int32(segStart)
	_, preLen, err := bf.PrefixEncode(do.decoder.Prefix)
	if err != nil {
		return nil, 0, err
	}

	eachLen := preLen + segNeed*(int(segSize)+int(2+(parityCount-1)/dataCount)*tagSize)

	bm, err := metainfo.NewBlockMeta(do.group.groupID, strconv.Itoa(int(do.bucketID)), strconv.Itoa(int(curStripe)), "")
	if err != nil {
		utils.MLogger.Error("Download failed: ", err)
		return nil, 0, err
	}

	needRepair := false //是否需要修复
	datas := make([][]byte, blockCount)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// download dataCount in parallel
	sm := semaphore.NewWeighted(int64(dataCount))
	wg := new(sync.WaitGroup)
	success := int32(0)
	fail := int32(0)
	// record money problem; redo it
	wrongMoney := int32(0)
	for i := 0; i < int(blockCount); i++ {
		// wait here;
		// 只在出错的时候以及下载满datacount后才会release资源
		err = sm.Acquire(ctx, 1)
		if err != nil {
			return nil, 0, err
		}

		// fails too many, no need to download
		if atomic.LoadInt32(&fail) > parityCount {
			utils.MLogger.Error("Download obeject failed too much: ", ErrCannotGetEnoughBlock)
			break
		}

		// enough, no need to download
		if atomic.LoadInt32(&success) >= dataCount {
			break
		}

		// need reapir in decoder
		if i >= int(dataCount) {
			needRepair = true
		}

		bm.SetCid(strconv.Itoa(i))
		ncid := bm.ToString()
		wg.Add(1)
		go func(inum int, chunkid string) {
			defer atomic.AddInt32(&fail, 1)
			defer wg.Done()
			var provider string
			pro, ok := do.cidMaps.Load(chunkid)
			if ok {
				provider = pro.(string)
			} else {
				providerID, _, err := do.group.getBlockProviders(ctx, chunkid)
				if err != nil || providerID == do.group.groupID {
					utils.MLogger.Warnf("Get Block %s 's provider from keeper failed: %s", chunkid, err)
					//在所有出错的地方释放资源，以便新的下载线程
					sm.Release(1)
					return
				}
				provider = providerID
				do.cidMaps.Store(chunkid, providerID)
			}

			pinfo, ok := do.group.providers[provider]
			if !ok {
				utils.MLogger.Warn(provider, " is not my provider")
				sm.Release(1)
				return
			}

			pinfo.Lock()

			//user给channel合约签名，发给provider
			mes, money, err := do.getChannelSign(pinfo, eachLen)
			if err != nil {
				if do.group.userID != do.group.groupID {
					utils.MLogger.Warnf("get channel fails: %s", err)
					pinfo.Unlock()
					sm.Release(1)
					return
				}
			}

			//获取数据块
			bgm, _ := metainfo.NewKey(chunkid, mpb.KeyType_Block, strconv.Itoa(int(segStart)), strconv.Itoa(segNeed))
			b, err := do.group.ds.GetBlock(ctx, bgm.ToString(), mes, provider)
			if err != nil {
				pinfo.Unlock()
				utils.MLogger.Warnf("Get Block %s from %s failed: %s", ncid, provider, err)
				if err.Error() == role.ErrWrongMoney.Error() {
					utils.MLogger.Infof("Try load channel value from %s", provider)
					atomic.AddInt32(&wrongMoney, 1)
				}

				if err.Error() == role.ErrNotEnoughBalance.Error() {
					atomic.AddInt32(&wrongMoney, 1)
					do.group.loadContracts(ctx, provider)
				}
				sm.Release(1)
				return
			}
			blkData := b.RawData()
			ok, err = dataformat.VerifyBlockLength(blkData, segStart, segNeed)
			if !ok || err != nil {
				utils.MLogger.Errorf("Verify Block %s from %s offset unmatched, Err: %s", chunkid, provider, err)
				pinfo.Unlock()
				sm.Release(1)
				return
			}

			_, _, _, ok = do.decoder.VerifyBlock(blkData, chunkid)
			if !ok {
				utils.MLogger.Warn("Fail to verify block: ", chunkid, " from:", provider)
				pinfo.Unlock()
				sm.Release(1)
				return
			}

			//下载数据成功，将内存的channel的value更改
			if pinfo.chanItem != nil {
				pinfo.chanItem.Value = money
				pinfo.chanItem.Sig = mes
				pinfo.chanItem.Dirty = true
				utils.MLogger.Info("Download success, change channel.value: ", pinfo.chanItem.ChannelID, " to: ", money.String())
				pinfo.Unlock()
				key, err := metainfo.NewKey(pinfo.providerID, mpb.KeyType_Channel, pinfo.chanItem.ChannelID)
				if err == nil {
					do.group.ds.PutKey(ctx, key.ToString(), mes, nil, "local")
				}
			} else {
				pinfo.Unlock()
			}

			datas[inum] = blkData
			atomic.AddInt32(&success, 1)
			atomic.AddInt32(&fail, -1)
			// download dataCount; release resource
			if atomic.LoadInt32(&success) >= dataCount {
				sm.Release(1)
			}
		}(i, ncid)
	}

	wg.Wait()

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
	data, err := do.decoder.Decode(datas, 0, int(length))
	if err != nil {
		utils.MLogger.Errorf("Download failed due to decode err: ", err)
		return nil, 0, err
	}

	utils.MLogger.Debugf("Download get length: %d, need %d, from %d", len(data), length, start)

	return data, int64(len(data)), nil
}

func (do *downloadTask) getChannelSign(pInfo *providerInfo, readLen int) ([]byte, *big.Int, error) {
	cItem := pInfo.chanItem
	hexSK := do.group.privKey
	channelID := do.group.groupID

	if cItem != nil {
		readPrice := big.NewInt(utils.READPRICE)
		weiRPrice := new(big.Float).SetInt64(utils.READPRICE)
		weiRPrice.Quo(weiRPrice, contracts.GetMemoPrice())
		weiRPrice.Int(readPrice)

		readPrice.Mul(readPrice, big.NewInt(int64(readLen)))
		readPrice.Quo(readPrice, big.NewInt(1024*1024))

		newValue := new(big.Int).Add(readPrice, cItem.Value)
		if newValue.Cmp(cItem.Money) > 0 {
			utils.MLogger.Infof("need to redeploy channel contract for %s, contract has balance %d, need %d ", cItem.ProID, cItem.Money, newValue)

			oldChanID := cItem.ChannelID

			chanID, err := role.DeployChannel(do.group.shareToID, do.group.groupID, pInfo.providerID, do.group.privKey, do.group.storeDays, do.group.storeSize/int64(do.group.providerSLA), true)
			if err != nil {
				return nil, nil, err
			}

			if chanID == oldChanID {
				utils.MLogger.Infof("channel %s has not changed", cItem.ChannelID)
				return nil, nil, role.ErrEmptyData
			}

			gotItem, err := role.GetChannelInfo(do.group.shareToID, chanID)
			if err != nil {
				return nil, nil, err
			}

			pInfo.chanItem = &gotItem

			cItem = pInfo.chanItem
			newValue = readPrice

			if newValue.Cmp(cItem.Money) > 0 {
				utils.MLogger.Infof("channel %s has money %d, but need %d", cItem.ChannelID, cItem.Money, newValue)
				return nil, nil, role.ErrNotEnoughBalance
			}
		}

		channelID = cItem.ChannelID

		mes, err := role.SignForChannel(channelID, hexSK, newValue)
		if err != nil {
			utils.MLogger.Errorf("Signature about channelID %s fails: %s", channelID, err)
			return nil, nil, err
		}

		return mes, newValue, nil
	}
	return nil, nil, role.ErrTestUser
}
