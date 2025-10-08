package etcd

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	gresolver "google.golang.org/grpc/resolver"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	"github.com/crazyfrankie/goim/pkg/lang/slice"
	"github.com/crazyfrankie/goim/pkg/logs"
)

type addrConn struct {
	conn        *grpc.ClientConn
	addr        string
	isConnected bool
}

type registryEtcdImpl struct {
	client        *clientv3.Client
	resolver      gresolver.Builder
	epManager     endpoints.Manager
	leaseID       clientv3.LeaseID
	dialOptions   []grpc.DialOption
	endpoint      endpoints.Endpoint
	serviceKey    string
	rootDirectory string
	watchNames    []string

	mu      sync.RWMutex
	connMap map[string][]*addrConn
}

func NewSvcDiscoveryRegistry(rootDirectory string, endpoints []string, watchNames []string,
	opts ...Option) (discovery.SvcDiscoveryRegistry, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
		// Increase keep-alive queue capacity and message size
		PermitWithoutStream: true,
		MaxCallSendMsgSize:  10 * 1024 * 1024, // 10 MB
	}

	// Apply provided options to the config
	for _, opt := range opts {
		opt(&cfg)
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	rl, err := resolver.NewBuilder(client)

	r := &registryEtcdImpl{
		client:        client,
		resolver:      rl,
		rootDirectory: rootDirectory,
		connMap:       make(map[string][]*addrConn),
		watchNames:    watchNames,
	}

	go r.watchServiceChanges()

	return r, nil
}

func (r *registryEtcdImpl) GetConn(ctx context.Context, serviceName string, opts ...grpc.DialOption) (grpc.ClientConnInterface, error) {
	target := fmt.Sprintf("etcd:///%s/%s", r.rootDirectory, serviceName)

	dialOpts := append(append(r.dialOptions, opts...), grpc.WithResolvers(r.resolver))

	return grpc.NewClient(target, dialOpts...)
}

func (r *registryEtcdImpl) GetConns(ctx context.Context, serviceName string, opts ...grpc.DialOption) ([]grpc.ClientConnInterface, error) {
	fullServiceKey := fmt.Sprintf("%s/%s", r.rootDirectory, serviceName)
	r.mu.RLock()
	if len(r.connMap) == 0 {
		r.mu.RUnlock()
		if err := r.initializeConnMap(opts...); err != nil {
			return nil, err
		}
		r.mu.RLock()
	}
	defer r.mu.RUnlock()

	return slice.Batch(func(t *addrConn) grpc.ClientConnInterface {
		return t.conn
	}, r.connMap[fullServiceKey]), nil
}

func (r *registryEtcdImpl) IsSelfNode(cc grpc.ClientConnInterface) bool {
	cli, ok := cc.(*grpc.ClientConn)
	if !ok {
		return false
	}
	return r.endpoint.Addr == cli.Target()
}

func (r *registryEtcdImpl) SetKey(ctx context.Context, key string, value []byte) error {
	if _, err := r.client.Put(ctx, key, string(value)); err != nil {
		return errorx.Wrapf(err, "etcd put err")
	}
	return nil
}

func (r *registryEtcdImpl) SetWithLease(ctx context.Context, key string, val []byte, ttl int64) error {
	leaseResp, err := r.client.Grant(ctx, ttl)
	if err != nil {
		return errorx.Wrapf(err, "etcd set with lease err")
	}

	if _, err := r.client.Put(ctx, key, conv.BytesToString(val), clientv3.WithLease(leaseResp.ID)); err != nil {
		return errorx.Wrapf(err, "etcd put with lease err")
	}

	return nil
}

