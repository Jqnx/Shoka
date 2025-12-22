// Package server implements all logic concerning the HTTP server
// e.g. routes, handlers, middleware
package server

import (
	"fmt"
	"net/http"
	"time"

	"Shoka/internal/config"
	"Shoka/internal/logger"
	"Shoka/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int64
	repo *repository.Queries
	db   *pgxpool.Pool
	log  logger.Logger
	app  *config.App
	// dm   *downloader.Manager
}

func NewServer(app *config.App) *http.Server {
	NewServer := &Server{
		port: app.Cfg.Server.Port,
		repo: app.Repo,
		db:   app.DB,
		log:  app.Log,
		app:  app,
		// dm:   dm,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
