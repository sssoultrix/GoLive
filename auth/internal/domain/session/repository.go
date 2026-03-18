package session

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, s RefreshSession) (RefreshSession, error)
	GetByRefreshTokenHash(ctx context.Context, tokenHash string) (RefreshSession, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	DeleteByRefreshTokenHash(ctx context.Context, tokenHash string) error
	DeleteExpired(ctx context.Context) (int64, error)
}

