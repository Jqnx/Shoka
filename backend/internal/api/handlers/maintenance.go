package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// minMaintenanceIntervalMinutes mirrors the library scan-interval floor: any
// non-zero schedule must be at least this coarse.
const minMaintenanceIntervalMinutes = 15

// MaintenanceTaskResponse is one scheduled housekeeping task, as returned by
// the maintenance endpoints.
type MaintenanceTaskResponse struct {
	Task            string `json:"task"`
	IntervalMinutes int64  `json:"interval_minutes"`
	// Omitted when the task has never run.
	LastRunAt string `json:"last_run_at,omitempty"`
}

// UpdateMaintenanceTaskRequest carries a new schedule for one task. A nil
// interval is rejected rather than treated as 0, so an empty body cannot
// silently disable a schedule.
type UpdateMaintenanceTaskRequest struct {
	IntervalMinutes *int64 `json:"interval_minutes"`
}

func toMaintenanceTaskResponse(t sqlc.MaintenanceTask) MaintenanceTaskResponse {
	resp := MaintenanceTaskResponse{
		Task:            t.Task,
		IntervalMinutes: t.IntervalMinutes,
	}

	if t.LastRunAt != nil {
		resp.LastRunAt = t.LastRunAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return resp
}

// GetMaintenanceTasks godoc
//
//	@Summary	List maintenance tasks
//	@Tags		admin
//	@Produce	json
//	@Success	200	{array}		MaintenanceTaskResponse
//	@Failure	403	{object}	response.Error
//	@Failure	500	{object}	response.Error
//	@Router		/api/admin/maintenance [get]
func (h *AdminHandler) GetMaintenanceTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Queries.GetMaintenanceTasks(r.Context())
	if err != nil {
		h.Log.Error("get maintenance tasks failed", "error", err)
		response.InternalError(w, "failed to get maintenance tasks")

		return
	}

	items := make([]MaintenanceTaskResponse, 0, len(rows))
	for _, t := range rows {
		items = append(items, toMaintenanceTaskResponse(t))
	}

	response.JSON(w, http.StatusOK, items)
}

// UpdateMaintenanceTask godoc
//
//	@Summary	Schedule a maintenance task
//	@Description	Sets how often the task runs. 0 disables scheduling; any other value must be at least 15 minutes.
//	@Tags		admin
//	@Accept		json
//	@Produce	json
//	@Param		task	path		string						true	"Maintenance task (cover or phash)"
//	@Param		body	body		UpdateMaintenanceTaskRequest	true	"New schedule"
//	@Success	200		{object}	MaintenanceTaskResponse
//	@Failure	400		{object}	response.Error
//	@Failure	403		{object}	response.Error
//	@Failure	404		{object}	response.Error
//	@Failure	500		{object}	response.Error
//	@Router		/api/admin/maintenance/{task} [patch]
func (h *AdminHandler) UpdateMaintenanceTask(w http.ResponseWriter, r *http.Request) {
	task := chi.URLParam(r, "task")

	if _, err := h.Queries.GetMaintenanceTask(r.Context(), task); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "maintenance task not found")
			return
		}

		h.Log.Error("get maintenance task failed", "task", task, "error", err)
		response.InternalError(w, "failed to get maintenance task")

		return
	}

	var body UpdateMaintenanceTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if body.IntervalMinutes == nil {
		response.BadRequest(w, "interval_minutes is required")
		return
	}

	interval := *body.IntervalMinutes
	if interval < 0 {
		response.BadRequest(w, "interval_minutes cannot be negative")
		return
	}

	if interval > 0 && interval < minMaintenanceIntervalMinutes {
		response.BadRequest(w, fmt.Sprintf("interval_minutes must be 0 (disabled) or at least %d", minMaintenanceIntervalMinutes))

		return
	}

	err := h.Queries.UpdateMaintenanceTaskInterval(r.Context(), sqlc.UpdateMaintenanceTaskIntervalParams{
		IntervalMinutes: interval,
		Task:            task,
	})
	if err != nil {
		h.Log.Error("update maintenance task failed", "task", task, "error", err)
		response.InternalError(w, "failed to update maintenance task")

		return
	}

	updated, err := h.Queries.GetMaintenanceTask(r.Context(), task)
	if err != nil {
		h.Log.Error("get maintenance task failed", "task", task, "error", err)
		response.InternalError(w, "failed to update maintenance task")

		return
	}

	response.JSON(w, http.StatusOK, toMaintenanceTaskResponse(updated))
}
