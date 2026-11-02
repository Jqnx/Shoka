package database

import (
	"Shoka/internal/database/sqlc"
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// openAt opens a second handle on the same database file, standing in for a job
// worker or another request.
func openAt(t *testing.T, path string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", connectionString(path))
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}

	t.Cleanup(func() { _ = db.Close() })

	return db
}

// startIncrement runs a counter increment on another connection and reports
// whether it committed.
func startIncrement(ctx context.Context, db *sql.DB) <-chan error {
	done := make(chan error, 1)

	go func() {
		_, err := db.ExecContext(ctx, "update artist set count = count + 1 where id = 1")
		done <- err
	}()

	return done
}

// TestConnectionStringReadBeforeWrite covers the property database.DeleteArchive
// and database.DeleteLibrary depend on: both read the relation ids they are
// about to decrement - through queries that consume every row, like the
// generated sqlc ones - and only then write, so their transaction promotes a
// read snapshot into a write one.
//
// On a deferred BEGIN (no _txlock) another connection committing in between
// invalidates that snapshot and the promotion fails with "database is locked";
// busy_timeout cannot retry it, so a user-facing DELETE would return 500
// whenever a running job wrote concurrently. With the immediate transactions
// this DSN asks for, the write lock is held from BEGIN, the concurrent writer
// waits its turn, and both statements succeed.
func TestConnectionStringReadBeforeWrite(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "shoka.db")

	db := openAt(t, path)

	if _, err := db.ExecContext(ctx, taxonomySchema+taxonomyFixture); err != nil {
		t.Fatalf("schema: %v", err)
	}

	other := openAt(t, path)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	defer func() { _ = tx.Rollback() }()

	// The read half of a delete, through the generated query: drained rows.
	ids, err := sqlc.New(db).WithTx(tx).ListArchiveIDsByLibrary(ctx, libraryOne)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	if len(ids) == 0 {
		t.Fatalf("fixture has no archives in %s", libraryOne)
	}

	concurrent := startIncrement(ctx, other)

	select {
	case err := <-concurrent:
		// Only reachable without the write lock: the increment got in first,
		// so the decrement below has to survive a snapshot that is now stale.
		if err != nil {
			t.Fatalf("concurrent increment: %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		// Expected: the increment is queued behind our transaction's write lock.
	}

	// The write half of a delete.
	if _, err := tx.ExecContext(ctx, "update artist set count = count - 1 where id in (1)"); err != nil {
		t.Fatalf("write after the read: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if err := <-concurrent; err != nil {
		t.Fatalf("concurrent increment: %v", err)
	}

	// Seeded 5, minus this transaction's decrement, plus the queued increment.
	if got := countRows(t, db, "select count from artist where id = 1"); got != 5 {
		t.Errorf("artist.count = %d, want 5", got)
	}
}
