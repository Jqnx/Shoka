package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/downloader"
	"Shoka/internal/fsutil"
	"Shoka/internal/logger"
	"Shoka/internal/notifier"
	"Shoka/internal/repository"
	"Shoka/internal/server"
	"Shoka/internal/websocket"
	"Shoka/internal/workers"
	"Shoka/internal/workers/tasks"

	"github.com/cavaliergopher/grab/v3"
	"github.com/hibiken/asynq"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	ctx := context.Background()

	// Starting new logger
	log := logger.NewLogger(config.Logger)

	// Loading Config
	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Fatal("failed to load config", "error", err)
	}
	log.Info("Config Loaded")

	// Create necessary directories
	if err := fsutil.CreateDirs(cfg); err != nil {
		log.Error("failed to create directories", "error", err)
	}

	// Creating db connection pool & connecting to db
	db, err := database.NewPool(ctx, cfg, log)
	if err != nil {
		log.Fatal("db: failed to connect to database", "error", err)
	}
	defer db.Close()
	log.Info("Connected to database.")

	// Setup listener
	li := notifier.NewListener(db)
	if err := li.Connect(ctx); err != nil {
		log.Fatal("listener: error connecting to database", "error", err)
	}

	// Setup notifier
	noti := notifier.NewNotifier(log, li)
	go noti.Run(ctx)

	// Initializing new repository
	repo := repository.New(db)
	log.Info("New repository initialized.")

	srv, err := workers.NewServer(cfg)
	if err != nil {
		log.Fatal("error connecting to redis", "error", err)
	}
	log.Info("Connected to redis.")

	// Start new Asynq client
	client, err := workers.NewAsynqClient(cfg)
	if err != nil {
		log.Fatal("error connecting to redis", "error", err)
	}

	// Initialize new WebSocket Hub
	hub := websocket.NewHub(log)
	log.Info("Initialized new WebSocket Hub.")

	grab := grab.NewClient()
	log.Info("Initialized new Grab client.")

	// Initializing new App
	app := config.NewApp(repo, log, db, cfg, noti, client, hub, grab)
	defer app.Close()

	dm := downloader.NewDownloadManager(&app)
	defer dm.Close()

	// Starting web server
	log.Info(fmt.Sprintf("starting server on :%v", cfg.Server.Port))
	server := server.NewServer(&app, dm)

	// Defining asynq handlers
	mux := asynq.NewServeMux()
	mux.Handle(tasks.TypeCreateArchive, tasks.NewArchiveProcessor(cfg, &app))
	mux.Handle(tasks.TypeCreateCover, tasks.NewCoverProcessor(&app))
	mux.Handle(tasks.TypeCreateThumbnail, tasks.NewThumbnailProcessor(&app))
	mux.Handle(tasks.TypeNewMetadata, tasks.NewMetadataProcessor(&app))
	mux.Handle(downloader.TypeDownload, downloader.NewDownloadProcessor(dm))

	// Start asynq server
	go func() {
		if err := srv.Run(mux); err != nil {
			log.Fatal(err.Error())
		}
	}()

	// Start workers
	w := workers.NewWorkers(&app, false, ctx)
	go w.NewArchives()
	go w.Cleanup()

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Info("Graceful shutdown complete.")
}
