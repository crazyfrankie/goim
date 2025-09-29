package token

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	
	"github.com/crazyfrankie/goim/infra/contract/cache"
	"github.com/crazyfrankie/goim/infra/contract/token"
	"github.com/crazyfrankie/goim/types/consts"
)

var (
	ErrUserRevoked = errors.New("user revoked")
)

const (
	RevokePrefix  = "user_access_revoked"
	RefreshPrefix = "refresh_token"
)

type TokenService struct {
	cmd       cache.Cmdable
	signAlgo  string
	secretKey []byte
}

func New(cmd cache.Cmdable) (token.Token, error) {
	signAlgo := os.Getenv(consts.JWTSignAlgo)
	secret := os.Getenv(consts.JWTSecretKey)

	return &TokenService{cmd: cmd, signAlgo: signAlgo, secretKey: []byte(secret)}, nil
}

func (s *TokenService) GenerateToken(uid int64, ua string) ([]string, error) {
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
	key := refreshKey(uid, ua)

	err = s.cmd.Set(context.Background(), key, refresh, time.Hour*24*30).Err()
	if err != nil {
		return nil, err
	}

	return res, nil
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

func (s *TokenService) ParseToken(tk string, isAccess bool) (*token.Claims, error) {
	t, err := jwt.ParseWithClaims(tk, &token.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*token.Claims)
	if !ok {
		return nil, errors.New("jwt is invalid")
	}

	if isAccess {
		key := revokeKey(claims.UID)
		revokedTime, err := s.cmd.Get(context.Background(), key).Result()
		if err == nil {
			issuedAt, _ := claims.GetIssuedAt()
			if revokedTime != "" && issuedAt.Unix() < parseRevokedTime(revokedTime) {
				return nil, ErrUserRevoked
			}
		}
	}

	return claims, nil
}

func (s *TokenService) TryRefresh(refresh string, ua string) ([]string, *token.Claims, error) {
	refreshClaims, err := s.ParseToken(refresh, false)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid refresh jwt")
	}

	res, err := s.cmd.Get(context.Background(), refreshKey(refreshClaims.UID, ua)).Result()
	if err != nil || res != refresh {
		return nil, nil, errors.New("jwt invalid or revoked")
	}

	access, err := s.newToken(refreshClaims.UID, time.Hour)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	issat, _ := refreshClaims.GetIssuedAt()
	expire, _ := refreshClaims.GetExpirationTime()
	if expire.Sub(now) < expire.Sub(issat.Time)/3 {
		// try refresh
		refresh, err = s.newToken(refreshClaims.UID, time.Hour*24*30)
		err = s.cmd.Set(context.Background(), refreshKey(refreshClaims.UID, ua), refresh, time.Hour*24*30).Err()
		if err != nil {
			return nil, nil, err
		}
	}

	return []string{access, refresh}, refreshClaims, nil
}

func (s *TokenService) CleanToken(ctx context.Context, uid int64, ua string) error {
	return s.cmd.Del(ctx, refreshKey(uid, ua)).Err()
}

func (s *TokenService) RevokeToken(ctx context.Context, uid int64) error {
	key := revokeKey(uid)
	return s.cmd.Set(ctx, key, time.Now().Unix(), time.Hour*24).Err()
}

func revokeKey(uid int64) string {
	return fmt.Sprintf("%s:%d", RevokePrefix, uid)
}

func refreshKey(uid int64, ua string) string {
	hash := hashUA(ua)
	return fmt.Sprintf("%s:%d:%s", RefreshPrefix, uid, hash)
}

func hashUA(ua string) string {
	sum := sha1.Sum([]byte(ua))
	return hex.EncodeToString(sum[:])
}

func parseRevokedTime(revokedTime string) int64 {
	res, err := strconv.ParseInt(revokedTime, 10, 64)
	if err != nil {
		return 0
	}

	return res
}
