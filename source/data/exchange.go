package data

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	metrics "github.com/ipfs/go-metrics-interface"
	p2phost "github.com/libp2p/go-libp2p-core/host"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/memoio/go-mefs/config"
	id "github.com/memoio/go-mefs/crypto/identity"
	dataformat "github.com/memoio/go-mefs/data-format"
	mpb "github.com/memoio/go-mefs/pb"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	ma "github.com/multiformats/go-multiaddr"
	mdns "github.com/multiformats/go-multiaddr-dns"
	mnet "github.com/multiformats/go-multiaddr-net"
)

var (
	// sort latencies in buckets with following upper bounds in seconds
	latencyBuckets = []float64{1e-4, 1e-3, 1e-2, 1e-1, 1}

	// sort sizes in buckets with following upper bounds in bytes
	sizeBuckets = []float64{1 << 6, 1 << 8, 1 << 12, 1 << 16, 1 << 20, 1 << 24}
)

var (
	errNoRouting    = errors.New("routing is not running")
	errNoConnection = errors.New("connection is offline")
	ErrRetry        = errors.New("ReTry Later")
	errKey          = errors.New("Key and value are mismatched")
)

type impl struct {
	netID   string // network address
	bstore  bs.Blockstore
	dstore  ds.Datastore
	aCache  *Cache
	rt      routing.Routing
	ph      p2phost.Host
	pubKeys *lru.ARCCache
	ms      *measure
}

type measure struct {
	getBlock    metrics.Counter
	putBlock    metrics.Counter
	appendBlock metrics.Counter
	delBlock    metrics.Counter
	getKey      metrics.Counter
	putKey      metrics.Counter
	appendKey   metrics.Counter
	delKey      metrics.Counter

	getLocalBlock    metrics.Counter
	putLocalBlock    metrics.Counter
	appendLocalBlock metrics.Counter
	delLocalBlock    metrics.Counter
	getLocalKey      metrics.Counter
	putLocalKey      metrics.Counter
	appendLocalKey   metrics.Counter
	delLocalKey      metrics.Counter

	getBlockErr    metrics.Counter
	putBlockErr    metrics.Counter
	appendBlockErr metrics.Counter
	delBlockErr    metrics.Counter
	getKeyErr      metrics.Counter
	putKeyErr      metrics.Counter
	appendKeyErr   metrics.Counter
	delKeyErr      metrics.Counter

	putLatency metrics.Histogram
	getLatency metrics.Histogram
	putSize    metrics.Histogram
	getSize    metrics.Histogram
}

