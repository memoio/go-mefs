package data

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	p2phost "github.com/libp2p/go-libp2p-core/host"
	inet "github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	swarm "github.com/libp2p/go-libp2p-swarm"
	"github.com/memoio/go-mefs/config"
	blocks "github.com/memoio/go-mefs/source/go-block-format"
	cid "github.com/memoio/go-mefs/source/go-cid"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dsq "github.com/memoio/go-mefs/source/go-datastore/query"
	bs "github.com/memoio/go-mefs/source/go-ipfs-blockstore"
	dht "github.com/memoio/go-mefs/source/go-libp2p-kad-dht"
	"github.com/memoio/go-mefs/utils"
	"github.com/memoio/go-mefs/utils/metainfo"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	errNoRouting = errors.New("routing is not running")
)

type impl struct {
	netID  string // network address
	bstore bs.Blockstore
	dstore ds.Datastore
	rt     routing.Routing
	ph     p2phost.Host
}

// New returns data.Service
func New(id string, b bs.Blockstore, d ds.Datastore, host p2phost.Host, r routing.Routing) Service {
	if r == nil {
		log.Println("network is not running.")
	}

	return &impl{
		netID:  id,
		rt:     r,
		ph:     host,
		dstore: d,
		bstore: b,
	}
}

func (n *impl) GetNetAddr() string {
	return n.netID
}

func (n *impl) SendMetaMessage(ctx context.Context, typ int32, key string, data, sig []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("SendMetaMessage: ", key, " to: ", to)

	p, err := peer.IDB58Decode(to)
	if err != nil {
		return err
	}

	n.Connect(ctx, to)

	return n.rt.(*dht.KadDHT).SendMessage(ctx, typ, key, data, sig, p)
}

func (n *impl) SendMetaRequest(ctx context.Context, typ int32, key string, data, sig []byte, to string) ([]byte, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	utils.MLogger.Debug("SendMetaRequest: ", key, " to: ", to)

	p, err := peer.IDB58Decode(to)
	if err != nil {
		return nil, err
	}

	n.Connect(ctx, to)

	return n.rt.(*dht.KadDHT).SendRequest(ctx, typ, key, data, sig, p)
}

func (n *impl) GetKey(ctx context.Context, key string, to string) ([]byte, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	utils.MLogger.Debug("GetKey: ", key, " from: ", to)

	if to != "local" && to != "" {
		n.Connect(ctx, to)
	}

	res, err := n.rt.(*dht.KadDHT).GetFrom(ctx, key, to)
	if err != nil && err != routing.ErrNotFound {
		utils.MLogger.Error("GetKey err:", err, ", key is: ", key, " from: ", to)
		return nil, err
	}
	return res, nil
}

func (n *impl) PutKey(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("PutKey: ", key, " to: ", to)

	if to != "local" && to != "" {
		n.Connect(ctx, to)
	}

	return n.rt.(*dht.KadDHT).PutTo(ctx, key, data, to)
}

// to modify
func (n *impl) AppendKey(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("AppendKey: ", key, " to: ", to)

	if to == "local" {
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

		return n.dstore.Append(ds.NewKey(bstr), data, s, len)

	}

	return n.SendMetaMessage(ctx, int32(metainfo.Append), key, data, nil, to)
}

func (n *impl) DeleteKey(ctx context.Context, key string, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("DeleteKey: ", key, " from: ", to)

	if to == "local" {
		return n.dstore.Delete(ds.NewKey(key))
	}

	return n.SendMetaMessage(ctx, int32(metainfo.Delete), key, nil, nil, to)
}

// GetBlock retrieves a particular block from the service,
// Getting it from the datastore using the key (hash).
func (n *impl) GetBlock(ctx context.Context, key string, sig []byte, to string) (blocks.Block, error) {
	if n.ph == nil || n.rt == nil {
		return nil, errNoRouting
	}

	utils.MLogger.Debug("GetBlock: ", key, " from: ", to)
	bids := strings.Split(key, metainfo.DELIMITER)
	if to == "local" {
		block, err := n.bstore.Get(cid.NewCidV2([]byte(bids[0])))
		if err == nil {
			return block, nil
		}

		return nil, err
	}

	if len(bids) == 1 {
		km, _ := metainfo.NewKeyMeta(bids[0], metainfo.Block)
		key = km.ToString()
	}

	bdata, err := n.SendMetaRequest(ctx, int32(metainfo.Get), key, nil, sig, to)
	if err != nil {
		return nil, err
	}

	c := cid.NewCidV2([]byte(key))
	b, err := blocks.NewBlockWithCid([]byte(bdata), c)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// key: blockID/"Block"
func (n *impl) PutBlock(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return errNoRouting
	}

	utils.MLogger.Debug("PutBlock: ", key, " to: ", to)

	bids := strings.Split(key, metainfo.DELIMITER)
	if to == "local" {
		bcid := cid.NewCidV2([]byte(bids[0]))
		b, err := blocks.NewBlockWithCid(data, bcid)
		if err != nil {
			return err
		}
		err = n.bstore.Put(b)
		if err != nil {
			return err
		}
		return nil
	}

	if len(bids) == 1 {
		km, _ := metainfo.NewKeyMeta(bids[0], metainfo.Block)
		key = km.ToString()
	}

	_, err := n.SendMetaRequest(ctx, int32(metainfo.Put), key, data, nil, to)
	if err != nil {
		return err
	}

	return nil
}

// key: blockID/"Block"/start/length (segSize)
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
		s, err := strconv.Atoi(skey[2])
		if err != nil {
			return err
		}

		len, err := strconv.Atoi(skey[3])
		if err != nil {
			return err
		}

		bcid := cid.NewCidV2([]byte(skey[0]))

		err = n.bstore.Append(bcid, data, s, len)
		if err != nil {
			return err
		}
		return nil
	}

	_, err := n.SendMetaRequest(ctx, int32(metainfo.Append), key, data, nil, to)
	if err != nil {
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
		bcid := cid.NewCidV2([]byte(bids[0]))
		return n.bstore.DeleteBlock(bcid)
	}

	if len(bids) == 1 {
		km, _ := metainfo.NewKeyMeta(bids[0], metainfo.Block)
		key = km.ToString()
	}

	_, err := n.SendMetaRequest(ctx, int32(metainfo.Delete), key, nil, nil, to)
	if err != nil {
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
	_, err := n.rt.(*dht.KadDHT).GetValue(ctx, key)
	return err
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

	km, err := metainfo.NewKeyMeta(to.Pretty(), metainfo.ExternalAddress)
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

		res, err := n.SendMetaRequest(ctx, int32(metainfo.Get), km.ToString(), nil, nil, pi.ID.Pretty())
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

func (n *impl) GetExternalAddr(p string) ([]byte, error) {
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
			return addr.Bytes(), nil
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
