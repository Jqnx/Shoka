package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/sources"

	"github.com/hibiken/asynq"
)

type MetadataPayload struct {
	Archive *repository.GetArchiveByIDRow
	Source  string
}

func NewMetadataTask(arch *repository.GetArchiveByIDRow, src string) (*asynq.Task, error) {
	payload, err := json.Marshal(MetadataPayload{Archive: arch, Source: src})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeNewMetadata, payload), nil
}

type MetadataProcessor struct {
	app *config.App
}

func (w *MetadataProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	// Create payload variable
	var payload MetadataPayload
	// Bind payload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	var src string
	var content []byte

	zip, err := fsutil.OpenArchive(payload.Archive.FilePath)
	if err != nil {
		return err
	}

	// TODO: Add more metadata file sources (Hentag, NHDownloader)
	switch payload.Source {
	case config.ComicInfoFile:
		fileContent, err := zip.ReadFile(config.ComicInfoFile)
		if err != nil {
			if errors.Is(err, config.ErrFileNotFound) {
				return nil
			} else {
				return err
			}
		}
		content = fileContent
		src = config.SourceComicInfo
	default:
		return fmt.Errorf("no valid metadata file used")
	}

	s, err := sources.NewSource(w.app.Cfg, src, nil)
	if err != nil {
		return err
	}
	if err := s.Unmarshal(content); err != nil {
		w.app.Log.Error("error unmarshalling metadata file:", "err", err.Error(), "file", src, "archive", payload.Archive.ID)
		return err
	}
	meta, err := s.GetMetadata()
	if err != nil {
		return err
	}

	arch := archive.NewArchive(w.app)
	arch.Update(payload.Archive.ID, &meta[0])
	if err := arch.UpdateInDB(ctx, w.app); err != nil {
		return err
	}
	return nil
}

func NewMetadataProcessor(app *config.App) *MetadataProcessor {
	return &MetadataProcessor{
		app: app,
	}
}
