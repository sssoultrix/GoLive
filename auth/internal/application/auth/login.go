package auth

import (
	"context"
	"errors"
	"time"

	"golive/auth/internal/domain/session"
	"golive/auth/internal/domain/user"
	"golive/auth/internal/infrastructure/jwt_agent"
	"golive/auth/pkg/bcrypt"
	"golive/auth/pkg/token"

	"github.com/jackc/pgx/v5"
)

type LoginInput struct {
	Login     string
	Password  string
	UserAgent *string
	IP        *string
	Now       time.Time
}

type LoginUseCase struct {
	Users      user.Repository
	Sessions   session.Repository
	JWT        *jwt_agent.Agent
	RefreshTTL time.Duration
}

func (uc *LoginUseCase) Execute(ctx context.Context, in LoginInput) (TokenPair, error) {
	if in.Login == "" || in.Password == "" {
		return TokenPair{}, errors.New("login and password are required")
	}

	u, err := uc.Users.GetByLogin(ctx, in.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TokenPair{}, errors.New("invalid credentials")
		}
		return TokenPair{}, err
	}
	if err := bcrypt.ComparePasswordHash(in.Password, u.PasswordHash); err != nil {
		return TokenPair{}, errors.New("invalid credentials")
	}

	refreshRaw := token.NewRefreshToken()
	refreshHash := token.Hash(refreshRaw)

	_, err = uc.Sessions.Create(ctx, session.RefreshSession{
		UserID:           u.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        in.UserAgent,
		IP:               in.IP,
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
		RefreshToken: refreshRaw,
		ExpiresAt:    exp,
	}, nil
}

