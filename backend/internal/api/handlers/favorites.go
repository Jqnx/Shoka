package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"log/slog"
	"net/http"
	"time"
)

// FavoriteHandler serves the cross-media favorites feed.
type FavoriteHandler struct {
	queries *sqlc.Queries
	logger  *slog.Logger
}

// NewFavoriteHandler returns a FavoriteHandler backed by queries.
func NewFavoriteHandler(queries *sqlc.Queries, log *slog.Logger) *FavoriteHandler {
	return &FavoriteHandler{
		queries: queries,
		logger:  log.With("handler", "favorites"),
	}
}

// FavoriteImageItem is one favorited archive page or audio gallery image in
// the cross-library feed. Kind is "archive" or "audio"; the id field for the
// other kind is empty, and PageIndex is that work's 0-based page/image index.
type FavoriteImageItem struct {
	Kind        string `json:"kind"`
	ArchiveID   string `json:"archive_id,omitempty"`
	AudioID     string `json:"audio_id,omitempty"`
	Title       string `json:"title"`
	PageCount   int    `json:"page_count"`
	PageIndex   int    `json:"page_index"`
	FavoritedAt string `json:"favorited_at"`
	// UpdatedAt is the source work's timestamp; audio previews use it to bust
	// the browser cache when a rescan replaces the image at the same index.
	UpdatedAt string `json:"updated_at"`
}

// FavoriteImageListResponse is the paginated favorite-images feed.
type FavoriteImageListResponse struct {
	Items []FavoriteImageItem `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// GetFavoritePages godoc
//
//	@Summary		List the current user's favorited archive pages and audio images
//	@Description	Newest first across both media. Items whose source has since shrunk past the index are omitted.
//	@Tags			favorites
//	@Produce		json
//	@Param			page	query		int	false	"Page number (default 1)"
//	@Param			limit	query		int	false	"Page size (default 24, max 100)"
//	@Success		200		{object}	FavoriteImageListResponse
//	@Failure		500		{object}	response.Error
//	@Router			/api/favorites/pages [get]
func (h *FavoriteHandler) GetFavoritePages(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	page, limit := parsePagination(r, defaultArchiveLimit)
	offset := int64((page - 1) * limit)

	scope := auth.LibraryScopeFromContext(r.Context()).Unrestricted()

	total, err := h.queries.CountUserFavoriteImages(r.Context(), sqlc.CountUserFavoriteImagesParams{
		Uid:          userID,
		AllLibraries: scope,
	})
	if err != nil {
		h.logger.Error("count favorite images failed", "error", err)
		response.InternalError(w, "failed to count favorite images")

		return
	}

	rows, err := h.queries.GetUserFavoriteImageList(r.Context(), sqlc.GetUserFavoriteImageListParams{
		Uid:          userID,
		AllLibraries: scope,
		Offset:       offset,
		Limit:        int64(limit),
	})
	if err != nil {
		h.logger.Error("get favorite images failed", "error", err)
		response.InternalError(w, "failed to get favorite images")

		return
	}

	items := make([]FavoriteImageItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, FavoriteImageItem{
			Kind:        row.Kind,
			ArchiveID:   row.ArchiveID,
			AudioID:     row.AudioID,
			Title:       row.Title,
			PageCount:   int(row.ItemCount),
			PageIndex:   int(row.ItemIndex),
			FavoritedAt: row.FavoritedAt.Format(time.RFC3339),
			UpdatedAt:   row.UpdatedAt.Format(time.RFC3339),
		})
	}

	response.JSON(w, http.StatusOK, FavoriteImageListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}
