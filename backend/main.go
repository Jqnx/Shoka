package main

import (
	"Shoka/internal/api"
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	"Shoka/internal/log"
	"Shoka/internal/util"
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/davidbyttow/govips/v2/vips"
)

func main() {
	// Create new Logger
	log := log.New()

	if err := util.EnsureDir("./data"); err != nil {
		log.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}
	log.Info("config loaded")

	// Connect to database
	db, err := database.New()
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	if err := database.Migrate(db); err != nil {
		log.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}
	queries := sqlc.New(db)
	log.Info("connected to database")

	// Initialize Services
	vips.LoggingSettings(nil, vips.LogLevelWarning)
	vips.Startup(nil)
	defer vips.Shutdown()
	queue := jobs.NewQueue(queries, log)
	worker := jobs.NewWorker(queue, log)
	scanner := library.NewScanner(queries, queue, log, cfg.LibraryDir)
	watcher := library.NewWatcher(queue, cfg.LibraryDir, log)
	images := image.NewProcessor(cfg.CacheDir, log)
	cache, err := image.NewCache(images, log)
	if err != nil {
		log.Error("failed to initialize image cache", "error", err)
		os.Exit(1)
	}
	api := api.New(queries, log, queue, cache, images)

	// Register worker handlers
	worker.Register(jobs.JobTypeScan, jobs.NewScanHandler(scanner, log), 1)
	worker.Register(jobs.JobTypeCover, jobs.NewCoverHandler(images, log), 5)
	worker.Register(jobs.JobTypeThumbnail, jobs.NewThumbnailHandler(images, log), 3)

	// Start Workers with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for range 10 {
		worker.Start(ctx)
	}

	// Start Initial Scan and File Watcher
	if err := queue.EnqueueOnce(ctx, jobs.JobTypeScan, jobs.ScanPayload{}); err != nil {
		log.Error("failed to enqueue initial scan", "error", err)
	}
	watcher.Start(ctx)

	// Start API Server
	host := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info(fmt.Sprintf("listening on %s", host))
	if err := http.ListenAndServe(host, api.Router); err != nil {
		log.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
