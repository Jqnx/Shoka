package workers

import (
	"Shoka/internal/config"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

func NewServer(cfg *config.Config) (*asynq.Server, error) {
	switch {
	case cfg.Workers.RedisHost == "":
		return nil, config.ErrNoRedisHost
	case cfg.Workers.RedisPort == "":
		return nil, config.ErrNoRedisPort
	}

	url := fmt.Sprintf("%v:%v", cfg.Workers.RedisHost, cfg.Workers.RedisPort)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: url},
		asynq.Config{
			Concurrency: cfg.Workers.Max,
			Queues: map[string]int{
				"critical": cfg.Workers.Max,
				"default":  cfg.Workers.Max,
				"low":      cfg.Workers.Max,
			},
			StrictPriority: true,
			RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
				return time.Second * 2
			},
		})
	return srv, nil
}
