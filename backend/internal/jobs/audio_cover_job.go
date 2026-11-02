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
// thumbnail from its first image.
const JobTypeAudioCover = "audio_cover"

// AudioCoverPayload selects the audio work whose cover should be generated.
type AudioCoverPayload struct {
	AudioID string `json:"audio_id"`
}

// NewAudioCoverHandler returns a Handler that renders an audio work's first
// image as its cover thumbnail. The image is read from the database at run
// time, so a stale payload can never write a cover that no longer matches the
// work's current images.
func NewAudioCoverHandler(processor *image.Processor, queries *sqlc.Queries, log *slog.Logger) Handler {
	log = log.With("component", "audio_cover_job")

	return func(ctx context.Context, job *Job) error {
		var p AudioCoverPayload
		if err := job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		audio, err := queries.GetAudioByID(ctx, p.AudioID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return processor.EvictAll(image.AudioCacheKey(p.AudioID))
			}

			return fmt.Errorf("get audio %s: %w", p.AudioID, err)
		}

		images, err := queries.GetAudioImages(ctx, p.AudioID)
		if err != nil {
			return fmt.Errorf("get audio images: %w", err)
		}

		if len(images) == 0 {
			return processor.EvictAll(image.AudioCacheKey(p.AudioID))
		}

		f, err := os.Open(filepath.Join(audio.FolderPath, images[0].RelativePath))
		if err != nil {
			if os.IsNotExist(err) {
				log.Warn("cover image missing, waiting for next scan",
					"audio_id", p.AudioID, "path", images[0].RelativePath)

				return nil
			}

			return fmt.Errorf("open cover image: %w", err)
		}
		defer func() { _ = f.Close() }()

		return processor.GenerateThumbnail(ctx, image.AudioCacheKey(p.AudioID), 0, f)
	}
}
