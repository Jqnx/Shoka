// Package workers contains all the logic regarding background workers
// and the queuing of their tasks.
package workers

import (
	"fmt"
	"time"

	"Shoka/internal/config"

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
	return srv, nil
}
