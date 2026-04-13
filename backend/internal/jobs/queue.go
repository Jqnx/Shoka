package jobs

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"Shoka/internal/database/sqlc"
)

type Queue struct {
	queries *sqlc.Queries
	log     *slog.Logger
}

func NewQueue(queries *sqlc.Queries, log *slog.Logger) *Queue {
	return &Queue{
		queries: queries,
		log:     log.With("component", "job_queue"),
	}
}

func (q *Queue) Enqueue(ctx context.Context, jobType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if err := q.queries.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		Type:    jobType,
		Payload: string(b),
	}); err != nil {
		q.log.Error("failed to enqueue job", "type", jobType, "error", err)
	}

	return err
}

func (q *Queue) EnqueueOnce(ctx context.Context, jobType string, payload any) error {
	exists, err := q.queries.HasPendingJob(ctx, jobType)
	if err != nil {
		return err
	}
	if exists {
		q.log.Info("job already queued, skipping", "type", jobType)
		return nil
	}
	return q.Enqueue(ctx, jobType, payload)
}

func (q *Queue) EnqueueAfter(ctx context.Context, jobType string, payload any, delay time.Duration) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	runAfter := time.Now().Add(delay)

	if err := q.queries.EnqueueJobAfter(ctx, sqlc.EnqueueJobAfterParams{
		Type:     jobType,
		Payload:  string(b),
		RunAfter: runAfter,
	}); err != nil {
		q.log.Error("failed to enqueue job", "type", jobType, "error", err)
	}

	return err
}

func (q *Queue) claim(ctx context.Context) (*Job, error) {
	row, err := q.queries.ClaimJob(ctx)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	job := Job{
		ID:          row.ID,
		Type:        row.Type,
		RawPayload:  row.Payload,
		Attempts:    row.Attempts,
		MaxAttempts: row.MaxAttempts,
	}

	return &job, nil
}

func (q *Queue) markDone(ctx context.Context, id int64) error {
	err := q.queries.MarkJobAsDone(ctx, id)
	return err
}

func (q *Queue) markFailed(ctx context.Context, id int64, jobErr error) error {
	jobError := jobErr.Error()
	err := q.queries.MarkJobAsFailed(ctx, sqlc.MarkJobAsFailedParams{
		Error: &jobError,
		ID:    id,
	})
	return err
}

func (q *Queue) requeueForRetry(ctx context.Context, id int64, attempts int64) error {
	backoff := time.Duration(attempts*attempts) * 30 * time.Second
	runAfter := time.Now().Add(backoff)

	err := q.queries.RequeueJob(ctx, sqlc.RequeueJobParams{
		RunAfter: runAfter,
		ID:       id,
	})

	return err
}
