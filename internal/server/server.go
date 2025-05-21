package server

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"riverqueue.com/riverui"
)

type Server struct {
	port     int
	repo     *repository.Queries
	db       *pgxpool.Pool
	log      *slog.Logger
	workerui *riverui.Server
	app      *config.App
}

func NewServer(
	config *config.Config,
	db *pgxpool.Pool,
	repo *repository.Queries,
	log *slog.Logger,
	app *config.App,
	// workerui *riverui.Server,
) *http.Server {
	port, _ := strconv.Atoi(config.Server.Port)
	NewServer := &Server{
		port: port,
		repo: repo,
		db:   db,
		log:  log,
		app:  app,
		// workerui: workerui,
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
