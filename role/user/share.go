package user

import (
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/memoio/go-mefs/crypto/aes"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/proto"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	b58 "github.com/mr-tron/base58/base58"
)

// GenShareObject constructs sharelink
func (l *LfsInfo) GenShareObject(ctx context.Context, bucketName, objectName string) (string, error) {
	utils.MLogger.Info("Download Object: ", objectName, " from bucket: ", bucketName)
	if !l.online || l.meta.bucketNameToID == nil {
		return "", ErrLfsServiceNotReady
	}

	bucketID, ok := l.meta.bucketNameToID[bucketName]
	if !ok {
		return "", ErrBucketNotExist
	}

	bucket, ok := l.meta.bucketByID[bucketID]
	if !ok || bucket == nil || bucket.Deletion {
		return "", ErrBucketNotExist
	}

	objectElement, ok := bucket.objects[objectName]
	if !ok {
		return "", ErrObjectNotExist
	}

	object, ok := objectElement.Value.(*objectInfo)
	if !ok || object == nil || object.Deletion {
		return "", ErrObjectNotExist
	}

	sl := &mpb.ShareLink{
		UserID:     l.userID,
		QueryID:    l.fsID,
		BucketName: bucketName,
		ObjectName: objectName,
		BOpts:      bucket.BOpts,
		BucketID:   bucket.BucketID,
		OParts:     make([]*mpb.ObjectPart, 1),
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
		decKey := aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.OPart.Start)
		sl.DecKey = decKey[:]
	}

	if l.fsID == l.userID {
		kmUser, err := metainfo.NewKey(l.fsID, mpb.KeyType_LFS, l.userID)
		if err != nil {
			return "", err
		}

		res, err := l.ds.GetKey(ctx, kmUser.ToString(), "local")
		if err != nil {
			return "", err
		}
		sl.KPs = string(res)
	}

	sByte, err := proto.Marshal(sl)
	if err != nil {
		return "", err
	}

	return b58.Encode(sByte), nil
}

// GetShareObject constructs lfs download process
func (u *Info) GetShareObject(ctx context.Context, writer io.Writer, completeFuncs []CompleteFunc, uid, localSk string, share string) error {
	utils.MLogger.Debug("Download Share Object")
	shareByte, err := b58.Decode(share)
	if err != nil {
		utils.MLogger.Warn("Download Share Object B58 decode failed: ", err)
		return err
	}

	sl := new(mpb.ShareLink)
	err = proto.Unmarshal(shareByte, sl)
	if err != nil {
		utils.MLogger.Warn("Download Share Object Unmarshal failed: ", err)
		return err
	}

	utils.MLogger.Info("Download Share Object: ", sl.GetObjectName(), " from bucket: ", sl.GetObjectName(), " from user: ", sl.GetUserID())

	if sl.UserID == sl.QueryID {
		kmUser, err := metainfo.NewKey(sl.QueryID, mpb.KeyType_LFS, sl.UserID)
		if err != nil {
			return err
		}

		err = u.ds.PutKey(ctx, kmUser.ToString(), []byte(sl.KPs), nil, "local")
		if err != nil {
			return err
		}
	}

	su, err := u.NewFS(sl.UserID, uid, sl.QueryID, localSk, 0, 0, 0, 0, 0, false)
	if err != nil {
		utils.MLogger.Errorf("create share user %s error: %s", sl.UserID, err)
		return err
	}

	err = su.Start(ctx)
	if err != nil {
		utils.MLogger.Errorf("share user %s started error: %s", sl.UserID, err)
		return err
	}

	sul := su.(*LfsInfo)
	sul.writable = false
	sul.privateKey = localSk

	bo := sl.BOpts
	decoder := dataformat.NewDataCoderWithBopts(bo, sul.keySet)
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
