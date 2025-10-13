package online

import (
	"context"

	msgv1 "github.com/crazyfrankie/goim/protocol/msg/v1"
	msggatewayv1 "github.com/crazyfrankie/goim/protocol/msggateway/v1"
)

type OnlinePusher interface {
	GetConnsAndOnlinePush(ctx context.Context, msg *msgv1.Message, pushToUserIDs []string) ([]*msggatewayv1.SingleMsgToUserResults, error)
	GetOnlinePushFailedUserIDs(ctx context.Context, msg *msgv1.Message, wsResults []*msggatewayv1.SingleMsgToUserResults, pushToUserIDs *[]string) []string
}
