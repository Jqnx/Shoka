package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"Shoka/internal/archive"
	"Shoka/internal/config"

	"github.com/hibiken/asynq"
)

type ArchivePayload struct {
	ArchivePath string
}

func NewCreateArchiveTask(path string) (*asynq.Task, error) {
	payload, err := json.Marshal(ArchivePayload{ArchivePath: path})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeCreateArchive, payload), nil
}

type ArchiveProcessor struct {
	cfg *config.Config
	app *config.App
}

func (a *ArchiveProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ArchivePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	c := context.Background()

	arch := archive.NewArchive(a.app)
	arch.New(payload.ArchivePath)
	if err := arch.Insert(c, a.app); err != nil {
		return err
	}
	a.app.Log.Info("archive created", "title:", arch.Title)

	return nil
}

func NewArchiveProcessor(cfg *config.Config, app *config.App) *ArchiveProcessor {
	return &ArchiveProcessor{
		cfg: cfg,
		app: app,
	}
}
