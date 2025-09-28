package service

import (
	"context"

	"github.com/crazyfrankie/goim/apps/user/domain/entity"
)

type userImpl struct {
}

func (u *userImpl) Create(ctx context.Context, req *CreateUserRequest) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) Login(ctx context.Context, email, password string) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) Logout(ctx context.Context, userID int64) (err error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) ResetPassword(ctx context.Context, email, password string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) GetUserInfo(ctx context.Context, userID int64) (user *entity.User, err error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) UpdateAvatar(ctx context.Context, userID int64, ext string, imagePayload []byte) (url string, err error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (err error) {
	//TODO implement me
	panic("implement me")
}

func (u *userImpl) MGetUserProfiles(ctx context.Context, userIDs []int64) (users []*entity.User, err error) {
	//TODO implement me
	panic("implement me")
}
