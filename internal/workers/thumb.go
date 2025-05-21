package workers

import (
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"Shoka/internal/tasks"
	"errors"
	"os"
	"path/filepath"

	"github.com/hibiken/asynq"
)

// TODO: REDO

func (w *Workers) Thumbs(arch *repository.Archive) {
	ac := w.NewAsynqClient()
	defer ac.Close()

	c := NewClient(ac, w.app, arch)
	c.NewThumb()
}

func (c *Client) NewThumb() {
	if c.arch.ThumbsPath != nil {
		// Get thumbs_path contents
		path, err := filepath.Abs(filepath.Join(*c.arch.ThumbsPath, "pages"))
		if err != nil {
			// If folder does not exist on filesystem, create it first
			// Else log error
			if errors.Is(err, os.ErrNotExist) {
				d := fsutil.NewThumbDir(c.app.Cfg.ThumbDir, c.arch.Type, *c.arch.Hash)
				_, err = d.CreatePageDir(*c.arch.ThumbsPath)
				if err != nil {
					return
				}
			} else {
				c.app.Log.Error("could not open thumbs dir:", "error", err.Error())
			}
		}

		// Checks if file exists on filesystem
		pages, err := os.ReadDir(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				cover, err := newThumbTask(c.arch, c.client)
				if err != nil {
					c.app.Log.Error("could not create task:", "error", err.Error())
				}
				c.app.Log.Info("queued task thumb page folder not exist:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
			} else {
				c.app.Log.Error("could not open pages dir:", "error", err.Error())
			}
		} else if len(pages) < int(c.arch.PageCount) {
			cover, err := newThumbTask(c.arch, c.client)
			if err != nil {
				c.app.Log.Error("could not create task:", "error", err.Error())
			}
			c.app.Log.Info("queued task thumb pagecount does not match:", "id", cover.ID, "queue", cover.Queue, "state", cover.State)
		}

	}

	if c.arch.ThumbsPath == nil {

		thumbs, err := newThumbTask(c.arch, c.client)
		if err != nil {
			c.app.Log.Error("could not generate thumbnail", "error:", err)
		}
		c.app.Log.Info("queued task thumb not on fs:", "id", thumbs.ID, "queue", thumbs.Queue, "state", thumbs.State)
	}
}

func newThumbTask(arch *repository.Archive, client *asynq.Client) (*asynq.TaskInfo, error) {
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
