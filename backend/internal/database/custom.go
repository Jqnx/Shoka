package database

import (
	"context"
	"fmt"

	"Shoka/internal/database/sqlc"
	"Shoka/internal/metadata"
)

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

	return result, nil
}
