package consts

const (
	TextMessageType = iota
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
	GroupChatType = iota
	SingleChatType
	WriteGroupChatType
	ReadGroupChatType
	NotificationChatType
)
