package sync

import (
	"testing"

	ds "github.com/memoio/go-mefs/source/go-datastore"
	dstest "github.com/memoio/go-mefs/source/go-datastore/test"
)

func TestSync(t *testing.T) {
	dstest.SubtestAll(t, MutexWrap(ds.NewMapDatastore()))
}
