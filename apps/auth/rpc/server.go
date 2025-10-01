package rpc

import (
	"context"

	"github.com/crazyfrankie/goim/apps/auth/application"
	"github.com/crazyfrankie/goim/apps/auth/domain/service"
	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/infra/impl/cache/redis"
	tokenimpl "github.com/crazyfrankie/goim/infra/impl/token"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
	"google.golang.org/grpc"
)

type basicServices struct {
	TokenGen token.Token
}

func initBasicServices(ctx context.Context) (*basicServices, error) {
	basic := &basicServices{}
	var err error

	cacheCli := redis.New()

	basic.TokenGen, err = tokenimpl.New(cacheCli)
	if err != nil {
		return nil, err
	}

	return basic, nil
}

func NewGRPCServer(ctx context.Context) (*grpc.Server, error) {
	basic, err := initBasicServices(ctx)
	if err != nil {
		return nil, err
	}
	authDomain := service.NewAuthDomain(&service.Components{
		TokenGen: basic.TokenGen,
	})
	appService := application.NewAuthApplicationService(authDomain)

	opts := gRPCServerOptions()

	srv := grpc.NewServer(opts...)
	authv1.RegisterAuthServiceServer(srv, appService)

	return srv, nil
}

func gRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.ResponseInterceptor()),
	}
}
