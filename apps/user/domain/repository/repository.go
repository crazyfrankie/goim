package repository

import (
	"context"

	"github.com/crazyfrankie/goim/apps/user/domain/internal/dal"
	"github.com/crazyfrankie/goim/apps/user/domain/internal/dal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUsersByEmail(ctx context.Context, email string) (*model.User, bool, error)
	UpdatePassword(ctx context.Context, email, password string) error
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
	UpdateAvatar(ctx context.Context, userID int64, iconURI string) error
	CheckUniqueNameExist(ctx context.Context, uniqueName string) (bool, error)
	UpdateProfile(ctx context.Context, userID int64, updates map[string]any) error
	CheckEmailExist(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *model.User) error
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]*model.User, error)
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return dal.NewUserDao(db)
}
