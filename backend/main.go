package main

import (
	"Shoka/internal/api"
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"Shoka/internal/library"
	applog "Shoka/internal/log"
	"Shoka/internal/metadata"
	"Shoka/internal/metadata/sources"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

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
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	log := applog.New(cfg)

	if err := run(cfg, log); err != nil {
		log.Error("fatal", "error", err)
		os.Exit(1)
	}
}

// run wires up every component and blocks serving the API. Returning an error
// (rather than calling os.Exit inline) lets the deferred db.Close and
// vips.Shutdown actually run on a startup failure.
func run(cfg *config.Config, log *slog.Logger) error {
	db, err := database.New()
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
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
	thumbnails := events.NewThumbnailBroadcaster()
	images := image.NewProcessor(cfg.Cache.Dir, log)

	cache, err := image.NewCache(images, log)
	if err != nil {
		return fmt.Errorf("initialize image cache: %w", err)
	}

	pipeline := metadata.NewPipeline(
		log,
		queries,
		sources.NewComicInfoSource(),
		sources.NewFilenameSource(),
		sources.NewEHentaiSource(),
		sources.NewNHentaiSource(),
	)

	apiServer := api.New(queries, log, queue, cache, images, pipeline, libraries, thumbnails, db)

	// Register worker handlers
	worker.Register(jobs.JobTypeScan, jobs.NewScanHandler(libraries, log), 1)
	worker.Register(jobs.JobTypeCover, jobs.NewCoverHandler(images, log), 5)
	worker.Register(jobs.JobTypePHash, jobs.NewPHashHandler(images, queries, log), 3)
	worker.Register(jobs.JobTypeThumbnail, jobs.NewThumbnailHandler(images, thumbnails, log), 3)
	worker.Register(jobs.JobTypeMetadata, jobs.NewMetadataHandler(pipeline, queries, db, queue, log), 5)
	worker.Register(jobs.JobTypeMetadataRemote, jobs.NewRemoteMetadataHandler(pipeline, queries, db, log), 1)

	// Start Workers with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for range 10 {
		worker.Start(ctx)
	}

	// Start watching every library and enqueue an initial scan for each
	if err := libraries.Start(ctx); err != nil {
		log.Error("failed to start library watchers", "error", err)
	}

	if err := libraries.EnqueueInitialScans(ctx); err != nil {
		log.Error("failed to enqueue initial scans", "error", err)
	}

	// Start API Server
	host := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Info("listening on " + host)

	srv := &http.Server{
		Addr:              host,
		Handler:           apiServer.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
