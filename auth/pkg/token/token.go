package token

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/google/uuid"
)

// NewRefreshToken returns an opaque refresh token string.
// We use UUIDv4 as a simple, transport-friendly token format.
func NewRefreshToken() string {
	return uuid.NewString()
}

// Hash returns deterministic SHA-256 hex hash (good for DB lookup).
func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

