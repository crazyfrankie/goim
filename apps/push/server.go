package push

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/push/application"
	"github.com/crazyfrankie/goim/infra/contract/discovery"
	pushv1 "github.com/crazyfrankie/goim/protocol/push/v1"
)

func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry, srv grpc.ServiceRegistrar) error {
	_, err := application.Init(ctx, client)
	if err != nil {
		return err
	}
	appService := application.NewPushApplicationService()

	pushv1.RegisterPushServiceServer(srv, appService)

	return nil
}
