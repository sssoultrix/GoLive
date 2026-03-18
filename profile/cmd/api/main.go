package main

import (
	"context"
	"log"
	"os"
	"time"

	"golive/profile/internal/application/profile"
	"golive/profile/internal/infrastructure/jwt_agent"
	"golive/profile/internal/infrastructure/postgres"
	"golive/profile/internal/infrastructure/user_events"
	"golive/profile/internal/interfaces/http/handlers"
	"golive/profile/internal/interfaces/http/middleware"
	"golive/profile/internal/interfaces/http/router"

	"github.com/gin-gonic/gin"
)

func main() {
	addr := getenv("PROFILE_HTTP_ADDR", "0.0.0.0:8081")
	logLevel := getenv("PROFILE_LOG_LEVEL", "info")

	accessTTL := mustDuration(getenv("JWT_ACCESS_TTL", "15m"))
	jwtSecret := getenv("JWT_SECRET_KEY", "")
	kafkaTopicUserEvents := getenv("KAFKA_TOPIC_USER_EVENTS", "user.events.v1")
	kafkaGroupProfile := getenv("KAFKA_GROUP_PROFILE", "profile-service")
	kafkaBrokers := getenv("KAFKA_BROKERS", "")

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, postgres.Config{
		Host:     getenv("POSTGRES_HOST", "postgres"),
		Port:     getenv("POSTGRES_PORT", "5432"),
		DB:       getenv("POSTGRES_DB", "profile"),
		User:     getenv("POSTGRES_USER", "profile"),
		Password: getenv("POSTGRES_PASSWORD", "profile"),
		SSLMode:  getenv("POSTGRES_SSLMODE", "disable"),
	})
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	profileRepo := postgres.NewProfileRepository(pool)
	jwtAgent := jwt_agent.New(jwtSecret, accessTTL)

	user_events.Start(ctx, user_events.Config{
		Brokers: splitCSV(kafkaBrokers),
		Group:   kafkaGroupProfile,
		Topic:   kafkaTopicUserEvents,
	}, profileRepo)

	getMeUC := &profile.GetMeUseCase{
		Profiles: profileRepo,
	}
	upsertMeUC := &profile.UpsertMeUseCase{
		Profiles: profileRepo,
	}
	getByLoginUC := &profile.GetByLoginUseCase{
		Profiles: profileRepo,
	}
	searchByLoginUC := &profile.SearchByLoginUseCase{
		Profiles: profileRepo,
	}

	if logLevel == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	h := &handlers.ProfileHandler{
		GetMe:         getMeUC,
		UpsertMe:      upsertMeUC,
		GetByLogin:    getByLoginUC,
		SearchByLogin: searchByLoginUC,
	}
	r := router.New(router.Deps{
		Profile:      h,
		ProfileGuard: middleware.RequireBearerJWT(jwtAgent),
	})

	log.Printf("profile api listening on %s", addr)
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
			s := trimSpaces(cur)
			if s != "" {
				out = append(out, s)
			}
			cur = ""
			continue
		}
		cur += string(v[i])
	}
	s := trimSpaces(cur)
	if s != "" {
		out = append(out, s)
	}
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

