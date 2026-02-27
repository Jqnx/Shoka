package workers

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
)

// TODO: Cleanup scheduler
// TODO: What needs to be cleaned up
// TODO: 2. Check if thumbs folder overall size is over a certain size, if so start removing the oldest last read

// Cleanup checks for any archives no longer on present on disk
// and removes their database entry and leftover files (thumbnails and cover)
func (w *Workers) Cleanup() error {
	ctx := context.Background()
	list := fsutil.ListArchives(w.app.Cfg.ContentDir)
	if len(list) == 0 {
		return nil
	}

	filepaths, err := w.app.Repo.GetAllFilePaths(ctx)
	if err != nil {
		return err
	}

	// Check if file from db is no longer on disk
	for _, item := range filepaths {
		if !slices.Contains(list, item.FilePath) {
			if err := w.CleanThumb(item.ID, true); err != nil {
				return err
			}

			if err := w.CleanDB(item); err != nil {
				return err
			}
		}
	}

	// TODO: Make amount of days (interval) configurable
	if err := w.CleanLastRead(14); err != nil {
		return err
	}

	if err := w.CleanDeleted(); err != nil {
		return err
	}

	return nil
}

// CleanDB removes an item of type repository.GetAllFilePathsRow
// from the database
func (w *Workers) CleanDB(item repository.GetAllFilePathsRow) error {
	ctx := context.Background()
	if err := w.app.Repo.DeleteArchiveByFilePath(ctx, item.FilePath); err != nil {
		return err
	}
	w.app.Log.Info("removed from db",
		"task", "cleanup",
		"file", filepath.Base(item.FilePath))
	return nil
}

// CleanThumb removes the thumbnails and
// optionally the cover of an item from disk
func (w *Workers) CleanThumb(id string, includeCover bool) error {
	ctx := context.Background()
	thumbPath, err := w.app.Repo.GetThumbPathByID(ctx, id)
	if err != nil {
		return err
	}
	if includeCover {
		if err := fsutil.RemoveContents(*thumbPath); err != nil {
			return err
		}
	} else {
		if err := fsutil.RemoveContents(filepath.Join(*thumbPath, "pages")); err != nil {
			return err
		}
	}
	w.app.Log.Info("removed thumbs",
		"task", "cleanup",
		"file", id)
	return nil
}

// CleanLastRead removes the thumbnails for items that have
// not been read in n days or more
func (w *Workers) CleanLastRead(days int) error {
	ctx := context.Background()
	lastRead, err := w.app.Repo.GetAllLastRead(ctx)
	if err != nil {
		return err
	}

	for _, item := range lastRead {
		d := time.Duration(days) * time.Hour * 24
		if time.Since(item.LastRead) > d {
			thumbPath, err := w.app.Repo.GetThumbPathByID(ctx, item.ArchiveID)
			if err != nil {
				return err
			}

			pages := filepath.Join(*thumbPath, "pages")
			files, err := os.ReadDir(pages)
			if err != nil {
				return err
			}

			if len(files) > 0 {
				if err := w.CleanThumb(item.ArchiveID, false); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (w *Workers) CleanDeleted() error {
	tmp, err := os.ReadDir(w.app.Cfg.TempDir)
	if err != nil {
		return err
	}

	for _, file := range tmp {
		if strings.Contains(file.Name(), ".deleted") {
			err := os.Remove(filepath.Join(w.app.Cfg.TempDir, file.Name()))
			if err != nil {
				w.app.Log.Error("failed to remove from filesystem",
					"task", "cleanup",
					"file", file.Name(),
					"error", err.Error())
				return err
			}
			w.app.Log.Info("removed from filesystem",
				"task", "cleanup",
				"file", file.Name())
		}
	}

	return nil
}
