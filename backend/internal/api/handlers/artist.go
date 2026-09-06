package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/language"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type ArtistHandler struct {
	queries   *sqlc.Queries
	db        *sql.DB
	logger    *slog.Logger
	processor *image.Processor
}

func NewArtistHandler(queries *sqlc.Queries, db *sql.DB, log *slog.Logger, processor *image.Processor) *ArtistHandler {
	return &ArtistHandler{
		queries:   queries,
		db:        db,
		processor: processor,
		logger:    log.With("handler", "artist"),
	}
}

type ArtistResponse struct {
	ID      int64    `json:"id"`
	Name    string   `json:"name"`
	Count   int64    `json:"count"`
	Aliases []string `json:"aliases,omitempty"`
	URLs    []string `json:"urls,omitempty"`
}

type ArtistListResponse struct {
	Items []ArtistResponse `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// GetArtists godoc
//
//	@Summary		List artists with pagination
//	@Tags			artists
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(50)
//	@Success		200	{object}	ArtistListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists [get]
func (h *ArtistHandler) GetArtists(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 50

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

	total, err := h.queries.TotalArtists(r.Context())
	if err != nil {
		h.logger.Error("count artists failed", "error", err)
		response.InternalError(w, "failed to count artists")

		return
	}

	rows, err := h.queries.GetArtistList(r.Context(), sqlc.GetArtistListParams{
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list artists failed", "error", err)
		response.InternalError(w, "failed to list artists")

		return
	}

	items := make([]ArtistResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ArtistResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
		})
	}

	response.JSON(w, http.StatusOK, ArtistListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetAllArtists godoc
//
//	@Summary		List every artist
//	@Tags			artists
//	@Produce		json
//	@Success		200	{array}		ArtistResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/all [get]
func (h *ArtistHandler) GetAllArtists(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetAllArtists(r.Context())
	if err != nil {
		h.logger.Error("get all artists failed", "error", err)
		response.InternalError(w, "failed to get artists")

		return
	}

	items := make([]ArtistResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ArtistResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
		})
	}

	response.JSON(w, http.StatusOK, items)
}

// artistDetail loads aliases + urls for artist and builds the full response.
func (h *ArtistHandler) artistDetail(ctx context.Context, artist sqlc.Artist) (ArtistResponse, error) {
	resp := ArtistResponse{
		ID:      artist.ID,
		Name:    artist.Name,
		Count:   artist.Count,
		Aliases: []string{},
		URLs:    []string{},
	}

	aliases, err := h.queries.GetArtistAliasesByID(ctx, artist.ID)
	if err != nil {
		return resp, err
	}

	for _, a := range aliases {
		resp.Aliases = append(resp.Aliases, a.Alias)
	}

	urls, err := h.queries.GetArtistUrlsByID(ctx, artist.ID)
	if err != nil {
		return resp, err
	}

	for _, u := range urls {
		resp.URLs = append(resp.URLs, u.Url)
	}

	return resp, nil
}

// GetArtist godoc
//
//	@Summary		Get a single artist
//	@Description	Includes the artist's aliases and URLs, unlike the list endpoints.
//	@Tags			artists
//	@Produce		json
//	@Param			id	path		int	true	"Artist ID"
//	@Success		200	{object}	ArtistResponse
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id} [get]
func (h *ArtistHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid artist id")
		return
	}

	artist, err := h.queries.GetArtistByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "artist not found")
			return
		}

		h.logger.Error("get artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to get artist")

		return
	}

	resp, err := h.artistDetail(r.Context(), artist)
	if err != nil {
		h.logger.Error("get artist detail failed", "id", id, "error", err)
		response.InternalError(w, "failed to get artist")

		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// addAliases inserts each non-blank alias for artistID. Returns the raw
// error (including unique constraint violations) for the caller to classify.
func (h *ArtistHandler) addAliases(ctx context.Context, q *sqlc.Queries, artistID int64, aliases []string) error {
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}

		if err := q.CreateAlias(ctx, sqlc.CreateAliasParams{Alias: alias, ArtistID: artistID}); err != nil {
			return err
		}
	}

	return nil
}

// addURLs inserts each non-blank URL for artistID. Returns the raw error
// (including unique constraint violations) for the caller to classify.
func (h *ArtistHandler) addURLs(ctx context.Context, q *sqlc.Queries, artistID int64, urls []string) error {
	for _, u := range urls {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}

		if err := q.CreateArtistUrl(ctx, sqlc.CreateArtistUrlParams{Url: u, ArtistID: artistID}); err != nil {
			return err
		}
	}

	return nil
}

type CreateArtistRequest struct {
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
	URLs    []string `json:"urls"`
}

// CreateArtist godoc
//
//	@Summary		Create an artist
//	@Tags			artists
//	@Accept			json
//	@Produce		json
//	@Param			body	body	CreateArtistRequest	true	"Artist to create"
//	@Success		201	{object}	ArtistResponse
//	@Failure		400	{object}	response.Error
//	@Failure		409	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists [post]
func (h *ArtistHandler) CreateArtist(w http.ResponseWriter, r *http.Request) {
	var body CreateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.Error("begin transaction failed", "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	artist, err := qtx.CreateArtist(r.Context(), sqlc.CreateArtistParams{
		Name:  body.Name,
		Count: 0,
	})
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			response.Conflict(w, "an artist with this name already exists")
			return
		}

		h.logger.Error("create artist failed", "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}

	if err := h.addAliases(r.Context(), qtx, artist.ID, body.Aliases); err != nil {
		if database.IsUniqueConstraintError(err) {
			response.Conflict(w, "an alias with this value is already in use")
			return
		}

		h.logger.Error("create artist aliases failed", "artist_id", artist.ID, "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}

	if err := h.addURLs(r.Context(), qtx, artist.ID, body.URLs); err != nil {
		if database.IsUniqueConstraintError(err) {
			response.Conflict(w, "a url with this value is already in use")
			return
		}

		h.logger.Error("create artist urls failed", "artist_id", artist.ID, "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}

	if err := tx.Commit(); err != nil {
		h.logger.Error("commit transaction failed", "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}

	resp, err := h.artistDetail(r.Context(), artist)
	if err != nil {
		h.logger.Error("get artist detail failed", "id", artist.ID, "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

type UpdateArtistRequest struct {
	Name *string `json:"name"`
	// Aliases/URLs: omit to leave unchanged, send [] to clear, or send a
	// full list to replace it entirely.
	Aliases []string `json:"aliases"`
	URLs    []string `json:"urls"`
}

// UpdateArtist godoc
//
//	@Summary		Update an artist
//	@Description	Partial update. Omit a field to leave it unchanged. For aliases/urls: omit to leave unchanged, send an empty array to clear, or send a full list to replace it — there's no way to add/remove a single value.
//	@Tags			artists
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int						true	"Artist ID"
//	@Param			body	body	UpdateArtistRequest	true	"Fields to update"
//	@Success		200	{object}	ArtistResponse
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		409	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id} [patch]
func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid artist id")
		return
	}

	current, err := h.queries.GetArtistByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "artist not found")
			return
		}

		h.logger.Error("get artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to get artist")

		return
	}

	var body UpdateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	name := current.Name
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
		if name == "" {
			response.BadRequest(w, "name cannot be empty")
			return
		}
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.Error("begin transaction failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}
	defer tx.Rollback()

	qtx := h.queries.WithTx(tx)

	artist, err := qtx.UpdateArtistByID(r.Context(), sqlc.UpdateArtistByIDParams{Name: name, ID: id})
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			response.Conflict(w, "an artist with this name already exists")
			return
		}

		h.logger.Error("update artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}

	if body.Aliases != nil {
		if err := qtx.RemoveArtistAliases(r.Context(), id); err != nil {
			h.logger.Error("clear artist aliases failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return
		}

		if err := h.addAliases(r.Context(), qtx, id, body.Aliases); err != nil {
			if database.IsUniqueConstraintError(err) {
				response.Conflict(w, "an alias with this value is already in use")
				return
			}

			h.logger.Error("update artist aliases failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return
		}
	}

	if body.URLs != nil {
		if err := qtx.RemoveArtistUrls(r.Context(), id); err != nil {
			h.logger.Error("clear artist urls failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return
		}

		if err := h.addURLs(r.Context(), qtx, id, body.URLs); err != nil {
			if database.IsUniqueConstraintError(err) {
				response.Conflict(w, "a url with this value is already in use")
				return
			}

			h.logger.Error("update artist urls failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return
		}
	}

	if err := tx.Commit(); err != nil {
		h.logger.Error("commit transaction failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}

	resp, err := h.artistDetail(r.Context(), artist)
	if err != nil {
		h.logger.Error("get artist detail failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// DeleteArtist godoc
//
//	@Summary		Delete an artist
//	@Description	Cascade-deletes the artist's aliases, URLs, and circle memberships, and unlinks it from every archive. Archive metadata/files are untouched.
//	@Tags			artists
//	@Param			id	path	int	true	"Artist ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id} [delete]
func (h *ArtistHandler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid artist id")
		return
	}

	if err := h.queries.DeleteArtist(r.Context(), id); err != nil {
		h.logger.Error("delete artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to delete artist")

		return
	}

	response.NoContent(w)
}

// GetArchivesByArtist godoc
//
//	@Summary		List archives for an artist
//	@Tags			artists
//	@Produce		json
//	@Param			id		path		int	true	"Artist ID"
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id}/archives [get]
func (h *ArtistHandler) GetArchivesByArtist(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid artist id")
		return
	}

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

	total, err := h.queries.TotalArchiveWithArtist(r.Context(), id)
	if err != nil {
		h.logger.Error("count archives by artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to count archives")

		return
	}

	rows, err := h.queries.GetArchivesByArtistID(r.Context(), sqlc.GetArchivesByArtistIDParams{
		ID:     id,
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list archives by artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to list archives")

		return
	}

	lc := language.NewLanguageConverter()

	items := make([]ArchiveResponse, 0, len(rows))
	for _, row := range rows {
		resp := ArchiveResponse{
			ID:          row.ID,
			Title:       row.Title,
			Summary:     row.Summary,
			Language:    lc.DisplayName(row.Language),
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
