package main

import (
	"context"
	"log"
	"os"
	"time"

	"golive/auth/internal/application/auth"
	"golive/auth/internal/infrastructure/jwt_agent"
	"golive/auth/internal/infrastructure/postgres"
	"golive/auth/internal/interfaces/http/handlers"
	"golive/auth/internal/interfaces/http/middleware"
	"golive/auth/internal/interfaces/http/router"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := getenv("AUTH_HTTP_ADDR", "0.0.0.0:8080")
	logLevel := getenv("AUTH_LOG_LEVEL", "info")

	accessTTL := mustDuration(getenv("JWT_ACCESS_TTL", "15m"))
	refreshTTL := mustDuration(getenv("JWT_REFRESH_TTL", "720h"))
	jwtSecret := getenv("JWT_SECRET_KEY", "")

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, postgres.Config{
		Host:     getenv("POSTGRES_HOST", "postgres"),
		Port:     getenv("POSTGRES_PORT", "5432"),
		DB:       getenv("POSTGRES_DB", "auth"),
		User:     getenv("POSTGRES_USER", "auth"),
		Password: getenv("POSTGRES_PASSWORD", "auth"),
		SSLMode:  getenv("POSTGRES_SSLMODE", "disable"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	userRepo := postgres.NewUserRepository(pool)
	sessRepo := postgres.NewRefreshSessionRepository(pool)
	jwtAgent := jwt_agent.New(jwtSecret, accessTTL)

	registerUC := &auth.RegisterUseCase{
		Users:      userRepo,
		Sessions:   sessRepo,
		JWT:        jwtAgent,
		RefreshTTL: refreshTTL,
	}
	loginUC := &auth.LoginUseCase{
		Users:      userRepo,
		Sessions:   sessRepo,
		JWT:        jwtAgent,
		RefreshTTL: refreshTTL,
	}
	refreshUC := &auth.RefreshUseCase{
		Users:      userRepo,
		Sessions:   sessRepo,
		JWT:        jwtAgent,
		RefreshTTL: refreshTTL,
	}
	logoutUC := &auth.LogoutUseCase{
		Sessions: sessRepo,
	}

	if logLevel == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	h := &handlers.AuthHandler{
		Register: registerUC,
		Login:    loginUC,
		Refresh:  refreshUC,
		Logout:   logoutUC,
	}
	r := router.New(router.Deps{
		Auth:      h,
		AuthGuard: middleware.RequireBearerJWT(jwtAgent),
	})

	log.Printf("auth api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("invalid duration %q: %v", v, err)
	}
	return d
}

