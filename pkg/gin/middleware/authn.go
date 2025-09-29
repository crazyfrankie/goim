package middleware

import (
	"context"
	"crypto/rsa"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"

	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/pkg/gin/response"
	"github.com/crazyfrankie/goim/pkg/util"
	userv1 "github.com/crazyfrankie/goim/protocol/user/v1"
	"github.com/crazyfrankie/goim/types/consts"
)

type AuthnHandler struct {
	noAuthPaths map[string]struct{}
	userClient  userv1.UserServiceClient
	publicKey   *rsa.PublicKey
}

func NewAuthnHandler(userClient userv1.UserServiceClient) (*AuthnHandler, error) {
	publicFile := os.Getenv(consts.JWTPublicKey)
	publicKey, _ := os.ReadFile(publicFile)
	key, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, err
	}

	return &AuthnHandler{userClient: userClient, noAuthPaths: make(map[string]struct{}), publicKey: key}, nil
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
		claims, err := h.parseToken(accessToken)
		if err == nil {
			c.Request = c.Request.WithContext(h.storeUserID(c.Request.Context(), claims.UID))

			c.Next()
			return
		}

		refreshToken, err := c.Cookie("goim_refresh")
		if err != nil {
			response.Unauthorized(c)
			return
		}

		res, err := h.userClient.RefreshToken(c.Request.Context(), &userv1.RefreshTokenRequest{RefreshToken: refreshToken})
		if err != nil {
			response.InternalServerError(c, err)
			return
		}

		util.SetAuthorization(c, res.AccessToken, res.RefreshToken)

		c.Next()
	}
}

func (h *AuthnHandler) parseToken(tk string) (*token.Claims, error) {
	t, err := jwt.ParseWithClaims(tk, &token.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return h.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*token.Claims)
	if !ok {
		return nil, errors.New("jwt is invalid")
	}

	return claims, nil
}

func (h *AuthnHandler) storeUserID(ctx context.Context, userID int64) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
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
