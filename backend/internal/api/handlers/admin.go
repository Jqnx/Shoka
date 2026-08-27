package handlers

import (
	"Shoka/internal/api/response"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/jobs"
	"log/slog"
	"math/bits"
	"net/http"
	"sort"
	"strconv"
)

type AdminHandler struct {
	Log     *slog.Logger
	Queries *sqlc.Queries
	Queue   *jobs.Queue
}

func NewAdminHandler(queries *sqlc.Queries, queue *jobs.Queue, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		Queries: queries,
		Queue:   queue,
		Log:     logger.With("handler", "admin"),
	}
}

// GenerateCovers godoc
//
//	@Summary		Enqueue cover generation for all archives
//	@Tags			admin
//	@Success		204
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/covers [post]
func (h *AdminHandler) GenerateCovers(w http.ResponseWriter, r *http.Request) {
	archives, err := h.Queries.GetAllArchiveFilePaths(r.Context())
	if err != nil {
		h.Log.Error("get archives failed", "error", err)
		response.InternalError(w, "failed to get archives")

		return
	}

	for _, a := range archives {
		if err := h.Queue.Enqueue(r.Context(), jobs.JobTypeCover, jobs.CoverPayload{
			ArchiveID: a.ID,
			FilePath:  a.FilePath,
		}); err != nil {
			h.Log.Error("enqueue cover job failed", "archive_id", a.ID, "error", err)
		}
	}

	response.NoContent(w)
}

// GeneratePHashes godoc
//
//	@Summary		Enqueue perceptual-hash computation for all archives
//	@Description	Backfills the hashes GET /api/admin/duplicates compares. Safe to re-run - each job overwrites that archive's hashes in place, so this is also how you re-hash the library after a change to which pages get sampled.
//	@Tags			admin
//	@Success		204
//	@Failure		500	{object}	response.Error
//	@Router			/api/admin/phashes [post]
func (h *AdminHandler) GeneratePHashes(w http.ResponseWriter, r *http.Request) {
	archives, err := h.Queries.GetAllArchiveFilePaths(r.Context())
	if err != nil {
		h.Log.Error("get archives failed", "error", err)
		response.InternalError(w, "failed to get archives")

		return
	}

	for _, a := range archives {
		if err := h.Queue.Enqueue(r.Context(), jobs.JobTypePHash, jobs.PHashPayload{
			ArchiveID: a.ID,
			FilePath:  a.FilePath,
		}); err != nil {
			h.Log.Error("enqueue phash job failed", "archive_id", a.ID, "error", err)
		}
	}

	response.NoContent(w)
}

// samplePointLabels name the sample points in GetAllArchivePHashes column
// order, for the per-point distance breakdown in the response.
var samplePointLabels = [4]string{"p0", "p25", "p50", "p75"}

type DuplicateArchive struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	LibraryID string `json:"library_id"`
	// Distance is the mean per-point Hamming distance to whichever *other*
	// member of the group this archive matches most closely - lower means
	// more visually similar. In a chained group (A close to B, B close to C)
	// this reports the real match, not the distance to an arbitrary anchor.
	Distance int `json:"distance"`
	// Distances breaks that down per sample point (p0 = cover, p25/p50/p75 =
	// that fraction through the archive), against that same closest member.
	// Points either archive hasn't hashed are omitted rather than reported
	// as 0.
	Distances map[string]int `json:"distances"`
}

type DuplicateGroup struct {
	Archives []DuplicateArchive `json:"archives"`
}

// comparePHashes returns the mean Hamming distance across the sample points
// both archives have hashed, how many points were comparable, and whether
// the pair counts as a match.
//
// A match needs a *majority* of comparable points within threshold rather
// than all of them: page counts rarely line up exactly between two releases
// of the same work (bonus pages, a title page, a different scan), so one
// sample point landing on misaligned content is expected and shouldn't veto
// an otherwise strong match. Requiring a majority still rules out the
// cover-coincidence false positives that a cover-only comparison produced.
func comparePHashes(a, b [4]*int64, threshold int) (mean, compared int, matched bool) {
	var total, within int

	for i := range a {
		if a[i] == nil || b[i] == nil {
			continue
		}

		//nolint:gosec // bit reinterpretation of a 64-bit perceptual hash stored as signed int64, not an arithmetic conversion
		d := bits.OnesCount64(uint64(*a[i]) ^ uint64(*b[i]))
		total += d
		compared++

		if d <= threshold {
			within++
		}
	}

	if compared == 0 {
		return 0, 0, false
	}
	// Two comparable points minimum, so a pair can never match on the cover
	// alone - that's the weak signal this replaced.
	return total / compared, compared, compared >= 2 && within*2 > compared
}

