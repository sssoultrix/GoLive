package postgres

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string
	Port     string
	DB       string
	User     string
	Password string
	SSLMode  string
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.DB,
	}
	q := u.Query()
	if cfg.SSLMode != "" {
		q.Set("sslmode", cfg.SSLMode)
	}
	u.RawQuery = q.Encode()

	pc, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, err
	}
	pc.MaxConnLifetime = time.Hour

	return pgxpool.NewWithConfig(ctx, pc)
}

