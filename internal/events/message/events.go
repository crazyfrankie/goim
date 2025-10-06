package message

type MessageEvent struct {
	EventType   EventType  `json:"event_type"`
	MessageID   int64      `json:"message_id"`
	UserID      int64      `json:"user_id"`
	Content     string     `json:"content"`
	TimestampMS int64      `json:"timestamp_ms"`
	Meta        *EventMeta `json:"meta,omitempty"`
}

type EventType int

const (
	MessageSent EventType = iota + 1
	MessageRead
	MessageDeleted
)

type EventMeta struct {
	TraceID    string `json:"trace_id,omitempty"`
	SendTimeMs int64  `json:"send_time_ms"`
}
