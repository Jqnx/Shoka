package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"Shoka/internal/api/response"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	"Shoka/internal/util"

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

type LibraryHandler struct {
	queries *sqlc.Queries
	queue   *jobs.Queue
	manager *library.Manager
	logger  *slog.Logger
}

func NewLibraryHandler(queries *sqlc.Queries, queue *jobs.Queue, manager *library.Manager, log *slog.Logger) *LibraryHandler {
	return &LibraryHandler{
		queries: queries,
		queue:   queue,
		manager: manager,
		logger:  log.With("handler", "library"),
	}
}

type LibraryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func toLibraryResponse(lib sqlc.Library) LibraryResponse {
	return LibraryResponse{
		ID:        lib.ID,
		Name:      lib.Name,
		Path:      lib.Path,
		Type:      lib.Type,
		CreatedAt: lib.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: lib.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
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

	items := make([]LibraryResponse, len(libs))
	for i, lib := range libs {
		items[i] = toLibraryResponse(lib)
	}

	response.JSON(w, http.StatusOK, items)
}

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
func (h *LibraryHandler) GetLibraryTypes(w http.ResponseWriter, r *http.Request) {
	items := make([]LibraryTypeResponse, 0, len(library.AllLibraryTypes))
	for _, t := range library.AllLibraryTypes {
		items = append(items, LibraryTypeResponse{
			Type:      t,
			Supported: h.manager.SupportsType(t),
		})
	}

	response.JSON(w, http.StatusOK, items)
}

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
//	@Param			body	body		CreateLibraryRequest	true	"Library to create"
//	@Success		201	{object}	LibraryResponse
//	@Failure		400	{object}	response.Error
//	@Failure		409	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries [post]
func (h *LibraryHandler) CreateLibrary(w http.ResponseWriter, r *http.Request) {
	var body CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Path = strings.TrimSpace(body.Path)
	if body.Type == "" {
		body.Type = "doujinshi"
	}

	if body.Name == "" || body.Path == "" {
		response.BadRequest(w, "name and path are required")
		return
	}

	if !h.manager.SupportsType(body.Type) {
		response.UnprocessableEntity(w, fmt.Sprintf("library type %q is not supported yet", body.Type))
		return
	}

	absPath, err := filepath.Abs(body.Path)
	if err != nil {
		response.BadRequest(w, "invalid path")
		return
	}

	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		response.BadRequest(w, "path does not exist or is not a directory")
		return
	}

	existing, err := h.queries.ListLibraries(r.Context())
	if err != nil {
		h.logger.Error("list libraries failed", "error", err)
		response.InternalError(w, "failed to validate library path")
		return
	}
	if overlap, ok := findOverlappingLibrary(existing, absPath); ok {
		response.BadRequest(w, fmt.Sprintf("path overlaps with existing library %q (%s)", overlap.Name, overlap.Path))
		return
	}

	var lib sqlc.Library
	for range 5 {
		id, err := util.GenerateID()
		if err != nil {
			response.InternalError(w, "failed to generate id")
			return
		}

		lib, err = h.queries.CreateLibrary(r.Context(), sqlc.CreateLibraryParams{
			ID:   id,
			Name: body.Name,
			Path: absPath,
			Type: body.Type,
		})
		if err == nil {
			break
		}
		if !database.IsUniqueConstraintError(err) {
			h.logger.Error("create library failed", "error", err)
			response.InternalError(w, "failed to create library")
			return
		}
		h.logger.Warn("library id collision, retrying")
	}
	if lib.ID == "" {
		response.InternalError(w, "failed to create library")
		return
	}

	for _, src := range defaultLibrarySources {
		if _, err := h.queries.UpsertLibrarySource(r.Context(), sqlc.UpsertLibrarySourceParams{
			LibraryID:         lib.ID,
			Source:            src.Name,
			Enabled:           boolToInt(src.Enabled),
			MagazineBlocklist: "[]",
			MiscBlocklist:     "[]",
		}); err != nil {
			h.logger.Error("seed library source failed", "library_id", lib.ID, "source", src.Name, "error", err)
		}
	}

	h.manager.OnLibraryChanged(r.Context(), lib)

	if err := h.queue.EnqueueOnce(r.Context(), jobs.JobTypeScan, jobs.ScanPayload{LibraryID: lib.ID}); err != nil {
		h.logger.Error("enqueue initial scan failed", "library_id", lib.ID, "error", err)
	}

	response.JSON(w, http.StatusCreated, toLibraryResponse(lib))
}

