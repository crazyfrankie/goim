package eventbus

import "context"

type Producer interface {
	Send(ctx context.Context, body []byte, opts ...SendOpt) error
	BatchSend(ctx context.Context, bodyArr [][]byte, opts ...SendOpt) error
}

type Message struct {
	Topic string
	Group string
	Body  []byte
}

type ConsumerHandler interface {
	HandleMessage(ctx context.Context, msg *Message) error
}

var defaultSVC ConsumerService

func SetDefaultSVC(svc ConsumerService) {
	defaultSVC = svc
}

func GetDefaultSVC() ConsumerService {
	return defaultSVC
}

type ConsumerService interface {
	RegisterConsumer(nameServer, topic, group string, consumerHandler ConsumerHandler, opts ...ConsumerOpt) error
}
