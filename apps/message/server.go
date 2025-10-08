package message

import (
	"context"

	"google.golang.org/grpc"

	"github.com/crazyfrankie/goim/apps/message/application"
	"github.com/crazyfrankie/goim/apps/message/domain/repository"
	"github.com/crazyfrankie/goim/apps/message/domain/service"
	"github.com/crazyfrankie/goim/infra/contract/discovery"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
)

func Start(ctx context.Context, client discovery.SvcDiscoveryRegistry, srv grpc.ServiceRegistrar) error {
	basic, err := application.Init(ctx)
	if err != nil {
		return err
	}
	messageRepo := repository.NewMessageRepository(basic.DB)
	messageDomain := service.NewMessageDomain(&service.Components{
		MessageRepo: messageRepo,
		IDGen:       basic.IDGen,
	})
	appService := application.NewMessageApplicationService(messageDomain)

	messagev1.RegisterMessageServiceServer(srv, appService)

	return nil
}
