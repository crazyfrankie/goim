package service

import (
	"context"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/internal/events/message"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/sonic"
)

type MessageEventHandler struct {
	// 推送相关的依赖
}

func NewMessageEventHandler() *MessageEventHandler {
	return &MessageEventHandler{}
}

func (h *MessageEventHandler) HandleMessage(ctx context.Context, msg *eventbus.Message) error {
	var event message.MessageEvent
	if err := sonic.Unmarshal(msg.Body, &event); err != nil {
		logs.Errorf("unmarshal message event failed: %v", err)
		return err
	}

	switch event.EventType {
	case message.MessageSent:
		return h.handleMessageSent(ctx, &event)
	case message.MessageRead:
		return h.handleMessageRead(ctx, &event)
	case message.MessageDeleted:
		return h.handleMessageDeleted(ctx, &event)
	default:
		logs.Warnf("unknown event type: %d", event.EventType)
		return nil
	}
}

func (h *MessageEventHandler) handleMessageSent(ctx context.Context, event *message.MessageEvent) error {
	// TODO , implement the message push logic
	logs.Infof("pushing message %d to user %d", event.MessageID, event.UserID)
	return nil
}

func (h *MessageEventHandler) handleMessageRead(ctx context.Context, event *message.MessageEvent) error {
	// 实现已读状态推送逻辑
	return nil
}

func (h *MessageEventHandler) handleMessageDeleted(ctx context.Context, event *message.MessageEvent) error {
	// TODO
	return nil
}
