package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"

	"github.com/go-chi/chi/v5"
)

type TagHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

func NewTagHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *TagHandler {
	return &TagHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "tag"),
	}
}

type TagResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Count       int64   `json:"count"`
}

type TagListResponse struct {
	Items []TagResponse `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

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

	total, err := h.queries.CountTags(r.Context())
	if err != nil {
		h.logger.Error("count tags failed", "error", err)
		response.InternalError(w, "failed to count tags")
		return
	}

	rows, err := h.queries.GetTagList(r.Context(), sqlc.GetTagListParams{
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list tags failed", "error", err)
		response.InternalError(w, "failed to list tags")
		return
	}

	items := make([]TagResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, TagResponse{
			ID:          row.ID,
			Name:        row.Name,
			Description: row.Description,
			Count:       row.Count,
		})
	}

	response.JSON(w, http.StatusOK, TagListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
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
	name := chi.URLParam(r, "name")
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

	total, err := h.queries.TotalArchiveWithTag(r.Context(), name)
	if err != nil {
		h.logger.Error("count archives by tag failed", "error", err)
		response.InternalError(w, "failed to count archives")
		return
	}

	rows, err := h.queries.GetArchivesByTagName(r.Context(), sqlc.GetArchivesByTagNameParams{
		Name:   name,
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list archives by tag failed", "error", err)
		response.InternalError(w, "failed to list archives")
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
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags/{id} [patch]
func (h *TagHandler) UpdateTagDescription(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid tag id")
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
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/tags/{id} [delete]
func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid tag id")
		return
	}

	if err := h.queries.DeleteTag(r.Context(), id); err != nil {
		h.logger.Error("delete tag failed", "error", err)
		response.InternalError(w, "failed to delete tag")
		return
	}

	response.NoContent(w)
}
