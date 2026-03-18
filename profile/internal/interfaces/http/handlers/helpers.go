package handlers

import domain "golive/profile/internal/domain/profile"

func toProfileResponse(p domain.Profile) profileResponse {
	return profileResponse{
		UserID:      p.UserID,
		Login:       p.Login,
		DisplayName: p.DisplayName,
		AvatarURL:   p.AvatarURL,
		Bio:         p.Bio,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

