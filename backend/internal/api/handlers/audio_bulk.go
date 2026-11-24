package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/metadata"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// BulkUpdateAudioPlayedRequest marks the listed audio works played or clears
// their played state for the caller.
type BulkUpdateAudioPlayedRequest struct {
	IDs    []string `json:"ids"`
	Played bool     `json:"played"`
}

// BulkUpdateAudiosRequest edits shared metadata across many audio works.
//
// Relations (artists/tags/voice_actors) MERGE into what each work already
// has, so each keeps its own and can only gain the submitted names. Scalars
// (language/circle) are SET, overwriting the current value; omit or null them
// to leave them alone, send "" to clear.
type BulkUpdateAudiosRequest struct {
	IDs         []string `json:"ids"`
	Artists     []string `json:"artists"`
	Tags        []string `json:"tags"`
	VoiceActors []string `json:"voice_actors"`
	Language    *string  `json:"language"`
	Circle      *string  `json:"circle"`
}

// hasUpdate reports whether the request asks for at least one change.
func (b *BulkUpdateAudiosRequest) hasUpdate() bool {
	return b.Artists != nil || b.Tags != nil || b.VoiceActors != nil ||
		b.Language != nil || b.Circle != nil
}

// canonicalizeBulkAudioNames snaps submitted names to the spelling already
// stored in each entity table. Doing it once per request rather than per work
// is enough because the mapping is global to the entity table.
func canonicalizeBulkAudioNames(r *http.Request, db *sql.DB, body *BulkUpdateAudiosRequest) error {
	for _, field := range []struct {
		table string
		names *[]string
	}{
		{"artist", &body.Artists},
		{"tag", &body.Tags},
		{"voice_actor", &body.VoiceActors},
	} {
		canonical, err := canonicalizeNames(r, db, field.table, *field.names)
		if err != nil {
			return fmt.Errorf("canonicalize %s: %w", field.table, err)
		}

		*field.names = canonical
	}

	if body.Circle != nil {
		canonical, err := canonicalizeNames(r, db, "circle", []string{*body.Circle})
		if err != nil {
			return fmt.Errorf("canonicalize circle: %w", err)
		}

		body.Circle = &canonical[0]
	}

	return nil
}

// partitionExistingAudios resolves which of ids are real audio works,
// reporting the rest as per-id failures in caller order.
func (h *AudioHandler) partitionExistingAudios(ctx context.Context, ids []string) ([]sqlc.GetAudiosByIDsRow, []BulkFailure, error) {
	rows, err := h.queries.GetAudiosByIDs(ctx, ids)
	if err != nil {
		return nil, nil, err
	}

	found := make(map[string]bool, len(rows))
	for _, row := range rows {
		found[row.ID] = true
	}

	// Non-nil so an all-succeeded batch serializes "failed": [] rather than
	// null.
	failed := []BulkFailure{}

	for _, id := range ids {
		if !found[id] {
			failed = append(failed, BulkFailure{ID: id, Error: "audio not found"})
		}
	}

	return rows, failed, nil
}

// checkAudioScope rejects the whole request with 403 when any resolved work
// belongs to a library the caller cannot access.
func checkAudioScope(w http.ResponseWriter, r *http.Request, rows []sqlc.GetAudiosByIDsRow) bool {
	scope := auth.LibraryScopeFromContext(r.Context())
	for _, row := range rows {
		if !scope.Allows(row.LibraryID) {
			response.Forbidden(w, "no access to one or more audio works")

			return false
		}
	}

	return true
}

// BulkUpdateAudioPlayed godoc
//
//	@Summary		Mark audio works played or unplayed in bulk
//	@Description	Marks every listed work played for the calling user, or clears its played state. Played state is per-user, so any authenticated user may call this. Ids that don't resolve to an audio work are reported in "failed" rather than failing the request. A batch containing a work the caller cannot access fails the whole request with 403.
//	@Tags			audio
//	@Accept			json
//	@Produce		json
//	@Param			body	body	BulkUpdateAudioPlayedRequest	true	"Audio works and desired played state"
//	@Success		200	{object}	BulkResult
//	@Failure		400	{object}	response.Error
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/bulk/played [put]
func (h *AudioHandler) BulkUpdateAudioPlayed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body BulkUpdateAudioPlayedRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ids, err := normalizeBulkIDs(body.IDs)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	rows, failed, err := h.partitionExistingAudios(ctx, ids)
	if err != nil {
		h.logger.Error("get audios by ids failed", "error", err)
		response.InternalError(w, "failed to load audio works")

		return
	}

	if !checkAudioScope(w, r, rows) {
		return
	}

	existingIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		existingIDs = append(existingIDs, row.ID)
	}

	if len(existingIDs) > 0 {
		if body.Played {
			err = h.queries.MarkAudioPlayedBulk(ctx, sqlc.MarkAudioPlayedBulkParams{
				Uid: auth.UserIDFromContext(ctx),
				Ids: existingIDs,
			})
		} else {
			err = h.queries.MarkAudioUnplayedBulk(ctx, sqlc.MarkAudioUnplayedBulkParams{
				Uid: auth.UserIDFromContext(ctx),
				Ids: existingIDs,
			})
		}

		if err != nil {
			h.logger.Error("bulk audio played update failed", "played", body.Played, "error", err)
			response.InternalError(w, "failed to update played state")

			return
		}
	}

	response.JSON(w, http.StatusOK, BulkResult{
		Requested: len(ids),
		Succeeded: len(existingIDs),
		Failed:    failed,
	})
}

