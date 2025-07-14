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

type CreateCoverPayload struct {
	Archive *repository.GetArchiveByIDRow
}

func NewCreateCoverTask(arch *repository.GetArchiveByIDRow) (*asynq.Task, error) {
	payload, err := json.Marshal(CreateCoverPayload{Archive: arch})
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

	// Create context
	c := context.Background()

	// Create new Cover
	co := thumb.NewCover(payload.Archive, w.app, "")
	// Create new Cover directory
	co.CreateDir(*payload.Archive.ThumbsPath)
	// Generate cover
	cover, err := co.Generate()
	if err != nil {
		return err
	}

	// Checks if ThumbsPath in DB is the same as generated thumbpath
	// If not then update db with new thumbdir
	if payload.Archive.CoverPath != &cover {
		if err := w.app.Repo.UpdateCoverPath(c, repository.UpdateCoverPathParams{
			CoverPath: &cover,
			ArchiveID: payload.Archive.ArchiveID,
		}); err != nil {
			return err
		}
	}

	since := time.Since(now)
	w.app.Log.Info("new cover", "fp:", cover, "elapsed:", since)

	return nil
}

func NewCoverProcessor(app *config.App) *CoverProcessor {
	return &CoverProcessor{
		app: app,
	}
}
