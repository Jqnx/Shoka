// Package workers contains all the logic regarding background workers
// and the queuing of their tasks.
package workers

import (
	"fmt"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/repository"

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

func Init(cfg *config.Config) (*asynq.Server, *asynq.Client, error) {
	switch {
	case cfg.Workers.RedisHost == "":
		return nil, nil, config.ErrNoRedisHost
	case cfg.Workers.RedisPort == "":
		return nil, nil, config.ErrNoRedisPort
	}

	url := fmt.Sprintf("%v:%v", cfg.Workers.RedisHost, cfg.Workers.RedisPort)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: url},
		asynq.Config{
			Concurrency: int(cfg.Workers.Max),
			Queues: map[string]int{
				"critical": int(cfg.Workers.Max),
				"default":  int(cfg.Workers.Max),
				"low":      int(cfg.Workers.Max),
			},
			StrictPriority: true,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Second * 2
			},
		})

	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr: url,
	})
	return srv, client, nil
}
