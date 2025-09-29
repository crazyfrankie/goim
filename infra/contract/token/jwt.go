package token

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UID int64 `json:"uid"`
	jwt.RegisteredClaims
}

type ResetClaims struct {
	Email string
	jwt.RegisteredClaims
}

type Token interface {
	ResetToken
	GenerateToken(uid int64, ua string) ([]string, error)
	ParseToken(token string, isAccess bool) (*Claims, error)
	TryRefresh(refresh string, ua string) ([]string, *Claims, error)
	CleanToken(ctx context.Context, uid int64, ua string) error
	RevokeToken(ctx context.Context, uid int64) error
}

type ResetToken interface {
	GenerateResetToken(email string, duration time.Duration) (string, error)
	ParseResetToken(token string) (*ResetClaims, error)
}
