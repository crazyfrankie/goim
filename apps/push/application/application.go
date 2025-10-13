package application

import (
	"context"
	"os"

	push "github.com/crazyfrankie/goim/apps/push/domain/service"
	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	eventbusimpl "github.com/crazyfrankie/goim/infra/impl/eventbus"
	"github.com/crazyfrankie/goim/infra/impl/push/online"
	"github.com/crazyfrankie/goim/types/consts"
)

type BasicServices struct {
}

func Init(ctx context.Context, client discovery.SvcDiscoveryRegistry) (*BasicServices, error) {
	var basic *BasicServices
	nameServer := os.Getenv(consts.MQServer)

	eventbus.SetDefaultSVC(eventbusimpl.NewConsumerService())
	onlinePusher := online.NewOnlinePusher(client)
	pushHandler := push.NewMessagePushHandler(onlinePusher)

	if err := eventbusimpl.DefaultSVC().
		RegisterConsumer(nameServer, consts.RMQTopicMessage, consts.RMQConsumeGroupMessage, pushHandler); err != nil {
		return nil, err
	}

	return basic, nil
}
