package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/repository"
	"Shoka/internal/thumb"

	"github.com/hibiken/asynq"
)

type CreateCoverPayload struct {
	Archive  *repository.GetArchiveByIDRow
	CoverDir string
}

func NewCreateCoverTask(arch *repository.GetArchiveByIDRow, coverdir string) (*asynq.Task, error) {
	payload, err := json.Marshal(CreateCoverPayload{Archive: arch, CoverDir: coverdir})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCreateCover, payload), nil
}

type CoverProcessor struct {
	app *config.App
}

func (w *CoverProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	// Create payload variable
	var payload CreateCoverPayload
	// Bind payload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	now := time.Now()

	// Create new Cover
	co := thumb.NewCover(payload.Archive, w.app, payload.CoverDir)
	// Generate cover
	cover, err := co.Generate()
	if err != nil {
		return err
	}

	since := time.Since(now).Round(time.Millisecond).String()
	w.app.Log.Info("task completed", "task", "generate cover", "filepath", cover, "elapsed", since)

	return nil
}

func NewCoverProcessor(app *config.App) *CoverProcessor {
	return &CoverProcessor{
		app: app,
	}
}
