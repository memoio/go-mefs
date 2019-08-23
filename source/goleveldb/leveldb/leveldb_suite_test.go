package leveldb

import (
	"testing"

	"github.com/memoio/go-mefs/source/goleveldb/leveldb/testutil"
)

func TestLevelDB(t *testing.T) {
	testutil.RunSuite(t, "LevelDB Suite")
}
