// Package database provides access to the SQLite database: connection setup,
// embedded migrations, generated sqlc queries, and hand-written helpers.
package database

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/metadata"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Session is the authenticated caller resolved from a Better Auth session
// token. Role and Banned come from the frontend-owned admin plugin columns.
type Session struct {
	UserID string
	// Role is "" when the column is NULL. That is the same thing the
	// permission matrix does with any unrecognised role - grants nothing - so
	// the empty string is the honest default rather than a substitute role.
	Role   string
	Banned bool
}

// GetSession looks up a Better Auth session token and returns the owning user.
// Returns (nil, nil) if the token does not exist or has expired.
func GetSession(ctx context.Context, db *sql.DB, token string) (*Session, error) {
	var (
		session   Session
		expiresAt int64
	)

	err := db.QueryRowContext(ctx, `
        SELECT s.user_id, s.expires_at,
               COALESCE(u.role, ''), COALESCE(u.banned, 0)
        FROM session s
        JOIN user u ON u.id = s.user_id
        WHERE s.token = ?
        LIMIT 1
    `, token).Scan(&session.UserID, &expiresAt, &session.Role, &session.Banned)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("session lookup: %w", err)
	}

	if time.Now().After(time.UnixMilli(expiresAt)) {
		return nil, nil
	}

	return &session, nil
}

// GetArchiveMetadata loads an archive and its complete metadata: the
// denormalized columns of the archive row plus the names of every related
// artist, tag, parody, circle and character, and its source URLs. It backs
// the archive metadata endpoint.
func GetArchiveMetadata(ctx context.Context, q *sqlc.Queries, archiveID string) (*metadata.Result, error) {
	archive, err := q.GetArchiveByID(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get archive: %w", err)
	}

	result := archiveMetadataFields(archive)

	if err := loadArchiveRelations(ctx, q, archiveID, result); err != nil {
		return nil, err
	}

	return result, nil
}

// archiveMetadataFields copies the scalar archive columns into a Result,
// leaving the relation slices for loadArchiveRelations. A nil column stays
// nil in the Result so an absent value is not reported as a zero one.
func archiveMetadataFields(archive sqlc.Archive) *metadata.Result {
	result := &metadata.Result{}

	if archive.Title != "" {
		result.Title = &archive.Title
	}

	if archive.Summary != nil {
		result.Summary = archive.Summary
	}

	if archive.Language != nil {
		result.Language = archive.Language
	}

	if archive.Category != nil {
		result.Category = archive.Category
	}

	if archive.PageCount > 0 {
		result.PageCount = &archive.PageCount
	}

	if archive.ReleaseDate != nil {
		result.ReleaseDate = archive.ReleaseDate
	}

	return result
}

// loadArchiveRelations fills the relation slices of result, stopping at the
// first query that fails. The load order is the order the queries always ran
// in, so callers still see the first (not an arbitrary) failure.
func loadArchiveRelations(ctx context.Context, q *sqlc.Queries, archiveID string, result *metadata.Result) error {
	var err error

	result.Artists, err = loadRelationNames(ctx, q.GetArchiveArtists, archiveID, "artists", func(a sqlc.Artist) string { return a.Name })
	if err != nil {
		return err
	}

	result.Tags, err = loadRelationNames(ctx, q.GetArchiveTag, archiveID, "tags", func(t sqlc.Tag) string { return t.Name })
	if err != nil {
		return err
	}

	result.Parodies, err = loadRelationNames(ctx, q.GetArchiveParody, archiveID, "parodies", func(p sqlc.Parody) string { return p.Name })
	if err != nil {
		return err
	}

	result.Circles, err = loadRelationNames(ctx, q.GetArchiveCircle, archiveID, "circles", func(c sqlc.Circle) string { return c.Name })
	if err != nil {
		return err
	}

	result.Characters, err = loadRelationNames(ctx, q.GetArchiveCharacters, archiveID, "characters", func(c sqlc.Character) string { return c.Name })
	if err != nil {
		return err
	}

	urls, err := q.GetArchiveUrls(ctx, archiveID)
	if err != nil {
		return fmt.Errorf("get urls: %w", err)
	}

	result.URLs = urls

	return nil
}

// loadRelationNames runs one archive-relation query and maps each row's Name
// into a plain string slice. The slice stays nil when there are no rows, so
// an archive without (say) parodies serialises the same as it always did.
func loadRelationNames[T any](
	ctx context.Context,
	query func(context.Context, string) ([]T, error),
	archiveID, label string,
	name func(T) string,
) ([]string, error) {
	rows, err := query(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", label, err)
	}

	var names []string

	for _, row := range rows {
		names = append(names, name(row))
	}

	return names, nil
}
