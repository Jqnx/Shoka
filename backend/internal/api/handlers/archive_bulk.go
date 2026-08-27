package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/auth"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"Shoka/internal/metadata"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// maxBulkArchives caps how many archives a single bulk request may touch.
// The cap exists mostly to keep the id list well clear of SQLite's bound
// parameter limit, and to stop one request from monopolising the worker
// pool - it is not a meaningful workflow limit at this app's scale.
const maxBulkArchives = 1000

// BulkFailure explains why one archive in a batch didn't get processed.
type BulkFailure struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// BulkResult is the shared response for every bulk endpoint. Bulk
// operations are deliberately partial-success: one bad archive reports
// itself in Failed rather than rolling back the ones that worked, since
// re-running a whole batch to get past a single broken archive is worse
// than knowing precisely which one to fix.
type BulkResult struct {
	Requested int           `json:"requested"`
	Succeeded int           `json:"succeeded"`
	Failed    []BulkFailure `json:"failed"`
}

// normalizeBulkIDs trims, drops blanks, and de-duplicates the caller's id
// list. De-duplication matters beyond tidiness: a repeated id would
// otherwise be counted twice in the result totals.
func normalizeBulkIDs(raw []string) ([]string, error) {
	seen := make(map[string]bool, len(raw))

	ids := make([]string, 0, len(raw))
	for _, id := range raw {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}

		seen[id] = true
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, errors.New("ids is required and must contain at least one archive id")
	}

	if len(ids) > maxBulkArchives {
		return nil, errors.New("too many ids in one request")
	}

	return ids, nil
}

// partitionExisting splits ids into those that resolve to a real archive
// and BulkFailures for those that don't, so a stale id in the caller's
// selection is reported rather than silently ignored.
func (h *ArchiveHandler) partitionExisting(r *http.Request, ids []string) ([]sqlc.GetArchivesByIDsRow, []BulkFailure, error) {
	rows, err := h.queries.GetArchivesByIDs(r.Context(), ids)
	if err != nil {
		return nil, nil, err
	}

	found := make(map[string]bool, len(rows))
	for _, row := range rows {
		found[row.ID] = true
	}

	// Non-nil so an all-succeeded batch serializes "failed": [] rather than
	// null - this codebase has been bitten before by Go's nil slices
	// reaching the frontend as null (see the note on Archive's relation
	// fields in types.ts).
	failed := []BulkFailure{}

	for _, id := range ids {
		if !found[id] {
			failed = append(failed, BulkFailure{ID: id, Error: "archive not found"})
		}
	}

	return rows, failed, nil
}

type BulkProgressRequest struct {
	IDs []string `json:"ids"`
	// Read true marks every archive finished (its own last page); false
	// clears progress entirely, the same as DELETE .../{id}/progress.
	Read bool `json:"read"`
}

// BulkUpdateProgress godoc
//
//	@Summary		Mark archives read or unread in bulk
//	@Description	Marking read jumps each archive to its own last page and flags it completed; marking unread deletes the progress row outright (page 0 would still count as in-progress). Ids that don't resolve to an archive are reported in "failed" rather than failing the request.
//	@Tags			archives
//	@Accept			json
//	@Produce		json
//	@Param			body	BulkProgressRequest	true	"Archives and desired read state"
//	@Success		200	{object}	BulkResult
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/bulk/progress [put]
func (h *ArchiveHandler) BulkUpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID := auth.UserIDFromContext(r.Context())

	var body BulkProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ids, err := normalizeBulkIDs(body.IDs)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	rows, failed, err := h.partitionExisting(r, ids)
	if err != nil {
		h.logger.Error("get archives by ids failed", "error", err)
		response.InternalError(w, "failed to load archives")

		return
	}

	existing := make([]string, 0, len(rows))
	for _, row := range rows {
		existing = append(existing, row.ID)
	}

	if len(existing) > 0 {
		if body.Read {
			err = h.queries.MarkArchivesRead(r.Context(), sqlc.MarkArchivesReadParams{
				Uid: userID,
				Ids: existing,
			})
		} else {
			err = h.queries.MarkArchivesUnread(r.Context(), sqlc.MarkArchivesUnreadParams{
				Uid: userID,
				Ids: existing,
			})
		}

		if err != nil {
			h.logger.Error("bulk progress update failed", "read", body.Read, "error", err)
			response.InternalError(w, "failed to update progress")

			return
		}
	}

	response.JSON(w, http.StatusOK, BulkResult{
		Requested: len(ids),
		Succeeded: len(existing),
		Failed:    failed,
	})
}

type BulkMetadataRequest struct {
	IDs []string `json:"ids"`
	// Source optionally pins the fetch to one named metadata source
	// (comicinfo, filename, e-hentai, nhentai). Omit it to run the normal
	// priority-ordered pipeline.
	Source string `json:"source"`
}

