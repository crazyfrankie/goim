package handler

import (
	"net/http"

	"github.com/crazyfrankie/goim/interfaces/user/api/model"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"

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
			c.JSON(http.StatusBadRequest, err)
			return
		}

		ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.New(map[string]string{
			"user_agent": c.Request.UserAgent(),
		}))

		res, err := h.userClient.Register(ctx, &userv1.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func (h *UserHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UserLoginReq
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		res, err := h.userClient.Login(c.Request.Context(), &userv1.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func (h *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