// New returns data.Service
func New(id string, b bs.Blockstore, d ds.Datastore, host p2phost.Host, r routing.Routing) Service {
	if r == nil {
		log.Println("network is not running.")
	}

	// cache public keys, key is userID
	pcache, err := lru.NewARC(2048)
	if err != nil {
		utils.MLogger.Error("new lru err:", err)
		return nil
	}

	mea := &measure{
		putKey:      metrics.New("data.putRemoteKey_total", "Total number of Data.PutRemoteKey calls").Counter(),
		putLocalKey: metrics.New("data.putLocalKey_total", "Total number of Data.PutLocalKey calls").Counter(),
		putKeyErr:   metrics.New("data.putKeyErr_total", "Total number of Data.PutKey errors").Counter(),

		getKey:      metrics.New("data.getRemoteKey_total", "Total number of Data.GetRemoteKey calls").Counter(),
		getLocalKey: metrics.New("data.getLocalKey_total", "Total number of Data.GetLocalKey calls").Counter(),
		getKeyErr:   metrics.New("data.getKeyErr_total", "Total number of Data.GetKey errors").Counter(),

		appendKey:      metrics.New("data.appendRemoteKey_total", "Total number of Data.AppendRemoteKey calls").Counter(),
		appendLocalKey: metrics.New("data.appendLocalKey_total", "Total number of Data.AppendLocalKey calls").Counter(),
		appendKeyErr:   metrics.New("data.appendKeyErr_total", "Total number of Data.AppendKey errors").Counter(),

		delKey:      metrics.New("data.delRemoteKey_total", "Total number of Data.DelRemoteKey calls").Counter(),
		delLocalKey: metrics.New("data.delLocalKey_total", "Total number of Data.DelLocalKey calls").Counter(),
		delKeyErr:   metrics.New("data.delKeyErr_total", "Total number of Data.DelKey errors").Counter(),

		putBlock:      metrics.New("data.putRemoteBlock_total", "Total number of Data.PutRemoteBlock calls").Counter(),
		putLocalBlock: metrics.New("data.putLocalBlock_total", "Total number of Data.PutLocalBlock calls").Counter(),
		putBlockErr:   metrics.New("data.putBlockErr_total", "Total number of Data.PutBlock errors").Counter(),

		getBlock:      metrics.New("data.getRemoteBlock_total", "Total number of Data.GetRemoteBlock calls").Counter(),
		getLocalBlock: metrics.New("data.getLocalBlock_total", "Total number of Data.GetLocalBlock calls").Counter(),
		getBlockErr:   metrics.New("data.getBlockErr_total", "Total number of Data.GetBlock errors").Counter(),

		appendBlock:      metrics.New("data.appendRemoteBlock_total", "Total number of Data.AppendRemoteBlock calls").Counter(),
		appendLocalBlock: metrics.New("data.appendLocalBlock_total", "Total number of Data.AppendLocalBlock calls").Counter(),
		appendBlockErr:   metrics.New("data.appendBlockErr_total", "Total number of Data.AppendBlock errors").Counter(),

		delBlock:      metrics.New("data.delRemoteBlock_total", "Total number of Data.DelRemoteBlock calls").Counter(),
		delLocalBlock: metrics.New("data.delLocalBlock_total", "Total number of Data.DelLocalBlock calls").Counter(),
		delBlockErr:   metrics.New("data.delBlockErr_total", "Total number of Data.DelBlock errors").Counter(),

		getLatency: metrics.New("data.get.latency_seconds",
			"Latency distribution of Data.Get calls").Histogram(latencyBuckets),
		putLatency: metrics.New("data.put.latency_seconds",
			"Latency distribution of Data.Put calls").Histogram(latencyBuckets),

		getSize: metrics.New("data.get.size_bytes",
			"Latency distribution of Data.Get calls").Histogram(sizeBuckets),
		putSize: metrics.New("data.put.size_bytes",
			"Latency distribution of Data.Put calls").Histogram(sizeBuckets),
	}

	return &impl{
		netID:   id,
		rt:      r,
		ph:      host,
		dstore:  d,
		bstore:  b,
		aCache:  NewCache(b),
		pubKeys: pcache,
		ms:      mea,
	}
}

func (n *impl) GetNetAddr() string {
	return n.netID
}

func recordLatency(h metrics.Histogram, start time.Time) {
	elapsed := time.Since(start)
	h.Observe(elapsed.Seconds())
}

func (n *impl) SendMetaMessage(ctx context.Context, typ int32, key string, data, sig []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	utils.MLogger.Debug("SendMetaMessage: ", key, " to: ", to)

	p, err := peer.IDB58Decode(to)
	if err != nil {
		return err
	}

	if !n.Connect(ctx, to) {
		return errNoConnection
	}

	return n.rt.(*dht.KadDHT).SendMessage(ctx, typ, key, data, sig, p)
}

func (n *impl) SendMetaRequest(ctx context.Context, typ int32, key string, data, sig []byte, to string) ([]byte, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	utils.MLogger.Debug("SendMetaRequest: ", key, " to: ", to)

	p, err := peer.IDB58Decode(to)
	if err != nil {
		return nil, err
	}

	if !n.Connect(ctx, to) {
		return nil, errNoConnection
	}

	switch typ {
	case int32(mpb.OpType_Put):
		n.ms.getSize.Observe(float64(len(data) + len(key) + len(sig)))
		defer recordLatency(n.ms.putLatency, time.Now())
	case int32(mpb.OpType_Get):
		n.ms.putSize.Observe(float64(len(key) + len(sig)))
		defer recordLatency(n.ms.getLatency, time.Now())
	default:
	}

	res, err := n.rt.(*dht.KadDHT).SendRequest(ctx, typ, key, data, sig, p)
	if err != nil {
		utils.MLogger.Errorf("SendMetaRequest %s to %s fails: %s", key, to, err)
	}

	switch typ {
	case int32(mpb.OpType_Get):
		n.ms.putSize.Observe(float64(len(res)))
	default:
	}

	return res, err
}

