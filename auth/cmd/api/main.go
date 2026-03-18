package main

import (
	"context"
	"log"
	"os"
	"time"

	"golive/auth/internal/application/auth"
	"golive/auth/internal/infrastructure/kafka_native"
	"golive/auth/internal/infrastructure/jwt_agent"
	"golive/auth/internal/infrastructure/outbox"
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
	profileBaseURL := getenv("PROFILE_BASE_URL", "")
	kafkaTopicUserEvents := getenv("KAFKA_TOPIC_USER_EVENTS", "user.events.v1")
	kafkaBrokers := getenv("KAFKA_BROKERS", "")

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
	registerTx := postgres.NewRegisterTx(pool)
	outboxRepo := postgres.NewOutboxRepository(pool)
	jwtAgent := jwt_agent.New(jwtSecret, accessTTL)

	var brokers []string
	for _, b := range splitCSV(kafkaBrokers) {
		if b != "" {
			brokers = append(brokers, b)
		}
	}
	pub, err := kafka_native.New(brokers, kafkaTopicUserEvents)
	if err != nil {
		log.Fatal(err)
	}
	defer pub.Close()

	outbox.Worker{Repo: outboxRepo, Publisher: pub}.Start(ctx)

	registerUC := &auth.RegisterUseCase{
		Users:      userRepo,
		Sessions:   sessRepo,
		RegisterTx: registerTx,
		JWT:        jwtAgent,
		RefreshTTL: refreshTTL,
		ProfileBaseURL:       profileBaseURL,
		KafkaTopicUserEvents: kafkaTopicUserEvents,
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

func splitCSV(v string) []string {
	var out []string
	cur := ""
	for i := 0; i < len(v); i++ {
		if v[i] == ',' {
			out = append(out, trimSpaces(cur))
			cur = ""
			continue
		}
		cur += string(v[i])
	}
	out = append(out, trimSpaces(cur))
	return out
}

func trimSpaces(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

