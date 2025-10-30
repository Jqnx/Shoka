package workers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/workers/tasks"

	"github.com/hibiken/asynq"
)

func (w *Workers) Thumbs(ch chan *asynq.TaskInfo, arch *repository.GetArchiveByIDRow) {
	c := NewClient(w.app.Client, w.app, arch)
	i := c.NewThumb(w.force)
	ch <- i
}

func (c *Client) NewThumb(force bool) *asynq.TaskInfo {
	ctx := context.Background()

	aDir := fsutil.NewArchiveDir(c.app.Cfg.ThumbDir, c.arch.Type, c.arch.Hash)
	tDir, err := aDir.GenThumbDir()
	if err != nil {
		c.app.Log.Error("could not generate thumbnail", "err", err.Error())
		return nil
	}
	pDir := filepath.Join(tDir, "pages")

	if c.arch.ThumbsPath == nil || c.arch.ThumbsPath != &tDir {
		if err := c.app.Repo.UpdateThumbPath(ctx, repository.UpdateThumbPathParams{
			ID:         c.arch.ID,
			ThumbsPath: &tDir,
			UpdatedAt:  time.Now(),
		}); err != nil {
			c.app.Log.Error("could not update thumbs_path in db", "err", err.Error(), "archive", c.arch.ID)
		}
		c.arch.ThumbsPath = &tDir
	}

	pages, err := os.ReadDir(pDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_, err := aDir.CreateDirs()
			if err != nil {
				c.app.Log.Error("could not create necessary directories", "err", err.Error(), "archive", c.arch.ID)
				return nil
			}
			thumbTask := c.newThumbTask()
			return thumbTask
		} else {
			c.app.Log.Error("could not open pages dir", "err", err.Error())
			return nil
		}
	}

	if len(pages) != int(c.arch.PageCount) {
		thumbTask := c.newThumbTask()
		return thumbTask
	}

	if force {
		thumbTask := c.newThumbTask()
		return thumbTask
	}
	return nil
}

func (c *Client) newThumbTask() *asynq.TaskInfo {
	newThumbs, err := tasks.NewThumbnailGenerateTask(c.arch)
	if err != nil {
		c.app.Log.Error("could not queue thumbnail task", "err", err.Error())
		return nil
	}
	thumb, err := c.client.Enqueue(newThumbs)
	if err != nil {
		c.app.Log.Error("could not queue thumbnail task", "err", err.Error())
		return nil
	}
	c.app.Log.Info("queued task", "task", "generate thumbnails", "archive", c.arch.ID, "id", thumb.ID, "queue", thumb.Queue, "state", thumb.State)
	return thumb
}
