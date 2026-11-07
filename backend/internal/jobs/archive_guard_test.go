package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"Shoka/internal/metadata"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// The guard only asks whether the archive row exists, so its fixture mirrors
// that single column rather than all 18 of the real archive table.
const archiveIDSchema = `
CREATE TABLE archive (id TEXT PRIMARY KEY NOT NULL);
`

// The metadata handlers read the whole row (GetArchiveByID), so their fixture
// mirrors every column that query scans, in table-definition order because
// SELECT * returns them that way: 00001 plus the columns added by 00021
// (library_id), 00029/00030 (phash_p0..p75) and 00034 (cover_page).
const archiveRowSchema = `
CREATE TABLE archive (
    id TEXT PRIMARY KEY NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    language TEXT,
    category TEXT,
    page_count INTEGER NOT NULL default 0,
    file_path TEXT UNIQUE NOT NULL,
    file_size INTEGER NOT NULL,
    mod_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL default (datetime('now')),
    updated_at DATETIME NOT NULL default (datetime('now')),
    release_date DATETIME,
    library_id TEXT NOT NULL,
    phash_p0 INTEGER,
    phash_p25 INTEGER,
    phash_p50 INTEGER,
    phash_p75 INTEGER,
    cover_page INTEGER NOT NULL DEFAULT 0
);
`

// liveID is the archive the fixture holds; goneID never exists, standing in
// for one a library or archive delete cascaded away.
const (
	liveID = "live-archive"
	goneID = "deleted-archive"
)

// missingFile is the payload FilePath: nothing is there, so a handler that
// gets past the guard fails at archive.Open instead of doing real work.
const missingFile = "/nonexistent/archive.zip"

// guardCase is one guarded handler: how to build it, and a payload targeting
// an archive id.
type guardCase struct {
	name string
	// reclaimsLeftover reports whether the handler also evicts a cache
	// directory written for an archive whose delete already ran. Only the
	// cache-writing handlers do; phash writes none.
	reclaimsLeftover bool
	newHandler       func(processor *image.Processor, queries *sqlc.Queries) Handler
	payload          func(id string) any
}

// guardCases covers every handler the archive guard was added to, except the
// metadata pair, which checks the row it already reads.
func guardCases() []guardCase {
	return []guardCase{
		{
			name:             "cover",
			reclaimsLeftover: true,
			newHandler: func(processor *image.Processor, queries *sqlc.Queries) Handler {
				return NewCoverHandler(processor, queries, testLogger())
			},
			payload: func(id string) any { return CoverPayload{ArchiveID: id, FilePath: missingFile} },
		},
		{
			name:             "thumbnail",
			reclaimsLeftover: true,
			newHandler: func(processor *image.Processor, queries *sqlc.Queries) Handler {
				return NewThumbnailHandler(processor, events.NewThumbnailBroadcaster(), queries, testLogger())
			},
			payload: func(id string) any { return ThumbnailPayload{ArchiveID: id, FilePath: missingFile} },
		},
		{
			name: "phash",
			newHandler: func(processor *image.Processor, queries *sqlc.Queries) Handler {
				return NewPHashHandler(processor, queries, testLogger())
			},
			payload: func(id string) any { return PHashPayload{ArchiveID: id, FilePath: missingFile} },
		},
	}
}

// TestArchiveGuardRunsLiveArchives pins the side of the guard that must not
// overreach: an archive that still exists is left to the job, and its cache
// directory is not touched.
func TestArchiveGuardRunsLiveArchives(t *testing.T) {
	t.Parallel()

	for _, tc := range guardCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := newJobsDB(t)
			processor := image.NewProcessor(t.TempDir(), testLogger())
			archiveDir := plantCache(t, processor, liveID)

			err := runGuardHandler(t, tc, db, processor, liveID)
			if err == nil {
				t.Fatal("handler returned nil, want the archive.Open failure that proves it did not skip a live archive")
			}

			if _, statErr := os.Stat(archiveDir); statErr != nil {
				t.Errorf("cache dir of a live archive: %v", statErr)
			}
		})
	}
}

