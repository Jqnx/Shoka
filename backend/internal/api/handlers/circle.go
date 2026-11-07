package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// CircleHandler serves circle listing, archive lookups and member artists.
type CircleHandler struct {
	queries   *sqlc.Queries
	db        *sql.DB
	logger    *slog.Logger
	processor *image.Processor
}

// NewCircleHandler returns a CircleHandler backed by queries, the shared
// database handle and the image processor.
func NewCircleHandler(queries *sqlc.Queries, db *sql.DB, log *slog.Logger, processor *image.Processor) *CircleHandler {
	return &CircleHandler{
		queries:   queries,
		db:        db,
		processor: processor,
		logger:    log.With("handler", "circle"),
	}
}

// CircleResponse is the circle payload returned by the list endpoints.
type CircleResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// CircleListResponse is a single page of circles plus the total count.
type CircleListResponse struct {
	Items []CircleResponse `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// CircleArtistResponse is one artist linked to a circle.
type CircleArtistResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// circleItem maps a circle row to the circle payload.
func circleItem(row sqlc.Circle) CircleResponse { return CircleResponse(row) }

// GetCircles godoc
//
//	@Summary		List circles with pagination
//	@Tags			circles
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(50)
//	@Success		200	{object}	CircleListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/circles [get]
func (h *CircleHandler) GetCircles(w http.ResponseWriter, r *http.Request) {
	page, ok := listTaxonomy(
		w, r, h.logger, "circles",
		h.queries.TotalCircle, h.queries.GetCircleList, circleItem,
	)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, CircleListResponse(page))
}

// GetAllCircles godoc
//
//	@Summary		List every circle
//	@Tags			circles
//	@Produce		json
//	@Success		200	{array}		CircleResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/circles/all [get]
func (h *CircleHandler) GetAllCircles(w http.ResponseWriter, r *http.Request) {
	items, ok := allTaxonomy(w, r, h.logger, "circles", h.queries.GetAllCircle, circleItem)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// GetArchivesByCircle godoc
//
//	@Summary		List archives for a circle
//	@Tags			circles
//	@Produce		json
//	@Param			name	path		string	true	"Circle name"
//	@Param			page	query		int		false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int		false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/circles/{name} [get]
func (h *CircleHandler) GetArchivesByCircle(w http.ResponseWriter, r *http.Request) {
	archivesByName(
		w, r, h.logger, "circle", chi.URLParam(r, "name"),
		h.queries.TotalArchiveWithCircle, h.queries.GetArchivesByCircleName, h.processor,
		h.enrichArtists,
	)
}

// enrichArtists fills each card's artists, which the archives-by-taxonomy
// queries do not select. Only the circle page shows them.
func (h *CircleHandler) enrichArtists(ctx context.Context, items []ArchiveResponse) error {
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}

	metaByID, err := database.GetBulkArchiveMetadata(ctx, h.db, ids)
	if err != nil {
		return fmt.Errorf("get bulk archive metadata: %w", err)
	}

	for i := range items {
		if meta, ok := metaByID[items[i].ID]; ok {
			items[i].Artists = meta.Artists
		}
	}

	return nil
}

// GetCircleArtists godoc
//
//	@Summary		List artists linked to a circle
//	@Tags			circles
//	@Produce		json
//	@Param			name	path	string	true	"Circle name"
//	@Success		200	{array}		CircleArtistResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/circles/{name}/artists [get]
func (h *CircleHandler) GetCircleArtists(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetCircleArtists(r.Context(), chi.URLParam(r, "name"))
	if err != nil {
		h.logger.Error("get circle artists failed", "error", err)
		response.InternalError(w, "failed to get circle artists")

		return
	}

	items := make([]CircleArtistResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, CircleArtistResponse{ID: row.ID, Name: row.Name})
	}

	response.JSON(w, http.StatusOK, items)
}

// DeleteCircle godoc
//
//	@Summary		Delete a circle
//	@Tags			circles
//	@Param			id	path	int	true	"Circle ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/circles/{id} [delete]
func (h *CircleHandler) DeleteCircle(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "circle")
	if !ok {
		return
	}

	if err := h.queries.DeleteCircleByID(r.Context(), id); err != nil {
		h.logger.Error("delete circle failed", "error", err)
		response.InternalError(w, "failed to delete circle")

		return
	}

	response.NoContent(w)
}
