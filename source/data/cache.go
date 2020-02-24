package data

import (
	"sync"
	"time"

	bf "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
)

const (
	defaultGCInterval time.Duration = 1 * time.Minute
	defaultExpiration time.Duration = 17 * time.Second
)

type Item struct {
	Key        string
	Value      []byte
	Start      int
	Length     int
	Expiration int64
}

type Cache struct {
	iMap       sync.Map
	gcInterval time.Duration // 过期数据项清理周期
	stopGc     chan bool
	bstore     bs.Blockstore
}

// 过期缓存数据项清理
func (c *Cache) gcLoop() {
	ticker := time.NewTicker(c.gcInterval)
	for {
		select {
		case <-ticker.C:
			c.Flush()
		case <-c.stopGc:
			ticker.Stop()
			return
		}
	}
}

func (c *Cache) Flush() {
	now := time.Now().UnixNano()
	var keys []string
	c.iMap.Range(func(k, v interface{}) bool {
		keys = append(keys, k.(string))
		return true
	})

	for _, key := range keys {
		ci, ok := c.iMap.Load(key)
		if !ok {
			continue
		}

		val := ci.(*Item)
		if val.Expiration < now {
			err := c.bstore.Append(cid.NewCidV2([]byte(val.Key)), val.Value, val.Start, val.Length)
			if err != nil {
				val.Expiration = time.Now().Add(defaultExpiration).UnixNano()
			} else {
				c.iMap.Delete(key)
			}
		}
	}
}

func (c *Cache) Summit(key string) {
	ci, ok := c.iMap.Load(key)
	if !ok {
		return
	}

	val := ci.(*Item)
	err := c.bstore.Append(cid.NewCidV2([]byte(val.Key)), val.Value, val.Start, val.Length)
	if err != nil {
		val.Expiration = time.Now().Add(defaultExpiration).UnixNano()
	} else {
		c.iMap.Delete(key)
	}
}

func (c *Cache) Has(k string) bool {
	_, found := c.iMap.Load(k)
	return found
}

func (c *Cache) Set(k string, val []byte, start, length int) error {
	e := time.Now().Add(defaultExpiration).UnixNano()
	it, found := c.iMap.Load(k)
	if !found {
		ni := &Item{
			Value:      val,
			Start:      start,
			Length:     length,
			Expiration: e,
		}
		c.iMap.Store(k, ni)
	}

	ni := it.(*Item)

	if start == ni.Start+ni.Length {
		ni.Start = start
		ni.Length = length
		_, preLen, err := bf.PrefixDecode(val)
		if err != nil {
			return err
		}
		ni.Value = append(ni.Value, val[preLen:]...)
	}

	return nil
}

func (c *Cache) StopGc() {
	c.stopGc <- true
}

func NewCache() *Cache {
	c := &Cache{
		gcInterval: defaultGCInterval,
		stopGc:     make(chan bool),
	}

	go c.gcLoop()
	return c
}
