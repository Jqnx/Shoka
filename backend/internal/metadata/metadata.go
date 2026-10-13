// Package metadata defines the metadata sources that enrich an archive
// (ComicInfo, filename parsing, gallery APIs) plus the pipeline that runs
// them by priority and merges their results into the archive's own tables.
package metadata

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Result holds all metadata that can be collected for an archive.
type Result struct {
	Title       *string    `json:"title"`
	Summary     *string    `json:"summary"`
	Language    *string    `json:"language"`
	Category    *string    `json:"category"`
	ReleaseDate *time.Time `json:"release_date"`
	PageCount   *int64     `json:"page_count"`
	Artists     []string   `json:"artists"`
	Tags        []string   `json:"tags"`
	Parodies    []string   `json:"parodies"`
	Circles     []string   `json:"circles"`
	Characters  []string   `json:"characters"`
	// URLs are the archive's source links (nhentai/e-hentai galleries,
	// ComicInfo <Web>, hand-pasted). Unlike every other slice here they
	// union across pipeline sources rather than first-wins — see merge().
	URLs []string `json:"urls"`
}

// Input is what every source receives to identify the archive.
type Input struct {
	ArchiveID string
	LibraryID string
	FilePath  string
	Title     string // filename-derived title, useful for search-based sources

	// SourceConfig holds the resolved per-library settings for whichever
	// source is currently being invoked. The Pipeline populates this field
	// fresh for each source before calling Fetch/Search, since each source
	// has its own settings row in the library_source table.
	SourceConfig SourceSettings
}

// SourceSettings is a source's per-library configuration, resolved from the
// library_source table. Not every source uses every field (e.g. ComicInfo
// uses none of them, filename uses only the blocklists).
type SourceSettings struct {
	Enabled           bool
	Cookies           string
	APIKey            string
	MagazineBlocklist []string
	MiscBlocklist     []string
}

// Source is the interface every metadata provider must implement.
type Source interface {
	// Name returns a human-readable identifier used in logs and admin UI.
	Name() string

	// Priority determines the order sources are tried.
	// Lower number = higher priority. File-based sources should have lower
	// numbers than API sources since they are more authoritative.
	Priority() int

	// IsLocal returns true for sources that read from the local filesystem
	// without network calls or rate limiting (e.g. ComicInfo, filename parsing).
	IsLocal() bool

	// Fetch attempts to retrieve metadata for the given archive.
	// Returns nil, nil if the source has no data for this archive — not an error.
	// Returns an error only if something genuinely went wrong (network failure, parse error).
	Fetch(ctx context.Context, input Input) (*Result, error)
}

// SearchableSource is a Source that can also be searched by title, returning
// candidate galleries for the user to pick from manually.
type SearchableSource interface {
	Source
	Search(ctx context.Context, input Input) ([]*SearchResult, error)
}

// SearchResult is one candidate gallery returned by a source's Search. It
// carries only the fields needed to render a picker row; Full stays nil until
// the user selects the result and FetchByID populates it.
type SearchResult struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	CoverURL   string     `json:"cover_url"`
	Language   string     `json:"language"`
	Tags       []string   `json:"tags"`
	PageCount  int        `json:"page_count"`
	UpdateDate *time.Time `json:"updated_at"`
	// Full is populated lazily when the user selects this result.
	// Nil until FetchByID is called.
	Full *Result `json:"full"`
}

// RemoteSource is a SearchableSource that can fetch the full metadata for a
// single gallery identified by the source-specific ID the user picked from
// search results.
type RemoteSource interface {
	SearchableSource
	// FetchByID fetches full metadata for a source-specific ID chosen by the
	// user from search results. input carries the resolved SourceConfig for
	// this source (cookies/API key) since fetching may need authentication.
	FetchByID(ctx context.Context, input Input, id string) (*Result, error)
}

// ApplyMetadata writes result into the archive's metadata tables in a single
// transaction: the flat archive columns are overwritten outright, while each
// relational field (artists, circles, tags, parodies, characters, and urls)
// is reconciled against the archive's current links.
//
// Relational fields are checked with != nil rather than len() > 0
// deliberately: a non-nil empty slice clears the relation entirely (e.g. a
// manual edit removing all tags), while a nil slice — the zero value, what
// pipeline sources leave a field at when they have nothing to contribute —
// means "don't touch this".
func ApplyMetadata(ctx context.Context, queries *sqlc.Queries, db *sql.DB, archiveID string, result *Result) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	qtx := queries.WithTx(tx)

	// Archive Details
	err = qtx.UpdateArchive(ctx, sqlc.UpdateArchiveParams{
		ID:          archiveID,
		Title:       result.Title,
		Summary:     result.Summary,
		Language:    result.Language,
		Category:    result.Category,
		PageCount:   result.PageCount,
		ReleaseDate: result.ReleaseDate,
	})
	if err != nil {
		return fmt.Errorf("update archive metadata: %w", err)
	}

	// Relational fields. Checking != nil rather than len() > 0 is
	// deliberate: it lets a caller clear a relation entirely by passing a
	// non-nil empty slice (e.g. a manual edit removing all tags), while a
	// nil slice (the zero value — what pipeline sources leave a field at
	// when they have nothing to contribute) still means "don't touch this."
	relations := []struct {
		names []string
		rel   relationSpec
	}{
		{result.Artists, artistsRelation},
		{result.Circles, circlesRelation},
		{result.Tags, tagsRelation},
		{result.Parodies, parodiesRelation},
		{result.Characters, charactersRelation},
	}

	for _, r := range relations {
		if r.names == nil {
			continue
		}

		if err := applyRelation(ctx, qtx, archiveID, r.names, r.rel); err != nil {
			return err
		}
	}

	if result.URLs != nil {
		if err := applyURLs(ctx, qtx, archiveID, result.URLs); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
