package workers

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"

	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
	app    *config.App
	arch   *repository.Archive
}

func NewClient(client *asynq.Client, app *config.App, arch *repository.Archive) *Client {
	return &Client{
		client: client,
		app:    app,
		arch:   arch,
	}
}
