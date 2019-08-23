package iterator_test

import (
	"testing"

	"github.com/memoio/go-mefs/source/goleveldb/leveldb/testutil"
)

func TestIterator(t *testing.T) {
	testutil.RunSuite(t, "Iterator Suite")
}
