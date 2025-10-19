package starthttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/run"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	discoveryimpl "github.com/crazyfrankie/goim/infra/impl/discovery"
	"github.com/crazyfrankie/goim/pkg/gin/middleware"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/metrics"
	"github.com/crazyfrankie/goim/pkg/tracing"
)

func init() {
	metrics.RegisterBFF()
}

type Config struct {
	ListenAddr      string
	ServiceName     string
	ServiceVer      string
	ShutdownTimeout time.Duration

	MetricAddr    string
	CollectorAddr string

	InitFunc func(ctx context.Context, client discovery.SvcDiscoveryRegistry, middlewares ...gin.HandlerFunc) (http.Handler, error)
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

	middlewares := []gin.HandlerFunc{
		middleware.Metric(),
	}

	if cfg.CollectorAddr != "" {
		traceProvider, err := tracing.GetTraceProvider(cfg.ServiceName, cfg.ServiceVer, cfg.CollectorAddr)
		if err != nil {
			return err
		}

		otel.SetTracerProvider(traceProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))

		middlewares = append([]gin.HandlerFunc{middleware.Trace(cfg.ServiceName)}, middlewares...)
	}

	engine, err := cfg.InitFunc(ctx, client, middlewares...)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: engine,
	}

	g.Add(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed: %w", err)
		}
		return nil
	}, func(err error) {
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logs.Errorf("failed to shutdown main server: %v", err)
		}
		logs.Infof("Server shutdown successfully")
	})

	if err := g.Run(); err != nil {
		logs.Infof("program interrupted, %v", err)
		return err
	}

	logs.Infof("Server exited gracefully")

	return nil
}
