package user

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/user/application"
	"github.com/crazyfrankie/goim/apps/user/domain/repository"
	"github.com/crazyfrankie/goim/apps/user/domain/service"
	"github.com/crazyfrankie/goim/infra/contract/discovery"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry, srv grpc.ServiceRegistrar) error {
	basic, err := application.Init(ctx)
	if err != nil {
		return err
	}
	userRepo := repository.NewUserRepository(basic.DB)
	userDomain := service.NewUserDomain(&service.Components{
		UserRepo: userRepo,
		IDGen:    basic.IDGen,
		IconOSS:  basic.IconOSS,
		TokenGen: basic.TokenGen,
	})
	appService := application.NewUserApplicationService(userDomain)

	userv1.RegisterUserServiceServer(srv, appService)

	return nil
}
