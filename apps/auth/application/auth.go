package application

import (
	"context"

	auth "github.com/crazyfrankie/goim/apps/auth/domain/service"
	"github.com/crazyfrankie/goim/pkg/grpc/ctxutil"
	authv1 "github.com/crazyfrankie/goim/protocol/auth/v1"
)

type AuthApplicationService struct {
	authDomain auth.Auth
	authv1.UnimplementedAuthServiceServer
}

func NewAuthApplicationService(authDomain auth.Auth) *AuthApplicationService {
	return &AuthApplicationService{authDomain: authDomain}
}

func (a *AuthApplicationService) GenerateConnToken(ctx context.Context, req *authv1.GenerateConnTokenRequest) (*authv1.GenerateConnTokenResponse, error) {
	userID := ctxutil.MustGetUserIDFromCtx(ctx)

	connToken, err := a.authDomain.GenerateConnToken(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &authv1.GenerateConnTokenResponse{Token: connToken}, nil
}

func (a *AuthApplicationService) ParseToken(ctx context.Context, req *authv1.ParseTokenRequest) (*authv1.ParseTokenResponse, error) {
	claims, err := a.authDomain.ParseToken(ctx, req.GetToken())
	if err != nil {
		return nil, err
	}

	return &authv1.ParseTokenResponse{UserID: claims.UID}, nil
}

func (a *AuthApplicationService) RefreshBizToken(ctx context.Context, req *authv1.RefreshBizTokenRequest) (*authv1.RefreshBizTokenResponse, error) {
	tokens, userID, err := a.authDomain.RefreshBizToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &authv1.RefreshBizTokenResponse{
		AccessToken:  tokens[0],
		RefreshToken: tokens[1],
		UserID:       userID,
	}, nil
}
