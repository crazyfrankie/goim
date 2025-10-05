package service

import (
	"context"
	"time"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/internal/events/message"
	"github.com/crazyfrankie/goim/pkg/sonic"
)

type messageEventPublisher struct {
	producer eventbus.Producer
}

func NewMessageEventPublisher(producer eventbus.Producer) message.PublishEventBus {
	return &messageEventPublisher{
		producer: producer,
	}
}

func (p *messageEventPublisher) PublishMessageEvent(ctx context.Context, event *message.MessageEvent) error {
	if event.Meta == nil {
		event.Meta = &message.EventMeta{}
	}
	event.Meta.SendTimeMs = time.Now().UnixMilli()

	bytes, err := sonic.Marshal(event)
	if err != nil {
		return err
	}

	return p.producer.Send(ctx, bytes)
}
