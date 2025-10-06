package application

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/infra/contract/idgen"
	"github.com/crazyfrankie/goim/infra/contract/storage"
	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/infra/impl/cache/redis"
	idgenimpl "github.com/crazyfrankie/goim/infra/impl/idgen"
	"github.com/crazyfrankie/goim/infra/impl/mysql"
	storageimpl "github.com/crazyfrankie/goim/infra/impl/storage"
	tokenimpl "github.com/crazyfrankie/goim/infra/impl/token"
)

type BasicServices struct {
	DB       *gorm.DB
	IDGen    idgen.IDGenerator
	IconOSS  storage.Storage
	TokenGen token.Token
}

func Init(ctx context.Context) (*BasicServices, error) {
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
