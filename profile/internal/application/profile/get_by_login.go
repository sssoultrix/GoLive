package profile

import (
	"context"
	"errors"
	"strings"

	domain "golive/profile/internal/domain/profile"

	"github.com/jackc/pgx/v5"
)

type GetByLoginUseCase struct {
	Profiles domain.Repository
}

type GetByLoginInput struct {
	Login string
}

func (uc *GetByLoginUseCase) Execute(ctx context.Context, in GetByLoginInput) (domain.Profile, error) {
	if strings.TrimSpace(in.Login) == "" {
		return domain.Profile{}, errors.New("login is required")
	}
	p, err := uc.Profiles.GetByLogin(ctx, in.Login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Profile{}, pgx.ErrNoRows
		}
		return domain.Profile{}, err
	}
	return p, nil
}

