package handlers

import (
	"Shoka/internal/database/sqlc"
	"net/http/httptest"
	"testing"
)

func TestParseDefaultAudioFolder(t *testing.T) {
	t.Parallel()

	audioLib := sqlc.Library{Type: "audio", DefaultAudioFolder: "SEなし"}
	doujinshiLib := sqlc.Library{Type: "doujinshi"}

	tests := []struct {
		name    string
		raw     *string
		current sqlc.Library
		want    string
		ok      bool
	}{
		{"omitted keeps stored", nil, audioLib, "SEなし", true},
		{"trims", new("  SEあり "), audioLib, "SEあり", true},
		{"clears", new(""), audioLib, "", true},
		{"rejects separator", new("SE/なし"), audioLib, "", false},
		{"rejects dot", new("."), audioLib, "", false},
		{"rejects parent reference", new(".."), audioLib, "", false},
		{"non-audio empty is harmless", new(""), doujinshiLib, "", true},
		{"rejects value on non-audio", new("SEなし"), doujinshiLib, "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()

			got, ok := parseDefaultAudioFolder(w, tc.raw, tc.current)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v (status %d)", ok, tc.ok, w.Code)
			}

			if got != tc.want {
				t.Errorf("folder = %q, want %q", got, tc.want)
			}

			if !ok && w.Code != 400 {
				t.Errorf("status = %d, want 400", w.Code)
			}
		})
	}
}
