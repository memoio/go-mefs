package namespace_test

import (
	"fmt"

	ds "github.com/memoio/go-mefs/source/go-datastore"
	nsds "github.com/memoio/go-mefs/source/go-datastore/namespace"
)

func Example() {
	mp := ds.NewMapDatastore()
	ns := nsds.Wrap(mp, ds.NewKey("/foo/bar"))

	k := ds.NewKey("/beep")
	v := "boop"

	if err := ns.Put(k, []byte(v)); err != nil {
		panic(err)
	}
	fmt.Printf("ns.Put %s %s\n", k, v)

	v2, _ := ns.Get(k)
	fmt.Printf("ns.Get %s -> %s\n", k, v2)

	k3 := ds.NewKey("/foo/bar/beep")
	v3, _ := mp.Get(k3)
	fmt.Printf("mp.Get %s -> %s\n", k3, v3)
	// Output:
	// ns.Put /beep boop
	// ns.Get /beep -> boop
	// mp.Get /foo/bar/beep -> boop
}
