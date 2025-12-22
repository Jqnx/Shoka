package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/thumb"

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

	arch := archive.RepoToArchive(payload.Archive, w.app)

	thumb := thumb.NewThumb(arch, w.app)
	pagePath := filepath.Join(*arch.ThumbPath, "pages")
	ok, err := fsutil.DirExists(pagePath)
	if err != nil {
		return err
	}
	if !ok {
		if err := arch.CreatePagesDir(); err != nil {
			return err
		}
	}

	if err := thumb.Generate(); err != nil {
		return err
	}

	since := time.Since(now)
	w.app.Log.Info("new thumbs", "fp", pagePath, "elapsed", since)
	return nil
}

func NewThumbnailProcessor(app *config.App) *ThumbnailProcessor {
	return &ThumbnailProcessor{
		app: app,
	}
}
