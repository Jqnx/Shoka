package tasks

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"Shoka/internal/thumb"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type ThumbnailPayload struct {
	// ArchivePath string
	Archive *repository.GetArchiveByIDRow
}

func NewThumbnailGenerateTask(arch *repository.GetArchiveByIDRow) (*asynq.Task, error) {
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

	// Create context
	c := context.Background()

	pd := ""
	td := ""
	thumb := thumb.NewThumb(payload.Archive, w.app, pd, td)
	thumb.GetThumbDir()
	thumb.CreatePageDir(thumb.ThumbDir)

	if err := thumb.Generate(); err != nil {
		return err
	}

	if payload.Archive.ThumbsPath != &thumb.ThumbDir {
		if err := w.app.Repo.UpdateThumbPath(c, repository.UpdateThumbPathParams{
			ThumbsPath: &thumb.ThumbDir,
			ArchiveID:  payload.Archive.ArchiveID,
		}); err != nil {
			return err
		}
	}

	since := time.Since(now)
	w.app.Log.Info("new thumbs", "fp:", thumb.PageDir, "elapsed:", since)
	return nil
}

func NewThumbnailProcessor(app *config.App) *ThumbnailProcessor {
	return &ThumbnailProcessor{
		app: app,
	}
}
