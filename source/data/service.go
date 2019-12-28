package data

import (
	"context"

	blocks "github.com/memoio/go-mefs/source/go-block-format"
)

// Service is for data
type Service interface {
	GetNetAddr() string
	GetKey(ctx context.Context, key, to string) ([]byte, error)
	PutKey(ctx context.Context, key string, data []byte, to string) error
	// AppendKey key is dtype/id/op1/op2
	AppendKey(ctx context.Context, key string, data []byte, to string) error
	DeleteKey(ctx context.Context, key, to string) error

	GetBlock(ctx context.Context, key string, sig []byte, to string) (blocks.Block, error)
	PutBlock(ctx context.Context, key string, data []byte, to string) error
	// AppendBlock key is dtype/id/op1/op2
	AppendBlock(ctx context.Context, key string, data []byte, to string) error
	DeleteBlock(ctx context.Context, key, to string) error

	SendMetaMessage(ctx context.Context, typ int32, key string, data, sig []byte, to string) error
	SendMetaRequest(ctx context.Context, typ int32, key string, data, sig []byte, to string) ([]byte, error)
	BroadcastMessage(ctx context.Context, key string) error
	Connect(ctx context.Context, to string) bool
	TestConnect() error
}
