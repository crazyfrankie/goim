package consts

const (
	TextMessageType = iota + 1
	PictureMessageType
	VoiceMessageType
	VideoMessageType
	FileMessageType
	AtTextMessageType
	MergerMessageType
	CardMessageType
	LocationMessageType
	CustomMessageType
	RevokeMessageType
	TypingMessageType
	QuoteMessageType
	MarkdownTextMessageType
	OANotification
)

const (
	GroupChatType = iota + 1
	SingleChatType
	WriteGroupChatType
	ReadGroupChatType
	NotificationChatType
)
