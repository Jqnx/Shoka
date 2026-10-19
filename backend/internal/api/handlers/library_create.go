package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/util"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// createLibraryParams is the decoded, validated form of a
// CreateLibraryRequest: trimmed, defaulted, and known to name a supported
// library type (but not yet known to point at a usable directory).
type createLibraryParams struct {
	name, libraryPath, libraryType string
}

// decodeCreateLibraryRequest decodes a CreateLibraryRequest and applies the
// request-only validations in the order the endpoint documents them. It
// writes the failure response itself and reports ok=false when the request
// isn't usable.
func (h *LibraryHandler) decodeCreateLibraryRequest(w http.ResponseWriter, r *http.Request) (createLibraryParams, bool) {
	var body CreateLibraryRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")

		return createLibraryParams{}, false
	}

	params := createLibraryParams{
		name:        strings.TrimSpace(body.Name),
		libraryPath: strings.TrimSpace(body.Path),
		libraryType: body.Type,
	}
	if params.libraryType == "" {
		params.libraryType = "doujinshi"
	}

	if params.name == "" || params.libraryPath == "" {
		response.BadRequest(w, "name and path are required")

		return createLibraryParams{}, false
	}

	if !h.manager.SupportsType(params.libraryType) {
		response.UnprocessableEntity(w, fmt.Sprintf("library type %q is not supported yet", params.libraryType))

		return createLibraryParams{}, false
	}

	return params, true
}

// resolveLibraryPath makes the requested path absolute and checks it is an
// existing directory, writing the failure response itself.
func resolveLibraryPath(w http.ResponseWriter, path string) (string, bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		response.BadRequest(w, "invalid path")

		return "", false
	}

	// #nosec G703 -- the caller-supplied path is intentional; both callers are
	// admin-gated.
	info, err := os.Stat(absPath)
	if err != nil || !info.IsDir() {
		response.BadRequest(w, "path does not exist or is not a directory")

		return "", false
	}

	return absPath, true
}

// checkLibraryOverlap rejects candidatePath when it is the same directory as,
// nested inside, or an ancestor of an existing library's path - two libraries
// over the same tree would scan and store every archive twice. It writes the
// failure response itself.
func (h *LibraryHandler) checkLibraryOverlap(w http.ResponseWriter, r *http.Request, candidatePath string) bool {
	existing, err := h.queries.ListLibraries(r.Context())
	if err != nil {
		h.logger.Error("list libraries failed", "error", err)
		response.InternalError(w, "failed to validate library path")

		return false
	}

	if overlap, ok := findOverlappingLibrary(existing, candidatePath); ok {
		response.BadRequest(w, fmt.Sprintf("path overlaps with existing library %q (%s)", overlap.Name, overlap.Path))

		return false
	}

	return true
}

// libraryIDAttempts is how many times a colliding generated library id is
// retried before the create is reported as failed. The generator is random,
// so a collision is vanishingly unlikely once, let alone five times.
const libraryIDAttempts = 5

// insertLibrary creates the library under a freshly generated id, retrying
// when that id already exists (the only recoverable failure - any other error
// is reported immediately). It writes the failure response itself and reports
// ok=false when no library was created.
func (h *LibraryHandler) insertLibrary(w http.ResponseWriter, r *http.Request, params createLibraryParams, absPath string) (sqlc.Library, bool) {
	for range libraryIDAttempts {
		id, err := util.GenerateID()
		if err != nil {
			response.InternalError(w, "failed to generate id")

			return sqlc.Library{}, false
		}

		lib, err := h.queries.CreateLibrary(r.Context(), sqlc.CreateLibraryParams{
			ID:   id,
			Name: params.name,
			Path: absPath,
			Type: params.libraryType,
		})
		if err == nil {
			return lib, true
		}

		if !database.IsUniqueConstraintError(err) {
			h.logger.Error("create library failed", "error", err)
			response.InternalError(w, "failed to create library")

			return sqlc.Library{}, false
		}

		h.logger.Warn("library id collision, retrying")
	}

	response.InternalError(w, "failed to create library")

	return sqlc.Library{}, false
}

// seedDefaultLibrarySources inserts the default per-source settings rows for a
// freshly created library. A failure here is logged and otherwise ignored:
// the library itself exists and can be scanned, and the settings dialog
// upserts a missing row the first time it is saved.
func (h *LibraryHandler) seedDefaultLibrarySources(ctx context.Context, libraryID string) {
	for _, src := range defaultLibrarySources {
		if _, err := h.queries.UpsertLibrarySource(ctx, sqlc.UpsertLibrarySourceParams{
			LibraryID:         libraryID,
			Source:            src.Name,
			Enabled:           boolToInt(src.Enabled),
			MagazineBlocklist: "[]",
			MiscBlocklist:     "[]",
		}); err != nil {
			h.logger.Error("seed library source failed", "library_id", libraryID, "source", src.Name, "error", err)
		}
	}
}
