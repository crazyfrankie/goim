package message

import "context"

type PublishEventBus interface {
	PublishMessageEvent(ctx context.Context, event *MessageEvent) error
}
