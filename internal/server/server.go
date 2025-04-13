package server

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
	"riverqueue.com/riverui"
)

type Server struct {
	port     int
	repo     *repository.Queries
	db       *pgx.Conn
	log      *slog.Logger
	workerui *riverui.Server
}

func NewServer(
	config *config.Config,
	db *pgx.Conn,
	repo *repository.Queries,
	log *slog.Logger,
	// workerui *riverui.Server,

) *http.Server {
	port, _ := strconv.Atoi(config.Server.Port)
	NewServer := &Server{
		port: port,
		repo: repo,
		db:   db,
		log:  log,
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
