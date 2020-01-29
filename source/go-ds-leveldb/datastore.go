package leveldb

import (
	"os"
	"path/filepath"

	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	"github.com/memoio/go-mefs/source/goleveldb/leveldb"
	"github.com/memoio/go-mefs/source/goleveldb/leveldb/errors"
	"github.com/memoio/go-mefs/source/goleveldb/leveldb/iterator"
	"github.com/memoio/go-mefs/source/goleveldb/leveldb/opt"
	"github.com/memoio/go-mefs/source/goleveldb/leveldb/storage"
	"github.com/memoio/go-mefs/source/goleveldb/leveldb/util"
)

type Datastore struct {
	*accessor
	DB   *leveldb.DB
	path string
}

var _ ds.Datastore = (*Datastore)(nil)
var _ ds.TxnDatastore = (*Datastore)(nil)

// Options is an alias of syndtr/goleveldb/opt.Options which might be extended
// in the future.
type Options opt.Options

// NewDatastore returns a new datastore backed by leveldb
//
// for path == "", an in memory bachend will be chosen
func NewDatastore(path string, opts *Options) (*Datastore, error) {
	var nopts opt.Options
	if opts != nil {
		nopts = opt.Options(*opts)
	}

	var err error
	var db *leveldb.DB

	if path == "" {
		db, err = leveldb.Open(storage.NewMemStorage(), &nopts)
	} else {
		db, err = leveldb.OpenFile(path, &nopts)
		if errors.IsCorrupted(err) && !nopts.GetReadOnly() {
			db, err = leveldb.RecoverFile(path, &nopts)
		}
	}

	if err != nil {
		return nil, err
	}

	return &Datastore{
		accessor: &accessor{ldb: db, syncWrites: true},
		DB:       db,
		path:     path,
	}, nil
}

// An extraction of the common interface between LevelDB Transactions and the DB itself.
//
// It allows to plug in either inside the `accessor`.
type levelDbOps interface {
	Put(key, value []byte, wo *opt.WriteOptions) error
	Get(key []byte, ro *opt.ReadOptions) (value []byte, err error)
	Has(key []byte, ro *opt.ReadOptions) (ret bool, err error)
	Delete(key []byte, wo *opt.WriteOptions) error
	NewIterator(slice *util.Range, ro *opt.ReadOptions) iterator.Iterator
}

// Datastore operations using either the DB or a transaction as the backend.
type accessor struct {
	ldb        levelDbOps
	syncWrites bool
}

func (a *accessor) Put(key ds.Key, value []byte) (err error) {
	return a.ldb.Put(key.Bytes(), value, &opt.WriteOptions{Sync: a.syncWrites})
}

func (a *accessor) Append(key ds.Key, value []byte, begin, length int) (err error) {
	return nil
}

func (a *accessor) Sync(prefix ds.Key) error {
	return nil
}

func (a *accessor) Get(key ds.Key) (value []byte, err error) {
	val, err := a.ldb.Get(key.Bytes(), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, ds.ErrNotFound
		}
		return nil, err
	}
	return val, nil
}

func (a *accessor) GetSegAndTag(key ds.Key, offset uint64) (segment []byte, tag []byte, err error) {
	return nil, nil, nil
}

func (a *accessor) Has(key ds.Key) (exists bool, err error) {
	return a.ldb.Has(key.Bytes(), nil)
}

func (d *accessor) GetSize(key ds.Key) (size int, err error) {
	return ds.GetBackedSize(d, key)
}

func (a *accessor) Delete(key ds.Key) (err error) {
	return a.ldb.Delete(key.Bytes(), &opt.WriteOptions{Sync: a.syncWrites})
}

func (a *accessor) Query(q dsq.Query) (dsq.Results, error) {
	var rnge *util.Range

	// make a copy of the query for the fallback naive query implementation.
	// don't modify the original so res.Query() returns the correct results.
	qNaive := q
	if q.Prefix != "" {
		rnge = util.BytesPrefix([]byte(q.Prefix))
		qNaive.Prefix = ""
	}
	i := a.ldb.NewIterator(rnge, nil)
	next := i.Next
	if len(q.Orders) > 0 {
		switch q.Orders[0].(type) {
		case dsq.OrderByKey, *dsq.OrderByKey:
			qNaive.Orders = nil
		case dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			next = func() bool {
				next = i.Prev
				return i.Last()
			}
			qNaive.Orders = nil
		default:
		}
	}
	r := dsq.ResultsFromIterator(q, dsq.Iterator{
		Next: func() (dsq.Result, bool) {
			if !next() {
				return dsq.Result{}, false
			}
			k := string(i.Key())
			e := dsq.Entry{Key: k, Size: len(i.Value())}

			if !q.KeysOnly {
				buf := make([]byte, len(i.Value()))
				copy(buf, i.Value())
				e.Value = buf
			}
			return dsq.Result{Entry: e}, true
		},
		Close: func() error {
			i.Release()
			return nil
		},
	})
	return dsq.NaiveQueryApply(qNaive, r), nil
}

// DiskUsage returns the current disk size used by this levelDB.
// For in-mem datastores, it will return 0.
func (d *Datastore) DiskUsage() (uint64, error) {
	if d.path == "" { // in-mem
		return 0, nil
	}

	var du uint64

	err := filepath.Walk(d.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		du += uint64(info.Size())
		return nil
	})

	if err != nil {
		return 0, err
	}

	return du, nil
}

// LevelDB needs to be closed.
func (d *Datastore) Close() (err error) {
	return d.DB.Close()
}

type leveldbBatch struct {
	b          *leveldb.Batch
	db         *leveldb.DB
	syncWrites bool
}

func (d *Datastore) Batch() (ds.Batch, error) {
	return &leveldbBatch{
		b:          new(leveldb.Batch),
		db:         d.DB,
		syncWrites: d.syncWrites,
	}, nil
}

func (b *leveldbBatch) Put(key ds.Key, value []byte) error {
	b.b.Put(key.Bytes(), value)
	return nil
}

func (b *leveldbBatch) Append(key ds.Key, value []byte, begin, length int) error {
	return nil
}

func (b *leveldbBatch) Commit() error {
	return b.db.Write(b.b, &opt.WriteOptions{Sync: b.syncWrites})
}

func (b *leveldbBatch) Delete(key ds.Key) error {
	b.b.Delete(key.Bytes())
	return nil
}

// A leveldb transaction embedding the accessor backed by the transaction.
type transaction struct {
	*accessor
	tx *leveldb.Transaction
}

func (t *transaction) Commit() error {
	return t.tx.Commit()
}

func (t *transaction) Discard() {
	t.tx.Discard()
}

func (d *Datastore) NewTransaction(readOnly bool) (ds.Txn, error) {
	tx, err := d.DB.OpenTransaction()
	if err != nil {
		return nil, err
	}
	accessor := &accessor{ldb: tx, syncWrites: false}
	return &transaction{accessor, tx}, nil
}
