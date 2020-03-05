package miniogw

import (
	"context"
	"errors"
	"io"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/memoio/go-mefs/core"
	"github.com/memoio/go-mefs/role/user"

	minio "github.com/minio/minio/cmd"
	"github.com/minio/minio/pkg/hash"
)

var (
	errAbort               = errors.New("MultipartUpload abort")
	errAbortByAnother      = errors.New("aborted by another upload to the same location")
	errUploadMissing       = errors.New("pending upload missing")
	errMismatch            = errors.New("pending upload bucket/object name mismatch")
	errPartAlreadyUploaded = errors.New("part already uploaded")
)

func (l *lfsGateway) NewMultipartUpload(ctx context.Context, bucket, object string, options minio.ObjectOptions) (uploadID string, err error) {
	lfs := core.LocalNode.Inst.(*user.Info).GetUser(l.userID)
	if lfs == nil || !lfs.Online() {
		return "", errLfsServiceNotReady
	}

	ctx, cancel := context.WithCancel(ctx)
	_, err = lfs.HeadBucket(ctx, bucket)
	if err != nil {
		cancel()
		return "", err
	}
	uploads := l.multipart

	upload, err := uploads.Create(bucket, object, cancel)
	if err != nil {
		return "", err
	}
	obj, err := lfs.PutObject(ctx, bucket, object, upload.Stream)
	if err != nil {
		return "", err
	}

	uploads.RemoveByID(upload.ID)
	if err != nil {
		upload.fail(err)
	} else {
		upload.complete(minio.ObjectInfo{
			Bucket:      bucket,
			Name:        object,
			IsDir:       obj.GetInfo().GetDir(),
			ETag:        obj.Parts[0].GetETag(),
			ContentType: obj.GetInfo().GetContentType(),
			Size:        obj.GetLength(),
		})
	}

	return upload.ID, nil
}

func (l *lfsGateway) PutObjectPart(ctx context.Context, bucket, object, uploadID string, partID int, data *minio.PutObjReader, options minio.ObjectOptions) (info minio.PartInfo, err error) {
	uploads := l.multipart
	upload, err := uploads.Get(bucket, object, uploadID)
	if err != nil {
		return minio.PartInfo{}, err
	}

	part, err := upload.Stream.AddPart(partID, data.Reader)
	if err != nil {
		return minio.PartInfo{}, err
	}

	err = <-part.Done
	if err != nil {
		return minio.PartInfo{}, err
	}

	partInfo := minio.PartInfo{
		PartNumber:   part.Number,
		LastModified: time.Now(),
		ETag:         data.SHA256HexString(),
		Size:         atomic.LoadInt64(&part.Size),
	}

	upload.addCompletedPart(partInfo)

	return partInfo, nil
}

func (l *lfsGateway) AbortMultipartUpload(ctx context.Context, bucket, object, uploadID string) (err error) {

	uploads := l.multipart

	upload, err := uploads.Remove(bucket, object, uploadID)
	if err != nil {
		return err
	}

	upload.Cancel()
	upload.Stream.Abort(errAbort)
	r := <-upload.Done
	if r.Error != errAbort {
		return r.Error
	}
	return nil
}

func (l *lfsGateway) CompleteMultipartUpload(ctx context.Context, bucket, object, uploadID string, uploadedParts []minio.CompletePart, options minio.ObjectOptions) (objInfo minio.ObjectInfo, err error) {
	uploads := l.multipart
	upload, err := uploads.Remove(bucket, object, uploadID)
	if err != nil {
		return minio.ObjectInfo{}, err
	}

	// notify stream that there aren't more parts coming
	upload.Stream.Close()
	// wait for completion
	result := <-upload.Done
	// return the final info
	return result.Info, result.Error
}

func (l *lfsGateway) ListObjectParts(ctx context.Context, bucket, object, uploadID string, partNumberMarker int, maxParts int, options minio.ObjectOptions) (result minio.ListPartsInfo, err error) {
	uploads := l.multipart
	upload, err := uploads.Get(bucket, object, uploadID)
	if err != nil {
		return minio.ListPartsInfo{}, err
	}

	list := minio.ListPartsInfo{}

	list.Bucket = bucket
	list.Object = object
	list.UploadID = uploadID
	list.PartNumberMarker = partNumberMarker
	list.MaxParts = maxParts
	list.Parts = upload.getCompletedParts()

	sort.Slice(list.Parts, func(i, k int) bool {
		return list.Parts[i].PartNumber < list.Parts[k].PartNumber
	})

	var first int
	for i, p := range list.Parts {
		first = i
		if partNumberMarker <= p.PartNumber {
			break
		}
	}

	list.Parts = list.Parts[first:]
	if len(list.Parts) > maxParts {
		list.NextPartNumberMarker = list.Parts[maxParts].PartNumber
		list.Parts = list.Parts[:maxParts]
		list.IsTruncated = true
	}

	return list, nil
}

