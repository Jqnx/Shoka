// Package handlers implements the HTTP handlers for Shoka's JSON API,
// exposing archive, library, metadata, job, and admin operations.
package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"Shoka/internal/maintenance"
	"log/slog"
	"net/http"
)

// AdminHandler exposes cross-library maintenance endpoints such as cover
// regeneration, perceptual hashing, and duplicate detection.
type AdminHandler struct {
	Log         *slog.Logger
	Queries     *sqlc.Queries
	Maintenance *maintenance.Manager
}

// NewAdminHandler returns an AdminHandler that runs whole-library maintenance
// through the manager and reads archives through queries.
func NewAdminHandler(queries *sqlc.Queries, maintenance *maintenance.Manager, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		Queries:     queries,
		Maintenance: maintenance,
		Log:         logger.With("handler", "admin"),
	}
}

// GenerateCovers godoc
//
//	@Summary		Enqueue cover generation for all archives
//	@Tags			admin
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/covers [post]
func (h *AdminHandler) GenerateCovers(w http.ResponseWriter, r *http.Request) {
	if err := h.Maintenance.RunTask(r.Context(), jobs.JobTypeCover); err != nil {
		h.Log.Error("run covers failed", "error", err)
		response.InternalError(w, "failed to enqueue cover generation")

		return
	}

	response.NoContent(w)
}

// GeneratePHashes godoc
//
//	@Summary		Enqueue perceptual-hash computation for all archives
//	@Description	Backfills the hashes GET /api/admin/duplicates compares. Safe to re-run - each job overwrites that archive's hashes in place, so this is also how you re-hash the library after a change to which pages get sampled.
//	@Tags			admin
//	@Success		204
//	@Failure		403	{object}	response.Error
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/phashes [post]
func (h *AdminHandler) GeneratePHashes(w http.ResponseWriter, r *http.Request) {
	if err := h.Maintenance.RunTask(r.Context(), jobs.JobTypePHash); err != nil {
		h.Log.Error("run phashes failed", "error", err)
		response.InternalError(w, "failed to enqueue hash generation")

		return
	}

	response.NoContent(w)
}

// GetDuplicates godoc
//
//	@Summary		Find likely-duplicate archives by perceptual hash
//	@Description	Cross-library maintenance action, like /admin/covers. Compares archives at four sample points (the cover plus 25/50/75% through the pages) and groups pairs where a majority of comparable points are within threshold - so a match has to hold across the body of the work, not just its cover. Groups are connected components (A close to B, B close to C groups all three even if A and C aren't directly close), so treat larger groups as progressively less certain. Archives the phash job hasn't processed yet are silently skipped. Note that positional sampling assumes the two archives' page counts roughly line up; a release padded with many extra pages can still slip through.
//	@Tags			admin
//	@Produce		json
//	@Param			threshold	query		int	false	"Max per-point Hamming distance (0-20) for a sample point to count as matching"	default(8)
//	@Success		200			{array}		DuplicateGroup
//	@Failure		403			{object}	response.Error
//	@Failure		500			{object}	response.Error
//	@Router			/api/admin/duplicates [get]
func (h *AdminHandler) GetDuplicates(w http.ResponseWriter, r *http.Request) {
	rows, err := h.Queries.GetAllArchivePHashes(r.Context())
	if err != nil {
		h.Log.Error("get archive phashes failed", "error", err)
		response.InternalError(w, "failed to get archive phashes")

		return
	}

	candidates := duplicateCandidates(rows)

	threshold := duplicateThreshold(r.URL.Query())
	components := duplicateComponents(candidates, threshold)

	groups := make([]DuplicateGroup, 0, len(components))

	for _, members := range components {
		groups = append(groups, duplicateGroupFor(members, candidates, threshold))
	}

	sortDuplicateGroupsBySize(groups)

	response.JSON(w, http.StatusOK, groups)
}
