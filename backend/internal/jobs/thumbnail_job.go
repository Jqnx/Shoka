package jobs

import (
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/library/archive"
	"context"
	"fmt"
	"log/slog"
)

const JobTypeThumbnail = "thumbnail"

type ThumbnailPayload struct {
	FilePath  string `json:"file_path"`
	ArchiveID string `json:"archive_id"`
}

// NewThumbnailHandler generates per-page thumbnails for an archive.
// broadcaster is notified as each page becomes ready (including pages
// skipped because they already exist) and once more when the whole
// archive reaches a terminal state, so SSE subscribers (see the
// /thumbnails/events endpoint) can reflect progress live.
//
// A failure is only reported as terminal (Done+Error) once this was the
// job's last allowed attempt (job.Attempts >= job.MaxAttempts) — the queue
// retries failed jobs with backoff, so an error on an earlier attempt
// isn't "done," it's "about to be retried," and subscribers shouldn't be
// told otherwise.
func NewThumbnailHandler(processor *image.Processor, broadcaster *events.ThumbnailBroadcaster, log *slog.Logger) Handler {
	log = log.With("component", "thumbnail_job")

	return func(ctx context.Context, job *Job) (err error) {
		var p ThumbnailPayload
		if err = job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		defer func() {
			if err != nil && job.Attempts >= job.MaxAttempts {
				broadcaster.Publish(p.ArchiveID, events.ThumbnailEvent{Done: true, Error: err.Error()})
			}
		}()

		a, err := archive.Open(p.FilePath)
		if err != nil {
			return fmt.Errorf("open archive: %w", err)
		}
		defer a.Close()

		pages, err := a.Pages()
		if err != nil {
			return fmt.Errorf("list pages: %w", err)
		}

		generated := 0

		for _, page := range pages {
			// Skip pages that already have a thumbnail (e.g. page 0, which
			// the cover job usually generates first) — checking before
			// extracting avoids paying the decompression cost again.
			if processor.ThumbPath(p.ArchiveID, page.Index) != "" {
				broadcaster.Publish(p.ArchiveID, events.ThumbnailEvent{Index: page.Index})
				continue
			}

			r, err := a.Extract(page)
			if err != nil {
				return fmt.Errorf("extract page %d: %w", page.Index, err)
			}

			err = processor.GenerateThumbnail(ctx, p.ArchiveID, page.Index, r)
			r.Close()

			if err != nil {
				return err
			}

			generated++

			broadcaster.Publish(p.ArchiveID, events.ThumbnailEvent{Index: page.Index})
		}

		log.Info("thumbnails generated", "archive_id", p.ArchiveID, "generated", generated, "total_pages", len(pages))
		broadcaster.Publish(p.ArchiveID, events.ThumbnailEvent{Done: true})

		return nil
	}
}