func (n *impl) GetUserPublicKey(key string) ([]byte, error) {
	pubKey, ok := n.pubKeys.Get(key)
	if ok {
		return pubKey.([]byte), nil
	}

	km, err := metainfo.NewKey(key, mpb.KeyType_PublicKey)
	if err != nil {
		return nil, err
	}
	pubRecByte, err := n.dstore.Get(ds.NewKey(km.ToString()))
	if err != nil {
		return nil, err
	}

	pubrec := new(mpb.Record)
	err = proto.Unmarshal(pubRecByte, pubrec)
	if err != nil {
		return nil, err
	}

	gotID, err := id.GetIDFromCompressPubKey(pubrec.GetValue())
	if err != nil {
		return nil, err
	}

	if gotID != key {
		return nil, errKey
	}

	n.pubKeys.Add(key, pubrec.GetValue())

	return pubrec.GetValue(), nil
}

func (n *impl) VerifyKey(ctx context.Context, key string, value, sig []byte) bool {
	keys := strings.Split(key, metainfo.DELIMITER)
	if len(keys) < 2 {
		utils.MLogger.Warn("key is wrong for: ", key)
		return false
	}

	switch keys[1] {
	case strconv.Itoa(int(mpb.KeyType_PublicKey)):
		gotID, err := id.GetIDFromCompressPubKey(value)
		if err != nil {
			utils.MLogger.Warn("convert public key to id fails: ", err)
			return false
		}

		if gotID != keys[0] {
			return false
		}
		return true
	default:
		if len(keys) == 2 {
			return true
		}

		gotID := keys[2]
		pubKey, err := n.GetUserPublicKey(gotID)
		if err != nil {
			return false
		}

		return id.VerifySigForKey(pubKey, key, value, sig)
	}
}

func (n *impl) GetKey(ctx context.Context, key string, to string) ([]byte, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	utils.MLogger.Debug("GetKey: ", key, " from: ", to)

	if to == "local" {
		n.ms.getLocalKey.Inc()
		recByte, err := n.dstore.Get(ds.NewKey(key))
		if err != nil {
			n.ms.getKeyErr.Inc()
			return nil, err
		}
		rec := new(mpb.Record)
		err = proto.Unmarshal(recByte, rec)
		if err != nil {
			return nil, err
		}

		// todo:verify sig
		return rec.GetValue(), nil
	}

	if to != "" {
		if !n.Connect(ctx, to) {
			return nil, errNoConnection
		}
	}

	n.ms.getKey.Inc()
	res, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, nil, nil, to)
	if err != nil {
		n.ms.getKeyErr.Inc()
		return nil, err
	}
	return res, nil

}

func (n *impl) PutKey(ctx context.Context, key string, data, sig []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("PutKey: ", key, " to: ", to)

	if to == "local" {
		n.ms.putLocalKey.Inc()
		rec := &mpb.Record{
			Key:       []byte(key),
			Value:     data,
			Signature: sig,
		}

		recByte, err := proto.Marshal(rec)
		if err != nil {
			return err
		}

		err = n.dstore.Put(ds.NewKey(key), recByte)
		if err != nil {
			n.ms.putKeyErr.Inc()
			return err
		}

		return nil
	}

	if to != "" {
		if !n.Connect(ctx, to) {
			return errNoConnection
		}
	}

	n.ms.putKey.Inc()
	_, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Put), key, data, sig, to)
	if err != nil {
		n.ms.putKeyErr.Inc()
		return err
	}

	return nil
}

