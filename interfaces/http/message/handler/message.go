package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/mitchellh/mapstructure"

	"github.com/crazyfrankie/goim/interfaces/http/message/model"
	"github.com/crazyfrankie/goim/pkg/apistruct"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	"github.com/crazyfrankie/goim/pkg/lang/encrypt"
	"github.com/crazyfrankie/goim/pkg/sonic"
	messagev1 "github.com/crazyfrankie/goim/protocol/message/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessageHandler struct {
	messageClient messagev1.MessageServiceClient
}

func NewMessageHandler(messageClient messagev1.MessageServiceClient) *MessageHandler {
	return &MessageHandler{messageClient: messageClient}
}

func (h *MessageHandler) RegisterRoute(r *gin.RouterGroup) {
	messageGroup := r.Group("message")
	{
		messageGroup.POST("send", h.SendMessage())
	}
}

func (h *MessageHandler) SendMessage() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.SendMsgReq
		if err := c.ShouldBind(&req); err != nil {
			response.InvalidParamError(c, err.Error())
			return
		}

		msgReq, err := h.getSendMsgReq(&req)
		if err != nil {
			response.InvalidParamError(c, err.Error())
			return
		}

		res, err := h.messageClient.SendMessage(c.Request.Context(), msgReq)
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		response.Success(c, res)
	}
}

func (h *MessageHandler) getSendMsgReq(req *model.SendMsgReq) (*messagev1.SendMessageRequest, error) {
	var data any
	switch req.ContentType {
	case consts.TextMessageType:
		data = &apistruct.TextElem{}
	case consts.PictureMessageType:
		data = &apistruct.PictureElem{}
	case consts.VoiceMessageType:
		data = &apistruct.SoundElem{}
	case consts.VideoMessageType:
		data = &apistruct.VideoElem{}
	case consts.FileMessageType:
		data = &apistruct.FileElem{}
	case consts.AtTextMessageType:
		data = &apistruct.AtElem{}
	case consts.CustomMessageType:
		data = &apistruct.CustomElem{}
	case consts.MarkdownTextMessageType:
		data = &apistruct.MarkdownTextElem{}
	case consts.QuoteMessageType:
		data = &apistruct.QuoteElem{}
	case consts.OANotification:
		data = &apistruct.OANotificationElem{}
		req.SessionType = consts.NotificationChatType
		//if err := h.userClient.GetNotificationByID(c, req.SendID); err != nil {
		//	return nil, err
		//}
	default:
		return nil, errorx.Wrapf(nil, "unsupported content type, contentType: %s", req.ContentType)
	}

	if err := mapstructure.WeakDecode(req.Content, data); err != nil {
		return nil, errorx.Wrapf(err, "failed to decode message content")
	}

	return h.newUserSendMsgReq(req, data), nil
}

func (h *MessageHandler) newUserSendMsgReq(req *model.SendMsgReq, data any) *messagev1.SendMessageRequest {
	sendID, _ := conv.StrToInt64(req.SendID)
	groupID, _ := conv.StrToInt64(req.GroupID)
	msgData := &messagev1.Message{
		SendID:      sendID,
		GroupID:     groupID,
		ClientMsgID: encrypt.Md5(req.SendID),
		SessionType: req.SessionType,
		ContentType: req.ContentType,
		SendTime:    req.SendTime,
	}
	var newContent string
	switch req.ContentType {
	//case consts.OANotification:
	//	notification := sdkws.NotificationElem{}
	//	notification.Detail, _ = sonic.MarshalString(params.Content)
	//	newContent, _ = sonic.MarshalString(&notification)
	case consts.TextMessageType:
		fallthrough
	//case consts.AtTextMessageType:
	//	if atElem, ok := data.(*apistruct.AtElem); ok {
	//		msgData.AtUserIDList = atElem.AtUserList
	//	}
	//	fallthrough
	case consts.PictureMessageType:
		fallthrough
	case consts.CustomMessageType:
		fallthrough
	case consts.VoiceMessageType:
		fallthrough
	case consts.VideoMessageType:
		fallthrough
	case consts.FileMessageType:
		fallthrough
	default:
		newContent, _ = sonic.MarshalString(req.Content)
	}
	msgData.Content = []byte(newContent)
	pbData := &messagev1.SendMessageRequest{
		Data: msgData,
	}

	return pbData
}
