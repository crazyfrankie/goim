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
	msgv1 "github.com/crazyfrankie/goim/protocol/msg/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type MessageHandler struct {
	messageClient msgv1.MessageServiceClient
}

func NewMessageHandler(messageClient msgv1.MessageServiceClient) *MessageHandler {
	return &MessageHandler{messageClient: messageClient}
}

func (h *MessageHandler) RegisterRoute(r *gin.RouterGroup) {
	messageGroup := r.Group("msg")
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

func (h *MessageHandler) getSendMsgReq(req *model.SendMsgReq) (*msgv1.SendMessageRequest, error) {
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
	case consts.CustomMessageType:
		data = &apistruct.CustomElem{}
	case consts.QuoteMessageType:
		data = &apistruct.QuoteElem{}
	default:
		return nil, errorx.Wrapf(nil, "unsupported content type, contentType: %s", req.ContentType)
	}

	if err := mapstructure.WeakDecode(req.Content, data); err != nil {
		return nil, errorx.Wrapf(err, "failed to decode msg content")
	}

	return h.newUserSendMsgReq(req, data), nil
}

func (h *MessageHandler) newUserSendMsgReq(req *model.SendMsgReq, data any) *msgv1.SendMessageRequest {
	sendID, _ := conv.StrToInt64(req.SendID)
	groupID, _ := conv.StrToInt64(req.GroupID)
	msgData := &msgv1.Message{
		SendID:      sendID,
		GroupID:     groupID,
		ClientMsgID: encrypt.Md5(req.SendID),
		SessionType: req.SessionType,
		ContentType: req.ContentType,
		SendTime:    req.SendTime,
	}
	var newContent string
	switch req.ContentType {
	case consts.TextMessageType:
		fallthrough
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
	pbData := &msgv1.SendMessageRequest{
		Data: msgData,
	}

	return pbData
}
