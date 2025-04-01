package dht

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/libp2p/go-libp2p-core/routing"

	"go.opencensus.io/tag"
	"golang.org/x/xerrors"

	proto "github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log"
	goprocess "github.com/jbenet/goprocess"
	goprocessctx "github.com/jbenet/goprocess/context"
	kb "github.com/libp2p/go-libp2p-kbucket"
	mpb "github.com/memoio/go-mefs/pb"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	"github.com/memoio/go-mefs/source/go-libp2p-kad-dht/metrics"
	opts "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/opts"
	pb "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/pb"
	providers "github.com/memoio/go-mefs/source/go-libp2p-kad-dht/providers"
	"github.com/memoio/go-mefs/source/instance"
	"github.com/memoio/go-mefs/utils/metainfo"
)

var logger = logging.Logger("dht")

// NumBootstrapQueries defines the number of random dht queries to do to
// collect members of the routing table.
const NumBootstrapQueries = 5

// KadDHT is an implementation of Kademlia with S/Kademlia modifications.
// It is used to implement the base Routing module.
type KadDHT struct {
	host      host.Host           // the network services we need
	self      peer.ID             // Local peer (yourself)
	peerstore peerstore.Peerstore // Peer Registry

	datastore ds.Datastore // Local data

	routingTable *kb.RoutingTable // Array of routing tables for differently distanced nodes
	providers    *providers.ProviderManager

	birth time.Time // When this peer started up

	ctx  context.Context
	proc goprocess.Process

	strmap map[peer.ID]*messageSender
	smlk   sync.Mutex

	plk sync.Mutex

	protocols   []protocol.ID // DHT protocols
	metahandler instance.Service
}

// Assert that MEFS assumptions about interfaces aren't broken. These aren't a
// guarantee, but we can use them to aid refactoring.
var (
	_ routing.ContentRouting = (*KadDHT)(nil)
	_ routing.Routing        = (*KadDHT)(nil)
	_ routing.PeerRouting    = (*KadDHT)(nil)
	_ routing.PubKeyFetcher  = (*KadDHT)(nil)
	_ routing.ValueStore     = (*KadDHT)(nil)
)

// New creates a new DHT with the specified host and options.
func New(ctx context.Context, h host.Host, options ...opts.Option) (*KadDHT, error) {
	var cfg opts.Options
	if err := cfg.Apply(append([]opts.Option{opts.Defaults}, options...)...); err != nil {
		return nil, err
	}
	dht := makeDHT(ctx, h, cfg.Datastore, cfg.Protocols)

	// register for network notifs.
	dht.host.Network().Notify((*netNotifiee)(dht))

	dht.proc = goprocessctx.WithContextAndTeardown(ctx, func() error {
		// remove ourselves from network notifs.
		dht.host.Network().StopNotify((*netNotifiee)(dht))
		return nil
	})

	dht.proc.AddChild(dht.providers.Process())
	if !cfg.Client {
		for _, p := range cfg.Protocols {
			h.SetStreamHandler(p, dht.handleNewStream)
		}
	}

	return dht, nil
}

// NewDHT creates a new DHT object with the given peer as the 'local' host.
// KadDHT's initialized with this function will respond to DHT requests,
// whereas KadDHT's initialized with NewDHTClient will not.
func NewDHT(ctx context.Context, h host.Host, dstore ds.Batching) *KadDHT {
	dht, err := New(ctx, h, opts.Datastore(dstore))
	if err != nil {
		panic(err)
	}
	return dht
}

// NewDHTClient creates a new DHT object with the given peer as the 'local'
// host. KadDHT clients initialized with this function will not respond to DHT
// requests. If you need a peer to respond to DHT requests, use NewDHT instead.
// NewDHTClient creates a new DHT object with the given peer as the 'local' host
func NewDHTClient(ctx context.Context, h host.Host, dstore ds.Batching) *KadDHT {
	dht, err := New(ctx, h, opts.Datastore(dstore), opts.Client(true))
	if err != nil {
		panic(err)
	}
	return dht
}

func makeDHT(ctx context.Context, h host.Host, dstore ds.Batching, protocols []protocol.ID) *KadDHT {
	rt := kb.NewRoutingTable(KValue, kb.ConvertPeerID(h.ID()), time.Minute, h.Peerstore())

	cmgr := h.ConnManager()
	rt.PeerAdded = func(p peer.ID) {
		cmgr.TagPeer(p, "kbucket", 5)
	}
	rt.PeerRemoved = func(p peer.ID) {
		cmgr.UntagPeer(p, "kbucket")
	}

	dht := &KadDHT{
		datastore:    dstore,
		self:         h.ID(),
		peerstore:    h.Peerstore(),
		host:         h,
		strmap:       make(map[peer.ID]*messageSender),
		ctx:          ctx,
		providers:    providers.NewProviderManager(ctx, h.ID(), dstore),
		birth:        time.Now(),
		routingTable: rt,
		protocols:    protocols,
	}

	dht.ctx = dht.newContextWithLocalTags(ctx)

	return dht
}

