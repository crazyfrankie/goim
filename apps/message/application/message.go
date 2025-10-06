package application

import (
	"context"
	"errors"

	"github.com/crazyfrankie/goim/apps/message/domain/entity"
	message "github.com/crazyfrankie/goim/apps/message/domain/service"
	eventbus "github.com/crazyfrankie/goim/internal/events/message"
	"github.com/crazyfrankie/goim/pkg/grpc/ctxutil"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessageApplicationService struct {
	messageDomain   message.Message
	messageEventBus eventbus.PublishEventBus
	messagev1.UnimplementedMessageServiceServer
}

func NewMessageApplicationService(messageDomain message.Message) *MessageApplicationService {
	return &MessageApplicationService{messageDomain: messageDomain}
}

func (m *MessageApplicationService) SendMessage(ctx context.Context, req *messagev1.SendMessageRequest) (*messagev1.SendMessageResponse, error) {
	if err := ctxutil.CheckAccess(ctx, req.SendID); err != nil {
		return nil, err
	}

	msg, err := m.messageDomain.Create(ctx, &message.CreateMessageRequest{
		SendID:      req.GetSendID(),
		RecvID:      req.GetRecvID(),
		GroupID:     req.GetGroupID(),
		ClientMsgID: req.GetClientMsgID(),
		Content:     string(req.GetContent()),
		SessionType: req.GetSessionType(),
		MessageFrom: req.GetMessageFrom(),
		ContentType: req.GetContentType(),
		// TODO add seq generated
		//Seq:       ,
		SendTime: req.GetSendTime(),
	})
	if err != nil {
		return nil, err
	}

	switch req.GetSessionType() {
	case consts.SingleChatType:
		return m.sendSingleChat(ctx, msg)
	case consts.GroupChatType:
		return m.sendGroupChat(ctx, msg)
	case consts.NotificationChatType:
		return m.sendNotificationChat(ctx, msg)
	default:
		return nil, errors.New("unsupported session type")
	}
}

func (m *MessageApplicationService) sendSingleChat(ctx context.Context, msg *entity.Message) (*messagev1.SendMessageResponse, error) {
	// TODO

	return &messagev1.SendMessageResponse{
		SendTime:    msg.SendTime,
		ServerMsgID: msg.MsgID,
		ClientMsgID: msg.ClientMsgID,
	}, nil
}

func (m *MessageApplicationService) sendGroupChat(ctx context.Context, msg *entity.Message) (*messagev1.SendMessageResponse, error) {
	// TODO

	return &messagev1.SendMessageResponse{
		SendTime:    msg.SendTime,
		ServerMsgID: msg.MsgID,
		ClientMsgID: msg.ClientMsgID,
	}, nil
}

func (m *MessageApplicationService) sendNotificationChat(ctx context.Context, msg *entity.Message) (*messagev1.SendMessageResponse, error) {
	// TODO

	return &messagev1.SendMessageResponse{
		SendTime:    msg.SendTime,
		ServerMsgID: msg.MsgID,
		ClientMsgID: msg.ClientMsgID,
	}, nil
}

func (m *MessageApplicationService) SetMessageStatus(ctx context.Context, req *messagev1.SetMessageStatusRequest) (*messagev1.SetMessageStatusResponse, error) {
	err := m.messageDomain.UpdateMessageStatus(ctx, req.GetStatus())
	if err != nil {
		return nil, err
	}

	return &messagev1.SetMessageStatusResponse{}, nil
}
