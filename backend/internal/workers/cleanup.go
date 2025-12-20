package workers

import (
	"context"
	"slices"

	"Shoka/internal/fsutil"
)

// TODO: Cleanup scheduler
// TODO: Figure out a way to differentiate between files being removed or directory not mounted/exists

// What needs to be cleaned up
// 1. Entries from db if they do not exist on disk anymore
//   - Be careful that it does NOT delete everything if folder is not properly mounted or cannot be found
//
// 2. Thumbnails & Cover if archive does not exist on drive anymore
//   - Be careful for same thing as above
//   - Should precede cleaning from db (as thumb_path and cover_path will be needed)
//
// 3. Session data if a session has expired
// 4. Thumbnails if an archive has not been read recently (2 weeks/14 days)
func (w *Workers) Cleanup() error {
	list := fsutil.ListArchives(w.app.Cfg.ContentDir)
	if err := w.CleanDB(list); err != nil {
		return err
	}
	return nil
}

func (w *Workers) CleanDB(list []string) error {
	ctx := context.Background()
	filepaths, err := w.app.Repo.GetAllFilePaths(ctx)
	if err != nil {
		return err
	}
	for _, item := range filepaths {
		if !slices.Contains(list, item.FilePath) {
			if err := w.app.Repo.DeleteArchiveByFilePath(ctx, item.FilePath); err != nil {
				return err
			}
			w.app.Log.Info("removed", "file", item)
		}
	}
	return nil
}
