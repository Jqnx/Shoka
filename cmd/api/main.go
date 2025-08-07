package main

import (
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/logger"
	"Shoka/internal/notifier"
	"Shoka/internal/repository"
	"Shoka/internal/server"
	"Shoka/internal/workers"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

// TODO: Create helper function to create every necessary directory

func main() {
	ctx := context.Background()

	// Starting new logger
	log := logger.NewSlog()

	// Loading Config
	cfg, err := config.LoadConfig(log)
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	log.Info("Config Loaded")

	// Creating db connection pool & connecting to db
	db, err := database.NewPool(ctx, cfg, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to database.")

	// Setup listener
	li := notifier.NewListener(db)
	if err := li.Connect(ctx); err != nil {
		log.Error("error connecting to database", "err", err)
		os.Exit(1)
	}

	// Setup notifier
	noti := notifier.NewNotifier(log, li)
	go noti.Run(ctx)

	// Initializing new repository
	repo := repository.New(db)
	log.Info("New repository initialized.")

	// Initializing new App
	app := config.NewApp(repo, log, db, cfg, noti)

	// Starting web server
	log.Info(fmt.Sprintf("starting server on :%v", cfg.Server.Port))
	server := server.NewServer(&app)

	// TODO: Add worker here in goroutine instead of its own binary

	// Start workers
	w := workers.NewWorkers(&app, false, ctx)
	go w.NewArchives()

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
