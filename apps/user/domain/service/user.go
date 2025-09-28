package service

import (
	"context"

	"github.com/crazyfrankie/goim/apps/user/domain/entity"
)

type CreateUserRequest struct {
	Email       string
	Password    string
	Name        string
	UniqueName  string
	Description string
}

type UpdateProfileRequest struct {
	UserID      int64
	Name        *string
	UniqueName  *string
	Description *string
}

type User interface {
	Create(ctx context.Context, req *CreateUserRequest) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	Logout(ctx context.Context, userID int64) (err error)
	ResetPassword(ctx context.Context, email, password string) (err error)
	GetUserInfo(ctx context.Context, userID int64) (user *entity.User, err error)
	UpdateAvatar(ctx context.Context, userID int64, ext string, imagePayload []byte) (url string, err error)
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (err error)
	MGetUserProfiles(ctx context.Context, userIDs []int64) (users []*entity.User, err error)
}