// BulkFetchMetadata godoc
//
//	@Summary		Fetch and apply metadata for archives in bulk
//	@Description	Enqueues the same metadata job the scanner uses for newly added archives: local sources run first and are applied, then a remote job is chained to fill remaining gaps. Pass "source" to pin the fetch to a single named source instead, in which case only that source runs and nothing is chained; network-backed sources are queued on the low-concurrency worker so a large batch doesn't hammer them. A source that is disabled for a given archive's library is skipped rather than retried. Returns as soon as the jobs are queued - "succeeded" counts archives queued, not archives whose metadata has been written yet. Unlike the single-archive fetch endpoints there is no preview/confirm step, so results are applied automatically.
//	@Tags			archives
//	@Accept			json
//	@Produce		json
//	@Param			body	BulkMetadataRequest	true	"Archives to identify, optionally pinned to one source"
//	@Success		202	{object}	BulkResult
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/bulk/metadata [post]
func (h *ArchiveHandler) BulkFetchMetadata(w http.ResponseWriter, r *http.Request) {
	var body BulkMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ids, err := normalizeBulkIDs(body.IDs)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	// Route by the source's own locality: remote sources go on the
	// single-slot worker (see main.go's Register limits) so a bulk fetch
	// can't fire five concurrent scrapes at the same site.
	jobType := jobs.JobTypeMetadata

	if source := strings.TrimSpace(body.Source); source != "" {
		isLocal, known := h.pipeline.SourceIsLocal(source)
		if !known {
			response.BadRequest(w, fmt.Sprintf("unknown metadata source %q", source))
			return
		}

		body.Source = source

		if !isLocal {
			jobType = jobs.JobTypeMetadataRemote
		}
	}

	rows, failed, err := h.partitionExisting(r, ids)
	if err != nil {
		h.logger.Error("get archives by ids failed", "error", err)
		response.InternalError(w, "failed to load archives")

		return
	}

	succeeded := 0

	for _, row := range rows {
		// EnqueueOnce so re-submitting a selection that's still processing
		// doesn't stack duplicate work for the same archive.
		if err := h.queue.EnqueueOnce(r.Context(), jobType, jobs.MetadataPayload{
			ArchiveID: row.ID,
			Source:    body.Source,
		}); err != nil {
			h.logger.Error("enqueue metadata job failed", "archive_id", row.ID, "error", err)
			failed = append(failed, BulkFailure{ID: row.ID, Error: "failed to enqueue metadata job"})

			continue
		}

		succeeded++
	}

	response.JSON(w, http.StatusAccepted, BulkResult{
		Requested: len(ids),
		Succeeded: succeeded,
		Failed:    failed,
	})
}

// BulkUpdateArchivesRequest edits shared metadata across many archives.
//
// Deliberately narrower than the single-archive PATCH: title, summary and
// release_date are facts about one specific archive, so stamping a single
// value across a selection is almost always a mistake rather than an
// intent, and there's no undo. Language and category are here because
// they're shared classifications rather than per-archive facts - setting a
// selection to "japanese" is a normal thing to want.
//
// The two kinds of field behave differently, which is the important thing
// to know about this endpoint:
//
//   - Relations (artists/tags/parodies/circles/characters) MERGE. Submitted
//     values are added to whatever each archive already has, so each keeps
//     its own. That means this endpoint can only add: an empty array is a
//     no-op rather than "clear", and bulk removal isn't supported.
//   - Scalars (language/category) are SET, overwriting the archive's
//     current value - there's nothing sensible to merge into. Omit them to
//     leave them alone; note that null also means "leave alone" (see
//     UpdateArchive's coalesce), so sending "" is how you blank one.
type BulkUpdateArchivesRequest struct {
	IDs        []string `json:"ids"`
	Artists    []string `json:"artists"`
	Tags       []string `json:"tags"`
	Parodies   []string `json:"parodies"`
	Circles    []string `json:"circles"`
	Characters []string `json:"characters"`
	Language   *string  `json:"language"`
	Category   *string  `json:"category"`
}

// mergeNames unions an archive's existing relation values with the ones
// submitted, preserving order (existing first) and returning nil when the
// caller omitted the field entirely so ApplyMetadata leaves it alone.
//
// De-duplication is case-insensitive; note that submitted names are also
// snapped to the spelling already in the entity table beforehand (see
// canonicalizeNames), so this only has to handle the per-archive overlap.
// canonicalizeNames rewrites submitted names to the spelling already
// stored in table, so bulk-adding "Big Breasts" to a library that already
// uses "big breasts" reuses that row rather than creating a second,
// case-variant one that then behaves as an unrelated tag when filtering.
// Per-archive de-duplication alone can't prevent this: an archive that
// doesn't yet carry the tag has nothing to compare against.
func canonicalizeNames(r *http.Request, db *sql.DB, table string, names []string) ([]string, error) {
	if len(names) == 0 {
		return names, nil
	}

	canonical, err := database.CanonicalNames(r.Context(), db, table, names)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(names))
	for i, name := range names {
		if existing, ok := canonical[strings.ToLower(strings.TrimSpace(name))]; ok {
			out[i] = existing
			continue
		}

		out[i] = name
	}

	return out, nil
}

