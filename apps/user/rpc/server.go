package rpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/user/application"
	"github.com/crazyfrankie/goim/apps/user/domain/repository"
	"github.com/crazyfrankie/goim/apps/user/domain/service"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

func NewGRPCServer(ctx context.Context) (*grpc.Server, error) {
	basic, err := application.Init(ctx)
	if err != nil {
		return nil, err
	}
	userRepo := repository.NewUserRepository(basic.DB)
	userDomain := service.NewUserDomain(&service.Components{
		UserRepo: userRepo,
		IDGen:    basic.IDGen,
		IconOSS:  basic.IconOSS,
		TokenGen: basic.TokenGen,
	})
	appService := application.NewUserApplicationService(userDomain)

	opts := gRPCServerOptions()

	srv := grpc.NewServer(opts...)
	userv1.RegisterUserServiceServer(srv, appService)

	return srv, nil
}

func gRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.ResponseInterceptor()),
	}
}
