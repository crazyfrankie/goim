package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/crazyfrankie/goim/apps/user/domain/entity"
	"github.com/crazyfrankie/goim/apps/user/domain/internal/dal/model"
	"github.com/crazyfrankie/goim/apps/user/domain/repository"
	"github.com/crazyfrankie/goim/infra/contract/idgen"
	"github.com/crazyfrankie/goim/infra/contract/storage"
	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/pkg/errorx"
	"github.com/crazyfrankie/goim/pkg/lang/ptr"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/types/consts"
	"github.com/crazyfrankie/goim/types/errno"
)

type Components struct {
	UserRepo repository.UserRepository
	IDGen    idgen.IDGenerator
	IconOSS  storage.Storage
	TokenGen token.Token
}

type userImpl struct {
	*Components
}

func NewUserDomain(c *Components) User {
	return &userImpl{c}
}

func (u *userImpl) Create(ctx context.Context, req *CreateUserRequest) (*entity.User, error) {
	exist, err := u.UserRepo.CheckEmailExist(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if exist {
		return nil, errorx.New(errno.ErrUserEmailAlreadyExistCode, errorx.KV("email", req.Email))
	}

	if req.UniqueName != "" {
		exist, err = u.UserRepo.CheckUniqueNameExist(ctx, req.UniqueName)
		if err != nil {
			return nil, err
		}
		if exist {
			return nil, errorx.New(errno.ErrUserUniqueNameAlreadyExistCode, errorx.KV("name", req.UniqueName))
		}
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	name := req.Name
	if name == "" {
		name = strings.Split(req.Email, "@")[0]
	}

	userID, err := u.IDGen.GenID(ctx)
	if err != nil {
		return nil, fmt.Errorf("generate id error: %w", err)
	}

	now := time.Now().UnixMilli()

	newUser := &model.User{
		ID:          userID,
		IconURI:     consts.UserIconURI,
		Name:        name,
		UniqueName:  u.getUniqueNameFormEmail(ctx, req.Email),
		Email:       req.Email,
		Password:    hashedPassword,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = u.UserRepo.CreateUser(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("insert user failed: %w", err)
	}

	iconURL, err := u.IconOSS.GetObjectUrl(ctx, newUser.IconURI)
	if err != nil {
		return nil, fmt.Errorf("get icon url failed: %w", err)
	}

	return userPO2DO(newUser, nil, iconURL), nil
}

func (u *userImpl) Login(ctx context.Context, email, password string) (*entity.User, error) {
	userModel, exist, err := u.UserRepo.GetUsersByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errorx.New(errno.ErrUserInfoInvalidCode)
	}

	valid := verifyPassword(password, userModel.Password)
	if !valid {
		return nil, errorx.New(errno.ErrUserInfoInvalidCode)
	}

	tokens, err := u.TokenGen.GenerateToken(userModel.ID)
	if err != nil {
		return nil, err
	}

	resURL, err := u.IconOSS.GetObjectUrl(ctx, userModel.IconURI)
	if err != nil {
		return nil, err
	}

	return userPO2DO(userModel, tokens, resURL), nil
}

func (u *userImpl) Logout(ctx context.Context, userID int64) error {
	return u.TokenGen.CleanToken(ctx, userID)
}

func (u *userImpl) ResetPassword(ctx context.Context, email, password string) error {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return err
	}

	return u.UserRepo.UpdatePassword(ctx, email, hashedPassword)
}

func (u *userImpl) GetUserInfo(ctx context.Context, userID int64) (*entity.User, error) {
	if userID <= 0 {
		return nil, errorx.New(errno.ErrUserInvalidParamCode,
			errorx.KVf("msg", "invalid user id : %d", userID))
	}

	userModel, err := u.UserRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resURL, err := u.IconOSS.GetObjectUrl(ctx, userModel.IconURI)
	if err != nil {
		return nil, err
	}

	return userPO2DO(userModel, nil, resURL), nil
}

func (u *userImpl) UpdateAvatar(ctx context.Context, userID int64, ext string, imagePayload []byte) (url string, err error) {
	avatarKey := "user_avatar/" + strconv.FormatInt(userID, 10) + "." + ext
	err = u.IconOSS.PutObject(ctx, avatarKey, imagePayload)
	if err != nil {
		return "", err
	}

	err = u.UserRepo.UpdateAvatar(ctx, userID, avatarKey)
	if err != nil {
		return "", err
	}

	url, err = u.IconOSS.GetObjectUrl(ctx, avatarKey)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (u *userImpl) ValidateProfileUpdate(ctx context.Context, req *ValidateProfileUpdateRequest) (
	resp *ValidateProfileUpdateResponse, err error,
) {
	if req.UniqueName == nil && req.Email == nil {
		return nil, errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "missing parameter"))
	}

	if req.UniqueName != nil {
		uniqueName := ptr.From(req.UniqueName)
		charNum := utf8.RuneCountInString(uniqueName)

		if charNum < 4 || charNum > 20 {
			return &ValidateProfileUpdateResponse{
				Code: UniqueNameTooShortOrTooLong,
				Msg:  "unique name length should be between 4 and 20",
			}, nil
		}

		exist, err := u.UserRepo.CheckUniqueNameExist(ctx, uniqueName)
		if err != nil {
			return nil, err
		}

		if exist {
			return &ValidateProfileUpdateResponse{
				Code: UniqueNameExist,
				Msg:  "unique name existed",
			}, nil
		}
	}

	return &ValidateProfileUpdateResponse{
		Code: ValidateSuccess,
		Msg:  "success",
	}, nil
}

