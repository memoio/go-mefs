package datastore

type op struct {
	delete bool
	begin  int
	length int
	value  []byte
}

// basicBatch implements the transaction interface for datastores who do
// not have any sort of underlying transactional support
type basicBatch struct {
	ops map[Key]op

	target Datastore
}

func NewBasicBatch(ds Datastore) Batch {
	return &basicBatch{
		ops:    make(map[Key]op),
		target: ds,
	}
}

func (bt *basicBatch) Put(key Key, val []byte) error {
	bt.ops[key] = op{value: val}
	return nil
}

func (bt *basicBatch) Append(key Key, val []byte, begin, length int) error {
	bt.ops[key] = op{
		value:  val,
		begin:  begin,
		length: length,
	}
	return nil
}

func (bt *basicBatch) Delete(key Key) error {
	bt.ops[key] = op{delete: true}
	return nil
}

func (bt *basicBatch) Commit() error {
	var err error
	for k, op := range bt.ops {
		if op.delete {
			err = bt.target.Delete(k)
		} else {
			if op.begin == 0 && op.length == 0 {
				err = bt.target.Put(k, op.value)
			} else {
				err = bt.target.Append(k, op.value, op.begin, op.length)
			}

		}
		if err != nil {
			break
		}
	}

	return err
}
