package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"Shoka/internal/repository"
	"Shoka/internal/thumb"
	"Shoka/internal/workers/tasks"

	"github.com/hibiken/asynq"
)

func (w *Workers) Thumbs(ch chan *asynq.TaskInfo, arch *repository.GetArchiveByIDRow) {
	ac := w.NewAsynqClient()
	defer ac.Close()

	// TODO: Use global client
	c := NewClient(ac, w.app, arch)
	i := c.NewThumb(w.force)
	ch <- i
}

func (c *Client) NewThumb(force bool) *asynq.TaskInfo {
	ctx := context.Background()

	if c.arch.ThumbsPath != nil {
		// Get thumbs_path contents
		path, _ := filepath.Abs(filepath.Join(*c.arch.ThumbsPath, "pages"))

		// Checks if folder exists on filesystem
		pages, err := os.ReadDir(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cover, err := newThumbTask(c.arch, c.client)
				if err != nil {
					c.app.Log.Error("could not create task:", "error", err.Error())
				}
				c.app.Log.Info("queued task thumb page folder not exist:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
				return cover
			} else {
				c.app.Log.Error("could not open pages dir:", "error", err.Error())
			}
		} else if len(pages) != int(c.arch.PageCount) {
			cover, err := newThumbTask(c.arch, c.client)
			if err != nil {
				c.app.Log.Error("could not create task:", "error", err.Error())
			}
			c.app.Log.Info("queued task thumb pagecount does not match:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
			return cover
		}
	} else {
		thumb := thumb.NewThumb(c.arch, c.app, "", "")
		thumb.GetThumbDir()

		pagesdir, _ := filepath.Abs(filepath.Join(thumb.ThumbDir, "pages"))
		pages, err := os.ReadDir(pagesdir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cover, err := newThumbTask(c.arch, c.client)
				if err != nil {
					c.app.Log.Error("could not create task:", "error", err.Error())
				}
				c.app.Log.Info("queued task pages directory not exist:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
				return cover
			} else {
				c.app.Log.Error("could not open pages dir:", "error", err.Error())
			}
		} else if len(pages) != int(c.arch.PageCount) {
			cover, err := newThumbTask(c.arch, c.client)
			if err != nil {
				c.app.Log.Error("could not create task:", "error", err.Error())
			}
			c.app.Log.Info("queued task thumb pagecount does not match:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
			return cover
		} else {
			if err := c.app.Repo.UpdateThumbPath(ctx, repository.UpdateThumbPathParams{
				ID:         c.arch.ID,
				ThumbsPath: &thumb.ThumbDir,
				UpdatedAt:  time.Now(),
			}); err != nil {
				c.app.Log.Error("could not update thumbs_path in db:", "error", err.Error())
			}
		}
	}

	if force {
		cover, err := newThumbTask(c.arch, c.client)
		if err != nil {
			c.app.Log.Error("could not create task:", "error", err.Error())
		}
		c.app.Log.Info("queued task thumb forced:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
		return cover
	}
	return nil
}

func newThumbTask(arch *repository.GetArchiveByIDRow, client *asynq.Client) (*asynq.TaskInfo, error) {
	newThumbs, err := tasks.NewThumbnailGenerateTask(arch)
	if err != nil {
		return nil, err
	}
	thumb, err := client.Enqueue(newThumbs)
	if err != nil {
		return nil, err
	}
	return thumb, nil
}
