package handler

import (
	"strconv"

	"github.com/crazyfrankie/goim/pkg/util"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/goim/interfaces/user/api/model"
	"github.com/crazyfrankie/goim/pkg/gin/response"
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
		userGroup.GET("info", h.GetUserInfo())
	}
}

func (h *UserHandler) Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UserRegisterReq
		if err := c.ShouldBind(&req); err != nil {
			response.InvalidParamError(c)
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
			response.InvalidParamError(c)
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
