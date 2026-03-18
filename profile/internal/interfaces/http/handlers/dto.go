package handlers

import "time"

type upsertMeRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
}

type profileResponse struct {
	UserID      string    `json:"user_id"`
	Login       string    `json:"login"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type searchResponse struct {
	Exact   *profileResponse  `json:"exact"`
	Matches []profileResponse `json:"matches"`
}

