package handlers

import "testing"

func TestTrackFolder(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"chapter.wav":                 "",
		"SEあり/chapter.wav":            "SEあり",
		"SEなし/disc1/chapter.wav":      "SEなし",
		"01_本編『SEありWAV』/track 01.wav": "01_本編『SEありWAV』",
	}

	for path, want := range tests {
		if got := trackFolder(path); got != want {
			t.Errorf("trackFolder(%q) = %q, want %q", path, got, want)
		}
	}
}
