package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/apps/message/domain/entity"
	"github.com/crazyfrankie/goim/apps/message/domain/internal/dal/model"
	"github.com/crazyfrankie/goim/apps/message/domain/repository"
	"github.com/crazyfrankie/goim/infra/contract/idgen"
)

type Components struct {
	MessageRepo repository.MessageRepository
	DB          *gorm.DB
	IDGen       idgen.IDGenerator
}

type messageImpl struct {
	*Components
}

func NewMessageDomain(c *Components) Message {
	return &messageImpl{c}
}

func (m *messageImpl) Create(ctx context.Context, req *CreateMessageRequest) (*entity.Message, error) {
	msgID, err := m.IDGen.GenID(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate id error: %w", err)
	}
	newMessage := &model.Message{
		ID:          msgID,
		SendID:      req.SendID,
		RecvID:      req.RecvID,
		GroupID:     req.GroupID,
		ClientMsgID: req.ClientMsgID,
		SessionType: req.SessionType,
		MessageFrom: req.MessageFrom,
		ContentType: req.ContentType,
		Content:     req.Content,
		Seq:         req.Seq,
		SendTime:    req.SendTime,
	}

	err = m.MessageRepo.Create(ctx, newMessage)
	if err != nil {
		return nil, err
	}

	return messagePO2DO(newMessage), nil
}

func (m *messageImpl) UpdateMessageStatus(ctx context.Context, status int32) error {
	err := m.MessageRepo.UpdateMessageStatus(ctx, status)
	if err != nil {
		return err
	}

	return nil
}

func messagePO2DO(msgPO *model.Message) *entity.Message {
	return &entity.Message{
		MsgID:       msgPO.ID,
		SendID:      msgPO.SendID,
		RecvID:      msgPO.RecvID,
		GroupID:     msgPO.GroupID,
		ClientMsgID: msgPO.ClientMsgID,
		Seq:         msgPO.Seq,
		Content:     msgPO.Content,
		SendTime:    msgPO.SendTime,
		Status:      msgPO.Status,
		IsRead:      msgPO.IsRead,
		CreatedTime: msgPO.CreatedTime,
		UpdatedTime: msgPO.UpdatedTime,
	}
}
