package measure

import (
	"testing"

	dstest "github.com/memoio/go-mefs/source/go-datastore/test"
)

func TestSuite(t *testing.T) {
	d := New("measure", dstest.NewTestDatastore(false))
	dstest.SubtestAll(t, d)
}
