package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/library/archive"
	"context"
	"fmt"
	"log/slog"
)

// JobTypeThumbnail is the job type that generates an archive's per-page
// thumbnails.
const JobTypeThumbnail = "thumbnail"

// ThumbnailPayload identifies the archive and backing file to thumbnail.
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
func NewThumbnailHandler(processor *image.Processor, broadcaster *events.ThumbnailBroadcaster, queries *sqlc.Queries, log *slog.Logger) Handler {
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

		// Outer return-path cleanup, so the skip below and a mid-loop abort
		// are both covered: a delete that lands between an existence check and
		// its thumbnail write leaves that write behind, and this reclaims it.
		defer func() {
			if gone, err := archiveGone(ctx, queries, p.ArchiveID); err == nil && gone {
				_ = processor.EvictAll(p.ArchiveID)
			}
		}()

		gone, err := archiveGone(ctx, queries, p.ArchiveID)
		if err != nil {
			return fmt.Errorf("check archive %s: %w", p.ArchiveID, err)
		}

		if gone {
			log.Info("archive deleted, skipping thumbnails", "archive_id", p.ArchiveID)
			broadcaster.Publish(p.ArchiveID, events.ThumbnailEvent{Done: true})

			return nil
		}

		a, err := archive.Open(p.FilePath)
		if err != nil {
			return fmt.Errorf("open archive: %w", err)
		}
		defer func() { _ = a.Close() }()

		pages, err := a.Pages()
		if err != nil {
			return fmt.Errorf("list pages: %w", err)
		}

		generated, err := thumbnailPages(ctx, thumbDeps{
			processor:   processor,
			queries:     queries,
			broadcaster: broadcaster,
			log:         log,
		}, p.ArchiveID, a, pages)
		if err != nil {
			return err
		}

		log.Info("thumbnails generated", "archive_id", p.ArchiveID, "generated", generated, "total_pages", len(pages))
		broadcaster.Publish(p.ArchiveID, events.ThumbnailEvent{Done: true})

		return nil
	}
}

// thumbDeps is what the per-page loop needs: the processor that owns the cache
// layout, the provider of the archive rows, and the SSE broadcaster.
type thumbDeps struct {
	processor   *image.Processor
	queries     *sqlc.Queries
	broadcaster *events.ThumbnailBroadcaster
	log         *slog.Logger
}

// thumbnailPages renders every page whose thumbnail is missing and reports how
// many it generated. It returns early, without error, once the archive has been
// deleted: that delete evicted this archive's cache directory, so the pages
// left would only recreate what was just removed. Publishing the archive's
// terminal Done event is the caller's job, so both endings send exactly one.
func thumbnailPages(ctx context.Context, deps thumbDeps, archiveID string, a archive.Archive, pages []archive.Page) (int, error) {
	generated := 0

	for _, page := range pages {
		gone, err := archiveGone(ctx, deps.queries, archiveID)
		if err != nil {
			return generated, err
		}

		if gone {
			deps.log.Info("archive deleted, stopping thumbnail generation",
				"archive_id", archiveID, "page", page.Index)

			return generated, nil
		}

		// Skip pages that already have a thumbnail (e.g. page 0, which the
		// cover job usually generates first) — checking before extracting
		// avoids paying the decompression cost again.
		if deps.processor.ThumbPath(archiveID, page.Index) != "" {
			deps.broadcaster.Publish(archiveID, events.ThumbnailEvent{Index: page.Index})

			continue
		}

		r, err := a.Extract(page)
		if err != nil {
			return generated, fmt.Errorf("extract page %d: %w", page.Index, err)
		}

		err = deps.processor.GenerateThumbnail(ctx, archiveID, page.Index, r)
		_ = r.Close()

		if err != nil {
			return generated, err
		}

		generated++

		deps.broadcaster.Publish(archiveID, events.ThumbnailEvent{Index: page.Index})
	}

	return generated, nil
}
