package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/language"
	"Shoka/internal/metadata"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	pipeline    *metadata.Pipeline
}

func NewArchiveHandler(queries *sqlc.Queries, db *sql.DB, log *slog.Logger, processor *image.Processor, cache *image.Cache, queue *jobs.Queue, broadcaster *events.ThumbnailBroadcaster, pipeline *metadata.Pipeline) *ArchiveHandler {
	return &ArchiveHandler{
		queries:     queries,
		db:          db,
		processor:   processor,
		cache:       cache,
		queue:       queue,
		broadcaster: broadcaster,
		pipeline:    pipeline,
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
	IsFavorited bool `json:"is_favorited"`
	Rating      *int `json:"rating"`
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

// favoritedSet fetches, for a set of archive ids, which ones the given user
// has favorited — one query instead of one ArchiveIsFavorited call per row.
func (h *ArchiveHandler) favoritedSet(ctx context.Context, userID string, ids []string) (map[string]bool, error) {
	favoritedIDs, err := h.queries.GetFavoritedArchiveIDs(ctx, sqlc.GetFavoritedArchiveIDsParams{
		Uid: userID,
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	favorited := make(map[string]bool, len(favoritedIDs))
	for _, id := range favoritedIDs {
		favorited[id] = true
	}

	return favorited, nil
}

// ratingSet fetches, for a set of archive ids, the given user's own rating
// (1-5) for each one that has been rated — one query instead of one
// GetArchiveRating call per row.
func (h *ArchiveHandler) ratingSet(ctx context.Context, userID string, ids []string) (map[string]int, error) {
	rows, err := h.queries.GetArchiveRatingsForIDs(ctx, sqlc.GetArchiveRatingsForIDsParams{
		Uid: userID,
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}

	ratings := make(map[string]int, len(rows))
	for _, row := range rows {
		ratings[row.ArchiveID] = int(row.Rating)
	}

	return ratings, nil
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

// GetCategories godoc
//
//	@Summary		List all archive categories
//	@Description	Returns every distinct category currently in use across all libraries, sorted. Purely reflects what's actually been scanned/tagged — nothing is seeded ahead of time.
//	@Tags			archives
//	@Produce		json
//	@Success		200	{array}		string
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/categories [get]
func (h *ArchiveHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetAllCategory(r.Context())
	if err != nil {
		h.logger.Error("get categories failed", "error", err)
		response.InternalError(w, "failed to get categories")

		return
	}

	categories := make([]string, 0, len(rows))
	for _, c := range rows {
		if c != nil {
			categories = append(categories, *c)
		}
	}

	response.JSON(w, http.StatusOK, categories)
}

type LanguageResponse struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// GetLanguages godoc
//
//	@Summary		List all languages
//	@Description	Returns every distinct language code currently in use across all libraries, paired with a display name (see language.LanguageConverter; falls back to the raw code if unrecognized). Purely reflects what's actually been scanned/tagged — nothing is seeded ahead of time.
//	@Tags			archives
//	@Produce		json
//	@Success		200	{array}		LanguageResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/languages [get]
func (h *ArchiveHandler) GetLanguages(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetAllLanguage(r.Context())
	if err != nil {
		h.logger.Error("get languages failed", "error", err)
		response.InternalError(w, "failed to get languages")

		return
	}

	lc := language.NewLanguageConverter()

	languages := make([]LanguageResponse, 0, len(rows))
	for _, code := range rows {
		if code == nil {
			continue
		}

		name, err := lc.ToName(*code)
		if err != nil {
			name = *code
		}

		languages = append(languages, LanguageResponse{Code: *code, Name: name})
	}

	response.JSON(w, http.StatusOK, languages)
}

// hasPHashParam parses the has_phash filter. An absent or unparseable value
// means "don't filter", matching how the other optional list params here
// treat junk input - a bad value shouldn't silently hide half the library.
func hasPHashParam(raw string) *bool {
	if raw == "" {
		return nil
	}

	v, err := strconv.ParseBool(raw)
	if err != nil {
		return nil
	}

	return &v
}

// GetArchives godoc
//
//	@Summary		List archives with pagination, filtering, and sorting
//	@Tags			archives
//	@Produce		json
//	@Param			library_id	query		string		true	"Library ID to list archives from"
//	@Param			q			query		string		false	"Full-text search over title, summary, artists, tags, parodies, circles, characters, category (substring match, min 3 characters - shorter values are ignored)"
//	@Param			page		query		int		false	"Page number (1-based)"							default(1)
//	@Param			limit		query		int		false	"Items per page"								default(24)
//	@Param			sort		query		string		false	"Sort order (title_asc, title_desc, release_date_asc, release_date_desc, created_at_asc, created_at_desc, page_count_asc, page_count_desc, rating_asc, rating_desc)"
//	@Param			artist		query		[]string	false	"Filter by artist name(s); archive must have all"
//	@Param			tag			query		[]string	false	"Filter by tag name(s); archive must have all"
//	@Param			character	query		[]string	false	"Filter by character name(s); archive must have all"
//	@Param			parody		query		[]string	false	"Filter by parody name(s); archive must have all"
//	@Param			language	query		string		false	"Filter by language code (e.g. en, ja)"
//	@Param			category	query		string		false	"Filter by category"
//	@Param			has_phash	query		bool		false	"Filter by whether the archive has been perceptually hashed (for duplicate detection). Omit to include both."
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
		Query:      q.Get("q"),
		Artists:    q["artist"],
		Tags:       q["tag"],
		Characters: q["character"],
		Parodies:   q["parody"],
		Language:   q.Get("language"),
		Category:   q.Get("category"),
		HasPHash:   hasPHashParam(q.Get("has_phash")),
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

	favorited, err := h.favoritedSet(r.Context(), userID, ids)
	if err != nil {
		h.logger.Error("get favorited archive ids failed", "error", err)
		response.InternalError(w, "failed to get favorite status")

		return
	}

	ratings, err := h.ratingSet(r.Context(), userID, ids)
	if err != nil {
		h.logger.Error("get archive ratings failed", "error", err)
		response.InternalError(w, "failed to get rating")

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
			IsFavorited: favorited[row.ID],
		}
		if rating, ok := ratings[row.ID]; ok {
			resp.Rating = &rating
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

// GetRecentlyRead godoc
//
//	@Summary		List recently read archives across all libraries
//	@Description	Returns archives with reading progress for the current user, most recently read first. Deliberately NOT scoped to a library - this is a continue-reading feed spanning the whole collection.
//	@Tags			archives
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/recently-read [get]
func (h *ArchiveHandler) GetRecentlyRead(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.queries.CountRecentlyReadArchives(r.Context(), userID)
	if err != nil {
		h.logger.Error("count recently read archives failed", "error", err)
		response.InternalError(w, "failed to count recently read archives")

		return
	}

	rows, err := h.queries.GetRecentlyReadArchives(r.Context(), sqlc.GetRecentlyReadArchivesParams{
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("get recently read archives failed", "error", err)
		response.InternalError(w, "failed to get recently read archives")

		return
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	favorited, err := h.favoritedSet(r.Context(), userID, ids)
	if err != nil {
		h.logger.Error("get favorited archive ids failed", "error", err)
		response.InternalError(w, "failed to get favorite status")

		return
	}

	ratings, err := h.ratingSet(r.Context(), userID, ids)
	if err != nil {
		h.logger.Error("get archive ratings failed", "error", err)
		response.InternalError(w, "failed to get rating")

		return
	}

	items := make([]ArchiveResponse, 0, len(rows))
	for _, row := range rows {
		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    row.Language,
			Category:    row.Category,
			ReleaseDate: row.ReleaseDate,
			PageCount:   int(row.PageCount),
			FilePath:    row.FilePath,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			ThumbsReady: h.processor.ThumbsReady(row.ID, int(row.PageCount)),
			IsFavorited: favorited[row.ID],
			Progress: &ProgressResponse{
				CurrentPage: int(row.Page),
				LastRead:    row.LastRead,
				Completed:   row.Completed,
			},
		}
		if rating, ok := ratings[row.ID]; ok {
			resp.Rating = &rating
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

// GetFavorites godoc
//
//	@Summary		List favorited archives across all libraries
//	@Description	Returns archives the current user has favorited, most recently favorited first. Deliberately NOT scoped to a library - this is a favorites feed spanning the whole collection.
//	@Tags			archives
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/favorites [get]
func (h *ArchiveHandler) GetFavorites(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.queries.CountUserFavoriteArchive(r.Context(), userID)
	if err != nil {
		h.logger.Error("count favorite archives failed", "error", err)
		response.InternalError(w, "failed to count favorite archives")

		return
	}

	rows, err := h.queries.GetUserFavoriteArchiveList(r.Context(), sqlc.GetUserFavoriteArchiveListParams{
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("get favorite archives failed", "error", err)
		response.InternalError(w, "failed to get favorite archives")

		return
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	ratings, err := h.ratingSet(r.Context(), userID, ids)
	if err != nil {
		h.logger.Error("get archive ratings failed", "error", err)
		response.InternalError(w, "failed to get rating")

		return
	}

	items := make([]ArchiveResponse, 0, len(rows))
	for _, row := range rows {
		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    row.Language,
			Category:    row.Category,
			ReleaseDate: row.ReleaseDate,
			PageCount:   int(row.PageCount),
			FilePath:    row.FilePath,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
			ThumbsReady: h.processor.ThumbsReady(row.ID, int(row.PageCount)),
			IsFavorited: true,
		}
		if rating, ok := ratings[row.ID]; ok {
			resp.Rating = &rating
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
		meta        *metadata.Result
		progress    *sqlc.ReadingProgress
		isFavorited bool
		rating      *int
		metaErr     error
		progErr     error
		favErr      error
		ratingErr   error
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

		res.isFavorited, res.favErr = h.queries.ArchiveIsFavorited(r.Context(), sqlc.ArchiveIsFavoritedParams{
			ArchiveID: archive.ID,
			Uid:       userID,
		})

		rating, err := h.queries.GetArchiveRating(r.Context(), sqlc.GetArchiveRatingParams{
			ArchiveID: archive.ID,
			Uid:       userID,
		})
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				res.ratingErr = err
			}
		} else {
			r := int(rating)
			res.rating = &r
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

	if res.favErr != nil {
		h.logger.Error("get favorite status failed", "id", id, "error", res.favErr)
		response.InternalError(w, "failed to get favorite status")

		return
	}

	if res.ratingErr != nil {
		h.logger.Error("get rating failed", "id", id, "error", res.ratingErr)
		response.InternalError(w, "failed to get rating")

		return
	}

	response.JSON(w, http.StatusOK, h.buildResponse(archive, res.meta, res.progress, res.isFavorited, res.rating))
}

// UpdateArchiveRequest is a partial update: omitted fields are left
// unchanged. For list fields (artists/tags/parodies/circles/characters),
// omit the key to leave it unchanged, send an empty array to clear it, or
// send a full list to replace it — there's no way to add/remove a single
// value, this always replaces the whole relation.
//
// page_count is deliberately not editable here — it's a structural fact
// about the archive file (used for pagination/thumbnail bounds checks
// elsewhere), not curated metadata.
type UpdateArchiveRequest struct {
	Title       *string    `json:"title"`
	Summary     *string    `json:"summary"`
	Language    *string    `json:"language"`
	Category    *string    `json:"category"`
	ReleaseDate *time.Time `json:"release_date"`
	Artists     []string   `json:"artists"`
	Tags        []string   `json:"tags"`
	Parodies    []string   `json:"parodies"`
	Circles     []string   `json:"circles"`
	Characters  []string   `json:"characters"`
}

// UpdateArchive godoc
//
//	@Summary		Update an archive
//	@Description	Partial update of an archive's metadata, including relations (artists/tags/parodies/circles/characters). This is a direct manual edit, not a metadata source — a subsequent "fetch metadata" call can still overwrite these values, there is currently no per-field locking.
//	@Tags			archives
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Archive ID"
//	@Param			body	UpdateArchiveRequest	true	"Fields to update"
//	@Success		200	{object}	ArchiveResponse
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id} [patch]
func (h *ArchiveHandler) UpdateArchive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	if _, err := h.queries.GetArchiveByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "archive not found")
			return
		}

		h.logger.Error("get archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to get archive")

		return
	}

	var body UpdateArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if body.Title != nil && strings.TrimSpace(*body.Title) == "" {
		response.BadRequest(w, "title cannot be empty")
		return
	}

	if err := metadata.ApplyMetadata(r.Context(), h.queries, h.db, id, &metadata.Result{
		Title:       body.Title,
		Summary:     body.Summary,
		Language:    body.Language,
		Category:    body.Category,
		ReleaseDate: body.ReleaseDate,
		Artists:     body.Artists,
		Tags:        body.Tags,
		Parodies:    body.Parodies,
		Circles:     body.Circles,
		Characters:  body.Characters,
	}); err != nil {
		h.logger.Error("update archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to update archive")

		return
	}

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		h.logger.Error("get updated archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to get updated archive")

		return
	}

	meta, err := database.GetArchiveMetadata(r.Context(), h.queries, id)
	if err != nil {
		h.logger.Error("get archive metadata failed", "id", id, "error", err)
		response.InternalError(w, "failed to get updated archive")

		return
	}

	var progress *sqlc.ReadingProgress
	if p, err := h.queries.GetProgressForArchive(r.Context(), sqlc.GetProgressForArchiveParams{
		ArchiveID: id,
		UserID:    userID,
	}); err == nil {
		progress = &p
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get progress failed", "id", id, "error", err)
	}

	isFavorited, err := h.queries.ArchiveIsFavorited(r.Context(), sqlc.ArchiveIsFavoritedParams{
		ArchiveID: id,
		Uid:       userID,
	})
	if err != nil {
		h.logger.Error("get favorite status failed", "id", id, "error", err)
		response.InternalError(w, "failed to get updated archive")

		return
	}

	var rating *int

	if v, err := h.queries.GetArchiveRating(r.Context(), sqlc.GetArchiveRatingParams{
		ArchiveID: id,
		Uid:       userID,
	}); err == nil {
		iv := int(v)
		rating = &iv
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get rating failed", "id", id, "error", err)
		response.InternalError(w, "failed to get updated archive")

		return
	}

	response.JSON(w, http.StatusOK, h.buildResponse(archive, meta, progress, isFavorited, rating))
}

// DeleteArchive godoc
//
//	@Summary		Delete an archive
//	@Description	Removes the archive and all its associated data (progress, favorites, ratings, metadata) from the library. By default the underlying file on disk is left untouched — pass delete_file=true to also remove it. Cache/thumbnail data is always evicted.
//	@Tags			archives
//	@Param			id			path	string	true	"Archive ID"
//	@Param			delete_file	query	bool	false	"Also delete the archive file from disk"	default(false)
//	@Success		204
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id} [delete]
func (h *ArchiveHandler) DeleteArchive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	deleteFile := r.URL.Query().Get("delete_file") == "true"

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

	if err := h.queries.DeleteArchive(r.Context(), id); err != nil {
		h.logger.Error("delete archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to delete archive")

		return
	}

	if err := h.processor.EvictAll(id); err != nil {
		h.logger.Warn("evict archive cache failed", "id", id, "error", err)
	}

	if deleteFile {
		if err := os.Remove(archive.FilePath); err != nil && !os.IsNotExist(err) {
			h.logger.Warn("delete archive file failed", "id", id, "path", archive.FilePath, "error", err)
		}
	}

	h.logger.Info("archive deleted", "id", id, "delete_file", deleteFile)
	response.NoContent(w)
}

type UpdateProgressRequest struct {
	Page int `json:"page"`
	// Completed: omit to auto-derive (true once page reaches the archives
	// last page), or set explicitly to override that (e.g. letting a user
	// mark something finished/unfinished regardless of page position).
	Completed *bool `json:"completed"`
}

// UpdateProgress godoc
//
//	@Summary		Record reading progress for an archive
//	@Description	Upserts the current users reading progress. Idempotent - safe to call on every page turn with the new absolute page number.
//	@Tags			archives
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"Archive ID"
//	@Param			body	UpdateProgressRequest	true	"Progress"
//	@Success		200	{object}	ProgressResponse
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/progress [put]
func (h *ArchiveHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
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

	var body UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if archive.PageCount > 0 && (body.Page < 0 || body.Page >= int(archive.PageCount)) {
		response.BadRequest(w, "page out of range")
		return
	}

	completed := archive.PageCount > 0 && body.Page >= int(archive.PageCount)-1
	if body.Completed != nil {
		completed = *body.Completed
	}

	row, err := h.queries.UpsertReadingProgress(r.Context(), sqlc.UpsertReadingProgressParams{
		ArchiveID: id,
		UserID:    userID,
		Page:      int64(body.Page),
		Completed: completed,
	})
	if err != nil {
		h.logger.Error("upsert reading progress failed", "id", id, "error", err)
		response.InternalError(w, "failed to update progress")

		return
	}

	response.JSON(w, http.StatusOK, ProgressResponse{
		CurrentPage: int(row.Page),
		LastRead:    row.LastRead,
		Completed:   row.Completed,
	})
}

// DeleteProgress godoc
//
//	@Summary		Reset reading progress for an archive
//	@Description	Removes the current users reading progress for this archive entirely (as opposed to setting page back to 0, which would still count as "in progress").
//	@Tags			archives
//	@Param			id	path	string	true	"Archive ID"
//	@Success		204
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/progress [delete]
func (h *ArchiveHandler) DeleteProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	if err := h.queries.DeleteReadingProgress(r.Context(), sqlc.DeleteReadingProgressParams{
		ArchiveID: id,
		UserID:    userID,
	}); err != nil {
		h.logger.Error("delete reading progress failed", "id", id, "error", err)
		response.InternalError(w, "failed to reset progress")

		return
	}

	response.NoContent(w)
}

// AddFavorite godoc
//
//	@Summary		Favorite an archive
//	@Description	Idempotent - adds the archive to the current users favorites. No-op if already favorited.
//	@Tags			archives
//	@Param			id	path	string	true	"Archive ID"
//	@Success		204
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/favorite [put]
func (h *ArchiveHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	if _, err := h.queries.GetArchiveByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "archive not found")
			return
		}

		h.logger.Error("get archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to get archive")

		return
	}

	if err := h.queries.AddFavoriteArchive(r.Context(), sqlc.AddFavoriteArchiveParams{
		ArchiveID: id,
		UserID:    userID,
	}); err != nil {
		h.logger.Error("add favorite archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to favorite archive")

		return
	}

	response.NoContent(w)
}

// RemoveFavorite godoc
//
//	@Summary		Unfavorite an archive
//	@Description	Removes the archive from the current users favorites. No-op if it wasn't favorited.
//	@Tags			archives
//	@Param			id	path	string	true	"Archive ID"
//	@Success		204
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/favorite [delete]
func (h *ArchiveHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	if err := h.queries.RemoveFavoriteArchive(r.Context(), sqlc.RemoveFavoriteArchiveParams{
		ArchiveID: id,
		Uid:       userID,
	}); err != nil {
		h.logger.Error("remove favorite archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to unfavorite archive")

		return
	}

	response.NoContent(w)
}

type SetRatingRequest struct {
	// Rating is 1-5, matching a classic 5-star rating widget. There's no
	// 0/"unset" value here — clearing a rating is DELETE .../rating.
	Rating int `json:"rating"`
}

// SetRating godoc
//
//	@Summary		Rate an archive
//	@Description	Idempotent - sets the current users star rating (1-5) for this archive, overwriting any previous rating.
//	@Tags			archives
//	@Accept			json
//	@Param			id		path	string				true	"Archive ID"
//	@Param			body	SetRatingRequest	true	"Rating"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/rating [put]
func (h *ArchiveHandler) SetRating(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	var body SetRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if body.Rating < 1 || body.Rating > 5 {
		response.BadRequest(w, "rating must be between 1 and 5")
		return
	}

	if _, err := h.queries.GetArchiveByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "archive not found")
			return
		}

		h.logger.Error("get archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to get archive")

		return
	}

	if _, err := h.queries.UpsertArchiveRating(r.Context(), sqlc.UpsertArchiveRatingParams{
		ArchiveID: id,
		UserID:    userID,
		Rating:    int64(body.Rating),
	}); err != nil {
		h.logger.Error("upsert archive rating failed", "id", id, "error", err)
		response.InternalError(w, "failed to rate archive")

		return
	}

	response.NoContent(w)
}

