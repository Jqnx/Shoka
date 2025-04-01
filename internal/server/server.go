package server

import (
	"Shoka/internal/database"
	"Shoka/internal/logger"
	"Shoka/internal/repository"
	"Shoka/internal/store"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/joho/godotenv/autoload"
)

type Server struct {
	port  int
	store store.Storage
	repo  *repository.Queries
	db    *pgx.Conn
	log   logger.Logger
}

func NewServer() *http.Server {
	l := logger.Get()
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	l.Logger.Info().Msg("Connected to database.")
	repo := repository.New(db)
	NewServer := &Server{
		port: port,
		// store:  store,
		repo: repo,
		db:   db,
		log:  l,
	}

	// Auto Migrations
	// store.Migrate.Migrate()
	// logger.Info("Migrations done.")

	// Scan for files
	//logger.Info("Scanning for files...")
	//err := store.Archive.CreateFromFile()
	//if err != nil {
	//	logger.Error(err)
	//} else {
	//	logger.Info("Scanned Successfully.")
	//}

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