// putValueToPeer stores the given key/value pair at the peer 'p'
func (dht *KadDHT) putValueToPeer(ctx context.Context, p peer.ID, rec *mpb.Record) error {
	pmes := pb.NewMessage(pb.Message_PUT_VALUE, rec.Key, 0)
	pmes.Record = rec
	rpmes, err := dht.sendRequest(ctx, p, pmes)
	if err != nil {
		logger.Debugf("putValueToPeer: %v. (peer: %s, key: %s)", err, p.Pretty(), loggableKey(string(rec.Key)))
		return err
	}

	if !bytes.Equal(rpmes.GetRecord().Value, pmes.GetRecord().Value) {
		logger.Warningf("putValueToPeer: value not put correctly. (%v != %v)", pmes, rpmes)
		return errors.New("value not put correctly")
	}

	return nil
}

var errInvalidRecord = errors.New("received invalid record")

// getValueOrPeers queries a particular peer p for the value for
// key. It returns either the value or a list of closer peers.
// NOTE: It will update the dht's peerstore with any new addresses
// it finds for the given peer.
func (dht *KadDHT) getValueOrPeers(ctx context.Context, p peer.ID, key string) (*mpb.Record, []*peer.AddrInfo, error) {
	pmes, err := dht.getValueSingle(ctx, p, key)
	if err != nil {
		return nil, nil, err
	}

	// Perhaps we were given closer peers
	peers := pb.PBPeersToPeerInfos(pmes.GetCloserPeers())

	if record := pmes.GetRecord(); record != nil {
		// Success! We were given the value
		logger.Debug("getValueOrPeers: got value")
		return record, peers, err
	}

	if len(peers) > 0 {
		logger.Debug("getValueOrPeers: peers")
		return nil, peers, nil
	}

	logger.Warning("getValueOrPeers: routing.ErrNotFound")
	return nil, nil, routing.ErrNotFound
}

// getValueSingle simply performs the get value RPC with the given parameters
func (dht *KadDHT) getValueSingle(ctx context.Context, p peer.ID, key string) (*pb.Message, error) {
	meta := logging.LoggableMap{
		"key":  key,
		"peer": p,
	}

	eip := logger.EventBegin(ctx, "getValueSingle", meta)
	defer eip.Done()

	pmes := new(pb.Message)
	pmes = pb.NewMessage(pb.Message_GET_VALUE, []byte(key), 0)

	bkey := strings.SplitN(key, metainfo.DELIMITER, 3)
	if len(bkey) == 3 {
		if bkey[1] == strconv.Itoa(int(mpb.KeyType_UserInit)) {
			pmes.OpType = int32(mpb.OpType_BroadCast)
		}
	}

	resp, err := dht.sendRequest(ctx, p, pmes)
	switch err {
	case nil:
		return resp, nil
	case ErrReadTimeout:
		logger.Warningf("getValueSingle: read timeout %s %s", p.Pretty(), key)
		fallthrough
	default:
		eip.SetError(err)
		return nil, err
	}
}

// getLocal attempts to retrieve the value from the datastore
func (dht *KadDHT) getLocal(key string) (*mpb.Record, error) {
	logger.Debugf("getLocal %s", key)
	rec, err := dht.getRecordFromDatastore(mkDsKey(key))
	if err != nil {
		logger.Warningf("getLocal: %s", err)
		return nil, err
	}

	// Double check the key. Can't hurt.
	if rec != nil && string(rec.GetKey()) != key {
		logger.Errorf("BUG getLocal: found a DHT record that didn't match it's key: %s != %s", rec.GetKey(), key)
		return nil, nil

	}
	return rec, nil
}

// putLocal stores the key value pair in the datastore
func (dht *KadDHT) putLocal(key string, rec *mpb.Record) error {
	logger.Debugf("putLocal: %v %v", key, rec)
	data, err := proto.Marshal(rec)
	if err != nil {
		logger.Warningf("putLocal: %s", err)
		return err
	}

	return dht.datastore.Put(mkDsKey(key), data)
}

// Update signals the routingTable to Update its last-seen status
// on the given peer.
func (dht *KadDHT) Update(ctx context.Context, p peer.ID) {
	logger.Event(ctx, "updatePeer", p)
	dht.routingTable.Update(p)
}

// FindLocal looks for a peer with a given ID connected to this dht and returns the peer and the table it was found in.
func (dht *KadDHT) FindLocal(id peer.ID) peer.AddrInfo {
	switch dht.host.Network().Connectedness(id) {
	case network.Connected, network.CanConnect:
		return dht.peerstore.PeerInfo(id)
	default:
		return peer.AddrInfo{}
	}
}

