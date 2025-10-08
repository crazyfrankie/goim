package types

const (
	WsUserID                = "sendID"
	OpUserID                = "userID"
	PlatformID              = "platformID"
	ConnID                  = "connID"
	Token                   = "token"
	OperationID             = "operationID"
	RemoteAddr              = "remoteAddr"
	Compression             = "compression"
	GzipCompressionProtocol = "gzip"
	BackgroundStatus        = "isBackground"
	SendResponse            = "isMsgResp"
	SDKType                 = "sdkType"
)

const (
	GoSDK = "go"
	JsSDK = "js"
)

const (
	WSGetNewestSeq        = 1001
	WSPullMsgBySeqList    = 1002
	WSSendMsg             = 1003
	WSSendSignalMsg       = 1004
	WSPullMsg             = 1005
	WSGetConvMaxReadSeq   = 1006
	WsPullConvLastMessage = 1007
	WSPushMsg             = 2001
	WSKickOnlineMsg       = 2002
	WsLogoutMsg           = 2003
	WsSetBackgroundStatus = 2004
	WsSubUserOnlineStatus = 2005
	WSDataError           = 3001
)
