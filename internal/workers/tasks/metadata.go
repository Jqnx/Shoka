package tasks

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/metadata"
	"Shoka/internal/repository"
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/bodgit/sevenzip"
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

	switch payload.Source {
	case "file":
		// Scan zip file for metadata files
		if fsutil.Is7z(*payload.Archive.FilePath) {
			zip, err := sevenzip.OpenReader(*payload.Archive.FilePath)
			if err != nil {
				return err
			}
			defer zip.Close()

			for _, file := range zip.File {
				switch filepath.Base(file.Name) {
				case config.ComicInfoFile:
					content, err := fsutil.Read7z(*file)
					if err != nil {
						return err
					}

					mb := metadata.GetBuilder("comicinfo")
					d := metadata.NewDirector(mb)
					meta, err := d.FetchMetadata(content)
					if err != nil {
						return err
					}
					ab := archive.GetBuilder(w.app)
					archive := ab.UpdateArchive(payload.Archive.ArchiveID, &meta[0])
					if err := archive.Update(ctx, w.app); err != nil {
						return err
					}
					return nil
					// TODO: Extra file-based metadata file support
				}
			}
		} else {
			zip, err := zip.OpenReader(*payload.Archive.FilePath)
			if err != nil {
				return err
			}
			defer zip.Close()

			for _, file := range zip.File {
				switch filepath.Base(file.Name) {
				case config.ComicInfoFile:
					content, err := fsutil.ReadZip(*file)
					if err != nil {
						return err
					}

					b := metadata.GetBuilder("comicinfo")
					d := metadata.NewDirector(b)
					meta, err := d.FetchMetadata(content)
					if err != nil {
						return err
					}
					ab := archive.GetBuilder(w.app)
					archive := ab.UpdateArchive(payload.Archive.ArchiveID, &meta[0])
					if err := archive.Update(ctx, w.app); err != nil {
						return err
					}
					return nil
					// TODO: Extra file-based metadata file support
				}
			}
		}
	}
	return nil
}

func NewMetadataProcessor(app *config.App) *MetadataProcessor {
	return &MetadataProcessor{
		app: app,
	}
}
