package application

import (
	"context"

	message "github.com/crazyfrankie/goim/apps/message/domain/service"
	eventbus "github.com/crazyfrankie/goim/internal/events/message"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
)

type MessageApplicationService struct {
	messageDomain   message.Message
	messageEventBus eventbus.PublishEventBus
	messagev1.UnimplementedMessageServiceServer
}

func NewMessageApplicationService(messageDomain message.Message) *MessageApplicationService {
	return &MessageApplicationService{messageDomain: messageDomain}
}

func (m *MessageApplicationService) SendMessage(ctx context.Context, req *messagev1.SendMessageReq) (*messagev1.SendMessageResponse, error) {
	panic("implement me")
}
