package downloader

/*
const (
	StatusPending     = "pending"
	StatusDownloading = "downloading"
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
	StatusCancelled   = "cancelled"
	StatusPaused      = "paused"
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
	grab            *grab.Client
}

type ActiveDownload struct {
	ID                 string
	Cancel             context.CancelFunc
	Progress           int32
	LastProgressUpdate time.Time
	Cancelled          bool
	FilePath           string
	BytesDownloaded    int64
	TotalSize          int64
	LastSpeedUpdate    time.Time
	LastBytesCount     int64
	DownloadSpeed      int64
	StartTime          time.Time
	Response           *grab.Response
	CanResume          bool
	ResumeSupported    bool
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
		grab:            app.Grab,
	}

	// Load existing downloads from database
	if err := dm.reQueueInterruptedDownloads(); err != nil {
		dm.log.Error("Failed to re-queue interrupted downloads", "error", err)
	}

	return dm
}
*/
