package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Bounds on the monitor's snapshot. A library-wide backfill can queue
// thousands of jobs; the view only needs enough to see what's happening now,
// and the counts below still report the true totals.
const (
	maxActiveJobs = 100
	maxFailedJobs = 20
	// Matches the workers' own poll interval - sampling faster would mostly
	// re-read an unchanged queue. A short job can still start and finish
	// between two samples and never be seen as running; the counts pick it
	// up either way, which is the right trade for a monitoring view.
	jobPollInterval = time.Second
	// Idle SSE connections can be dropped by intermediaries; a comment frame
	// keeps them open without emitting a spurious snapshot.
	jobHeartbeatInterval = 20 * time.Second
)

// jobStatuses is every value the job.status CHECK constraint allows. Used to
// zero-fill the counts map so the UI renders all four tiles even on a fresh
// database, where GetJobCounts returns only statuses that have rows.
var jobStatuses = [4]string{"pending", "running", "done", "failed"}

type JobHandler struct {
	Log     *slog.Logger
	Queries *sqlc.Queries
}

func NewJobHandler(queries *sqlc.Queries, logger *slog.Logger) *JobHandler {
	return &JobHandler{
		Queries: queries,
		Log:     logger.With("handler", "job"),
	}
}

type JobView struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
	// Target is a human-readable description of what the job acts on,
	// derived from its payload - e.g. "archive:a3f9..." or "library:main".
	// Empty when the payload has no recognisable target.
	Target      string    `json:"target"`
	Attempts    int64     `json:"attempts"`
	MaxAttempts int64     `json:"max_attempts"`
	Error       *string   `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type JobSnapshot struct {
	// Counts is keyed by status and always contains all four statuses.
	Counts map[string]int64 `json:"counts"`
	// Active is pending + running jobs, running first. Capped at
	// maxActiveJobs - Counts reports the true totals.
	Active         []JobView `json:"active"`
	RecentFailures []JobView `json:"recent_failures"`
}

// payloadTarget derives a display target from a job's JSON payload.
//
// Payload key casing is inconsistent across job types: cover/phash/thumbnail
// tag their fields (`archive_id`), while ScanPayload.LibraryID and
// MetadataPayload.ArchiveID have no tags and marshal as Go field names. Both
// spellings are checked here so the frontend never has to know the difference.
func payloadTarget(payload string) string {
	var fields map[string]any
	if err := json.Unmarshal([]byte(payload), &fields); err != nil {
		return ""
	}

	// Ordered so the most specific identifier wins for payloads carrying
	// more than one (none currently do, but scan-of-archive would).
	lookups := []struct {
		label string
		keys  [2]string
	}{
		{"archive", [2]string{"archive_id", "ArchiveID"}},
		{"library", [2]string{"library_id", "LibraryID"}},
	}

	for _, lookup := range lookups {
		for _, key := range lookup.keys {
			if v, ok := fields[key].(string); ok && v != "" {
				return fmt.Sprintf("%s:%s", lookup.label, v)
			}
		}
	}

	return ""
}

func toJobViews(rows []sqlc.Job) []JobView {
	views := make([]JobView, 0, len(rows))
	for _, row := range rows {
		views = append(views, JobView{
			ID:          row.ID,
			Type:        row.Type,
			Status:      row.Status,
			Target:      payloadTarget(row.Payload),
			Attempts:    row.Attempts,
			MaxAttempts: row.MaxAttempts,
			Error:       row.Error,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return views
}

// snapshot gathers the full monitor payload in one pass.
func (h *JobHandler) snapshot(ctx context.Context) (*JobSnapshot, error) {
	countRows, err := h.Queries.GetJobCounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get job counts: %w", err)
	}

	counts := make(map[string]int64, len(jobStatuses))
	for _, status := range jobStatuses {
		counts[status] = 0
	}

	for _, row := range countRows {
		counts[row.Status] = row.Count
	}

	active, err := h.Queries.GetActiveJobs(ctx, maxActiveJobs)
	if err != nil {
		return nil, fmt.Errorf("get active jobs: %w", err)
	}

	failed, err := h.Queries.GetRecentFailedJobs(ctx, maxFailedJobs)
	if err != nil {
		return nil, fmt.Errorf("get recent failed jobs: %w", err)
	}

	return &JobSnapshot{
		Counts:         counts,
		Active:         toJobViews(active),
		RecentFailures: toJobViews(failed),
	}, nil
}

// GetJobs godoc
//
//	@Summary		Get a snapshot of the job queue
//	@Description	Status counts across the whole queue, plus the currently pending/running jobs and the most recent failures. Point-in-time - see GET /api/admin/jobs/events to follow it live.
//	@Tags			admin
//	@Produce		json
//	@Success		200	{object}	JobSnapshot
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/jobs [get]
func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.snapshot(r.Context())
	if err != nil {
		h.Log.Error("get job snapshot failed", "error", err)
		response.InternalError(w, "failed to get jobs")

		return
	}

	response.JSON(w, http.StatusOK, snapshot)
}

// StreamJobEvents godoc
//
//	@Summary		Stream job queue activity
//	@Description	Server-Sent Events. Emits a "snapshot" event ({counts, active, recent_failures}, same shape as GET /api/admin/jobs) immediately on connect, then again whenever the queue changes. Polls once a second, matching the workers' own tick, so a job that starts and finishes inside one interval may never be observed as running - the counts still account for it. Identical consecutive snapshots are suppressed; a comment heartbeat keeps idle connections alive.
//	@Tags			admin
//	@Produce		text/event-stream
//	@Success		200
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/jobs/events [get]
func (h *JobHandler) StreamJobEvents(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		response.InternalError(w, "streaming unsupported")
		return
	}

	ctx := r.Context()

	// Take the first snapshot before writing the SSE headers so a failure
	// here can still be reported as a normal 500 rather than as an empty
	// 200 stream the client would sit waiting on.
	snapshot, err := h.snapshot(ctx)
	if err != nil {
		h.Log.Error("get job snapshot failed", "error", err)
		response.InternalError(w, "failed to get jobs")

		return
	}

	response.SSEHeaders(w)

	// Compare the encoded form rather than the struct: it has to be
	// marshalled to be sent anyway, and it sidesteps map iteration order
	// mattering to a reflect.DeepEqual on Counts.
	lastSent, err := json.Marshal(snapshot)
	if err != nil {
		h.Log.Error("marshal job snapshot failed", "error", err)
		return
	}

	if err := response.SSEEvent(w, "snapshot", snapshot); err != nil {
		return
	}

	flusher.Flush()

	poll := time.NewTicker(jobPollInterval)
	defer poll.Stop()

	heartbeat := time.NewTicker(jobHeartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}

			flusher.Flush()

		case <-poll.C:
			snapshot, err := h.snapshot(ctx)
			if err != nil {
				// Transient read failure (a locked database, say): log it
				// and keep the stream open rather than dropping the client,
				// which would otherwise reconnect into the same error.
				h.Log.Error("get job snapshot failed", "error", err)
				continue
			}

			encoded, err := json.Marshal(snapshot)
			if err != nil {
				h.Log.Error("marshal job snapshot failed", "error", err)
				continue
			}

			if bytes.Equal(encoded, lastSent) {
				continue
			}

			lastSent = encoded

			if err := response.SSEEvent(w, "snapshot", snapshot); err != nil {
				return
			}

			flusher.Flush()
		}
	}
}
