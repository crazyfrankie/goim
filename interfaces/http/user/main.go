package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/oklog/run"

	"github.com/crazyfrankie/goim/interfaces/http/user/api"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	engine, err := api.InitEngine()
	if err != nil {
		panic(err)
	}

	initLogger()

	addr := os.Getenv("SERVER_ADDR")
	srv := &http.Server{
		Handler: engine,
		Addr:    addr,
	}
	g := &run.Group{}

	g.Add(func() error {
		logs.Infof("http bff running at, %s", addr)
		return srv.ListenAndServe()
	}, func(err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logs.Errorf("failed to shutdown main server: %v", err)
		}
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		logs.Warnf("all programs down")
	}
}

func initLogger() {
	logger := logs.NewLogger(os.Stdout)
	logger = logs.With(logger, "bff", "http")
	logs.SetGlobalLogger(logger)
	setLogLevel(logger)
}

func setLogLevel(logger logs.FullLogger) {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))

	logs.Infof("[log level: %s]", level)
	switch level {
	case "trace":
		logger.SetLevel(logs.LevelTrace)
	case "debug":
		logger.SetLevel(logs.LevelDebug)
	case "info":
		logger.SetLevel(logs.LevelInfo)
	case "notice":
		logger.SetLevel(logs.LevelNotice)
	case "warn":
		logger.SetLevel(logs.LevelWarn)
	case "error":
		logger.SetLevel(logs.LevelError)
	case "fatal":
		logger.SetLevel(logs.LevelFatal)
	default:
		logger.SetLevel(logs.LevelInfo)
	}
}
