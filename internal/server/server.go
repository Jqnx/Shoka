package server

import (
	"Shoka/internal/database"
	"Shoka/internal/store"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	port   int
	store  store.Storage
	db     *gorm.DB
	logger *zap.SugaredLogger
}

func NewServer() *http.Server {
	logger := zap.Must(zap.NewProduction()).Sugar()
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	db := database.New()
	logger.Info("Connected to database.")
	store := store.NewStorage(db)
	NewServer := &Server{
		port:   port,
		store:  store,
		db:     db,
		logger: logger,
	}

	// Auto Migrations
	store.Archive.Migrate()
	store.Tags.Migrate()
	store.Artist.Migrate()
	store.Group.Migrate()
	logger.Info("Migrations done.")

	// Scan for files
	logger.Info("Scanning for files...")
	err := store.Archive.CreateFromFile()
	if err != nil {
		logger.Error(err)
	} else {
		logger.Info("Scanned Successfully.")
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