// findPeerSingle asks peer 'p' if they know where the peer with id 'id' is
func (dht *KadDHT) findPeerSingle(ctx context.Context, p peer.ID, id peer.ID) (*pb.Message, error) {
	eip := logger.EventBegin(ctx, "findPeerSingle",
		logging.LoggableMap{
			"peer":   p,
			"target": id,
		})
	defer eip.Done()

	pmes := pb.NewMessage(pb.Message_FIND_NODE, []byte(id), 0)
	resp, err := dht.sendRequest(ctx, p, pmes)
	switch err {
	case nil:
		return resp, nil
	case ErrReadTimeout:
		logger.Warningf("read timeout: %s %s", p.Pretty(), id)
		fallthrough
	default:
		eip.SetError(err)
		return nil, err
	}
}

func (dht *KadDHT) findProvidersSingle(ctx context.Context, p peer.ID, key cid.Cid) (*pb.Message, error) {
	eip := logger.EventBegin(ctx, "findProvidersSingle", p, key)
	defer eip.Done()

	pmes := pb.NewMessage(pb.Message_GET_PROVIDERS, key.Bytes(), 0)
	resp, err := dht.sendRequest(ctx, p, pmes)
	switch err {
	case nil:
		return resp, nil
	case ErrReadTimeout:
		logger.Warningf("read timeout: %s %s", p.Pretty(), key)
		fallthrough
	default:
		eip.SetError(err)
		return nil, err
	}
}

// nearestPeersToQuery returns the routing tables closest peers.
func (dht *KadDHT) nearestPeersToQuery(pmes *pb.Message, count int) []peer.ID {
	closer := dht.routingTable.NearestPeers(kb.ConvertKey(string(pmes.GetKey())), count)
	return closer
}

// betterPeersToQuery returns nearestPeersToQuery, but if and only if closer than self.
func (dht *KadDHT) betterPeersToQuery(pmes *pb.Message, p peer.ID, count int) []peer.ID {
	closer := dht.nearestPeersToQuery(pmes, count)

	// no node? nil
	if closer == nil {
		logger.Warning("betterPeersToQuery: no closer peers to send:", p)
		return nil
	}

	filtered := make([]peer.ID, 0, len(closer))
	for _, clp := range closer {

		// == to self? thats bad
		if clp == dht.self {
			logger.Error("BUG betterPeersToQuery: attempted to return self! this shouldn't happen...")
			return nil
		}
		// Dont send a peer back themselves
		if clp == p {
			continue
		}

		filtered = append(filtered, clp)
	}

	// ok seems like closer nodes
	return filtered
}

// Context return dht's context
func (dht *KadDHT) Context() context.Context {
	return dht.ctx
}

// Process return dht's process
func (dht *KadDHT) Process() goprocess.Process {
	return dht.proc
}

// RoutingTable return dht's routingTable
func (dht *KadDHT) RoutingTable() *kb.RoutingTable {
	return dht.routingTable
}

// Close calls Process Close
func (dht *KadDHT) Close() error {
	return dht.proc.Close()
}

func (dht *KadDHT) protocolStrs() []string {
	pstrs := make([]string, len(dht.protocols))
	for idx, proto := range dht.protocols {
		pstrs[idx] = string(proto)
	}

	return pstrs
}

func mkDsKey(s string) ds.Key {
	//return ds.NewKey("/data/" + base32.RawStdEncoding.EncodeToString([]byte(s)))
	return ds.NewKey(s)
}

func (dht *KadDHT) PeerID() peer.ID {
	return dht.self
}

func (dht *KadDHT) PeerKey() []byte {
	return kb.ConvertPeerID(dht.self)
}

func (dht *KadDHT) Host() host.Host {
	return dht.host
}

func (dht *KadDHT) Ping(ctx context.Context, p peer.ID) error {
	req := pb.NewMessage(pb.Message_PING, nil, 0)
	resp, err := dht.sendRequest(ctx, p, req)
	if err != nil {
		return xerrors.Errorf("sending request: %w", err)
	}
	if resp.Type != pb.Message_PING {
		return xerrors.Errorf("got unexpected response type: %v", resp.Type)
	}
	return nil
}

// newContextWithLocalTags returns a new context.Context with the InstanceID and
// PeerID keys populated. It will also take any extra tags that need adding to
// the context as tag.Mutators.
func (dht *KadDHT) newContextWithLocalTags(ctx context.Context, extraTags ...tag.Mutator) context.Context {
	extraTags = append(
		extraTags,
		tag.Upsert(metrics.KeyPeerID, dht.self.Pretty()),
		tag.Upsert(metrics.KeyInstanceID, fmt.Sprintf("%p", dht)),
	)
	ctx, _ = tag.New(
		ctx,
		extraTags...,
	) // ignoring error as it is unrelated to the actual function of this code.
	return ctx
}
