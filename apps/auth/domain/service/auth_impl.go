package service

import (
	"context"

	"github.com/crazyfrankie/goim/infra/contract/token"
)

type Components struct {
	TokenGen token.Token
}

type authImpl struct {
	*Components
}

func NewAuthDomain(c *Components) Auth {
	return &authImpl{c}
}

func (a *authImpl) GenerateConnToken(ctx context.Context, userID int64) (string, error) {
	token, err := a.TokenGen.GenerateConnToken(userID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *authImpl) ParseToken(ctx context.Context, token string) (*token.Claims, error) {
	claims, err := a.TokenGen.ParseToken(token)
	if err != nil {
		return nil, err
	}

	return claims, err
}

func (a *authImpl) RefreshBizToken(ctx context.Context, refreshToken string) ([]string, error) {
	tokens, err := a.TokenGen.TryRefresh(refreshToken)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
