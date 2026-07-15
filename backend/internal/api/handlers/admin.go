package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"log/slog"
	"net/http"
)

type AdminHandler struct {
	Log     *slog.Logger
	Queries *sqlc.Queries
	Queue   *jobs.Queue
}

func NewAdminHandler(queries *sqlc.Queries, queue *jobs.Queue, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		Queries: queries,
		Queue:   queue,
		Log:     logger.With("handler", "admin"),
	}
}

// GenerateCovers godoc
//
//	@Summary		Enqueue cover generation for all archives
//	@Tags			admin
//	@Success		204
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/covers [post]
func (h *AdminHandler) GenerateCovers(w http.ResponseWriter, r *http.Request) {
	archives, err := h.Queries.GetAllArchiveFilePaths(r.Context())
	if err != nil {
		h.Log.Error("get archives failed", "error", err)
		response.InternalError(w, "failed to get archives")
		return
	}

	for _, a := range archives {
		if err := h.Queue.Enqueue(r.Context(), jobs.JobTypeCover, jobs.CoverPayload{
			ArchiveID: a.ID,
			FilePath:  a.FilePath,
		}); err != nil {
			h.Log.Error("enqueue cover job failed", "archive_id", a.ID, "error", err)
		}
	}

	response.NoContent(w)
}
