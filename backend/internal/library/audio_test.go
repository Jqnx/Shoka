package library

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestClassifyFiles(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		stored  []string
		scanned []string
		want    audioChange
	}{
		{"same order", []string{"1", "2", "10"}, []string{"1", "2", "10"}, audioUnchanged},
		{"reordered", []string{"1", "10", "2"}, []string{"1", "2", "10"}, audioReordered},
		{"file replaced", []string{"1", "2"}, []string{"1", "3"}, audioChanged},
		{"file added", []string{"1"}, []string{"1", "2"}, audioChanged},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := classifyFiles(tc.stored, tc.scanned); got != tc.want {
				t.Errorf("classifyFiles(%v, %v) = %v, want %v", tc.stored, tc.scanned, got, tc.want)
			}
		})
	}
}

// TestScanWorkNaturalTrackOrder pins numeric ordering inside a work: a rescan
// must keep "第2章" before "第10章" so the player lists tracks in playback
// order rather than lexicographically.
func TestScanWorkNaturalTrackOrder(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	names := []string{"第10章.wav", "第2章.wav", "第1章.wav", "第20章.wav"}

	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("audio"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	var scanner AudioScanner

	work, err := scanner.scanWork(dir)
	if err != nil {
		t.Fatalf("scan work: %v", err)
	}

	want := []string{"第1章.wav", "第2章.wav", "第10章.wav", "第20章.wav"}
	if !slices.Equal(work.audio, want) {
		t.Errorf("track order = %v, want %v", work.audio, want)
	}
}
