package auth

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/auth/application"
	"github.com/crazyfrankie/goim/apps/auth/domain/service"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
)

func Start(ctx context.Context, srv grpc.ServiceRegistrar) error {
	basic, err := application.Init(ctx)
	if err != nil {
		return err
	}
	authDomain := service.NewAuthDomain(&service.Components{
		TokenGen: basic.TokenGen,
	})
	appService := application.NewAuthApplicationService(authDomain)

	authv1.RegisterAuthServiceServer(srv, appService)

	return nil
}
