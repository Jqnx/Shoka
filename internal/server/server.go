package server

import (
	"Shoka/internal/config"
	"Shoka/internal/repository"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port int
	repo *repository.Queries
	db   *pgxpool.Pool
	log  *slog.Logger
	app  *config.App
}

func NewServer(app *config.App) *http.Server {
	NewServer := &Server{
		port: app.Cfg.Server.Port,
		repo: app.Repo,
		db:   app.DB,
		log:  app.Log,
		app:  app,
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
