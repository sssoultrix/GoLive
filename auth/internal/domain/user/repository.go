package user

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, u User) (User, error)
	GetByLogin(ctx context.Context, login string) (User, error)
	GetByID(ctx context.Context, id uuid.UUID) (User, error)
}

