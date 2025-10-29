package workers

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"Shoka/internal/archive"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/workers/tasks"

	"github.com/hibiken/asynq"
)

func (w *Workers) Covers(ch chan *asynq.TaskInfo, arch *repository.GetArchiveByIDRow) {
	ac := w.NewAsynqClient()
	defer ac.Close()

	// TODO: Use global client
	c := NewClient(ac, w.app, arch)
	i := c.NewCover(w.force)
	ch <- i
}

// TODO: Also think about deleting generated thumbnails if archive has not been read recently
// TODO: LastRead column in db, gets updated when user GETs archive pages

func (c *Client) NewCover(force bool) *asynq.TaskInfo {
	// Check if cover directory exists on filesystem for a single archive
	path := filepath.Join(*c.arch.ThumbsPath, "cover")
	exists, err := fsutil.DirExists(path)
	if err != nil {
		c.app.Log.Error("could not create task:", "error", err.Error())
		return nil
	}

	if !exists {
		// If cover directory does not exist
		// create new one and create new cover

		// 1. Create new directory
		ar := archive.RepoToArchive(c.arch, c.app)
		err := ar.CreateCoverDir()
		if err != nil {
			c.app.Log.Error("could not create task:", "error", err.Error())
			return nil
		}

		// 2. Queue new cover task
		cover, err := newCoverTask(c.arch, c.client, path)
		if err != nil {
			c.app.Log.Error("could not create task:", "error", err.Error())
			return nil
		}
		c.app.Log.Info("queued task: cover did not exist:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
		return cover
	} else {
		// If cover directory exists,
		// check if cover exists,
		// if not create a new cover.

		// 1. Checks if file exists on filesystem
		_, err := os.Stat(filepath.Join(path, fmt.Sprintf("%s.webp", c.arch.Hash)))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cover, err := newCoverTask(c.arch, c.client, path)
				if err != nil {
					c.app.Log.Error("could not create task:", "error", err.Error())
				}
				c.app.Log.Info("queued task: cover not found on filesystem:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
				return cover
			} else {
				c.app.Log.Error("could not check if file exists:", "error", err.Error())
			}
		}
	}

	// Option to force the creation of new covers.
	if exists && force {
		cover, err := newCoverTask(c.arch, c.client, path)
		if err != nil {
			c.app.Log.Error("could not create task:", "error", err.Error())
		}
		c.app.Log.Info("queued task cover forced:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
		return cover
	}

	return nil
}

// newCoverTask creates and queues a NewCreateCoverTask
// returns the TaskInfo and an error
func newCoverTask(arch *repository.GetArchiveByIDRow, client *asynq.Client, coverdir string) (*asynq.TaskInfo, error) {
	newCovers, err := tasks.NewCreateCoverTask(arch, coverdir)
	if err != nil {
		return nil, err
	}
	cover, err := client.Enqueue(newCovers)
	if err != nil {
		return nil, err
	}
	return cover, nil
}
