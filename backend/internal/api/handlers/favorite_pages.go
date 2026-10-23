package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// FavoritePageItem is one favorited page in the cross-library feed.
type FavoritePageItem struct {
	ArchiveID   string `json:"archive_id"`
	Title       string `json:"title"`
	PageCount   int    `json:"page_count"`
	PageIndex   int    `json:"page_index"`
	FavoritedAt string `json:"favorited_at"`
}

// FavoritePageListResponse is the paginated favorite-pages feed.
type FavoritePageListResponse struct {
	Items []FavoritePageItem `json:"items"`
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Limit int                `json:"limit"`
}

// FavoritePageIndicesResponse lists the current user's favorited page indices
// for one archive, ascending.
type FavoritePageIndicesResponse struct {
	Indices []int32 `json:"indices"`
}

// favoriteArchive validates the archive id and loads the archive, writing the
// error response itself when that fails.
func (h *ArchiveHandler) favoriteArchive(w http.ResponseWriter, r *http.Request) (sqlc.Archive, bool) {
	id := chi.URLParam(r, "id")
	if !validArchiveID(id) {
		response.BadRequest(w, "invalid archive id")

		return sqlc.Archive{}, false
	}

	archive, err := h.queries.GetArchiveByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "archive not found")
			return sqlc.Archive{}, false
		}

		h.logger.Error("get archive failed", "id", id, "error", err)
		response.InternalError(w, "failed to get archive")

		return sqlc.Archive{}, false
	}

	return archive, true
}

// favoritePageTarget validates the archive and page index path parameters,
// without checking the index against the archive's page count - callers that
// need that (the add path) do it themselves so removing a stale favorite stays
// possible.
func (h *ArchiveHandler) favoritePageTarget(w http.ResponseWriter, r *http.Request) (sqlc.Archive, int, bool) {
	index, err := strconv.Atoi(chi.URLParam(r, "index"))
	if err != nil || index < 0 {
		response.BadRequest(w, "invalid page index")

		return sqlc.Archive{}, 0, false
	}

	archive, ok := h.favoriteArchive(w, r)
	if !ok {
		return sqlc.Archive{}, 0, false
	}

	return archive, index, true
}

// AddFavoritePage godoc
//
//	@Summary		Favorite one page of an archive
//	@Description	Idempotent. Adding the same page twice leaves one row.
//	@Tags			archives
//	@Param			id		path	string	true	"Archive ID"
//	@Param			index	path	int		true	"Page index (0-based)"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/pages/{index}/favorite [put]
func (h *ArchiveHandler) AddFavoritePage(w http.ResponseWriter, r *http.Request) {
	archive, index, ok := h.favoritePageTarget(w, r)
	if !ok {
		return
	}

	if index >= int(archive.PageCount) {
		response.NotFound(w, "page not found")
		return
	}

	if err := h.queries.AddFavoritePage(r.Context(), sqlc.AddFavoritePageParams{
		ArchiveID: archive.ID,
		Uid:       auth.UserIDFromContext(r.Context()),
		PageIndex: int64(index),
	}); err != nil {
		h.logger.Error("add favorite page failed", "id", archive.ID, "index", index, "error", err)
		response.InternalError(w, "failed to favorite page")

		return
	}

	response.NoContent(w)
}

// RemoveFavoritePage godoc
//
//	@Summary		Unfavorite one page of an archive
//	@Description	Idempotent no-op if the page wasn't favorited. Works for a page index the archive no longer has, so stale favorites can still be cleared.
//	@Tags			archives
//	@Param			id		path	string	true	"Archive ID"
//	@Param			index	path	int		true	"Page index (0-based)"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/pages/{index}/favorite [delete]
func (h *ArchiveHandler) RemoveFavoritePage(w http.ResponseWriter, r *http.Request) {
	archive, index, ok := h.favoritePageTarget(w, r)
	if !ok {
		return
	}

	if err := h.queries.RemoveFavoritePage(r.Context(), sqlc.RemoveFavoritePageParams{
		ArchiveID: archive.ID,
		Uid:       auth.UserIDFromContext(r.Context()),
		PageIndex: int64(index),
	}); err != nil {
		h.logger.Error("remove favorite page failed", "id", archive.ID, "index", index, "error", err)
		response.InternalError(w, "failed to unfavorite page")

		return
	}

	response.NoContent(w)
}

// GetFavoritePageIndices godoc
//
//	@Summary		List the current user's favorited page indices for an archive
//	@Tags			archives
//	@Produce		json
//	@Param			id	path		string	true	"Archive ID"
//	@Success		200	{object}	FavoritePageIndicesResponse
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/{id}/favorite-pages [get]
func (h *ArchiveHandler) GetFavoritePageIndices(w http.ResponseWriter, r *http.Request) {
	archive, ok := h.favoriteArchive(w, r)
	if !ok {
		return
	}

	rows, err := h.queries.GetFavoritePageIndices(r.Context(), sqlc.GetFavoritePageIndicesParams{
		ArchiveID: archive.ID,
		Uid:       auth.UserIDFromContext(r.Context()),
	})
	if err != nil {
		h.logger.Error("get favorite page indices failed", "id", archive.ID, "error", err)
		response.InternalError(w, "failed to get favorite pages")

		return
	}

	indices := make([]int32, len(rows))
	for i, index := range rows {
		//nolint:gosec // page_index is CHECK (page_index >= 0) and bounded by every archive's page count
		indices[i] = int32(index)
	}

	response.JSON(w, http.StatusOK, FavoritePageIndicesResponse{Indices: indices})
}

// GetFavoritePages godoc
//
//	@Summary		List the current user's favorited pages across all archives
//	@Description	Newest first. Pages an archive has since shrunk past are omitted.
//	@Tags			archives
//	@Produce		json
//	@Param			page	query		int	false	"Page number (default 1)"
//	@Param			limit	query		int	false	"Page size (default 24, max 100)"
//	@Success		200		{object}	FavoritePageListResponse
//	@Failure		500		{object}	response.Error
//	@Router			/api/archives/favorites/pages [get]
func (h *ArchiveHandler) GetFavoritePages(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	page, limit := parsePagination(r, defaultArchiveLimit)
	offset := int64((page - 1) * limit)

	total, err := h.queries.CountUserFavoritePages(r.Context(), userID)
	if err != nil {
		h.logger.Error("count favorite pages failed", "error", err)
		response.InternalError(w, "failed to count favorite pages")

		return
	}

	rows, err := h.queries.GetUserFavoritePageList(r.Context(), sqlc.GetUserFavoritePageListParams{
		Uid:    userID,
		Offset: offset,
		Limit:  int64(limit),
	})
	if err != nil {
		h.logger.Error("get favorite pages failed", "error", err)
		response.InternalError(w, "failed to get favorite pages")

		return
	}

	items := make([]FavoritePageItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, FavoritePageItem{
			ArchiveID:   row.ArchiveID,
			Title:       row.Title,
			PageCount:   int(row.PageCount),
			PageIndex:   int(row.PageIndex),
			FavoritedAt: row.FavoritedAt.Format(time.RFC3339),
		})
	}

	response.JSON(w, http.StatusOK, FavoritePageListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
