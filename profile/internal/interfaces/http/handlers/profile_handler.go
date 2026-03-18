package handlers

import "golive/profile/internal/application/profile"

type ProfileHandler struct {
	GetMe        *profile.GetMeUseCase
	UpsertMe     *profile.UpsertMeUseCase
	GetByLogin   *profile.GetByLoginUseCase
	SearchByLogin *profile.SearchByLoginUseCase
}

