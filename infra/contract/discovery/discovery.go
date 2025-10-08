package discovery

import (
	"context"

	"google.golang.org/grpc"
)

type Conn interface {
	GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (grpc.ClientConnInterface, error)
	GetConns(ctx context.Context, serviceName string, opts ...grpc.DialOption) ([]grpc.ClientConnInterface, error)
	IsSelfNode(cc grpc.ClientConnInterface) bool
}

type WatchKey struct {
	Value []byte
}

type WatchKeyHandler func(data *WatchKey) error

type KeyValue interface {
	SetKey(ctx context.Context, key string, value []byte) error
	SetWithLease(ctx context.Context, key string, val []byte, ttl int64) error
	GetKey(ctx context.Context, key string) ([]byte, error)
	GetKeyWithPrefix(ctx context.Context, key string) ([][]byte, error)
	WatchKey(ctx context.Context, key string, fn WatchKeyHandler) error
}

type SvcDiscoveryRegistry interface {
	Conn
	KeyValue
	AppendOption(opts ...grpc.DialOption)
	Register(ctx context.Context, serviceName string, host, port string) error
	Close()
}
