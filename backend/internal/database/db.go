package database

import (
	"Shoka/internal/config"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pressly/goose/v3"

	_ "github.com/mattn/go-sqlite3" // registers the "sqlite3" driver for database/sql
)

// DatabaseFile is the name of the SQLite database file stored in config.DataDir.
var DatabaseFile = "shoka.db"

// New opens the SQLite database in config.DataDir, enabling WAL mode, foreign
// keys, and a busy timeout, and verifies the connection is usable.
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

// Migrate applies all embedded goose migrations to the given database.
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