// TestArchiveGuardSkipsDeletedArchives pins that a job whose archive is gone
// succeeds without opening the archive. The payload points at a file that does
// not exist, so an error here means the guard let the job through.
func TestArchiveGuardSkipsDeletedArchives(t *testing.T) {
	t.Parallel()

	for _, tc := range guardCases() {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := newJobsDB(t)
			processor := image.NewProcessor(t.TempDir(), testLogger())
			archiveDir := plantCache(t, processor, goneID)

			if err := runGuardHandler(t, tc, db, processor, goneID); err != nil {
				t.Fatalf("handler returned %v, want nil for a deleted archive", err)
			}

			_, statErr := os.Stat(archiveDir)

			switch {
			case tc.reclaimsLeftover && statErr == nil:
				t.Error("cache dir of the deleted archive survived; the job's deferred check must evict it")
			case !tc.reclaimsLeftover && statErr != nil:
				t.Errorf("handler evicted a cache dir it never writes: %v", statErr)
			}
		})
	}
}

// TestMetadataHandlersSkipDeletedArchive covers the metadata handlers' own
// version of the guard: they read the row they need, so a missing one is a
// skip rather than a scrape attempt that retries and ends up failed.
func TestMetadataHandlersSkipDeletedArchive(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		newHandler func(pipeline *metadata.Pipeline, queries *sqlc.Queries, db *sql.DB) Handler
	}{
		{
			name: "local",
			newHandler: func(pipeline *metadata.Pipeline, queries *sqlc.Queries, db *sql.DB) Handler {
				return NewMetadataHandler(pipeline, queries, db, NewQueue(queries, testLogger()), testLogger())
			},
		},
		{
			name: "remote",
			newHandler: func(pipeline *metadata.Pipeline, queries *sqlc.Queries, db *sql.DB) Handler {
				return NewRemoteMetadataHandler(pipeline, queries, db, testLogger())
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := newArchiveRowDB(t)
			queries := sqlc.New(db)
			log := testLogger()
			handler := tc.newHandler(metadata.NewPipeline(log, queries), queries, db)

			if err := handler(t.Context(), newJob(t, tc.name, MetadataPayload{ArchiveID: goneID})); err != nil {
				t.Fatalf("handler returned %v, want nil for a deleted archive", err)
			}
		})
	}
}

// newJobsDB returns an in-memory database holding the guard fixture: the
// archive table's id column and the single row for liveID.
func newJobsDB(t *testing.T) *sql.DB {
	t.Helper()

	db := openMemoryDB(t, archiveIDSchema)

	if _, err := db.ExecContext(t.Context(), "INSERT INTO archive (id) VALUES (?)", liveID); err != nil {
		t.Fatalf("seed archive: %v", err)
	}

	return db
}

// newArchiveRowDB returns an in-memory database holding the archive table as
// the metadata handlers read it, with no rows.
func newArchiveRowDB(t *testing.T) *sql.DB {
	t.Helper()

	return openMemoryDB(t, archiveRowSchema)
}

// openMemoryDB returns an in-memory database with schema applied.
func openMemoryDB(t *testing.T, schema string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })

	// :memory: is per-connection, so the fixture would vanish behind the pool.
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), schema); err != nil {
		t.Fatalf("schema: %v", err)
	}

	return db
}

// runGuardHandler encodes the case's payload the way the queue does and runs
// its handler once against the fixture database.
func runGuardHandler(t *testing.T, tc guardCase, db *sql.DB, processor *image.Processor, archiveID string) error {
	t.Helper()

	handler := tc.newHandler(processor, sqlc.New(db))

	return handler(t.Context(), newJob(t, tc.name, tc.payload(archiveID)))
}

// newJob builds the job row a handler would decode, with a payload it can be
// given without going through the queue.
func newJob(t *testing.T, jobType string, payload any) *Job {
	t.Helper()

	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	return &Job{ID: 1, Type: jobType, RawPayload: string(raw), Attempts: 1, MaxAttempts: 3}
}

// plantCache writes the thumbnail the processor would generate for archiveID
// and returns that archive's cache directory - the path a delete evicts, so a
// test can tell "left alone" from "reclaimed".
func plantCache(t *testing.T, processor *image.Processor, archiveID string) string {
	t.Helper()

	thumb := filepath.Join(processor.ThumbDir(archiveID), "001.webp")
	if err := os.MkdirAll(filepath.Dir(thumb), 0o750); err != nil {
		t.Fatalf("create thumbs dir: %v", err)
	}

	if err := os.WriteFile(thumb, []byte("webp"), 0o600); err != nil {
		t.Fatalf("write thumbnail: %v", err)
	}

	return filepath.Dir(processor.ThumbDir(archiveID))
}

// testLogger discards the handlers' own log output.
func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
