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

// connectionString builds the DSN for a database file: WAL mode, foreign keys,
// a busy timeout, and immediate transactions.
//
// _txlock=immediate makes every BeginTx take the write lock up front. Without
// it a transaction that reads before it writes - database.DeleteArchive and
// database.DeleteLibrary both read the relation ids they are about to
// decrement - loses its snapshot to a concurrent writer and fails with "database
// is locked" instead of waiting, because the busy timeout cannot retry a
// snapshot that is already stale. Every transaction in this codebase writes,
// so taking the lock at BEGIN costs nothing. db_test.go pins that behaviour.
func connectionString(path string) string {
	return fmt.Sprintf("file:%s?_journal_mode=WAL&_fk=ON&_busy_timeout=5000&_txlock=immediate", path)
}

// New opens the SQLite database in config.DataDir and verifies the connection
// is usable.
func New() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connectionString(filepath.Join(config.DataDir, DatabaseFile)))
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
