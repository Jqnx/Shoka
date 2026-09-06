package jobs

import (
	"Shoka/internal/image"
	"Shoka/internal/library/archive"
	"context"
	"fmt"
	"log/slog"
)

const JobTypeCover = "cover"

type CoverPayload struct {
	FilePath  string `json:"file_path"`
	ArchiveID string `json:"archive_id"`
	// PageIndex is the 0-based page to use as the cover (archive.cover_page).
	// Zero value is page 0 - the historical, only behaviour before covers
	// became choosable - so every pre-existing enqueue site (scan, bulk
	// regenerate) keeps working unchanged.
	PageIndex int `json:"page_index"`
}

func NewCoverHandler(processor *image.Processor, log *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		var p CoverPayload
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
			r.Close()
			return err
		}

		r.Close()

		return nil
	}
}