// RemoveRating godoc
//
//	@Summary		Clear an archive's rating
//	@Description	Removes the current users star rating for this archive. No-op if it wasn't rated.
//	@Tags			archives
//	@Param			id	path	string	true	"Archive ID"
//	@Success		204
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/rating [delete]
func (h *ArchiveHandler) RemoveRating(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.UserIDFromContext(r.Context())

	if err := h.queries.RemoveArchiveRating(r.Context(), sqlc.RemoveArchiveRatingParams{
		ArchiveID: id,
		Uid:       userID,
	}); err != nil {
		h.logger.Error("remove archive rating failed", "id", id, "error", err)
		response.InternalError(w, "failed to clear rating")

		return
	}

	response.NoContent(w)
}

func (h *ArchiveHandler) buildResponse(
	archive sqlc.Archive,
	meta *metadata.Result,
	progress *sqlc.ReadingProgress,
	isFavorited bool,
	rating *int,
) ArchiveResponse {
	return buildArchiveResponse(h.processor, archive, meta, progress, isFavorited, rating)
}

// buildArchiveResponse is the shared ArchiveResponse builder — a
// package-level function (rather than an ArchiveHandler method) so other
// handlers that hold their own *image.Processor can build the same
// response shape, e.g. MetadataHandler previewing a fetched metadata
// result without persisting it.
func buildArchiveResponse(
	processor *image.Processor,
	archive sqlc.Archive,
	meta *metadata.Result,
	progress *sqlc.ReadingProgress,
	isFavorited bool,
	rating *int,
) ArchiveResponse {
	lc := language.NewLanguageConverter()

	var (
		lang string
		err  error
	)
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
		ThumbsReady: processor.ThumbsReady(archive.ID, int(archive.PageCount)),
		IsFavorited: isFavorited,
		Rating:      rating,
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

// overlayResult returns a copy of archive with any non-nil scalar fields
// from result overlaid on top. Used to preview what an archive would look
// like with a fetched metadata result applied, without writing anything to
// the database — the real archive row is never mutated by this.
func overlayResult(archive sqlc.Archive, result *metadata.Result) sqlc.Archive {
	preview := archive
	if result.Title != nil {
		preview.Title = *result.Title
	}

	if result.Summary != nil {
		preview.Summary = result.Summary
	}

	if result.Language != nil {
		preview.Language = result.Language
	}

	if result.Category != nil {
		preview.Category = result.Category
	}

	if result.ReleaseDate != nil {
		preview.ReleaseDate = result.ReleaseDate
	}

	if result.PageCount != nil {
		preview.PageCount = *result.PageCount
	}

	return preview
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

	for i := range pageCount {
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
