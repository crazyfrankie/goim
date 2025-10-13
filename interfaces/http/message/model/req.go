package model

type SendMsgReq struct {
	RecvID      string         `json:"recvID,omitempty"`
	SendID      string         `json:"sendID,omitempty" binding:"required"`
	GroupID     string         `json:"groupID,omitempty" binding:"required_if=SessionType 2|required_if=SessionType 3"`
	Content     map[string]any `json:"content,omitempty" binding:"required" swaggerignore:"true"`
	ContentType int32          `json:"contentType" binding:"required"`
	SessionType int32          `json:"sessionType,omitempty" binding:"required"`
	SendTime    int64          `json:"sendTime,omitempty"`
}
