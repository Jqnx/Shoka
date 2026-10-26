package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// librarySourceResponse maps a stored per-library source row to its API shape.
// Shared by the read endpoint and the update path so the two can't drift.
func librarySourceResponse(row sqlc.LibrarySource) LibrarySourceResponse {
	resp := LibrarySourceResponse{
		Source:            row.Source,
		Enabled:           row.Enabled != 0,
		MagazineBlocklist: unmarshalStringSlice(row.MagazineBlocklist),
		MiscBlocklist:     unmarshalStringSlice(row.MiscBlocklist),
		Priority:          row.Priority,
	}

	if row.Cookies != nil {
		resp.Cookies = *row.Cookies
	}

	if row.ApiKey != nil {
		resp.APIKey = *row.ApiKey
	}

	return resp
}

// applyLibrarySourceUpdate overlays the request's present fields onto params,
// which the caller seeds from the library's existing source row. Pointer
// fields are why a partial PATCH can't wipe a stored value the caller never
// mentioned (see UpdateLibrarySourceRequest).
func applyLibrarySourceUpdate(params *sqlc.UpsertLibrarySourceParams, body UpdateLibrarySourceRequest) {
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
}

// orEmptyJSONList returns raw, or the empty JSON list when raw is blank, so a
// source row that predates a blocklist column can't be stored unparsable.
func orEmptyJSONList(raw string) string {
	if raw == "" {
		return "[]"
	}

	return raw
}

// ReorderLibrarySourcesRequest lists every remote source in its new run
// order. Local sources are pinned and must not appear.
type ReorderLibrarySourcesRequest struct {
	Sources []string `json:"sources"`
}

// ReorderLibrarySources godoc
//
//	@Summary		Reorder a library's remote metadata sources
//	@Description	Stores the order remote sources run in for this library. Local sources (ComicInfo, filename) always run first and are rejected here.
//	@Tags			admin
//	@Accept			json
//	@Param			id		path	string							true	"Library ID"
//	@Param			body	body	ReorderLibrarySourcesRequest	true	"Every remote source name, in the new order"
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/libraries/{id}/sources/order [put]
func (h *LibraryHandler) ReorderLibrarySources(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body ReorderLibrarySourcesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if !h.validSourceOrder(body.Sources) {
		response.BadRequest(w, "sources must be a duplicate-free list of every remote source in the new order")
		return
	}

	if err := h.applySourceOrder(r.Context(), id, body.Sources); err != nil {
		h.logger.Error("update source order failed", "id", id, "error", err)
		response.InternalError(w, "failed to update source order")

		return
	}

	response.NoContent(w)
}

// validSourceOrder reports whether names is exactly the set of registered
// remote source names, without duplicates.
func (h *LibraryHandler) validSourceOrder(names []string) bool {
	remote := h.pipeline.RemoteSourceNames()
	if len(names) == 0 || len(names) != len(remote) {
		return false
	}

	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		isLocal, known := h.pipeline.SourceIsLocal(name)
		if !known || isLocal {
			return false
		}

		if _, duplicate := seen[name]; duplicate {
			return false
		}

		seen[name] = struct{}{}
	}

	for _, name := range remote {
		if _, ok := seen[name]; !ok {
			return false
		}
	}

	return true
}

// applySourceOrder writes priorities 1..N in one transaction. A partial
// failure must not leave the remote rows half in the old order and half in
// the new one.
func (h *LibraryHandler) applySourceOrder(ctx context.Context, libraryID string, names []string) error {
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	qtx := h.queries.WithTx(tx)

	for i, name := range names {
		if err := h.ensureLibrarySource(ctx, qtx, libraryID, name); err != nil {
			return err
		}

		priority := int64(i + 1)
		if err := qtx.SetLibrarySourcePriority(ctx, sqlc.SetLibrarySourcePriorityParams{
			Priority:  &priority,
			LibraryID: libraryID,
			Source:    name,
		}); err != nil {
			return fmt.Errorf("set priority for %s: %w", name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

// ensureLibrarySource seeds a missing row with the same defaults a new
// library gets, so a source added after this library was created (or a row
// deleted out from under it) can still receive an order.
func (h *LibraryHandler) ensureLibrarySource(ctx context.Context, qtx *sqlc.Queries, libraryID, name string) error {
	if _, err := qtx.GetLibrarySource(ctx, sqlc.GetLibrarySourceParams{
		LibraryID: libraryID,
		Source:    name,
	}); err == nil {
		return nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("get source %s: %w", name, err)
	}

	enabled := false

	for _, def := range defaultLibrarySources {
		if def.Name == name {
			enabled = def.Enabled

			break
		}
	}

	if _, err := qtx.UpsertLibrarySource(ctx, sqlc.UpsertLibrarySourceParams{
		LibraryID:         libraryID,
		Source:            name,
		Enabled:           boolToInt(enabled),
		Cookies:           nil,
		ApiKey:            nil,
		MagazineBlocklist: "[]",
		MiscBlocklist:     "[]",
	}); err != nil {
		return fmt.Errorf("seed source %s: %w", name, err)
	}

	return nil
}
