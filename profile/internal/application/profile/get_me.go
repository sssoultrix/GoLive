package profile

import (
	"context"
	"errors"
	"strings"

	domain "golive/profile/internal/domain/profile"

	"github.com/jackc/pgx/v5"
)

type GetMeUseCase struct {
	Profiles domain.Repository
}

type GetMeInput struct {
	UserID string
}

func (uc *GetMeUseCase) Execute(ctx context.Context, in GetMeInput) (domain.Profile, error) {
	if strings.TrimSpace(in.UserID) == "" {
		return domain.Profile{}, errors.New("user id is required")
	}
	p, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Profile{}, pgx.ErrNoRows
		}
		return domain.Profile{}, err
	}
	return p, nil
}

