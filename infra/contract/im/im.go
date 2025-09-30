package im

import (
	"context"
)

// IMEngine IM引擎接口
type IMEngine interface {
	// PushToUser 推送消息给指定用户
	PushToUser(ctx context.Context, userID int64, message *Message) error
	// PushToUsers 推送消息给多个用户
	PushToUsers(ctx context.Context, userIDs []int64, message *Message) error
	// PushToRoom 推送消息给房间(群组)
	PushToRoom(ctx context.Context, roomID string, message *Message) error
	// GetOnlineUsers 获取在线用户列表
	GetOnlineUsers(ctx context.Context, userIDs []int64) ([]int64, error)
	// IsUserOnline 检查用户是否在线
	IsUserOnline(ctx context.Context, userID int64) (bool, error)
}

// Message 消息结构
type Message struct {
	ID       int64       `json:"id"`
	Type     MessageType `json:"type"`
	Content  string      `json:"content"`
	FromUser int64       `json:"from_user"`
	ToUser   int64       `json:"to_user,omitempty"`
	GroupID  int64       `json:"group_id,omitempty"`
	Extra    interface{} `json:"extra,omitempty"`
}

// MessageType 消息类型
type MessageType int32

const (
	MessageTypeText  MessageType = 1 // 文本消息
	MessageTypeImage MessageType = 2 // 图片消息
	MessageTypeFile  MessageType = 3 // 文件消息
	MessageTypeAudio MessageType = 4 // 语音消息
	MessageTypeVideo MessageType = 5 // 视频消息
)

// PushResult 推送结果
type PushResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
