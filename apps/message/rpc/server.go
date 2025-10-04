package rpc

import (
	"context"

	"google.golang.org/grpc"
	
	"github.com/crazyfrankie/goim/apps/message/application"
	"github.com/crazyfrankie/goim/apps/message/domain/repository"
	"github.com/crazyfrankie/goim/apps/message/domain/service"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
)

func NewGRPCServer(ctx context.Context) (*grpc.Server, error) {
	basic, err := application.Init(ctx)
	if err != nil {
		return nil, err
	}
	messageRepo := repository.NewMessageRepository(basic.DB)
	messageDomain := service.NewMessageDomain(&service.Components{
		MessageRepo: messageRepo,
		IDGen:       basic.IDGen,
	})
	appService := application.NewMessageApplicationService(messageDomain)

	opts := gRPCServerOptions()

	srv := grpc.NewServer(opts...)
	messagev1.RegisterMessageServiceServer(srv, appService)

	return srv, nil
}

func gRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.ResponseInterceptor()),
	}
}
