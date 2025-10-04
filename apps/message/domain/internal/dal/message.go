package dal

import (
	"context"

	"gorm.io/gorm"

	"github.com/crazyfrankie/goim/apps/message/domain/internal/dal/model"
	"github.com/crazyfrankie/goim/apps/message/domain/internal/dal/query"
)

type MessageDao struct {
	query *query.Query
}

func NewMessageDao(db *gorm.DB) *MessageDao {
	return &MessageDao{query: query.Use(db)}
}

func (m *MessageDao) Create(ctx context.Context, message *model.Message) error {
	return m.query.Message.WithContext(ctx).Create(message)
}
