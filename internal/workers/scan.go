package workers

import (
	"Shoka/internal/archive"
	"Shoka/internal/config"
	"Shoka/internal/fsutil"
	"Shoka/internal/repository"
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type ScanArgs struct {
	Dir string `json:"dir"`
}

func (ScanArgs) Kind() string { return "scan" }

type ScanWorker struct {
	query *repository.Queries
	log   *slog.Logger
	river.WorkerDefaults[ScanArgs]
}

func (w *ScanWorker) Work(ctx context.Context, job *river.Job[ScanArgs]) error {
	list := fsutil.ListArchives(job.Args.Dir)
	for _, item := range list {
		err := archive.CreateFromFile(ctx, item, w.query, w.log)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewScan(
	ctx context.Context,
	log *slog.Logger,
	db *pgxpool.Pool,
	cfg *config.Config,
	q *repository.Queries,
	workers *river.Workers,
) error {
	river.AddWorker(workers, &ScanWorker{
		query: q,
		log:   log,
	})

	client, err := river.NewClient(riverpgxv5.New(db), &river.Config{
		Logger: log,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: cfg.Workers.Max},
		},
		Workers: workers,
	})
	if err != nil {
		log.ErrorContext(ctx, err.Error())
		panic(err)
	}

	defer func() {
		if err := client.Stop(ctx); err != nil {
			log.ErrorContext(ctx, err.Error())
			panic(err)
		}
	}()

	if err := client.Start(ctx); err != nil {
		log.ErrorContext(ctx, err.Error())
		panic(err)
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		log.ErrorContext(ctx, err.Error())
		panic(err)
	}
	defer tx.Rollback(ctx)

	_, err = client.InsertTx(ctx, tx, &ScanArgs{Dir: cfg.Dir}, nil)
	if err != nil {
		log.ErrorContext(ctx, err.Error())
		panic(err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.ErrorContext(ctx, err.Error())
		panic(err)
	}

	return nil
}