// to modify
func (n *impl) AppendKey(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("AppendKey: ", key, " to: ", to)
	if to == "local" {
		n.ms.appendLocalKey.Inc()
		skey := strings.Split(key, metainfo.DELIMITER)
		if len(skey) < 4 {
			return metainfo.ErrIllegalKey
		}

		s, err := strconv.Atoi(skey[2])
		if err != nil {
			return err
		}

		len, err := strconv.Atoi(skey[3])
		if err != nil {
			return err
		}

		bstr := strings.Join(skey[:2], metainfo.DELIMITER)

		err = n.dstore.Append(ds.NewKey(bstr), data, s, len)
		if err != nil {
			n.ms.appendKeyErr.Inc()
			return err
		}

		return nil
	}

	n.ms.appendKey.Inc()
	err := n.SendMetaMessage(ctx, int32(mpb.OpType_Append), key, data, nil, to)
	if err != nil {
		n.ms.appendKeyErr.Inc()
		return err
	}

	return nil
}

func (n *impl) DeleteKey(ctx context.Context, key string, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("DeleteKey: ", key, " from: ", to)

	if to == "local" {
		n.ms.delKey.Inc()
		err := n.dstore.Delete(ds.NewKey(key))
		if err != nil {
			n.ms.delKeyErr.Inc()
			return err
		}
		return nil
	}

	n.ms.delKey.Inc()
	err := n.SendMetaMessage(ctx, int32(mpb.OpType_Delete), key, nil, nil, to)
	if err != nil {
		n.ms.delKeyErr.Inc()
		return err
	}
	return nil
}

// GetBlock retrieves a particular partial block from the service,
// Getting it from the datastore using the key (hash).
// key: blockID/"Block"/start/length (todo)
func (n *impl) GetBlock(ctx context.Context, key string, sig []byte, to string) (blocks.Block, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	skey := strings.Split(key, metainfo.DELIMITER)
	utils.MLogger.Debug("GetBlock: ", key, " from: ", to)
	if to == "local" {
		n.ms.getLocalBlock.Inc()
		if len(skey) == 4 {
			s, err := strconv.Atoi(skey[2])
			if err != nil {
				return nil, err
			}

			len, err := strconv.Atoi(skey[3])
			if err != nil {
				return nil, err
			}
			// get from cache
			res, err := n.aCache.Get(skey[0], s, len)
			if err == nil {
				c := cid.NewCidV2([]byte(skey[0]))
				b, err := blocks.NewBlockWithCid(res, c)
				if err != nil {
					return nil, err
				}
				return b, nil
			}
		}

		block, err := n.bstore.Get(cid.NewCidV2([]byte(key)))
		if err == nil {
			return block, nil
		}
		if err.Error() == dataformat.ErrDataTooShort.Error() {
			go n.aCache.Summit(skey[0])
		}
		n.ms.getBlockErr.Inc()
		return nil, err
	}

	if len(skey) == 1 {
		km, _ := metainfo.NewKey(skey[0], mpb.KeyType_Block)
		key = km.ToString()
	}

	n.ms.getBlock.Inc()
	retry := 0
	for {
		retry++
		bdata, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, nil, sig, to)
		if err != nil {
			if err.Error() == ErrRetry.Error() {
				if retry > 5 {
					return nil, err
				}
				time.Sleep(time.Duration(retry) * 30 * time.Second)
				utils.MLogger.Debug("Retry GetBlock: ", key, " from: ", to)
				continue
			}
			return nil, err
		}

		c := cid.NewCidV2([]byte(key))
		b, err := blocks.NewBlockWithCid(bdata, c)
		if err != nil {
			n.ms.getBlockErr.Inc()
			return nil, err
		}
		return b, nil
	}
}

// key: blockID/"Block"
func (n *impl) PutBlock(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("PutBlock: ", key, " to: ", to)

	bids := strings.Split(key, metainfo.DELIMITER)
	if to == "local" {
		n.ms.putLocalBlock.Inc()
		bcid := cid.NewCidV2([]byte(bids[0]))
		b, err := blocks.NewBlockWithCid(data, bcid)
		if err != nil {
			return err
		}
		err = n.bstore.Put(b)
		if err != nil {
			n.ms.putBlockErr.Inc()
			return err
		}
		return nil
	}

	if len(bids) == 1 {
		km, _ := metainfo.NewKey(bids[0], mpb.KeyType_Block)
		key = km.ToString()
	}

	n.ms.putBlock.Inc()
	_, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Put), key, data, nil, to)
	if err != nil {
		n.ms.putBlockErr.Inc()
		return err
	}

	return nil
}

