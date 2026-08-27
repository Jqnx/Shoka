package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ParodyHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

func NewParodyHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *ParodyHandler {
	return &ParodyHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "parody"),
	}
}

type ParodyResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type ParodyListResponse struct {
	Items []ParodyResponse `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// GetParodies godoc
//
//	@Summary		List parodies with pagination
//	@Tags			parodies
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(50)
//	@Success		200	{object}	ParodyListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/parodies [get]
func (h *ParodyHandler) GetParodies(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.queries.CountParodies(r.Context())
	if err != nil {
		h.logger.Error("count parodies failed", "error", err)
		response.InternalError(w, "failed to count parodies")

		return
	}

	rows, err := h.queries.GetParodyList(r.Context(), sqlc.GetParodyListParams{
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list parodies failed", "error", err)
		response.InternalError(w, "failed to list parodies")

		return
	}

	items := make([]ParodyResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ParodyResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
		})
	}

	response.JSON(w, http.StatusOK, ParodyListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetAllParodies godoc
//
//	@Summary		List every parody
//	@Tags			parodies
//	@Produce		json
//	@Success		200	{array}		ParodyResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/parodies/all [get]
func (h *ParodyHandler) GetAllParodies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetAllParody(r.Context())
	if err != nil {
		h.logger.Error("get all parodies failed", "error", err)
		response.InternalError(w, "failed to get parodies")

		return
	}

	items := make([]ParodyResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ParodyResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
		})
	}

	response.JSON(w, http.StatusOK, items)
}

// GetArchivesByParody godoc
//
//	@Summary		List archives for a parody
//	@Tags			parodies
//	@Produce		json
//	@Param			name	path		string	true	"Parody name"
//	@Param			page	query		int		false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int		false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/parodies/{name} [get]
func (h *ParodyHandler) GetArchivesByParody(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.queries.TotalArchiveWithParody(r.Context(), name)
	if err != nil {
		h.logger.Error("count archives by parody failed", "error", err)
		response.InternalError(w, "failed to count archives")

		return
	}

	rows, err := h.queries.GetArchivesByParodyName(r.Context(), sqlc.GetArchivesByParodyNameParams{
		Name:   name,
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list archives by parody failed", "error", err)
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

// DeleteParody godoc
//
//	@Summary		Delete a parody
//	@Tags			parodies
//	@Param			id	path	int	true	"Parody ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/parodies/{id} [delete]
func (h *ParodyHandler) DeleteParody(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid parody id")
		return
	}

	if err := h.queries.DeleteParody(r.Context(), id); err != nil {
		h.logger.Error("delete parody failed", "error", err)
		response.InternalError(w, "failed to delete parody")

		return
	}

	response.NoContent(w)
}
