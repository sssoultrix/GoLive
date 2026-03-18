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
)

type RegisterInput struct {
	Login     string
	Password  string
	UserAgent *string
	IP        *string
	Now       time.Time
}

type RegisterUseCase struct {
	Users      user.Repository
	Sessions   session.Repository
	JWT        *jwt_agent.Agent
	RefreshTTL time.Duration
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if in.Login == "" || in.Password == "" {
		return TokenPair{}, errors.New("login and password are required")
	}

	passHash, err := bcrypt.HashPassword(in.Password)
	if err != nil {
		return TokenPair{}, err
	}

	createdUser, err := uc.Users.Create(ctx, user.User{
		Login:        in.Login,
		PasswordHash: passHash,
	})
	if err != nil {
		return TokenPair{}, err
	}

	refreshRaw := token.NewRefreshToken()
	refreshHash := token.Hash(refreshRaw)

	_, err = uc.Sessions.Create(ctx, session.RefreshSession{
		UserID:           createdUser.ID,
		RefreshTokenHash: refreshHash,
		UserAgent:        in.UserAgent,
		IP:               in.IP,
		ExpiresAt:        in.Now.Add(uc.RefreshTTL),
	})
	if err != nil {
		return TokenPair{}, err
	}

	access, exp, err := uc.JWT.MintAccessToken(createdUser.ID, createdUser.Login, in.Now)
	if err != nil {
		return TokenPair{}, err
	}

	return TokenPair{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		ExpiresAt:    exp,
	}, nil
}

