package startrpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/crazyfrankie/goim/pkg/tracing"
	"github.com/oklog/run"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	discoveryimpl "github.com/crazyfrankie/goim/infra/impl/discovery"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/metrics"
)

type Config struct {
	ListenIP        string
	ListenPort      string
	RegisterIP      string
	RPCRegisterName string
	RPCServiceVer   string

	MetricAddr    string
	CollectorAddr string

	RPCStart   func(ctx context.Context, client discovery.SvcDiscoveryRegistry, srv grpc.ServiceRegistrar) error
	ServerOpts []grpc.ServerOption
}

func Start(ctx context.Context, cfg *Config) error {
	client, err := discoveryimpl.NewDiscoveryRegister()
	if err != nil {
		return err
	}
	defer client.Close()

	client.AppendOption(
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, "round_robin")),
		grpc.WithChainUnaryInterceptor(interceptor.ClientLogInterceptor()),
	)

	g := &run.Group{}

	// Signal handler
	g.Add(func() error {
		return signal.CtxWaitExit(ctx)
	}, func(err error) {

	})

	// Prometheus metrics server
	if cfg.MetricAddr != "" {
		g.Add(func() error {
			listener, err := net.Listen("tcp", cfg.MetricAddr)
			if err != nil {
				return err
			}

			return metrics.Start(listener)
		}, func(err error) {

		})
	}

	// RPC server
	var (
		rpcServer   *grpc.Server
		rpcListener net.Listener
	)

	if cfg.CollectorAddr != "" {
		traceProvider, err := tracing.GetTraceProvider(cfg.RPCRegisterName, cfg.RPCServiceVer, cfg.CollectorAddr)
		if err != nil {
			return err
		}
		otel.SetTracerProvider(traceProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
	}

	onRegisterService := func(desc *grpc.ServiceDesc, impl any) {
		if rpcServer != nil {
			rpcServer.RegisterService(desc, impl)
			return
		}

		rpcListenAddr := net.JoinHostPort(cfg.ListenIP, cfg.ListenPort)

		var err error
		rpcListener, err = net.Listen("tcp", rpcListenAddr)
		if err != nil {
			logs.CtxErrorf(ctx, "listen rpc failed, rpcRegisterName: %s, rpcListenAddr: %s", cfg.RPCRegisterName, rpcListenAddr)
			return
		}

		rpcServer = grpc.NewServer(cfg.ServerOpts...)
		rpcServer.RegisterService(desc, impl)
		logs.CtxDebugf(ctx, "rpc start register, rpcRegisterName: %s, registerIP: %s, listenPort: %s", cfg.RPCRegisterName, cfg.RegisterIP, cfg.ListenPort)

		g.Add(func() error {
			// Register service
			if err := client.Register(ctx, cfg.RPCRegisterName, cfg.RegisterIP, cfg.ListenPort); err != nil {
				return fmt.Errorf("rpc register %s: %w", cfg.RPCRegisterName, err)
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

	if err := cfg.RPCStart(ctx, client, &grpcServiceRegistrar{onRegisterService: onRegisterService}); err != nil {
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
