package rpc

import (
	"context"

	"google.golang.org/grpc"
	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/apps/user/application"
	"github.com/crazyfrankie/goim/apps/user/domain/repository"
	"github.com/crazyfrankie/goim/apps/user/domain/service"
	"github.com/crazyfrankie/goim/infra/contract/idgen"
	"github.com/crazyfrankie/goim/infra/contract/storage"
	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/infra/impl/cache/redis"
	idgenimpl "github.com/crazyfrankie/goim/infra/impl/idgen"
	"github.com/crazyfrankie/goim/infra/impl/mysql"
	storageimpl "github.com/crazyfrankie/goim/infra/impl/storage"
	tokenimpl "github.com/crazyfrankie/goim/infra/impl/token"
	"github.com/crazyfrankie/goim/pkg/grpc/interceptor"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

type basicServices struct {
	DB       *gorm.DB
	IDGen    idgen.IDGenerator
	IconOSS  storage.Storage
	TokenGen token.Token
}

func initBasicServices(ctx context.Context) (*basicServices, error) {
	basic := &basicServices{}
	var err error

	basic.DB, err = mysql.New()
	if err != nil {
		return nil, err
	}

	cacheCli := redis.New()

	basic.IDGen, err = idgenimpl.New(cacheCli)
	if err != nil {
		return nil, err
	}

	basic.TokenGen, err = tokenimpl.New(cacheCli)
	if err != nil {
		return nil, err
	}

	basic.IconOSS, err = storageimpl.New(ctx)
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
