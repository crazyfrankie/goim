package main

import (
	"context"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/oklog/run"

	"github.com/crazyfrankie/goim/apps/user/rpc"
	"github.com/crazyfrankie/goim/pkg/logs"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	srv, err := rpc.NewGRPCServer(context.Background())
	if err != nil {
		panic(err)
	}

	initLogger()

	g := &run.Group{}

	g.Add(func() error {
		addr := os.Getenv("SERVER_ADDR")
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		return srv.Serve(listener)
	}, func(err error) {
		srv.GracefulStop()
		srv.Stop()
	})

	g.Add(run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM))

	if err := g.Run(); err != nil {
		logs.Warnf("all programs down")
	}
}

func initLogger() {
	logger := logs.NewLogger(os.Stdout)
	logger = logs.With(logger, "service.name", "user")
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
