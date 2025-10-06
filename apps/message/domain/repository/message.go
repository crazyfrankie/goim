package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/apps/message/domain/internal/dal"
	"github.com/crazyfrankie/goim/apps/message/domain/internal/dal/model"
)

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return dal.NewMessageDao(db)
}

type MessageRepository interface {
	Create(ctx context.Context, message *model.Message) error
	UpdateMessageStatus(ctx context.Context, status int32) error
}
