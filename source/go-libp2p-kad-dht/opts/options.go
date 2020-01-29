package dhtopts

import (
	"fmt"

	protocol "github.com/libp2p/go-libp2p-core/protocol"
	ds "github.com/memoio/go-mefs/source/go-datastore"
	dssync "github.com/memoio/go-mefs/source/go-datastore/sync"
)

var (
	ProtocolDHT      protocol.ID = "/bcs/kad/1.0.0"
	DefaultProtocols             = []protocol.ID{ProtocolDHT}
)

// Options is a structure containing all the options that can be used when constructing a DHT.
type Options struct {
	Datastore ds.Batching
	Client    bool
	Protocols []protocol.ID
}

// Apply applies the given options to this Option
func (o *Options) Apply(opts ...Option) error {
	for i, opt := range opts {
		if err := opt(o); err != nil {
			return fmt.Errorf("dht option %d failed: %s", i, err)
		}
	}
	return nil
}

// Option DHT option type.
type Option func(*Options) error

// Defaults are the default DHT options. This option will be automatically
// prepended to any options you pass to the DHT constructor.
var Defaults = func(o *Options) error {
	o.Datastore = dssync.MutexWrap(ds.NewMapDatastore())
	o.Protocols = DefaultProtocols
	return nil
}

// Datastore configures the DHT to use the specified datastore.
//
// Defaults to an in-memory (temporary) map.
func Datastore(ds ds.Batching) Option {
	return func(o *Options) error {
		o.Datastore = ds
		return nil
	}
}

// Client configures whether or not the DHT operates in client-only mode.
//
// Defaults to false.
func Client(only bool) Option {
	return func(o *Options) error {
		o.Client = only
		return nil
	}
}

// Validator configures the DHT to use the specified validator.
//
// Defaults to a namespaced validator that can only validate public keys.

// Protocols sets the protocols for the DHT
//
// Defaults to dht.DefaultProtocols
func Protocols(protocols ...protocol.ID) Option {
	return func(o *Options) error {
		o.Protocols = protocols
		return nil
	}
}
