package handlers

import (
	"Shoka/internal/api/response"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// LibraryBrowseEntry is one subdirectory of a browsed path.
type LibraryBrowseEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// LibraryBrowseResponse lists the subdirectories of one server path.
type LibraryBrowseResponse struct {
	Path    string               `json:"path"`
	Parent  string               `json:"parent,omitempty"`
	Entries []LibraryBrowseEntry `json:"entries"`
}

// BrowseLibraryPath godoc
//
//	@Summary		Browse server folders
//	@Tags			admin
//	@Produce		json
//	@Param			path query string false "Directory to list (defaults to /)"
//	@Success		200	{object}	LibraryBrowseResponse
//	@Failure		400	{object}	response.Error
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/browse [get]
func (h *LibraryHandler) BrowseLibraryPath(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSpace(r.URL.Query().Get("path"))
	if path == "" {
		path = "/"
	}

	absPath, ok := resolveLibraryPath(w, path)
	if !ok {
		return
	}

	// #nosec G703 -- browsing an arbitrary server directory is this admin-only
	// endpoint's purpose.
	dirents, err := os.ReadDir(absPath)
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			response.Forbidden(w, "permission denied reading directory")

			return
		}

		h.logger.Error("browse library path failed", "path", absPath, "error", err)
		response.InternalError(w, "failed to read directory")

		return
	}

	entries := make([]LibraryBrowseEntry, 0, len(dirents))

	for _, entry := range dirents {
		fullPath := filepath.Join(absPath, entry.Name())

		// Follow symlinks: a link to a directory is browsable, while a broken
		// link or a link to a file is not.
		// #nosec G703 -- the path is derived from the browsed directory, which
		// this admin-only endpoint is meant to expose.
		info, err := os.Stat(fullPath)
		if err != nil || !info.IsDir() {
			continue
		}

		entries = append(entries, LibraryBrowseEntry{Name: entry.Name(), Path: fullPath})
	}

	resp := LibraryBrowseResponse{Path: absPath, Entries: entries}
	if parent := filepath.Dir(absPath); parent != absPath {
		resp.Parent = parent
	}

	response.JSON(w, http.StatusOK, resp)
}
