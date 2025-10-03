package rpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/auth/application"
	"github.com/crazyfrankie/goim/apps/auth/domain/service"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
)

func NewGRPCServer(ctx context.Context) (*grpc.Server, error) {
	basic, err := application.Init(ctx)
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
