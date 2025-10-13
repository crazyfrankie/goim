package application

import (
	"context"
	"errors"

	"github.com/crazyfrankie/goim/apps/message/domain/entity"
	message "github.com/crazyfrankie/goim/apps/message/domain/service"
	eventbus "github.com/crazyfrankie/goim/internal/events/message"
	"github.com/crazyfrankie/goim/pkg/grpc/ctxutil"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	msgv1 "github.com/crazyfrankie/goim/protocol/msg/v1"
	pushv1 "github.com/crazyfrankie/goim/protocol/push/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessageApplicationService struct {
	messageDomain   message.Message
	messageEventBus eventbus.PushEventBus
	msgv1.UnimplementedMessageServiceServer
}

func NewMessageApplicationService(messageDomain message.Message) *MessageApplicationService {
	return &MessageApplicationService{messageDomain: messageDomain}
}

func (m *MessageApplicationService) SendMessage(ctx context.Context, req *msgv1.SendMessageRequest) (*msgv1.SendMessageResponse, error) {
	if err := ctxutil.CheckAccess(ctx, req.GetData().GetSendID()); err != nil {
		return nil, err
	}

	msg, err := m.messageDomain.Create(ctx, &message.CreateMessageRequest{
		SendID:      req.GetData().GetSendID(),
		RecvID:      req.GetData().GetRecvID(),
		GroupID:     req.GetData().GetGroupID(),
		ClientMsgID: req.GetData().GetClientMsgID(),
		Content:     string(req.GetData().GetContent()),
		SessionType: req.GetData().GetSessionType(),
		MessageFrom: req.GetData().GetMessageFrom(),
		ContentType: req.GetData().GetContentType(),
		// TODO add seq generated
		//Seq:       ,
		SendTime: req.GetData().GetSendTime(),
	})
	if err != nil {
		return nil, err
	}

	switch req.GetData().GetSessionType() {
	case consts.SingleChatType:
		return m.sendSingleChat(ctx, req.GetData(), msg)
	case consts.GroupChatType:
		return m.sendGroupChat(ctx, msg)
	case consts.NotificationChatType:
		return m.sendNotificationChat(ctx, msg)
	default:
		return nil, errors.New("unsupported session type")
	}
}

func (m *MessageApplicationService) sendSingleChat(ctx context.Context, req *msgv1.Message, msg *entity.Message) (*msgv1.SendMessageResponse, error) {
	err := m.messageEventBus.PushMessageEvent(ctx, &pushv1.PushMsgRequest{
		Message:        req,
		ConversationID: "",
		UserIDs:        []string{conv.Int64ToStr(msg.RecvID)},
	})
	if err != nil {
		return nil, err
	}

	return &msgv1.SendMessageResponse{
		SendTime:    msg.SendTime,
		ServerMsgID: msg.MsgID,
		ClientMsgID: msg.ClientMsgID,
	}, nil
}

func (m *MessageApplicationService) sendGroupChat(ctx context.Context, msg *entity.Message) (*msgv1.SendMessageResponse, error) {
	// TODO

	return &msgv1.SendMessageResponse{
		SendTime:    msg.SendTime,
		ServerMsgID: msg.MsgID,
		ClientMsgID: msg.ClientMsgID,
	}, nil
}

func (m *MessageApplicationService) sendNotificationChat(ctx context.Context, msg *entity.Message) (*msgv1.SendMessageResponse, error) {
	// TODO

	return &msgv1.SendMessageResponse{
		SendTime:    msg.SendTime,
		ServerMsgID: msg.MsgID,
		ClientMsgID: msg.ClientMsgID,
	}, nil
}

func (m *MessageApplicationService) SetMessageStatus(ctx context.Context, req *msgv1.SetMessageStatusRequest) (*msgv1.SetMessageStatusResponse, error) {
	err := m.messageDomain.UpdateMessageStatus(ctx, req.GetStatus())
	if err != nil {
		return nil, err
	}

	return &msgv1.SetMessageStatusResponse{}, nil
}
