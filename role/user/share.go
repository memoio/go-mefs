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

// GenShareObject constructs sharelink
func (l *LfsInfo) GenShareObject(ctx context.Context, bucketName, objectName string) ([]byte, error) {
	utils.MLogger.Info("Download Object: ", objectName, " from bucket: ", bucketName)
	if !l.online || l.meta.bucketNameToID == nil {
		return nil, ErrLfsServiceNotReady
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

	sl := &pb.ShareLink{
		UserID:     l.userID,
		QueryID:    l.fsID,
		BucketName: bucketName,
		ObjectName: objectName,
		BOpts:      bucket.BOpts,
		BucketID:   bucket.BucketID,
		OParts:     make([]*pb.ObjectPart, 1),
	}

	opart := object.GetOPart()
	sl.OParts[0] = opart
	for opart.GetNextPart() != "" {
		objectElement, ok := bucket.objects[opart.GetNextPart()]
		if !ok {
			break
		}

		object, ok = objectElement.Value.(*objectInfo)
		if !ok || object == nil || object.Deletion {
			break
		}
		opart = object.GetOPart()
		sl.OParts = append(sl.OParts, opart)
	}

	if bucket.BOpts.Encryption == 1 {
		decKey := CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.OPart.Start)
		sl.DecKey = decKey[:]
	}

	sByte, err := proto.Marshal(sl)
	if err != nil {
		return nil, err
	}

	return sByte, nil
}

// GetShareObject constructs lfs download process
func (u *Info) GetShareObject(ctx context.Context, writer io.Writer, completeFuncs []CompleteFunc, localKey string, share []byte) error {

	sl := new(pb.ShareLink)

	err := proto.Unmarshal(share, sl)
	if err != nil {
		return err
	}

	utils.MLogger.Info("Download Share Object: ", sl.GetObjectName(), " from bucket: ", sl.GetObjectName(), " from user: ", sl.GetUserID())

	su, err := u.NewFS(sl.UserID, sl.QueryID, "", 0, 0, 0, 0, 0, false)
	if err != nil {
		utils.MLogger.Errorf("create share user %s error: %s", sl.UserID, err)
		return err
	}

	err = su.Start(ctx)
	if err != nil {
		utils.MLogger.Errorf("share user %s started error: %s", sl.UserID, err)
		return err
	}

	bo := sl.BOpts

	su.(*LfsInfo).privateKey = localKey

	decoder := dataformat.NewDataCoderWithBopts(bo, su.(*LfsInfo).keySet)
	segSize := int64(bo.GetSegmentSize())
	stripeSize := int64(bo.SegmentCount * bo.SegmentSize * bo.GetDataCount())

	for i := 0; i < len(sl.GetOParts()); i++ {
		opart := sl.GetOParts()[i]
		// 下载的开始条带
		stripePos := opart.Start / stripeSize
		// 下载开始的segment
		segPos := (opart.Start % stripeSize) / segSize
		// segment的偏移
		offsetPos := opart.Start % segSize

		dl := &downloadTask{
			fsID:         sl.QueryID,
			bucketID:     sl.BucketID,
			group:        su.(*LfsInfo).gInfo,
			decoder:      decoder,
			state:        Pending,
			startTime:    time.Now(),
			curStripe:    stripePos,
			segOffset:    segPos,
			dStart:       offsetPos,
			dLength:      opart.Length,
			encrypt:      bo.Encryption,
			writer:       writer,
			completeFunc: completeFuncs,
		}

		if bo.Encryption == 1 {
			copy(dl.sKey[:], sl.DecKey[:32])
		}

		err := dl.Start(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
