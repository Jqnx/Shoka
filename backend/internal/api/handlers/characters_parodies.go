package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Characters and parodies are the same shape of taxonomy entity - both are
// just {id, name, count} behind the four standard endpoints - so their
// handlers share this file. Tags and artists carry extra fields or endpoints
// and live in tag.go and artist.go.

// CharacterHandler serves character listing and archive lookups.
type CharacterHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

// NewCharacterHandler returns a CharacterHandler backed by queries and the
// shared image processor.
func NewCharacterHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *CharacterHandler {
	return &CharacterHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "character"),
	}
}

// CharacterResponse is the character payload returned by the list endpoints.
type CharacterResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// CharacterListResponse is a single page of characters plus the total count.
type CharacterListResponse struct {
	Items []CharacterResponse `json:"items"`
	Total int64               `json:"total"`
	Page  int                 `json:"page"`
	Limit int                 `json:"limit"`
}

// characterItem maps a character row to the character payload.
func characterItem(row sqlc.Character) CharacterResponse { return CharacterResponse(row) }

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
	page, ok := listTaxonomy(
		w, r, h.logger, "characters",
		h.queries.CountCharacters, h.queries.GetCharacterList, characterItem,
	)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, CharacterListResponse(page))
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
	items, ok := allTaxonomy(w, r, h.logger, "characters", h.queries.GetAllCharacter, characterItem)
	if !ok {
		return
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
	archivesByName(
		w, r, h.logger, "character", chi.URLParam(r, "name"),
		h.queries.TotalArchiveWithCharacter, h.queries.GetArchivesByCharacterName, h.processor,
	)
}

// DeleteCharacter godoc
//
//	@Summary		Delete a character
//	@Tags			characters
//	@Param			id	path	int	true	"Character ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/characters/{id} [delete]
func (h *CharacterHandler) DeleteCharacter(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "character")
	if !ok {
		return
	}

	if err := h.queries.DeleteCharacter(r.Context(), id); err != nil {
		h.logger.Error("delete character failed", "error", err)
		response.InternalError(w, "failed to delete character")

		return
	}

	response.NoContent(w)
}

// ParodyHandler serves parody listing and archive lookups.
type ParodyHandler struct {
	queries   *sqlc.Queries
	logger    *slog.Logger
	processor *image.Processor
}

// NewParodyHandler returns a ParodyHandler backed by queries and the shared
// image processor.
func NewParodyHandler(queries *sqlc.Queries, log *slog.Logger, processor *image.Processor) *ParodyHandler {
	return &ParodyHandler{
		queries:   queries,
		processor: processor,
		logger:    log.With("handler", "parody"),
	}
}

// ParodyResponse is the parody payload returned by the list endpoints.
type ParodyResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// ParodyListResponse is a single page of parodies plus the total count.
type ParodyListResponse struct {
	Items []ParodyResponse `json:"items"`
	Total int64            `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

// parodyItem maps a parody row to the parody payload.
func parodyItem(row sqlc.Parody) ParodyResponse { return ParodyResponse(row) }

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
	page, ok := listTaxonomy(
		w, r, h.logger, "parodies",
		h.queries.CountParodies, h.queries.GetParodyList, parodyItem,
	)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, ParodyListResponse(page))
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
	items, ok := allTaxonomy(w, r, h.logger, "parodies", h.queries.GetAllParody, parodyItem)
	if !ok {
		return
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
	archivesByName(
		w, r, h.logger, "parody", chi.URLParam(r, "name"),
		h.queries.TotalArchiveWithParody, h.queries.GetArchivesByParodyName, h.processor,
	)
}

// DeleteParody godoc
//
//	@Summary		Delete a parody
//	@Tags			parodies
//	@Param			id	path	int	true	"Parody ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/parodies/{id} [delete]
func (h *ParodyHandler) DeleteParody(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "parody")
	if !ok {
		return
	}

	if err := h.queries.DeleteParody(r.Context(), id); err != nil {
		h.logger.Error("delete parody failed", "error", err)
		response.InternalError(w, "failed to delete parody")

		return
	}

	response.NoContent(w)
}
