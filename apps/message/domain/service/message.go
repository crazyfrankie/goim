package service

import "context"

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
	Status      int32  // Message Status (e.g., delivered, seen)
	IsRead      bool   // Read Status
}

type Message interface {
	Create(ctx context.Context, req *CreateMessageRequest) error
}
