package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golive/auth/internal/domain/session"
	"golive/auth/internal/domain/user"
	"golive/auth/internal/infrastructure/jwt_agent"
	"golive/auth/internal/infrastructure/postgres"
	"golive/auth/pkg/bcrypt"
	"golive/auth/pkg/token"
)

type RegisterInput struct {
	Login     string
	Password  string
	UserAgent *string
	IP        *string
	Now       time.Time
}

type RegisterUseCase struct {
	Users      user.Repository
	Sessions   session.Repository
	RegisterTx *postgres.RegisterTx
	JWT        *jwt_agent.Agent
	RefreshTTL time.Duration

	// Optional: publish to Kafka via REST Proxy.
	KafkaRestURL         string
	KafkaTopicUserEvents string

	// Optional: HTTP fallback to create profile directly.
	ProfileBaseURL string

	HTTPClient *http.Client
}

func (uc *RegisterUseCase) Execute(ctx context.Context, in RegisterInput) (TokenPair, error) {
	if in.Login == "" || in.Password == "" {
		return TokenPair{}, errors.New("login and password are required")
	}

	passHash, err := bcrypt.HashPassword(in.Password)
	if err != nil {
		return TokenPair{}, err
	}

	refreshRaw := token.NewRefreshToken()
	refreshHash := token.Hash(refreshRaw)

	var createdUser user.User
	if uc.RegisterTx != nil {
		createdUser, err = uc.RegisterTx.Execute(ctx, postgres.RegisterTxInput{
			Login:            in.Login,
			PasswordHash:     passHash,
			RefreshTokenHash: refreshHash,
			UserAgent:        in.UserAgent,
			IP:               in.IP,
			ExpiresAt:        in.Now.Add(uc.RefreshTTL),
			OccurredAt:       in.Now,
		})
		if err != nil {
			return TokenPair{}, err
		}
	} else {
		// Fallback without outbox/transaction (should not be used in production).
		createdUser, err = uc.Users.Create(ctx, user.User{
			Login:        in.Login,
			PasswordHash: passHash,
		})
		if err != nil {
			return TokenPair{}, err
		}
		_, err = uc.Sessions.Create(ctx, session.RefreshSession{
			UserID:           createdUser.ID,
			RefreshTokenHash: refreshHash,
			UserAgent:        in.UserAgent,
			IP:               in.IP,
			ExpiresAt:        in.Now.Add(uc.RefreshTTL),
		})
		if err != nil {
			return TokenPair{}, err
		}
	}

	access, exp, err := uc.JWT.MintAccessToken(createdUser.ID, createdUser.Login, in.Now)
	if err != nil {
		return TokenPair{}, err
	}

	// Event publishing is handled by outbox worker after commit.
	uc.tryCreateProfile(ctx, access)

	return TokenPair{
		AccessToken:  access,
		RefreshToken: refreshRaw,
		ExpiresAt:    exp,
	}, nil
}

func (uc *RegisterUseCase) httpClient() *http.Client {
	if uc.HTTPClient != nil {
		return uc.HTTPClient
	}
	return &http.Client{Timeout: 3 * time.Second}
}

func (uc *RegisterUseCase) tryCreateProfile(ctx context.Context, accessToken string) {
	base := strings.TrimSpace(uc.ProfileBaseURL)
	if base == "" {
		return
	}

	url := strings.TrimRight(base, "/") + "/api/v1/profile/me"
	body, _ := json.Marshal(map[string]any{})

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		log.Printf("profile upsert (register): build request failed: %v", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := uc.httpClient().Do(req)
	if err != nil {
		log.Printf("profile upsert (register): request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return
	}

	b, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	log.Printf("profile upsert (register): unexpected status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(b)))
}

