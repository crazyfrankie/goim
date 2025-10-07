package startrpc

import (
	"context"
	"net"
	"time"

	"github.com/oklog/run"
	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func Start(ctx context.Context, listenIP, registerIP, listenPort, rpcRegisterName string,
	rpcStart func(ctx context.Context, srv grpc.ServiceRegistrar) error, opts ...grpc.ServerOption) error {

	g := &run.Group{}

	// Signal handler
	{
		ctx, cancel := context.WithCancel(ctx)
		g.Add(func() error {
			return signal.CtxWaitExit(ctx)
		}, func(err error) {
			cancel()
		})
	}

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
			// TODO
			// Register service

			// Start serving
			return rpcServer.Serve(rpcListener)
		}, func(err error) {
			if rpcServer != nil {
				// Graceful stop with timeout
				stopped := make(chan struct{})
				go func() {
					rpcServer.GracefulStop()
					close(stopped)
				}()

				select {
				case <-stopped:
				case <-time.After(15 * time.Second):
					logs.CtxWarnf(ctx, "rcp graceful stop timeout")
				}
			}
			if rpcListener != nil {
				rpcListener.Close()
			}
		})
	}

	if err := rpcStart(ctx, &grpcServiceRegistrar{onRegisterService: onRegisterService}); err != nil {
		return err
	}

	// Run all services
	if err := g.Run(); err != nil {
		logs.CtxDebugf(ctx, "program interrupted, %v", err)
		return err
	}

	logs.CtxDebugf(ctx, "Server exited gracefully")

	return nil
}

type grpcServiceRegistrar struct {
	onRegisterService func(desc *grpc.ServiceDesc, impl any)
}

func (x *grpcServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl any) {
	x.onRegisterService(desc, impl)
}
