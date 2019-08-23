// Package delayed wraps a datastore allowing to artificially
// delay all operations.
package delayed

import (
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	delay "github.com/ipfs/go-ipfs-delay"
)

// New returns a new delayed datastore.
func New(ds ds.Datastore, delay delay.D) ds.Datastore {
	return &delayed{ds: ds, delay: delay}
}

type delayed struct {
	ds    ds.Datastore
	delay delay.D
}

func (dds *delayed) Put(key ds.Key, value []byte) (err error) {
	dds.delay.Wait()
	return dds.ds.Put(key, value)
}

func (dds *delayed) Append(key ds.Key, value []byte, beginoffset, endoffset int) (err error) {
	dds.delay.Wait()
	return dds.ds.Append(key, value, beginoffset, endoffset)
}

func (dds *delayed) Get(key ds.Key) (value []byte, err error) {
	dds.delay.Wait()
	return dds.ds.Get(key)
}

func (dds *delayed) GetSegAndTag(key ds.Key, offset uint64) (segment []byte, tag []byte, err error) {
	dds.delay.Wait()
	return dds.ds.GetSegAndTag(key, offset)
}

func (dds *delayed) Has(key ds.Key) (exists bool, err error) {
	dds.delay.Wait()
	return dds.ds.Has(key)
}

func (dds *delayed) GetSize(key ds.Key) (size int, err error) {
	dds.delay.Wait()
	return dds.ds.GetSize(key)
}

func (dds *delayed) Delete(key ds.Key) (err error) {
	dds.delay.Wait()
	return dds.ds.Delete(key)
}

func (dds *delayed) Query(q dsq.Query) (dsq.Results, error) {
	dds.delay.Wait()
	return dds.ds.Query(q)
}

func (dds *delayed) Batch() (ds.Batch, error) {
	return ds.NewBasicBatch(dds), nil
}

func (dds *delayed) DiskUsage() (uint64, error) {
	dds.delay.Wait()
	return ds.DiskUsage(dds.ds)
}
