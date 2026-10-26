package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"net/http"
)

// MaintenanceTaskResponse is one scheduled housekeeping task, as returned by
// the maintenance endpoints.
type MaintenanceTaskResponse struct {
	Task            string `json:"task"`
	IntervalMinutes int64  `json:"interval_minutes"`
	// Omitted when the task has never run.
	LastRunAt string `json:"last_run_at,omitempty"`
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
