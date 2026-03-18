package outbox

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"golive/auth/internal/infrastructure/postgres"
)

type Publisher interface {
	Publish(ctx context.Context, key string, value []byte) error
}

type Worker struct {
	Repo      *postgres.OutboxRepository
	Publisher Publisher

	PollInterval time.Duration
	BatchSize    int
}

type userRegisteredEvent struct {
	Type       string    `json:"type"`
	UserID     string    `json:"user_id"`
	Login      string    `json:"login"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (w Worker) Start(ctx context.Context) {
	if w.Repo == nil {
		return
	}
	if w.PollInterval <= 0 {
		w.PollInterval = 500 * time.Millisecond
	}
	if w.BatchSize <= 0 {
		w.BatchSize = 50
	}

	go func() {
		t := time.NewTicker(w.PollInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w Worker) tick(ctx context.Context) {
	msgs, err := w.Repo.ClaimPending(ctx, w.BatchSize)
	if err != nil {
		log.Printf("outbox: claim pending failed: %v", err)
		return
	}
	if len(msgs) == 0 {
		return
	}
	for _, m := range msgs {
		if err := w.publish(ctx, m); err != nil {
			attempt := m.Attempts + 1
			delay := backoff(attempt)
			nextAt := time.Now().UTC().Add(delay)
			_ = w.Repo.MarkFailed(ctx, m.ID, attempt, truncate(err.Error(), 1000), nextAt)
			continue
		}
		_ = w.Repo.MarkPublished(ctx, m.ID)
	}
}

func (w Worker) publish(ctx context.Context, m postgres.OutboxMessage) error {
	switch strings.ToLower(strings.TrimSpace(m.EventType)) {
	case "user.registered":
		var ev userRegisteredEvent
		if err := json.Unmarshal(m.Payload, &ev); err != nil {
			return err
		}
		return w.Publisher.Publish(ctx, ev.UserID, m.Payload)
	default:
		// Unknown event type — mark as published to avoid endless retries.
		return nil
	}
}

func backoff(attempt int) time.Duration {
	// 1s, 2s, 4s, ... up to 30s
	d := time.Second << max(0, min(attempt-1, 5))
	if d > 30*time.Second {
		return 30 * time.Second
	}
	return d
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

