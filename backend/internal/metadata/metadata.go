package metadata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"Shoka/internal/database/sqlc"
	"Shoka/internal/util"
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

	// IsLocal returns true for sources that read from the local filesystem
	// without network calls or rate limiting (e.g. ComicInfo, filename parsing).
	IsLocal() bool

	// Fetch attempts to retrieve metadata for the given archive.
	// Returns nil, nil if the source has no data for this archive — not an error.
	// Returns an error only if something genuinely went wrong (network failure, parse error).
	Fetch(ctx context.Context, input Input) (*Result, error)
}

type SearchableSource interface {
	Source
	Search(ctx context.Context, input Input) ([]*SearchResult, error)
}

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

type RemoteSource interface {
	SearchableSource
	FetchByID(ctx context.Context, id string) (*Result, error)
}

func ApplyMetadata(ctx context.Context, queries *sqlc.Queries, db *sql.DB, archiveID string, result *Result) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()
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

	// Relational fields
	if len(result.Artists) > 0 {
		if err := applyArtists(ctx, qtx, archiveID, result.Artists); err != nil {
			return err
		}
	}

	if len(result.Circles) > 0 {
		if err := applyCircles(ctx, qtx, archiveID, result.Circles); err != nil {
			return err
		}
	}

	if len(result.Tags) > 0 {
		if err := applyTags(ctx, qtx, archiveID, result.Tags); err != nil {
			return err
		}
	}

	if len(result.Parodies) > 0 {
		if err := applyParodies(ctx, qtx, archiveID, result.Parodies); err != nil {
			return err
		}
	}

	if len(result.Characters) > 0 {
		if err := applyCharacters(ctx, qtx, archiveID, result.Characters); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func applyArtists(ctx context.Context, q *sqlc.Queries, archiveID string, artists []string) error {
	artistsJson, _ := json.Marshal(artists)
	err := q.BulkAddArtists(ctx, artistsJson)
	if err != nil {
		return fmt.Errorf("upsert artists: %w", err)
	}

	rows, err := q.BulkGetArtists(ctx, artists)
	if err != nil {
		return fmt.Errorf("get artists: %w", err)
	}

	desiredIDs := make([]int64, len(rows))
	for i, row := range rows {
		desiredIDs[i] = row.ID
	}

	currentIDs, err := q.GetArchiveArtistIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get current artists: %w", err)
	}

	toAdd, toRemove := util.DiffIDs(currentIDs, desiredIDs)

	if len(toAdd) > 0 {
		toAddJson, _ := json.Marshal(toAdd)
		if err := q.BulkAddArchiveArtists(ctx, sqlc.BulkAddArchiveArtistsParams{
			ArchiveID: &archiveID,
			Artists:   toAddJson,
		}); err != nil {
			return fmt.Errorf("link artists: %w", err)
		}
	}

	if len(toRemove) > 0 {
		toRemoveJson, _ := json.Marshal(toRemove)
		if err := q.BulkRemoveArtistFromArchive(ctx, sqlc.BulkRemoveArtistFromArchiveParams{
			ArchiveID: &archiveID,
			Artists:   toRemoveJson,
		}); err != nil {
			return fmt.Errorf("unlink artists: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := q.IncrementArtistCount(ctx, toAdd); err != nil {
			return fmt.Errorf("increment artist counts: %w", err)
		}
	}
	if len(toRemove) > 0 {
		if err := q.DecrementArtistCount(ctx, toRemove); err != nil {
			return fmt.Errorf("decrement artist counts: %w", err)
		}
	}

	return nil
}

func applyCircles(ctx context.Context, q *sqlc.Queries, archiveID string, circles []string) error {
	circleJson, _ := json.Marshal(circles)
	err := q.BulkAddCircle(ctx, circleJson)
	if err != nil {
		return fmt.Errorf("upsert circles: %w", err)
	}

	rows, err := q.BulkGetCircle(ctx, circles)
	if err != nil {
		return fmt.Errorf("get circles: %w", err)
	}

	desiredIDs := make([]int64, len(rows))
	for i, row := range rows {
		desiredIDs[i] = row.ID
	}

	currentIDs, err := q.GetArchiveCircleIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get current circles: %w", err)
	}

	toAdd, toRemove := util.DiffIDs(currentIDs, desiredIDs)

	if len(toAdd) > 0 {
		toAddJson, _ := json.Marshal(toAdd)
		if err := q.BulkAddArchiveCircle(ctx, sqlc.BulkAddArchiveCircleParams{
			ArchiveID: &archiveID,
			Circle:    toAddJson,
		}); err != nil {
			return fmt.Errorf("link circles: %w", err)
		}
	}

	if len(toRemove) > 0 {
		toRemoveJson, _ := json.Marshal(toRemove)
		if err := q.BulkRemoveCircleFromArchive(ctx, sqlc.BulkRemoveCircleFromArchiveParams{
			ArchiveID: &archiveID,
			Circle:    toRemoveJson,
		}); err != nil {
			return fmt.Errorf("unlink circles: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := q.IncrementCircleCount(ctx, toAdd); err != nil {
			return fmt.Errorf("increment circle counts: %w", err)
		}
	}
	if len(toRemove) > 0 {
		if err := q.DecrementCircleCount(ctx, toRemove); err != nil {
			return fmt.Errorf("decrement circle counts: %w", err)
		}
	}

	return nil
}

func applyTags(ctx context.Context, q *sqlc.Queries, archiveID string, tags []string) error {
	tagsJson, _ := json.Marshal(tags)
	err := q.BulkAddTags(ctx, tagsJson)
	if err != nil {
		return fmt.Errorf("upsert tags: %w", err)
	}

	rows, err := q.BulkGetTags(ctx, tags)
	if err != nil {
		return fmt.Errorf("get tags: %w", err)
	}

	desiredIDs := make([]int64, len(rows))
	for i, row := range rows {
		desiredIDs[i] = row.ID
	}

	currentIDs, err := q.GetArchiveTagIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get current tags: %w", err)
	}

	toAdd, toRemove := util.DiffIDs(currentIDs, desiredIDs)

	if len(toAdd) > 0 {
		toAddJson, _ := json.Marshal(toAdd)
		if err := q.BulkAddArchiveTags(ctx, sqlc.BulkAddArchiveTagsParams{
			ArchiveID: &archiveID,
			Tags:      toAddJson,
		}); err != nil {
			return fmt.Errorf("link tags: %w", err)
		}
	}

	if len(toRemove) > 0 {
		toRemoveJson, _ := json.Marshal(toRemove)
		if err := q.BulkRemoveTagFromArchive(ctx, sqlc.BulkRemoveTagFromArchiveParams{
			ArchiveID: &archiveID,
			Tags:      toRemoveJson,
		}); err != nil {
			return fmt.Errorf("unlink tags: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := q.IncrementTagCount(ctx, toAdd); err != nil {
			return fmt.Errorf("increment tag counts: %w", err)
		}
	}
	if len(toRemove) > 0 {
		if err := q.DecrementTagCount(ctx, toRemove); err != nil {
			return fmt.Errorf("decrement tag counts: %w", err)
		}
	}

	return nil
}

func applyCharacters(ctx context.Context, q *sqlc.Queries, archiveID string, characters []string) error {
	charactersJson, _ := json.Marshal(characters)
	err := q.BulkAddCharacters(ctx, charactersJson)
	if err != nil {
		return fmt.Errorf("upsert characters: %w", err)
	}

	rows, err := q.BulkGetCharacters(ctx, characters)
	if err != nil {
		return fmt.Errorf("get characters: %w", err)
	}

	desiredIDs := make([]int64, len(rows))
	for i, row := range rows {
		desiredIDs[i] = row.ID
	}

	currentIDs, err := q.GetArchiveCharacterIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get current characters: %w", err)
	}

	toAdd, toRemove := util.DiffIDs(currentIDs, desiredIDs)

	if len(toAdd) > 0 {
		toAddJson, _ := json.Marshal(toAdd)
		if err := q.BulkAddArchiveCharacters(ctx, sqlc.BulkAddArchiveCharactersParams{
			ArchiveID:  &archiveID,
			Characters: toAddJson,
		}); err != nil {
			return fmt.Errorf("link characters: %w", err)
		}
	}

	if len(toRemove) > 0 {
		toRemoveJson, _ := json.Marshal(toRemove)
		if err := q.BulkRemoveCharactersFromArchive(ctx, sqlc.BulkRemoveCharactersFromArchiveParams{
			ArchiveID:  &archiveID,
			Characters: toRemoveJson,
		}); err != nil {
			return fmt.Errorf("unlink characters: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := q.IncrementCharacterCount(ctx, toAdd); err != nil {
			return fmt.Errorf("increment character counts: %w", err)
		}
	}
	if len(toRemove) > 0 {
		if err := q.DecrementCharacterCount(ctx, toRemove); err != nil {
			return fmt.Errorf("decrement character counts: %w", err)
		}
	}

	return nil
}

func applyParodies(ctx context.Context, q *sqlc.Queries, archiveID string, parodies []string) error {
	parodiesJson, _ := json.Marshal(parodies)
	err := q.BulkAddParodies(ctx, parodiesJson)
	if err != nil {
		return fmt.Errorf("upsert parodies: %w", err)
	}

	rows, err := q.BulkGetParodies(ctx, parodies)
	if err != nil {
		return fmt.Errorf("get parodies: %w", err)
	}

	desiredIDs := make([]int64, len(rows))
	for i, row := range rows {
		desiredIDs[i] = row.ID
	}

	currentIDs, err := q.GetArchiveParodyIDs(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get current parodies: %w", err)
	}

	toAdd, toRemove := util.DiffIDs(currentIDs, desiredIDs)

	if len(toAdd) > 0 {
		toAddJson, _ := json.Marshal(toAdd)
		if err := q.BulkAddArchiveParodies(ctx, sqlc.BulkAddArchiveParodiesParams{
			ArchiveID: &archiveID,
			Parodies:  toAddJson,
		}); err != nil {
			return fmt.Errorf("link parodies: %w", err)
		}
	}

	if len(toRemove) > 0 {
		toRemoveJson, _ := json.Marshal(toRemove)
		if err := q.BulkRemoveParodiesFromArchive(ctx, sqlc.BulkRemoveParodiesFromArchiveParams{
			ArchiveID: &archiveID,
			Parodies:  toRemoveJson,
		}); err != nil {
			return fmt.Errorf("unlink parodies: %w", err)
		}
	}

	if len(toAdd) > 0 {
		if err := q.IncrementParodyCount(ctx, toAdd); err != nil {
			return fmt.Errorf("increment parody counts: %w", err)
		}
	}
	if len(toRemove) > 0 {
		if err := q.DecrementParodyCount(ctx, toRemove); err != nil {
			return fmt.Errorf("decrement parody counts: %w", err)
		}
	}

	return nil
}
