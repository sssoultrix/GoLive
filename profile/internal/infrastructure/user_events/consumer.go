package user_events

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"golive/profile/internal/domain/profile"
	"golive/profile/internal/infrastructure/kafka_native"

	"github.com/twmb/franz-go/pkg/kgo"
)

type UserRegisteredEvent struct {
	Type       string    `json:"type"`
	UserID     string    `json:"user_id"`
	Login      string    `json:"login"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Config struct {
	Brokers []string
	Group   string
	Topic   string
}

func Start(ctx context.Context, cfg Config, repo profile.Repository) {
	group := strings.TrimSpace(cfg.Group)
	topic := strings.TrimSpace(cfg.Topic)
	if len(cfg.Brokers) == 0 || group == "" || topic == "" || repo == nil {
		return
	}

	go run(ctx, cfg, repo)
}

func run(ctx context.Context, cfg Config, repo profile.Repository) {
	c, err := kafka_native.NewConsumer(cfg.Brokers, cfg.Group, cfg.Topic)
	if err != nil {
		log.Printf("user events consumer: create kafka consumer failed: %v", err)
		return
	}
	defer c.Close()

	for ctx.Err() == nil {
		fetches := c.Client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			log.Printf("user events consumer: poll errors: %v", errs[0].Err)
		}
		fetches.EachRecord(func(r *kgo.Record) {
			var ev UserRegisteredEvent
			if err := json.Unmarshal(r.Value, &ev); err != nil {
				log.Printf("user events consumer: bad json: %v", err)
				return
			}
			if !strings.EqualFold(ev.Type, "user.registered") || ev.UserID == "" || ev.Login == "" {
				return
			}
			if _, err := repo.UpsertByUserID(ctx, profile.UpsertParams{
				UserID:      ev.UserID,
				Login:       ev.Login,
				DisplayName: ev.Login,
				AvatarURL:   "",
				Bio:         "Описание профиля",
			}); err != nil {
				log.Printf("user events consumer: upsert profile failed user_id=%s login=%s: %v", ev.UserID, ev.Login, err)
				return
			}
		})
		// commit offsets after processing
		if err := c.Client.CommitUncommittedOffsets(ctx); err != nil {
			log.Printf("user events consumer: commit offsets failed: %v", err)
		}
	}
}

