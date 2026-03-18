package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxMessage struct {
	ID          uuid.UUID
	AggregateID string
	EventType   string
	Payload     []byte
	Attempts    int
}

type OutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *OutboxRepository {
	return &OutboxRepository{pool: pool}
}

// ClaimPending locks and returns up to limit pending messages for processing.
// Uses SKIP LOCKED to allow multiple workers.
func (r *OutboxRepository) ClaimPending(ctx context.Context, limit int) ([]OutboxMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	const q = `
WITH cte AS (
  SELECT id
  FROM outbox
  WHERE status='pending' AND next_attempt_at <= now()
  ORDER BY created_at ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
)
SELECT o.id, o.aggregate_id, o.event_type, o.payload, o.attempts
FROM outbox o
JOIN cte ON cte.id = o.id;
`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]OutboxMessage, 0, limit)
	for rows.Next() {
		var m OutboxMessage
		if err := rows.Scan(&m.ID, &m.AggregateID, &m.EventType, &m.Payload, &m.Attempts); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
UPDATE outbox
SET status='published', published_at=now(), last_error=NULL
WHERE id=$1;
`, id)
	return err
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uuid.UUID, attempt int, lastError string, nextAttemptAt time.Time) error {
	_, err := r.pool.Exec(ctx, `
UPDATE outbox
SET status='pending',
    attempts=$2,
    last_error=$3,
    next_attempt_at=$4
WHERE id=$1;
`, id, attempt, lastError, nextAttemptAt)
	return err
}

