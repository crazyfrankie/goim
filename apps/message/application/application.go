package application

import (
	"context"
	"fmt"
	"os"

	"github.com/crazyfrankie/goim/infra/contract/discovery"
	"gorm.io/gorm"

	message "github.com/crazyfrankie/goim/apps/message/domain/service"
	"github.com/crazyfrankie/goim/infra/contract/idgen"
	"github.com/crazyfrankie/goim/infra/impl/cache/redis"
	"github.com/crazyfrankie/goim/infra/impl/eventbus"
	idgenimpl "github.com/crazyfrankie/goim/infra/impl/idgen"
	"github.com/crazyfrankie/goim/infra/impl/mysql"
	messageevent "github.com/crazyfrankie/goim/internal/events/message"
	"github.com/crazyfrankie/goim/types/consts"
)

type BasicServices struct {
	DB              *gorm.DB
	IDGen           idgen.IDGenerator
	MessageEventBus messageevent.PublishEventBus
}

func Init(ctx context.Context, client discovery.SvcDiscoveryRegistry) (*BasicServices, error) {
	basic := &BasicServices{}
	var err error

	basic.DB, err = mysql.New("TIDB_DSN")
	if err != nil {
		return nil, err
	}

	cacheCli := redis.New()

	basic.IDGen, err = idgenimpl.New(cacheCli)
	if err != nil {
		return nil, err
	}

	appEventProducer, err := initAppEventProducer()
	if err != nil {
		return nil, err
	}

	basic.MessageEventBus = message.NewMessageEventPublisher(appEventProducer)

	return basic, nil
}

func initAppEventProducer() (eventbus.Producer, error) {
	nameServer := os.Getenv(consts.MQServer)
	messageEventProducer, err := eventbus.NewProducer(nameServer, consts.RMQTopicMessage, consts.RMQConsumeGroupMessage, 1)
	if err != nil {
		return nil, fmt.Errorf("init message producer failed, err=%w", err)
	}

	return messageEventProducer, nil
}
