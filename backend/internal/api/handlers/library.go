package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// defaultLibrarySources seeds a newly created library's per-library
// metadata source settings with the same defaults the app previously used
// as global config.
var defaultLibrarySources = []struct {
	Name    string
	Enabled bool
}{
	{"comicinfo", true},
	{"filename", true},
	{"e-hentai", false},
	{"nhentai", false},
}

// LibraryHandler manages library CRUD, scans, covers, and per-library
// metadata source settings.
type LibraryHandler struct {
	queries   *sqlc.Queries
	db        *sql.DB
	queue     *jobs.Queue
	manager   *library.Manager
	processor *image.Processor
	logger    *slog.Logger
}

// NewLibraryHandler returns a LibraryHandler backed by queries, the job
// queue, the library manager, and the image processor that owns the
// per-archive cache directories.
func NewLibraryHandler(queries *sqlc.Queries, db *sql.DB, queue *jobs.Queue, manager *library.Manager, processor *image.Processor, log *slog.Logger) *LibraryHandler {
	return &LibraryHandler{
		queries:   queries,
		db:        db,
		queue:     queue,
		manager:   manager,
		processor: processor,
		logger:    log.With("handler", "library"),
	}
}

// LibraryResponse is the library payload returned by the library endpoints.
type LibraryResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
	// 0 means periodic scanning is off.
	ScanIntervalMinutes int64 `json:"scan_interval_minutes"`
	WatchEnabled        bool  `json:"watch_enabled"`
	// Omitted when the library has never been scanned.
	LastScannedAt string `json:"last_scanned_at,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

func toLibraryResponse(lib sqlc.Library) LibraryResponse {
	resp := LibraryResponse{
		ID:                  lib.ID,
		Name:                lib.Name,
		Path:                lib.Path,
		Type:                lib.Type,
		ScanIntervalMinutes: lib.ScanIntervalMinutes,
		WatchEnabled:        lib.WatchEnabled != 0,
		CreatedAt:           lib.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:           lib.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if lib.LastScannedAt != nil {
		resp.LastScannedAt = lib.LastScannedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return resp
}

// GetLibraries godoc
//
//	@Summary		List libraries
//	@Tags			libraries
//	@Produce		json
//	@Success		200	{array}		LibraryResponse
//	@Failure		500	{object}	response.Error
//	@Router			/api/libraries [get]
func (h *LibraryHandler) GetLibraries(w http.ResponseWriter, r *http.Request) {
	libs, err := h.queries.ListLibraries(r.Context())
	if err != nil {
		h.logger.Error("list libraries failed", "error", err)
		response.InternalError(w, "failed to list libraries")

		return
	}

	scope := auth.LibraryScopeFromContext(r.Context())

	items := make([]LibraryResponse, 0, len(libs))
	for _, lib := range libs {
		if !scope.Allows(lib.ID) {
			continue
		}

		items = append(items, toLibraryResponse(lib))
	}

	response.JSON(w, http.StatusOK, items)
}

// LibraryTypeResponse describes a library type and whether a scanner
// currently supports it.
type LibraryTypeResponse struct {
	Type      string `json:"type"`
	Supported bool   `json:"supported"`
}

// GetLibraryTypes godoc
//
//	@Summary		List library types
//	@Description	Returns every library type known to the schema and whether it's currently supported (has a working scanner). Unsupported types are reserved for future use and are rejected by library creation.
//	@Tags			libraries
//	@Produce		json
//	@Success		200	{array}		LibraryTypeResponse
//	@Router			/api/libraries/types [get]
func (h *LibraryHandler) GetLibraryTypes(w http.ResponseWriter, _ *http.Request) {
	items := make([]LibraryTypeResponse, 0, len(library.AllLibraryTypes))
	for _, t := range library.AllLibraryTypes {
		items = append(items, LibraryTypeResponse{
			Type:      t,
			Supported: h.manager.SupportsType(t),
		})
	}

	response.JSON(w, http.StatusOK, items)
}

// CreateLibraryRequest is the JSON body accepted by CreateLibrary.
type CreateLibraryRequest struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

// CreateLibrary godoc
//
//	@Summary		Create a library
//	@Tags			admin
//	@Accept			json
//	@Param			body	body	CreateLibraryRequest	true	"Library to create"
//	@Success		201	{object}	LibraryResponse
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		409	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries [post]
func (h *LibraryHandler) CreateLibrary(w http.ResponseWriter, r *http.Request) {
	body, ok := h.decodeCreateLibraryRequest(w, r)
	if !ok {
		return
	}

	absPath, ok := resolveLibraryPath(w, body.libraryPath)
	if !ok {
		return
	}

	if !h.checkLibraryOverlap(w, r, absPath) {
		return
	}

	lib, ok := h.insertLibrary(w, r, body, absPath)
	if !ok {
		return
	}

	h.seedDefaultLibrarySources(r.Context(), lib.ID)

	h.manager.OnLibraryChanged(lib)

	if err := h.queue.EnqueueOnce(r.Context(), jobs.JobTypeScan, jobs.ScanPayload{LibraryID: lib.ID}); err != nil {
		h.logger.Error("enqueue initial scan failed", "library_id", lib.ID, "error", err)
	}

	response.JSON(w, http.StatusCreated, toLibraryResponse(lib))
}

// Floor for a non-zero interval, so a hand-rolled request can't schedule a
// full library scan every minute.
const minScanIntervalMinutes = 15

// UpdateLibraryRequest is a partial update: omitted fields are unchanged.
// Pointers so saving the name alone can't reset the scan schedule to its
// zero value — same convention as UpdateLibrarySourceRequest.
type UpdateLibraryRequest struct {
	Name *string `json:"name"`
	// 0 disables periodic scanning; anything else must be >= 15.
	ScanIntervalMinutes *int64 `json:"scan_interval_minutes"`
	WatchEnabled        *bool  `json:"watch_enabled"`
}

// UpdateLibrary godoc
//
//	@Summary		Update a library's name and scan settings
//	@Description	Partial update — omitted fields are left unchanged. The library's path and type are immutable after creation. Set scan_interval_minutes to 0 to disable periodic scanning, or to at least 15 to enable it.
//	@Tags			admin
//	@Accept			json
//	@Param			id		path		string					true	"Library ID"
//	@Param			body	body	UpdateLibraryRequest	true	"Fields to update"
//	@Success		200	{object}	LibraryResponse
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id} [patch]
func (h *LibraryHandler) UpdateLibrary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	current, err := h.queries.GetLibraryByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "library not found")
			return
		}

		h.logger.Error("get library failed", "id", id, "error", err)
		response.InternalError(w, "failed to get library")

		return
	}

	var body UpdateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	name := current.Name
	if body.Name != nil {
		name = strings.TrimSpace(*body.Name)
		if name == "" {
			response.BadRequest(w, "name cannot be empty")
			return
		}
	}

	scanInterval := current.ScanIntervalMinutes
	if body.ScanIntervalMinutes != nil {
		scanInterval = *body.ScanIntervalMinutes
		if scanInterval < 0 {
			response.BadRequest(w, "scan_interval_minutes cannot be negative")
			return
		}

		if scanInterval > 0 && scanInterval < minScanIntervalMinutes {
			response.BadRequest(w, fmt.Sprintf("scan_interval_minutes must be 0 (disabled) or at least %d", minScanIntervalMinutes))
			return
		}
	}

	watchEnabled := current.WatchEnabled
	if body.WatchEnabled != nil {
		watchEnabled = boolToInt(*body.WatchEnabled)
	}

	updated, err := h.queries.UpdateLibrary(r.Context(), sqlc.UpdateLibraryParams{
		ID:                  id,
		Name:                name,
		ScanIntervalMinutes: scanInterval,
		WatchEnabled:        watchEnabled,
	})
	if err != nil {
		h.logger.Error("update library failed", "id", id, "error", err)
		response.InternalError(w, "failed to update library")

		return
	}

	h.manager.OnLibraryChanged(updated)

	response.JSON(w, http.StatusOK, toLibraryResponse(updated))
}

// DeleteLibrary godoc
//
//	@Summary		Delete a library
//	@Description	Cascade-deletes every archive and cached data belonging to this library. Does not touch files on disk.
//	@Tags			admin
//	@Param			id	path	string	true	"Library ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id} [delete]
func (h *LibraryHandler) DeleteLibrary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := h.queries.GetLibraryByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "library not found")
			return
		}

		h.logger.Error("get library failed", "id", id, "error", err)
		response.InternalError(w, "failed to get library")

		return
	}

	archiveIDs, err := database.DeleteLibrary(r.Context(), h.queries, h.db, id)
	if err != nil {
		h.logger.Error("delete library failed", "id", id, "error", err)
		response.InternalError(w, "failed to delete library")

		return
	}

	// After the commit, so a failed delete evicts nothing.
	for _, archiveID := range archiveIDs {
		if err := h.processor.EvictAll(archiveID); err != nil {
			h.logger.Warn("evict archive cache failed", "id", archiveID, "error", err)
		}
	}

	h.manager.OnLibraryDeleted(id)

	response.NoContent(w)
}

// ScanLibrary godoc
//
//	@Summary		Trigger a scan of a single library
//	@Tags			admin
//	@Param			id	path	string	true	"Library ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id}/scan [post]
func (h *LibraryHandler) ScanLibrary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := h.queries.GetLibraryByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "library not found")
			return
		}

		h.logger.Error("get library failed", "id", id, "error", err)
		response.InternalError(w, "failed to get library")

		return
	}

	if err := h.queue.EnqueueOnce(r.Context(), jobs.JobTypeScan, jobs.ScanPayload{LibraryID: id}); err != nil {
		h.logger.Error("enqueue scan failed", "id", id, "error", err)
		response.InternalError(w, "failed to enqueue scan")

		return
	}

	response.NoContent(w)
}

// ScanAllLibraries godoc
//
//	@Summary		Trigger a scan of every library
//	@Description	Enqueues one scan job per library, skipping types no scanner supports. Safe to repeat - a library whose scan is already pending or running is skipped.
//	@Tags			admin
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/scan [post]
func (h *LibraryHandler) ScanAllLibraries(w http.ResponseWriter, r *http.Request) {
	if err := h.manager.EnqueueAllLibraryScans(r.Context()); err != nil {
		h.logger.Error("enqueue all scans failed", "error", err)
		response.InternalError(w, "failed to enqueue scans")

		return
	}

	response.NoContent(w)
}

// GenerateLibraryCovers godoc
//
//	@Summary		Enqueue cover generation for one library's archives
//	@Description	Library-scoped counterpart to /api/admin/covers, which does the same for every library at once.
//	@Tags			admin
//	@Param			id	path	string	true	"Library ID"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id}/covers [post]
func (h *LibraryHandler) GenerateLibraryCovers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, err := h.queries.GetLibraryByID(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			response.NotFound(w, "library not found")
			return
		}

		h.logger.Error("get library failed", "id", id, "error", err)
		response.InternalError(w, "failed to get library")

		return
	}

	archives, err := h.queries.GetArchiveFilePathsByLibrary(r.Context(), id)
	if err != nil {
		h.logger.Error("get library archives failed", "id", id, "error", err)
		response.InternalError(w, "failed to get archives")

		return
	}

	for _, a := range archives {
		if err := h.queue.Enqueue(r.Context(), jobs.JobTypeCover, jobs.CoverPayload{
			ArchiveID: a.ID,
			FilePath:  a.FilePath,
			PageIndex: int(a.CoverPage),
		}); err != nil {
			h.logger.Error("enqueue cover job failed", "archive_id", a.ID, "error", err)
		}
	}

	response.NoContent(w)
}

// LibrarySourceResponse is a library's per-source metadata settings.
type LibrarySourceResponse struct {
	Source            string   `json:"source"`
	Enabled           bool     `json:"enabled"`
	Cookies           string   `json:"cookies,omitempty"`
	APIKey            string   `json:"api_key,omitempty"`
	MagazineBlocklist []string `json:"magazine_blocklist"`
	MiscBlocklist     []string `json:"misc_blocklist"`
}

// GetLibrarySources godoc
//
//	@Summary		Get a library's per-library metadata source settings
//	@Tags			admin
//	@Produce		json
//	@Param			id	path		string	true	"Library ID"
//	@Success		200	{array}		LibrarySourceResponse
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id}/sources [get]
func (h *LibraryHandler) GetLibrarySources(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	rows, err := h.queries.GetLibrarySources(r.Context(), id)
	if err != nil {
		h.logger.Error("get library sources failed", "id", id, "error", err)
		response.InternalError(w, "failed to get library sources")

		return
	}

	items := make([]LibrarySourceResponse, len(rows))
	for i, row := range rows {
		items[i] = librarySourceResponse(row)
	}

	response.JSON(w, http.StatusOK, items)
}

// UpdateLibrarySourceRequest is a partial update: omitted fields are left
// unchanged. Enabled/Cookies/APIKey use pointers so, e.g., toggling
// "enabled" alone (the settings dialog never opened) doesn't wipe a
// previously-saved API key — send an explicit value (empty string clears
// cookies/api_key) to change it. For the blocklists, omit the key to leave
// it unchanged, send an empty array to clear it, or send a full list to
// replace it.
type UpdateLibrarySourceRequest struct {
	Enabled           *bool    `json:"enabled"`
	Cookies           *string  `json:"cookies"`
	APIKey            *string  `json:"api_key"`
	MagazineBlocklist []string `json:"magazine_blocklist"`
	MiscBlocklist     []string `json:"misc_blocklist"`
}

// UpdateLibrarySource godoc
//
//	@Summary		Configure a metadata source for a single library
//	@Tags			admin
//	@Accept			json
//	@Param			id		path		string						true	"Library ID"
//	@Param			source	path		string						true	"Source name (e.g. nhentai, e-hentai, comicinfo, filename)"
//	@Param			body	body	UpdateLibrarySourceRequest	true	"Source settings"
//	@Success		200	{object}	LibrarySourceResponse
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id}/sources/{source} [patch]
func (h *LibraryHandler) UpdateLibrarySource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	source := chi.URLParam(r, "source")

	var body UpdateLibrarySourceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	existing, err := h.queries.GetLibrarySource(r.Context(), sqlc.GetLibrarySourceParams{
		LibraryID: id,
		Source:    source,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.logger.Error("get library source failed", "id", id, "source", source, "error", err)
		response.InternalError(w, "failed to update library source")

		return
	}

	params := sqlc.UpsertLibrarySourceParams{
		LibraryID:         id,
		Source:            source,
		Enabled:           existing.Enabled,
		Cookies:           existing.Cookies,
		ApiKey:            existing.ApiKey,
		MagazineBlocklist: orEmptyJSONList(existing.MagazineBlocklist),
		MiscBlocklist:     orEmptyJSONList(existing.MiscBlocklist),
	}

	applyLibrarySourceUpdate(&params, body)

	row, err := h.queries.UpsertLibrarySource(r.Context(), params)
	if err != nil {
		h.logger.Error("update library source failed", "id", id, "source", source, "error", err)
		response.InternalError(w, "failed to update library source")

		return
	}

	response.JSON(w, http.StatusOK, librarySourceResponse(row))
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}

	return 0
}

func unmarshalStringSlice(raw string) []string {
	if raw == "" {
		return []string{}
	}

	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return []string{}
	}

	return out
}

// findOverlappingLibrary returns the first existing library whose path is
// an ancestor of, equal to, or a descendant of candidatePath.
func findOverlappingLibrary(libs []sqlc.Library, candidatePath string) (sqlc.Library, bool) {
	for _, lib := range libs {
		if isSubPath(lib.Path, candidatePath) || isSubPath(candidatePath, lib.Path) {
			return lib, true
		}
	}

	return sqlc.Library{}, false
}

// isSubPath reports whether child is equal to, or nested inside, parent.
func isSubPath(parent, child string) bool {
	parent = filepath.Clean(parent)
	child = filepath.Clean(child)

	if parent == child {
		return true
	}

	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
