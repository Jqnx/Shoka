package metadata

import (
	"context"
	"time"
)

// Result holds all metadata that can be collected for an archive.
type Result struct {
	Title       *string
	Summary     *string
	Language    *string
	Category    *string
	ReleaseDate *time.Time
	PageCount   *int64
	Artists     []string
	Tags        []string
	Parodies    []string
	Circles     []string
	Characters  []string
}

// Input is what every source receives to identify the archive.
type Input struct {
	ArchiveID string
	FilePath  string
	Title     string // filename-derived title, useful for search-based sources
}

// SourceConfig wraps a source with its enabled state.
type SourceConfig struct {
	Source  Source
	Enabled bool
}

// Source is the interface every metadata provider must implement.
type Source interface {
	// Name returns a human-readable identifier used in logs and admin UI.
	Name() string

	// Priority determines the order sources are tried.
	// Lower number = higher priority. File-based sources should have lower
	// numbers than API sources since they are more authoritative.
	Priority() int

	// Fetch attempts to retrieve metadata for the given archive.
	// Returns nil, nil if the source has no data for this archive — not an error.
	// Returns an error only if something genuinely went wrong (network failure, parse error).
	Fetch(ctx context.Context, input Input) (*Result, error)
}
