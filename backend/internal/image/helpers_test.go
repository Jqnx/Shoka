package image

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func newTestProcessor(t *testing.T) *Processor {
	t.Helper()

	return NewProcessor(t.TempDir(), slog.New(slog.DiscardHandler))
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
