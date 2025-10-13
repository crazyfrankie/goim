package message

import (
	"context"

	pushv1 "github.com/crazyfrankie/goim/protocol/push/v1"
)

type PushEventBus interface {
	PushMessageEvent(ctx context.Context, event *pushv1.PushMsgRequest) error
}
