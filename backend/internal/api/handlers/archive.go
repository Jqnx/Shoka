package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/metadata"

	"github.com/go-chi/chi/v5"
)

type ArchiveHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

func NewArchiveHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *ArchiveHandler {
	return &ArchiveHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "archive"),
	}
}

type ArchiveResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Summary     *string    `json:"summary"`
	Language    *string    `json:"language"`
	Category    *string    `json:"category"`
	ReleaseDate *time.Time `json:"release_date"`
	PageCount   int        `json:"page_count"`
	FilePath    string     `json:"file_path"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Artists    []string `json:"artists"`
	Tags       []string `json:"tags"`
	Parodies   []string `json:"parodies"`
	Circles    []string `json:"circles"`
	Characters []string `json:"characters"`

	Progress *ProgressResponse `json:"progress"`

	ThumbsReady bool `json:"thumbs_ready"`
}

type ProgressResponse struct {
	CurrentPage int       `json:"current_page"`
	LastRead    time.Time `json:"last_read"`
	Completed   bool      `json:"completed"`
}

// GetArchive godoc
//
//	@Summary		Get a single archive
//	@Tags			archives
//	@Produce		json
//	@Param			id	path		string	true	"Archive ID"
//	@Success		200	{object}	ArchiveResponse
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/archives/{id} [get]
func (h *ArchiveHandler) GetArchive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "archive not found")
			return
		}
		h.logger.Error("get archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to get archive")
		return
	}

	// fetch relational metadata and progress concurrently
	type result struct {
		meta     *metadata.Result
		progress *sqlc.Progress
		metaErr  error
		progErr  error
	}

	ch := make(chan result, 1)
	go func() {
		var res result
		res.meta, res.metaErr = database.GetArchiveMetadata(r.Context(), h.queries, id)
		progress, err := h.queries.GetProgressForArchive(r.Context(), sqlc.GetProgressForArchiveParams{
			ArchiveID: archive.ID,
			UserID:    userID,
		})

		if err != nil {
			res.progress = nil
			res.progErr = err
		} else {
			res.progress = &progress
			res.progErr = err
		}
		ch <- res
	}()

	res := <-ch

	if res.metaErr != nil {
		h.logger.Error("get archive metadata failed", "id", id, "error", res.metaErr)
		response.InternalError(w, "failed to get archive metadata")
		return
	}

	if res.progErr != nil && !errors.Is(res.progErr, sql.ErrNoRows) {
		h.logger.Error("get progress failed", "id", id, "error", res.progErr)
		response.InternalError(w, "failed to get reading progress")
		return
	}

	response.JSON(w, http.StatusOK, h.buildResponse(archive, res.meta, res.progress))
}

func (h *ArchiveHandler) buildResponse(
	archive sqlc.Archive,
	meta *metadata.Result,
	progress *sqlc.Progress,
) ArchiveResponse {
	resp := ArchiveResponse{
		ID:          archive.ID,
		Title:       archive.Title,
		Summary:     archive.Summary,
		Language:    archive.Language,
		Category:    archive.Category,
		ReleaseDate: archive.ReleaseDate,
		PageCount:   int(archive.PageCount),
		CreatedAt:   archive.CreatedAt,
		UpdatedAt:   archive.UpdatedAt,
		ThumbsReady: h.processor.ThumbsReady(archive.ID, int(archive.PageCount)),
	}

	if meta != nil {
		resp.Artists = meta.Artists
		resp.Tags = meta.Tags
		resp.Parodies = meta.Parodies
		resp.Circles = meta.Circles
		resp.Characters = meta.Characters
	}

	if progress != nil {
		resp.Progress = &ProgressResponse{
			CurrentPage: int(progress.Page),
			LastRead:    progress.LastRead,
			Completed:   progress.Completed,
		}
	}

	return resp
}
