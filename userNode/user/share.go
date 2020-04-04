package user

import (
	"context"
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
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
	if !l.online || l.meta.buckets == nil {
		return "", ErrLfsServiceNotReady
	}

	bucket, ok := l.meta.buckets[bucketName]
	if !ok || bucket == nil || bucket.Deletion {
		return "", ErrBucketNotExist
	}

	object, ok := bucket.Objects.Find(MetaName(objectName)).(*ObjectInfo)
	if !ok || object.Deletion {
		return "", ErrObjectNotExist
	}
	sl := &mpb.ShareLink{
		UserID:     l.userID,
		QueryID:    l.fsID,
		BucketName: bucketName,
		ObjectName: objectName,
		BOpts:      bucket.BOpts,
		BucketID:   bucket.BucketID,
		OParts:     make([]*mpb.ObjectPart, object.GetPartCount()),
	}

	for i := 0; i < int(object.GetPartCount()); i++ {
		sl.OParts[i] = object.Parts[i]
	}

	if bucket.BOpts.Encryption == 1 {
		decKey := aes.CreateAesKey([]byte(l.privateKey), []byte(l.fsID), bucket.BucketID, object.GetInfo().GetObjectID())
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

	utils.MLogger.Info("Download Share Object: ", sl.GetObjectName(), " from bucket: ", sl.GetBucketName(), " from user: ", sl.GetUserID())

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

	su, err := u.NewFS(sl.UserID, uid, sl.QueryID, localSk, 0, 0, 0, 0, 0, false, false)
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

	bopt := &mpb.BlockOptions{
		Bopts:   bo,
		Start:   0,
		UserID:  sl.UserID,
		QueryID: sl.QueryID,
	}

	decoder := dataformat.NewDataCoderWithPrefix(sul.keySet, bopt)

	dl := &downloadTask{
		bucketID:     sl.BucketID,
		group:        su.(*LfsInfo).gInfo,
		decoder:      decoder,
		startTime:    time.Now(),
		encrypt:      bo.Encryption,
		writer:       writer,
		completeFunc: completeFuncs,
	}

	if bo.Encryption == 1 {
		copy(dl.sKey[:], sl.DecKey[:32])
	}

	readLen := int64(0)
	pStart := int64(0)
	length := int64(0)
	for i := 0; i < len(sl.GetOParts()); i++ {
		dl.start = sl.OParts[i].GetStart() + pStart
		dl.length = sl.OParts[i].GetLength() - pStart
		if length > 0 && length-readLen < dl.length {
			dl.length = length - readLen
		}
		err := dl.Start(ctx)
		if err != nil {
			return err
		}
		pStart = 0
		readLen += dl.length
		if length > 0 && length >= readLen {
			break
		}
	}

	return nil
}
