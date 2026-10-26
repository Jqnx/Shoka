// Package maintenance schedules and runs the whole-library housekeeping jobs
// exposed on /admin/maintenance (cover regeneration, perceptual hashing).
package maintenance

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Intervals are configured in minutes, so a minute is the finest useful tick.
const schedulerTick = time.Minute

// Manager owns the maintenance_task table: it runs a scheduled task on demand
// and fans a run out into one job per archive.
type Manager struct {
	queries *sqlc.Queries
	queue   *jobs.Queue
	log     *slog.Logger
}

// NewManager returns a Manager that reads schedules through queries and
// enqueues per-archive jobs on queue.
func NewManager(queries *sqlc.Queries, queue *jobs.Queue, log *slog.Logger) *Manager {
	return &Manager{queries: queries, queue: queue, log: log.With("component", "maintenance_manager")}
}

// Start launches the scheduler goroutine. Stop by cancelling ctx. State lives
// in the maintenance_task table rather than timers, so schedules survive
// restarts and edits apply on the next tick.
func (m *Manager) Start(ctx context.Context) {
	go m.runScheduler(ctx)
}

func (m *Manager) runScheduler(ctx context.Context) {
	ticker := time.NewTicker(schedulerTick)
	defer ticker.Stop()

	m.log.Info("maintenance scheduler started", "tick", schedulerTick)

	for {
		select {
		case <-ctx.Done():
			m.log.Info("maintenance scheduler stopped")
			return
		case <-ticker.C:
			m.enqueueDueTasks(ctx)
		}
	}
}

// enqueueDueTasks enqueues every due task, skipping one whose previous run is
// still pending or running: a run fans out one job per archive with plain
// Enqueue, so without the skip a slow run would stack a second full pass.
// Skipped tasks are not stamped, so they retry on the next tick.
func (m *Manager) enqueueDueTasks(ctx context.Context) {
	tasks, err := m.queries.ListMaintenanceTasksDue(ctx)
	if err != nil {
		m.log.Error("failed to list maintenance tasks due", "error", err)
		return
	}

	for _, t := range tasks {
		pending, err := m.queries.HasPendingJobByType(ctx, t.Task)
		if err != nil {
			m.log.Error("failed to check pending jobs", "task", t.Task, "error", err)
			continue
		}

		if pending {
			m.log.Info("scheduled maintenance skipped, previous run still in flight", "task", t.Task)
			continue
		}

		m.log.Info("enqueueing scheduled maintenance",
			"task", t.Task,
			"interval_minutes", t.IntervalMinutes,
		)

		if err := m.RunTask(ctx, t.Task); err != nil {
			m.log.Error("failed to run scheduled maintenance", "task", t.Task, "error", err)
		}
	}
}

// RunTask fans one job out per archive and records the attempt. Task ids are
// job types. The attempt is stamped even when individual enqueues fail: a
// task that always fails must not be re-picked every tick.
func (m *Manager) RunTask(ctx context.Context, task string) error {
	switch task {
	case jobs.JobTypeCover, jobs.JobTypePHash:
	default:
		return fmt.Errorf("unknown maintenance task %q", task)
	}

	archives, err := m.queries.GetAllArchiveFilePaths(ctx)
	if err != nil {
		return fmt.Errorf("list archives: %w", err)
	}

	for _, a := range archives {
		var payload any
		if task == jobs.JobTypeCover {
			payload = jobs.CoverPayload{ArchiveID: a.ID, FilePath: a.FilePath, PageIndex: int(a.CoverPage)}
		} else {
			payload = jobs.PHashPayload{ArchiveID: a.ID, FilePath: a.FilePath}
		}

		if err := m.queue.Enqueue(ctx, task, payload); err != nil {
			m.log.Error("enqueue maintenance job failed", "task", task, "archive_id", a.ID, "error", err)
		}
	}

	if err := m.queries.MarkMaintenanceTaskRun(ctx, task); err != nil {
		m.log.Error("failed to record maintenance run", "task", task, "error", err)
	}

	return nil
}
