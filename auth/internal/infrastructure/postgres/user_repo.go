package postgres

import (
	"context"
	"errors"
	"time"

	"golive/auth/internal/domain/user"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	const q = `
INSERT INTO users (login, password_hash)
VALUES ($1, $2)
RETURNING id, login, password_hash, created_at;
`
	var out user.User
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, q, u.Login, u.PasswordHash).Scan(&out.ID, &out.Login, &out.PasswordHash, &createdAt)
	if err != nil {
		return user.User{}, err
	}
	out.CreatedAt = createdAt
	return out, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (user.User, error) {
	const q = `SELECT id, login, password_hash, created_at FROM users WHERE login=$1;`
	var out user.User
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, q, login).Scan(&out.ID, &out.Login, &out.PasswordHash, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, pgx.ErrNoRows
		}
		return user.User{}, err
	}
	out.CreatedAt = createdAt
	return out, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (user.User, error) {
	const q = `SELECT id, login, password_hash, created_at FROM users WHERE id=$1;`
	var out user.User
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, q, id).Scan(&out.ID, &out.Login, &out.PasswordHash, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, pgx.ErrNoRows
		}
		return user.User{}, err
	}
	out.CreatedAt = createdAt
	return out, nil
}

