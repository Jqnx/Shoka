package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/spf13/viper"
)

type AdminHandler struct {
	Log     *slog.Logger
	Queries *sqlc.Queries
}

func NewAdminHandler(queries *sqlc.Queries, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		Queries: queries,
		Log:     logger.With("handler", "admin"),
	}
}

type UpdateSourceEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

// UpdateSourceEnabled godoc
//
//	@Summary		Enable or disable a metadata source
//	@Description	Toggles the enabled state of a named metadata source and persists the change to config
//	@Tags			admin
//	@Accept			json
//	@Param			source	path		string						true	"Source name (e.g. nhentai, e-hentai, comicinfo)"
//	@Param			body	body		UpdateSourceEnabledRequest	true	"Enabled state"
//	@Success		204
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/sources/{source} [patch]
func (h *AdminHandler) UpdateSourceEnabled(w http.ResponseWriter, r *http.Request) {
	sourceName := chi.URLParam(r, "source")

	var body UpdateSourceEnabledRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	key := fmt.Sprintf("metadata.sources.%s.enabled", sourceName)
	viper.Set(key, body.Enabled)

	if err := viper.WriteConfig(); err != nil {
		h.Log.Error("failed to write config", "error", err)
		response.InternalError(w, "failed to save config")
		return
	}

	response.NoContent(w)
}
