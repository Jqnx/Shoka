package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"Shoka/internal/database/sqlc"
	"Shoka/internal/metadata"
)

const (
	JobTypeMetadata       = "metadata"
	JobTypeMetadataRemote = "metadata/remote"
)

type MetadataPayload struct {
	ArchiveID string
}

// NewMetadataHandler runs local sources only, applies the result, then enqueues
// a remote job to fill any remaining gaps via network sources.
func NewMetadataHandler(pipeline *metadata.Pipeline, queries *sqlc.Queries, db *sql.DB, queue *Queue, logger *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		var p MetadataPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		archive, err := queries.GetArchiveByID(ctx, p.ArchiveID)
		if err != nil {
			return fmt.Errorf("get archive: %w", err)
		}

		input := metadata.Input{
			ArchiveID: archive.ID,
			FilePath:  archive.FilePath,
			Title:     archive.Title,
		}

		result, err := pipeline.RunLocal(ctx, input)
		if err != nil {
			return fmt.Errorf("run local metadata pipeline: %w", err)
		}

		if err := metadata.ApplyMetadata(ctx, queries, db, archive.ID, result); err != nil {
			return fmt.Errorf("apply local metadata: %w", err)
		}

		return queue.Enqueue(ctx, JobTypeMetadataRemote, MetadataPayload{ArchiveID: p.ArchiveID})
	}
}

// NewRemoteMetadataHandler runs the full pipeline (local + remote sources) and
// applies whatever gaps remote sources can fill. Intended for low-concurrency
// execution so network sources are not hammered simultaneously.
func NewRemoteMetadataHandler(pipeline *metadata.Pipeline, queries *sqlc.Queries, db *sql.DB, logger *slog.Logger) Handler {
	return func(ctx context.Context, job *Job) error {
		var p MetadataPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		archive, err := queries.GetArchiveByID(ctx, p.ArchiveID)
		if err != nil {
			return fmt.Errorf("get archive: %w", err)
		}

		result, err := pipeline.Run(ctx, metadata.Input{
			ArchiveID: archive.ID,
			FilePath:  archive.FilePath,
			Title:     archive.Title,
		})
		if err != nil {
			return fmt.Errorf("run metadata pipeline: %w", err)
		}

		return metadata.ApplyMetadata(ctx, queries, db, archive.ID, result)
	}
}