// key: blockID/"Block"
func (n *impl) AppendBlock(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("AppendBlock: ", key, " to: ", to)

	skey := strings.Split(key, metainfo.DELIMITER)
	if len(skey) < 4 {
		return metainfo.ErrIllegalKey
	}
	if to == "local" {
		n.ms.appendLocalBlock.Inc()
		s, err := strconv.Atoi(skey[2])
		if err != nil {
			return err
		}

		len, err := strconv.Atoi(skey[3])
		if err != nil {
			return err
		}

		if n.aCache.Has(skey[0]) {
			return n.aCache.Set(skey[0], data, s, len)
		}

		bcid := cid.NewCidV2([]byte(skey[0]))
		err = n.bstore.Append(bcid, data, s, len)
		if err != nil {
			utils.MLogger.Infof("AppendBlock %s to local fails %s", key, err)
			n.ms.appendBlockErr.Inc()
			return n.aCache.Set(skey[0], data, s, len)
		}
		return nil
	}

	n.ms.appendBlock.Inc()
	_, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Append), key, data, nil, to)
	if err != nil {
		n.ms.appendBlockErr.Inc()
		return err
	}

	return nil
}

// DeleteBlock deletes a block in the blockservice from the datastore
func (n *impl) DeleteBlock(ctx context.Context, key, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("DeleteBlock: ", key, " from: ", to)

	bids := strings.Split(key, metainfo.DELIMITER)
	if to == "local" {
		n.ms.delLocalBlock.Inc()
		bcid := cid.NewCidV2([]byte(bids[0]))
		err := n.bstore.DeleteBlock(bcid)
		if err != nil {
			n.ms.delBlockErr.Inc()
			return err
		}
		return nil
	}

	if len(bids) == 1 {
		km, _ := metainfo.NewKey(bids[0], mpb.KeyType_Block)
		key = km.ToString()
	}

	n.ms.delBlock.Inc()
	_, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Delete), key, nil, nil, to)
	if err != nil {
		n.ms.delBlockErr.Inc()
		return err
	}

	return nil
}

// BroadcastMessage
func (n *impl) BroadcastMessage(ctx context.Context, key string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("BroadcastMessage: ", key)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	n.rt.(*dht.KadDHT).GetValue(ctx, key)

	for _, pid := range n.ph.Network().Peers() {
		// send to all connectness
		n.SendMetaRequest(ctx, int32(mpb.OpType_Get), key, nil, nil, pid.Pretty())
	}

	return nil
}

func (n *impl) TestConnect() error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("Test Connect")

	waitTime := 0 //进行网络连接\

	if n.ph == nil {
		return errNoRouting
	}

	for {
		if waitTime > 60 { //连不上网？
			utils.MLogger.Error("No network, please add bootstrap peers restart.")
			return errNoRouting
		}
		if connPeers := n.ph.Network().Peers(); len(connPeers) != 0 { //刚启动还没连接节点，等等
			break //连上网了，退出
		} else {
			utils.MLogger.Error("waiting for network connection...")
			utils.MLogger.Error("run: mefs bootstrap add <node address>")
			time.Sleep(10 * time.Second) //没联网，等联网
		}
		waitTime++
	}
	return nil
}

//连接试三次
func (n *impl) FastConnect(ctx context.Context, to string) bool {
	if n.ph == nil || n.rt == nil {
		return false
	}

	id, err := peer.IDB58Decode(to)
	if err != nil {
		return false
	}

	if n.ph.Network().Connectedness(id) == inet.Connected {
		return true
	}

	pi, err := n.rt.FindPeer(ctx, id)
	if err != nil {
		return false
	}

	if swrm, ok := n.ph.Network().(*swarm.Swarm); ok {
		swrm.Backoff().Clear(pi.ID)
	}

	err = n.ph.Connect(ctx, pi)
	if err == nil {
		if n.ph.Network().Connectedness(id) == inet.Connected {
			return true
		}
	}

	return false
}

