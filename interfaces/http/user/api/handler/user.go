package handler

import (
	"io"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/goim/interfaces/http/user/api/model"
	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/util"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
)

type UserHandler struct {
	userClient userv1.UserServiceClient
}

func NewUserHandler(userClient userv1.UserServiceClient) *UserHandler {
	return &UserHandler{userClient: userClient}
}

func (h *UserHandler) RegisterRoute(r *gin.RouterGroup) {
	userGroup := r.Group("user")
	{
		userGroup.POST("register", h.Register())
		userGroup.POST("login", h.Login())
		userGroup.GET("logout", h.Logout())
		userGroup.GET("profile", h.GetUserInfo())
		userGroup.PUT("profile", h.UpdateProfile())
		userGroup.POST("reset-password", h.ResetPassword())
	}
}

func (h *UserHandler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UserRegisterReq
		if err := c.ShouldBind(&req); err != nil {
			response.InvalidParamError(c, err.Error())
			return
		}

		res, err := h.userClient.Register(c.Request.Context(), &userv1.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		util.SetAuthorization(c, res.Data.AccessToken, res.Data.RefreshToken)

		response.Success(c, userDTO2VO(res.Data))
	}
}

func (h *UserHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UserLoginReq
		if err := c.ShouldBind(&req); err != nil {
			response.InvalidParamError(c, err.Error())
			return
		}

		res, err := h.userClient.Login(c.Request.Context(), &userv1.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		util.SetAuthorization(c, res.Data.AccessToken, res.Data.RefreshToken)

		response.Success(c, userDTO2VO(res.Data))
	}
}

func (h *UserHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := h.userClient.Logout(c.Request.Context(), &userv1.LogoutRequest{})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (h *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		res, err := h.userClient.GetUserInfo(c.Request.Context(), &userv1.GetUserInfoRequest{})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		response.Success(c, userDTO2VO(res.Data))
	}
}

func (h *UserHandler) UpdateAvatar() gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("avatar")
		if err != nil {
			logs.CtxErrorf(c.Request.Context(), "Get Avatar Fail failed, err=%v", err)
			response.InvalidParamError(c, "missing avatar file")
			return
		}

		// Check file type
		if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
			response.InvalidParamError(c, "invalid file type, only image allowed")
			return
		}

		// Read file content
		src, err := file.Open()
		if err != nil {
			response.InternalServerError(c, err)
			return
		}
		defer src.Close()

		fileContent, err := io.ReadAll(src)
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		res, err := h.userClient.UpdateAvatar(c.Request.Context(), &userv1.UpdateAvatarRequest{
			Avatar:   fileContent,
			MimeType: file.Header.Get("Content-Type"),
		})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		response.Success(c, res)
	}
}

func (h *UserHandler) UpdateProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UserUpdateProfileReq
		if err := c.ShouldBind(&req); err != nil {
			response.InvalidParamError(c, err.Error())
			return
		}

		_, err := h.userClient.UpdateProfile(c.Request.Context(), &userv1.UpdateProfileRequest{
			Name:           req.Name,
			UserUniqueName: req.UniqueName,
			Description:    req.Description,
			Sex:            (*userv1.Sex)(req.Sex),
		})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func (h *UserHandler) ResetPassword() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UserResetPasswordReq
		if err := c.ShouldBind(&req); err != nil {
			response.InvalidParamError(c, err.Error())
			return
		}

		_, err := h.userClient.ResetPassword(c.Request.Context(), &userv1.ResetPasswordRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		response.Success(c, nil)
	}
}

func userDTO2VO(userDto *userv1.User) *model.UserInfoResp {
	return &model.UserInfoResp{
		UserID:         strconv.FormatInt(userDto.UserId, 10),
		Name:           userDto.Name,
		UserUniqueName: userDto.UserUniqueName,
		Email:          userDto.Email,
		Description:    userDto.Description,
		Avatar:         userDto.AvatarUrl,
		UserCreateTime: userDto.UserCreateTime,
	}
}