// TODO: implement
// ListMultipartUploads lists all multipart uploads.
func (l *lfsGateway) ListMultipartUploads(ctx context.Context, bucket, prefix, keyMarker, uploadIDMarker, delimiter string, maxUploads int) (lmi minio.ListMultipartsInfo, err error) {
	return lmi, nil
}

// CopyObjectPart creates a part in a multipart upload by copying
// existing object or a part of it.
func (l *lfsGateway) CopyObjectPart(ctx context.Context, srcBucket, srcObject, destBucket, destObject, uploadID string,
	partID int, startOffset, length int64, srcInfo minio.ObjectInfo, srcOpts, dstOpts minio.ObjectOptions) (p minio.PartInfo, err error) {
	return p, nil
}

// MultipartUploads manages pending multipart uploads
type MultipartUploads struct {
	mu      sync.RWMutex
	lastID  int
	pending map[string]*MultipartUpload
}

// NewMultipartUploads creates new MultipartUploads
func NewMultipartUploads() *MultipartUploads {
	return &MultipartUploads{
		pending: make(map[string]*MultipartUpload),
	}
}

// Create creates a new upload
func (uploads *MultipartUploads) Create(bucket, object string, cancel context.CancelFunc) (*MultipartUpload, error) {
	uploads.mu.Lock()
	defer uploads.mu.Unlock()

	for id, upload := range uploads.pending {
		if upload.Bucket == bucket && upload.Object == object {
			upload.Stream.Abort(errAbortByAnother)
			delete(uploads.pending, id)
		}
	}

	uploads.lastID++
	uploadID := "Upload" + strconv.Itoa(uploads.lastID)
	upload := NewMultipartUpload(uploadID, bucket, object, cancel)
	uploads.pending[uploadID] = upload
	return upload, nil
}

// Get finds a pending upload
func (uploads *MultipartUploads) Get(bucket, object, uploadID string) (*MultipartUpload, error) {
	uploads.mu.Lock()
	defer uploads.mu.Unlock()

	upload, ok := uploads.pending[uploadID]
	if !ok {
		return nil, errUploadMissing
	}
	if upload.Bucket != bucket || upload.Object != object {
		return nil, errMismatch
	}

	return upload, nil
}

// Remove returns and removes a pending upload
func (uploads *MultipartUploads) Remove(bucket, object, uploadID string) (*MultipartUpload, error) {
	uploads.mu.RLock()
	defer uploads.mu.RUnlock()

	upload, ok := uploads.pending[uploadID]
	if !ok {
		return nil, errUploadMissing
	}
	if upload.Bucket != bucket || upload.Object != object {
		return nil, errMismatch
	}

	delete(uploads.pending, uploadID)

	return upload, nil
}

// RemoveByID removes pending upload by id
func (uploads *MultipartUploads) RemoveByID(uploadID string) {
	uploads.mu.RLock()
	defer uploads.mu.RUnlock()
	delete(uploads.pending, uploadID)
}

// MultipartUpload is partial info about a pending upload
type MultipartUpload struct {
	ID     string
	Bucket string
	Object string
	Cancel context.CancelFunc
	// Metadata map[string]string
	Done   chan (*MultipartUploadResult)
	Stream *MultipartStream

	mu        sync.Mutex
	completed []minio.PartInfo
}

// MultipartUploadResult contains either an Error or the uploaded ObjectInfo
type MultipartUploadResult struct {
	Error error
	Info  minio.ObjectInfo
}

// NewMultipartUpload creates a new MultipartUpload
func NewMultipartUpload(uploadID string, bucket, object string, cancel context.CancelFunc) *MultipartUpload {
	upload := &MultipartUpload{
		ID:     uploadID,
		Bucket: bucket,
		Object: object,
		Cancel: cancel,
		Done:   make(chan *MultipartUploadResult, 1),
		Stream: NewMultipartStream(),
	}
	return upload
}

