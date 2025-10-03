package eventbus

import (
	"fmt"
	"os"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/infra/impl/eventbus/kafka"
	"github.com/crazyfrankie/goim/infra/impl/eventbus/rmq"
	"github.com/crazyfrankie/goim/types/consts"
)

type (
	Producer        = eventbus.Producer
	ConsumerService = eventbus.ConsumerService
	ConsumerHandler = eventbus.ConsumerHandler
	ConsumerOpt     = eventbus.ConsumerOpt
	Message         = eventbus.Message
)

type consumerServiceImpl struct{}

func NewConsumerService() ConsumerService {
	return &consumerServiceImpl{}
}

func DefaultSVC() ConsumerService {
	return eventbus.GetDefaultSVC()
}

func (consumerServiceImpl) RegisterConsumer(nameServer, topic, group string, consumerHandler eventbus.ConsumerHandler, opts ...eventbus.ConsumerOpt) error {
	tp := os.Getenv(consts.MQTypeKey)
	switch tp {
	case "kafka":
		return kafka.RegisterConsumer(nameServer, topic, group, consumerHandler, opts...)
	case "rmq":
		return rmq.RegisterConsumer(nameServer, topic, group, consumerHandler, opts...)
	}

	return fmt.Errorf("invalid mq type: %s , only support nsq, kafka, rmq", tp)
}

func NewProducer(nameServer, topic, group string, retries int) (eventbus.Producer, error) {
	tp := os.Getenv(consts.MQTypeKey)
	switch tp {
	case "kafka":
		return kafka.NewProducer(nameServer, topic)
	case "rmq":
		return rmq.NewProducer(nameServer, topic, group, retries)
	}

	return nil, fmt.Errorf("invalid mq type: %s , only support nsq, kafka, rmq", tp)
}
