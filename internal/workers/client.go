package workers

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"fmt"

	"github.com/hibiken/asynq"
)

type Client struct {
	client *asynq.Client
	app    *config.App
	arch   *repository.GetArchiveByIDRow
}

func NewClient(client *asynq.Client, app *config.App, arch *repository.GetArchiveByIDRow) *Client {
	return &Client{
		client: client,
		app:    app,
		arch:   arch,
	}
}

func NewAsynqClient(cfg *config.Config) (*asynq.Client, error) {
	switch {
	case cfg.Workers.RedisHost == "":
		return nil, config.ErrNoRedisHost
	case cfg.Workers.RedisPort == "":
		return nil, config.ErrNoRedisPort
	}

	url := fmt.Sprintf("%v:%v", cfg.Workers.RedisHost, cfg.Workers.RedisPort)
	opt := asynq.RedisClientOpt{
		Addr: url,
	}
	client := asynq.NewClient(opt)
	return client, nil
}
