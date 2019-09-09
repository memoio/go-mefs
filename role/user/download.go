package user

import (
	"bufio"
	"context"
	"crypto/sha256"
	"io"
	"log"
	"math/big"
	"strconv"
	"sync"
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

const DefaultBufSize = 1024 * 1024 * 4

const DefaultJobsCount = 3

type writeFunc func([]byte, int32, error)

type notif struct {
	err error
	id  int32
}

//下载一个stripe的下载任务
type downloadStripe struct {
	jobID          int32 //此任务为所有Stripe任务里的第几个子任务
	bucket         *Bucket
	blockCompleted int                     //下载成功了几个块
	decoder        *dataformat.DataDecoder //用于解码数据
	Lfs            *LfsService             //lfsservice
	Group          *GroupService           //GroupService
	Contract       *ContractService        //合约用于支付
	writeFunc      writeFunc               //用于调度下载，以及返回是否出错
}

func newDownloadStripe(jobID int32, bucket *Bucket, decoder *dataformat.DataDecoder, lfs *LfsService, group *GroupService, contract *ContractService, writeFunc writeFunc) (*downloadStripe, error) {
	return &downloadStripe{
		jobID:     jobID,
		bucket:    bucket,
		decoder:   decoder,
		Lfs:       lfs,
		Group:     group,
		Contract:  contract,
		writeFunc: writeFunc,
	}, nil
}

//下载一整个对象的下载任务
type downloadObject struct {
	Bucket       *Bucket
	LfsService   *LfsService      //lfsService
	Group        *GroupService    //GroupService
	Contract     *ContractService //合约用于支付
	Object       *Object
	decoder      *dataformat.DataDecoder //用于解码数据
	buffer       [][]byte
	State        TaskState
	jobs         []*downloadStripe
	curJob       int32
	jobsCount    int32 //同时进行几个子job
	curStripe    int32 //当前已进行到哪个stripe
	sizeReceived int   //可以统计下载进度
	startTime    int64
	pipeReader   io.Reader
	pipeWriter   io.Writer
	completeFunc []CompleteFunc //完成任务或出错的通知函数
}

func (lfs *LfsService) ConstructDownload(bucketName, objectName string) (Job, io.Reader, error) {
	err := isStart(lfs.UserID)
	if err != nil {
		return nil, nil, err
	}
	if lfs.CurrentLog.BucketNameToID == nil {
		return nil, nil, ErrBucketNotExist
	}

	bucketID, ok := lfs.CurrentLog.BucketNameToID[bucketName]
	if !ok {
		return nil, nil, ErrBucketNotExist
	}
	bucket, ok := lfs.CurrentLog.BucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return nil, nil, ErrBucketNotExist
	}

	object, ok := bucket.Objects[objectName]
	if !ok || object == nil || object.Deletion {
		return nil, nil, ErrObjectNotExist
	}
	piper, pipew := io.Pipe()
	bufw := bufio.NewWriterSize(pipew, DefaultBufSize)
	checkErrAndClosePipe := func(err error) error {
		if err != nil {
			err = pipew.CloseWithError(err)
			return err
		}
		err = pipew.Close()
		return err
	}
	var complete []CompleteFunc
	complete = append(complete, checkErrAndClosePipe)
	contract := GetContractService(lfs.UserID)
	group := GetGroupService(lfs.UserID)

	decoder, _ := dataformat.NewDataDecoder(bucket.Policy, bucket.DataCount, bucket.ParityCount)
	return &downloadObject{
		Bucket:       bucket,
		LfsService:   lfs,
		Group:        group,
		Contract:     contract,
		Object:       object,
		pipeReader:   piper,
		pipeWriter:   bufw,
		completeFunc: complete,
		decoder:      decoder,
		State:        Pending,
		jobsCount:    DefaultJobsCount,
		curJob:       0,
		startTime:    time.Now().Unix(),
	}, piper, nil
}

