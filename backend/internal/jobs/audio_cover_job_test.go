package jobs

import (
	"Shoka/internal/database/sqlc"
	"Shoka/internal/image"
	"bytes"
	stdimage "image"
	"image/color"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// audioCoverFixture seeds one audio work with two differently sized PNGs so a
// test can tell which one the render used.
type audioCoverFixture struct {
	queries   *sqlc.Queries
	processor *image.Processor
	folder    string
}

func newAudioCoverFixture(t *testing.T) audioCoverFixture {
	t.Helper()

	db := openMemoryDB(t, audioSchema)
	folder := t.TempDir()

	if _, err := db.ExecContext(t.Context(),
		`INSERT INTO audio (id, library_id, title, folder_path, mod_time) VALUES (?, ?, ?, ?, ?)`,
		audioTestWorkID, "lib1", "work", folder, "2026-01-01 00:00:00"); err != nil {
		t.Fatalf("seed audio: %v", err)
	}

	for index, spec := range []struct {
		name          string
		width, height int
	}{
		{name: "001.png", width: 64, height: 48},
		{name: "002.png", width: 120, height: 80},
	} {
		img := stdimage.NewRGBA(stdimage.Rect(0, 0, spec.width, spec.height))

		for x := range spec.width {
			for y := range spec.height {
				img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 0x80, A: 0xff})
			}
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			t.Fatalf("encode image: %v", err)
		}

		if err := os.WriteFile(filepath.Join(folder, spec.name), buf.Bytes(), 0o600); err != nil {
			t.Fatalf("write image: %v", err)
		}

		if _, err := db.ExecContext(t.Context(),
			`INSERT INTO audio_image (audio_id, relative_path, sort_order) VALUES (?, ?, ?)`,
			audioTestWorkID, spec.name, index); err != nil {
			t.Fatalf("seed image: %v", err)
		}
	}

	return audioCoverFixture{
		queries:   sqlc.New(db),
		processor: image.NewProcessor(t.TempDir(), slog.New(slog.DiscardHandler)),
		folder:    folder,
	}
}

func TestAudioCoverJobUsesChosenImage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cover     int64
		wantWidth int
	}{
		{name: "chosen second image", cover: 1, wantWidth: 120},
		{name: "out-of-range falls back to the first image", cover: 9, wantWidth: 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fx := newAudioCoverFixture(t)
			if err := fx.queries.UpdateAudioCoverIndex(t.Context(), sqlc.UpdateAudioCoverIndexParams{
				ID:         audioTestWorkID,
				CoverIndex: tt.cover,
			}); err != nil {
				t.Fatalf("set cover index: %v", err)
			}

			handler := NewAudioCoverHandler(fx.processor, fx.queries, slog.New(slog.DiscardHandler))
			if err := handler(t.Context(), newJob(t, JobTypeAudioCover, AudioCoverPayload{AudioID: audioTestWorkID})); err != nil {
				t.Fatalf("run audio cover job: %v", err)
			}

			if cover := fx.processor.CachedAudioCover(audioTestWorkID); cover.Width != tt.wantWidth {
				t.Errorf("cover width = %d, want %d", cover.Width, tt.wantWidth)
			}
		})
	}
}

func TestAudioCoverCurrentDetectsChangedSelection(t *testing.T) {
	t.Parallel()

	fx := newAudioCoverFixture(t)
	handler := NewAudioCoverHandler(fx.processor, fx.queries, slog.New(slog.DiscardHandler))

	if err := handler(t.Context(), newJob(t, JobTypeAudioCover, AudioCoverPayload{AudioID: audioTestWorkID})); err != nil {
		t.Fatalf("run audio cover job: %v", err)
	}

	current, err := audioCoverCurrent(t.Context(), fx.queries, audioTestWorkID, "001.png")
	if err != nil || !current {
		t.Fatalf("rendered image 001.png reported current=%t err=%v, want true", current, err)
	}

	if err := fx.queries.UpdateAudioCoverIndex(t.Context(), sqlc.UpdateAudioCoverIndexParams{
		ID:         audioTestWorkID,
		CoverIndex: 1,
	}); err != nil {
		t.Fatalf("change cover index: %v", err)
	}

	// The render was made from 001.png while the row now names 002.png.
	current, err = audioCoverCurrent(t.Context(), fx.queries, audioTestWorkID, "001.png")
	if err != nil {
		t.Fatalf("recheck selection: %v", err)
	}

	if current {
		t.Fatal("audioCoverCurrent accepted a render for the previous selection")
	}

	// A second run renders the new selection, which then reports current.
	if err := handler(t.Context(), newJob(t, JobTypeAudioCover, AudioCoverPayload{AudioID: audioTestWorkID})); err != nil {
		t.Fatalf("rerun audio cover job: %v", err)
	}

	if current, err := audioCoverCurrent(t.Context(), fx.queries, audioTestWorkID, "002.png"); err != nil || !current {
		t.Fatalf("rendered image 002.png reported current=%t err=%v, want true", current, err)
	}

	if cover := fx.processor.CachedAudioCover(audioTestWorkID); cover.Width != 120 {
		t.Errorf("cover width = %d, want the chosen image's 120", cover.Width)
	}
}
