package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

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

func (m *messageImpl) Create(ctx context.Context, req *CreateMessageRequest) error {
	msgID, err := m.IDGen.GenID(ctx)
	if err != nil {
		return fmt.Errorf("generate id error: %w", err)
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
		Status:      req.Status,
		IsRead:      req.IsRead,
	}

	err = m.MessageRepo.Create(ctx, newMessage)
	if err != nil {
		return err
	}

	return nil
}
