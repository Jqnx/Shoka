package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// JobTypeAudioThumbnail is the job type that generates an audio work's gallery
// thumbnails.
const JobTypeAudioThumbnail = "audio_thumbnail"

// AudioThumbnailPayload selects the audio work whose gallery should be
// thumbnailed.
type AudioThumbnailPayload struct {
	AudioID string `json:"audio_id"`
}

// NewAudioThumbnailHandler returns a Handler that renders every image of an
// audio work as a 360px gallery thumbnail. Images are read from disk through
// the database's relative paths at run time, so a stale payload can never
// write thumbnails for a work's previous image set.
//
// Each ready thumbnail and the terminal state are published on the work's
// audio cache key, which is what the /thumbnails/events stream subscribes to.
func NewAudioThumbnailHandler(
	processor *image.Processor,
	broadcaster *events.ThumbnailBroadcaster,
	queries *sqlc.Queries,
	log *slog.Logger,
) Handler {
	log = log.With("component", "audio_thumbnail_job")

	return func(ctx context.Context, job *Job) (err error) {
		var p AudioThumbnailPayload
		if err = job.Decode(&p); err != nil {
			return fmt.Errorf("decode payload: %w", err)
		}

		key := image.AudioCacheKey(p.AudioID)

		// Same terminal-only rule as NewThumbnailHandler: an error on an
		// earlier attempt is followed by a retry, so subscribers are only told
		// generation gave up once this was the last attempt.
		defer func() {
			if err != nil && job.Attempts >= job.MaxAttempts {
				broadcaster.Publish(key, events.ThumbnailEvent{Done: true, Error: err.Error()})
			}
		}()

		audio, err := queries.GetAudioByID(ctx, p.AudioID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				broadcaster.Publish(key, events.ThumbnailEvent{Done: true})

				return nil
			}

			return fmt.Errorf("get audio %s: %w", p.AudioID, err)
		}

		images, err := queries.GetAudioImages(ctx, p.AudioID)
		if err != nil {
			return fmt.Errorf("get audio images: %w", err)
		}

		if err := resetStaleAudioGallery(processor, p.AudioID); err != nil {
			return err
		}

		generated, failed, err := generateAudioThumbs(ctx, processor, broadcaster, log, p.AudioID, audio, images)
		if err != nil {
			return err
		}

		return finishAudioThumbs(ctx, processor, broadcaster, queries, log, p.AudioID, images, generated, failed)
	}
}

// resetStaleAudioGallery wipes a gallery rendered with different settings. A
// stale gallery belongs to another render pass, so its thumbs cannot be
// resumed. Writing the current marker right after the wipe makes a retry
// after a partial run skip the wipe and keep the thumbs that run produced.
func resetStaleAudioGallery(processor *image.Processor, audioID string) error {
	if !processor.AudioGalleryStale(audioID) {
		return nil
	}

	if err := processor.EvictAudioGallery(audioID); err != nil {
		return fmt.Errorf("evict stale audio gallery: %w", err)
	}

	if err := processor.WriteAudioGalleryVersion(audioID); err != nil {
		return fmt.Errorf("write audio gallery version: %w", err)
	}

	return nil
}

// generateAudioThumbs renders every image whose thumbnail is missing, skipping
// those already on disk and reporting how many it generated and how many it
// could not render. A single undecodable image must not block the rest of the
// work, so its failure is logged and counted rather than returned.
func generateAudioThumbs(
	ctx context.Context,
	processor *image.Processor,
	broadcaster *events.ThumbnailBroadcaster,
	log *slog.Logger,
	audioID string,
	audio sqlc.Audio,
	images []sqlc.AudioImage,
) (generated, failed int, err error) {
	key := image.AudioCacheKey(audioID)

	for index := range images {
		if err := ctx.Err(); err != nil {
			return generated, failed, err
		}

		if processor.AudioThumbPath(audioID, index) != "" {
			broadcaster.Publish(key, events.ThumbnailEvent{Index: index})

			continue
		}

		relativePath := images[index].RelativePath
		if !filepath.IsLocal(relativePath) {
			log.Warn("invalid audio image path", "audio_id", audioID, "path", relativePath)

			failed++

			continue
		}

		//nolint:gosec // G304: RelativePath passed filepath.IsLocal above
		f, err := os.Open(filepath.Join(audio.FolderPath, relativePath))
		if err != nil {
			log.Warn("open audio image failed", "audio_id", audioID, "path", relativePath, "error", err)

			failed++

			continue
		}

		err = processor.GenerateAudioThumb(ctx, audioID, index, f)
		_ = f.Close()

		if err != nil {
			log.Warn("generate audio thumbnail failed",
				"audio_id", audioID, "path", relativePath, "index", index, "error", err)

			failed++

			continue
		}

		generated++

		broadcaster.Publish(key, events.ThumbnailEvent{Index: index})
	}

	return generated, failed, nil
}

// finishAudioThumbs validates the gallery before reporting success. A scan
// that landed mid-generation rebuilt the image list and wiped the cache
// directory, and index-keyed thumbs for the old list would then be wrong; a
// vanished thumbnail or a scanner eviction likewise leaves the gallery
// incomplete. Both cases evict and fail the job so the retry rebuilds against
// the current state.
func finishAudioThumbs(
	ctx context.Context,
	processor *image.Processor,
	broadcaster *events.ThumbnailBroadcaster,
	queries *sqlc.Queries,
	log *slog.Logger,
	audioID string,
	images []sqlc.AudioImage,
	generated, failed int,
) error {
	latest, err := queries.GetAudioImages(ctx, audioID)
	if err != nil {
		return fmt.Errorf("recheck audio images: %w", err)
	}

	if !sameImagePaths(images, latest) {
		if err := processor.EvictAudioGallery(audioID); err != nil {
			return fmt.Errorf("evict audio gallery: %w", err)
		}

		return fmt.Errorf("audio %s images changed during thumbnail generation", audioID)
	}

	// The frontend fills the whole grid on a successful done, so a success
	// must not be sent while any thumbnail is missing.
	if failed == 0 && !processor.AudioGalleryReady(audioID, len(images)) {
		if err := processor.EvictAudioGallery(audioID); err != nil {
			return fmt.Errorf("evict audio gallery: %w", err)
		}

		return fmt.Errorf("audio %s gallery incomplete after generation", audioID)
	}

	log.Info("audio thumbnails generated", "audio_id", audioID,
		"generated", generated, "failed", failed, "total_images", len(images))

	key := image.AudioCacheKey(audioID)

	if failed > 0 {
		// Permanently corrupt files do not deserve a queue retry: the version
		// marker is current, so the next gallery visit re-enqueues and retries
		// only the missing thumbs.
		broadcaster.Publish(key, events.ThumbnailEvent{
			Done:  true,
			Error: fmt.Sprintf("%d audio image(s) could not be thumbnailed", failed),
		})

		return nil
	}

	broadcaster.Publish(key, events.ThumbnailEvent{Done: true})

	return nil
}

// sameImagePaths reports whether two image lists name the same relative paths
// in the same order - an index-keyed thumbnail is only valid for the list it
// was rendered from.
func sameImagePaths(a, b []sqlc.AudioImage) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].RelativePath != b[i].RelativePath {
			return false
		}
	}

	return true
}