func (do *downloadObject) Start(ctx context.Context) error {
	dataCount := do.Bucket.DataCount
	do.curStripe = do.Object.StripeStart
	offsetStart := do.Object.OffsetStart
	stripeSize := utils.BlockSize * dataCount
	var lock sync.Mutex

	//下载的第一个stripe前已经有多少数据，等于此文件追加在后面
	extraData := do.Object.OffsetStart * int32(do.Bucket.SegmentSize) * dataCount

	//如果本此下载任务不需要设定的线程数目
	if (do.Object.ObjectSize+extraData-1)/stripeSize+1 < do.jobsCount {
		do.jobsCount = (do.Object.ObjectSize+extraData-1)/stripeSize + 1
	}
	do.jobs = make([]*downloadStripe, do.jobsCount)
	do.buffer = make([][]byte, do.jobsCount)
	notifChan := make(chan notif, do.jobsCount)

	//匿名函数用于调度下载与缓冲，所有的下载都通过匿名函数进行
	//用变量加方法，是不是会更好一点
	write := func(data []byte, job int32, err error) {
		lock.Lock()
		defer lock.Unlock()
		if err != nil {
			notifChan <- notif{err, job}
			return
		}
		//轮到本stripe写数据，直接写
		if do.curJob == job {
			_, err := do.pipeWriter.Write(data)
			if err != nil {
				notifChan <- notif{err, job}
				return
			}
			//通知此Job已写入
			notifChan <- notif{nil, job}
			do.sizeReceived += len(data)
			//将后续先下载的缓冲数据一并写了
			for i := (job + 1) % do.jobsCount; i < do.jobsCount; i = (i + 1) % do.jobsCount {
				if do.buffer[i] == nil || i == job {
					do.curJob = i
					break
				}
				if do.buffer[i] != nil {
					_, err := do.pipeWriter.Write(do.buffer[i])
					if err != nil {
						notifChan <- notif{err, i}
						return
					}
					//通知此Job已写入
					notifChan <- notif{nil, i}
					do.sizeReceived += len(do.buffer[i])
					//数据写完置空
					do.buffer[i] = nil
				}
			}
		} else { //还没轮到写数据，先缓存
			do.buffer[job] = data
		}
	}

	//构造任务并运行
	for i := 0; i < int(do.jobsCount); i++ {
		do.jobs[i], _ = newDownloadStripe(int32(i), do.Bucket, do.decoder, do.LfsService, do.Group, do.Contract, write)
		var remain int32
		if do.Object.OffsetStart == 0 {
			if do.Object.ObjectSize >= (do.curStripe-do.Object.StripeStart+1)*stripeSize {
				//填满一整个stripe
				remain = stripeSize
			} else {
				//最后剩下一部分
				remain = do.Object.ObjectSize % (stripeSize)
			}
		} else {
			if do.Object.ObjectSize <= stripeSize-extraData {
				//第一种情况，处在某一个stripe的中间
				remain = do.Object.ObjectSize
			} else if do.Object.ObjectSize >= (do.curStripe-do.Object.StripeStart+1)*stripeSize-extraData {
				//第二种情况，此stripe填到末尾
				remain = stripeSize - offsetStart*int32(do.Bucket.SegmentSize)*dataCount
			} else {
				//第三种情况，此stripe从零开始，但没填到末尾
				remain = (do.Object.ObjectSize + extraData) % (stripeSize)
			}
		}
		go do.jobs[i].Run(ctx, do.curStripe, offsetStart, remain)
		do.curStripe++
		offsetStart = 0
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case notif := <-notifChan:
			//某个stripe下载出现错误
			if notif.err != nil {
				do.Complete(notif.err)
				return notif.err
			}
			//已经下载结束了
			if do.sizeReceived >= int(do.Object.ObjectSize) {
				//将最后的缓冲刷掉
				do.pipeWriter.(*bufio.Writer).Flush()
				do.Complete(nil)
				return nil
			}
			//如果当前在下载的块已经足够，不需要开新下载，否则为完成的Job开一个新任务
			if do.Object.ObjectSize > (do.curStripe-do.Object.StripeStart)*stripeSize-extraData {
				var remain int32
				if do.Object.OffsetStart == 0 {
					if do.Object.ObjectSize >= (do.curStripe-do.Object.StripeStart+1)*stripeSize {
						//填满一整个stripe
						remain = stripeSize
					} else {
						//最后剩下一部分
						remain = do.Object.ObjectSize % (stripeSize)
					}
				} else {
					if do.Object.ObjectSize <= stripeSize-extraData {
						//第一种情况，处在某一个stripe的中间
						remain = do.Object.ObjectSize
					} else if do.Object.ObjectSize >= (do.curStripe-do.Object.StripeStart+1)*stripeSize-extraData {
						//第二种情况，此stripe填到末尾
						remain = stripeSize - offsetStart*int32(do.Bucket.SegmentSize)*dataCount
					} else {
						//第三种情况，此stripe从零开始，但没填到末尾
						remain = (do.Object.ObjectSize + extraData) % (stripeSize)
					}
				}
				//复用结构体运行
				go do.jobs[notif.id].Run(ctx, do.curStripe, offsetStart, remain)
				do.curStripe++
			}
		}
	}
}

