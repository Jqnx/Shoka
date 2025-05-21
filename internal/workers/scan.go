package workers

import (
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/workers/tasks"
	"context"
	"fmt"

	"github.com/hibiken/asynq"
)

// TODO:
// Also think about deleting generated thumbnails if archive has not been read recently
// LastRead column in db, gets updated when user GETs archive pages

func (w *Workers) NewScanClient() {
	url := fmt.Sprintf("%v:%v", w.app.Cfg.Workers.RedisHost, w.app.Cfg.Workers.RedisPort)
	opt := asynq.RedisClientOpt{
		Addr: url,
	}
	client := asynq.NewClient(opt)
	defer client.Close()

	// TASKS
	// Scanning

	// Check media type based extension & folder name

	list := fsutil.ListArchives(w.app.Cfg.ContentDir)
	for _, item := range list {
		if fsutil.MatchExtension(item, config.ArchiveExtensions) {

			ctx := context.Background()

			exists, err := w.app.Repo.FilePathExists(ctx, &item)
			if err != nil {
				w.app.Log.Error("error checking if file path in db:", "error", err.Error())
			}

			if exists.RowsAffected() == 0 {
				// Create Archive
				newArch, err := tasks.NewCreateArchiveTask(item)
				if err != nil {
					w.app.Log.Error("could not create task:", "error", err.Error())
				}
				arch, err := client.Enqueue(newArch, asynq.Queue("critical"))
				if err != nil {
					w.app.Log.Error("could not queue task:", "error", err.Error())
				}
				w.app.Log.Info("archive found:", "id", arch.ID, "queue", arch.Queue, "state", arch.State)
			}

		}
	}
}
