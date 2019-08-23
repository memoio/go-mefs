package memdb

import (
	"testing"

	"github.com/memoio/go-mefs/source/goleveldb/leveldb/testutil"
)

func TestMemDB(t *testing.T) {
	testutil.RunSuite(t, "MemDB Suite")
}
