package database

import (
	"Shoka/internal/auth"
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
               COALESCE(u.role, ?), COALESCE(u.banned, 0)
        FROM session s
        JOIN user u ON u.id = s.user_id
        WHERE s.token = ?
        LIMIT 1
    `, auth.RoleUser, token).Scan(&session.UserID, &expiresAt, &session.Role, &session.Banned)

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

func GetArchiveMetadata(ctx context.Context, q *sqlc.Queries, archiveID string) (*metadata.Result, error) {
	archive, err := q.GetArchiveByID(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get archive: %w", err)
	}

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

	artists, err := q.GetArchiveArtists(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get artists: %w", err)
	}

	for _, a := range artists {
		result.Artists = append(result.Artists, a.Name)
	}

	tags, err := q.GetArchiveTag(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}

	for _, t := range tags {
		result.Tags = append(result.Tags, t.Name)
	}

	parodies, err := q.GetArchiveParody(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get parodies: %w", err)
	}

	for _, p := range parodies {
		result.Parodies = append(result.Parodies, p.Name)
	}

	circles, err := q.GetArchiveCircle(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get circles: %w", err)
	}

	for _, c := range circles {
		result.Circles = append(result.Circles, c.Name)
	}

	characters, err := q.GetArchiveCharacters(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get characters: %w", err)
	}

	for _, c := range characters {
		result.Characters = append(result.Characters, c.Name)
	}

	urls, err := q.GetArchiveUrls(ctx, archiveID)
	if err != nil {
		return nil, fmt.Errorf("get urls: %w", err)
	}

	result.URLs = urls

	return result, nil
}
