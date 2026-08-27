package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"slices"
)

type ReaderSettingsHandler struct {
	queries *sqlc.Queries
	logger  *slog.Logger
}

func NewReaderSettingsHandler(queries *sqlc.Queries, log *slog.Logger) *ReaderSettingsHandler {
	return &ReaderSettingsHandler{
		queries: queries,
		logger:  log.With("handler", "reader_settings"),
	}
}

// defaultReaderSettings is what a user with no reader_settings row gets - a
// missing row means "never customized", not "has no settings", so GET always
// returns a full, usable set of values either way.
var defaultReaderSettings = ReaderSettingsResponse{
	ViewMode:         "paged",
	ReadingDirection: "rtl",
	PageLayout:       "single",
	FitMode:          "width",
	Background:       "black",
}

var (
	// Scroll-vs-paged lives entirely in ViewMode now - ReadingDirection only
	// picks which physical side is "next" while paged (it's irrelevant in
	// scroll mode, which always scrolls top-to-bottom).
	validViewModes         = []string{"paged", "scroll"}
	validReadingDirections = []string{"ltr", "rtl"}
	validPageLayouts       = []string{"single", "double"}
	validFitModes          = []string{"width", "height", "original"}
	validBackgrounds       = []string{"black", "white", "gray"}
)

type ReaderSettingsResponse struct {
	ViewMode         string `json:"view_mode"`
	ReadingDirection string `json:"reading_direction"`
	PageLayout       string `json:"page_layout"`
	FitMode          string `json:"fit_mode"`
	Background       string `json:"background"`
}

func readerSettingsFromRow(row sqlc.ReaderSetting) ReaderSettingsResponse {
	return ReaderSettingsResponse{
		ViewMode:         row.ViewMode,
		ReadingDirection: row.ReadingDirection,
		PageLayout:       row.PageLayout,
		FitMode:          row.FitMode,
		Background:       row.Background,
	}
}

// GetReaderSettings godoc
//
//	@Summary		Get the current user's reader settings
//	@Description	Returns default values if the user hasn't customized anything yet.
//	@Tags			reader-settings
//	@Produce		json
//	@Success		200	{object}	ReaderSettingsResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/reader-settings [get]
func (h *ReaderSettingsHandler) GetReaderSettings(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	row, err := h.queries.GetReaderSettings(r.Context(), userID)
	if errors.Is(err, sql.ErrNoRows) {
		response.JSON(w, http.StatusOK, defaultReaderSettings)
		return
	}

	if err != nil {
		h.logger.Error("get reader settings failed", "error", err)
		response.InternalError(w, "failed to get reader settings")

		return
	}

	response.JSON(w, http.StatusOK, readerSettingsFromRow(row))
}

type UpdateReaderSettingsRequest struct {
	ViewMode         *string `json:"view_mode"`
	ReadingDirection *string `json:"reading_direction"`
	PageLayout       *string `json:"page_layout"`
	FitMode          *string `json:"fit_mode"`
	Background       *string `json:"background"`
}

// UpdateReaderSettings godoc
//
//	@Summary		Update the current user's reader settings
//	@Description	Partial update - omitted fields keep their current (or default) value. Creates the row on first call.
//	@Tags			reader-settings
//	@Accept			json
//	@Produce		json
//	@Param			body	UpdateReaderSettingsRequest	true	"Fields to change"
//	@Success		200		{object}	ReaderSettingsResponse
//	@Failure		400		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Router			/api/reader-settings [patch]
func (h *ReaderSettingsHandler) UpdateReaderSettings(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var body UpdateReaderSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := validateOneOf("view_mode", body.ViewMode, validViewModes); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := validateOneOf("reading_direction", body.ReadingDirection, validReadingDirections); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := validateOneOf("page_layout", body.PageLayout, validPageLayouts); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := validateOneOf("fit_mode", body.FitMode, validFitModes); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if err := validateOneOf("background", body.Background, validBackgrounds); err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Overlay the request onto whatever's current (falling back to defaults
	// for a user who's never customized anything) so the upsert below always
	// writes a complete row, even though the request itself is a partial
	// PATCH.
	current := defaultReaderSettings
	if row, err := h.queries.GetReaderSettings(r.Context(), userID); err == nil {
		current = readerSettingsFromRow(row)
	} else if !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get reader settings failed", "error", err)
		response.InternalError(w, "failed to update reader settings")

		return
	}

	if body.ViewMode != nil {
		current.ViewMode = *body.ViewMode
	}

	if body.ReadingDirection != nil {
		current.ReadingDirection = *body.ReadingDirection
	}

	if body.PageLayout != nil {
		current.PageLayout = *body.PageLayout
	}

	if body.FitMode != nil {
		current.FitMode = *body.FitMode
	}

	if body.Background != nil {
		current.Background = *body.Background
	}

	row, err := h.queries.UpsertReaderSettings(r.Context(), sqlc.UpsertReaderSettingsParams{
		UserID:           userID,
		ViewMode:         current.ViewMode,
		ReadingDirection: current.ReadingDirection,
		PageLayout:       current.PageLayout,
		FitMode:          current.FitMode,
		Background:       current.Background,
	})
	if err != nil {
		h.logger.Error("upsert reader settings failed", "error", err)
		response.InternalError(w, "failed to update reader settings")

		return
	}

	response.JSON(w, http.StatusOK, readerSettingsFromRow(row))
}

func validateOneOf(field string, value *string, allowed []string) error {
	if value == nil {
		return nil
	}

	if slices.Contains(allowed, *value) {
		return nil
	}

	return errors.New("invalid " + field)
}
