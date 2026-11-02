package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
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

// ArtistHandler serves artist CRUD and the archives linked to an artist.
type ArtistHandler struct {
	queries   *sqlc.Queries
	db        *sql.DB
	logger    *slog.Logger
	processor *image.Processor
}

// NewArtistHandler returns an ArtistHandler backed by queries, db, and the
// shared image processor.
func NewArtistHandler(queries *sqlc.Queries, db *sql.DB, log *slog.Logger, processor *image.Processor) *ArtistHandler {
	return &ArtistHandler{
		queries:   queries,
		db:        db,
		processor: processor,
		logger:    log.With("handler", "artist"),
	}
}

// ArtistCoverResponse identifies the archive whose cover stands in for the
// artist, plus the inputs of its cover URL (cover_page busts the image cache).
type ArtistCoverResponse struct {
	ArchiveID   string `json:"archive_id"`
	Page        int    `json:"page"`
	ThumbsReady bool   `json:"thumbs_ready"`
}

// ArtistResponse is the artist payload returned by the list endpoints and
// CreateArtist/UpdateArtist.
type ArtistResponse struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Count       int64                  `json:"count"`
	Aliases     []string               `json:"aliases,omitempty"`
	URLs        []string               `json:"urls,omitempty"`
	Circles     []CircleArtistResponse `json:"circles,omitempty"`
	Cover       *ArtistCoverResponse   `json:"cover,omitempty"`
	IsFavorited bool                   `json:"is_favorited"`
	Rating      *int                   `json:"rating"`
}

