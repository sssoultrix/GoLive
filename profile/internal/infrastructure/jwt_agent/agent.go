package jwt_agent

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Agent struct {
	secret    []byte
	accessTTL time.Duration
}

func New(secret string, accessTTL time.Duration) *Agent {
	return &Agent{
		secret:    []byte(secret),
		accessTTL: accessTTL,
	}
}

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

func (a *Agent) MintAccessToken(userID uuid.UUID, login string, now time.Time) (token string, expiresAt time.Time, err error) {
	if len(a.secret) == 0 {
		return "", time.Time{}, errors.New("jwt secret is empty")
	}
	expiresAt = now.Add(a.accessTTL)
	claims := Claims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := t.SignedString(a.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return s, expiresAt, nil
}

func (a *Agent) ParseAndVerifyAccessToken(raw string) (Claims, error) {
	var out Claims
	parsed, err := jwt.ParseWithClaims(raw, &out, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		if len(a.secret) == 0 {
			return nil, errors.New("jwt secret is empty")
		}
		return a.secret, nil
	})
	if err != nil {
		return Claims{}, err
	}
	if !parsed.Valid {
		return Claims{}, errors.New("invalid token")
	}
	return out, nil
}