// addCompletedPart adds a completed part to the list
func (upload *MultipartUpload) addCompletedPart(part minio.PartInfo) {
	upload.mu.Lock()
	defer upload.mu.Unlock()

	upload.completed = append(upload.completed, part)
}

func (upload *MultipartUpload) getCompletedParts() []minio.PartInfo {
	upload.mu.Lock()
	defer upload.mu.Unlock()

	return append([]minio.PartInfo{}, upload.completed...)
}

// fail aborts the upload with an error
func (upload *MultipartUpload) fail(err error) {
	upload.Done <- &MultipartUploadResult{Error: err}
	close(upload.Done)
}

// complete completes the upload
func (upload *MultipartUpload) complete(info minio.ObjectInfo) {
	upload.Done <- &MultipartUploadResult{Info: info}
	close(upload.Done)
}

// MultipartStream serializes multiple readers into a single reader
type MultipartStream struct {
	mu          sync.Mutex
	moreParts   sync.Cond
	err         error
	closed      bool
	finished    bool
	nextID      int
	nextNumber  int
	currentPart *StreamPart
	parts       []*StreamPart
}

// StreamPart is a reader waiting in MultipartStream
type StreamPart struct {
	Number int
	ID     int
	Size   int64
	Reader *hash.Reader
	Done   chan error
}

// NewMultipartStream creates a new MultipartStream
func NewMultipartStream() *MultipartStream {
	stream := &MultipartStream{}
	stream.moreParts.L = &stream.mu
	stream.nextID = 1
	return stream
}

// Abort aborts the stream reading
func (stream *MultipartStream) Abort(err error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.finished {
		return
	}

	if stream.err == nil {
		stream.err = err
	}
	stream.finished = true
	stream.closed = true

	for _, part := range stream.parts {
		part.Done <- err
		close(part.Done)
	}
	stream.parts = nil

	stream.moreParts.Broadcast()
}

// Close closes the stream, but lets it complete
func (stream *MultipartStream) Close() {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	stream.closed = true
	stream.moreParts.Broadcast()
}

// Read implements io.Reader interface, blocking when there's no part
func (stream *MultipartStream) Read(data []byte) (n int, err error) {
	stream.mu.Lock()
	for {
		// has an error occurred?
		if stream.err != nil {
			stream.mu.Unlock()
			return 0, stream.err
		}
		// still uploading the current part?
		if stream.currentPart != nil {
			break
		}
		// do we have the next part?
		if len(stream.parts) > 0 && stream.nextID == stream.parts[0].ID {
			stream.currentPart = stream.parts[0]
			stream.parts = stream.parts[1:]
			stream.nextID++
			break
		}
		// we don't have the next part and are closed, hence we are complete
		if stream.closed {
			stream.finished = true
			stream.mu.Unlock()
			return 0, io.EOF
		}

		stream.moreParts.Wait()
	}
	stream.mu.Unlock()

	// read as much as we can
	n, err = stream.currentPart.Reader.Read(data)
	atomic.AddInt64(&stream.currentPart.Size, int64(n))
	if err == io.EOF {
		// the part completed, hence advance to the next one
		err = nil
		close(stream.currentPart.Done)
		stream.currentPart = nil
	} else if err != nil {
		// something bad happened, abort the whole thing
		stream.Abort(err)
		return n, err
	}

	return n, err
}

// AddPart adds a new part to the stream to wait
func (stream *MultipartStream) AddPart(partID int, data *hash.Reader) (*StreamPart, error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if partID < stream.nextID {
		return nil, errPartAlreadyUploaded
	}

	for _, p := range stream.parts {
		if p.ID == partID {
			// Replace the reader of this part with the new one.
			// This could happen if the read timeout for this part has expired
			// and the client tries to upload the part again.
			p.Reader = data
			return p, nil
		}
	}

	stream.nextNumber++
	part := &StreamPart{
		Number: stream.nextNumber - 1,
		ID:     partID,
		Size:   0,
		Reader: data,
		Done:   make(chan error, 1),
	}

	stream.parts = append(stream.parts, part)
	sort.Slice(stream.parts, func(i, k int) bool {
		return stream.parts[i].ID < stream.parts[k].ID
	})

	stream.moreParts.Broadcast()

	return part, nil
}
