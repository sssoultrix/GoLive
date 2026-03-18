package profile

import (
	"context"
	"errors"
	"strings"

	domain "golive/profile/internal/domain/profile"
)

type SearchByLoginUseCase struct {
	Profiles domain.Repository
}

type SearchByLoginInput struct {
	Query string
	Limit int
}

func (uc *SearchByLoginUseCase) Execute(ctx context.Context, in SearchByLoginInput) ([]domain.Profile, error) {
	if strings.TrimSpace(in.Query) == "" {
		return []domain.Profile{}, errors.New("query is required")
	}
	return uc.Profiles.SearchByLogin(ctx, in.Query, in.Limit)
}

