package starthttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/oklog/run"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	discoveryimpl "github.com/crazyfrankie/goim/infra/impl/discovery"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func Start(ctx context.Context, listenAddr string, initFn func(ctx context.Context, client discovery.SvcDiscoveryRegistry) (http.Handler, error), shutdownTimeout time.Duration) error {
	client, err := discoveryimpl.NewDiscoveryRegister()
	if err != nil {
		return err
	}

	g := &run.Group{}

	// Signal handler
	g.Add(func() error {
		return signal.CtxWaitExit(ctx)
	}, func(err error) {

	})

	engine, err := initFn(ctx, client)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: engine,
	}

	g.Add(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server failed: %w", err)
		}
		return nil
	}, func(err error) {
		shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
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
