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

// NewWatcher returns a Watcher for a single library's directory. The
// returned logger is tagged with the watcher component and library ID.
func NewWatcher(queue JobQueue, libraryID, dir string, log *slog.Logger) *Watcher {
	return &Watcher{
		queue:     queue,
		libraryID: libraryID,
		dir:       dir,
		log:       log.With("component", "watcher", "library_id", libraryID),
	}
}

// Start registers fsnotify watches for w.dir and all of its
// subdirectories, then spawns the goroutine that rescans the library on
// changes until ctx is cancelled.
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
		_ = watcher.Close()
		return err
	}

	w.log.Info("watching library", "dir", w.dir, "watched_dirs", added)

	go w.run(ctx, watcher)

	return nil
}

func (w *Watcher) run(ctx context.Context, watcher *fsnotify.Watcher) {
	// run owns watcher for its whole lifetime; a close failure on shutdown
	// has no caller left to report to.
	defer func() { _ = watcher.Close() }()

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

			debounce = w.handleEvent(ctx, watcher, event, debounce)

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}

			w.log.Error("watcher error", "error", err)
		}
	}
}

// handleEvent reacts to a single fsnotify event: newly created directories
// are added to the watch set, and create/remove/rename events (re)arm the
// debounce timer that triggers a scoped rescan. It returns the timer the
// run loop should keep, so the caller stays a plain select loop.
func (w *Watcher) handleEvent(ctx context.Context, watcher *fsnotify.Watcher, event fsnotify.Event, debounce *time.Timer) *time.Timer {
	// New directories need to be watched explicitly so files copied
	// into them are picked up too.
	if event.Has(fsnotify.Create) {
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			w.watchNewDir(watcher, event.Name)
		}
	}

	if !event.Has(fsnotify.Create) && !event.Has(fsnotify.Remove) && !event.Has(fsnotify.Rename) {
		return debounce
	}

	w.log.Debug("fs event detected", "op", event.Op, "path", event.Name)

	if debounce != nil {
		debounce.Stop()
	}

	return time.AfterFunc(5*time.Second, func() {
		w.log.Info("triggering scan after fs event")

		if err := w.queue.EnqueueOnce(ctx, jobs.JobTypeScan, jobs.ScanPayload{LibraryID: w.libraryID}); err != nil {
			w.log.Error("scan triggered by watcher failed", "error", err)
		}
	})
}

// watchNewDir adds newly created directories (and any subdirectories they
// already contain, e.g. a whole folder moved in at once) to the watcher.
func (w *Watcher) watchNewDir(watcher *fsnotify.Watcher, dir string) {
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			w.log.Warn("could not access path while watching new directory", "path", path, "error", err)
			return nil
		}

		if !d.IsDir() {
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
