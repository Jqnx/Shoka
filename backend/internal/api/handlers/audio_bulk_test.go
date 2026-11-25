package handlers

import (
	"Shoka/internal/auth"
	"Shoka/internal/database/sqlc"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// bulkAudioSchema mirrors the tables the audio bulk path touches. It is
// inline rather than database.Migrate's full schema because the migration
// runner expects the frontend-owned `user` table and FTS5.
const bulkAudioSchema = `
CREATE TABLE audio (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL,
    title TEXT NOT NULL,
    folder_path TEXT UNIQUE NOT NULL,
    mod_time DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    release_date DATETIME,
    language TEXT,
    cover_index INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE audio_played (
    audio_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    played_at DATETIME NOT NULL,
    completed INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (audio_id, user_id)
);

CREATE TABLE artist (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE tag (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE circle (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE voice_actor (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    count INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE artist_alias (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    artist_id INTEGER NOT NULL,
    alias TEXT UNIQUE NOT NULL
);

CREATE TABLE artist_circle (
    artist_id INTEGER,
    circle_id INTEGER,
    PRIMARY KEY (artist_id, circle_id)
);

CREATE TABLE audio_artist (
    audio_id TEXT NOT NULL,
    artist_id INTEGER NOT NULL,
    PRIMARY KEY (audio_id, artist_id)
);

CREATE TABLE audio_tag (
    audio_id TEXT NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (audio_id, tag_id)
);

CREATE TABLE audio_circle (
    audio_id TEXT NOT NULL,
    circle_id INTEGER NOT NULL,
    PRIMARY KEY (audio_id, circle_id)
);

CREATE TABLE audio_voice_actor (
    audio_id TEXT NOT NULL,
    voice_actor_id INTEGER NOT NULL,
    PRIMARY KEY (audio_id, voice_actor_id)
);
`

func newBulkAudioHandler(t *testing.T) (*AudioHandler, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "bulk.db"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	db.SetMaxOpenConns(1)

	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.ExecContext(context.Background(), bulkAudioSchema); err != nil {
		t.Fatalf("seed schema: %v", err)
	}

	h := NewAudioHandler(sqlc.New(db), db, nil, nil, nil, nil, slog.New(slog.DiscardHandler))

	return h, db
}

// seedBulkAudio inserts two works whose only shared relation is nothing: each
// has its own tag, so a merge's preservation is observable.
func seedBulkAudio(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx := context.Background()

	stmts := []string{
		`INSERT INTO audio (id, library_id, title, folder_path, mod_time)
		 VALUES ('a', 'lib', 'Work A', '/a', '2024-01-01'), ('b', 'lib', 'Work B', '/b', '2024-01-01')`,
		`INSERT INTO tag (id, name, count) VALUES (1, 'alpha', 1), (2, 'beta', 1)`,
		`INSERT INTO audio_tag (audio_id, tag_id) VALUES ('a', 1), ('b', 2)`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seed audio: %v", err)
		}
	}
}

func bulkAudioRequest(t *testing.T, h *AudioHandler, method, target, body string) *httptest.ResponseRecorder {
	t.Helper()

	ctx := auth.WithLibraryScope(auth.WithUserID(context.Background(), "u1"), auth.LibraryScope{})
	req := httptest.NewRequestWithContext(ctx, method, target, bytes.NewBufferString(body))

	req.Header.Set("content-type", "application/json")

	rec := httptest.NewRecorder()

	switch target {
	case "/api/audio/bulk":
		h.BulkUpdateAudios(rec, req)
	case "/api/audio/bulk/played":
		h.BulkUpdateAudioPlayed(rec, req)
	default:
		t.Fatalf("unknown target %q", target)
	}

	return rec
}

func decodeBulkAudioResult(t *testing.T, rec *httptest.ResponseRecorder) BulkResult {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	var result BulkResult
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	return result
}

