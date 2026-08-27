package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/library/archive"
	"context"
	"fmt"
	"io"
	"log/slog"
)

const JobTypePHash = "phash"

type PHashPayload struct {
	FilePath  string `json:"file_path"`
	ArchiveID string `json:"archive_id"`
}

// samplePositions are the points through an archive that get hashed, as a
// fraction of its page count. The cover alone was too weak a signal for
// duplicate detection in both directions - unrelated works with similar
// cover art matched, and re-releases with a new cover didn't - so matching
// samples the body of the work too.
var samplePositions = [4]float64{0, 0.25, 0.50, 0.75}

// sampleIndices maps samplePositions onto concrete page indices for an
// archive of pageCount pages. Short archives can map several positions onto
// the same page (a 1-page archive maps all four to page 0); that's harmless,
// it just means those points carry the same information.
func sampleIndices(pageCount int) [4]int {
	var idx [4]int
	if pageCount <= 0 {
		return idx
	}

	for i, pos := range samplePositions {
		n := int(float64(pageCount) * pos)
		if n >= pageCount {
			n = pageCount - 1
		}

		idx[i] = n
	}

	return idx
}

// NewPHashHandler computes perceptual hashes at several points through an
// archive. Kept separate from the cover job (which also decodes page 0) so
// re-hashing the whole library after a change to the sampling/matching
// doesn't force every thumbnail to be regenerated along with it.
func NewPHashHandler(processor *image.Processor, queries *sqlc.Queries, log *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		var p PHashPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		a, err := archive.Open(p.FilePath)
		if err != nil {
			return fmt.Errorf("open archive: %w", err)
		}
		defer a.Close()

		pages, err := a.Pages()
		if err != nil {
			return fmt.Errorf("list pages: %w", err)
		}

		if len(pages) == 0 {
			log.Warn("archive has no pages, skipping phash", "archive_id", p.ArchiveID)
			return nil
		}

		// Several sample positions can resolve to the same page on short
		// archives - hash each distinct page once and reuse the result.
		indices := sampleIndices(len(pages))
		hashByIndex := make(map[int]uint64, len(indices))
		hashes := make([]*int64, len(indices))

		for i, pageIdx := range indices {
			if hash, done := hashByIndex[pageIdx]; done {
				signed := int64(hash)
				hashes[i] = &signed

				continue
			}

			hash, err := hashPage(processor, a, pages[pageIdx])
			if err != nil {
				// One unreadable page shouldn't discard the points that did
				// work - the comparison tolerates missing sample points.
				log.Warn("hash page failed",
					"archive_id", p.ArchiveID, "page", pageIdx, "error", err)

				continue
			}

			hashByIndex[pageIdx] = hash
			signed := int64(hash)
			hashes[i] = &signed
		}

		if err := queries.UpdateArchivePHashes(ctx, sqlc.UpdateArchivePHashesParams{
			PhashP0:  hashes[0],
			PhashP25: hashes[1],
			PhashP50: hashes[2],
			PhashP75: hashes[3],
			ID:       p.ArchiveID,
		}); err != nil {
			return fmt.Errorf("store phashes: %w", err)
		}

		return nil
	}
}

func hashPage(processor *image.Processor, a archive.Archive, page archive.Page) (uint64, error) {
	r, err := a.Extract(page)
	if err != nil {
		return 0, fmt.Errorf("extract page: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return 0, fmt.Errorf("read page: %w", err)
	}

	return processor.ComputePHash(data)
}
