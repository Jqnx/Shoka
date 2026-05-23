package handlers

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/language"
	"Shoka/internal/metadata"

	"github.com/go-chi/chi/v5"
)

type ArchiveHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
	cache     *image.Cache
}

func NewArchiveHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor, cache *image.Cache) *ArchiveHandler {
	return &ArchiveHandler{
		queries:   queries,
		processor: processor,
		cache:     cache,
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

type ArchiveListResponse struct {
	Items []ArchiveResponse `json:"items"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// GetArchives godoc
//
//	@Summary		List archives with pagination
//	@Tags			archives
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives [get]
func (h *ArchiveHandler) GetArchives(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	page := 1
	limit := 24

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	offset := int64((page - 1) * limit)

	total, err := h.queries.CountArchives(r.Context())
	if err != nil {
		h.logger.Error("count archives failed", "error", err)
		response.InternalError(w, "failed to count archives")
		return
	}

	rows, err := h.queries.GetArchiveList(r.Context(), sqlc.GetArchiveListParams{
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list archives failed", "error", err)
		response.InternalError(w, "failed to list archives")
		return
	}

	items := make([]ArchiveResponse, 0, len(rows))
	lc := language.NewLanguageConverter()
	for _, row := range rows {
		var lang string
		if row.Language != nil {
			lang, err = lc.ToName(*row.Language)
			if err != nil {
				lang = *row.Language
			}
		}
		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    &lang,
			Category:    row.Category,
			ReleaseDate: row.ReleaseDate,
			PageCount:   int(row.PageCount),
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			ThumbsReady: h.processor.ThumbsReady(row.ID, int(row.PageCount)),
		}
		if row.Page != nil {
			resp.Progress = &ProgressResponse{
				CurrentPage: int(*row.Page),
				Completed:   row.Completed != nil && *row.Completed,
			}
			if row.LastRead != nil {
				resp.Progress.LastRead = *row.LastRead
			}
		}
		items = append(items, resp)
	}

	response.JSON(w, http.StatusOK, ArchiveListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
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
	lc := language.NewLanguageConverter()
	var lang string
	var err error
	if archive.Language != nil {
		lang, err = lc.ToName(*archive.Language)
		if err != nil {
			lang = *archive.Language
		}
	}
	resp := ArchiveResponse{
		ID:          archive.ID,
		Title:       archive.Title,
		Summary:     archive.Summary,
		Language:    &lang,
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

// GetCover godoc
//
//	@Summary		Get the cover thumbnail for an archive
//	@Tags			archives
//	@Produce		image/webp
//	@Param			id	path	string	true	"Archive ID"
//	@Success		200
//	@Failure		404	{object}	response.Error
//	@Router			/api/archives/{id}/cover [get]
func (h *ArchiveHandler) GetCover(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	path := h.processor.ThumbPath(id, 0)
	if path == "" {
		response.NotFound(w, "cover not ready")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, path)
}

// GetPage godoc
//
//	@Summary		Get a full-resolution page for an archive
//	@Tags			archives
//	@Produce		image/webp
//	@Param			id		path	string	true	"Archive ID"
//	@Param			index	path	int		true	"Page index (0-based)"
//	@Success		200
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/pages/{index} [get]
func (h *ArchiveHandler) GetPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil || index < 0 {
		response.BadRequest(w, "invalid page index")
		return
	}

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

	if index >= int(archive.PageCount) {
		response.NotFound(w, "page not found")
		return
	}

	data, err := h.cache.GetPage(id, archive.FilePath, index)
	if err != nil {
		h.logger.Error("get page failed", "id", id, "index", index, "error", err)
		response.InternalError(w, "failed to get page")
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
