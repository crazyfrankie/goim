package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"

	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/lang/conv"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
)

type AuthnHandler struct {
	noAuthPaths map[string]struct{}
	authClient  authv1.AuthServiceClient
}

func NewAuthnHandler(authClient authv1.AuthServiceClient) (*AuthnHandler, error) {
	return &AuthnHandler{authClient: authClient, noAuthPaths: make(map[string]struct{})}, nil
}

func (h *AuthnHandler) IgnorePath(paths []string) *AuthnHandler {
	for _, path := range paths {
		h.noAuthPaths[path] = struct{}{}
	}
	return h
}

func (h *AuthnHandler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		currentPath := c.Request.URL.Path

		if _, ok := h.noAuthPaths[currentPath]; ok {
			c.Next()
			return
		}

		accessToken, err := getAccessToken(c)
		if err != nil {
			response.Unauthorized(c)
			return
		}
		parseRes, err := h.authClient.ParseToken(c.Request.Context(), &authv1.ParseTokenRequest{Token: accessToken})
		if err == nil {
			c.Request = c.Request.WithContext(h.storeUserID(c.Request.Context(), parseRes.GetUserID()))

			c.Next()
			return
		}

		refreshToken, err := c.Cookie("goim_refresh")
		if err != nil {
			response.Unauthorized(c)
			return
		}

		refreshRes, err := h.authClient.RefreshBizToken(c.Request.Context(), &authv1.RefreshBizTokenRequest{RefreshToken: refreshToken})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}
		c.Request = c.Request.WithContext(h.storeUserID(c.Request.Context(), refreshRes.GetUserID()))

		response.SetAuthorization(c, refreshRes.AccessToken, refreshRes.RefreshToken)

		c.Next()
	}
}

func (h *AuthnHandler) storeUserID(ctx context.Context, userID int64) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"user_id": conv.Int64ToStr(userID),
	}))
}

func getAccessToken(c *gin.Context) (string, error) {
	tokenHeader := c.GetHeader("Authorization")
	if tokenHeader == "" {
		return "", errors.New("no auth")
	}

	strs := strings.Split(tokenHeader, " ")
	if len(strs) != 2 || strs[0] != "Bearer" {
		return "", errors.New("header is invalid")
	}

	if strs[1] == "" {
		return "", errors.New("jwt is empty")
	}

	return strs[1], nil
}