//连接试三次
func (n *impl) Connect(ctx context.Context, to string) bool {
	if n.ph == nil || n.rt == nil {
		return false
	}

	id, err := peer.IDB58Decode(to)
	if err != nil {
		return false
	}

	if n.ph.Network().Connectedness(id) == inet.Connected {
		return true
	}

	connectTryCount := 1
	for i := 0; i < connectTryCount; i++ {
		pi, err := n.rt.FindPeer(ctx, id)
		if err != nil {
			break
		}

		for j := 0; j < 3; j++ {
			if swrm, ok := n.ph.Network().(*swarm.Swarm); ok {
				swrm.Backoff().Clear(pi.ID)
			}

			err = n.ph.Connect(ctx, pi)
			if err == nil {
				if n.ph.Network().Connectedness(id) == inet.Connected {
					return true
				}
			}
		}
	}

	for i := 0; i < connectTryCount; i++ {
		res := n.getAddrAndConnect(ctx, id)
		if res {
			return true
		}
	}
	return false
}

func (n *impl) getAddrAndConnect(ctx context.Context, to peer.ID) bool {
	if n.ph == nil || n.rt == nil {
		return false
	}

	km, err := metainfo.NewKey(to.Pretty(), mpb.KeyType_ExternalAddress)
	if err != nil {
		return false
	}

	for _, defaultBootstrapAddress := range config.DefaultBootstrapAddresses {
		bi, err := ma.NewMultiaddr(defaultBootstrapAddress)
		if err != nil {
			continue
		}

		pi, err := peer.AddrInfoFromP2pAddr(bi)
		if err != nil {
			continue
		}

		npi := peer.AddrInfo{
			ID:    pi.ID,
			Addrs: pi.Addrs,
		}

		err = n.ph.Connect(ctx, npi)
		if err != nil {
			continue
		}

		res, err := n.SendMetaRequest(ctx, int32(mpb.OpType_Get), km.ToString(), nil, nil, pi.ID.Pretty())
		if err != nil {
			continue
		}

		pai, err := ma.NewMultiaddrBytes(res)
		if err != nil {
			utils.MLogger.Errorf("multiaddr %s failed to parse: %s", string(res), err)
			continue
		}

		npi = peer.AddrInfo{
			ID:    to,
			Addrs: []ma.Multiaddr{pai},
		}

		utils.MLogger.Debug(to, "has extern ip: ", npi.String())

		for i := 0; i < 3; i++ {
			if swrm, ok := n.ph.Network().(*swarm.Swarm); ok {
				swrm.Backoff().Clear(npi.ID)
			}

			err = n.ph.Connect(ctx, npi)
			if err == nil {
				return true
			}
		}
	}
	return false
}

func (n *impl) Itererate(prefix string) ([]dsq.Entry, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	utils.MLogger.Debug("Itererate: ", prefix)

	q := dsq.Query{Prefix: prefix}
	qr, _ := n.dstore.Query(q) //进行查询操作
	es, _ := qr.Rest()

	return es, nil
}

func (n *impl) GetPeers() ([]peer.ID, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}
	return n.ph.Network().Peers(), nil
}

func (n *impl) GetExternalAddr(p string) (ma.Multiaddr, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	utils.MLogger.Debug("GetExternalAddr: ", p)

	pid, err := peer.IDB58Decode(p)
	if err != nil {
		return nil, err
	}
	for _, c := range n.ph.Network().ConnsToPeer(pid) {
		rid := c.RemotePeer()
		if rid.Pretty() == p {
			addr := c.RemoteMultiaddr()
			utils.MLogger.Debug(p, " has extern ip: ", addr.String())
			ok := mnet.IsThinWaist(addr)
			if ok {
				// is ip4/tcp or ip4/udp
				ok = mnet.IsPrivateAddr(addr)
				if !ok {
					// is public addr
					return addr, nil
				}
			} else {
				// is /dns/...
				addrs, err := mdns.Resolve(context.Background(), addr)
				if err != nil {
					return nil, err
				}

				for _, maddr := range addrs {
					ok = mnet.IsPrivateAddr(maddr)
					if !ok {
						return maddr, nil
					}
				}
			}

		}
	}

	return nil, errors.New("No remote address")
}

func (n *impl) BlockStore() bs.Blockstore {
	return n.bstore
}

func (n *impl) DataStore() ds.Datastore {
	return n.dstore
}
