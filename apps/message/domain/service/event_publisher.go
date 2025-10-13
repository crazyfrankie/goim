package service

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/internal/events/message"
	pushv1 "github.com/crazyfrankie/goim/protocol/push/v1"
)

type messageEventPusher struct {
	producer eventbus.Producer
}

func NewMessageEventPusher(producer eventbus.Producer) message.PushEventBus {
	return &messageEventPusher{
		producer: producer,
	}
}

func (p *messageEventPusher) PushMessageEvent(ctx context.Context, event *pushv1.PushMsgRequest) error {
	bytes, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return p.producer.Send(ctx, bytes)
}
