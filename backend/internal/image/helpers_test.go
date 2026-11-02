package image

import (
	"bytes"
	"context"
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

func newTestProcessor(t *testing.T) *Processor {
	t.Helper()

	return NewProcessor(t.TempDir(), slog.New(slog.DiscardHandler))
}

func TestGenerateAudioCover(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	img := stdimage.NewRGBA(stdimage.Rect(0, 0, 40, 30))

	for x := range 40 {
		for y := range 30 {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 0x40, A: 0xff})
		}
	}

	var src bytes.Buffer
	if err := png.Encode(&src, img); err != nil {
		t.Fatalf("encode source png: %v", err)
	}

	if err := p.GenerateAudioCover(context.Background(), "work", &src); err != nil {
		t.Fatalf("generate audio cover: %v", err)
	}

	cover := p.CachedAudioCover("work")
	if cover.Path == "" {
		t.Fatal("cover path is empty after a successful render")
	}

	if cover.Width != 40 || cover.Height != 30 {
		t.Errorf("recorded cover size = %dx%d, want 40x30", cover.Width, cover.Height)
	}

	if AudioCoverStale(p.cacheDir, "work") {
		t.Error("a successfully rendered cover is reported stale")
	}
}

func TestGenerateAudioCoverFailureLeavesNoMarker(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	// The first render establishes a cover and marker; the second must fail
	// and leave the recorded metadata alone rather than tainting it.
	img := stdimage.NewRGBA(stdimage.Rect(0, 0, 8, 6))

	var src bytes.Buffer
	if err := png.Encode(&src, img); err != nil {
		t.Fatalf("encode source png: %v", err)
	}

	if err := p.GenerateAudioCover(context.Background(), "work", &src); err != nil {
		t.Fatalf("generate audio cover: %v", err)
	}

	if err := p.GenerateAudioCover(context.Background(), "work", bytes.NewReader([]byte("not an image"))); err == nil {
		t.Fatal("generating a cover from invalid data succeeded")
	}

	if AudioCoverStale(p.cacheDir, "work") {
		t.Error("a failed re-render marked the still-valid cover stale")
	}
}

func TestAudioCoverStale(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	if !AudioCoverStale(p.cacheDir, "work") {
		t.Error("a cover that was never rendered is reported fresh")
	}

	if err := os.MkdirAll(p.ThumbDir(AudioCacheKey("work")), 0o700); err != nil {
		t.Fatalf("create thumb dir: %v", err)
	}

	if err := os.WriteFile(audioCoverMarkerPath(p.cacheDir, "work"), []byte(audioCoverMarker()), 0o600); err != nil {
		t.Fatalf("write cover marker: %v", err)
	}

	if !AudioCoverStale(p.cacheDir, "work") {
		t.Error("a marker without a cover image is reported fresh")
	}

	if err := os.WriteFile(thumbPathIn(p.cacheDir, AudioCacheKey("work"), 0), []byte("image"), 0o600); err != nil {
		t.Fatalf("write cover image: %v", err)
	}

	if AudioCoverStale(p.cacheDir, "work") {
		t.Error("an existing cover with the current settings is reported stale")
	}

	if err := os.WriteFile(audioCoverMarkerPath(p.cacheDir, "work"), []byte("360x75"), 0o600); err != nil {
		t.Fatalf("overwrite cover marker: %v", err)
	}

	if !AudioCoverStale(p.cacheDir, "work") {
		t.Error("a cover rendered with older settings is reported fresh")
	}

	if err := os.Remove(thumbPathIn(p.cacheDir, AudioCacheKey("work"), 0)); err != nil {
		t.Fatalf("remove cover image: %v", err)
	}

	if !AudioCoverStale(p.cacheDir, "work") {
		t.Error("a cover image lost behind a current marker is reported fresh")
	}
}

func TestCachedAudioCover(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	if cover := p.CachedAudioCover("work"); cover.Path != "" || cover.ModTime != nil || cover.Width != 0 {
		t.Errorf("cover for an unrendered work = %+v, want zero value", cover)
	}

	if err := os.MkdirAll(p.ThumbDir(AudioCacheKey("work")), 0o700); err != nil {
		t.Fatalf("create thumb dir: %v", err)
	}

	wantPath := thumbPathIn(p.cacheDir, AudioCacheKey("work"), 0)
	if err := os.WriteFile(wantPath, []byte("not an image"), 0o600); err != nil {
		t.Fatalf("write cover image: %v", err)
	}

	cover := p.CachedAudioCover("work")
	if cover.Path != wantPath {
		t.Errorf("cover path = %q, want %q", cover.Path, wantPath)
	}

	if cover.ModTime == nil {
		t.Error("cover mtime is nil for an existing cover")
	}

	if cover.Width != 0 || cover.Height != 0 {
		t.Errorf("dimensions for an undecodable cover = %dx%d, want 0x0", cover.Width, cover.Height)
	}
}

func TestThumbsReady(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	if p.ThumbsReady("archive", 3) {
		t.Error("archive the job never finished is reported ready")
	}

	if err := p.MarkThumbsReady("archive", 3); err != nil {
		t.Fatalf("mark thumbs ready: %v", err)
	}

	if !p.ThumbsReady("archive", 3) {
		t.Error("archive the job finished is reported not ready")
	}

	// The marker records the count the job thumbnailed - the file's - while
	// callers pass the stored archive.page_count, which a metadata source can
	// overwrite with a lower one. Every page the lower count names does have a
	// thumbnail, so that archive is still ready.
	if !p.ThumbsReady("archive", 1) {
		t.Error("a page count below the thumbnailed one is reported not ready")
	}

	if p.ThumbsReady("archive", 4) {
		t.Error("marker for a 3-page archive vouches for a 4-page one")
	}

	if !p.ThumbsReady("archive", 0) {
		t.Error("an archive with no pages has nothing to wait for")
	}
}

func TestThumbsReadyRejectsUnreadableMarker(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	if err := os.MkdirAll(p.ThumbDir("archive"), 0o700); err != nil {
		t.Fatalf("create thumb dir: %v", err)
	}

	if err := os.WriteFile(p.thumbReadyMarker("archive"), []byte("not a count"), 0o600); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	if p.ThumbsReady("archive", 1) {
		t.Error("a marker that doesn't hold a count is reported ready")
	}
}

// TestMarkThumbsReadyLeavesNoTempFiles guards the atomic write: the marker the
// archive list reads has to be complete the moment it exists, and it must not
// leave litter behind in the thumbnail directory either.
func TestMarkThumbsReadyLeavesNoTempFiles(t *testing.T) {
	t.Parallel()

	p := newTestProcessor(t)

	if err := p.MarkThumbsReady("archive", 2); err != nil {
		t.Fatalf("mark thumbs ready: %v", err)
	}

	entries, err := os.ReadDir(p.ThumbDir("archive"))
	if err != nil {
		t.Fatalf("read thumb dir: %v", err)
	}

	if len(entries) != 1 || entries[0].Name() != filepath.Base(p.thumbReadyMarker("archive")) {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}

		t.Errorf("thumb dir holds %v, want only the marker", names)
	}
}
