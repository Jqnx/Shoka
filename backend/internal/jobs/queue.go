package jobs

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"time"
)

// Queue wraps the job-table queries and implements the enqueue, claim,
// retry, and completion transitions the worker relies on.
type Queue struct {
	queries *sqlc.Queries
	log     *slog.Logger
}

// NewQueue returns a Queue backed by the given job-table queries.
func NewQueue(queries *sqlc.Queries, log *slog.Logger) *Queue {
	return &Queue{
		queries: queries,
		log:     log.With("component", "job_queue"),
	}
}

// Enqueue marshals payload and inserts a new job of jobType that becomes
// immediately claimable by a worker.
func (q *Queue) Enqueue(ctx context.Context, jobType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return q.enqueue(ctx, jobType, string(b))
}

func (q *Queue) enqueue(ctx context.Context, jobType, payload string) error {
	if err := q.queries.EnqueueJob(ctx, sqlc.EnqueueJobParams{
		Type:    jobType,
		Payload: payload,
	}); err != nil {
		q.log.Error("failed to enqueue job", "type", jobType, "error", err)
		return err
	}

	return nil
}

// HasPendingJob reports whether a job of jobType with the exact same payload
// is pending or running. Streaming handlers use it to tell "generation is
// still coming" from "the job that would report progress already finished".
func (q *Queue) HasPendingJob(ctx context.Context, jobType string, payload any) (bool, error) {
	b, err := json.Marshal(payload)
	if err != nil {
		return false, err
	}

	return q.queries.HasPendingJob(ctx, sqlc.HasPendingJobParams{
		Type:    jobType,
		Payload: string(b),
	})
}

// EnqueueOnce enqueues a job unless one with the same type AND payload is
// already pending or running — e.g. two triggers to scan the same library,
// or to generate thumbnails for the same archive, shouldn't both queue up.
// Jobs with a different payload (a different library/archive/etc.) are
// never suppressed by this check.
func (q *Queue) EnqueueOnce(ctx context.Context, jobType string, payload any) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	exists, err := q.queries.HasPendingJob(ctx, sqlc.HasPendingJobParams{
		Type:    jobType,
		Payload: string(b),
	})
	if err != nil {
		return err
	}

	if exists {
		q.log.Info("job already queued, skipping", "type", jobType)
		return nil
	}

	return q.enqueue(ctx, jobType, string(b))
}

// EnqueueAfter inserts a job of jobType that only becomes claimable once
// delay has elapsed.
func (q *Queue) EnqueueAfter(ctx context.Context, jobType string, payload any, delay time.Duration) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// UTC so the stored value matches the column's datetime('now')
	// default and ClaimJob's comparison basis.
	runAfter := time.Now().UTC().Add(delay)

	err = q.queries.EnqueueJobAfter(ctx, sqlc.EnqueueJobAfterParams{
		Type:     jobType,
		Payload:  string(b),
		RunAfter: runAfter,
	})
	if err != nil {
		q.log.Error("failed to enqueue job", "type", jobType, "error", err)
	}

	return err
}

func (q *Queue) claim(ctx context.Context) (*Job, error) {
	row, err := q.queries.ClaimJob(ctx)
	if errors.Is(err, sql.ErrNoRows) {
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
	return q.queries.MarkJobAsDone(ctx, id)
}

func (q *Queue) markFailed(ctx context.Context, id int64, jobErr error) error {
	jobError := jobErr.Error()

	return q.queries.MarkJobAsFailed(ctx, sqlc.MarkJobAsFailedParams{
		Error: &jobError,
		ID:    id,
	})
}

func (q *Queue) requeueForRetry(ctx context.Context, id, attempts int64) error {
	backoff := time.Duration(attempts*attempts) * 30 * time.Second
	runAfter := time.Now().UTC().Add(backoff)

	return q.queries.RequeueJob(ctx, sqlc.RequeueJobParams{
		RunAfter: runAfter,
		ID:       id,
	})
}
