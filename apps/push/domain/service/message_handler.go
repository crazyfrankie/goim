package service

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/infra/contract/push/online"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	"github.com/crazyfrankie/goim/pkg/logs"
	msgv1 "github.com/crazyfrankie/goim/protocol/msg/v1"
	pushv1 "github.com/crazyfrankie/goim/protocol/push/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessagePushHandler struct {
	onlinePusher online.OnlinePusher
}

func NewMessagePushHandler(onlinePusher online.OnlinePusher) *MessagePushHandler {
	return &MessagePushHandler{onlinePusher: onlinePusher}
}

func (h *MessagePushHandler) HandleMessage(ctx context.Context, msg *eventbus.Message) error {
	var event pushv1.PushMsgRequest
	if err := proto.Unmarshal(msg.Body, &event); err != nil {
		logs.Errorf("unmarshal msg event failed: %v", err)
		return err
	}

	switch event.GetMessage().GetSessionType() {
	case consts.SingleChatType:
		pushUserList := []string{conv.Int64ToStr(event.GetMessage().RecvID), conv.Int64ToStr(event.GetMessage().SendID)}

		return h.handleSingleChat(ctx, pushUserList, event.GetMessage())
	//case consts.GroupChatType:
	default:
		logs.Warnf("unknown session type: %d", event.GetMessage().GetSessionType())
		return nil
	}
}

func (h *MessagePushHandler) handleSingleChat(ctx context.Context, pushUserList []string, msg *msgv1.Message) error {
	// first online push

	logs.Infof("receive msg, %v", msg)

	//res, err := h.onlinePusher.GetConnsAndOnlinePush(ctx, msg, pushUserList)
	//if err != nil {
	//	return nil
	//}

	//logs.Infof("online push results, %v", res)

	// TODO offline push

	return nil
}
