// Package jobs implements the background job system: the payload types and
// handlers for each job type, the queue that persists them, and the worker
// pool that claims and executes them.
package jobs

import (
	"Shoka/internal/image"
	"Shoka/internal/library/archive"
	"context"
	"fmt"
	"log/slog"
)

// JobTypeCover is the job type that regenerates an archive's cover thumbnail.
const JobTypeCover = "cover"

// CoverPayload selects the archive whose cover should be generated and the
// page to use as the cover.
type CoverPayload struct {
	FilePath  string `json:"file_path"`
	ArchiveID string `json:"archive_id"`
	// PageIndex is the 0-based page to use as the cover (archive.cover_page).
	// Zero value is page 0 - the historical, only behaviour before covers
	// became choosable - so every pre-existing enqueue site (scan, bulk
	// regenerate) keeps working unchanged.
	PageIndex int `json:"page_index"`
}

// NewCoverHandler returns a Handler that renders the chosen page as the
// archive's cover thumbnail, falling back to page 0 when the requested index
// is out of range.
func NewCoverHandler(processor *image.Processor, _ *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		var p CoverPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
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

		if len(pages) == 0 {
			return fmt.Errorf("archive %s has no pages", p.ArchiveID)
		}

		index := p.PageIndex
		if index < 0 || index >= len(pages) {
			index = 0
		}

		page := pages[index]

		r, err := a.Extract(page)
		if err != nil {
			return fmt.Errorf("extract page %d: %w", page.Index, err)
		}

		if err := processor.GenerateThumbnail(ctx, p.ArchiveID, page.Index, r); err != nil {
			_ = r.Close()
			return err
		}

		_ = r.Close()

		return nil
	}
}
