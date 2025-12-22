package workers

import (
	"Shoka/internal/repository"
	"Shoka/internal/workers/tasks"

	"github.com/hibiken/asynq"
)

func (w *Workers) Metadata(arch *repository.GetArchiveByIDRow, src string) {
	c := NewClient(w.app.Client, w.app, arch)
	c.NewMetadata(src)
}

// TODO: Somehow get metadata from inside job to outside client/worker
func (c *Client) NewMetadata(src string) {
	meta, err := newMetadataTask(c.arch, src, c.client)
	if err != nil {
		c.app.Log.Error("could not create task:", "error", err.Error())
	}
	c.app.Log.Info("queued task new metadata:", "id", meta.ID, "queue", meta.Queue, "state", meta.State)
}

func newMetadataTask(arch *repository.GetArchiveByIDRow, src string, client *asynq.Client) (*asynq.TaskInfo, error) {
	newMeta, err := tasks.NewMetadataTask(arch, src)
	if err != nil {
		return nil, err
	}
	meta, err := client.Enqueue(newMeta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
