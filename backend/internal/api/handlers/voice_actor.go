package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// VoiceActorHandler serves the voice actor taxonomy endpoints.
type VoiceActorHandler struct {
	queries *sqlc.Queries
	logger  *slog.Logger
}

// NewVoiceActorHandler returns a VoiceActorHandler backed by queries.
func NewVoiceActorHandler(queries *sqlc.Queries, log *slog.Logger) *VoiceActorHandler {
	return &VoiceActorHandler{
		queries: queries,
		logger:  log.With("handler", "voice_actor"),
	}
}

// VoiceActorResponse is the voice actor payload returned by the list
// endpoints.
type VoiceActorResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// VoiceActorListResponse is a single page of voice actors plus the total
// count.
type VoiceActorListResponse struct {
	Items []VoiceActorResponse `json:"items"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}

// voiceActorItem maps a voice actor row to the response payload.
func voiceActorItem(row sqlc.VoiceActor) VoiceActorResponse {
	return VoiceActorResponse{ID: row.ID, Name: row.Name, Count: row.Count}
}

// GetVoiceActors godoc
//
//	@Summary		List voice actors with pagination
//	@Tags			voice-actors
//	@Produce		json
//	@Param			page	query		int	false	"Page number (1-based)"	default(1)
//	@Param			limit	query		int	false	"Items per page"		default(50)
//	@Success		200	{object}	VoiceActorListResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/voice-actors [get]
func (h *VoiceActorHandler) GetVoiceActors(w http.ResponseWriter, r *http.Request) {
	page, ok := listTaxonomy(
		w, r, h.logger, "voice actors",
		h.queries.TotalVoiceActors, h.queries.GetVoiceActorList, voiceActorItem,
	)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, VoiceActorListResponse(page))
}

// GetAllVoiceActors godoc
//
//	@Summary		List every voice actor
//	@Tags			voice-actors
//	@Produce		json
//	@Success		200	{array}		VoiceActorResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/voice-actors/all [get]
func (h *VoiceActorHandler) GetAllVoiceActors(w http.ResponseWriter, r *http.Request) {
	items, ok := allTaxonomy(w, r, h.logger, "voice actors", h.queries.GetAllVoiceActors, voiceActorItem)
	if !ok {
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// DeleteVoiceActor godoc
//
//	@Summary		Delete a voice actor
//	@Description	Removes the voice actor. Its audio links cascade away; the works themselves remain.
//	@Tags			voice-actors
//	@Param			id	path	int	true	"Voice actor ID"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/voice-actors/{id} [delete]
func (h *VoiceActorHandler) DeleteVoiceActor(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r, "voice actor")
	if !ok {
		return
	}

	if err := h.queries.DeleteVoiceActor(r.Context(), id); err != nil {
		h.logger.Error("delete voice actor failed", "id", id, "error", err)
		response.InternalError(w, "failed to delete voice actor")

		return
	}

	response.NoContent(w)
}

// keep chi imported for the path parameter contract documented above.
var _ = chi.URLParam
