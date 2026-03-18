package session

import (
	"time"

	"github.com/google/uuid"
)

// RefreshSession represents one refresh-token session (device/browser).
type RefreshSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        *string
	IP               *string
	ExpiresAt        time.Time
	CreatedAt        time.Time
}

