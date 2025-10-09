package startrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/oklog/run"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	discoveryimpl "github.com/crazyfrankie/goim/infra/impl/discovery"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func Start(ctx context.Context, listenIP, registerIP, listenPort, rpcRegisterName string,
	rpcStart func(ctx context.Context, client discovery.SvcDiscoveryRegistry, srv grpc.ServiceRegistrar) error,
	opts ...grpc.ServerOption) error {

	client, err := discoveryimpl.NewDiscoveryRegister()
	if err != nil {
		return err
	}
	defer client.Close()

	g := &run.Group{}

	// Signal handler
	g.Add(func() error {
		return signal.CtxWaitExit(ctx)
	}, func(err error) {

	})

	// Prometheus metrics server
	// TODO

	// RPC server
	var (
		rpcServer   *grpc.Server
		rpcListener net.Listener
	)

	onRegisterService := func(desc *grpc.ServiceDesc, impl any) {
		if rpcServer != nil {
			rpcServer.RegisterService(desc, impl)
			return
		}

		rpcListenAddr := net.JoinHostPort(listenIP, listenPort)

		var err error
		rpcListener, err = net.Listen("tcp", rpcListenAddr)
		if err != nil {
			logs.CtxErrorf(ctx, "listen rpc failed, rpcRegisterName: %s, rpcListenAddr: %s", rpcRegisterName, rpcListenAddr)
			return
		}

		rpcServer = grpc.NewServer(opts...)
		rpcServer.RegisterService(desc, impl)
		logs.CtxDebugf(ctx, "rpc start register, rpcRegisterName: %s, registerIP: %s, listenPort: %s", rpcRegisterName, registerIP, listenPort)

		g.Add(func() error {
			// Register service
			if err := client.Register(ctx, rpcRegisterName, registerIP, listenPort); err != nil {
				return fmt.Errorf("rpc register %s: %w", rpcRegisterName, err)
			}

			// Start serving
			return rpcServer.Serve(rpcListener)
		}, func(err error) {
			if rpcServer != nil {
				// Graceful stop with timeout
				stopCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
				defer cancel()

				done := make(chan struct{})
				go func() {
					rpcServer.GracefulStop()
					close(done)
				}()

				select {
				case <-done:
					logs.CtxInfof(ctx, "gRPC server stopped gracefully")
				case <-stopCtx.Done():
					logs.CtxWarnf(ctx, "gRPC server graceful stop timeout, forcing shutdown")
					rpcServer.Stop()
				}

				if rpcListener != nil {
					rpcListener.Close()
				}
			}
			if rpcListener != nil {
				rpcListener.Close()
			}
		})
	}

	if err := rpcStart(ctx, client, &grpcServiceRegistrar{onRegisterService: onRegisterService}); err != nil {
		return err
	}

	// Run all services
	if err := g.Run(); err != nil {
		logs.Infof("program interrupted, %v", err)
		return err
	}

	logs.Infof("Server exited gracefully")

	return nil
}

type grpcServiceRegistrar struct {
	onRegisterService func(desc *grpc.ServiceDesc, impl any)
}

func (x *grpcServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl any) {
	x.onRegisterService(desc, impl)
}
