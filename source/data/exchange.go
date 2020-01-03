package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	iaddr "github.com/ipfs/go-ipfs-addr"
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
	"github.com/memoio/go-mefs/utils/metainfo"
	ma "github.com/multiformats/go-multiaddr"
)

var (
	ErrNoRouting = errors.New("routing is not running")
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
		return ErrNoRouting
	}

	p, err := peer.IDB58Decode(to)
	if err != nil {
		return err
	}

	n.Connect(ctx, to)

	return n.rt.(*dht.KadDHT).SendMessage(ctx, typ, key, data, sig, p)
}

func (n *impl) SendMetaRequest(ctx context.Context, typ int32, key string, data, sig []byte, to string) ([]byte, error) {
	if n.ph == nil || n.rt == nil {
		return nil, ErrNoRouting
	}

	p, err := peer.IDB58Decode(to)
	if err != nil {
		return nil, err
	}

	n.Connect(ctx, to)

	return n.rt.(*dht.KadDHT).SendRequest(ctx, typ, key, data, sig, p)
}

func (n *impl) GetKey(ctx context.Context, key string, to string) ([]byte, error) {
	if n.ph == nil || n.rt == nil {
		return nil, ErrNoRouting
	}

	if to != "local" && to != "" {
		n.Connect(ctx, to)
	}

	res, err := n.rt.(*dht.KadDHT).GetFrom(ctx, key, to)
	if err != nil && err != routing.ErrNotFound {
		log.Println("GetKey err:", err, "key is: ", key, "from: ", to)
		return nil, err
	}
	return res, nil
}

func (n *impl) PutKey(ctx context.Context, key string, data []byte, to string) error {
	if n.ph == nil || n.rt == nil {
		return ErrNoRouting
	}

	if to != "local" && to != "" {
		n.Connect(ctx, to)
	}

	return n.rt.(*dht.KadDHT).PutTo(ctx, key, data, to)
}

// to modify
func (n *impl) AppendKey(ctx context.Context, key string, data []byte, to string) error {
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
	if to == "local" {
		return n.dstore.Delete(ds.NewKey(key))
	}

	return n.SendMetaMessage(ctx, int32(metainfo.Delete), key, nil, nil, to)
}

// GetBlock retrieves a particular block from the service,
// Getting it from the datastore using the key (hash).
func (n *impl) GetBlock(ctx context.Context, key string, sig []byte, to string) (blocks.Block, error) {
	if to == "local" {
		block, err := n.bstore.Get(cid.NewCidV2([]byte(key)))
		if err == nil {
			return block, nil
		}

		return nil, err
	}

	bdata, err := n.SendMetaRequest(ctx, int32(metainfo.Get), key, nil, sig, to)
	if err != nil {
		return nil, err
	}

	if string(bdata) == "complete" {
		return nil, errors.New("get block failed")
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

// key: blockID/"Block"/start/offset
func (n *impl) AppendBlock(ctx context.Context, key string, data []byte, to string) error {
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

		bstr := strings.Join(skey[:2], metainfo.DELIMITER)

		bcid := cid.NewCidV2([]byte(bstr))

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

	return errors.New("Routing is nil")
}

// DeleteBlock deletes a block in the blockservice from the datastore
func (n *impl) DeleteBlock(ctx context.Context, key, to string) error {
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err := n.rt.(*dht.KadDHT).GetValue(ctx, key)
	return err
}

func (n *impl) TestConnect() error {
	waitTime := 0 //进行网络连接\

	if n.ph == nil {
		return ErrNoRouting
	}

	for {
		if waitTime > 60 { //连不上网？
			log.Println("No network, please add bootstrap peers restart.")
			return ErrNoRouting
		}
		if connPeers := n.ph.Network().Peers(); len(connPeers) != 0 { //刚启动还没连接节点，等等
			break //连上网了，退出
		} else {
			log.Println("waiting for network connection...")
			log.Println("run: mefs bootstrap add <node address>")
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

	var retry = false
	connectTryCount := 3
	for i := 0; i <= connectTryCount; i++ {
		if retry { // retry three times
			ctx = context.WithValue(ctx, "ExternIP", true)
		}

		pi, err := n.rt.FindPeer(ctx, id)
		if err != nil {
			fmt.Printf("findpeer err: %s\n", err)
			return false
		}

		if swrm, ok := n.ph.Network().(*swarm.Swarm); ok {
			swrm.Backoff().Clear(pi.ID)
		}

		err = n.ph.Connect(ctx, pi)
		if err == nil {
			return true
		}
		retry = true
	}

	for i := 0; i <= connectTryCount; i++ {
		res := n.getAddrAndConnect(ctx, to)
		if res {
			return true
		}
	}
	return false
}

func (n *impl) getAddrAndConnect(ctx context.Context, to string) bool {
	km, err := metainfo.NewKeyMeta(to, metainfo.ExternalAddress)
	if err != nil {
		return false
	}

	for _, defaultBootstrapAddress := range config.DefaultBootstrapAddresses {
		addr := strings.Split(defaultBootstrapAddress, "/")
		peerID := addr[len(addr)-1]
		res, err := n.SendMetaRequest(ctx, int32(metainfo.Get), km.ToString(), nil, nil, peerID)
		if err != nil {
			continue
		}
		paddr := string(res) + "/p2p/" + to

		fmt.Println("try to connect: ", paddr)

		pai, err := peersWithAddresses(paddr)
		if err != nil {
			continue
		}

		if swrm, ok := n.ph.Network().(*swarm.Swarm); ok {
			swrm.Backoff().Clear(pai.ID)
		}

		err = n.ph.Connect(ctx, pai)
		if err == nil {
			return true
		}
	}
	return false
}

// peersWithAddresses is a function that takes in a slice of string peer addresses
// (multiaddr + peerid) and returns a slice of properly constructed peers
func peersWithAddresses(addrs string) (peer.AddrInfo, error) {
	pai := peer.AddrInfo{}

	iaddrs, err := iaddr.ParseString(addrs)
	if err != nil {
		return pai, err
	}

	pai.ID = iaddrs.ID()
	tpt := iaddrs.Transport()

	if tpt != nil {
		pai.Addrs = []ma.Multiaddr{tpt}
	}

	return pai, nil
}

func (n *impl) Itererate(prefix string) ([]dsq.Entry, error) {
	q := dsq.Query{Prefix: prefix}
	qr, _ := n.dstore.Query(q) //进行查询操作
	es, _ := qr.Rest()

	return es, nil
}

func (n *impl) GetPeers() []peer.ID {
	return n.ph.Network().Peers()
}

func (n *impl) GetExternalAddr(p string) ([]byte, error) {
	pid, err := peer.IDB58Decode(p)
	if err != nil {
		return nil, err
	}
	for _, c := range n.ph.Network().ConnsToPeer(pid) {
		rid := c.RemotePeer()
		if rid.Pretty() == p {
			addr := c.RemoteMultiaddr()
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
