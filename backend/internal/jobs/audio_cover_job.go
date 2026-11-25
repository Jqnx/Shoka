package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// JobTypeAudioCover is the job type that regenerates an audio work's cover
// thumbnail from its chosen image.
const JobTypeAudioCover = "audio_cover"

// AudioCoverPayload selects the audio work whose cover should be generated.
type AudioCoverPayload struct {
	AudioID string `json:"audio_id"`
}

// NewAudioCoverHandler returns a Handler that renders the work's chosen image
// as its cover thumbnail. Both the image list and audio.cover_index are read
// from the database at run time, so a stale payload can never write a cover
// that no longer matches the work's current images or selection.
func NewAudioCoverHandler(processor *image.Processor, queries *sqlc.Queries, log *slog.Logger) Handler {
	log = log.With("component", "audio_cover_job")

	return func(ctx context.Context, job *Job) error {
		var p AudioCoverPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		// Registered before the existence check so it also covers a delete
		// that lands during GenerateAudioCover: that delete already evicted the
		// cache directory, so this reclaims whatever was written after it.
		defer func() {
			if _, err := queries.GetAudioByID(ctx, p.AudioID); errors.Is(err, sql.ErrNoRows) {
				_ = processor.EvictAll(image.AudioCacheKey(p.AudioID))
			}
		}()

		// A scan or SetCover can change the selection while an image decodes.
		// One extra pass renders the latest choice in this invocation instead
		// of leaving the cover absent until a queue retry (the first retry
		// backs off 30s).
		for range 2 {
			changed, err := renderChosenAudioCover(ctx, processor, queries, log, p.AudioID)
			if err != nil {
				return err
			}

			if !changed {
				return nil
			}
		}

		if err := processor.EvictAll(image.AudioCacheKey(p.AudioID)); err != nil {
			return err
		}

		return fmt.Errorf("audio %s cover selection changed during generation", p.AudioID)
	}
}

// renderChosenAudioCover renders the work's currently-selected image. changed
// reports that the selection or image list moved while the image was decoding,
// so the caller should render again from the latest row. Missing images and
// deleted works are a no-op: there is nothing this job can do until a scan.
func renderChosenAudioCover(
	ctx context.Context,
	processor *image.Processor,
	queries *sqlc.Queries,
	log *slog.Logger,
	audioID string,
) (changed bool, err error) {
	audio, err := queries.GetAudioByID(ctx, audioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, processor.EvictAll(image.AudioCacheKey(audioID))
		}

		return false, fmt.Errorf("get audio %s: %w", audioID, err)
	}

	images, err := queries.GetAudioImages(ctx, audioID)
	if err != nil {
		return false, fmt.Errorf("get audio images: %w", err)
	}

	if len(images) == 0 {
		return false, processor.EvictAll(image.AudioCacheKey(audioID))
	}

	index := image.AudioCoverIndex(audio.CoverIndex, len(images))
	relativePath := images[index].RelativePath

	//nolint:gosec // G304: relativePath comes from the scanner's audio_image rows, which are built under the work's folder
	f, err := os.Open(filepath.Join(audio.FolderPath, relativePath))
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("cover image missing, waiting for next scan",
				"audio_id", audioID, "path", relativePath)

			return false, nil
		}

		return false, fmt.Errorf("open cover image: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := processor.GenerateAudioCover(ctx, audioID, f); err != nil {
		return false, err
	}

	current, err := audioCoverCurrent(ctx, queries, audioID, relativePath)
	if err != nil {
		return false, err
	}

	return !current, nil
}

// audioCoverCurrent reports whether relativePath is still the image the row
// selects. GenerateAudioCover writes atomically, but it renders whatever the
// selection was before the check ran, so a changed selection or image list
// would otherwise leave the wrong cover in place.
func audioCoverCurrent(
	ctx context.Context,
	queries *sqlc.Queries,
	audioID, relativePath string,
) (bool, error) {
	latest, err := queries.GetAudioImages(ctx, audioID)
	if err != nil {
		return false, fmt.Errorf("recheck audio images: %w", err)
	}

	if len(latest) == 0 {
		return false, nil
	}

	audio, err := queries.GetAudioByID(ctx, audioID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("recheck audio %s: %w", audioID, err)
	}

	index := image.AudioCoverIndex(audio.CoverIndex, len(latest))

	return latest[index].RelativePath == relativePath, nil
}
