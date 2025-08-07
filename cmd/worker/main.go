package main

import (
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/logger"
	"Shoka/internal/notifier"
	"Shoka/internal/repository"
	"Shoka/internal/workers/tasks"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
)

// TODO: Graceful shutdown

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

	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	log.Info("Config Loaded")

	// Creating db connection pool & connecting to db
	db, err := database.NewPool(ctx, cfg, log)
	if err != nil {
		log.Error("db: failed to connect to database", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to database.")

	// Setup listener
	li := notifier.NewListener(db)
	if err := li.Connect(ctx); err != nil {
		log.Error("listener: error connecting to database", "error", err)
		os.Exit(1)
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
	mux.Handle(tasks.TypeCreateArchive, tasks.NewArchiveProcessor(cfg, &app))
	mux.Handle(tasks.TypeCreateCover, tasks.NewCoverProcessor(&app))
	mux.Handle(tasks.TypeCreateThumbnail, tasks.NewThumbnailProcessor(&app))
	mux.Handle(tasks.TypeNewMetadata, tasks.NewMetadataProcessor(&app))

	// Start asynq server
	if err := srv.Run(mux); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
