package tasks

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/thumb"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// TODO: REDO

type ThumbnailPayload struct {
	// ArchivePath string
	Archive *repository.Archive
}

func NewThumbnailGenerateTask(arch *repository.Archive) (*asynq.Task, error) {
	// payload, err := json.Marshal(ThumbnailPayload{ArchivePath: archivepath, Archive: arch})
	payload, err := json.Marshal(ThumbnailPayload{Archive: arch})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCreateThumbnail, payload), nil
}

type ThumbnailProcessor struct {
	app *config.App
}

func (w *ThumbnailProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	// Create payload variable
	var payload ThumbnailPayload
	// Bind payload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	now := time.Now()

	// TODO: Make sure pagedir always exists, it breaks the watching of thumbs on request if it does not exist

	// Create context
	c := context.Background()

	// TODO: Move page dir check/creation to handler
	// Check if dir exists, if not create

	// Initialize new ThumbDir
	d := fsutil.NewThumbDir(
		w.app.Cfg.ThumbDir,
		payload.Archive.Type,
		*payload.Archive.Hash)

	thumbdir, err := d.GetThumbDir()
	if err != nil {
		return err
	}

	pagedir, err := d.CreatePageDir(thumbdir)
	if err != nil {
		return err
	}

	if err := thumb.GenerateThumbs(payload.Archive, pagedir, w.app); err != nil {
		return err
	}

	if payload.Archive.ThumbsPath != &thumbdir {
		if err := w.app.Repo.UpdateThumbPath(c, repository.UpdateThumbPathParams{
			ThumbsPath: &thumbdir,
			ArchiveID:  payload.Archive.ArchiveID,
		}); err != nil {
			return err
		}
	}

	since := time.Since(now)
	w.app.Log.Info("new thumbs", "fp:", pagedir, "elapsed:", since)
	return nil
}

func NewThumbnailProcessor(app *config.App) *ThumbnailProcessor {
	return &ThumbnailProcessor{
		app: app,
	}
}
