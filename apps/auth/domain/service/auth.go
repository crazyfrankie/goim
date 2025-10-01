package service

import (
	"context"

	"github.com/crazyfrankie/goim/infra/contract/token"
)

type Auth interface {
	GenerateConnToken(ctx context.Context, userID int64) (string, error)
	ParseToken(ctx context.Context, token string) (*token.Claims, error)
	RefreshBizToken(ctx context.Context, refreshToken string) ([]string, error)
}