func (r *registryEtcdImpl) GetKey(ctx context.Context, key string) ([]byte, error) {
	resp, err := r.client.Get(ctx, key)
	if err != nil {
		return nil, errorx.Wrapf(err, "etcd get err")
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	return resp.Kvs[0].Value, nil
}

func (r *registryEtcdImpl) GetKeyWithPrefix(ctx context.Context, key string) ([][]byte, error) {
	resp, err := r.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, errorx.Wrapf(err, "etcd get with prefix err")
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	return slice.Batch(func(kv *mvccpb.KeyValue) []byte {
		return kv.Value
	}, resp.Kvs), nil
}

func (r *registryEtcdImpl) WatchKey(ctx context.Context, key string, fn discovery.WatchKeyHandler) error {
	watchChan := r.client.Watch(ctx, key)
	for watchResp := range watchChan {
		for _, event := range watchResp.Events {
			if event.IsModify() && string(event.Kv.Key) == key {
				if err := fn(&discovery.WatchKey{Value: event.Kv.Value}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *registryEtcdImpl) AppendOption(opts ...grpc.DialOption) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resetConnMap()
	r.dialOptions = append(r.dialOptions, opts...)
}

func (r *registryEtcdImpl) Register(ctx context.Context, serviceName string, host, port string) error {
	r.serviceKey = fmt.Sprintf("etcd:///%s/%s", r.rootDirectory, serviceName)
	em, err := endpoints.NewManager(r.client, r.rootDirectory+"/"+serviceName)
	if err != nil {
		return err
	}
	r.epManager = em

	leaseResp, err := r.client.Grant(ctx, 30)
	if err != nil {
		return err
	}
	r.leaseID = leaseResp.ID

	addr := net.JoinHostPort(host, port)
	r.endpoint = endpoints.Endpoint{Addr: addr}

	err = r.epManager.AddEndpoint(ctx, r.serviceKey, r.endpoint, clientv3.WithLease(r.leaseID))
	if err != nil {
		return err
	}

	go r.keepAlive()

	return nil
}

func (r *registryEtcdImpl) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resetConnMap()
	if r.client != nil {
		_ = r.client.Close()
	}
}

func (r *registryEtcdImpl) initializeConnMap(opts ...grpc.DialOption) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ctx := context.Background()
	for _, name := range r.watchNames {
		fullPrefix := fmt.Sprintf("%s/%s", r.rootDirectory, name)
		resp, err := r.client.Get(ctx, fullPrefix, clientv3.WithPrefix())
		if err != nil {
			return err
		}

		oldList := r.connMap[fullPrefix]

		addrMap := make(map[string]*addrConn, len(oldList))
		for _, conn := range oldList {
			addrMap[conn.addr] = conn
		}
		newList := make([]*addrConn, 0, len(oldList))
		for _, kv := range resp.Kvs {
			prefix, addr := r.splitEndpoint(string(kv.Key))
			if addr == "" {
				continue
			}
			if _, _, err = net.SplitHostPort(addr); err != nil {
				continue
			}
			if prefix != fullPrefix {
				continue
			}

			if conn, ok := addrMap[addr]; ok {
				conn.isConnected = true
				continue
			}

			dialOpts := append(append(r.dialOptions, opts...), grpc.WithResolvers(r.resolver))

			conn, err := grpc.NewClient(addr, dialOpts...)
			if err != nil {
				continue
			}
			newList = append(newList, &addrConn{conn: conn, addr: addr, isConnected: false})
		}
		for _, conn := range oldList {
			if conn.isConnected {
				conn.isConnected = false
				newList = append(newList, conn)
				continue
			}
			if err = conn.conn.Close(); err != nil {
				logs.CtxWarnf(ctx, "close conn err, %v", err)
			}
		}
		r.connMap[fullPrefix] = newList
	}

	return nil
}

func (r *registryEtcdImpl) resetConnMap() {
	ctx := context.Background()
	for _, conn := range r.connMap {
		for _, c := range conn {
			if err := c.conn.Close(); err != nil {
				logs.CtxWarnf(ctx, "failed to close conn, err: %v", err)
			}
		}
	}
	r.connMap = make(map[string][]*addrConn)
}

func (r *registryEtcdImpl) keepAlive() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := r.client.KeepAlive(ctx, r.leaseID)
	if err != nil {
		logs.CtxFatalf(ctx, "KeepAlive failed: %v", err)
	}

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				logs.Info("KeepAlive channel closed")
				return
			}
			logs.Info("Lease renewed")
		case <-ctx.Done():
			return
		}
	}
}

// watchServiceChanges watches for changes in the service directory
func (r *registryEtcdImpl) watchServiceChanges() {
	for _, s := range r.watchNames {
		go func() {
			watchChan := r.client.Watch(context.Background(), r.rootDirectory+"/"+s, clientv3.WithPrefix())
			for range watchChan {
				if err := r.initializeConnMap(); err != nil {
					logs.Warnf("initializeConnMap in watch err, %v", err)
				}
			}
		}()
	}
}

// splitEndpoint splits the endpoint string into prefix and address
func (r *registryEtcdImpl) splitEndpoint(input string) (string, string) {
	lastSlashIndex := strings.LastIndex(input, "/")
	if lastSlashIndex != -1 {
		part1 := input[:lastSlashIndex]
		part2 := input[lastSlashIndex+1:]
		return part1, part2
	}
	return input, ""
}
