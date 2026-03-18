package postgres

import (
	"context"
	"errors"
	"time"

	"golive/auth/internal/domain/session"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RefreshSessionRepository struct {
	pool *pgxpool.Pool
}

func NewRefreshSessionRepository(pool *pgxpool.Pool) *RefreshSessionRepository {
	return &RefreshSessionRepository{pool: pool}
}

func (r *RefreshSessionRepository) Create(ctx context.Context, s session.RefreshSession) (session.RefreshSession, error) {
	const q = `
INSERT INTO refresh_sessions (user_id, refresh_token_hash, user_agent, ip, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, refresh_token_hash, user_agent, ip, expires_at, created_at;
`
	var out session.RefreshSession
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, q, s.UserID, s.RefreshTokenHash, s.UserAgent, s.IP, s.ExpiresAt).
		Scan(&out.ID, &out.UserID, &out.RefreshTokenHash, &out.UserAgent, &out.IP, &out.ExpiresAt, &createdAt)
	if err != nil {
		return session.RefreshSession{}, err
	}
	out.CreatedAt = createdAt
	return out, nil
}

func (r *RefreshSessionRepository) GetByRefreshTokenHash(ctx context.Context, tokenHash string) (session.RefreshSession, error) {
	const q = `
SELECT id, user_id, refresh_token_hash, user_agent, ip, expires_at, created_at
FROM refresh_sessions
WHERE refresh_token_hash=$1;
`
	var out session.RefreshSession
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, q, tokenHash).
		Scan(&out.ID, &out.UserID, &out.RefreshTokenHash, &out.UserAgent, &out.IP, &out.ExpiresAt, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return session.RefreshSession{}, pgx.ErrNoRows
		}
		return session.RefreshSession{}, err
	}
	out.CreatedAt = createdAt
	return out, nil
}

func (r *RefreshSessionRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_sessions WHERE id=$1;`, id)
	return err
}

func (r *RefreshSessionRepository) DeleteByRefreshTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM refresh_sessions WHERE refresh_token_hash=$1;`, tokenHash)
	return err
}

func (r *RefreshSessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM refresh_sessions WHERE expires_at < now();`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

