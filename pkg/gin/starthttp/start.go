package starthttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/oklog/run"
	
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func Start(ctx context.Context, listenAddr string, initFn func() (http.Handler, error), shutdownTimeout time.Duration) error {
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

	engine, err := initFn()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: engine,
	}

	g.Add(func() error {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}, func(err error) {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logs.Errorf("failed to shutdown main server: %v", err)
		}
	})

	return g.Run()
}
