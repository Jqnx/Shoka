package database

import (
	"Shoka/internal/config"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

var DatabaseFile string = "shoka.db"

func New() (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_fk=ON&_busy_timeout=5000", filepath.Join(config.DataDir, DatabaseFile))

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
