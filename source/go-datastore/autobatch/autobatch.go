// Package autobatch provides a go-datastore implementation that
// automatically batches together writes by holding puts in memory until
// a certain threshold is met.
package autobatch

import (
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
)

// Datastore implements a go-datastore.
type Datastore struct {
	child ds.Batching

	// TODO: discuss making ds.Batch implement the full ds.Datastore interface
	buffer           map[ds.Key]op
	maxBufferEntries int
}

type op struct {
	delete bool
	begin  int
	length int
	value  []byte
}

// NewAutoBatching returns a new datastore that automatically
// batches writes using the given Batching datastore. The size
// of the memory pool is given by size.
func NewAutoBatching(d ds.Batching, size int) *Datastore {
	return &Datastore{
		child:            d,
		buffer:           make(map[ds.Key]op, size),
		maxBufferEntries: size,
	}
}

// Delete deletes a key/value
func (d *Datastore) Delete(k ds.Key) error {
	d.buffer[k] = op{delete: true}
	if len(d.buffer) > d.maxBufferEntries {
		return d.Flush()
	}
	return nil
}

// Get retrieves a value given a key.
func (d *Datastore) Get(k ds.Key) ([]byte, error) {
	o, ok := d.buffer[k]
	if ok {
		if o.delete {
			return nil, ds.ErrNotFound
		}
		return o.value, nil
	}

	return d.child.Get(k)
}

// GetSegAndTag get
func (d *Datastore) GetSegAndTag(k ds.Key, offset uint64) ([]byte, []byte, error) {
	return d.child.GetSegAndTag(k, offset)
}

// Put stores a key/value.
func (d *Datastore) Put(k ds.Key, val []byte) error {
	d.buffer[k] = op{value: val}
	if len(d.buffer) > d.maxBufferEntries {
		return d.Flush()
	}
	return nil
}

// Append appends block
func (d *Datastore) Append(k ds.Key, val []byte, beginoffset, length int) error {
	d.Flush()

	d.buffer[k] = op{
		value:  val,
		begin:  beginoffset,
		length: length,
	}

	if len(d.buffer) > d.maxBufferEntries {
		return d.Flush()
	}

	return nil
}

// Sync flushes all operations on keys at or under the prefix
// from the current batch to the underlying datastore
func (d *Datastore) Sync(prefix ds.Key) error {
	b, err := d.child.Batch()
	if err != nil {
		return err
	}

	for k, o := range d.buffer {
		if !(k.Equal(prefix) || k.IsDescendantOf(prefix)) {
			continue
		}

		var err error
		if o.delete {
			err = b.Delete(k)
		} else {
			if o.begin == 0 && o.length == 0 {
				err = b.Put(k, o.value)
			} else {
				err = b.Append(k, o.value, o.begin, o.length)
			}
		}
		if err != nil {
			return err
		}

		delete(d.buffer, k)
	}

	return b.Commit()
}

// Flush flushes the current batch to the underlying datastore.
func (d *Datastore) Flush() error {
	b, err := d.child.Batch()
	if err != nil {
		return err
	}

	for k, o := range d.buffer {
		var err error
		if o.delete {
			err = b.Delete(k)
		} else {
			if o.begin == 0 && o.length == 0 {
				err = b.Put(k, o.value)
			} else {
				err = b.Append(k, o.value, o.begin, o.length)
			}

		}
		if err != nil {
			return err
		}
	}
	// clear out buffer
	d.buffer = make(map[ds.Key]op, d.maxBufferEntries)

	return b.Commit()
}

// Has checks if a key is stored.
func (d *Datastore) Has(k ds.Key) (bool, error) {
	o, ok := d.buffer[k]
	if ok {
		return !o.delete, nil
	}

	return d.child.Has(k)
}

// GetSize implements Datastore.GetSize
func (d *Datastore) GetSize(k ds.Key) (int, error) {
	o, ok := d.buffer[k]
	if ok {
		if o.delete {
			return -1, ds.ErrNotFound
		}
		return len(o.value), nil
	}

	return d.child.GetSize(k)
}

// Query performs a query
func (d *Datastore) Query(q dsq.Query) (dsq.Results, error) {
	err := d.Flush()
	if err != nil {
		return nil, err
	}

	return d.child.Query(q)
}

// DiskUsage implements the PersistentDatastore interface.
func (d *Datastore) DiskUsage() (uint64, error) {
	return ds.DiskUsage(d.child)
}

func (d *Datastore) Close() error {
	err1 := d.Flush()
	err2 := d.child.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}
