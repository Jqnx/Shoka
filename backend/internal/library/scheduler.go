package library

import (
	"Shoka/internal/jobs"
	"context"
	"time"
)

// Intervals are configured in minutes, so a minute is the finest useful tick.
const schedulerTick = time.Minute

// runScheduler enqueues a scan for each library whose interval has elapsed.
// Started by Manager.Start, stopped by cancelling ctx.
//
// State lives in the library table rather than per-library timers, so
// schedules survive restarts and edits apply on the next tick without any
// timer teardown.
func (m *Manager) runScheduler(ctx context.Context) {
	ticker := time.NewTicker(schedulerTick)
	defer ticker.Stop()

	m.log.Info("scan scheduler started", "tick", schedulerTick)

	for {
		select {
		case <-ctx.Done():
			m.log.Info("scan scheduler stopped")
			return
		case <-ticker.C:
			m.enqueueDueScans(ctx)
		}
	}
}

// enqueueDueScans enqueues every due library. EnqueueOnce suppresses a
// library already pending or running, so a slow scan never stacks up.
func (m *Manager) enqueueDueScans(ctx context.Context) {
	libs, err := m.queries.ListLibrariesDueForScan(ctx)
	if err != nil {
		m.log.Error("failed to list libraries due for scan", "error", err)
		return
	}

	for _, lib := range libs {
		if _, ok := m.scanners[lib.Type]; !ok {
			continue
		}

		m.log.Info("enqueueing scheduled scan",
			"library_id", lib.ID,
			"library", lib.Name,
			"interval_minutes", lib.ScanIntervalMinutes,
		)

		if err := m.queue.EnqueueOnce(ctx, jobs.JobTypeScan, jobs.ScanPayload{LibraryID: lib.ID}); err != nil {
			m.log.Error("failed to enqueue scheduled scan", "library_id", lib.ID, "error", err)
		}
	}
}
