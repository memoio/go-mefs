package providers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	u "github.com/ipfs/go-ipfs-util"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/ipfs/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	dssync "github.com/memoio/go-mefs/source/go-datastore/sync"
	//
	// used by TestLargeProvidersSet: do not remove
	// lds "github.com/ipfs/go-ds-leveldb"
)

func TestProviderManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mid := peer.ID("testing")
	p := NewProviderManager(ctx, mid, dssync.MutexWrap(ds.NewMapDatastore()))
	a := cid.NewCidV0(u.Hash([]byte("test")))
	p.AddProvider(ctx, a, peer.ID("testingprovider"))

	// Not cached
	resp := p.GetProviders(ctx, a)
	if len(resp) != 1 {
		t.Fatal("Could not retrieve provider.")
	}

	// Cached
	resp = p.GetProviders(ctx, a)
	if len(resp) != 1 {
		t.Fatal("Could not retrieve provider.")
	}
	p.proc.Close()
}

func TestProvidersDatastore(t *testing.T) {
	old := lruCacheSize
	lruCacheSize = 10
	defer func() { lruCacheSize = old }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mid := peer.ID("testing")
	p := NewProviderManager(ctx, mid, dssync.MutexWrap(ds.NewMapDatastore()))
	defer p.proc.Close()

	friend := peer.ID("friend")
	var cids []cid.Cid
	for i := 0; i < 100; i++ {
		c := cid.NewCidV0(u.Hash([]byte(fmt.Sprint(i))))
		cids = append(cids, c)
		p.AddProvider(ctx, c, friend)
	}

	for _, c := range cids {
		resp := p.GetProviders(ctx, c)
		if len(resp) != 1 {
			t.Fatal("Could not retrieve provider.")
		}
		if resp[0] != friend {
			t.Fatal("expected provider to be 'friend'")
		}
	}
}

func TestProvidersSerialization(t *testing.T) {
	dstore := dssync.MutexWrap(ds.NewMapDatastore())

	k := cid.NewCidV0(u.Hash(([]byte("my key!"))))
	p1 := peer.ID("peer one")
	p2 := peer.ID("peer two")
	pt1 := time.Now()
	pt2 := pt1.Add(time.Hour)

	err := writeProviderEntry(dstore, k, p1, pt1)
	if err != nil {
		t.Fatal(err)
	}

	err = writeProviderEntry(dstore, k, p2, pt2)
	if err != nil {
		t.Fatal(err)
	}

	pset, err := loadProvSet(dstore, k)
	if err != nil {
		t.Fatal(err)
	}

	lt1, ok := pset.set[p1]
	if !ok {
		t.Fatal("failed to load set correctly")
	}

	if !pt1.Equal(lt1) {
		t.Fatalf("time wasnt serialized correctly, %v != %v", pt1, lt1)
	}

	lt2, ok := pset.set[p2]
	if !ok {
		t.Fatal("failed to load set correctly")
	}

	if !pt2.Equal(lt2) {
		t.Fatalf("time wasnt serialized correctly, %v != %v", pt1, lt1)
	}
}

