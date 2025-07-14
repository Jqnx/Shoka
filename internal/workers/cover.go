package workers

import (
	"Shoka/internal/repository"
	"Shoka/internal/workers/tasks"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"
)

func (w *Workers) Covers(ch chan *asynq.TaskInfo, arch *repository.GetArchiveByIDRow) {
	ac := w.NewAsynqClient()
	defer ac.Close()

	c := NewClient(ac, w.app, arch)
	i := c.NewCover(w.force)
	ch <- i
}

// TODO: Also think about deleting generated thumbnails if archive has not been read recently
// TODO: LastRead column in db, gets updated when user GETs archive pages

func (c *Client) NewCover(force bool) *asynq.TaskInfo {
	ctx := context.Background()

	// If cover_path is in db, check if covers exist on file system.
	// If they do not, create covers.
	if c.arch.CoverPath != nil {
		// Get absolute cover path
		path, _ := filepath.Abs(filepath.Join(*c.arch.ThumbsPath, "cover"))

		// Checks if file exists on filesystem
		_, err := os.Stat(filepath.Join(path, *c.arch.CoverPath))
		if err != nil {
			// If cover does not exists, create new one
			// Else log error
			if errors.Is(err, os.ErrNotExist) {
				cover, err := newCoverTask(c.arch, c.client)
				if err != nil {
					c.app.Log.Error("could not create task:", "error", err.Error())
				}
				c.app.Log.Info("queued task cover not on fs:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
				return cover
			} else {
				c.app.Log.Error("could not check if file exists:", "error", err.Error())
			}
		}
	} else {
		// If cover_path is not in db create covers
		// Default behavior

		// Checks if file exists on filesystem first
		// TODO: Change coverpath to a full path
		path := filepath.Join(*c.arch.ThumbsPath, "cover")
		name := fmt.Sprintf("%v.webp", c.arch.Hash)
		full, _ := filepath.Abs(filepath.Join(path, name))
		_, err := os.Stat(full)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cover, err := newCoverTask(c.arch, c.client)
				if err != nil {
					c.app.Log.Error("could not create task:", "error", err.Error())
				}
				c.app.Log.Info("queued task cover not on fs:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
				return cover
			} else {
				c.app.Log.Error("could not check if file exists:", "error", err.Error())
			}
		}
		if err := c.app.Repo.UpdateCoverPath(ctx, repository.UpdateCoverPathParams{
			CoverPath: &name,
			ArchiveID: c.arch.ArchiveID,
		}); err != nil {
			c.app.Log.Error("could not update cover_path in db:", "error", err.Error())
		}
	}

	// Option to force the creation of new covers.
	if force {
		cover, err := newCoverTask(c.arch, c.client)
		if err != nil {
			c.app.Log.Error("could not create task:", "error", err.Error())
		}
		c.app.Log.Info("queued task cover forced:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
		return cover
	}

	//}
	return nil
}

// newCoverTask creates and queues a NewCreateCoverTask
// returns the TaskInfo and an error
func newCoverTask(arch *repository.GetArchiveByIDRow, client *asynq.Client) (*asynq.TaskInfo, error) {
	newCovers, err := tasks.NewCreateCoverTask(arch)
	if err != nil {
		return nil, err
	}
	cover, err := client.Enqueue(newCovers)
	if err != nil {
		return nil, err
	}
	return cover, nil
}
