package model

type SendMsgReq struct {
	RecvID      string         `json:"recv_id"`
	SendID      string         `json:"sendID" binding:"required"`
	GroupID     string         `json:"groupID" binding:"required_if=SessionType 2|required_if=SessionType 3"`
	Content     map[string]any `json:"content" binding:"required" swaggerignore:"true"`
	ContentType int32          `json:"contentType" binding:"required"`
	SessionType int32          `json:"sessionType" binding:"required"`
	SendTime    int64          `json:"sendTime"`
}
