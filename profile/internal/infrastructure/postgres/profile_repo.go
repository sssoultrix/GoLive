package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"golive/profile/internal/domain/profile"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

func (r *ProfileRepository) GetByUserID(ctx context.Context, userID string) (profile.Profile, error) {
	const q = `
SELECT user_id, login, display_name, avatar_url, bio, created_at, updated_at
FROM profiles
WHERE user_id=$1;
`
	var out profile.Profile
	var createdAt time.Time
	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&out.UserID,
		&out.Login,
		&out.DisplayName,
		&out.AvatarURL,
		&out.Bio,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return profile.Profile{}, pgx.ErrNoRows
		}
		return profile.Profile{}, err
	}
	out.CreatedAt = createdAt
	out.UpdatedAt = updatedAt
	return out, nil
}

func (r *ProfileRepository) GetByLogin(ctx context.Context, login string) (profile.Profile, error) {
	const q = `
SELECT user_id, login, display_name, avatar_url, bio, created_at, updated_at
FROM profiles
WHERE login=$1;
`
	var out profile.Profile
	var createdAt time.Time
	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, q, login).Scan(
		&out.UserID,
		&out.Login,
		&out.DisplayName,
		&out.AvatarURL,
		&out.Bio,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return profile.Profile{}, pgx.ErrNoRows
		}
		return profile.Profile{}, err
	}
	out.CreatedAt = createdAt
	out.UpdatedAt = updatedAt
	return out, nil
}

func (r *ProfileRepository) UpsertByUserID(ctx context.Context, p profile.UpsertParams) (profile.Profile, error) {
	const q = `
INSERT INTO profiles (user_id, login, display_name, avatar_url, bio)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE SET
  login=EXCLUDED.login,
  display_name=EXCLUDED.display_name,
  avatar_url=EXCLUDED.avatar_url,
  bio=EXCLUDED.bio,
  updated_at=now()
RETURNING user_id, login, display_name, avatar_url, bio, created_at, updated_at;
`
	var out profile.Profile
	var createdAt time.Time
	var updatedAt time.Time
	err := r.pool.QueryRow(ctx, q, p.UserID, p.Login, p.DisplayName, p.AvatarURL, p.Bio).Scan(
		&out.UserID,
		&out.Login,
		&out.DisplayName,
		&out.AvatarURL,
		&out.Bio,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return profile.Profile{}, err
	}
	out.CreatedAt = createdAt
	out.UpdatedAt = updatedAt
	return out, nil
}

func (r *ProfileRepository) SearchByLogin(ctx context.Context, query string, limit int) ([]profile.Profile, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return []profile.Profile{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	const sql = `
SELECT user_id, login, display_name, avatar_url, bio, created_at, updated_at
FROM profiles
WHERE login ILIKE '%' || $1 || '%'
ORDER BY login ASC
LIMIT $2;
`
	rows, err := r.pool.Query(ctx, sql, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]profile.Profile, 0, limit)
	for rows.Next() {
		var p profile.Profile
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(
			&p.UserID,
			&p.Login,
			&p.DisplayName,
			&p.AvatarURL,
			&p.Bio,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		p.CreatedAt = createdAt
		p.UpdatedAt = updatedAt
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

