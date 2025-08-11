package config

import (
	"Shoka/internal/logger"
	"Shoka/internal/notifier"
	"Shoka/internal/repository"
	"Shoka/internal/websocket"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Repo   *repository.Queries
	Log    logger.Logger
	DB     *pgxpool.Pool
	Cfg    *Config
	Noti   notifier.Notifier
	Client *asynq.Client
	Hub    *websocket.Hub
}

func NewApp(repo *repository.Queries,
	log logger.Logger,
	db *pgxpool.Pool,
	cfg *Config,
	noti notifier.Notifier,
	client *asynq.Client,
	hub *websocket.Hub,
) App {
	return App{
		Repo:   repo,
		Log:    log,
		DB:     db,
		Cfg:    cfg,
		Noti:   noti,
		Client: client,
		Hub:    hub,
	}
}

func (a *App) Close() {
	if a.Client != nil {
		a.Client.Close()
	}
}
