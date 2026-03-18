package postgres

import (
	"context"
	"encoding/json"
	"time"

	"golive/auth/internal/domain/session"
	"golive/auth/internal/domain/user"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RegisterTx struct {
	pool *pgxpool.Pool
}

func NewRegisterTx(pool *pgxpool.Pool) *RegisterTx {
	return &RegisterTx{pool: pool}
}

type RegisterTxInput struct {
	Login        string
	PasswordHash string

	RefreshTokenHash string
	UserAgent        *string
	IP               *string
	ExpiresAt        time.Time

	OccurredAt time.Time
}

type userRegisteredEvent struct {
	Type       string    `json:"type"`
	UserID     string    `json:"user_id"`
	Login      string    `json:"login"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (r *RegisterTx) Execute(ctx context.Context, in RegisterTxInput) (createdUser user.User, err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return user.User{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	// 1) Create user
	const createUserQ = `
INSERT INTO users (login, password_hash)
VALUES ($1, $2)
RETURNING id, login, password_hash, created_at;
`
	var createdAt time.Time
	if err := tx.QueryRow(ctx, createUserQ, in.Login, in.PasswordHash).Scan(&createdUser.ID, &createdUser.Login, &createdUser.PasswordHash, &createdAt); err != nil {
		return user.User{}, err
	}
	createdUser.CreatedAt = createdAt

	// 2) Create refresh session
	const createSessionQ = `
INSERT INTO refresh_sessions (user_id, refresh_token_hash, user_agent, ip, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, refresh_token_hash, user_agent, ip, expires_at, created_at;
`
	var s session.RefreshSession
	var sCreatedAt time.Time
	if err := tx.QueryRow(ctx, createSessionQ, createdUser.ID, in.RefreshTokenHash, in.UserAgent, in.IP, in.ExpiresAt).
		Scan(&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IP, &s.ExpiresAt, &sCreatedAt); err != nil {
		return user.User{}, err
	}

	// 3) Insert outbox event
	ev := userRegisteredEvent{
		Type:       "user.registered",
		UserID:     createdUser.ID.String(),
		Login:      createdUser.Login,
		OccurredAt: in.OccurredAt,
	}
	payload, err := json.Marshal(ev)
	if err != nil {
		return user.User{}, err
	}

	const insertOutboxQ = `
INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload, status, next_attempt_at)
VALUES ($1, $2, $3, $4::jsonb, 'pending', now());
`
	if _, err := tx.Exec(ctx, insertOutboxQ, "user", createdUser.ID.String(), "user.registered", payload); err != nil {
		return user.User{}, err
	}

	return createdUser, nil
}

