package profile

import "context"

type UpsertParams struct {
	UserID      string
	Login       string
	DisplayName string
	AvatarURL   string
	Bio         string
}

type Repository interface {
	GetByUserID(ctx context.Context, userID string) (Profile, error)
	GetByLogin(ctx context.Context, login string) (Profile, error)
	UpsertByUserID(ctx context.Context, p UpsertParams) (Profile, error)
	SearchByLogin(ctx context.Context, query string, limit int) ([]Profile, error)
}

