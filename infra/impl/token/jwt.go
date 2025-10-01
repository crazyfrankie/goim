package token

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/crazyfrankie/goim/infra/contract/cache"
	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/types/consts"
)

const (
	RevokePrefix  = "user_access_revoked"
	RefreshPrefix = "refresh_token"
)

type TokenService struct {
	cmd       cache.Cmdable
	signAlgo  string
	secretKey *rsa.PrivateKey
	publicKey *rsa.PublicKey
}

func New(cmd cache.Cmdable) (token.Token, error) {
	signAlgo := os.Getenv(consts.JWTSignAlgo)
	secretPath := os.Getenv(consts.JWTSecretKey)
	publicPath := os.Getenv(consts.JWTPublicKey)

	privateKey, err := os.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}
	private, _ := jwt.ParseRSAPrivateKeyFromPEM(privateKey)

	publicKey, _ := os.ReadFile(publicPath)
	public, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, err
	}

	return &TokenService{cmd: cmd, signAlgo: signAlgo, secretKey: private, publicKey: public}, nil
}

func (s *TokenService) GenerateToken(uid int64) ([]string, error) {
	res := make([]string, 2)
	access, err := s.newToken(uid, time.Minute*15)
	if err != nil {
		return res, err
	}
	res[0] = access
	refresh, err := s.newToken(uid, time.Hour*24*30)
	if err != nil {
		return res, err
	}
	res[1] = refresh

	// set refresh in redis
	key := refreshKey(uid)

	err = s.cmd.Set(context.Background(), key, refresh, time.Hour*24*30).Err()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *TokenService) GenerateConnToken(uid int64) (string, error) {
	return s.newToken(uid, time.Hour*24)
}

func (s *TokenService) newToken(uid int64, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &token.Claims{
		UID: uid,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}
	tk := jwt.NewWithClaims(jwt.GetSigningMethod(s.signAlgo), claims)
	str, err := tk.SignedString(s.secretKey)

	return str, err
}

func (s *TokenService) ParseToken(tk string) (*token.Claims, error) {
	t, err := jwt.ParseWithClaims(tk, &token.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.publicKey, nil
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

func (s *TokenService) TryRefresh(refresh string) ([]string, error) {
	refreshClaims, err := s.ParseToken(refresh)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh jwt")
	}

	res, err := s.cmd.Get(context.Background(), refreshKey(refreshClaims.UID)).Result()
	if err != nil || res != refresh {
		return nil, errors.New("jwt invalid or revoked")
	}

	access, err := s.newToken(refreshClaims.UID, time.Hour)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	issat, _ := refreshClaims.GetIssuedAt()
	expire, _ := refreshClaims.GetExpirationTime()
	if expire.Sub(now) < expire.Sub(issat.Time)/3 {
		// try refresh
		refresh, err = s.newToken(refreshClaims.UID, time.Hour*24*30)
		err = s.cmd.Set(context.Background(), refreshKey(refreshClaims.UID), refresh, time.Hour*24*30).Err()
		if err != nil {
			return nil, err
		}
	}

	return []string{access, refresh}, nil
}

func (s *TokenService) CleanToken(ctx context.Context, uid int64) error {
	return s.cmd.Del(ctx, refreshKey(uid)).Err()
}

func (s *TokenService) RevokeToken(ctx context.Context, uid int64) error {
	key := revokeKey(uid)
	return s.cmd.Set(ctx, key, time.Now().Unix(), time.Hour*24).Err()
}

func revokeKey(uid int64) string {
	return fmt.Sprintf("%s:%d", RevokePrefix, uid)
}

func refreshKey(uid int64) string {
	return fmt.Sprintf("%s:%d", RefreshPrefix, uid)
}
