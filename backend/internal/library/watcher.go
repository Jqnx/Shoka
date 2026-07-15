package library

import (
	"Shoka/internal/jobs"
	"context"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches a single library's root directory (recursively) for
// filesystem changes and triggers a scoped rescan of that library.
type Watcher struct {
	queue     JobQueue
	libraryID string
	dir       string
	log       *slog.Logger
}

func NewWatcher(queue JobQueue, libraryID, dir string, log *slog.Logger) *Watcher {
	return &Watcher{
		queue:     queue,
		libraryID: libraryID,
		dir:       dir,
		log:       log.With("component", "watcher", "library_id", libraryID),
	}
}

func (w *Watcher) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// fsnotify only watches the directories you explicitly Add — it is not
	// recursive. Doujinshi libraries are typically nested (artist/series
	// folders), so walk the tree and watch every subdirectory up front.
	added := 0
	if err := filepath.WalkDir(w.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			w.log.Warn("could not access path while setting up watches", "path", path, "error", err)
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if err := watcher.Add(path); err != nil {
			w.log.Warn("failed to watch directory", "path", path, "error", err)
			return nil
		}
		added++
		return nil
	}); err != nil {
		watcher.Close()
		return err
	}

	w.log.Info("watching library", "dir", w.dir, "watched_dirs", added)

	go w.run(ctx, watcher)
	return nil
}

func (w *Watcher) run(ctx context.Context, watcher *fsnotify.Watcher) {
	defer watcher.Close()

	// Debounce Timer
	// We wait 5s after the last event before scanning
	// this handles the case where copying a large file fires dozens of events.
	var debounce *time.Timer

	for {
		select {
		case <-ctx.Done():
			if debounce != nil {
				debounce.Stop()
			}
			return

		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// New directories need to be watched explicitly so files copied
			// into them are picked up too.
			if event.Has(fsnotify.Create) {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					w.watchNewDir(watcher, event.Name)
				}
			}

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				w.log.Debug("fs event detected", "op", event.Op, "path", event.Name)

				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(5*time.Second, func() {
					w.log.Info("triggering scan after fs event")
					if err := w.queue.EnqueueOnce(ctx, jobs.JobTypeScan, jobs.ScanPayload{LibraryID: w.libraryID}); err != nil {
						w.log.Error("scan triggered by watcher failed", "error", err)
					}
				})
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			w.log.Error("watcher error", "error", err)
		}
	}
}

// watchNewDir adds newly created directories (and any subdirectories they
// already contain, e.g. a whole folder moved in at once) to the watcher.
func (w *Watcher) watchNewDir(watcher *fsnotify.Watcher, dir string) {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		if len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}
		if err := watcher.Add(path); err != nil {
			w.log.Warn("failed to watch new directory", "path", path, "error", err)
		}
		return nil
	})
}