// ArtistListResponse is a single page of artists plus the total count.
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
//	@Param			page		query		int		false	"Page number (1-based)"	default(1)
//	@Param			limit		query		int		false	"Items per page"		default(50)
//	@Param			favorites	query		bool	false	"Only artists the current user has favorited"
//	@Param			rating		query		int		false	"Only artists the current user rated this many stars (1-5)"
//	@Success		200	{object}	ArtistListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists [get]
func (h *ArtistHandler) GetArtists(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r, defaultListLimit)
	offset := int64((page - 1) * limit)

	favoritesOnly := int64(0)
	if r.URL.Query().Get("favorites") == "true" {
		favoritesOnly = 1
	}

	// Out-of-range or unparseable values are ignored rather than rejected:
	// filtering is a view concern and a stale bookmark shouldn't 400 a page.
	filterRating := ratingFilterParam(r.URL.Query().Get("rating"))

	uid := auth.UserIDFromContext(r.Context())
	allLibraries := allLibrariesFlag(r.Context())

	total, err := h.queries.TotalArtists(r.Context(), sqlc.TotalArtistsParams{
		Uid:           uid,
		FavoritesOnly: favoritesOnly,
		FilterRating:  filterRating,
	})
	if err != nil {
		h.logger.Error("count artists failed", "error", err)
		response.InternalError(w, "failed to count artists")

		return
	}

	rows, err := h.queries.GetArtistList(r.Context(), sqlc.GetArtistListParams{
		CoverAllLibs:  allLibraries,
		Uid:           uid,
		FavoritesOnly: favoritesOnly,
		FilterRating:  filterRating,
		Limit:         int64(limit),
		Offset:        offset,
	})
	if err != nil {
		h.logger.Error("list artists failed", "error", err)
		response.InternalError(w, "failed to list artists")

		return
	}

	items := make([]ArtistResponse, 0, len(rows))

	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
		items = append(items, ArtistResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
			Cover: h.coverFromRow(row),
		})
	}

	// Both enrichments are one query per page rather than one per artist.
	// sqlc renders an empty slice as `in (NULL)`, which matches nothing.
	favorited, err := h.queries.GetFavoritedArtistIDs(r.Context(), sqlc.GetFavoritedArtistIDsParams{
		Uid: uid,
		Ids: ids,
	})
	if err != nil {
		h.logger.Error("get favorited artists failed", "error", err)
		response.InternalError(w, "failed to list artists")

		return
	}

	favoritedSet := make(map[int64]struct{}, len(favorited))
	for _, id := range favorited {
		favoritedSet[id] = struct{}{}
	}

	for i := range items {
		_, items[i].IsFavorited = favoritedSet[items[i].ID]
	}

	ratings, err := h.queries.GetArtistRatingsForIDs(r.Context(), sqlc.GetArtistRatingsForIDsParams{
		Uid: uid,
		Ids: ids,
	})
	if err != nil {
		h.logger.Error("get artist ratings failed", "error", err)
		response.InternalError(w, "failed to list artists")

		return
	}

	ratingByArtist := make(map[int64]int, len(ratings))

	for _, rating := range ratings {
		ratingByArtist[rating.ArtistID] = int(rating.Rating)
	}

	for i := range items {
		v, ok := ratingByArtist[items[i].ID]
		if !ok {
			continue
		}

		// Copy before taking the address so entries don't share one.
		items[i].Rating = &v
	}

	response.JSON(w, http.StatusOK, ArtistListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// allLibrariesFlag renders the request's library scope as the 1/0 flag the
// cover queries compare against.
func allLibrariesFlag(ctx context.Context) int64 {
	if auth.LibraryScopeFromContext(ctx).Unrestricted() {
		return 1
	}

	return 0
}

// coverFromRow builds the cover payload from a list row's joined columns.
func (h *ArtistHandler) coverFromRow(row sqlc.GetArtistListRow) *ArtistCoverResponse {
	if row.CoverArchiveID == nil || row.CoverPage == nil || row.CoverPageCount == nil {
		return nil
	}

	return &ArtistCoverResponse{
		ArchiveID:   *row.CoverArchiveID,
		Page:        int(*row.CoverPage),
		ThumbsReady: h.processor.ThumbsReady(*row.CoverArchiveID, int(*row.CoverPageCount)),
	}
}

// artistCover resolves the artist's current cover within this user's library
// access; no cover is not an error.
func (h *ArtistHandler) artistCover(ctx context.Context, id int64, uid string) (*ArtistCoverResponse, error) {
	row, err := h.queries.GetArtistCoverArchive(ctx, sqlc.GetArtistCoverArchiveParams{
		CoverAllLibs: allLibrariesFlag(ctx),
		Uid:          uid,
		ArtistID:     id,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	if row.ArchiveID == nil || row.CoverPage == nil || row.PageCount == nil {
		return nil, nil
	}

	return &ArtistCoverResponse{
		ArchiveID:   *row.ArchiveID,
		Page:        int(*row.CoverPage),
		ThumbsReady: h.processor.ThumbsReady(*row.ArchiveID, int(*row.PageCount)),
	}, nil
}

// GetAllArtists godoc
//
//	@Summary		List every artist
//	@Tags			artists
//	@Produce		json
//	@Success		200	{array}		ArtistResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/all [get]
//
// Cover, IsFavorited, and Rating are deliberately left zero: every consumer
// of this endpoint is a name picker, and computing them per artist would cost
// the same enrichment as GetArtists for a payload nothing reads.
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

// artistDetail loads aliases, urls, circles, cover, and this user's
// favorite/rating state for artist and builds the full response.
func (h *ArtistHandler) artistDetail(ctx context.Context, artist sqlc.Artist, uid string) (ArtistResponse, error) {
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

	circles, err := h.queries.GetArtistCircles(ctx, artist.ID)
	if err != nil {
		return resp, err
	}

	for _, c := range circles {
		resp.Circles = append(resp.Circles, CircleArtistResponse{ID: c.ID, Name: c.Name})
	}

	cover, err := h.artistCover(ctx, artist.ID, uid)
	if err != nil {
		return resp, err
	}

	resp.Cover = cover

	isFavorited, err := h.queries.ArtistIsFavorited(ctx, sqlc.ArtistIsFavoritedParams{
		ArtistID: artist.ID,
		Uid:      uid,
	})
	if err != nil {
		return resp, err
	}

	resp.IsFavorited = isFavorited

	rating, err := h.queries.GetArtistRating(ctx, sqlc.GetArtistRatingParams{
		ArtistID: artist.ID,
		Uid:      uid,
	})
	if err == nil {
		v := int(rating)
		resp.Rating = &v
	} else if !errors.Is(err, sql.ErrNoRows) {
		return resp, err
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

	resp, err := h.artistDetail(r.Context(), artist, auth.UserIDFromContext(r.Context()))
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

// CreateArtistRequest is the JSON body accepted by CreateArtist.
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
//	@Failure		403	{object}	response.Error
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
	defer func() { _ = tx.Rollback() }()

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

	resp, err := h.artistDetail(r.Context(), artist, auth.UserIDFromContext(r.Context()))
	if err != nil {
		h.logger.Error("get artist detail failed", "id", artist.ID, "error", err)
		response.InternalError(w, "failed to create artist")

		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

// UpdateArtistRequest is the JSON body accepted by UpdateArtist.
type UpdateArtistRequest struct {
	Name *string `json:"name"`
	// Aliases/URLs: omit to leave unchanged, send [] to clear, or send a
	// full list to replace it entirely.
	Aliases []string `json:"aliases"`
	URLs    []string `json:"urls"`
	// CoverArchiveID picks which of the artist's archives supplies the
	// thumbnail. Omitted leaves it unchanged; "" restores the default
	// (the artist's oldest archive).
	CoverArchiveID *string `json:"cover_archive_id"`
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
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		409	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id} [patch]
func (h *ArtistHandler) UpdateArtist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "artist")
	if !ok {
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

	name, body, ok := updateArtistBody(w, r, current.Name)
	if !ok {
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.Error("begin transaction failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}
	defer func() { _ = tx.Rollback() }()

	artist, ok := h.applyArtistUpdate(w, r, h.queries.WithTx(tx), id, name, body)
	if !ok {
		return
	}

	if err := tx.Commit(); err != nil {
		h.logger.Error("commit transaction failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}

	resp, err := h.artistDetail(r.Context(), artist, auth.UserIDFromContext(r.Context()))
	if err != nil {
		h.logger.Error("get artist detail failed", "id", artist.ID, "error", err)
		response.InternalError(w, "failed to update artist")

		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// updateArtistBody decodes the PATCH body and resolves the artist's new name,
// keeping current when the body omits it. It writes the 400 response and
// returns false on a bad request.
func updateArtistBody(w http.ResponseWriter, r *http.Request, current string) (string, UpdateArtistRequest, bool) {
	var body UpdateArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")

		return "", body, false
	}

	name := current
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
		if name == "" {
			response.BadRequest(w, "name cannot be empty")

			return "", body, false
		}
	}

	return name, body, true
}

// applyArtistUpdate renames the artist and reconciles the aliases and urls
// the request asked to replace, all inside qtx. It writes the failure
// response and returns false on error.
func (h *ArtistHandler) applyArtistUpdate(w http.ResponseWriter, r *http.Request, qtx *sqlc.Queries, id int64, name string, body UpdateArtistRequest) (sqlc.Artist, bool) {
	artist, err := qtx.UpdateArtistByID(r.Context(), sqlc.UpdateArtistByIDParams{Name: name, ID: id})
	if err != nil {
		if database.IsUniqueConstraintError(err) {
			response.Conflict(w, "an artist with this name already exists")

			return sqlc.Artist{}, false
		}

		h.logger.Error("update artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return sqlc.Artist{}, false
	}

	// Aliases and urls are replaced the same way, so they share one loop.
	links := []struct {
		plural   string
		values   []string
		clear    func(context.Context, int64) error
		insert   func(context.Context, *sqlc.Queries, int64, []string) error
		conflict string
	}{
		{"aliases", body.Aliases, qtx.RemoveArtistAliases, h.addAliases, "an alias with this value is already in use"},
		{"urls", body.URLs, qtx.RemoveArtistUrls, h.addURLs, "a url with this value is already in use"},
	}

	for _, link := range links {
		if link.values == nil {
			continue
		}

		if err := link.clear(r.Context(), id); err != nil {
			h.logger.Error("clear artist "+link.plural+" failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return sqlc.Artist{}, false
		}

		if err := link.insert(r.Context(), qtx, id, link.values); err != nil {
			if database.IsUniqueConstraintError(err) {
				response.Conflict(w, link.conflict)

				return sqlc.Artist{}, false
			}

			h.logger.Error("update artist "+link.plural+" failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return sqlc.Artist{}, false
		}
	}

	if !h.applyArtistCover(w, r, qtx, id, body.CoverArchiveID) {
		return sqlc.Artist{}, false
	}

	return artist, true
}

// applyArtistCover stores the chosen cover archive, or clears it back to the
// default when raw is "". It writes the failure response and returns false.
func (h *ArtistHandler) applyArtistCover(w http.ResponseWriter, r *http.Request, qtx *sqlc.Queries, id int64, raw *string) bool {
	if raw == nil {
		return true
	}

	cover := strings.TrimSpace(*raw)
	if cover != "" {
		linked, err := qtx.ArchiveArtistLinkExists(r.Context(), sqlc.ArchiveArtistLinkExistsParams{
			ArchiveID: cover,
			ArtistID:  id,
		})
		if err != nil {
			h.logger.Error("check artist cover link failed", "id", id, "error", err)
			response.InternalError(w, "failed to update artist")

			return false
		}

		if !linked {
			response.BadRequest(w, "archive does not belong to this artist")

			return false
		}
	}

	var coverID *string
	if cover != "" {
		coverID = &cover
	}

	if err := qtx.UpdateArtistCover(r.Context(), sqlc.UpdateArtistCoverParams{
		CoverArchiveID: coverID,
		ID:             id,
	}); err != nil {
		h.logger.Error("update artist cover failed", "id", id, "error", err)
		response.InternalError(w, "failed to update artist")

		return false
	}

	return true
}

// DeleteArtist godoc
//
//	@Summary		Delete an artist
//	@Description	Cascade-deletes the artist's aliases, URLs, and circle memberships, and unlinks it from every archive. Archive metadata/files are untouched.
//	@Tags			artists
//	@Param			id	path	int	true	"Artist ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id} [delete]
func (h *ArtistHandler) DeleteArtist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "artist")
	if !ok {
		return
	}

	if err := h.queries.DeleteArtist(r.Context(), id); err != nil {
		h.logger.Error("delete artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to delete artist")

		return
	}

	response.NoContent(w)
}

// MergeArtistRequest is the JSON body accepted by MergeArtist.
type MergeArtistRequest struct {
	TargetID int64 `json:"target_id"`
}

// MergeArtist godoc
//
//	@Summary		Merge an artist into another
//	@Description	Moves the artist's archives, audio works, URLs, circles, favorites and ratings to the target; adds its name and aliases to the target; deletes the artist. An alias that already belongs to a different artist fails the merge with 409.
//	@Tags			artists
//	@Accept			json
//	@Param			id		path	int					true	"Artist ID"
//	@Param			body	body	MergeArtistRequest	true	"Target artist"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		403	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		409	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id}/merge [post]
func (h *ArtistHandler) MergeArtist(w http.ResponseWriter, r *http.Request) {
	sourceID, targetID, ok := mergeTarget(w, r)
	if !ok {
		return
	}

	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.Error("begin transaction failed", "source_id", sourceID, "target_id", targetID, "error", err)
		response.InternalError(w, "failed to merge artists")

		return
	}
	defer func() { _ = tx.Rollback() }()

	if !h.mergeArtists(r.Context(), w, h.queries.WithTx(tx), sourceID, targetID) {
		return
	}

	if err := tx.Commit(); err != nil {
		h.logger.Error("commit transaction failed", "source_id", sourceID, "target_id", targetID, "error", err)
		response.InternalError(w, "failed to merge artists")

		return
	}

	response.NoContent(w)
}

// mergeArtists performs the merge inside q, writing the failure response and
// returning false when the caller must roll back.
func (h *ArtistHandler) mergeArtists(ctx context.Context, w http.ResponseWriter, q *sqlc.Queries, sourceID, targetID int64) bool {
	source, target, ok := h.mergeArtistRows(ctx, w, q, sourceID, targetID)
	if !ok {
		return false
	}

	aliases, conflict, err := h.mergeArtistAliases(ctx, q, source, target)
	if err != nil {
		h.logger.Error("plan artist aliases failed", "source_id", sourceID, "target_id", targetID, "error", err)
		response.InternalError(w, "failed to merge artists")

		return false
	}

	if conflict != "" {
		response.Conflict(w, `alias "`+conflict+`" is already used by another artist`)

		return false
	}

	if !h.moveArtistData(ctx, w, q, sourceID, targetID) {
		return false
	}

	if !h.addMergedAliases(ctx, w, q, targetID, aliases) {
		return false
	}

	if err := q.RecomputeArtistCount(ctx, targetID); err != nil {
		h.logger.Error("recompute artist count failed", "target_id", targetID, "error", err)
		response.InternalError(w, "failed to merge artists")

		return false
	}

	return true
}

// moveArtistData copies the source's relations onto the target and then
// deletes the source row. It writes the failure response and returns false
// on the first failure.
func (h *ArtistHandler) moveArtistData(ctx context.Context, w http.ResponseWriter, q *sqlc.Queries, sourceID, targetID int64) bool {
	params := sqlc.RepointArtistArchiveLinksParams{TargetID: targetID, SourceID: sourceID}

	steps := []struct {
		name string
		run  func() error
	}{
		{"repoint archive links", func() error { return q.RepointArtistArchiveLinks(ctx, params) }},
		{"repoint audio links", func() error { return q.RepointAudioArtistLinks(ctx, sqlc.RepointAudioArtistLinksParams(params)) }},
		{"move urls", func() error { return q.MoveArtistUrls(ctx, sqlc.MoveArtistUrlsParams(params)) }},
		{"copy circles", func() error { return q.CopyArtistCircles(ctx, sqlc.CopyArtistCirclesParams(params)) }},
		{"copy favorites", func() error { return q.CopyArtistFavorites(ctx, sqlc.CopyArtistFavoritesParams(params)) }},
		{"copy ratings", func() error { return q.CopyArtistRatings(ctx, sqlc.CopyArtistRatingsParams(params)) }},
		{"adopt cover", func() error {
			return q.AdoptArtistCover(ctx, sqlc.AdoptArtistCoverParams{SourceID: sourceID, TargetID: targetID})
		}},
		{"delete source", func() error { return q.DeleteArtist(ctx, sourceID) }},
	}

	for _, step := range steps {
		if !h.mergeStep(w, step.name, sourceID, targetID, step.run()) {
			return false
		}
	}

	return true
}

// mergeTarget parses the source id and the request body, writing the 400
// response and returning false on a bad request.
func mergeTarget(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	sourceID, ok := pathID(w, r, "artist")
	if !ok {
		return 0, 0, false
	}

	var body MergeArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")

		return 0, 0, false
	}

	if body.TargetID < 1 {
		response.BadRequest(w, "invalid target artist id")

		return 0, 0, false
	}

	if body.TargetID == sourceID {
		response.BadRequest(w, "cannot merge an artist into itself")

		return 0, 0, false
	}

	return sourceID, body.TargetID, true
}

// mergeArtistRows loads the source and target rows, writing the 404/500
// response and returning false when either is missing.
func (h *ArtistHandler) mergeArtistRows(ctx context.Context, w http.ResponseWriter, q *sqlc.Queries, sourceID, targetID int64) (sqlc.Artist, sqlc.Artist, bool) {
	source, ok := h.mergeArtistRow(ctx, w, q, sourceID, "artist not found")
	if !ok {
		return sqlc.Artist{}, sqlc.Artist{}, false
	}

	target, ok := h.mergeArtistRow(ctx, w, q, targetID, "target artist not found")
	if !ok {
		return sqlc.Artist{}, sqlc.Artist{}, false
	}

	return source, target, true
}

// mergeArtistRow loads one artist, writing the response on failure.
func (h *ArtistHandler) mergeArtistRow(ctx context.Context, w http.ResponseWriter, q *sqlc.Queries, id int64, notFound string) (sqlc.Artist, bool) {
	artist, err := q.GetArtistByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, notFound)

			return sqlc.Artist{}, false
		}

		h.logger.Error("get artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to merge artists")

		return sqlc.Artist{}, false
	}

	return artist, true
}

// mergeArtistAliases returns the trimmed, case-insensitively deduped
// aliases that should be added to target. conflict is non-empty when an
// alias already belongs to an artist other than source or target.
func (h *ArtistHandler) mergeArtistAliases(ctx context.Context, q *sqlc.Queries, source, target sqlc.Artist) (aliases []string, conflict string, err error) {
	sourceAliases, err := q.GetArtistAliasesByID(ctx, source.ID)
	if err != nil {
		return nil, "", err
	}

	targetAliases, err := q.GetArtistAliasesByID(ctx, target.ID)
	if err != nil {
		return nil, "", err
	}

	ownership, err := q.GetArtistAliasOwnership(ctx)
	if err != nil {
		return nil, "", err
	}

	// Fold the global alias ownership with Go's Unicode-aware lower-casing,
	// the same fold the resolver uses; SQLite's lower() is ASCII-only.
	owners := make(map[string][]int64, len(ownership))

	for _, entry := range ownership {
		key := strings.ToLower(strings.TrimSpace(entry.Alias))
		owners[key] = append(owners[key], entry.ArtistID)
	}

	taken := make(map[string]struct{}, len(targetAliases)+1)
	taken[strings.ToLower(strings.TrimSpace(target.Name))] = struct{}{}

	for _, alias := range targetAliases {
		taken[strings.ToLower(strings.TrimSpace(alias.Alias))] = struct{}{}
	}

	seen := make(map[string]struct{}, len(sourceAliases)+1)

	for _, candidate := range append([]string{source.Name}, aliasValues(sourceAliases)...) {
		alias := strings.TrimSpace(candidate)
		key := strings.ToLower(alias)

		if key == "" {
			continue
		}

		if _, ok := taken[key]; ok {
			continue
		}

		if _, ok := seen[key]; ok {
			continue
		}

		if otherArtistOwns(owners[key], source.ID, target.ID) {
			return nil, alias, nil
		}

		seen[key] = struct{}{}

		aliases = append(aliases, alias)
	}

	return aliases, "", nil
}

// aliasValues extracts the alias strings from GetArtistAliasesByID rows.
func aliasValues(rows []sqlc.GetArtistAliasesByIDRow) []string {
	aliases := make([]string, len(rows))
	for i, row := range rows {
		aliases[i] = row.Alias
	}

	return aliases
}

// otherArtistOwns reports whether any owner is neither the source nor the
// target artist.
func otherArtistOwns(owners []int64, sourceID, targetID int64) bool {
	for _, owner := range owners {
		if owner != sourceID && owner != targetID {
			return true
		}
	}

	return false
}

// addMergedAliases inserts the planned aliases for the target artist,
// writing the 409/500 response and returning false on failure.
func (h *ArtistHandler) addMergedAliases(ctx context.Context, w http.ResponseWriter, q *sqlc.Queries, targetID int64, aliases []string) bool {
	for _, alias := range aliases {
		if err := q.CreateAlias(ctx, sqlc.CreateAliasParams{Alias: alias, ArtistID: targetID}); err != nil {
			if database.IsUniqueConstraintError(err) {
				response.Conflict(w, `alias "`+alias+`" is already in use`)

				return false
			}

			h.logger.Error("create merged alias failed", "target_id", targetID, "error", err)
			response.InternalError(w, "failed to merge artists")

			return false
		}
	}

	return true
}

// mergeStep logs a failed relation move, writes the 500 response, and
// reports whether the merge may continue.
func (h *ArtistHandler) mergeStep(w http.ResponseWriter, name string, sourceID, targetID int64, err error) bool {
	if err == nil {
		return true
	}

	h.logger.Error("merge artist: "+name+" failed", "source_id", sourceID, "target_id", targetID, "error", err)
	response.InternalError(w, "failed to merge artists")

	return false
}

// existingArtist resolves the path id and writes the 400/404 response,
// returning false when the caller should stop.
func (h *ArtistHandler) existingArtist(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := pathID(w, r, "artist")
	if !ok {
		return 0, false
	}

	if _, err := h.queries.GetArtistByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "artist not found")

			return 0, false
		}

		h.logger.Error("get artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to get artist")

		return 0, false
	}

	return id, true
}

// AddFavorite godoc
//
//	@Summary		Favorite an artist
//	@Tags			artists
//	@Param			id	path	int	true	"Artist ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id}/favorite [put]
func (h *ArtistHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := h.existingArtist(w, r)
	if !ok {
		return
	}

	if err := h.queries.AddFavoriteArtist(r.Context(), sqlc.AddFavoriteArtistParams{
		ArtistID: id,
		UserID:   auth.UserIDFromContext(r.Context()),
	}); err != nil {
		h.logger.Error("add favorite artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to favorite artist")

		return
	}

	response.NoContent(w)
}

// RemoveFavorite godoc
//
//	@Summary		Unfavorite an artist
//	@Description	Removes the artist from the current user's favorites. No-op if it wasn't favorited.
//	@Tags			artists
//	@Param			id	path	int	true	"Artist ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id}/favorite [delete]
func (h *ArtistHandler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	id, ok := h.existingArtist(w, r)
	if !ok {
		return
	}

	if err := h.queries.RemoveFavoriteArtist(r.Context(), sqlc.RemoveFavoriteArtistParams{
		ArtistID: id,
		Uid:      auth.UserIDFromContext(r.Context()),
	}); err != nil {
		h.logger.Error("remove favorite artist failed", "id", id, "error", err)
		response.InternalError(w, "failed to unfavorite artist")

		return
	}

	response.NoContent(w)
}

// SetRating godoc
//
//	@Summary		Rate an artist
//	@Description	Idempotent - sets the current user's star rating (1-5) for this artist, overwriting any previous rating.
//	@Tags			artists
//	@Accept			json
//	@Param			id		path	int					true	"Artist ID"
//	@Param			body	body	SetRatingRequest	true	"Rating"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id}/rating [put]
func (h *ArtistHandler) SetRating(w http.ResponseWriter, r *http.Request) {
	var body SetRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if body.Rating < 1 || body.Rating > 5 {
		response.BadRequest(w, "rating must be between 1 and 5")
		return
	}

	id, ok := h.existingArtist(w, r)
	if !ok {
		return
	}

	if _, err := h.queries.UpsertArtistRating(r.Context(), sqlc.UpsertArtistRatingParams{
		ArtistID: id,
		UserID:   auth.UserIDFromContext(r.Context()),
		Rating:   int64(body.Rating),
	}); err != nil {
		h.logger.Error("upsert artist rating failed", "id", id, "error", err)
		response.InternalError(w, "failed to rate artist")

		return
	}

	response.NoContent(w)
}

// RemoveRating godoc
//
//	@Summary		Clear an artist's rating
//	@Description	Removes the current user's star rating for this artist. No-op if it wasn't rated.
//	@Tags			artists
//	@Param			id	path	int	true	"Artist ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/artists/{id}/rating [delete]
func (h *ArtistHandler) RemoveRating(w http.ResponseWriter, r *http.Request) {
	id, ok := h.existingArtist(w, r)
	if !ok {
		return
	}

	if err := h.queries.RemoveArtistRating(r.Context(), sqlc.RemoveArtistRatingParams{
		ArtistID: id,
		Uid:      auth.UserIDFromContext(r.Context()),
	}); err != nil {
		h.logger.Error("remove artist rating failed", "id", id, "error", err)
		response.InternalError(w, "failed to clear rating")

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
	id, ok := pathID(w, r, "artist")
	if !ok {
		return
	}

	archivesByID(w, r, h.logger, id, h.queries.TotalArchiveWithArtist, h.queries.GetArchivesByArtistID, h.processor)
}