func audioTagNames(t *testing.T, db *sql.DB, audioID string) []string {
	t.Helper()

	rows, err := db.QueryContext(
		context.Background(),
		`SELECT tag.name FROM audio_tag JOIN tag ON tag.id = audio_tag.tag_id
		 WHERE audio_tag.audio_id = ? ORDER BY tag.name`,
		audioID,
	)
	if err != nil {
		t.Fatalf("query tags: %v", err)
	}

	defer func() { _ = rows.Close() }()

	var names []string

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan tag: %v", err)
		}

		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tags: %v", err)
	}

	return names
}

func TestBulkUpdateAudiosMergesRelations(t *testing.T) {
	t.Parallel()

	h, db := newBulkAudioHandler(t)
	seedBulkAudio(t, db)

	rec := bulkAudioRequest(t, h, http.MethodPatch, "/api/audio/bulk",
		`{"ids":["a","b"],"tags":["shared"]}`)
	result := decodeBulkAudioResult(t, rec)

	if result.Requested != 2 || result.Succeeded != 2 || len(result.Failed) != 0 {
		t.Fatalf("result = %+v, want requested 2, succeeded 2, no failures", result)
	}

	wantTags := map[string][]string{
		"a": {"alpha", "shared"},
		"b": {"beta", "shared"},
	}

	for id, want := range wantTags {
		got := audioTagNames(t, db, id)

		if len(got) != len(want) {
			t.Fatalf("tags for %s = %v, want %v", id, got, want)
		}

		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("tags for %s = %v, want %v", id, got, want)
			}
		}
	}

	for wantID, wantTitle := range map[string]string{"a": "Work A", "b": "Work B"} {
		var title string

		if err := db.QueryRowContext(context.Background(), `SELECT title FROM audio WHERE id = ?`, wantID).Scan(&title); err != nil {
			t.Fatalf("get title %s: %v", wantID, err)
		}

		if title != wantTitle {
			t.Errorf("title for %s = %q, want %q", wantID, title, wantTitle)
		}
	}
}

