package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/events"
	"Shoka/internal/image"
	"bytes"
	stdimage "image"
	"image/color"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestMain(m *testing.M) {
	_ = vips.Startup(nil)
	code := m.Run()

	vips.Shutdown()

	os.Exit(code)
}

// audioSchema mirrors migration 00044 for the two tables the audio thumbnail
// job reads: GetAudioByID and GetAudioImages both select *, in table order.
const audioSchema = `
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

// newAudioThumbFixture seeds one audio work rooted at a temp dir holding two
// synthetic PNGs, mimicking the scanner's image rows.
func newAudioThumbFixture(t *testing.T) (*sqlc.Queries, *image.Processor, string) {
	t.Helper()

	db := openMemoryDB(t, audioSchema)
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

	for index, name := range []string{"001.png", "002.png"} {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatalf("encode image %d: %v", index, err)
		}

		if err := os.WriteFile(filepath.Join(folder, name), buf.Bytes(), 0o600); err != nil {
			t.Fatalf("write image %d: %v", index, err)
		}

		if _, err := db.ExecContext(t.Context(),
			`INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES (?, ?, ?)`,
			"work1", name, index); err != nil {
			t.Fatalf("seed image %d: %v", index, err)
		}
	}

	return sqlc.New(db), image.NewProcessor(t.TempDir(), slog.New(slog.DiscardHandler)), folder
}

// runAudioThumbJob runs the handler once and returns the ready indexes it
// published plus the terminal event.
func runAudioThumbJob(t *testing.T, queries *sqlc.Queries, processor *image.Processor) ([]int, events.ThumbnailEvent) {
	t.Helper()

	broadcaster := events.NewThumbnailBroadcaster()

	ch, cancel := broadcaster.Subscribe(image.AudioCacheKey("work1"))

	defer cancel()

	handler := NewAudioThumbnailHandler(processor, broadcaster, queries, slog.New(slog.DiscardHandler))
	if err := handler(t.Context(), newJob(t, JobTypeAudioThumbnail, AudioThumbnailPayload{AudioID: "work1"})); err != nil {
		t.Fatalf("run audio thumbnail job: %v", err)
	}

	var indexes []int

	for {
		event, ok := <-ch
		if !ok {
			t.Fatal("broadcaster closed before a done event")
		}

		if event.Done {
			return indexes, event
		}

		indexes = append(indexes, event.Index)
	}
}

func TestAudioThumbnailHandler(t *testing.T) {
	t.Parallel()

	queries, processor, _ := newAudioThumbFixture(t)

	indexes, done := runAudioThumbJob(t, queries, processor)
	if done.Error != "" {
		t.Fatalf("done event carries an error: %q", done.Error)
	}

	if len(indexes) != 2 || indexes[0] != 0 || indexes[1] != 1 {
		t.Errorf("ready indexes = %v, want [0 1]", indexes)
	}

	for index := range 2 {
		if path := processor.AudioThumbPath("work1", index); path == "" {
			t.Fatalf("thumbnail %d missing after generation", index)
		}
	}

	if !processor.AudioGalleryReady("work1", 2) {
		t.Error("gallery is not ready after a complete run")
	}

	// A second run skips existing thumbs but still reports both as ready
	// before the terminal event.
	indexes, done = runAudioThumbJob(t, queries, processor)
	if done.Error != "" {
		t.Fatalf("second run done event carries an error: %q", done.Error)
	}

	if len(indexes) != 2 || indexes[0] != 0 || indexes[1] != 1 {
		t.Errorf("skip-path ready indexes = %v, want [0 1]", indexes)
	}
}

func TestAudioThumbnailHandlerReportsFailedImage(t *testing.T) {
	t.Parallel()

	queries, processor, folder := newAudioThumbFixture(t)

	if _, done := runAudioThumbJob(t, queries, processor); done.Error != "" {
		t.Fatalf("initial run done event carries an error: %q", done.Error)
	}

	// Corrupt image 1 and drop its thumbnail, so this run has to retry it.
	if err := os.WriteFile(filepath.Join(folder, "002.png"), []byte("not an image"), 0o600); err != nil {
		t.Fatalf("corrupt image: %v", err)
	}

	if err := os.Remove(filepath.Join(processor.AudioGalleryDir("work1"), "002.webp")); err != nil {
		t.Fatalf("remove thumbnail: %v", err)
	}

	indexes, done := runAudioThumbJob(t, queries, processor)
	if done.Error == "" {
		t.Fatal("done event has no error after an undecodable image")
	}

	if len(indexes) != 1 || indexes[0] != 0 {
		t.Errorf("ready indexes = %v, want [0]", indexes)
	}

	if path := processor.AudioThumbPath("work1", 0); path == "" {
		t.Error("one undecodable image removed the other work's thumbnail")
	}
}