func (do *downloadObject) Stop(ctx context.Context) error {
	return nil
}

func (do *downloadObject) Cancel(ctx context.Context) error {
	return nil
}

func (do *downloadObject) Complete(err error) {
	for _, f := range do.completeFunc {
		f(err)
	}
}

func (do *downloadObject) Info() (interface{}, error) {
	return nil, nil
}

type (
	//日后用于区分是直接写到某路径，还是通过api的
	Destination struct {
		typ    int
		path   string
		writer io.Writer
	}

	// a function type that is called when the download completed.
	CompleteFunc func(error) error
)

func (ds *downloadStripe) Run(ctx context.Context, stripeID, offsetStart, remain int32) {
	//首先设置一些本次stripe下载的基本参数
	dataCount := ds.bucket.DataCount
	parityCount := ds.bucket.ParityCount
	blockCount := dataCount + parityCount
	bm, err := metainfo.NewBlockMeta(ds.Lfs.UserID, strconv.Itoa(int(ds.bucket.BucketID)), strconv.Itoa(int(stripeID)), "")
	if err != nil {
		ds.writeFunc(nil, ds.jobID, err)
		return
	}
	// 构建user的privatekey+bucketid的key，对key进行sha256后作为加密的key
	tmpkey := ds.Lfs.PrivateKey
	tmpkey = append(tmpkey, byte(ds.bucket.BucketID))
	skey := sha256.Sum256(tmpkey)
	cfg, err := localNode.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed, err: ", err)
		ds.writeFunc(nil, ds.jobID, err)
		return
	}
	datas := make([][]byte, blockCount)
	var tempReceiveBlockCount int32
	var needRepair = true //是否需要修复
	var data []byte
	for i := 0; i < int(blockCount); i++ {
		select {
		case <-ctx.Done(): //取消了本次下载
			return
		default:
			bm.SetBid(strconv.Itoa(i))
			ncid := bm.ToString()
			provider, _, err := ds.Group.GetBlockProviders(ncid)
			if err != nil || provider == ds.Lfs.UserID {
				log.Printf("Get Block %s's provider from keeper failed, Err: %v\n", ncid, err)
				continue
			}

			//user给channel合约签名，发给provider
			mes, money, err := ds.getMessage(ncid, provider)
			if err != nil {
				continue
			}

			//获取数据块
			b, err := localNode.Blocks.GetBlockFrom(ctx, provider, ncid, DefaultGetBlockDelay, mes)
			if err != nil {
				log.Printf("Get Block %s from %s failed, Err: %v\n", ncid, provider, err)
				continue
			}
			blkData := b.RawData()
			//需要检查数据块的长度也没问题
			ok, err := dataformat.VerifyBlockLength(blkData, int(offsetStart), int(ds.bucket.TagFlag), int(ds.bucket.SegmentSize), int(dataCount), int(parityCount), int(remain), ds.bucket.Policy)
			if !ok {
				log.Printf("Block %s from %s offset unmatched, Err: %v\n", ncid, provider, err)
				continue
			}

			if ok := dataformat.VerifyBlock(blkData, ncid, ds.Group.GetKeyset().Pk); !ok || err != nil {
				log.Println("Verify Block failed.", ncid, "from:", provider)
				continue
			}

			if !cfg.Test {
				//下载数据成功，将内存的channel的value更改
				cs := ds.Contract
				if cs == nil {
					ds.writeFunc(nil, ds.jobID, err)
					return
				}
				channel, err := cs.GetChannelItem(provider)
				if err != nil {
					ds.writeFunc(nil, ds.jobID, err)
					return
				}
				log.Println("下载成功，更改内存中channel.value", channel.ChannelAddr, money.String())
				channel.Value = money
				cs.channelBook[provider] = channel
			}
			datas[i] = blkData
			tempReceiveBlockCount++
			if tempReceiveBlockCount >= dataCount {
				if i == int(dataCount)-1 {
					needRepair = false
				}
				break
			}
		}
	}
	if tempReceiveBlockCount < dataCount {
		ds.writeFunc(nil, ds.jobID, ErrCannotGetEnoughBlock)
		return
	}
	data, err = ds.decoder.Decode(datas, int(offsetStart), needRepair)
	if err != nil {
		ds.writeFunc(nil, ds.jobID, err)
		return
	}
	if int32(len(data)) > remain {
		padding := aes.BlockSize - ((remain-1)%aes.BlockSize + 1)
		data = data[:remain+padding] //此处时因为获取的为整块，而文件所需只占里面一部分
		// 先解密，再去padding
		data, err = aes.AesDecrypt(data, skey[:])
		if err != nil {
			ds.writeFunc(nil, ds.jobID, err)
			return
		}
		data = data[:len(data)-int(padding)]
		//使用匿名函数调度
		ds.writeFunc(data, ds.jobID, nil)
		return
	}

	data, err = aes.AesDecrypt(data, skey[:])
	if err != nil {
		ds.writeFunc(nil, ds.jobID, err)
		return
	}
	//使用匿名函数调度
	ds.writeFunc(data, ds.jobID, nil)
	return
}

