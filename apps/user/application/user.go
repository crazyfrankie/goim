package application

import (
	"context"
	"net/mail"
	"strconv"

	"github.com/crazyfrankie/goim/apps/user/domain/entity"
	user "github.com/crazyfrankie/goim/apps/user/domain/service"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/grpc/ctxutil"
	langslice "github.com/crazyfrankie/goim/pkg/lang/slice"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
	"github.com/crazyfrankie/goim/types/errno"
)

type UserApplicationService struct {
	userDomain user.User
	userv1.UnimplementedUserServiceServer
}

func NewUserApplicationService(userDomain user.User) userv1.UserServiceServer {
	return &UserApplicationService{userDomain: userDomain}
}

func isValidEmail(email string) bool {
	// If the email string is not in the correct format, it will return an error.
	_, err := mail.ParseAddress(email)
	return err == nil
}

func (u *UserApplicationService) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	if !isValidEmail(req.GetEmail()) {
		return nil, errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "invalid email"))
	}

	userInfo, err := u.userDomain.Create(ctx, &user.CreateUserRequest{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
		Name:     req.GetName(),
	})
	if err != nil {
		return nil, err
	}

	userInfo, err = u.userDomain.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &userv1.RegisterResponse{
		Data: userDO2DTO(userInfo),
	}, nil
}

func (u *UserApplicationService) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	userInfo, err := u.userDomain.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &userv1.LoginResponse{
		Data: userDO2DTO(userInfo),
	}, nil
}

func (u *UserApplicationService) GetUserInfo(ctx context.Context, req *userv1.GetUserInfoRequest) (*userv1.GetUserInfoResponse, error) {
	userID := ctxutil.MustGetUserIDFromCtx(ctx)

	userInfo, err := u.userDomain.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &userv1.GetUserInfoResponse{Data: userDO2DTO(userInfo)}, nil
}

func (u *UserApplicationService) MGetUserInfo(ctx context.Context, req *userv1.MGetUserInfoRequest) (*userv1.MGetUserInfoResponse, error) {
	userIDs, err := langslice.TransformWithErrorCheck(req.GetUserIds(), func(s string) (int64, error) {
		return strconv.ParseInt(s, 10, 64)
	})
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.ErrUserInvalidParamCode, errorx.KV("msg", "invalid user id"))
	}

	userInfos, err := u.userDomain.MGetUserProfiles(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	return &userv1.MGetUserInfoResponse{
		Data: langslice.ToMap(userInfos, func(userInfo *entity.User) (string, *userv1.User) {
			return strconv.FormatInt(userInfo.UserID, 10), userDO2DTO(userInfo)
		}),
	}, nil
}

func (u *UserApplicationService) ResetPassword(ctx context.Context, req *userv1.ResetPasswordRequest) (*userv1.ResetPasswordResponse, error) {
	err := u.userDomain.ResetPassword(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (u *UserApplicationService) UpdateAvatar(ctx context.Context, req *userv1.UpdateAvatarRequest) (*userv1.UpdateAvatarResponse, error) {
	var ext string
	var err error
	switch req.GetMimeType() {
	case "image/jpeg", "image/jpg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/gif":
		ext = "gif"
	case "image/webp":
		ext = "webp"
	default:
		return nil, errorx.WrapByCode(err, errno.ErrUserInvalidParamCode,
			errorx.KV("msg", "unsupported image type"))
	}

	userID := ctxutil.MustGetUserIDFromCtx(ctx)

	iconUrl, err := u.userDomain.UpdateAvatar(ctx, userID, ext, req.GetAvatar())
	if err != nil {
		return nil, err
	}

	return &userv1.UpdateAvatarResponse{AvatarUrl: iconUrl}, nil
}

func (u *UserApplicationService) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	userID := ctxutil.MustGetUserIDFromCtx(ctx)

	err := u.userDomain.UpdateProfile(ctx, &user.UpdateProfileRequest{
		UserID:      userID,
		Name:        req.Name,
		UniqueName:  req.UserUniqueName,
		Description: req.Description,
		Sex:         (*int32)(req.Sex),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func userDO2DTO(userDo *entity.User) *userv1.User {
	return &userv1.User{
		UserId:         userDo.UserID,
		Name:           userDo.Name,
		Email:          userDo.Email,
		UserUniqueName: userDo.UniqueName,
		AvatarUrl:      userDo.IconURL,
		Description:    userDo.Description,
		Sex:            userv1.Sex(userDo.Sex),
		AccessToken:    userDo.AccessToken,
		RefreshToken:   userDo.RefreshToken,

		UserCreateTime: userDo.CreatedAt / 1000,
	}
}
