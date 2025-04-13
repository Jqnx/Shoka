package main

import (
	"Shoka/internal/config"
	"Shoka/internal/database"
	"Shoka/internal/logger"
	"Shoka/internal/repository"
	"Shoka/internal/server"
	"Shoka/internal/workers"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/riverqueue/river"
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
	log := logger.NewSlog()

	// Loading Config
	config := config.LoadConfig()
	log.Info("Config Loaded")

	// Creating db connection pool & connecting to db
	pool := database.NewPool(ctx, config, log)
	db := database.NewConn(ctx, pool, log)
	log.Info("Connected to database.")

	// Initializing new repository
	repo := repository.New(db)
	log.Info("New repository initialized.")

	// Start workers
	boys := river.NewWorkers()
	workers.NewScan(ctx, log, pool, config, repo, boys)
	fmt.Println(boys)

	// Starting web server
	log.Info(fmt.Sprintf("starting server on :%v", config.Server.Port))
	server := server.NewServer(
		config,
		db,
		repo,
		log,
		// workerui,
	)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Info("Graceful shutdown complete.")
}
