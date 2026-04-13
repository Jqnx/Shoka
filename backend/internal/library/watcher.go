package library

import (
	"context"
	"log/slog"
	"time"

	"Shoka/internal/jobs"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	queue JobQueue
	dirs  []string
	log   *slog.Logger
}

func NewWatcher(queue JobQueue, dirs []string, log *slog.Logger) *Watcher {
	return &Watcher{
		queue: queue,
		dirs:  dirs,
		log:   log.With("component", "watcher"),
	}
}

func (w *Watcher) Start(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	for _, dir := range w.dirs {
		if err := watcher.Add(dir); err != nil {
			return err
		}
		w.log.Info("watching directory", "dir", dir)
	}

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

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				w.log.Debug("fs event detected", "op", event.Op, "path", event.Name)

				if debounce != nil {
					debounce.Stop()
				}
				debounce = time.AfterFunc(5*time.Second, func() {
					w.log.Info("triggering scan after fs event")
					if err := w.queue.EnqueueOnce(ctx, jobs.JobTypeScan, jobs.ScanPayload{}); err != nil {
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
