package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/platform/metrics"

	"github.com/google/uuid"
)

// AttemptInfo is injected into the handler context after the job row's attempt counter is incremented.
// Domains use this for budget exhaustion without reading platform.jobs columns.
type AttemptInfo struct {
	Attempt     int // 1-based current attempt (after increment)
	MaxAttempts int
}

type attemptCtxKey struct{}

func ContextWithAttempt(ctx context.Context, info AttemptInfo) context.Context {
	return context.WithValue(ctx, attemptCtxKey{}, info)
}

func AttemptFromContext(ctx context.Context) (AttemptInfo, bool) {
	info, ok := ctx.Value(attemptCtxKey{}).(AttemptInfo)
	return info, ok
}

type Handler func(ctx context.Context, payload json.RawMessage) error

type Queue struct {
	db          *sql.DB
	handlers    map[string]Handler
	maxAttempts int
	pollEvery   time.Duration
}

func NewQueue(db *sql.DB, maxAttempts int, pollEvery time.Duration) *Queue {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if pollEvery <= 0 {
		pollEvery = 2 * time.Second
	}
	return &Queue{db: db, handlers: make(map[string]Handler), maxAttempts: maxAttempts, pollEvery: pollEvery}
}

func (q *Queue) Register(jobType string, h Handler) {
	q.handlers[jobType] = h
}

func (q *Queue) Enqueue(ctx context.Context, jobType string, payload any, runAt time.Time) (uuid.UUID, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return uuid.Nil, err
	}
	id := uuid.New()
	if runAt.IsZero() {
		runAt = time.Now().UTC()
	}
	_, err = q.db.ExecContext(ctx, `
		INSERT INTO platform.jobs (id, type, payload, run_at, attempts, max_attempts, created_at)
		VALUES ($1, $2, $3, $4, 0, $5, NOW())
	`, id, jobType, body, runAt, q.maxAttempts)
	return id, err
}

func (q *Queue) Ready(ctx context.Context) error {
	return q.db.PingContext(ctx)
}

// Run polls and processes jobs until ctx is cancelled.
func (q *Queue) Run(ctx context.Context) {
	log := logging.WithContext(ctx, "jobs")
	ticker := time.NewTicker(q.pollEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("job worker stopping", nil)
			return
		case <-ticker.C:
			if err := q.processOne(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Error("job processing error", err, nil)
			}
		}
	}
}

func (q *Queue) processOne(ctx context.Context) error {
	tx, err := q.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var (
		id         uuid.UUID
		jobType    string
		payload    []byte
		attempts   int
		maxAttempt int
	)
	err = tx.QueryRowContext(ctx, `
		SELECT id, type, payload, attempts, max_attempts
		FROM platform.jobs
		WHERE completed_at IS NULL
		  AND (locked_until IS NULL OR locked_until < NOW())
		  AND run_at <= NOW()
		  AND attempts < max_attempts
		ORDER BY run_at
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`).Scan(&id, &jobType, &payload, &attempts, &maxAttempt)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE platform.jobs
		SET locked_until = NOW() + INTERVAL '5 minutes', attempts = attempts + 1
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	h, ok := q.handlers[jobType]
	if !ok {
		metrics.JobProcessed(jobType, "unknown")
		_, _ = q.db.ExecContext(ctx, `
			UPDATE platform.jobs SET last_error = $2, locked_until = NULL, completed_at = NOW() WHERE id = $1
		`, id, fmt.Sprintf("no handler for type %s", jobType))
		return nil
	}

	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	runCtx = ContextWithAttempt(runCtx, AttemptInfo{
		Attempt:     attempts + 1,
		MaxAttempts: maxAttempt,
	})
	if err := h(runCtx, payload); err != nil {
		metrics.JobProcessed(jobType, "error")
		_, _ = q.db.ExecContext(ctx, `
			UPDATE platform.jobs
			SET last_error = $2, locked_until = NULL,
			    run_at = NOW() + (INTERVAL '1 minute' * attempts),
			    completed_at = CASE WHEN attempts >= max_attempts THEN NOW() ELSE NULL END
			WHERE id = $1
		`, id, err.Error())
		return err
	}

	metrics.JobProcessed(jobType, "ok")
	_, err = q.db.ExecContext(ctx, `
		UPDATE platform.jobs SET completed_at = NOW(), locked_until = NULL, last_error = NULL WHERE id = $1
	`, id)
	return err
}
