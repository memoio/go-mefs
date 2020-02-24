package data

import (
	"sort"
	"sync"
	"time"

	dataformat "github.com/memoio/go-mefs/data-format"
	bf "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
)

const (
	defaultGCInterval time.Duration = 1 * time.Minute
	defaultExpiration time.Duration = 37 * time.Second
)

type seg struct {
	Value  []byte
	Start  int
	Length int
}

type Item struct {
	sync.RWMutex
	Segs       []*seg
	Key        []byte
	Expiration int64
	tries      int
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

	deletion := false
	for _, key := range keys {
		ci, ok := c.iMap.Load(key)
		if !ok {
			continue
		}

		val := ci.(*Item)
		deletion = false
		val.Lock()
		if val.Expiration < now && len(val.Segs) > 0 {
			err := c.bstore.Append(cid.NewCidV2(val.Key), val.Segs[0].Value, val.Segs[0].Start, val.Segs[0].Length)
			if err != nil {
				val.Expiration = time.Now().Add(defaultExpiration).UnixNano()
				val.tries++
			} else {
				val.tries = 0
				if len(val.Segs) > 1 {
					val.Expiration = time.Now().Add(defaultExpiration).UnixNano()
					val.Segs = val.Segs[1:]
				}
			}
		}

		if val.tries > 10 {
			deletion = true
		}

		val.Unlock()
		if deletion {
			c.iMap.Delete(key)
		}
	}
}

func (c *Cache) Summit(key string) {
	ci, ok := c.iMap.Load(key)
	if !ok {
		return
	}

	deletion := false
	val := ci.(*Item)
	val.Lock()
	if len(val.Segs) > 0 {
		err := c.bstore.Append(cid.NewCidV2(val.Key), val.Segs[0].Value, val.Segs[0].Start, val.Segs[0].Length)
		if err != nil {
			val.Expiration = time.Now().Add(defaultExpiration).UnixNano()
			val.tries++
		} else {
			if len(val.Segs) > 1 {
				val.Expiration = time.Now().Add(defaultExpiration).UnixNano()
				val.Segs = val.Segs[1:]
			}
		}
	}
	val.Unlock()
	if deletion {
		c.iMap.Delete(key)
	}
}

func (c *Cache) Has(k string) bool {
	_, found := c.iMap.Load(k)
	return found
}

func (c *Cache) Get(k string, start, length int) ([]byte, error) {
	ci, ok := c.iMap.Load(k)
	if ok {
		val := ci.(*Item)
		val.RLock()
		defer val.RUnlock()
		for i := 0; i < len(val.Segs); i++ {
			if start >= val.Segs[i].Start && start+length <= val.Segs[i].Start+val.Segs[i].Length {
				pre, preLen, err := bf.PrefixDecode(val.Segs[i].Value)
				if err != nil {
					return nil, err
				}

				tagSize, ok := dataformat.TagMap[int(pre.Bopts.TagFlag)]
				if !ok {
					return nil, dataformat.ErrWrongTagFlag
				}

				tagNum := 2 + (pre.Bopts.ParityCount-1)/pre.Bopts.DataCount
				fieldSize := int(pre.Bopts.SegmentSize + tagNum*int32(tagSize))

				pre.Start = int32(start)
				prebuf, npreLen, err := bf.PrefixEncode(pre)
				if err != nil {
					return nil, err
				}

				res := make([]byte, npreLen+fieldSize*length)
				copy(res, prebuf)
				copy(res[preLen:], val.Segs[i].Value[preLen+(start-val.Segs[i].Start)*fieldSize:preLen+(start-val.Segs[i].Start+length)*fieldSize])
				return res, nil
			}
			if start < val.Segs[i].Start {
				break
			}
		}

	}

	return nil, ErrRetry
}

func (c *Cache) Set(k string, val []byte, start, length int) error {
	e := time.Now().Add(defaultExpiration).UnixNano()
	it, found := c.iMap.Load(k)
	if !found {
		ni := &Item{
			Key:        []byte(k),
			Expiration: e,
			Segs:       make([]*seg, 1),
		}
		ns := &seg{
			Value:  val,
			Start:  start,
			Length: length,
		}

		ni.Segs[0] = ns
		c.iMap.Store(k, ni)
		return nil
	}

	ni := it.(*Item)
	ni.Lock()
	defer ni.Unlock()

	ns := &seg{
		Start:  start,
		Length: length,
		Value:  val,
	}

	ni.Segs = append(ni.Segs, ns)

	sort.Slice(ni.Segs, func(i, j int) bool {
		return ni.Segs[i].Start < ni.Segs[j].Start
	})

	if len(ni.Segs) <= 1 {
		return nil
	}

	seBefore := ni.Segs[0]
	has := 1
	for i := 1; i < len(ni.Segs); i++ {
		seAfter := ni.Segs[i]
		if seAfter.Start == seBefore.Start+seBefore.Length {
			seBefore.Length += seAfter.Length
			_, preLen, err := bf.PrefixDecode(seAfter.Value)
			if err != nil {
				return err
			}
			seBefore.Value = append(seBefore.Value, seAfter.Value[preLen:]...)
			seAfter.Start = int(^uint(0) >> 1) // put to last
		} else {
			seBefore = seAfter // to next
			has++
		}
	}

	sort.Slice(ni.Segs, func(i, j int) bool {
		return ni.Segs[i].Start < ni.Segs[j].Start
	})

	ni.Segs = ni.Segs[:has]

	return nil
}

func (c *Cache) StopGc() {
	c.stopGc <- true
}

func NewCache(bstore bs.Blockstore) *Cache {
	c := &Cache{
		gcInterval: defaultGCInterval,
		stopGc:     make(chan bool),
		bstore:     bstore,
	}

	go c.gcLoop()
	return c
}