func TestProvidesExpire(t *testing.T) {
	pval := ProvideValidity
	cleanup := defaultCleanupInterval
	ProvideValidity = time.Second / 2
	defaultCleanupInterval = time.Second / 2
	defer func() {
		ProvideValidity = pval
		defaultCleanupInterval = cleanup
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ds := dssync.MutexWrap(ds.NewMapDatastore())
	mid := peer.ID("testing")
	p := NewProviderManager(ctx, mid, ds)

	peers := []peer.ID{"a", "b"}
	var cids []cid.Cid
	for i := 0; i < 10; i++ {
		c := cid.NewCidV0(u.Hash([]byte(fmt.Sprint(i))))
		cids = append(cids, c)
	}

	for _, c := range cids[:5] {
		p.AddProvider(ctx, c, peers[0])
		p.AddProvider(ctx, c, peers[1])
	}

	time.Sleep(time.Second / 4)

	for _, c := range cids[5:] {
		p.AddProvider(ctx, c, peers[0])
		p.AddProvider(ctx, c, peers[1])
	}

	for _, c := range cids {
		out := p.GetProviders(ctx, c)
		if len(out) != 2 {
			t.Fatal("expected providers to still be there")
		}
	}

	time.Sleep(3 * time.Second / 8)

	for _, c := range cids[:5] {
		out := p.GetProviders(ctx, c)
		if len(out) > 0 {
			t.Fatal("expected providers to be cleaned up, got: ", out)
		}
	}

	for _, c := range cids[5:] {
		out := p.GetProviders(ctx, c)
		if len(out) != 2 {
			t.Fatal("expected providers to still be there")
		}
	}

	time.Sleep(time.Second / 2)

	// Stop to prevent data races
	p.Process().Close()

	if p.providers.Len() != 0 {
		t.Fatal("providers map not cleaned up")
	}

	res, err := ds.Query(dsq.Query{Prefix: providersKeyPrefix})
	if err != nil {
		t.Fatal(err)
	}
	rest, err := res.Rest()
	if err != nil {
		t.Fatal(err)
	}
	if len(rest) > 0 {
		t.Fatal("expected everything to be cleaned out of the datastore")
	}
}

var _ = ioutil.NopCloser
var _ = os.DevNull

/* This can be used for profiling. Keeping it commented out for now to avoid incurring extra CI time
func TestLargeProvidersSet(t *testing.T) {
	old := lruCacheSize
	lruCacheSize = 10
	defer func() { lruCacheSize = old }()

	dirn, err := ioutil.TempDir("", "provtest")
	if err != nil {
		t.Fatal(err)
	}

	opts := &lds.Options{
		NoSync:      true,
		Compression: 1,
	}
	lds, err := lds.NewDatastore(dirn, opts)
	if err != nil {
		t.Fatal(err)
	}
	_ = lds

	defer func() {
		os.RemoveAll(dirn)
	}()

	ctx := context.Background()
	var peers []peer.ID
	for i := 0; i < 3000; i++ {
		peers = append(peers, peer.ID(fmt.Sprint(i)))
	}

	mid := peer.ID("myself")
	p := NewProviderManager(ctx, mid, lds)
	defer p.proc.Close()

	var cids []cid.Cid
	for i := 0; i < 1000; i++ {
		c := cid.NewCidV0(u.Hash([]byte(fmt.Sprint(i))))
		cids = append(cids, c)
		for _, pid := range peers {
			p.AddProvider(ctx, c, pid)
		}
	}

	for i := 0; i < 5; i++ {
		start := time.Now()
		for _, c := range cids {
			_ = p.GetProviders(ctx, c)
		}
		elapsed := time.Since(start)
		fmt.Printf("query %f ms\n", elapsed.Seconds()*1000)
	}
}
*/

func TestUponCacheMissProvidersAreReadFromDatastore(t *testing.T) {
	old := lruCacheSize
	lruCacheSize = 1
	defer func() { lruCacheSize = old }()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p1, p2 := peer.ID("a"), peer.ID("b")
	c1 := cid.NewCidV1(cid.DagCBOR, u.Hash([]byte("1")))
	c2 := cid.NewCidV1(cid.DagCBOR, u.Hash([]byte("2")))
	pm := NewProviderManager(ctx, p1, dssync.MutexWrap(ds.NewMapDatastore()))

	// add provider
	pm.AddProvider(ctx, c1, p1)
	// make the cached provider for c1 go to datastore
	pm.AddProvider(ctx, c2, p1)
	// now just offloaded record should be brought back and joined with p2
	pm.AddProvider(ctx, c1, p2)

	c1Provs := pm.GetProviders(ctx, c1)
	if len(c1Provs) != 2 {
		t.Fatalf("expected c1 to be provided by 2 peers, is by %d", len(c1Provs))
	}
}

func TestWriteUpdatesCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	p1, p2 := peer.ID("a"), peer.ID("b")
	c1 := cid.NewCidV1(cid.DagCBOR, u.Hash([]byte("1")))
	pm := NewProviderManager(ctx, p1, dssync.MutexWrap(ds.NewMapDatastore()))

	// add provider
	pm.AddProvider(ctx, c1, p1)
	// force into the cache
	pm.GetProviders(ctx, c1)
	// add a second provider
	pm.AddProvider(ctx, c1, p2)

	c1Provs := pm.GetProviders(ctx, c1)
	if len(c1Provs) != 2 {
		t.Fatalf("expected c1 to be provided by 2 peers, is by %d", len(c1Provs))
	}
}
