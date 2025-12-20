package workers

import (
	"context"

	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/workers/tasks"

	"github.com/hibiken/asynq"
)

// TODO: Also think about deleting generated thumbnails if archive has not been read recently
// TODO: LastRead column in db, gets updated when user GETs archive pages

// Scan takes in a list of archive filepaths,
// checks if they already exist in the database
// and queues a new create archive task if they do not
func (w *Workers) Scan(list []string) {
	// TASKS
	// Scanning

	// Check media type based extension & folder name
	for _, item := range list {
		if fsutil.MatchExtension(item, config.ArchiveExtensions) {
			ctx := context.Background()

			exists, err := w.app.Repo.FilePathExists(ctx, item)
			if err != nil {
				w.app.Log.Error("error checking if file path in db:", "error", err.Error())
			}

			if exists.RowsAffected() == 0 {
				// Create Archive
				newArch, err := tasks.NewCreateArchiveTask(item)
				if err != nil {
					w.app.Log.Error("could not create task:", "error", err.Error())
				}
				arch, err := w.app.Client.Enqueue(newArch, asynq.Queue("critical"))
				if err != nil {
					w.app.Log.Error("could not queue task:", "error", err.Error())
				}
				w.app.Log.Info("archive found:", "id", arch.ID, "queue", arch.Queue, "state", arch.State)
			}

		}
	}
}