func (ds *downloadStripe) getMessage(ncid string, provider string) ([]byte, *big.Int, error) {
	money := big.NewInt(0)
	//user给channel合约签名，发给provider
	userID := ds.Lfs.UserID
	privateKey := ds.Lfs.PrivateKey
	localAddress, providerAddress, hexSK, err := buildSignParams(userID, provider, privateKey)
	if err != nil {
		log.Printf("buildSignParams about Block %s from %s failed.\n", ncid, provider)
		return nil, nil, err
	}

	var channelAddr common.Address

	//判断是不是测试user，如果是，就将channelAddress设为0
	node := localNode
	cfg, err := node.Repo.Config()
	if err != nil {
		log.Println("get config from Download failed.")
		return nil, nil, err
	}
	if cfg.Test {
		channelAddress := contracts.InvalidAddr
		channelAddr = common.HexToAddress(channelAddress)
		//设置此次下载需要签名的金额，money此时不用变仍为0
	} else {
		cs := ds.Contract
		if cs == nil {
			return nil, nil, ErrUserNotExist
		}
		Item, err := cs.GetChannelItem(provider)
		if err != nil {
			return nil, nil, err
		}
		channelAddr = common.HexToAddress(Item.ChannelAddr)
		// 此次下载需要签名的金额，在valueBase的基础上再加上此次下载需要支付的money，就是此次签名的value
		addValue := int64((utils.BlockSize / (1024 * 1024)) * utils.READPRICEPERMB)
		money = money.Add(Item.Value, big.NewInt(addValue)) //100 + valueBase
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
