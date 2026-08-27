package jobs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

const JobTypeScan = "scan"

// ScanPayload always targets a single library — the app has no "scan
// everything" job; scanning multiple libraries just means enqueueing one
// ScanPayload per library and letting the worker pool parallelize them.
type ScanPayload struct {
	LibraryID string
}

// Scannable is implemented by the library manager. It resolves the library
// by ID and dispatches to the Scanner implementation for its type. Defined
// here (rather than depending on internal/library) to avoid an import cycle,
// since internal/library already imports internal/jobs for job payloads.
type Scannable interface {
	ScanLibrary(ctx context.Context, libraryID string) error
}

func NewScanHandler(scanner Scannable, log *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		var p ScanPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		if p.LibraryID == "" {
			return errors.New("scan job missing library_id")
		}

		return scanner.ScanLibrary(ctx, p.LibraryID)
	}
}
