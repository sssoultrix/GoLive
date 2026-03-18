package profile

import "time"

type Profile struct {
	UserID      string
	Login       string
	DisplayName string
	AvatarURL   string
	Bio         string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