func TestBulkUpdateAudiosSetsLanguageAndCircle(t *testing.T) {
	t.Parallel()

	h, db := newBulkAudioHandler(t)
	seedBulkAudio(t, db)

	rec := bulkAudioRequest(t, h, http.MethodPatch, "/api/audio/bulk",
		`{"ids":["a","b"],"language":"japanese","circle":"New Circle"}`)
	result := decodeBulkAudioResult(t, rec)

	if result.Succeeded != 2 {
		t.Fatalf("succeeded = %d, want 2; failures: %+v", result.Succeeded, result.Failed)
	}

	for _, id := range []string{"a", "b"} {
		var language, circle string

		if err := db.QueryRowContext(
			context.Background(),
			`SELECT audio.language, circle.name FROM audio
			 JOIN audio_circle ON audio_circle.audio_id = audio.id
			 JOIN circle ON circle.id = audio_circle.circle_id
			 WHERE audio.id = ?`, id,
		).Scan(&language, &circle); err != nil {
			t.Fatalf("get metadata for %s: %v", id, err)
		}

		if language != "japanese" || circle != "New Circle" {
			t.Errorf("%s = language %q, circle %q; want japanese / New Circle", id, language, circle)
		}
	}

	// "" clears the circle while the omitted language stays put.
	rec = bulkAudioRequest(t, h, http.MethodPatch, "/api/audio/bulk", `{"ids":["a","b"],"circle":""}`)
	result = decodeBulkAudioResult(t, rec)

	if result.Succeeded != 2 {
		t.Fatalf("succeeded = %d, want 2; failures: %+v", result.Succeeded, result.Failed)
	}

	var circleRows int

	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM audio_circle`).Scan(&circleRows); err != nil {
		t.Fatalf("count circles: %v", err)
	}

	if circleRows != 0 {
		t.Errorf("audio_circle rows = %d, want 0", circleRows)
	}

	for _, id := range []string{"a", "b"} {
		var language string

		if err := db.QueryRowContext(context.Background(), `SELECT language FROM audio WHERE id = ?`, id).Scan(&language); err != nil {
			t.Fatalf("get language for %s: %v", id, err)
		}

		if language != "japanese" {
			t.Errorf("language for %s = %q, want japanese", id, language)
		}
	}
}

func TestBulkUpdateAudiosClearsLanguageToNull(t *testing.T) {
	t.Parallel()

	h, db := newBulkAudioHandler(t)
	seedBulkAudio(t, db)

	rec := bulkAudioRequest(t, h, http.MethodPatch, "/api/audio/bulk",
		`{"ids":["a","b"],"language":"japanese"}`)
	if result := decodeBulkAudioResult(t, rec); result.Succeeded != 2 {
		t.Fatalf("set language succeeded = %d, want 2; failures: %+v", result.Succeeded, result.Failed)
	}

	// "" clears the language to NULL rather than storing an empty code, which
	// the language facet would otherwise advertise as a blank option.
	rec = bulkAudioRequest(t, h, http.MethodPatch, "/api/audio/bulk", `{"ids":["a","b"],"language":""}`)
	result := decodeBulkAudioResult(t, rec)

	if result.Succeeded != 2 {
		t.Fatalf("clear language succeeded = %d, want 2; failures: %+v", result.Succeeded, result.Failed)
	}

	for _, id := range []string{"a", "b"} {
		var language sql.NullString

		if err := db.QueryRowContext(context.Background(), `SELECT language FROM audio WHERE id = ?`, id).Scan(&language); err != nil {
			t.Fatalf("get language for %s: %v", id, err)
		}

		if language.Valid {
			t.Errorf("language for %s = %q, want NULL", id, language.String)
		}
	}
}

func TestBulkUpdateAudiosReportsMissingIDs(t *testing.T) {
	t.Parallel()

	h, db := newBulkAudioHandler(t)
	seedBulkAudio(t, db)

	rec := bulkAudioRequest(t, h, http.MethodPatch, "/api/audio/bulk",
		`{"ids":["a","missing"],"tags":["shared"]}`)
	result := decodeBulkAudioResult(t, rec)

	if result.Requested != 2 || result.Succeeded != 1 {
		t.Fatalf("result = %+v, want requested 2, succeeded 1", result)
	}

	if len(result.Failed) != 1 || result.Failed[0].ID != "missing" || result.Failed[0].Error != "audio not found" {
		t.Fatalf("failures = %+v, want one audio not found for missing", result.Failed)
	}
}

func TestBulkUpdateAudioPlayed(t *testing.T) {
	t.Parallel()

	h, db := newBulkAudioHandler(t)
	seedBulkAudio(t, db)

	rec := bulkAudioRequest(t, h, http.MethodPut, "/api/audio/bulk/played",
		`{"ids":["a","b"],"played":true}`)
	result := decodeBulkAudioResult(t, rec)

	if result.Requested != 2 || result.Succeeded != 2 {
		t.Fatalf("result = %+v, want requested 2, succeeded 2", result)
	}

	var played, finished int

	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*), coalesce(SUM(completed = 1), 0) FROM audio_played WHERE user_id = 'u1'`,
	).Scan(&played, &finished); err != nil {
		t.Fatalf("count played: %v", err)
	}

	if played != 2 || finished != 2 {
		t.Fatalf("played rows = %d, finished = %d, want 2 and 2", played, finished)
	}

	rec = bulkAudioRequest(t, h, http.MethodPut, "/api/audio/bulk/played",
		`{"ids":["a","b"],"played":false}`)
	result = decodeBulkAudioResult(t, rec)

	if result.Succeeded != 2 {
		t.Fatalf("unplayed succeeded = %d, want 2", result.Succeeded)
	}

	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM audio_played WHERE user_id = 'u1'`).Scan(&played); err != nil {
		t.Fatalf("count played: %v", err)
	}

	if played != 0 {
		t.Errorf("played rows = %d, want 0", played)
	}
}
