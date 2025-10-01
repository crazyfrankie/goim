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

type ValidateProfileUpdateRequest struct {
	UniqueName *string
	Email      *string
}

type ValidateProfileUpdateResult int

const (
	ValidateSuccess             ValidateProfileUpdateResult = 0
	UniqueNameExist             ValidateProfileUpdateResult = 2
	UniqueNameTooShortOrTooLong ValidateProfileUpdateResult = 3
	EmailExist                  ValidateProfileUpdateResult = 5
)

type ValidateProfileUpdateResponse struct {
	Code ValidateProfileUpdateResult
	Msg  string
}

type UpdateProfileRequest struct {
	UserID      int64
	Name        *string
	UniqueName  *string
	Description *string
	Sex         *int32
}

type User interface {
	Create(ctx context.Context, req *CreateUserRequest) (*entity.User, error)
	Login(ctx context.Context, email, password string) (*entity.User, error)
	Logout(ctx context.Context, userID int64) error
	ResetPassword(ctx context.Context, email, password string) error
	GetUserInfo(ctx context.Context, userID int64) (user *entity.User, err error)
	ValidateProfileUpdate(ctx context.Context, req *ValidateProfileUpdateRequest) (resp *ValidateProfileUpdateResponse, err error)
	UpdateAvatar(ctx context.Context, userID int64, ext string, imagePayload []byte) (url string, err error)
	UpdateProfile(ctx context.Context, req *UpdateProfileRequest) error
	MGetUserProfiles(ctx context.Context, userIDs []int64) (users []*entity.User, err error)
}
