package storage

import (
	"context"
	"os"

	"github.com/crazyfrankie/goim/infra/contract/storage"
	"github.com/crazyfrankie/goim/infra/impl/storage/minio"
	"github.com/crazyfrankie/goim/types/consts"
)

type Storage = storage.Storage

func New(ctx context.Context) (Storage, error) {
	return minio.New(
		ctx,
		os.Getenv(consts.MinIOEndpoint),
		os.Getenv(consts.MinIOAK),
		os.Getenv(consts.MinIOSK),
		os.Getenv(consts.StorageBucket),
		false,
	)
}
