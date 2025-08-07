package config

import (
	"Shoka/internal/logger"
	"Shoka/internal/notifier"
	"Shoka/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Repo *repository.Queries
	Log  logger.Logger
	DB   *pgxpool.Pool
	Cfg  *Config
	Noti notifier.Notifier
}

func NewApp(repo *repository.Queries, log logger.Logger, db *pgxpool.Pool, cfg *Config, noti notifier.Notifier) App {
	return App{
		Repo: repo,
		Log:  log,
		DB:   db,
		Cfg:  cfg,
		Noti: noti,
	}
}