func mergeNames(existing, add []string) []string {
	if add == nil {
		return nil
	}

	merged := make([]string, 0, len(existing)+len(add))

	seen := make(map[string]bool, len(existing)+len(add))
	for _, list := range [][]string{existing, add} {
		for _, name := range list {
			name = strings.TrimSpace(name)

			key := strings.ToLower(name)
			if name == "" || seen[key] {
				continue
			}

			seen[key] = true

			merged = append(merged, name)
		}
	}

	return merged
}

// BulkUpdateArchives godoc
//
//	@Summary		Update shared metadata on archives in bulk
//	@Description	Relations (artists/tags/parodies/circles/characters) MERGE into what each archive already has rather than replacing it, so an archive keeps its own tags and gains the submitted ones; de-duplication is case-insensitive and submitted names are snapped to the spelling already in use. Because it only adds, an empty array is a no-op rather than "clear" and bulk removal is unsupported. Scalars (language/category) are SET, overwriting the current value; omit or null them to leave them alone, send "" to blank them. Title, summary and release_date are intentionally not editable here since they describe one specific archive - use PATCH /api/archives/{id} for those. Each archive is applied in its own transaction, so a failure part-way through leaves earlier archives updated and reports the rest in "failed".
//	@Tags			archives
//	@Accept			json
//	@Produce		json
//	@Param			body	BulkUpdateArchivesRequest	true	"Archives and relations to add"
//	@Success		200	{object}	BulkResult
//	@Failure		400	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/archives/bulk [patch]
func (h *ArchiveHandler) BulkUpdateArchives(w http.ResponseWriter, r *http.Request) {
	var body BulkUpdateArchivesRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ids, err := normalizeBulkIDs(body.IDs)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	if body.Artists == nil && body.Tags == nil && body.Parodies == nil &&
		body.Circles == nil && body.Characters == nil &&
		body.Language == nil && body.Category == nil {
		response.BadRequest(w, "at least one of artists, tags, parodies, circles, characters, language or category is required")
		return
	}

	rows, failed, err := h.partitionExisting(r, ids)
	if err != nil {
		h.logger.Error("get archives by ids failed", "error", err)
		response.InternalError(w, "failed to load archives")

		return
	}

	existingIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		existingIDs = append(existingIDs, row.ID)
	}

	// One bulk read of current relations rather than a lookup per archive -
	// the merge needs every archive's existing values.
	current, err := database.GetBulkArchiveMetadata(r.Context(), h.db, existingIDs)
	if err != nil {
		h.logger.Error("get bulk archive metadata failed", "error", err)
		response.InternalError(w, "failed to load current metadata")

		return
	}

	// Once per request rather than per archive - the mapping is global to
	// the entity table, not archive-specific.
	for _, field := range []struct {
		table string
		names *[]string
	}{
		{"artist", &body.Artists},
		{"tag", &body.Tags},
		{"parody", &body.Parodies},
		{"circle", &body.Circles},
		{"character", &body.Characters},
	} {
		canonical, err := canonicalizeNames(r, h.db, field.table, *field.names)
		if err != nil {
			h.logger.Error("canonicalize names failed", "table", field.table, "error", err)
			response.InternalError(w, "failed to resolve existing metadata names")

			return
		}

		*field.names = canonical
	}

	succeeded := 0

	for _, id := range existingIDs {
		existing := current[id]
		if existing == nil {
			existing = &metadata.Result{}
		}

		result := &metadata.Result{
			// Scalars are passed straight through: ApplyMetadata's
			// coalesce leaves a nil one untouched.
			Language:   body.Language,
			Category:   body.Category,
			Artists:    mergeNames(existing.Artists, body.Artists),
			Tags:       mergeNames(existing.Tags, body.Tags),
			Parodies:   mergeNames(existing.Parodies, body.Parodies),
			Circles:    mergeNames(existing.Circles, body.Circles),
			Characters: mergeNames(existing.Characters, body.Characters),
		}

		if err := metadata.ApplyMetadata(r.Context(), h.queries, h.db, id, result); err != nil {
			h.logger.Error("bulk update archive failed", "archive_id", id, "error", err)
			failed = append(failed, BulkFailure{ID: id, Error: "failed to update archive"})

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
