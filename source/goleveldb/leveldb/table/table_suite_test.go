package table

import (
	"testing"

	"github.com/memoio/go-mefs/source/goleveldb/leveldb/testutil"
)

func TestTable(t *testing.T) {
	testutil.RunSuite(t, "Table Suite")
}
