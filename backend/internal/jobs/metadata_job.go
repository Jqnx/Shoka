package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/metadata"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

const (
	JobTypeMetadata       = "metadata"
	JobTypeMetadataRemote = "metadata/remote"
)

type MetadataPayload struct {
	ArchiveID string
	// Source optionally pins the job to one named metadata source instead
	// of running the whole priority-ordered pipeline. Empty means "run the
	// pipeline", which is what the scanner enqueues.
	Source string `json:",omitempty"`
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
			LibraryID: archive.LibraryID,
			FilePath:  archive.FilePath,
			Title:     archive.Title,
		}

		// Pinned to one source: run only that, and don't chain the remote
		// job - the caller asked for this source, not "whatever fills the
		// gaps".
		if p.Source != "" {
			result, err := pipeline.FetchWithSource(ctx, p.Source, input)
			if err != nil {
				return skipOrFail(logger, "local", p, err)
			}

			if result == nil {
				return nil
			}

			return metadata.ApplyMetadata(ctx, queries, db, archive.ID, result)
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

		input := metadata.Input{
			ArchiveID: archive.ID,
			LibraryID: archive.LibraryID,
			FilePath:  archive.FilePath,
			Title:     archive.Title,
		}

		if p.Source != "" {
			result, err := pipeline.FetchWithSource(ctx, p.Source, input)
			if err != nil {
				return skipOrFail(logger, "remote", p, err)
			}

			if result == nil {
				return nil
			}

			return metadata.ApplyMetadata(ctx, queries, db, archive.ID, result)
		}

		result, err := pipeline.Run(ctx, input)
		if err != nil {
			return fmt.Errorf("run metadata pipeline: %w", err)
		}

		return metadata.ApplyMetadata(ctx, queries, db, archive.ID, result)
	}
}

// skipOrFail decides whether a pinned-source failure is worth retrying.
// A source that's unknown, disabled for this archive's library, or not
// searchable will fail identically on every attempt, so those are logged
// and swallowed rather than burning the job's retry budget. This matters
// for bulk fetches, where one disabled library would otherwise produce a
// retry storm across every archive in it.
func skipOrFail(logger *slog.Logger, stage string, p MetadataPayload, err error) error {
	switch {
	case errors.Is(err, metadata.ErrUnknownSource),
		errors.Is(err, metadata.ErrDisabledSource),
		errors.Is(err, metadata.ErrNotSearchable):
		logger.Warn("skipping metadata source",
			"stage", stage, "archive_id", p.ArchiveID, "source", p.Source, "reason", err)

		return nil
	default:
		return fmt.Errorf("fetch from source %q: %w", p.Source, err)
	}
}
