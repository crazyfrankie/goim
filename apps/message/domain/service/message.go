package service

import (
	"context"

	"github.com/crazyfrankie/goim/apps/message/domain/entity"
)

type CreateMessageRequest struct {
	SendID      int64  // Sender ID
	RecvID      int64  // Receiver ID
	GroupID     int64  // Group ID (optional, for group messages)
	ClientMsgID int64  // Client's unique message ID
	Content     string // Message Content
	SessionType int32  // Session Type (1: private, 2: group)
	MessageFrom int32  // Source of the message (e.g., user, system)
	ContentType int32  // Type of content (e.g., text, image, etc.)
	Seq         int64  // Message Sequence Number
	SendTime    int64  // Send Time (Milliseconds)
}

type Message interface {
	Create(ctx context.Context, req *CreateMessageRequest) (*entity.Message, error)
	UpdateMessageStatus(ctx context.Context, status int32) error
}
