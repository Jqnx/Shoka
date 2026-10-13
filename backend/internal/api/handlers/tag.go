package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// TagHandler serves tag listing, description updates, and archive lookups.
type TagHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

// NewTagHandler returns a TagHandler backed by queries and the shared image
// processor.
func NewTagHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *TagHandler {
	return &TagHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "tag"),
	}
}

// TagResponse is the tag payload returned by the list endpoints.
type TagResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Count       int64   `json:"count"`
}

// TagListResponse is a single page of tags plus the total count.
type TagListResponse struct {
	Items []TagResponse `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

// tagItem maps a tag row to the tag payload.
func tagItem(row sqlc.Tag) TagResponse { return TagResponse(row) }

// GetTags godoc
//
//	@Summary		List tags with pagination
//	@Tags			tags
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(50)
//	@Success		200	{object}	TagListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags [get]
func (h *TagHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	page, ok := listTaxonomy(
		w, r, h.logger, "tags",
		h.queries.CountTags, h.queries.GetTagList, tagItem,
	)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, TagListResponse(page))
}

// GetAllTags godoc
//
//	@Summary		List every tag
//	@Tags			tags
//	@Produce		json
//	@Success		200	{array}		TagResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags/all [get]
func (h *TagHandler) GetAllTags(w http.ResponseWriter, r *http.Request) {
	items, ok := allTaxonomy(w, r, h.logger, "tags", h.queries.GetAllTags, tagItem)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// GetArchivesByTag godoc
//
//	@Summary		List archives for a tag
//	@Tags			tags
//	@Produce		json
//	@Param			name	path		string	true	"Tag name"
//	@Param			page	query		int		false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int		false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags/{name} [get]
func (h *TagHandler) GetArchivesByTag(w http.ResponseWriter, r *http.Request) {
	archivesByName(
		w, r, h.logger, "tag", chi.URLParam(r, "name"),
		h.queries.TotalArchiveWithTag, h.queries.GetArchivesByTagName, h.processor,
	)
}

// UpdateTagDescriptionRequest is the JSON body accepted by
// UpdateTagDescription.
type UpdateTagDescriptionRequest struct {
	Description *string `json:"description"`
}

// UpdateTagDescription godoc
//
//	@Summary		Update a tag's description
//	@Tags			tags
//	@Accept			json
//	@Param			id		path	int							true	"Tag ID"
//	@Param			body	body	UpdateTagDescriptionRequest	true	"Description"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags/{id} [patch]
func (h *TagHandler) UpdateTagDescription(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "tag")
	if !ok {
		return
	}

	var body UpdateTagDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.queries.UpdateTagDescription(r.Context(), sqlc.UpdateTagDescriptionParams{
		Description: body.Description,
		ID:          id,
	}); err != nil {
		h.logger.Error("update tag description failed", "error", err)
		response.InternalError(w, "failed to update tag")

		return
	}

	response.NoContent(w)
}

// DeleteTag godoc
//
//	@Summary		Delete a tag
//	@Tags			tags
//	@Param			id	path	int	true	"Tag ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags/{id} [delete]
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "tag")
	if !ok {
		return
	}

	if err := h.queries.DeleteTag(r.Context(), id); err != nil {
		h.logger.Error("delete tag failed", "error", err)
		response.InternalError(w, "failed to delete tag")

		return
	}

	response.NoContent(w)
}
