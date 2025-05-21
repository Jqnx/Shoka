package workers

import (
	"fmt"

	"github.com/hibiken/asynq"
)

func (w *Workers) NewAsynqClient() *asynq.Client {
	url := fmt.Sprintf("%v:%v", w.app.Cfg.Workers.RedisHost, w.app.Cfg.Workers.RedisPort)
	opt := asynq.RedisClientOpt{
		Addr: url,
	}
	client := asynq.NewClient(opt)
	return client
}
