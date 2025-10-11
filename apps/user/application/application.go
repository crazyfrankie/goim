package application

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/infra/contract/idgen"
	"github.com/crazyfrankie/goim/infra/contract/storage"
	"github.com/crazyfrankie/goim/infra/impl/cache/redis"
	idgenimpl "github.com/crazyfrankie/goim/infra/impl/idgen"
	"github.com/crazyfrankie/goim/infra/impl/mysql"
	storageimpl "github.com/crazyfrankie/goim/infra/impl/storage"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type BasicServices struct {
	DB      *gorm.DB
	IDGen   idgen.IDGenerator
	IconOSS storage.Storage
	AuthCli authv1.AuthServiceClient
}

func Init(ctx context.Context, client discovery.SvcDiscoveryRegistry) (*BasicServices, error) {
	basic := &BasicServices{}
	var err error

	basic.DB, err = mysql.New("MYSQL_DSN")
	if err != nil {
		return nil, err
	}

	cacheCli := redis.New()

	basic.IDGen, err = idgenimpl.New(cacheCli)
	if err != nil {
		return nil, err
	}

	authCC, err := client.GetConn(ctx, consts.AuthServiceName)
	if err != nil {
		return nil, err
	}

	basic.AuthCli = authv1.NewAuthServiceClient(authCC)

	basic.IconOSS, err = storageimpl.New(ctx)
	if err != nil {
		return nil, err
	}

	return basic, nil
}
