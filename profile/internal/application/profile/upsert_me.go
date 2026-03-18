package profile

import (
	"context"
	"errors"
	"strings"
	"time"

	domain "golive/profile/internal/domain/profile"
)

type UpsertMeUseCase struct {
	Profiles domain.Repository
}

type UpsertMeInput struct {
	UserID      string
	Login       string
	DisplayName string
	AvatarURL   string
	Bio         string
	Now         time.Time
}

func (uc *UpsertMeUseCase) Execute(ctx context.Context, in UpsertMeInput) (domain.Profile, error) {
	if strings.TrimSpace(in.UserID) == "" {
		return domain.Profile{}, errors.New("user id is required")
	}
	if strings.TrimSpace(in.Login) == "" {
		return domain.Profile{}, errors.New("login is required")
	}
	return uc.Profiles.UpsertByUserID(ctx, domain.UpsertParams{
		UserID:      in.UserID,
		Login:       in.Login,
		DisplayName: strings.TrimSpace(in.DisplayName),
		AvatarURL:   strings.TrimSpace(in.AvatarURL),
		Bio:         strings.TrimSpace(in.Bio),
	})
}

