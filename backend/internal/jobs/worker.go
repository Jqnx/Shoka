package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Every job type must implement this function signature.
type Handler func(ctx context.Context, job *Job) error

type Worker struct {
	queue      *Queue
	handlers   map[string]Handler
	semaphores map[string]chan struct{}
	log        *slog.Logger
	interval   time.Duration
}

func NewWorker(queue *Queue, log *slog.Logger) *Worker {
	return &Worker{
		queue:      queue,
		handlers:   make(map[string]Handler),
		semaphores: make(map[string]chan struct{}),
		log:        log.With("component", "job_worker"),
		interval:   time.Second,
	}
}

// Register() associates a job type with its handler function.
func (w *Worker) Register(jobType string, handler Handler, limit int) {
	w.handlers[jobType] = handler
	if limit > 0 {
		w.semaphores[jobType] = make(chan struct{}, limit)
	}
}

// Start() begins polling in a goroutine. Cancel the context to stop it.
func (w *Worker) Start(ctx context.Context) {
	go w.run(ctx)

	w.log.Info("job worker started", "poll_interval", w.interval)
}

func (w *Worker) run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.log.Info("job worker stopped")
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

// processNext() pulls the next job off the queue and processes it.
func (w *Worker) processNext(ctx context.Context) {
	job, err := w.queue.claim(ctx)
	if err != nil {
		w.log.Error("failed to claim job", "error", err)
		return
	}

	if job == nil {
		return // nothing to do
	}

	log := w.log.With("job_id", job.ID, "job_type", job.Type, "attempt", job.Attempts)
	log.Info("processing job")

	handler, ok := w.handlers[job.Type]
	if !ok {
		log.Error("no handler registered for job type")
		w.queue.markFailed(ctx, job.ID, fmt.Errorf("no handler for type: %s", job.Type))

		return
	}

	if sem, ok := w.semaphores[job.Type]; ok {
		sem <- struct{}{}
		defer func() { <-sem }()
	}

	if err := handler(ctx, job); err != nil {
		log.Warn("job failed", "error", err)

		if job.Attempts < job.MaxAttempts {
			w.queue.requeueForRetry(ctx, job.ID, job.Attempts)
			log.Info("job requeued for retry", "backoff_seconds", job.Attempts*job.Attempts*30)
		} else {
			w.queue.markFailed(ctx, job.ID, err)
			log.Error("job exceeded max attempts, marked failed")
		}

		return
	}

	w.queue.markDone(ctx, job.ID)
	log.Info("job completed")
}