// BulkUpdateAudios godoc
//
//	@Summary		Update shared metadata on audio works in bulk
//	@Description	Relations (artists/tags/voice_actors) MERGE into what each work already has rather than replacing it, so a work keeps its own values and gains the submitted ones; de-duplication is case-insensitive and submitted names are snapped to the spelling already in use. Because it only adds, an empty array is a no-op and bulk removal is unsupported. Scalars (language/circle) are SET, overwriting the current value; omit or null them to leave them alone, send "" to clear. Circle accepts at most one value. Each work is applied in its own transaction, so a failure part-way through leaves earlier works updated and is reported per-id in "failed".
//	@Tags			audio
//	@Accept			json
//	@Produce		json
//	@Param			body	body	BulkUpdateAudiosRequest	true	"Audio works and metadata to merge or set"
//	@Success		200	{object}	BulkResult
//	@Failure		400	{object}	response.Error
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/audio/bulk [patch]
func (h *AudioHandler) BulkUpdateAudios(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body BulkUpdateAudiosRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ids, err := normalizeBulkIDs(body.IDs)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if !body.hasUpdate() {
		response.BadRequest(w, "at least one of artists, tags, voice_actors, language or circle is required")
		return
	}

	rows, failed, err := h.partitionExistingAudios(ctx, ids)
	if err != nil {
		h.logger.Error("get audios by ids failed", "error", err)
		response.InternalError(w, "failed to load audio works")

		return
	}

	if !checkAudioScope(w, r, rows) {
		return
	}

	existingIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		existingIDs = append(existingIDs, row.ID)
	}

	// One bulk read of current relations rather than a lookup per work - the
	// merge needs every work's existing values.
	current, err := database.GetBulkAudioMetadata(ctx, h.db, existingIDs)
	if err != nil {
		h.logger.Error("get bulk audio metadata failed", "error", err)
		response.InternalError(w, "failed to load current metadata")

		return
	}

	if err := canonicalizeBulkAudioNames(r, h.db, &body); err != nil {
		h.logger.Error("canonicalize names failed", "error", err)
		response.InternalError(w, "failed to resolve existing metadata names")

		return
	}

	succeeded := 0

	for _, row := range rows {
		if err := metadata.ApplyAudioMetadata(ctx, h.queries, h.db, row.ID, bulkAudioResult(row, current[row.ID], &body)); err != nil {
			h.logger.Error("bulk update audio failed", "audio_id", row.ID, "error", err)
			failed = append(failed, BulkFailure{ID: row.ID, Error: "failed to update audio work"})

			continue
		}

		succeeded++
	}

	response.JSON(w, http.StatusOK, BulkResult{
		Requested: len(ids),
		Succeeded: succeeded,
		Failed:    failed,
	})
}

// bulkAudioResult builds the metadata to apply to one audio work. Relations
// are merged into the work's current values and preserved when the request
// omitted them, because ApplyAudioMetadata replaces the full set.
func bulkAudioResult(row sqlc.GetAudiosByIDsRow, current *metadata.AudioResult, body *BulkUpdateAudiosRequest) *metadata.AudioResult {
	if current == nil {
		current = &metadata.AudioResult{}
	}

	language := row.Language
	if body.Language != nil {
		if name := strings.TrimSpace(*body.Language); name == "" {
			language = nil
		} else {
			language = &name
		}
	}

	circles := current.Circles

	if body.Circle != nil {
		if name := strings.TrimSpace(*body.Circle); name == "" {
			circles = nil
		} else {
			circles = []string{name}
		}
	}

	return &metadata.AudioResult{
		Title:       row.Title,
		ReleaseDate: row.ReleaseDate,
		Language:    language,
		Artists:     mergeAudioNames(current.Artists, body.Artists),
		Tags:        mergeAudioNames(current.Tags, body.Tags),
		Circles:     circles,
		VoiceActors: mergeAudioNames(current.VoiceActors, body.VoiceActors),
	}
}

// mergeAudioNames keeps the existing list when the request omitted a relation;
// ApplyAudioMetadata treats nil and empty alike, as "clear".
func mergeAudioNames(existing, add []string) []string {
	if add == nil {
		return existing
	}

	return mergeNames(existing, add)
}
