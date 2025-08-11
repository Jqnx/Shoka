package downloader

import (
	"Shoka/internal/config"
	"Shoka/internal/logger"
	"Shoka/internal/repository"
	"Shoka/internal/websocket"
	"context"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	StatusPending     = "pending"
	StatusDownloading = "downloading"
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
	StatusCancelled   = "cancelled"
)

type Manager struct {
	ActiveDownloads map[string]*ActiveDownload
	ActiveMutex     sync.RWMutex
	hub             *websocket.Hub
	asynqClient     *asynq.Client
	queries         *repository.Queries
	pool            *pgxpool.Pool
	log             logger.Logger
	cfg             *config.Config
}

type ActiveDownload struct {
	ID                 string
	Cancel             context.CancelFunc
	Progress           int32
	LastProgressUpdate time.Time
	Cancelled          bool
	FilePath           string
}

func NewDownloadManager(app *config.App) *Manager {
	dm := &Manager{
		log:             app.Log,
		ActiveDownloads: make(map[string]*ActiveDownload),
		hub:             app.Hub,
		asynqClient:     app.Client,
		queries:         app.Repo,
		pool:            app.DB,
		cfg:             app.Cfg,
	}

	// Load existing downloads from database
	if err := dm.reQueueInterruptedDownloads(); err != nil {
		dm.log.Error("Failed to re-queue interrupted downloads", "error", err)
	}

	return dm
}
