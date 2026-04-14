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
		page := pages[0]

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
