package types

import (
	"errors"
)

type Message struct {
	Type      int32       `json:"type"`
	Operation int32       `json:"operation"`
	Seq       int32       `json:"seq"`
	Data      []byte      `json:"data"`
	Meta      interface{} `json:"meta,omitempty"`
}

const (
	MessageTypeHeartbeat = 1
	MessageTypeData      = 2
	MessageTypeNotify    = 3
	MessageTypeBinary    = 4
	MessageTypeAuth      = 7
	MessageTypeAuthReply = 8
)

const (
	OpHeartbeat      = 2
	OpHeartbeatReply = 3
	OpMessage        = 5
	OpAuth           = 7
	OpAuthReply      = 8
	OpChangeRoom     = 12
	OpSub            = 13
	OpUnsub          = 14

	OpMessageAck       = 101
	OpGroupMessage     = 200
	OpUserStatus       = 300
	OpUserStatusNotify = 301
)

const (
	OpChangeRoomReply = OpChangeRoom + 1000
	OpSubReply        = OpSub + 1000
	OpUnsubReply      = OpUnsub + 1000
)

// BroadcastReq Broadcast Request
type BroadcastReq struct {
	RoomID  string
	Message *Message
}

// UserState User Status
type UserState struct {
	UserID  string
	Online  []int32
	Offline []int32
}

var (
	ErrClientClosed           = errors.New("client closed")
	ErrSendQueueFull          = errors.New("send queue full")
	ErrRingFull               = errors.New("ring buffer full")
	ErrRingEmpty              = errors.New("ring buffer empty")
	ErrConnClosed             = errors.New("connection closed")
	ErrInvalidMessage         = errors.New("invalid message")
	ErrAuthFailed             = errors.New("authentication failed")
	ErrUnsupportedMessageType = errors.New("unsupported message type")
	ErrInvalidMessageFormat   = errors.New("invalid message format")
	ErrNotSubscribed          = errors.New("not subscribed to user")
	ErrUnsupportedOperation   = errors.New("unsupported operation")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrClientNotFound         = errors.New("client not found")
)

// DefaultServerConfig Default Server Configuration
//func DefaultServerConfig() *ServerConfig {
//	return &ServerConfig{
//		BucketNum:     32,
//		MaxConnNum:    100000,
//		HeartbeatTime: time.Second * 30,
//		Bucket: &BucketConfig{
//			ChannelSize:   10000,
//			RoomSize:      1000,
//			RoutineAmount: 32,
//			RoutineSize:   1024,
//			Round: &RoundConfig{
//				ReaderNum:     32,
//				ReaderBuf:     1024,
//				ReaderBufSize: 8192,
//				WriterNum:     32,
//				WriterBuf:     1024,
//				WriterBufSize: 8192,
//				TimerNum:      32,
//				TimerSize:     2048,
//			},
//		},
//		Client: &ClientConfig{
//			SendQueueSize: 1000,
//			RecvRingSize:  512,
//			WriteTimeout:  time.Second * 10,
//			ReadTimeout:   time.Second * 30,
//		},
//	}
//}
