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
	c := NewClient(w.app.Client, w.app, arch)
	i := c.NewCover(w.force)
	ch <- i
}

// TODO: Also think about deleting generated thumbnails if archive has not been read recently
// TODO: LastRead column in db, gets updated when user GETs archive pages

func (c *Client) NewCover(force bool) *asynq.TaskInfo {
	// Check if archive has valid ThumbPath
	if c.arch.ThumbPath == nil || *c.arch.ThumbPath == "" {
		c.app.Log.Error("could not create task", "task", "generate cover", "archive", c.arch.ID, "err", "thumbpath does not exist")
		return nil
	}

	// Check if cover directory exists on filesystem for a single archive
	path := filepath.Join(*c.arch.ThumbPath, "cover")
	exists, err := fsutil.DirExists(path)
	if err != nil {
		c.app.Log.Error("could not create task", "err", err.Error())
		return nil
	}

	if !exists {
		// If cover directory does not exist
		// create new one and create new cover

		// 1. Create new directory
		ar := archive.RepoToArchive(c.arch, c.app)
		err := ar.CreateCoverDir()
		if err != nil {
			c.app.Log.Error("could not create cover directory", "err", err.Error())
			return nil
		}

		// 2. Queue new cover task
		cover := c.newCoverTask(path)
		return cover
	} else {
		// If cover directory exists,
		// check if cover exists,
		// if not create a new cover.

		// 1. Checks if file exists on filesystem
		fileName := fmt.Sprintf("%s.webp", c.arch.FileHash)
		_, err := os.Stat(filepath.Join(path, fileName))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cover := c.newCoverTask(path)
				return cover
			} else {
				c.app.Log.Error("could not check if file exists", "err", err.Error())
			}
		}
	}

	// Option to force the creation of new covers.
	if exists && force {
		cover := c.newCoverTask(path)
		return cover
	}

	return nil
}

// newCoverTask creates and queues a NewCreateCoverTask
// returns the TaskInfo and an error
func (c *Client) newCoverTask(coverdir string) *asynq.TaskInfo {
	newCovers, err := tasks.NewCreateCoverTask(c.arch, coverdir)
	if err != nil {
		c.app.Log.Error("could not create task", "err", err.Error())
		return nil
	}
	cover, err := c.client.Enqueue(newCovers)
	if err != nil {
		c.app.Log.Error("could not queue task", "err", err.Error())
		return nil
	}
	c.app.Log.Info("queued task", "task", "generate cover", "archive", c.arch.ID, "id", cover.ID, "queue", cover.Queue, "state", cover.State)
	return cover
}
