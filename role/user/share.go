package user

import (
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	dataformat "github.com/memoio/go-mefs/data-format"
	pb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
)

// GetShareObject constructs lfs download process
func (l *LfsInfo) GetShareObject(ctx context.Context, writer io.Writer, completeFuncs []CompleteFunc, share []byte) error {

	sl := new(pb.ShareLink)

	err := proto.Unmarshal(share, sl)
	if err != nil {
		return err
	}

	utils.MLogger.Info("Download Share Object: ", sl.GetObjectName(), " from bucket: ", sl.GetObjectName(), " from user: ", sl.GetUserID())

	bucket := sl.BOptions

	stripeSize := int64(utils.BlockSize * bucket.GetDataCount())
	segStripeSize := int64(bucket.GetSegmentSize()) * int64(bucket.GetDataCount())
	// 下载的开始条带
	stripePos := sl.Start / stripeSize
	// 下载开始的segment
	segPos := (sl.Start % stripeSize) / segStripeSize
	// segment的偏移
	offsetPos := sl.Start % segStripeSize

	decoder := dataformat.NewDataCoder(int(bucket.Policy), int(bucket.DataCount), dataformat.CurrentVersion, int(bucket.ParityCount), int(bucket.TagFlag), int(bucket.SegmentSize), dataformat.DefaultSegmentCount, l.keySet)

	gi := &groupInfo{}

	dl := &downloadTask{
		fsID:         sl.QueryID,
		bucketID:     sl.BucketID,
		group:        gi,
		decoder:      decoder,
		state:        Pending,
		startTime:    time.Now().Unix(),
		curStripe:    stripePos,
		segOffset:    segPos,
		dStart:       offsetPos,
		dLength:      sl.Length,
		encrypt:      sl.Encryption,
		writer:       writer,
		completeFunc: completeFuncs,
	}

	if sl.Encryption {
		copy(dl.sKey[:32], sl.DecKey[:32])
	}

	return dl.Start(ctx)
}