func (u *userImpl) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) error {
	updates := map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
	}

	if req.UniqueName != nil {
		resp, err := u.ValidateProfileUpdate(ctx, &ValidateProfileUpdateRequest{
			UniqueName: req.UniqueName,
		})
		if err != nil {
			return err
		}

		if resp.Code != ValidateSuccess {
			return errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", resp.Msg))
		}

		updates["unique_name"] = ptr.From(req.UniqueName)
	}

	if req.Name != nil {
		updates["name"] = ptr.From(req.Name)
	}

	if req.Description != nil {
		updates["description"] = ptr.From(req.Description)
	}

	if req.Sex != nil {
		updates["sex"] = ptr.From(req.Sex)
	}

	err := u.UserRepo.UpdateProfile(ctx, req.UserID, updates)
	if err != nil {
		return err
	}

	return nil
}

func (u *userImpl) MGetUserProfiles(ctx context.Context, userIDs []int64) (users []*entity.User, err error) {
	userModels, err := u.UserRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	users = make([]*entity.User, 0, len(userModels))
	for _, um := range userModels {
		// Get image URL
		resURL, err := u.IconOSS.GetObjectUrl(ctx, um.IconURI)
		if err != nil {
			continue // If getting the image URL fails, skip the user
		}

		users = append(users, userPO2DO(um, nil, resURL))
	}

	return users, nil
}

func (u *userImpl) getUniqueNameFormEmail(ctx context.Context, email string) string {
	arr := strings.Split(email, "@")
	if len(arr) != 2 {
		return email
	}

	username := arr[0]

	exist, err := u.UserRepo.CheckUniqueNameExist(ctx, username)
	if err != nil {
		logs.CtxWarnf(ctx, "check unique name exist failed: %v", err)
		return email
	}

	if exist {
		logs.CtxWarnf(ctx, "unique name %s already exist", username)

		return email
	}

	return username
}

func userPO2DO(model *model.User, tokens []string, iconURL string) *entity.User {
	res := &entity.User{
		UserID:      model.ID,
		Name:        model.Name,
		UniqueName:  model.UniqueName,
		Email:       model.Email,
		Description: model.Description,
		Sex:         model.Sex,
		IconURI:     model.IconURI,
		IconURL:     iconURL,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
	if tokens != nil {
		res.AccessToken = tokens[0]
		res.RefreshToken = tokens[1]
	}

	return res
}

func hashPassword(password string) (string, error) {
	hashPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashPass), nil
}

func verifyPassword(password, encodedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(encodedHash), []byte(password))
	if err != nil {
		return false
	}

	return true
}
