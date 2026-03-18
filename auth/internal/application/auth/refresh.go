package auth

import (
	"context"
	"errors"
	"time"

	"golive/auth/internal/domain/session"
	"golive/auth/internal/domain/user"
	"golive/auth/internal/infrastructure/jwt_agent"
	"golive/auth/pkg/token"

	"github.com/jackc/pgx/v5"
)

type RefreshInput struct {
	RefreshToken string
	Now          time.Time
}

type RefreshUseCase struct {
	Users      user.Repository
	Sessions   session.Repository
	JWT        *jwt_agent.Agent
	RefreshTTL time.Duration
}

func (uc *RefreshUseCase) Execute(ctx context.Context, in RefreshInput) (TokenPair, error) {
	if in.RefreshToken == "" {
		return TokenPair{}, errors.New("refresh_token is required")
	}

	oldHash := token.Hash(in.RefreshToken)
	sess, err := uc.Sessions.GetByRefreshTokenHash(ctx, oldHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, errors.New("invalid refresh token")
		}
		return TokenPair{}, err
	}
	if !sess.ExpiresAt.After(in.Now) {
		_ = uc.Sessions.DeleteByID(ctx, sess.ID)
		return TokenPair{}, errors.New("refresh token expired")
	}

	u, err := uc.Users.GetByID(ctx, sess.UserID)
	if err != nil {
		return TokenPair{}, err
	}

	// Rotate refresh token: delete old session and create a new one.
	if err := uc.Sessions.DeleteByID(ctx, sess.ID); err != nil {
		return TokenPair{}, err
	}

	newRefreshRaw := token.NewRefreshToken()
	newRefreshHash := token.Hash(newRefreshRaw)
	_, err = uc.Sessions.Create(ctx, session.RefreshSession{
		UserID:           u.ID,
		RefreshTokenHash: newRefreshHash,
		UserAgent:        sess.UserAgent,
		IP:               sess.IP,
		ExpiresAt:        in.Now.Add(uc.RefreshTTL),
	})
	if err != nil {
		return TokenPair{}, err
	}

	access, exp, err := uc.JWT.MintAccessToken(u.ID, u.Login, in.Now)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  access,
		RefreshToken: newRefreshRaw,
		ExpiresAt:    exp,
	}, nil
}

