package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"Shoka/internal/api"
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	"Shoka/internal/log"
	"Shoka/internal/metadata"
	"Shoka/internal/metadata/sources"
	"Shoka/internal/util"

	"github.com/davidbyttow/govips/v2/vips"
)

// @title Shoka API
// @version 1.0
// @description Self-hosted doujinshi library manager API
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Create new Logger
	log := log.New(cfg)

	if err := util.EnsureDir("./data"); err != nil {
		log.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

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
	libraries := library.NewManager(cfg, queries, queue, log)
	images := image.NewProcessor(cfg.Cache.Dir, log)
	cache, err := image.NewCache(images, log)
	pipeline := metadata.NewPipeline(
		log,
		queries,
		sources.NewComicInfoSource(),
		sources.NewFilenameSource(),
		sources.NewEHentaiSource(),
		sources.NewNHentaiSource(),
	)
	if err != nil {
		log.Error("failed to initialize image cache", "error", err)
		os.Exit(1)
	}
	api := api.New(queries, log, queue, cache, images, pipeline, libraries, db)

	// Register worker handlers
	worker.Register(jobs.JobTypeScan, jobs.NewScanHandler(libraries, log), 1)
	worker.Register(jobs.JobTypeCover, jobs.NewCoverHandler(images, log), 5)
	worker.Register(jobs.JobTypeThumbnail, jobs.NewThumbnailHandler(images, log), 3)
	worker.Register(jobs.JobTypeMetadata, jobs.NewMetadataHandler(pipeline, queries, db, queue, log), 5)
	worker.Register(jobs.JobTypeMetadataRemote, jobs.NewRemoteMetadataHandler(pipeline, queries, db, log), 1)

	// Start Workers with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for range 10 {
		worker.Start(ctx)
	}

	// Start watching every enabled library and enqueue an initial scan for each
	if err := libraries.Start(ctx); err != nil {
		log.Error("failed to start library watchers", "error", err)
	}
	if err := libraries.EnqueueInitialScans(ctx); err != nil {
		log.Error("failed to enqueue initial scans", "error", err)
	}

	// Start API Server
	host := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info(fmt.Sprintf("listening on %s", host))
	if err := http.ListenAndServe(host, api.Router); err != nil {
		log.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
