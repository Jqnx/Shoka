package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
)

const accessSaveFailure = "failed to save library access"

// UserAccessHandler manages the per-user library allowlist. Every endpoint is
// admin-only; the access list of an admin is inert because admins bypass it.
type UserAccessHandler struct {
	queries *sqlc.Queries
	db      *sql.DB
	logger  *slog.Logger
}

// NewUserAccessHandler returns a UserAccessHandler backed by the given
// queries, database, and logger.
func NewUserAccessHandler(queries *sqlc.Queries, db *sql.DB, logger *slog.Logger) *UserAccessHandler {
	return &UserAccessHandler{
		queries: queries,
		db:      db,
		logger:  logger.With("handler", "user_access"),
	}
}

// LibraryAccessResponse is one user's granted library ids.
type LibraryAccessResponse struct {
	UserID     string   `json:"user_id"`
	LibraryIDs []string `json:"library_ids"`
}

// libraryAccessRequest is the body of PUT /api/admin/users/{id}/library-access.
type libraryAccessRequest struct {
	LibraryIDs []string `json:"library_ids"`
}

// requireUser answers 404 when the user does not exist and 500 when the check
// itself fails, reporting whether the handler should continue.
func (h *UserAccessHandler) requireUser(w http.ResponseWriter, r *http.Request, id, failure string) bool {
	exists, err := database.UserExists(r.Context(), h.db, id)
	if err != nil {
		h.logger.Error("check user exists failed", "user_id", id, "error", err)
		response.InternalError(w, failure)

		return false
	}

	if !exists {
		response.NotFound(w, "user not found")
		return false
	}

	return true
}

// uniqueLibraryIDs returns the caller's ids sorted and deduplicated, so the
// response and the rows written below describe the same set regardless of how
// they were ordered or repeated.
func uniqueLibraryIDs(raw []string) []string {
	ids := slices.Clone(raw)
	slices.Sort(ids)
	ids = slices.Compact(ids)

	// Callers echo this slice back to the client, where an omitted
	// "library_ids" must still serialize as [] rather than null.
	if ids == nil {
		return []string{}
	}

	return ids
}

// validateLibraries answers 400 for an id that names no library. Without this
// check a ghost id would reach the foreign key and surface as a 500.
func (h *UserAccessHandler) validateLibraries(w http.ResponseWriter, r *http.Request, ids []string) bool {
	if len(ids) == 0 {
		return true
	}

	libraries, err := h.queries.ListLibraries(r.Context())
	if err != nil {
		h.logger.Error("list libraries failed", "error", err)
		response.InternalError(w, accessSaveFailure)

		return false
	}

	known := make(map[string]bool, len(libraries))
	for _, library := range libraries {
		known[library.ID] = true
	}

	for _, id := range ids {
		if !known[id] {
			response.BadRequest(w, "unknown library: "+id)

			return false
		}
	}

	return true
}

// replaceAccess swaps the user's grants for ids in one transaction, so a
// failure part-way through cannot leave the user with a truncated allowlist.
func (h *UserAccessHandler) replaceAccess(w http.ResponseWriter, r *http.Request, id string, ids []string) bool {
	tx, err := h.db.BeginTx(r.Context(), nil)
	if err != nil {
		h.logger.Error("begin transaction failed", "user_id", id, "error", err)
		response.InternalError(w, accessSaveFailure)

		return false
	}
	defer func() { _ = tx.Rollback() }()

	qtx := h.queries.WithTx(tx)

	if err := qtx.DeleteLibraryAccessForUser(r.Context(), id); err != nil {
		h.logger.Error("delete library access failed", "user_id", id, "error", err)
		response.InternalError(w, accessSaveFailure)

		return false
	}

	for _, libraryID := range ids {
		if err := qtx.InsertLibraryAccess(r.Context(), sqlc.InsertLibraryAccessParams{
			UserID:    id,
			LibraryID: libraryID,
		}); err != nil {
			h.logger.Error("insert library access failed", "user_id", id, "library_id", libraryID, "error", err)
			response.InternalError(w, accessSaveFailure)

			return false
		}
	}

	if err := tx.Commit(); err != nil {
		h.logger.Error("commit library access failed", "user_id", id, "error", err)
		response.InternalError(w, accessSaveFailure)

		return false
	}

	return true
}

// GetLibraryAccess godoc
//
//	@Summary		Get a user's library access
//	@Description	Returns the library ids granted to the user. An admin's list is inert: admins can always access every library.
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	LibraryAccessResponse
//	@Failure		403	{object}	response.Error
//	@Failure		404	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/users/{id}/library-access [get]
func (h *UserAccessHandler) GetLibraryAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !h.requireUser(w, r, id, "failed to get library access") {
		return
	}

	ids, err := database.AccessibleLibraryIDs(r.Context(), h.db, id)
	if err != nil {
		h.logger.Error("list library access failed", "user_id", id, "error", err)
		response.InternalError(w, "failed to get library access")

		return
	}

	response.JSON(w, http.StatusOK, LibraryAccessResponse{UserID: id, LibraryIDs: ids})
}

// SetLibraryAccess godoc
//
//	@Summary		Replace a user's library access
//	@Description	Replaces the set of libraries the user may access. Unknown library ids are rejected. An empty list revokes every grant, so a non-admin sees nothing. Editing an admin's list is allowed but has no effect: admins bypass the access list.
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string					true	"User ID"
//	@Param			body	body		libraryAccessRequest	true	"Granted library ids"
//	@Success		200		{object}	LibraryAccessResponse
//	@Failure		403		{object}	response.Error
//	@Failure		400		{object}	response.Error
//	@Failure		404		{object}	response.Error
//	@Failure		500		{object}	response.Error
//	@Router			/api/admin/users/{id}/library-access [put]
func (h *UserAccessHandler) SetLibraryAccess(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if !h.requireUser(w, r, id, accessSaveFailure) {
		return
	}

	var body libraryAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ids := uniqueLibraryIDs(body.LibraryIDs)

	if !h.validateLibraries(w, r, ids) || !h.replaceAccess(w, r, id, ids) {
		return
	}

	response.JSON(w, http.StatusOK, LibraryAccessResponse{UserID: id, LibraryIDs: ids})
}