// GetDuplicates godoc
//
//	@Summary		Find likely-duplicate archives by perceptual hash
//	@Description	Cross-library maintenance action, like /admin/covers. Compares archives at four sample points (the cover plus 25/50/75% through the pages) and groups pairs where a majority of comparable points are within threshold - so a match has to hold across the body of the work, not just its cover. Groups are connected components (A close to B, B close to C groups all three even if A and C aren't directly close), so treat larger groups as progressively less certain. Archives the phash job hasn't processed yet are silently skipped. Note that positional sampling assumes the two archives' page counts roughly line up; a release padded with many extra pages can still slip through.
//	@Tags			admin
//	@Produce		json
//	@Param			threshold	query		int	false	"Max per-point Hamming distance (0-20) for a sample point to count as matching"	default(8)
//	@Success		200			{array}		DuplicateGroup
//	@Failure		500			{object}	response.Error
//	@Router			/api/admin/duplicates [get]
func (h *AdminHandler) GetDuplicates(w http.ResponseWriter, r *http.Request) {
	threshold := 8

	if v := r.URL.Query().Get("threshold"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 20 {
			threshold = n
		}
	}

	rows, err := h.Queries.GetAllArchivePHashes(r.Context())
	if err != nil {
		h.Log.Error("get archive phashes failed", "error", err)
		response.InternalError(w, "failed to get archive phashes")

		return
	}

	type candidate struct {
		id, title, libraryID string
		hashes               [4]*int64
	}

	candidates := make([]candidate, 0, len(rows))
	for _, row := range rows {
		candidates = append(candidates, candidate{
			id:        row.ID,
			title:     row.Title,
			libraryID: row.LibraryID,
			hashes:    [4]*int64{row.PhashP0, row.PhashP25, row.PhashP50, row.PhashP75},
		})
	}

	// Union-find over every pair within threshold, an O(n^2) scan - fine at
	// this app's scale (a personal library, not millions of archives) and
	// far simpler than an approximate-nearest-neighbor index.
	parent := make([]int, len(candidates))
	for i := range parent {
		parent[i] = i
	}

	find := func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}

		return x
	}

	union := func(a, b int) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}

	for i := range candidates {
		for j := i + 1; j < len(candidates); j++ {
			if _, _, matched := comparePHashes(candidates[i].hashes, candidates[j].hashes, threshold); matched {
				union(i, j)
			}
		}
	}

	membersByRoot := make(map[int][]int)

	for i := range candidates {
		root := find(i)
		membersByRoot[root] = append(membersByRoot[root], i)
	}

	groups := make([]DuplicateGroup, 0, len(membersByRoot))
	for _, members := range membersByRoot {
		if len(members) < 2 {
			continue
		}

		archives := make([]DuplicateArchive, 0, len(members))
		for _, idx := range members {
			c := candidates[idx]

			// Compare against whichever other group member c is actually
			// closest to, rather than a fixed anchor: union-find chains
			// A-B and B-C into one group even when A and C are a poor
			// direct match, and reporting c's distance to that anchor
			// would look like a false positive.
			mean := 0
			nearest := -1

			for _, otherIdx := range members {
				if otherIdx == idx {
					continue
				}

				m, compared, _ := comparePHashes(c.hashes, candidates[otherIdx].hashes, threshold)
				if compared > 0 && (nearest == -1 || m < mean) {
					mean = m
					nearest = otherIdx
				}
			}

			distances := make(map[string]int, len(samplePointLabels))
			if nearest != -1 {
				other := candidates[nearest]
				for k, label := range samplePointLabels {
					if c.hashes[k] == nil || other.hashes[k] == nil {
						continue
					}

					//nolint:gosec // bit reinterpretation of a 64-bit perceptual hash stored as signed int64, not an arithmetic conversion
					distances[label] = bits.OnesCount64(uint64(*c.hashes[k]) ^ uint64(*other.hashes[k]))
				}
			}

			archives = append(archives, DuplicateArchive{
				ID:        c.id,
				Title:     c.title,
				LibraryID: c.libraryID,
				Distance:  mean,
				Distances: distances,
			})
		}

		sort.Slice(archives, func(i, j int) bool { return archives[i].Distance < archives[j].Distance })
		groups = append(groups, DuplicateGroup{Archives: archives})
	}

	sort.Slice(groups, func(i, j int) bool {
		// Bigger groups (more copies of the same thing) and tighter matches
		// are the most actionable, so surface those first.
		if len(groups[i].Archives) != len(groups[j].Archives) {
			return len(groups[i].Archives) > len(groups[j].Archives)
		}

		return groups[i].Archives[1].Distance < groups[j].Archives[1].Distance
	})

	response.JSON(w, http.StatusOK, groups)
}