type UpdateLibraryRequest struct {
	Name *string `json:"name"`
}

// UpdateLibrary godoc
//
//	@Summary		Rename a library
//	@Description	The library's path and type are immutable after creation.
//	@Tags			admin
//	@Accept			json
//	@Param			id		path		string					true	"Library ID"
//	@Param			body	body		UpdateLibraryRequest	true	"Fields to update"
//	@Success		200	{object}	LibraryResponse
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

	updated, err := h.queries.UpdateLibrary(r.Context(), sqlc.UpdateLibraryParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		h.logger.Error("update library failed", "id", id, "error", err)
		response.InternalError(w, "failed to update library")
		return
	}

	h.manager.OnLibraryChanged(r.Context(), updated)

	response.JSON(w, http.StatusOK, toLibraryResponse(updated))
}

// DeleteLibrary godoc
//
//	@Summary		Delete a library
//	@Description	Cascade-deletes every archive and cached data belonging to this library. Does not touch files on disk.
//	@Tags			admin
//	@Param			id	path	string	true	"Library ID"
//	@Success		204
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

	if err := h.queries.DeleteLibrary(r.Context(), id); err != nil {
		h.logger.Error("delete library failed", "id", id, "error", err)
		response.InternalError(w, "failed to delete library")
		return
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
		items[i] = LibrarySourceResponse{
			Source:            row.Source,
			Enabled:           row.Enabled != 0,
			MagazineBlocklist: unmarshalStringSlice(row.MagazineBlocklist),
			MiscBlocklist:     unmarshalStringSlice(row.MiscBlocklist),
		}
		if row.Cookies != nil {
			items[i].Cookies = *row.Cookies
		}
		if row.ApiKey != nil {
			items[i].APIKey = *row.ApiKey
		}
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
//	@Param			body	body		UpdateLibrarySourceRequest	true	"Source settings"
//	@Success		200	{object}	LibrarySourceResponse
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
		MagazineBlocklist: existing.MagazineBlocklist,
		MiscBlocklist:     existing.MiscBlocklist,
	}
	if params.MagazineBlocklist == "" {
		params.MagazineBlocklist = "[]"
	}
	if params.MiscBlocklist == "" {
		params.MiscBlocklist = "[]"
	}

	if body.Enabled != nil {
		params.Enabled = boolToInt(*body.Enabled)
	}
	if body.Cookies != nil {
		params.Cookies = body.Cookies
	}
	if body.APIKey != nil {
		params.ApiKey = body.APIKey
	}
	if body.MagazineBlocklist != nil {
		magazineJSON, _ := json.Marshal(body.MagazineBlocklist)
		params.MagazineBlocklist = string(magazineJSON)
	}
	if body.MiscBlocklist != nil {
		miscJSON, _ := json.Marshal(body.MiscBlocklist)
		params.MiscBlocklist = string(miscJSON)
	}

	row, err := h.queries.UpsertLibrarySource(r.Context(), params)
	if err != nil {
		h.logger.Error("update library source failed", "id", id, "source", source, "error", err)
		response.InternalError(w, "failed to update library source")
		return
	}

	resp := LibrarySourceResponse{
		Source:            row.Source,
		Enabled:           row.Enabled != 0,
		MagazineBlocklist: unmarshalStringSlice(row.MagazineBlocklist),
		MiscBlocklist:     unmarshalStringSlice(row.MiscBlocklist),
	}
	if row.Cookies != nil {
		resp.Cookies = *row.Cookies
	}
	if row.ApiKey != nil {
		resp.APIKey = *row.ApiKey
	}

	response.JSON(w, http.StatusOK, resp)
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
