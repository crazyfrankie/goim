package entity

type Message struct {
	MsgID       int64 // Server-generated message ID
	SendID      int64
	RecvID      int64
	GroupID     int64  // Group ID (for group messages)
	Seq         int64  // Message Sequence Number
	ClientMsgID string // Client-generated message ID
	Content     string // Message Content
	SendTime    int64  // Send Time (Milliseconds)
	Status      int32  // Message Status
	IsRead      bool   // Read Status
	CreatedTime int64  // Creation Time
	UpdatedTime int64  // Updated Time
}
