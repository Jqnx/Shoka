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
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/language"
	"Shoka/internal/metadata"

	"github.com/go-chi/chi/v5"
)

type ArchiveHandler struct {
	queries     *sqlc.Queries
	db          *sql.DB
	logger      *slog.Logger
	processor   *image.Processor
	cache       *image.Cache
	queue       *jobs.Queue
	broadcaster *events.ThumbnailBroadcaster
}

func NewArchiveHandler(queries *sqlc.Queries, db *sql.DB, log *slog.Logger, processor *image.Processor, cache *image.Cache, queue *jobs.Queue, broadcaster *events.ThumbnailBroadcaster) *ArchiveHandler {
	return &ArchiveHandler{
		queries:     queries,
		db:          db,
		processor:   processor,
		cache:       cache,
		queue:       queue,
		broadcaster: broadcaster,
		logger:      log.With("handler", "archive"),
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

type SortOptionResponse struct {
	Value       string `json:"value"`
	DisplayName string `json:"display_name"`
}

// GetArchiveSortOptions godoc
//
//	@Summary		List archive sort options
//	@Description	Returns every valid value for the "sort" query param on GET /api/archives, with a display name for UI dropdowns
//	@Tags			archives
//	@Produce		json
//	@Success		200	{array}		SortOptionResponse
//	@Router			/api/archives/sort-options [get]
func (h *ArchiveHandler) GetArchiveSortOptions(w http.ResponseWriter, r *http.Request) {
	items := make([]SortOptionResponse, 0, len(database.SortOptions))
	for _, opt := range database.SortOptions {
		items = append(items, SortOptionResponse{
			Value:       opt.Value,
			DisplayName: opt.DisplayName,
		})
	}

	response.JSON(w, http.StatusOK, items)
}

// GetArchives godoc
//
//	@Summary		List archives with pagination, filtering, and sorting
//	@Tags			archives
//	@Produce		json
//	@Param			library_id	query		string		true	"Library ID to list archives from"
//	@Param			page		query		int		false	"Page number (1-based)"							default(1)
//	@Param			limit		query		int		false	"Items per page"								default(24)
//	@Param			sort		query		string		false	"Sort order (title_asc, title_desc, release_date_asc, release_date_desc, created_at_asc, created_at_desc, page_count_asc, page_count_desc)"
//	@Param			artist		query		[]string	false	"Filter by artist name(s); archive must have all"
//	@Param			tag			query		[]string	false	"Filter by tag name(s); archive must have all"
//	@Param			character	query		[]string	false	"Filter by character name(s); archive must have all"
//	@Param			parody		query		[]string	false	"Filter by parody name(s); archive must have all"
//	@Param			language	query		string		false	"Filter by language code (e.g. en, ja)"
//	@Param			category	query		string		false	"Filter by category"
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives [get]
func (h *ArchiveHandler) GetArchives(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())
	q := r.URL.Query()

	libraryID := q.Get("library_id")
	if libraryID == "" {
		response.BadRequest(w, "library_id is required")
		return
	}

	page := 1
	limit := 24

	if p := q.Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := q.Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	filter := database.ArchiveFilter{
		LibraryID:  libraryID,
		Artists:    q["artist"],
		Tags:       q["tag"],
		Characters: q["character"],
		Parodies:   q["parody"],
		Language:   q.Get("language"),
		Category:   q.Get("category"),
		Sort:       q.Get("sort"),
		Limit:      int64(limit),
		Offset:     int64((page - 1) * limit),
		UserID:     userID,
	}

	rows, total, err := database.ListArchives(r.Context(), h.db, filter)
	if err != nil {
		h.logger.Error("list archives failed", "error", err)
		response.InternalError(w, "failed to list archives")
		return
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	metaByID, err := database.GetBulkArchiveMetadata(r.Context(), h.db, ids)
	if err != nil {
		h.logger.Error("get bulk archive metadata failed", "error", err)
		response.InternalError(w, "failed to get archive metadata")
		return
	}

	lc := language.NewLanguageConverter()
	items := make([]ArchiveResponse, 0, len(rows))
	for _, row := range rows {
		var lang string
		if row.Language != nil {
			if name, err := lc.ToName(*row.Language); err == nil {
				lang = name
			} else {
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
		if meta, ok := metaByID[row.ID]; ok {
			resp.Artists = meta.Artists
			resp.Tags = meta.Tags
			resp.Parodies = meta.Parodies
			resp.Circles = meta.Circles
			resp.Characters = meta.Characters
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

// GetPageThumbnail godoc
//
//	@Summary		Get a page's thumbnail for an archive
//	@Description	Read-only — serves a thumbnail only if it has already been generated. Does not trigger generation; call POST .../thumbnails for that.
//	@Tags			archives
//	@Produce		image/webp
//	@Param			id		path	string	true	"Archive ID"
//	@Param			index	path	int		true	"Page index (0-based)"
//	@Success		200
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/pages/{index}/thumbnail [get]
func (h *ArchiveHandler) GetPageThumbnail(w http.ResponseWriter, r *http.Request) {
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

	path := h.processor.ThumbPath(id, index)
	if path == "" {
		response.NotFound(w, "thumbnail not ready")
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	http.ServeFile(w, r, path)
}

// GenerateThumbnails godoc
//
//	@Summary		Trigger thumbnail generation for an archive
//	@Description	Enqueues background generation of this archive's per-page thumbnails and returns immediately. Safe to call repeatedly — a job already pending/running for this archive is not duplicated. This is the only thing that triggers thumbnail generation; the GET endpoints only ever serve what already exists, so simply fetching a thumbnail URL (e.g. from curl/Postman) can't spin up generation work. Intended to be called by the frontend when a user opens an archive.
//	@Tags			archives
//	@Param			id	path	string	true	"Archive ID"
//	@Success		202	{object}	nil	"generation started"
//	@Success		204	{object}	nil	"thumbnails already ready"
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/thumbnails [post]
func (h *ArchiveHandler) GenerateThumbnails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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

	if h.processor.ThumbsReady(archive.ID, int(archive.PageCount)) {
		response.NoContent(w)
		return
	}

	if err := h.queue.EnqueueOnce(r.Context(), jobs.JobTypeThumbnail, jobs.ThumbnailPayload{
		ArchiveID: archive.ID,
		FilePath:  archive.FilePath,
	}); err != nil {
		h.logger.Error("enqueue thumbnail job failed", "id", id, "error", err)
		response.InternalError(w, "failed to trigger thumbnail generation")
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// thumbnailDoneEvent is the payload of the terminal "done" SSE event. Error
// is omitted on success and set when generation permanently failed (all
// retries exhausted) — see NewThumbnailHandler in internal/jobs.
type thumbnailDoneEvent struct {
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// StreamThumbnailEvents godoc
//
//	@Summary		Stream thumbnail generation progress
//	@Description	Server-Sent Events. Emits a "ready" event ({"index": N}) for each page thumbnail as it becomes available — including an immediate snapshot of pages already ready when the connection opens — followed by a terminal "done" event ({"done": true} on success, {"done": true, "error": "..."} if generation permanently failed after exhausting retries). Read-only: does not trigger generation, call POST .../thumbnails for that.
//	@Tags			archives
//	@Produce		text/event-stream
//	@Param			id	path	string	true	"Archive ID"
//	@Success		200
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/thumbnails/events [get]
func (h *ArchiveHandler) StreamThumbnailEvents(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.InternalError(w, "streaming unsupported")
		return
	}

	// Subscribe before reading the on-disk snapshot below so an event
	// published in the gap between them can't be missed — worst case it's
	// reported twice (once in the snapshot, once live), which is harmless.
	ch, cancel := h.broadcaster.Subscribe(archive.ID)
	defer cancel()

	response.SSEHeaders(w)

	pageCount := int(archive.PageCount)
	allReady := true
	for i := 0; i < pageCount; i++ {
		if h.processor.ThumbPath(archive.ID, i) != "" {
			response.SSEEvent(w, "ready", map[string]int{"index": i})
		} else {
			allReady = false
		}
	}
	flusher.Flush()

	if allReady {
		response.SSEEvent(w, "done", thumbnailDoneEvent{Done: true})
		flusher.Flush()
		return
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			if event.Done {
				response.SSEEvent(w, "done", thumbnailDoneEvent{Done: true, Error: event.Error})
				flusher.Flush()
				return
			}
			response.SSEEvent(w, "ready", map[string]int{"index": event.Index})
			flusher.Flush()
		}
	}
}
