package auth

import (
	"context"
	"errors"

	"golive/auth/internal/domain/session"
	"golive/auth/pkg/token"
)

type LogoutInput struct {
	RefreshToken string
}

type LogoutUseCase struct {
	Sessions session.Repository
}

func (uc *LogoutUseCase) Execute(ctx context.Context, in LogoutInput) error {
	if in.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return uc.Sessions.DeleteByRefreshTokenHash(ctx, token.Hash(in.RefreshToken))
}

