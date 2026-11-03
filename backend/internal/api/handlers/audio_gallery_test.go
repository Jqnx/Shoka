package handlers

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"bytes"
	"context"
	"database/sql"
	stdimage "image"
	"image/color"
	"image/png"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"
)

// audioGallerySchema is the part of migration 00044 the gallery handlers read.
const audioGallerySchema = `
CREATE TABLE audio (
    id TEXT PRIMARY KEY NOT NULL,
    library_id TEXT NOT NULL,
    title TEXT NOT NULL,
    folder_path TEXT UNIQUE NOT NULL,
    mod_time DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    release_date DATETIME,
    language TEXT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE audio_image (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    audio_id TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    UNIQUE (audio_id, relative_path)
);
`

// newAudioGalleryHandler seeds work "work1" with one 64x48 PNG.
func newAudioGalleryHandler(t *testing.T) (*AudioHandler, *image.Processor, string) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(t.Context(), audioGallerySchema); err != nil {
		t.Fatalf("schema: %v", err)
	}

	folder := t.TempDir()

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO audio (id, library_id, title, folder_path, mod_time) VALUES (?, ?, ?, ?, ?)`,
		"work1", "lib1", "work", folder, "2026-01-01 00:00:00"); err != nil {
		t.Fatalf("seed audio: %v", err)
	}

	img := stdimage.NewRGBA(stdimage.Rect(0, 0, 64, 48))

	for x := range 64 {
		for y := range 48 {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 0x40, A: 0xff})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode image: %v", err)
	}

	if err := os.WriteFile(filepath.Join(folder, "001.png"), buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write image: %v", err)
	}

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES (?, ?, ?)`,
		"work1", "001.png", 0); err != nil {
		t.Fatalf("seed image: %v", err)
	}

	logger := slog.New(slog.DiscardHandler)
	processor := image.NewProcessor(t.TempDir(), logger)

	cache, err := image.NewCache(processor, logger)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}

	return NewAudioHandler(sqlc.New(db), db, processor, cache, nil, nil, logger), processor, folder
}

// audioGalleryGet builds a request carrying chi's id/index path params.
func audioGalleryGet(t *testing.T, id, index string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/audio/"+id+"/images/"+index, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	rctx.URLParams.Add("index", index)

	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestAudioGetImageThumbnail(t *testing.T) {
	t.Parallel()

	h, processor, folder := newAudioGalleryHandler(t)

	before := httptest.NewRecorder()
	h.GetImageThumbnail(before, audioGalleryGet(t, "work1", "0"))

	if before.Code != http.StatusNotFound {
		t.Errorf("status before generation = %d, want 404", before.Code)
	}

	if got := before.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control before generation = %q, want no-store", got)
	}

	src, err := os.Open(filepath.Join(folder, "001.png")) //nolint:gosec // test fixture path
	if err != nil {
		t.Fatalf("open source image: %v", err)
	}

	err = processor.GenerateAudioThumb(context.Background(), "work1", 0, src)
	_ = src.Close()

	if err != nil {
		t.Fatalf("generate thumbnail: %v", err)
	}

	after := httptest.NewRecorder()
	h.GetImageThumbnail(after, audioGalleryGet(t, "work1", "0"))

	if after.Code != http.StatusOK {
		t.Fatalf("status after generation = %d, want 200", after.Code)
	}

	if got := after.Header().Get("Content-Type"); got != "image/webp" {
		t.Errorf("Content-Type = %q, want image/webp", got)
	}

	if got := after.Header().Get("Cache-Control"); got != "private, max-age=86400" {
		t.Errorf("Cache-Control = %q, want private, max-age=86400", got)
	}

	if after.Body.Len() == 0 {
		t.Error("thumbnail body is empty")
	}

	outOfRange := httptest.NewRecorder()
	h.GetImageThumbnail(outOfRange, audioGalleryGet(t, "work1", "99"))

	if outOfRange.Code != http.StatusNotFound {
		t.Errorf("status for index 99 = %d, want 404", outOfRange.Code)
	}
}

func TestAudioGetImagePreview(t *testing.T) {
	t.Parallel()

	h, _, folder := newAudioGalleryHandler(t)

	first := httptest.NewRecorder()
	h.GetImagePreview(first, audioGalleryGet(t, "work1", "0"))

	if first.Code != http.StatusOK {
		t.Fatalf("preview status = %d, want 200", first.Code)
	}

	if got := first.Header().Get("Content-Type"); got != "image/webp" {
		t.Errorf("preview Content-Type = %q, want image/webp", got)
	}

	if first.Body.Len() == 0 {
		t.Fatal("preview body is empty")
	}

	// The second request is served from the in-memory LRU: removing the source
	// must not turn it into a 404.
	if err := os.Remove(filepath.Join(folder, "001.png")); err != nil {
		t.Fatalf("remove source image: %v", err)
	}

	second := httptest.NewRecorder()
	h.GetImagePreview(second, audioGalleryGet(t, "work1", "0"))

	if second.Code != http.StatusOK {
		t.Fatalf("cached preview status = %d, want 200", second.Code)
	}

	if !bytes.Equal(first.Body.Bytes(), second.Body.Bytes()) {
		t.Error("cached preview bytes differ from the rendered preview")
	}
}
