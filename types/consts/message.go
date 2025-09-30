package consts

type MessageType int

const (
	TextMessageType MessageType = iota
)

type ChatType int

const (
	GroupChatType ChatType = iota + 1
	SingleChatType
)
