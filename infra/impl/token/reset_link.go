package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/crazyfrankie/goim/infra/contract/token"
)

func (s *TokenService) GenerateResetToken(email string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &token.ResetClaims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}
	tk := jwt.NewWithClaims(jwt.GetSigningMethod(s.signAlgo), claims)
	str, err := tk.SignedString(s.secretKey)

	return str, err
}

func (s *TokenService) ParseResetToken(tk string) (*token.ResetClaims, error) {
	t, err := jwt.ParseWithClaims(tk, &token.ResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := t.Claims.(*token.ResetClaims)
	if ok {
		return claims, nil
	}

	return nil, errors.New("jwt is invalid")
}
