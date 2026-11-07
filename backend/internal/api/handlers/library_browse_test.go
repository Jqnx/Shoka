package handlers

import (
	"Shoka/internal/api/response"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestBrowseLibraryPath(t *testing.T) {
	t.Parallel()

	dir, _ := newBrowseFixture(t)

	tests := []struct {
		name         string
		query        string
		wantPath     string
		wantParent   string
		wantNoParent bool
		wantNames    []string
	}{
		{
			name:         "defaults to root",
			wantPath:     "/",
			wantNoParent: true,
		},
		{
			name:         "empty path is root",
			query:        "?path=",
			wantPath:     "/",
			wantNoParent: true,
		},
		{
			name:       "lists directories and follows directory symlinks",
			query:      "?path=" + url.QueryEscape(dir),
			wantPath:   dir,
			wantParent: filepath.Dir(dir),
			wantNames:  []string{"a", "b", "link"},
		},
	}

	handler := &LibraryHandler{logger: slog.New(slog.DiscardHandler)}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := browseRequest(t, handler, tt.query)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusOK, rec.Body)
			}

			body := decodeBrowseResponse(t, rec)
			if body.Path != tt.wantPath {
				t.Fatalf("path = %q, want %q", body.Path, tt.wantPath)
			}

			if body.Parent != tt.wantParent {
				t.Fatalf("parent = %q, want %q", body.Parent, tt.wantParent)
			}

			if body.Entries == nil {
				t.Fatal("entries is null, want an array")
			}

			assertParentKey(t, rec, !tt.wantNoParent)

			if tt.wantNames != nil {
				assertEntryNames(t, body.Entries, tt.wantPath, tt.wantNames)
			}
		})
	}
}

func TestBrowseLibraryPathInvalid(t *testing.T) {
	t.Parallel()

	_, filePath := newBrowseFixture(t)

	tests := []struct {
		name  string
		query string
	}{
		{name: "missing path", query: "?path=" + url.QueryEscape("/nonexistent-xyz-123")},
		{name: "file path", query: "?path=" + url.QueryEscape(filePath)},
	}

	handler := &LibraryHandler{logger: slog.New(slog.DiscardHandler)}

	const wantMessage = "path does not exist or is not a directory"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := browseRequest(t, handler, tt.query)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusBadRequest, rec.Body)
			}

			var body response.Error
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error body: %v", err)
			}

			if body.Message != wantMessage {
				t.Fatalf("message = %q, want %q", body.Message, wantMessage)
			}
		})
	}
}

func TestBrowseLibraryPathPermissionDenied(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permissions")
	}

	locked := filepath.Join(t.TempDir(), "locked")
	if err := os.Mkdir(locked, 0o300); err != nil {
		t.Fatalf("create locked dir: %v", err)
	}

	handler := &LibraryHandler{logger: slog.New(slog.DiscardHandler)}

	rec := browseRequest(t, handler, "?path="+url.QueryEscape(locked))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (body %s)", rec.Code, http.StatusForbidden, rec.Body)
	}

	var body response.Error
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}

	const wantMessage = "permission denied reading directory"
	if body.Message != wantMessage {
		t.Fatalf("message = %q, want %q", body.Message, wantMessage)
	}
}

// newBrowseFixture builds a directory holding subdirectories a and b, a file,
// a symlink to a directory, and a symlink to a file.
func newBrowseFixture(t *testing.T) (dir, filePath string) {
	t.Helper()

	dir = t.TempDir()

	for _, name := range []string{"a", "b"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o750); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	filePath = filepath.Join(dir, "c.txt")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("create file: %v", err)
	}

	// A symlink to a directory is browsable; a link to a file is not.
	if err := os.Symlink(filepath.Join(dir, "a"), filepath.Join(dir, "link")); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	if err := os.Symlink(filePath, filepath.Join(dir, "linkfile")); err != nil {
		t.Fatalf("create file symlink: %v", err)
	}

	return dir, filePath
}

func browseRequest(t *testing.T, handler *LibraryHandler, query string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		"/api/admin/libraries/browse"+query,
		nil,
	)
	rec := httptest.NewRecorder()
	handler.BrowseLibraryPath(rec, req)

	return rec
}

func decodeBrowseResponse(t *testing.T, rec *httptest.ResponseRecorder) LibraryBrowseResponse {
	t.Helper()

	var body LibraryBrowseResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	return body
}

// assertParentKey checks whether the JSON body carries a "parent" key at all;
// root listings omit it rather than sending an empty string.
func assertParentKey(t *testing.T, rec *httptest.ResponseRecorder, want bool) {
	t.Helper()

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw body: %v", err)
	}

	if _, ok := raw["parent"]; ok != want {
		t.Fatalf("parent key present = %t, want %t", ok, want)
	}
}

func assertEntryNames(t *testing.T, entries []LibraryBrowseEntry, base string, want []string) {
	t.Helper()

	if len(entries) != len(want) {
		t.Fatalf("entries = %v, want names %v", entries, want)
	}

	for i, name := range want {
		if entries[i].Name != name {
			t.Fatalf("entry %d name = %q, want %q", i, entries[i].Name, name)
		}

		if wantPath := filepath.Join(base, name); entries[i].Path != wantPath {
			t.Fatalf("entry %d path = %q, want %q", i, entries[i].Path, wantPath)
		}
	}
}
