package application

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/infra/contract/idgen"
	"github.com/crazyfrankie/goim/infra/impl/cache/redis"
	idgenimpl "github.com/crazyfrankie/goim/infra/impl/idgen"
	"github.com/crazyfrankie/goim/infra/impl/mysql"
)

type BasicServices struct {
	DB    *gorm.DB
	IDGen idgen.IDGenerator
}

func Init(ctx context.Context) (*BasicServices, error) {
	basic := &BasicServices{}
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

	return basic, nil
}
