package application

import (
	"context"

	auth "github.com/crazyfrankie/goim/apps/auth/domain/service"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
)

type AuthApplicationService struct {
	authDomain auth.Auth
	authv1.UnimplementedAuthServiceServer
}

func NewAuthApplicationService(authDomain auth.Auth) *AuthApplicationService {
	return &AuthApplicationService{authDomain: authDomain}
}

func (a *AuthApplicationService) GenerateToken(ctx context.Context, req *authv1.GenerateTokenRequest) (*authv1.GenerateTokenResponse, error) {
	tokens, err := a.authDomain.GenerateToken(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	return &authv1.GenerateTokenResponse{
		AccessToken:  tokens[0],
		RefreshToken: tokens[1],
	}, nil
}

func (a *AuthApplicationService) ParseToken(ctx context.Context, req *authv1.ParseTokenRequest) (*authv1.ParseTokenResponse, error) {
	claims, err := a.authDomain.ParseToken(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}

	return &authv1.ParseTokenResponse{UserID: claims.UID}, nil
}

func (a *AuthApplicationService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	tokens, userID, err := a.authDomain.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:  tokens[0],
		RefreshToken: tokens[1],
		UserID:       userID,
	}, nil
}
