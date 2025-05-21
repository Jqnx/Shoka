package main

import (
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/logger"
	"Shoka/internal/notifier"
	"Shoka/internal/repository"
	"Shoka/internal/tasks"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

// TODO:
// Graceful shutdown

func main() {
	ctx := context.Background()

	// Starting new logger
	log := logger.NewSlog()

	// Loading Config
	err := godotenv.Load()
	if err != nil {
		log.Error("Error loading .env file")
		os.Exit(1)
	}
	cfg := config.LoadConfig()
	log.Info("Config Loaded")

	// Creating db connection pool & connecting to db
	// db := database.NewConn(ctx, cfg, log)
	db := database.NewPool(ctx, cfg, log)
	log.Info("Connected to database.")

	// Setup listener
	li := notifier.NewListener(db)
	if err := li.Connect(ctx); err != nil {
		panic(err)
	}

	noti := notifier.NewNotifier(log, li)
	go noti.Run(ctx)

	// Initializing new repository
	repo := repository.New(db)
	log.Info("New repository initialized.")

	app := config.NewApp(repo, log, db, cfg, noti)

	// Initializing asynq server
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

	// Defining asynq handlers
	mux := asynq.NewServeMux()
	mux.Handle(tasks.TypeScan, tasks.NewScanProcessor(cfg, &app))
	mux.Handle(tasks.TypeCreateArchive, tasks.NewArchiveProcessor(cfg, &app))
	mux.Handle(tasks.TypeCreateCover, tasks.NewCoverProcessor(&app))
	mux.Handle(tasks.TypeCreateThumbnail, tasks.NewThumbnailProcessor(&app))

	// Start asynq server
	if err := srv.Run(mux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
