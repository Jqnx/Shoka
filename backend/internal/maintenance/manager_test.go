package maintenance

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"Shoka/internal/jobs"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestAlreadyDone(t *testing.T) {
	t.Parallel()

	processor := image.NewProcessor(t.TempDir(), slog.New(slog.DiscardHandler))
	manager := &Manager{processor: processor}

	if err := os.MkdirAll(processor.ThumbDir("has-cover"), 0o750); err != nil {
		t.Fatalf("make thumb dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(processor.ThumbDir("has-cover"), "001.webp"), nil, 0o600); err != nil {
		t.Fatalf("write cover thumbnail: %v", err)
	}

	hash := int64(1)

	tests := []struct {
		name string
		task string
		row  sqlc.GetArchivesForMaintenanceRow
		want bool
	}{
		{"cover present", jobs.JobTypeCover, sqlc.GetArchivesForMaintenanceRow{ID: "has-cover", PageCount: 1}, true},
		{"cover missing", jobs.JobTypeCover, sqlc.GetArchivesForMaintenanceRow{ID: "no-cover", PageCount: 1}, false},
		{"cover on another page", jobs.JobTypeCover, sqlc.GetArchivesForMaintenanceRow{ID: "has-cover", CoverPage: 1, PageCount: 3}, false},
		{"cover page out of range clamps to page 0", jobs.JobTypeCover, sqlc.GetArchivesForMaintenanceRow{ID: "has-cover", CoverPage: 5, PageCount: 2}, true},
		{"phash with one point", jobs.JobTypePHash, sqlc.GetArchivesForMaintenanceRow{ID: "hashed", PhashP25: &hash}, true},
		{"phash with no points", jobs.JobTypePHash, sqlc.GetArchivesForMaintenanceRow{ID: "unhashed"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := manager.alreadyDone(tt.task, tt.row); got != tt.want {
				t.Errorf("alreadyDone(%q, %+v) = %v, want %v", tt.task, tt.row, got, tt.want)
			}
		})
	}
}

const (
	doneArchive = "done"
	todoArchive = "todo"
)

func TestRunTaskSkipsCompletedWorkUnlessForced(t *testing.T) {
	t.Parallel()

	t.Run("scheduled cover only queues the missing one", func(t *testing.T) {
		t.Parallel()

		manager, db := newMaintenanceFixture(t)

		if err := manager.RunTask(t.Context(), jobs.JobTypeCover, false); err != nil {
			t.Fatalf("run scheduled covers: %v", err)
		}

		assertJobArchives(t, db, jobs.JobTypeCover, []string{todoArchive})
	})

	t.Run("scheduled phash only queues the unhashed one", func(t *testing.T) {
		t.Parallel()

		manager, db := newMaintenanceFixture(t)

		if err := manager.RunTask(t.Context(), jobs.JobTypePHash, false); err != nil {
			t.Fatalf("run scheduled phashes: %v", err)
		}

		assertJobArchives(t, db, jobs.JobTypePHash, []string{todoArchive})
	})

	t.Run("forced runs queue every archive", func(t *testing.T) {
		t.Parallel()

		manager, db := newMaintenanceFixture(t)

		if err := manager.RunTask(t.Context(), jobs.JobTypeCover, true); err != nil {
			t.Fatalf("run forced covers: %v", err)
		}

		if err := manager.RunTask(t.Context(), jobs.JobTypePHash, true); err != nil {
			t.Fatalf("run forced phashes: %v", err)
		}

		assertJobArchives(t, db, jobs.JobTypeCover, []string{doneArchive, todoArchive})
		assertJobArchives(t, db, jobs.JobTypePHash, []string{doneArchive, todoArchive})
	})

	t.Run("every attempt is stamped", func(t *testing.T) {
		t.Parallel()

		manager, db := newMaintenanceFixture(t)

		if err := manager.RunTask(t.Context(), jobs.JobTypeCover, false); err != nil {
			t.Fatalf("run scheduled covers: %v", err)
		}

		if err := manager.RunTask(t.Context(), jobs.JobTypePHash, false); err != nil {
			t.Fatalf("run scheduled phashes: %v", err)
		}

		var unstamped int

		if err := db.QueryRowContext(t.Context(),
			"SELECT count(*) FROM maintenance_task WHERE last_run_at IS NULL").Scan(&unstamped); err != nil {
			t.Fatalf("count stamps: %v", err)
		}

		if unstamped != 0 {
			t.Errorf("%d maintenance tasks were not stamped", unstamped)
		}
	})
}

// newMaintenanceFixture returns a manager over a fresh in-memory database with
// one completed archive ("done": hashed, cover thumbnail on disk) and one
// untouched archive ("todo").
func newMaintenanceFixture(t *testing.T) (*Manager, *sql.DB) {
	t.Helper()

	log := slog.New(slog.DiscardHandler)

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	// :memory: is per-connection, so the fixture would vanish behind the pool.
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), `
		CREATE TABLE archive (
			id TEXT PRIMARY KEY,
			file_path TEXT NOT NULL,
			cover_page INTEGER NOT NULL DEFAULT 0,
			page_count INTEGER NOT NULL DEFAULT 0,
			phash_p0 INTEGER,
			phash_p25 INTEGER,
			phash_p50 INTEGER,
			phash_p75 INTEGER
		);
		CREATE TABLE job (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			payload TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'pending'
				CHECK (status IN ('pending', 'running', 'done', 'failed')),
			attempts INTEGER NOT NULL DEFAULT 0,
			max_attempts INTEGER NOT NULL DEFAULT 3,
			error TEXT,
			created_at DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
			run_after DATETIME NOT NULL DEFAULT (datetime('now'))
		);
		CREATE TABLE maintenance_task (
			task TEXT PRIMARY KEY,
			last_run_at DATETIME
		);
		INSERT INTO archive (id, file_path, phash_p0, phash_p25, phash_p50, phash_p75)
		VALUES ('done', '/done.zip', 1, 2, 3, 4);
		INSERT INTO archive (id, file_path) VALUES ('todo', '/todo.zip');
		INSERT INTO maintenance_task (task) VALUES ('cover'), ('phash');
	`); err != nil {
		t.Fatalf("seed db: %v", err)
	}

	processor := image.NewProcessor(t.TempDir(), log)
	if err := os.MkdirAll(processor.ThumbDir("done"), 0o750); err != nil {
		t.Fatalf("make thumb dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(processor.ThumbDir("done"), "001.webp"), nil, 0o600); err != nil {
		t.Fatalf("write cover thumbnail: %v", err)
	}

	queries := sqlc.New(db)

	return NewManager(queries, jobs.NewQueue(queries, log), processor, log), db
}

// assertJobArchives compares the archive ids of a type's queued jobs, sorted,
// so the assertion is independent of enqueue order.
func assertJobArchives(t *testing.T, db *sql.DB, jobType string, want []string) {
	t.Helper()

	rows, err := db.QueryContext(t.Context(), "SELECT payload FROM job WHERE type = ?", jobType)
	if err != nil {
		t.Fatalf("query %s jobs: %v", jobType, err)
	}
	defer func() { _ = rows.Close() }()

	var got []string

	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			t.Fatalf("scan payload: %v", err)
		}

		var payload struct {
			ArchiveID string `json:"archive_id"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("decode payload %q: %v", raw, err)
		}

		got = append(got, payload.ArchiveID)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate jobs: %v", err)
	}

	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Errorf("%s jobs = %v, want %v", jobType, got, want)
	}
}
