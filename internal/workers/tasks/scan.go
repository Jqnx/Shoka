package tasks

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type ScanPayload struct {
	FilePath string
}

func NewScanTask(path string) (*asynq.Task, error) {
	payload, err := json.Marshal(ScanPayload{FilePath: path})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeScan, payload), nil
}

type ScanProcessor struct {
	cfg *config.Config
	app *config.App
}

func (w *ScanProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload ScanPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	c := context.Background()

	_, err := archive.CreateFromFile(c, payload.FilePath, w.app)
	if err != nil {
		return err
	}

	return nil
}

func NewScanProcessor(cfg *config.Config, app *config.App) *ScanProcessor {
	return &ScanProcessor{
		cfg: cfg,
		app: app,
	}
}

func ArchiveScan(ctx context.Context, path string, client *asynq.Client, a *config.App) error {
	// TODO:
	// Create folders for media types
	// Use archive folder here

	// If file path exists but does not have a thumbnail, then generate a thumbnail here

	// Also update metadata here

	return nil
}
