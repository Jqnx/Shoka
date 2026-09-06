package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/language"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CharacterHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

func NewCharacterHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *CharacterHandler {
	return &CharacterHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "character"),
	}
}

type CharacterResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type CharacterListResponse struct {
	Items []CharacterResponse `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// GetCharacters godoc
//
//	@Summary		List characters with pagination
//	@Tags			characters
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(50)
//	@Success		200	{object}	CharacterListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/characters [get]
func (h *CharacterHandler) GetCharacters(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.queries.CountCharacters(r.Context())
	if err != nil {
		h.logger.Error("count characters failed", "error", err)
		response.InternalError(w, "failed to count characters")

		return
	}

	rows, err := h.queries.GetCharacterList(r.Context(), sqlc.GetCharacterListParams{
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list characters failed", "error", err)
		response.InternalError(w, "failed to list characters")

		return
	}

	items := make([]CharacterResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, CharacterResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
		})
	}

	response.JSON(w, http.StatusOK, CharacterListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetAllCharacters godoc
//
//	@Summary		List every character
//	@Tags			characters
//	@Produce		json
//	@Success		200	{array}		CharacterResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/characters/all [get]
func (h *CharacterHandler) GetAllCharacters(w http.ResponseWriter, r *http.Request) {
	rows, err := h.queries.GetAllCharacter(r.Context())
	if err != nil {
		h.logger.Error("get all characters failed", "error", err)
		response.InternalError(w, "failed to get characters")

		return
	}

	items := make([]CharacterResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, CharacterResponse{
			ID:    row.ID,
			Name:  row.Name,
			Count: row.Count,
		})
	}

	response.JSON(w, http.StatusOK, items)
}

// GetArchivesByCharacter godoc
//
//	@Summary		List archives for a character
//	@Tags			characters
//	@Produce		json
//	@Param			name	path		string	true	"Character name"
//	@Param			page	query		int		false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int		false	"Items per page"		default(24)
//	@Success		200	{object}	ArchiveListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/characters/{name} [get]
func (h *CharacterHandler) GetArchivesByCharacter(w http.ResponseWriter, r *http.Request) {
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

	total, err := h.queries.TotalArchiveWithCharacter(r.Context(), name)
	if err != nil {
		h.logger.Error("count archives by character failed", "error", err)
		response.InternalError(w, "failed to count archives")

		return
	}

	rows, err := h.queries.GetArchivesByCharacterName(r.Context(), sqlc.GetArchivesByCharacterNameParams{
		Name:   name,
		Uid:    userID,
		Limit:  int64(limit),
		Offset: offset,
	})
	if err != nil {
		h.logger.Error("list archives by character failed", "error", err)
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

// DeleteCharacter godoc
//
//	@Summary		Delete a character
//	@Tags			characters
//	@Param			id	path	int	true	"Character ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/characters/{id} [delete]
func (h *CharacterHandler) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid character id")
		return
	}

	if err := h.queries.DeleteCharacter(r.Context(), id); err != nil {
		h.logger.Error("delete character failed", "error", err)
		response.InternalError(w, "failed to delete character")

		return
	}

	response.NoContent(w)
}
